import { createHash } from "node:crypto";
import { existsSync, readFileSync } from "node:fs";
import { join } from "node:path";
import type { Plugin } from "vite";

// Hand-rolled service-worker build step.
//
// Why not vite-plugin-pwa: adding it means adding it *and* its Workbox
// dependency tree to package-lock.json, and this repo's CI runs `npm ci`,
// which fails hard on a lockfile that was not produced by npm. The machine
// that wrote this branch has no Node toolchain, so a hand-edited lockfile
// entry was the alternative -- a worse risk than owning ~200 lines of
// service worker. The strategies below are also deliberately narrower than
// anything a generator would produce, because most of this app's traffic
// must never be cached at all. See ADR 0031.
//
// The plugin does three things at build time:
//   1. builds the precache list from the real bundle output plus a checked
//      list of static files in public/,
//   2. derives a build id from the contents of everything precached, so the
//      cache name changes exactly when the cached bytes change,
//   3. stamps both into src/sw/service-worker.js and emits it as /sw.js.

// Placeholders in src/sw/service-worker.js. The template keeps them as valid
// JavaScript literals so the file lints and unit-tests unmodified;
// serviceWorker.test.ts substitutes them the same way this plugin does.
export const BUILD_ID_TOKEN = '"__BUILD_ID__"';
export const PRECACHE_TOKEN = '["__PRECACHE__"]';

// Static files under public/ that are worth having offline. Everything here
// must exist or the build fails: a precache entry that 404s makes
// cache.addAll reject, which makes the service worker fail to install.
//
// Deliberately NOT here: public/sounds/*.mp3. That directory is 17MB, most of
// it one ambient track, and audio is worthless without a live game anyway.
const STATIC_PRECACHE = [
  "/offline.html",
  "/manifest.webmanifest",
  "/icons/favicon.svg",
  "/icons/icon-192.png",
  // Every face-down card in the game renders this, so it is the one image
  // worth having before the first request for it.
  "/card-back.jpg",
  "/card-back-small.jpg",
];

function sha256(data: string | Buffer): string {
  return createHash("sha256").update(data).digest("hex");
}

export function serviceWorkerPlugin(): Plugin {
  let publicDir = "";
  let templatePath = "";

  return {
    name: "cmdctrl:service-worker",
    // Dev never registers a worker (lib/pwa.ts is production-only), so there
    // is nothing to serve from `vite dev` and no stale-cache trap for
    // developers. Test the real thing with `npm run build && npm run preview`.
    apply: "build",
    // Run after Vite's own html/asset plugins so index.html and the hashed
    // chunks are already in the bundle.
    enforce: "post",

    configResolved(config) {
      publicDir = config.publicDir;
      templatePath = join(config.root, "src/sw/service-worker.js");
    },

    generateBundle(_options, bundle) {
      const template = readFileSync(templatePath, "utf8");
      for (const token of [BUILD_ID_TOKEN, PRECACHE_TOKEN]) {
        if (!template.includes(token)) {
          this.error(`src/sw/service-worker.js no longer contains the ${token} placeholder`);
        }
      }

      // 1. Hashed build output, straight from the bundle.
      const emitted = Object.keys(bundle)
        .filter((name) => /\.(js|css)$/.test(name) || name === "index.html")
        .map((name) => "/" + name)
        .sort();

      // This plugin runs `enforce: "post"`, so Vite's html plugin has already
      // emitted index.html. If that ever stops being true the shell would
      // quietly stop being precached and cold starts would silently go back
      // to the network -- fail the build instead.
      if (!emitted.includes("/index.html")) {
        this.error("index.html is not in the bundle; the service worker cannot precache the shell");
      }

      // 2. Static files, verified on disk and hashed into the build id.
      const staticDigests: string[] = [];
      for (const path of STATIC_PRECACHE) {
        const onDisk = join(publicDir, path);
        if (!existsSync(onDisk)) {
          this.error(
            `service worker precache lists ${path}, which does not exist in ${publicDir}. ` +
              `Add the file or drop it from STATIC_PRECACHE in vite-plugin-sw.ts.`,
          );
        }
        staticDigests.push(path + ":" + sha256(readFileSync(onDisk)));
      }

      // 3. Guard the manifest's own promises: a manifest that names a missing
      //    icon is silently un-installable, which is exactly the kind of
      //    failure nobody notices until a user tries to install the app.
      const manifestPath = join(publicDir, "manifest.webmanifest");
      const manifest = JSON.parse(readFileSync(manifestPath, "utf8")) as {
        icons?: { src: string }[];
      };
      for (const icon of manifest.icons ?? []) {
        if (!existsSync(join(publicDir, icon.src))) {
          this.error(
            `manifest.webmanifest references ${icon.src}, which does not exist in ${publicDir}. ` +
              `Run client/tools/gen-icons.sh.`,
          );
        }
      }

      const precache = [...emitted, ...STATIC_PRECACHE];

      // The build id covers the hashed filenames (which encode their own
      // contents), the static files' contents, and this worker's own source.
      // Rebuilding an unchanged tree reproduces the same id and therefore
      // keeps the warm cache; changing any precached byte invalidates it.
      const buildID = sha256([...emitted, ...staticDigests, template].join("\n")).slice(0, 12);

      const source = template
        .replace(BUILD_ID_TOKEN, JSON.stringify(buildID))
        .replace(PRECACHE_TOKEN, JSON.stringify(precache, null, 2));

      this.emitFile({ type: "asset", fileName: "sw.js", source });

      // Printed on purpose: the precache list is generated, so the build log
      // is the only place anyone can see what a given deploy will hold
      // offline -- and the build id is what a stale-cache report should be
      // matched against.
      console.log(
        `\n[sw] emitted /sw.js — build ${buildID}, ${precache.length} precached files:\n` +
          precache.map((path) => `     ${path}`).join("\n"),
      );
    },
  };
}
