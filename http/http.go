package http

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

var (
	crlf       = []byte("\r\n")
	colonSpace = []byte(": ")
)

// NewConn is a constructor of minws HTTP connection
func NewConn(rwc net.Conn) *Conn {
	return &Conn{
		rwc: rwc,
		r:   bufio.NewReader(rwc),
	}
}

// Conn is server-side HTTP connection
type Conn struct {
	rwc net.Conn
	r   *bufio.Reader
}

// ReadRequest parses request
func (c *Conn) ReadRequest() (*Response, error) {
	req := &Request{}

	firstLine, err := c.readLineByteSlice()
	if err != nil {
		return nil, err
	}
	var ok bool
	req.Method, req.RequestURI, req.Proto, ok = parseRequestLine(string(firstLine))
	if !ok {
		return nil, fmt.Errorf("Invalid first line %s", string(firstLine))
	}
	if req.ProtoMajor, req.ProtoMinor, ok = parseHTTPVersion(req.Proto); !ok {
		return nil, fmt.Errorf("Invalid proto version %s", req.Proto)
	}

	req.Header, err = c.readHeader()
	if err != nil {
		return nil, err
	}

	res := &Response{
		Req:    req,
		w:      bufio.NewWriter(c.rwc),
		c:      c,
		header: make(Header),
	}

	return res, nil
}

func (c *Conn) readLineByteSlice() ([]byte, error) {
	var line []byte
	for {
		l, more, err := c.r.ReadLine()
		if err != nil {
			return nil, err
		}
		line = append(line, l...)
		if !more {
			break
		}
	}
	return line, nil
}

func (c *Conn) readHeader() (Header, error) {
	res := make(Header)
	for {
		kv, err := c.readLineByteSlice()
		if len(kv) == 0 {
			return res, err
		}

		i := bytes.IndexByte(kv, ':')
		if i < 0 {
			return res, fmt.Errorf("Invalid header line: %s", kv)
		}
		// Todo: validation
		key := string(kv[:i])
		if key == "" {
			return res, fmt.Errorf("Invalid header key: %s", kv)
		}
		// skip initial spaces in value
		i++ // skip colon
		for i < len(kv) && (kv[i] == ' ' || kv[i] == '\t') {
			i++
		}
		value := string(kv[i:])
		res[key] = value

		if err != nil {
			return res, err
		}
	}
}

func parseRequestLine(line string) (method, requestURI, proto string, ok bool) {
	s1 := strings.Index(line, " ")
	s2 := strings.Index(line[s1+1:], " ")
	if s1 < 0 || s2 < 0 {
		return
	}
	s2 += s1 + 1
	return line[0:s1], line[s1+1 : s2], line[s2+1:], true
}

func parseHTTPVersion(vers string) (major, minor int, ok bool) {
	Big := 10000
	if !strings.HasPrefix(vers, "HTTP/") {
		return 0, 0, false
	}
	dot := strings.Index(vers, ".")
	if dot < 0 {
		return 0, 0, false
	}
	major, err := strconv.Atoi(vers[5:dot])
	if err != nil || major < 0 || major > Big {
		return 0, 0, false
	}
	minor, err = strconv.Atoi(vers[dot+1:])
	if err != nil || minor < 0 || minor > Big {
		return 0, 0, false
	}
	return major, minor, true
}

func headerValueContainsToken(values []string, token string) bool {
	for _, v := range values {
		// Todo: parse value
		if strings.Index(v, token) != -1 {
			return true
		}
	}
	return false
}

// Request represents HTTP Request
type Request struct {
	Method     string
	RequestURI string
	Proto      string
	ProtoMajor int
	ProtoMinor int
	Header     Header
}

// Header represents HTTP header
type Header map[string]string

// Get returns header value
func (h Header) Get(key string) string {
	if v := h[key]; len(v) > 0 {
		return v
	}
	return ""
}

// Has checks header contains key
func (h Header) Has(key string) bool {
	_, ok := h[key]
	return ok
}

func (h Header) set(key, value string) {
	h[key] = value
}

// ResponseWriter is writer for HTTP response
type ResponseWriter interface {
	Write(p []byte) (n int, err error)
	SetStatus(code int)
	SetHeader(key, value string)
}

// Response represents HTTP Response
type Response struct {
	Req           *Request
	w             *bufio.Writer
	c             *Conn
	wroteHeader   bool
	status        int
	header        Header
	contentLength int
}

func (r *Response) Write(p []byte) (int, error) {
	if !r.wroteHeader {
		if r.status == 0 {
			r.SetStatus(StatusOK)
		}
		r.writeHeader(p)
	}
	i, err := r.w.Write(p)
	if err != nil {
		return i, err
	}
	return i, nil
}

// FinishRequest finalizes request and send respnose to client
func (r *Response) FinishRequest() {
	if !r.wroteHeader {
		if r.status == 0 {
			r.SetStatus(StatusOK)
		}
		r.writeHeader(nil)
	}
	r.w.Flush()
}

// SetStatus set response status code
func (r *Response) SetStatus(code int) {
	r.status = code
}

// SetHeader set response header
func (r *Response) SetHeader(key, value string) {
	r.header.set(key, value)
}

func (r *Response) writeHeader(p []byte) {
	if r.wroteHeader {
		return
	}
	r.wroteHeader = true

	// status line
	ver := "1.1"
	fmt.Fprintf(r.w, "HTTP/%s %d %s\r\n", ver, r.status, statusText[r.status])

	exh := extraHeader{}

	// content-type
	exh.contentType = "text/plain; charset=utf-8"

	// content-length
	r.contentLength = len(p)
	exh.contentLength = strconv.AppendInt(exh.contentLength, int64(len(p)), 10)

	// connection
	exh.connection = "keep-alive"
	if r.header.Has("Connection") {
		exh.connection = r.header.Get("Connection")
	}

	// date
	if r.header.Has("Date") {
		exh.date = []byte(r.header.Get("Date"))
	} else {
		exh.date = appendTime(exh.date, time.Now())
	}

	// other headers
	for k, v := range r.header {
		fmt.Fprintf(r.w, "%s: %s\r\n", k, v)
	}

	exh.Write(r.w)
	r.w.Write(crlf)
}

type extraHeader struct {
	contentType      string
	connection       string
	transferEncoding string
	date             []byte // written if not nil
	contentLength    []byte // written if not nil
}

var extraHeaderKeys = [][]byte{
	[]byte("Content-Type"),
	[]byte("Connection"),
	[]byte("Transfer-Encoding"),
}

var (
	headerContentLength = []byte("Content-Length: ")
	headerDate          = []byte("Date: ")
)

func (h extraHeader) Write(w *bufio.Writer) {
	if h.date != nil {
		w.Write(headerDate)
		w.Write(h.date)
		w.Write(crlf)
	}
	if h.contentLength != nil {
		w.Write(headerContentLength)
		w.Write(h.contentLength)
		w.Write(crlf)
	}
	for i, v := range []string{h.contentType, h.connection, h.transferEncoding} {
		if v != "" {
			w.Write(extraHeaderKeys[i])
			w.Write(colonSpace)
			w.WriteString(v)
			w.Write(crlf)
		}
	}
}

func appendTime(b []byte, t time.Time) []byte {
	const days = "SunMonTueWedThuFriSat"
	const months = "JanFebMarAprMayJunJulAugSepOctNovDec"

	t = t.UTC()
	yy, mm, dd := t.Date()
	hh, mn, ss := t.Clock()
	day := days[3*t.Weekday():]
	mon := months[3*(mm-1):]

	return append(b,
		day[0], day[1], day[2], ',', ' ',
		byte('0'+dd/10), byte('0'+dd%10), ' ',
		mon[0], mon[1], mon[2], ' ',
		byte('0'+yy/1000), byte('0'+(yy/100)%10), byte('0'+(yy/10)%10), byte('0'+yy%10), ' ',
		byte('0'+hh/10), byte('0'+hh%10), ':',
		byte('0'+mn/10), byte('0'+mn%10), ':',
		byte('0'+ss/10), byte('0'+ss%10), ' ',
		'G', 'M', 'T')
}
