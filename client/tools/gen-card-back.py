#!/usr/bin/env python3
"""Draw the card back the table shows for a face-down or hidden card.

Writes client/public/card-back.jpg (488x680) and card-back-small.jpg
(146x204). The back is our own: the C2 mark (the four Cs and the diamond
of src/lib/components/Wordmark.svelte and public/icons/icon.svg) on the
default skin's ink, in a black card border. It replaced a bundled scan of
Wizards of the Coast's card back, whose Magic: The Gathering logo has no
place on a site that says it is unofficial fan content (#2982).

Needs Pillow only. The icon's one path is flattened here because the
repo has no SVG rasteriser; if you edit the mark in icon.svg, mirror the
change in C_PATH below and re-run:

    python3 client/tools/gen-card-back.py
"""

import math
import os

from PIL import Image, ImageDraw, ImageFilter

INK = (7, 13, 20)
ORANGE = (255, 122, 26)
PAPER = (232, 239, 247)
BLACK = (10, 10, 12)

W, H = 488, 680
SS = 4  # supersampling factor


def arc_points(x0, y0, rx, ry, large, sweep, x1, y1, steps=24):
    """SVG endpoint-arc to a list of points (circles only: rx == ry)."""
    # Centre parameterisation for a circular arc (SVG 1.1 F.6.5).
    dx, dy = (x0 - x1) / 2, (y0 - y1) / 2
    r = rx
    d2 = dx * dx + dy * dy
    if d2 > r * r:
        r = math.sqrt(d2)
    k2 = max(0.0, (r * r - d2) / d2)
    k = math.sqrt(k2) * (-1 if large == sweep else 1)
    cxp, cyp = k * dy, -k * dx
    cx, cy = cxp + (x0 + x1) / 2, cyp + (y0 + y1) / 2
    a0 = math.atan2(y0 - cy, x0 - cx)
    a1 = math.atan2(y1 - cy, x1 - cx)
    if sweep and a1 < a0:
        a1 += 2 * math.pi
    if not sweep and a1 > a0:
        a1 -= 2 * math.pi
    return [
        (cx + r * math.cos(a0 + (a1 - a0) * i / steps), cy + r * math.sin(a0 + (a1 - a0) * i / steps))
        for i in range(1, steps + 1)
    ]


def c_polygon():
    """The icon's C, M-10-40V-80L-30-100H-52A48 48 0 0 0 -100-52 ... Z."""
    pts = [(-10, -40)]

    def line(x, y):
        pts.append((x, y))

    def arc(r, sweep, x, y):
        px, py = pts[-1]
        pts.extend(arc_points(px, py, r, r, 0, sweep, x, y))

    line(-10, -80)
    line(-30, -100)
    line(-52, -100)
    arc(48, 0, -100, -52)
    line(-100, -30)
    line(-80, -10)
    line(-40, -10)
    line(-33, -17)
    line(-48, -32)
    line(-71, -32)
    arc(7, 1, -78, -39)
    line(-78, -52)
    arc(26, 1, -52, -78)
    line(-39, -78)
    arc(7, 1, -32, -71)
    line(-32, -48)
    line(-17, -33)
    return pts


def render() -> Image.Image:
    w, h = W * SS, H * SS
    img = Image.new("RGB", (w, h), BLACK)
    d = ImageDraw.Draw(img)

    # The field, inside a black card border like a printed card's.
    border = 22 * SS
    d.rounded_rectangle((border, border, w - border, h - border), radius=14 * SS, fill=INK)
    # A hairline in the accent, inset from the field's edge.
    inset = border + 12 * SS
    d.rounded_rectangle(
        (inset, inset, w - inset, h - inset), radius=8 * SS, outline=ORANGE, width=2 * SS
    )

    # The pool of light from the icon, behind the mark.
    glow = Image.new("RGB", (w, h), INK)
    gd = ImageDraw.Draw(glow)
    cx, cy = w // 2, h // 2
    gd.ellipse((cx - 190 * SS, cy - 190 * SS, cx + 190 * SS, cy + 190 * SS), fill=(40, 26, 14))
    glow = glow.filter(ImageFilter.GaussianBlur(70 * SS))
    mask = Image.new("L", (w, h), 0)
    ImageDraw.Draw(mask).rounded_rectangle(
        (inset + 4 * SS, inset + 4 * SS, w - inset - 4 * SS, h - inset - 4 * SS),
        radius=6 * SS,
        fill=255,
    )
    img.paste(glow, (0, 0), mask)
    d = ImageDraw.Draw(img)

    # The mark: four Cs and a diamond, on the icon's 208-unit grid.
    scale = 1.3 * SS
    poly = c_polygon()
    for sx, sy in ((1, 1), (-1, 1), (1, -1), (-1, -1)):
        d.polygon([(cx + sx * x * scale, cy + sy * y * scale) for x, y in poly], fill=ORANGE)
    diamond = [(0, -38), (38, 0), (0, 38), (-38, 0)]
    d.polygon([(cx + x * scale, cy + y * scale) for x, y in diamond], fill=PAPER)

    return img.resize((W, H), Image.LANCZOS)


def main() -> None:
    out = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "public")
    full = render()
    full.save(os.path.join(out, "card-back.jpg"), "JPEG", quality=88, optimize=True, progressive=True)
    small = full.resize((146, 204), Image.LANCZOS)
    small.save(os.path.join(out, "card-back-small.jpg"), "JPEG", quality=88, optimize=True)
    print("wrote card-back.jpg and card-back-small.jpg")


if __name__ == "__main__":
    main()
