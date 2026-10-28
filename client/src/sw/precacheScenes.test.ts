// ADR 0209 §4 (#4397): signature scenes load the first time their card
// appears in a game, not when the app starts, so the service worker's
// precache leaves their chunks out. vite.config.ts names them with
// chunkFileName; the plugin's precache list skips what precached rejects.

import { describe, expect, it } from "vitest";

import { LAZY_CHUNK_PREFIX, chunkFileName, precached } from "../../vite-plugin-sw";

describe("signature scene chunks", () => {
  it("a chunk made only of fx/scenes modules is named as a scene's", () => {
    expect(
      chunkFileName({
        moduleIds: ["/repo/client/src/lib/fx/scenes/blasphemous_act.ts"],
      }),
    ).toBe(`${LAZY_CHUNK_PREFIX}[name]-[hash].js`);
    expect(
      chunkFileName({
        moduleIds: ["C:\\repo\\client\\src\\lib\\fx\\scenes\\kit.ts"],
      }),
    ).toBe(`${LAZY_CHUNK_PREFIX}[name]-[hash].js`);
  });

  it("any other chunk keeps Vite's own name", () => {
    expect(chunkFileName({ moduleIds: ["/repo/client/src/lib/boardShot.ts"] })).toBe(
      "assets/[name]-[hash].js",
    );
    expect(
      chunkFileName({
        moduleIds: ["/repo/client/src/lib/fx/scenes/kit.ts", "/repo/client/src/lib/motion.ts"],
      }),
    ).toBe("assets/[name]-[hash].js");
    expect(chunkFileName({ moduleIds: [] })).toBe("assets/[name]-[hash].js");
  });

  it("the precache takes the shell and every other chunk, not a scene's", () => {
    expect(precached("index.html")).toBe(true);
    expect(precached("assets/index-abc.js")).toBe(true);
    expect(precached("assets/index-abc.css")).toBe(true);
    expect(precached(`${LAZY_CHUNK_PREFIX}toxic_deluge-abc.js`)).toBe(false);
    expect(precached("assets/logo.png")).toBe(false);
  });
});
