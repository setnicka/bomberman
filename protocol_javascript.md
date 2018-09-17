## Implementace v Javascriptu

Nějak takto se dá navázat komunikace se serverem (v prohlížeči)

```javascript

let connection;
function init() {
    connection = new WebSocket("ws://server:8000");

    connection.onopen = function (event) {
        connection.send("player:password")
        console.log("Connected");
    }


    connection.onclose = (event) => {
        setTimeout(() => init(), 400);
    }
}

init();

connection.onmessage = (message) => {
    const state = JSON.parse(message.data);
    if ('points_per_wall' in state) {
        // první zpráva obsahuje konfiguraci, ne stav hry
        console.log("Konfigurace hry:", state);
        return;
    }

    // naše souřadnice
    const X = state.X,
          Y = state.Y;
    // toto je políčko na pravo od nás
    const polickoVPravo = state.Board[X + 1][Y];
    console.log(polickoVPravo);
    // connection.send("up")
    // connection.send("down")
    // connection.send("right")
    // connection.send("left")
    // connection.send("bomb")
}
```
