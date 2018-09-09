const WallDrawer = function(ctx, name, x, y, maxX, maxY) {
    ctx.fillStyle = "black";
    ctx.fillRect(x, y, maxX, maxY)
};

const GroundDrawer = function(ctx, name, x, y, maxX, maxY) {
    ctx.fillStyle = "white";
    ctx.fillRect(x, y, maxX, maxY)
};

const RockDrawer = function(ctx, name, x, y, maxX, maxY) {
    const img = document.getElementById(`img-wall`) as HTMLImageElement
    ctx.drawImage(img, x, y, maxX, maxY);
};

const BombDrawer = function(ctx, name, x, y, maxX, maxY) {
    const img = document.getElementById(`img-bomb`) as HTMLImageElement
    ctx.drawImage(img, x, y, maxX, maxY);
};

const FlameDrawer = function(ctx : CanvasRenderingContext2D, name, x, y, maxX, maxY) {
    const imgId = Math.ceil(Math.random() * 4);
    const img = document.getElementById(`img-flame${imgId}`) as HTMLImageElement
    ctx.drawImage(img, x, y, maxX, maxY);
};

const PlayerDrawer = function(ctx: CanvasRenderingContext2D, name, x, y, maxX, maxY) {
    ctx.fillStyle = "cyan";
    ctx.fillRect(x, y, maxX, maxY)
    ctx.font = "100px monospace"
    if (name[0] = "p") {
        name = name.substr(1);
    }
    const measuredFont = ctx.measureText(name)
    ctx.font = `${Math.min(maxY, maxX / (measuredFont.width / 100))}px monospace`
    ctx.fillStyle = "black"
    const measuredFont2 = ctx.measureText(name)
    const freeSpace = maxX - measuredFont2.width
    ctx.fillText(name, x + freeSpace / 2, y + (maxY * 0.8));
};

const BombPUDrawer = function(ctx, name, x, y, maxX, maxY) {
    ctx.fillStyle = "green";
    ctx.fillRect(x, y, maxX, maxY)
};

const RadiusPUDrawer = function(ctx, name, x, y, maxX, maxY) {
    ctx.fillStyle = "pink";
    ctx.fillRect(x, y, maxX, maxY)
};

const createDrawer = (players: string[]) => {
    const drawFunc = {
        "Wall": WallDrawer,
        "Ground": GroundDrawer,
        "Rock": RockDrawer,
        "Bomb": BombDrawer,
        "Flame": FlameDrawer,
        "PowerUp(Bomb)": BombPUDrawer,
        "PowerUp(Radius)": RadiusPUDrawer,
    }
    for (const p of players) {
        drawFunc[p] = PlayerDrawer
    }
    return (ctx, name, x, y, mx, my) => {
        return drawFunc[name](ctx, name, x, y, mx, my)
    }
}