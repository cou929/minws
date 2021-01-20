package http

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

var (
	crlf       = []byte("\r\n")
	colonSpace = []byte(": ")
)

// Server represents server to handle http
// processes handshake
type Server struct{}

// NewServer is constructor of Server
func NewServer() *Server {
	return &Server{}
}

// HandShake handles handshake process of websocket
func (srv *Server) HandShake(rwc net.Conn) {
	c := srv.newConn(rwc)
	defer func() {
		c.rwc.Close()
	}()
	c.r = bufio.NewReader(c.rwc)
	for {
		w, err := c.readRequest()
		if err != nil {
			if isCommonNetReadError(err) {
				return
			}
			log.Printf("readRequest error %v", err)
			return
		}
		srv.handleHandShake(w, w.req)
		w.finishRequest()
	}
}

func (srv *Server) newConn(rwc net.Conn) *conn {
	return &conn{
		server: srv,
		rwc:    rwc,
	}
}

func (srv *Server) handleHandShake(w ResponseWriter, req *Request) {
	return
}

// conn is server-side HTTP connection
type conn struct {
	server *Server
	rwc    net.Conn
	r      *bufio.Reader
}

func (c *conn) readRequest() (*response, error) {
	req := &Request{}

	// ex. `GET /index.html HTTP/1.0`
	firstLine, err := c.readLine()
	if err != nil {
		return nil, err
	}
	var ok bool
	req.Method, req.RequestURI, req.Proto, ok = parseRequestLine(firstLine)
	if !ok {
		return nil, fmt.Errorf("Invalid first line %s", firstLine)
	}
	if req.ProtoMajor, req.ProtoMinor, ok = parseHTTPVersion(req.Proto); !ok {
		return nil, fmt.Errorf("Invalid proto version %s", req.Proto)
	}

	// Todo: parse request uri

	req.Header, err = c.readHeader()
	if err != nil {
		return nil, err
	}

	req.Close = shouldClose(req.ProtoMajor, req.ProtoMinor, req.Header)

	// Todo: read body

	log.Println(req)

	res := &response{
		req: req,
		w:   bufio.NewWriter(c.rwc),
		c:   c,
	}

	return res, nil
}

func (c *conn) readLine() (string, error) {
	line, err := c.readLineByteSlice()
	if err != nil {
		return "", err
	}
	return string(line), nil
}

func (c *conn) readLineByteSlice() ([]byte, error) {
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

func (c *conn) readHeader() (Header, error) {
	res := make(Header)
	for {
		// Todo: read continued line
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
		res[key] = append(res[key], value)

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

func shouldClose(major, minor int, header Header) bool {
	// HTTP 0.9 has no keepalive
	if major < 1 {
		return true
	}
	conv := header["Connection"]
	hasClose := headerValueContainsToken(conv, "close")
	if major == 1 && minor == 0 {
		return hasClose || !headerValueContainsToken(conv, "keep-alive")
	}
	// HTTP/1.1+ are default keep-alive
	return hasClose
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

func isCommonNetReadError(err error) bool {
	if err == io.EOF { // read EOF
		return true
	}
	if neterr, ok := err.(net.Error); ok && neterr.Timeout() { // tcp timeout error
		return true
	}
	if oe, ok := err.(*net.OpError); ok && oe.Op == "read" { // tcp read error
		return true
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
	Close      bool
}

// Header represents HTTP header
type Header map[string][]string

func (h Header) get(key string) string {
	if v := h[key]; len(v) > 0 {
		return v[0]
	}
	return ""
}

func (h Header) has(key string) bool {
	_, ok := h[key]
	return ok
}

// ResponseWriter is writer for HTTP response
type ResponseWriter interface {
	Write(p []byte) (n int, err error)
	SetStatus(code int)
}

// Response represents HTTP Response
type response struct {
	req             *Request
	w               *bufio.Writer
	c               *conn
	wroteHeader     bool
	status          int
	header          Header
	contentLength   int
	closeAfterReply bool
}

func (r *response) Write(p []byte) (int, error) {
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

func (r *response) finishRequest() {
	if !r.wroteHeader {
		if r.status == 0 {
			r.SetStatus(StatusOK)
		}
		r.writeHeader(nil)
	}
	r.w.Flush()
}

func (r *response) SetStatus(code int) {
	r.status = code
}

func (r *response) writeHeader(p []byte) {
	if r.wroteHeader {
		return
	}
	r.wroteHeader = true

	// status line
	ver := "1.1"
	if r.req.ProtoMajor == 1 && r.req.ProtoMinor < 1 {
		ver = "1.0"
	}
	fmt.Fprintf(r.w, "HTTP/%s %d %s\r\n", ver, r.status, statusText[r.status])

	exh := extraHeader{}

	// content-type
	// todo: DetectContentType
	exh.contentType = "text/plain; charset=utf-8"

	// content-length
	r.contentLength = len(p)
	exh.contentLength = strconv.AppendInt(exh.contentLength, int64(len(p)), 10)

	// connection
	exh.connection = "keep-alive"
	if r.header.has("Connection") {
		exh.connection = r.header.get("Connection")
	}
	if r.closeAfterReply {
		exh.connection = "close"
	}

	// date
	if r.header.has("Date") {
		exh.date = []byte(r.header.get("Date"))
	} else {
		exh.date = appendTime(exh.date, time.Now())
	}

	// other headers
	for k, v := range r.header {
		fmt.Fprintf(r.w, "%s: %s\r\n", k, v[0])
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
