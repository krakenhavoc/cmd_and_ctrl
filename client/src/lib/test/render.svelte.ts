// render.svelte.ts — the client's component render harness (#689).
//
// Mounts a real Svelte 5 component into a real DOM so a test can ask
// about the things that only exist once markup is rendered: roles and
// ARIA, focus, event propagation, conditional rendering. Everything
// that can be a pure function should stay one (see client/README.md);
// this is for the rest.
//
// Svelte's own `mount` / `unmount` / `flushSync` rather than
// `@testing-library/svelte`: `mount` is the API the app itself boots
// with (`src/main.ts`), so a render test exercises the same entry point
// as production, and it adds no third-party peer-dependency matrix to
// keep in step with vite-plugin-svelte — which is what stopped #624
// from getting a render test in the first place.
//
// The file is `.svelte.ts`, not `.ts`, because it needs `$state`:
// `mount` only tracks reads of props that live on a reactive proxy, so
// a plain object passed as `props` is frozen for the life of the mount
// and `setProps` would be a lie. vite-plugin-svelte compiles runes in
// `.svelte.ts` modules; a plain `.ts` file would ship `$state` through
// untouched.
//
// Requires a DOM: every test file that imports this must carry
// `// @vitest-environment jsdom` in a docblock at the top.

import { flushSync, mount, unmount, type Component } from "svelte";

/** A mounted component, and the handles a test needs on it. */
export interface Rendered<P extends Record<string, unknown>> {
  /** The element the component was mounted into. */
  container: HTMLElement;
  /**
   * The live props proxy. Mutating a field re-renders; prefer
   * `setProps` so the flush is not forgotten.
   */
  props: P;
  /** Merge new prop values in and flush the resulting update. */
  setProps(next: Partial<P>): void;
  /** Unmount and detach the container. */
  destroy(): void;
}

// Everything this module has mounted and not yet torn down, newest
// last, so `cleanup()` in an afterEach can reach mounts a failing test
// abandoned. A leaked mount is not harmless: its `$effect`s keep
// running against the next test's jsdom.
const live: Array<{ destroy(): void }> = [];

/**
 * render mounts `component` with `props` into a fresh container
 * appended to `document.body`, flushes the first render, and hands
 * back the container plus a reactive prop setter.
 */
export function render<P extends Record<string, unknown>>(
  component: Component<P, Record<string, never>>,
  props: P,
): Rendered<P> {
  const container = document.createElement("div");
  document.body.appendChild(container);

  // A shallow spread is enough: `$state` proxies lazily on read, so
  // nested objects become reactive when the component reaches them.
  const reactive = $state({ ...props }) as P;
  const instance = mount(component, { target: container, props: reactive });
  flushSync();

  const handle: Rendered<P> = {
    container,
    props: reactive,
    setProps(next: Partial<P>): void {
      Object.assign(reactive, next);
      flushSync();
    },
    destroy(): void {
      const at = live.indexOf(handle);
      if (at !== -1) live.splice(at, 1);
      void unmount(instance, { outro: false });
      container.remove();
    },
  };
  live.push(handle);
  return handle;
}

/**
 * cleanup unmounts everything still mounted. Call it from an
 * `afterEach` so one failing assertion cannot leak effects into the
 * next test.
 */
export function cleanup(): void {
  while (live.length > 0) live[live.length - 1].destroy();
  document.body.innerHTML = "";
}

/**
 * click dispatches a bubbling, cancellable click the way a real
 * pointer does. `el.click()` also bubbles, but takes no modifier keys
 * and is easy to confuse with a direct handler call; this keeps every
 * test on one path through the DOM's own dispatch, which is the point
 * when what is under test is propagation.
 */
export function click(el: Element, init: MouseEventInit = {}): void {
  el.dispatchEvent(new MouseEvent("click", { bubbles: true, cancelable: true, ...init }));
  flushSync();
}

/** Re-exported so tests can flush after mutating state directly. */
export { flushSync };
