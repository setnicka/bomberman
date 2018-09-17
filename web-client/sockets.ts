interface BombermanWS {
    send(event: string): void
}

const WS = function(url: string, token: string, receiveFn: (msg: MessageEvent, conn: BombermanWS) => void) {
    var closed = false;
    const out = {
        sckt: new WebSocket(url),
        send(event: string) {
            out.sckt.send(event)
        },
        close() {
            closed = true;
            out.sckt.close();
        }
    }

    const init = function() {
        out.sckt.onmessage = (msg) => receiveFn(msg, out)

        out.sckt.onopen = function (event) {
            out.sckt.send(token)
            console.log("Connected" + event);
        }


        let tryAgainTimeout = 0;
        const tryAgain = (event: Event) => {
            if (closed) return;
            if (!tryAgainTimeout)
                tryAgainTimeout = setTimeout(() => {
                    tryAgainTimeout = 0;
                    out.sckt = new WebSocket(url);
                    init();
                }, 400);
        }

        out.sckt.onclose = tryAgain
        // out.sckt.onerror = tryAgain
    }
    init()
    return out
}
