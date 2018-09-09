import websocket
import sys
import random
import time
import json

try:
    import thread
except ImportError:
    import _thread as thread

directions = [
        ['up', 0, -1],
        ['down', 0, 1],
        ['left', -1, 0],
        ['right', 1, 0],
]

def on_message(ws, message):
    print("{}\tReceived message ({} bytes)".format(time.time(), len(message)))
    state = json.loads(message)
    (X, Y) = (state['X'], state['Y'])
    # print(state)

    moves = []
    for (move, xChange, yChange) in directions:
        if state['Board'][X + xChange][Y + yChange] == ' ':
            moves.append(move)
    move = random.choice(moves)

    print("{}\tSending: {}".format(time.time(), move))
    ws.send(move)

def on_error(ws, error):
    print(error)

def on_close(ws):
    print("### closed ###")

def on_open(ws):
    ws.send(auth_string)


if __name__ == "__main__":
    port = sys.argv[1]
    auth_string = sys.argv[2]
    #websocket.enableTrace(True)
    ws = websocket.WebSocketApp(
        "ws://localhost:" + port + "/",
        on_message=on_message,
        on_error=on_error,
        on_close=on_close
    )
    ws.on_open = on_open
    ws.run_forever()
