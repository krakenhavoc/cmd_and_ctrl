// The fx preview harness's own Vite config (#4868): `npm run fx-preview`.
//
// It is a separate root, so the app's `vite build` (client/vite.config.ts,
// input client/index.html) never reaches it and nothing here ships. It
// serves client/public (card backs, fonts, sounds) and draws a placeholder
// for every card image the harness asks for, so it needs no Go server and
// no Scryfall: table.ts gives each card the Scryfall ID `fx-<its name>`.

import { fileURLToPath } from "node:url";
import { defineConfig, type Plugin } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";

const client = fileURLToPath(new URL("../..", import.meta.url));

// A card's placeholder art: its name on a colour of its own, the back face
// darker, so a flip or a slide can be told apart at a glance.
function cardArt(name: string, face: number): string {
  let h = 0;
  for (const ch of name) h = (h * 31 + ch.codePointAt(0)!) % 360;
  const light = face > 0 ? 22 : 38;
  const esc = name.replace(/[&<>"]/g, (c) => `&#${c.charCodeAt(0)};`);
  // A line of about eleven characters reads at a tile's size.
  const lines: string[] = [];
  for (const w of esc.split(" ")) {
    const last = lines.length - 1;
    if (last >= 0 && lines[last].length + w.length < 11) lines[last] += ` ${w}`;
    else lines.push(w);
  }
  const text = lines.map((l, i) => `<tspan x="73" dy="${i === 0 ? 0 : 23}">${l}</tspan>`).join("");
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 146 204" width="146" height="204">
<rect width="146" height="204" rx="8" fill="hsl(${h} 45% ${light}%)"/>
<rect x="7" y="7" width="132" height="190" rx="5" fill="none" stroke="hsl(${h} 60% 70%)" stroke-width="2"/>
<text x="73" y="52" fill="#fff" font-family="sans-serif" font-size="20" font-weight="700" text-anchor="middle">${text}</text>
<text x="73" y="182" fill="hsl(${h} 60% 80%)" font-family="sans-serif" font-size="18" text-anchor="middle">${face > 0 ? "back face" : ""}</text>
</svg>`;
}

function placeholderArt(): Plugin {
  return {
    name: "fx-preview-card-art",
    configureServer(server) {
      server.middlewares.use((req, res, next) => {
        const url = new URL(req.url ?? "/", "http://localhost");
        const m = /^\/cards\/fx-([^/]+)\/image$/.exec(url.pathname);
        if (!m) return next();
        res.setHeader("Content-Type", "image/svg+xml");
        res.end(cardArt(decodeURIComponent(m[1]), Number(url.searchParams.get("face") ?? 0)));
      });
    },
  };
}

export default defineConfig({
  root: fileURLToPath(new URL(".", import.meta.url)),
  publicDir: `${client}public`,
  plugins: [svelte({ configFile: `${client}svelte.config.js` }), placeholderArt()],
  server: {
    // Beside the app's 5173; `npm run fx-preview -- --port N` picks another.
    port: 5174,
    fs: { allow: [client] },
  },
});
