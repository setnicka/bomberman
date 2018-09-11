#!/usr/bin/python3

# Required libraries: pip3 install websocket-client

import websocket
import sys
import random
import time
import json

directions = [
    ['up', 0, -1],
    ['down', 0, 1],
    ['left', -1, 0],
    ['right', 1, 0],
]
bomb_probability = 0.1


def on_message(ws, message):
    print("{}\tReceived message ({} bytes)".format(time.time(), len(message)))
    state = json.loads(message)
    if 'points_per_wall' in state:
        print("Received configuration: ")
        print(state)
        return

    (X, Y) = (state['X'], state['Y'])
    print(state)
    moves = []
    for (move, xChange, yChange) in directions:
        if state['Board'][X + xChange][Y + yChange] == ' ':
            moves.append(move)

    if random.random() < bomb_probability:
        move = "bomb"
    else:
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

    ws = websocket.WebSocketApp(
        "ws://localhost:" + port + "/",
        on_message=on_message,
        on_error=on_error,
        on_close=on_close,
        on_open=on_open
    )
    ws.run_forever()
