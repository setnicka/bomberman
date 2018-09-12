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

const PlayerDrawer = (color: string) => function(ctx: CanvasRenderingContext2D, name: string, x, y, maxX, maxY) {
    ctx.fillStyle = color;
    ctx.fillRect(x, y, maxX, maxY)
    ctx.font = "100px monospace"
    if (name[0].toLowerCase() == "p" && name.length > 1) {
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
    const img = document.getElementById(`img-power-radius`) as HTMLImageElement
    ctx.drawImage(img, x, y, maxX, maxY);
};

const RadiusPUDrawer = function(ctx, name, x, y, maxX, maxY) {
    const img = document.getElementById(`img-power-bombs`) as HTMLImageElement
    ctx.drawImage(img, x, y, maxX, maxY);
};

const createDrawer = (players: [string, string][]) => {
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
    }
    for (const p of players) {
        drawFunc[p[0]] = PlayerDrawer(p[1])
    }
    return (ctx, name, x, y, mx, my) => {
        if (name in drawFunc)
            return drawFunc[name](ctx, name, x, y, mx, my)
        else {
            console.error(`Can't draw ${name}`)
        }
    }
}
