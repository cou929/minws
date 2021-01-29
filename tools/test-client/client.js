var connection = null;

function connect() {
    const port = 5001;
    const serverUrl = 'ws://' + document.location.hostname + ':' + port;
    connection = new WebSocket(serverUrl, "json");
    console.log("***CREATED WEBSOCKET");

    connection.onopen = function (evt) {
        console.log("***ONOPEN");
    };
    console.log("***CREATED ONOPEN");

    connection.onmessage = function (evt) {
        console.log("***ONMESSAGE");
        document.getElementById("debug").innerHTML = evt.data;
    }

    console.log("***CREATED ONMESSAGE");
}

function send() {
    console.log("***SEND");
    const msg = {
        text: document.getElementById('text').value,
        type: "message",
        date: Date.now()
    };
    let payload = document.getElementById('useBinary').checked ? new Blob([msg]) : JSON.stringify(msg);
    connection.send(payload);
    document.getElementById("debug").innerHTML = "";
}
