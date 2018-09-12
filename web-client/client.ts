const setupDaDScrolling = (panel: HTMLElement) => {
  panel.onmousemove = (ev) => {
    if (ev.buttons) {
      panel.scrollBy(-ev.movementX, -ev.movementY);
    }
  }
}


const BomberClient = function(canvasId: string, playerName: string, raddr: string) {
  var canvas = document.getElementById(canvasId) as HTMLCanvasElement
  const scrollPanel = canvas.parentElement!
  setupDaDScrolling(scrollPanel)
  var ctx = canvas.getContext('2d')!

  let zoom = 1;
  const boardCache : { lastZoom: number | null; board: string[][] | null; players: string } = {
    lastZoom: null,
    board: null,
    players: ""
  }
  let packet : any = null
  const renderPlayers = (players: any[], colors: string[]) => {
    const serializedPlayers = JSON.stringify(players)
    if (serializedPlayers == boardCache.players) return;
    boardCache.players = serializedPlayers
    const box = document.getElementById("players") as HTMLUListElement
    while (box.firstChild) box.removeChild(box.firstChild)
    for (let i = 0; i < players.length; i++) {
      const p = players[i]
      const move =
        p.X == p.LastX && p.Y == p.LastY ? "◯" :
        p.X == p.LastX - 1 && p.Y == p.LastY ? "🡐 " :
        p.X == p.LastX && p.Y == p.LastY - 1 ? "🡑" :
        p.X == p.LastX + 1 && p.Y == p.LastY ? "🡒" :
        p.X == p.LastX && p.Y == p.LastY + 1 ? "🡓" :
        "";
      const li = document.createElement("li")
      // li.innerText += `#${i} - `
      const elName = document.createElement(p.Alive ? "span" : "strike")
      elName.classList.add("playerName")
      elName.style.textShadow = "1px 1px white"
      elName.style.backgroundColor = colors[i]
      elName.appendChild(document.createTextNode(p.Name))
      li.appendChild(elName)
      const elMove = document.createElement("span")
      elMove.classList.add("playerMove")
      elMove.innerText = move
      li.appendChild(elMove)
      li.appendChild(document.createElement("br"))
      li.appendChild(document.createTextNode(`${p.Points} bodů (💣 ${p.Bombs}/${p.MaxBomb}, 🔥 ${p.Radius})`))
      box.appendChild(li)
    }
  }
  const draw = () => {
    if (packet == null) return;
    if (packet.Board == null) return;
    const players: any[] = packet.Players
    players.forEach((p, i) => p.Index = i)
    players.sort((a, b) => b.Points - a.Points)
    const colors = players.map((p) => `hsl(${p.Index * (360 / players.length)}, 100%, 50%)`)
    const board = packet.Board
    const width = board.length
    const height = board[0].length
    const tileSize = 24*zoom;
    let redraw = false
    if (canvas.width != width * tileSize) {
      canvas.width = width * tileSize
      redraw = true
    }
    if (canvas.height != height * tileSize) {
      canvas.height = height * tileSize
      redraw = true
    }
    renderPlayers(players, colors)
    const tileDrawer = createDrawer(players.map((p, i) => ["P" + i, colors[i]]))

    if (boardCache.board == null) {
      boardCache.board = Array.from(new Array(height)).map(_ => Array.from(new Array(width)))
    }

    for (var i = height - 1; i >= 0; i--) {
      for (var j = width - 1; j >= 0; j--) {
        const cell = board[j][i];
        const rawName = cell.Name || cell;
        const name = rawName == "P" ?
                     rawName + players.findIndex(p => p.X == j && p.Y == i) :
                     rawName
        if (!redraw && boardCache.board[i][j] == name && boardCache.lastZoom == zoom && name != "Flame" && name != "F") {
          continue
        }
        boardCache.board[i][j] = name

        const x = Math.floor(j * tileSize)
        const y = Math.floor(i * tileSize)

        tileDrawer(ctx, name, x, y, Math.ceil(tileSize), Math.ceil(tileSize))
      }
    }
    boardCache.lastZoom = zoom
  }
  setInterval(() => {
    draw()
  }, 100)

  const that = {
    zoomIn() {
      zoom *= 2
      draw()
      scrollPanel.scrollLeft *= 2
      scrollPanel.scrollTop *= 2
    },
    zoomOut() {
      zoom *= 0.5
      draw()
      scrollPanel.scrollLeft *= 0.5
      scrollPanel.scrollTop *= 0.5
    }
  }
  draw()

  var endpoint = "ws://"+raddr+"";

  var updateSrv = WS(endpoint, playerName, function(e, conn) {
    var state = JSON.parse(e.data);
    packet = state
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
