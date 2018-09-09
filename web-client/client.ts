const setupDaDScrolling = (panel: HTMLElement) => {
  panel.onmousemove = (ev) => {
    if (ev.buttons) {
      panel.scrollBy(-ev.movementX, -ev.movementY);
    }
  }
}


const BomberClient = function(canvasId, playerName, raddr) {
  // Get my canvas yo!
  var canvas = document.getElementById(canvasId) as HTMLCanvasElement
  const scrollPanel = canvas.parentElement!
  setupDaDScrolling(scrollPanel)
  var ctx = canvas.getContext('2d')!

  let zoom = 1;
  const boardCache : { lastZoom: number | null; board: string[][] | null } = {
    lastZoom: null,
    board: null
  }
  let packet : any = null
  const draw = () => {
    if (packet == null) return;
    const board = packet.Board
    const height = board.length
    const width = board[0].length
    const tileSize = 50*zoom;
    let redraw = false
    if (canvas.width != width * tileSize) {
      canvas.width = width * tileSize
      redraw = true
    }
    if (canvas.height != height * tileSize) {
      canvas.height = height * tileSize
      redraw = true
    }
    const tileDrawer = createDrawer(["p1", "p2", "p3", "p4", "P1", "P2", "P3", "P4"])

    if (boardCache.board == null) {
      boardCache.board = Array.from(new Array(height)).map(_ => Array.from(new Array(width)))
    }

    for (var i = height - 1; i >= 0; i--) {
      for (var j = width - 1; j >= 0; j--) {
        const cell = board[i][j];
        const name = cell.Name;
        if (!redraw && boardCache.board[i][j] == name && boardCache.lastZoom == zoom && name != "Flame") {
          continue
        }
        boardCache.board[i][j] = name

        const x = Math.floor(j * tileSize);
        const y = Math.floor(i * tileSize);

        tileDrawer(ctx, name, x, y, Math.ceil(tileSize), Math.ceil(tileSize));
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
