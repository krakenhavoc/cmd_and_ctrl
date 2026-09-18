import js from "@eslint/js";
import tseslint from "typescript-eslint";
import svelte from "eslint-plugin-svelte";
import globals from "globals";

export default [
  js.configs.recommended,
  ...tseslint.configs.recommended,
  ...svelte.configs["flat/recommended"],
  {
    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.node,
      },
    },
  },
  {
    // The service-worker template runs in a ServiceWorkerGlobalScope, not a
    // window: `clients`, `skipWaiting` and a self-typed `self` are globals
    // there and undefined everywhere else.
    files: ["src/sw/service-worker.js"],
    languageOptions: {
      globals: {
        ...globals.serviceworker,
      },
    },
  },
  {
    // `**/*.svelte.ts` as well as `**/*.svelte`: eslint-plugin-svelte 3
    // lints Svelte MODULES (runes outside a component) too, and
    // svelte-eslint-parser needs the TypeScript parser handed to it for
    // both. Without this, `src/lib/test/render.svelte.ts` fails to
    // parse on its first type import.
    files: ["**/*.svelte", "**/*.svelte.ts", "**/*.svelte.js"],
    languageOptions: {
      parserOptions: {
        parser: tseslint.parser,
      },
    },
  },
  {
    // Rules eslint-plugin-svelte 3 added that the existing client does
    // not satisfy. The upgrade itself was not optional: version 2's
    // parser cannot read `<svelte:boundary>` at all and failed on
    // Game.svelte with "Unknown type:SvelteBoundary", so the #720 error
    // boundary could not be linted, never mind merged.
    //
    // Turning these three off is a deliberate hold, not a verdict. Each
    // is a real finding in ~10 pre-existing files across the board
    // components, and clearing them is its own change with its own
    // reasoning — `prefer-svelte-reactivity` in particular wants a
    // judgement per site, because a plain Map built inside a
    // `$derived.by` is already reactive through its dependencies and
    // does not want a SvelteMap. Doing it here would bury the boundary
    // in unrelated churn and collide with everything else in flight in
    // client/.
    rules: {
      "svelte/prefer-svelte-reactivity": "off",
      "svelte/require-each-key": "off",
      "svelte/no-useless-mustaches": "off",
    },
  },
  {
    rules: {
      "@typescript-eslint/no-unused-vars": [
        "error",
        { argsIgnorePattern: "^_", varsIgnorePattern: "^_" },
      ],
    },
  },
  {
    // #720: no new unguarded stores. svelte/store's `subscriber_queue`
    // is module-GLOBAL and is never reset after a callback throws, so
    // ONE unguarded store anywhere in the app can stall every store in
    // it — including the guarded ones. That is #266, and it is not a
    // hazard a reviewer should have to remember.
    //
    // `get`, `Readable` and `Writable` are unaffected; only the three
    // constructors are.
    files: ["src/**/*.ts", "src/**/*.svelte", "src/**/*.svelte.ts"],
    ignores: ["src/lib/guardedStore.ts", "src/**/*.test.ts"],
    rules: {
      "no-restricted-imports": [
        "error",
        {
          paths: [
            {
              name: "svelte/store",
              importNames: ["writable", "derived", "readable"],
              message:
                "Use guardedWritable / guardedDerived from lib/guardedStore instead (#266 / #720): svelte/store's subscriber_queue is global, so one throwing subscriber on an unguarded store freezes every store in the app.",
            },
          ],
        },
      ],
    },
  },
  {
    ignores: ["dist/**", "node_modules/**", ".svelte-kit/**", ".vite/**"],
  },
];
