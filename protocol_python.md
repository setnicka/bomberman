## Implementace v Pythonu

je potřeba stáhnout knihovnu na websocket - `pip3 install websocket-client`. Pak se používá nějak takto:

```python
import websocket
import json


def on_message(ws, message):
    state = json.loads(message)
    if 'points_per_wall' in state:
        # první zpráva obsahuje konfiguraci, ne stav hry
        print("Konfigurace hry: ")
        print(state)
        return

    # naše souřadnice
    (X, Y) = (state['X'], state['Y'])
    # toto je políčko na pravo od nás
    policko_v_pravo = state['Board'][X + 1][Y]
    # ws.send("up")
    # ws.send("down")
    # ws.send("right")
    # ws.send("left")
    # ws.send("bomb")


def on_error(ws, error):
    print(error)


def on_close(ws):
    print("### closed ###")


auth_string = "jmeno:heslo"
def on_open(ws):
    ws.send(auth_string)



ws = websocket.WebSocketApp(
    "ws://server:8000/",
    on_message=on_message,
    on_error=on_error,
    on_close=on_close,
    on_open=on_open
)
ws.run_forever()
```