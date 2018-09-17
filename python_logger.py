import websocket
import json


def on_message(ws, message):
    print(message)

def on_error(ws, error):
    print("Error: ", file=sys.stderr)


def on_close(ws):
    pass


auth_string = "public"
def on_open(ws):
    ws.send(auth_string)


ws = websocket.WebSocketApp(
    "ws://localhost:8000/",
    on_message=on_message,
    on_error=on_error,
    on_close=on_close,
    on_open=on_open
)
while True:
    ws.run_forever()
