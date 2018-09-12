var Keyboard = function () {
    var that = {};
    that.W = 87,
        that.A = 65,
        that.S = 83,
        that.D = 68,
        that.Up = 38,
        that.Down = 40,
        that.Left = 37,
        that.Right = 39,
        that.Space = 32;
    var mapping = {};
    var handler = function (event) { };
    document.onkeydown = function (ev) {
        if (mapping[ev.keyCode] != null) {
            handler(mapping[ev.keyCode]);
            return false;
        }
        return true;
    };
    that.map = function (key, val) {
        mapping[key] = val;
        return that;
    };
    that.handler = function (f) {
        handler = f;
    };
    return that;
};
const WS = function (url, token, receiveFn) {
    const out = {
        sckt: new WebSocket(url),
        send(event) {
            out.sckt.send(event);
        }
    };
    const init = function () {
        out.sckt.onmessage = (msg) => receiveFn(msg, out);
        out.sckt.onopen = function (event) {
            out.sckt.send(token);
            console.log("Connected" + event);
        };
        let tryAgainTimeout = 0;
        const tryAgain = (event) => {
            if (!tryAgainTimeout)
                tryAgainTimeout = setTimeout(() => {
                    tryAgainTimeout = 0;
                    out.sckt = new WebSocket(url);
                    init();
                }, 400);
        };
        out.sckt.onclose = tryAgain;
        // out.sckt.onerror = tryAgain
    };
    init();
    return out;
};
const WallDrawer = function (ctx, name, x, y, maxX, maxY) {
    ctx.fillStyle = "black";
    ctx.fillRect(x, y, maxX, maxY);
};
const GroundDrawer = function (ctx, name, x, y, maxX, maxY) {
    ctx.fillStyle = "white";
    ctx.fillRect(x, y, maxX, maxY);
};
const RockDrawer = function (ctx, name, x, y, maxX, maxY) {
    const img = document.getElementById(`img-wall`);
    ctx.drawImage(img, x, y, maxX, maxY);
};
const BombDrawer = function (ctx, name, x, y, maxX, maxY) {
    const img = document.getElementById(`img-bomb`);
    ctx.drawImage(img, x, y, maxX, maxY);
};
const FlameDrawer = function (ctx, name, x, y, maxX, maxY) {
    const imgId = Math.ceil(Math.random() * 4);
    const img = document.getElementById(`img-flame${imgId}`);
    ctx.drawImage(img, x, y, maxX, maxY);
};
const PlayerDrawer = (color) => function (ctx, name, x, y, maxX, maxY) {
    ctx.fillStyle = color;
    ctx.fillRect(x, y, maxX, maxY);
    ctx.font = "100px monospace";
    if (name[0].toLowerCase() == "p" && name.length > 1) {
        name = name.substr(1);
    }
    const measuredFont = ctx.measureText(name);
    ctx.font = `${Math.min(maxY, maxX / (measuredFont.width / 100))}px monospace`;
    ctx.fillStyle = "black";
    const measuredFont2 = ctx.measureText(name);
    const freeSpace = maxX - measuredFont2.width;
    ctx.fillText(name, x + freeSpace / 2, y + (maxY * 0.8));
};
const BombPUDrawer = function (ctx, name, x, y, maxX, maxY) {
    const img = document.getElementById(`img-power-radius`);
    ctx.drawImage(img, x, y, maxX, maxY);
};
const RadiusPUDrawer = function (ctx, name, x, y, maxX, maxY) {
    const img = document.getElementById(`img-power-bombs`);
    ctx.drawImage(img, x, y, maxX, maxY);
};
const createDrawer = (players) => {
    const drawFunc = {
        "Wall": WallDrawer,
        "#": WallDrawer,
        "Ground": GroundDrawer,
        " ": GroundDrawer,
        "Rock": RockDrawer,
        ".": RockDrawer,
        "Bomb": BombDrawer,
        "B": BombDrawer,
        "Flame": FlameDrawer,
        "F": FlameDrawer,
        "PowerUp(Bomb)": BombPUDrawer,
        "n": BombPUDrawer,
        "PowerUp(Radius)": RadiusPUDrawer,
        "r": RadiusPUDrawer,
    };
    for (const p of players) {
        drawFunc[p[0]] = PlayerDrawer(p[1]);
    }
    return (ctx, name, x, y, mx, my) => {
        if (name in drawFunc)
            return drawFunc[name](ctx, name, x, y, mx, my);
        else {
            console.error(`Can't draw ${name}`);
        }
    };
};
const setupDaDScrolling = (panel) => {
    panel.onmousemove = (ev) => {
        if (ev.buttons) {
            panel.scrollBy(-ev.movementX, -ev.movementY);
        }
    };
};
const BomberClient = function (canvasId, playerName, raddr) {
    var canvas = document.getElementById(canvasId);
    const scrollPanel = canvas.parentElement;
    setupDaDScrolling(scrollPanel);
    var ctx = canvas.getContext('2d');
    let zoom = 1;
    const boardCache = {
        lastZoom: null,
        board: null,
        players: ""
    };
    let packet = null;
    const renderPlayers = (players, colors) => {
        const serializedPlayers = JSON.stringify(players);
        if (serializedPlayers == boardCache.players)
            return;
        boardCache.players = serializedPlayers;
        const box = document.getElementById("players");
        while (box.firstChild)
            box.removeChild(box.firstChild);
        for (let i = 0; i < players.length; i++) {
            const p = players[i];
            const move = p.X == p.LastX && p.Y == p.LastY ? "◯" :
                p.X == p.LastX - 1 && p.Y == p.LastY ? "🡐 " :
                    p.X == p.LastX && p.Y == p.LastY - 1 ? "🡑" :
                        p.X == p.LastX + 1 && p.Y == p.LastY ? "🡒" :
                            p.X == p.LastX && p.Y == p.LastY + 1 ? "🡓" :
                                "";
            const li = document.createElement("li");
            // li.innerText += `#${i} - `
            const elName = document.createElement(p.Alive ? "span" : "strike");
            elName.classList.add("playerName");
            elName.style.textShadow = "1px 1px white";
            elName.style.backgroundColor = colors[i];
            elName.appendChild(document.createTextNode(p.Name));
            li.appendChild(elName);
            const elMove = document.createElement("span");
            elMove.classList.add("playerMove");
            elMove.innerText = move;
            li.appendChild(elMove);
            li.appendChild(document.createElement("br"));
            li.appendChild(document.createTextNode(`${p.Points} bodů (💣 ${p.Bombs}/${p.MaxBomb}, 🔥 ${p.Radius})`));
            box.appendChild(li);
        }
    };
    const draw = () => {
        if (packet == null)
            return;
        if (packet.Board == null)
            return;
        const players = packet.Players;
        players.forEach((p, i) => p.Index = i);
        players.sort((a, b) => b.Points - a.Points);
        const colors = players.map((p) => `hsl(${p.Index * (360 / players.length)}, 100%, 50%)`);
        const board = packet.Board;
        const width = board.length;
        const height = board[0].length;
        const tileSize = 24 * zoom;
        let redraw = false;
        if (canvas.width != width * tileSize) {
            canvas.width = width * tileSize;
            redraw = true;
        }
        if (canvas.height != height * tileSize) {
            canvas.height = height * tileSize;
            redraw = true;
        }
        renderPlayers(players, colors);
        const tileDrawer = createDrawer(players.map((p, i) => ["P" + i, colors[i]]));
        if (boardCache.board == null) {
            boardCache.board = Array.from(new Array(height)).map(_ => Array.from(new Array(width)));
        }
        for (var i = height - 1; i >= 0; i--) {
            for (var j = width - 1; j >= 0; j--) {
                const cell = board[j][i];
                const rawName = cell.Name || cell;
                const name = rawName == "P" ?
                    rawName + players.findIndex(p => p.X == j && p.Y == i) :
                    rawName;
                if (!redraw && boardCache.board[i][j] == name && boardCache.lastZoom == zoom && name != "Flame" && name != "F") {
                    continue;
                }
                boardCache.board[i][j] = name;
                const x = Math.floor(j * tileSize);
                const y = Math.floor(i * tileSize);
                tileDrawer(ctx, name, x, y, Math.ceil(tileSize), Math.ceil(tileSize));
            }
        }
        boardCache.lastZoom = zoom;
    };
    setInterval(() => {
        draw();
    }, 100);
    const that = {
        zoomIn() {
            zoom *= 2;
            draw();
            scrollPanel.scrollLeft *= 2;
            scrollPanel.scrollTop *= 2;
        },
        zoomOut() {
            zoom *= 0.5;
            draw();
            scrollPanel.scrollLeft *= 0.5;
            scrollPanel.scrollTop *= 0.5;
        }
    };
    draw();
    var endpoint = "ws://" + raddr + "";
    var updateSrv = WS(endpoint, playerName, function (e, conn) {
        var state = JSON.parse(e.data);
        packet = state;
        draw();
    });
    // var moveSrv = WS(endpoint+"/move");
    // var kb = Keyboard();
    // kb.map(kb.Up, "up")
    //   .map(kb.Down, "down")
    //   .map(kb.Left, "left")
    //   .map(kb.Right, "right")
    //   .map(kb.Space, "bomb")
    //   .handler(moveSrv.send);
    return that;
};
//# sourceMappingURL=script.js.map