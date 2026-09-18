import { defineConfig } from "vitest/config";
import { svelte } from "@sveltejs/vite-plugin-svelte";

// The test pipeline is deliberately its own config rather than a `test`
// block bolted onto vite.config.ts (#689). Vitest prefers this file when
// both exist, so vite.config.ts stays purely about running the app: the
// dev proxy table and the build-only service-worker plugin have no place
// in a test run, and `resolve.conditions` below must not leak into a
// production build.
//
// Two things here are load-bearing, and both are the answer to #624's
// failed attempt at a component render test:
//
//  1. `resolve.conditions: ["browser"]`. Vitest picks a transform mode
//     from the environment — `ssr` for `node`, `web` for `jsdom` — and
//     under the ssr pipeline the `svelte` package resolves to its server
//     export, which compiles components for string rendering and has no
//     `mount`. Asking for the browser condition is what makes a real
//     client-side mount possible. Because the mode is per-file, this
//     only applies to the files that opt into jsdom; the ~890 pure-logic
//     tests keep resolving exactly as they did before.
//
//  2. `environment: "node"` stays the default. A DOM environment is
//     opt-in per file with a `// @vitest-environment jsdom` docblock at
//     the top of the file. Flipping the whole suite to jsdom would give
//     every pure-logic test a `window`, `localStorage` and a `document`
//     it never asked for — and several of those tests exist precisely
//     to pin the no-`localStorage` branch (settings.ts, session.ts).
//     Making them run in a browser-shaped world would quietly stop
//     testing that.
//
// jsdom over happy-dom: the behaviour that can only be tested by
// rendering is roles and ARIA, focus, and event propagation, and jsdom
// is the more spec-complete implementation of those three in
// particular. happy-dom's advantage is start-up speed, which buys
// nothing when the environment is opt-in for a handful of files.
export default defineConfig({
  plugins: [
    svelte({
      // This is the blocker #689 called out — #624's render test died
      // here. svelte.config.js uses `vitePreprocess()`, whose STYLE
      // half hands each `<style>` block to Vite's own `preprocessCSS`.
      // Under Vitest that call throws "Cannot create proxy with a
      // non-object as target or handler": Vite 6's `preprocessCSS`
      // builds a `PartialEnvironment` from `config.environments`, and
      // the config Vitest resolves has no `environments` entry for it
      // to read. It fails during preprocessing, so the component never
      // compiles and the whole file reports "0 test".
      //
      // The style half is only needed for `<style lang="scss">` and
      // friends; every component here writes plain CSS, which Svelte's
      // own compiler scopes without help. And `vitePreprocess()` with
      // no arguments adds NO script preprocessor — that half is opt-in
      // (`{ script: true }`) — so the production build already relies
      // on Svelte 5 stripping `lang="ts"` itself. Which makes the whole
      // preprocessor a no-op for tests, and an empty list the honest
      // spelling of that.
      //
      // It has to be an ARRAY, not `vitePreprocess({ style: false })`:
      // the plugin deep-merges svelte.config.js under the inline
      // options, and merging two preprocessor OBJECTS would let
      // svelte.config.js's `style` key survive. Arrays replace.
      //
      // Should a component ever adopt a CSS preprocessor, this is the
      // line that has to change.
      preprocess: [],
      // Inject each component's CSS at runtime instead of emitting a
      // separate virtual CSS module. Vitest does not process CSS, and
      // an emitted module it stubs out is one more moving part for no
      // gain — a render test asserts about roles and text, never about
      // computed styles.
      emitCss: false,
    }),
  ],
  resolve: { conditions: ["browser"] },
  test: {
    environment: "node",
    include: ["src/**/*.test.ts"],
  },
});
