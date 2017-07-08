
/* ------------   Connexion au serveur et attente de message  ---------------- */
function connect(){
    var socket = null;
    try {
        socket = new WebSocket("ws://localhost:8080/ws");

    } catch (exception) {
        console.error(exception);
    }

    socket.onerror = function(error) {
        console.error(error);
    };

    // Lorsque la connexion est établie.
    socket.onopen = function(event) {
        console.log("Connected.");
        var t = d3.select("body")
          .append("p")
          .text("Connected");

        this.onclose = function(event) {
        console.log("Disconnected.");
        };

        this.onmessage = function(event) {
            console.log(event)
            console.log(event.data);
            var data = JSON.parse(event.data);
            console.log(data);
            var t = d3.select("body")
              .append("p")
              .text(data.Name);

           // ConvertJSON(event);
            // var t = d3.select("body")
            //   .append("p")
            //   .text("message");
        };
        // Envoi d'un message vers le serveur.
        //this.send("Hello world!");
    };
}
connect();


/* ----------------------------    Message reçu ----------------------------    */
function ConvertJSON(message) {


}
