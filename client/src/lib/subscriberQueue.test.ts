import { describe, it, expect } from "vitest";
import { writable } from "svelte/store";

// A tripwire on the upstream bug the whole of #266 / #720 exists to
// contain, and the only place in the suite that deliberately triggers
// it.
//
// svelte/store keeps ONE module-global `subscriber_queue` and drains it
// like this:
//
//     for (let i = 0; i < subscriber_queue.length; i += 2) {
//       subscriber_queue[i][0](subscriber_queue[i + 1]);
//     }
//     subscriber_queue.length = 0;
//
// A callback that throws skips the remaining callbacks AND the reset.
// The queue is then non-empty forever, so every later `set()` on EVERY
// store enqueues and returns without flushing — one throw anywhere
// freezes the whole UI for the life of the page.
//
// ONE test, in its own file, and last-resort by design: poisoning the
// queue is irreversible within a module registry, so anything that ran
// after it in this file would be testing a broken svelte/store. Vitest
// isolates module state per file, which is what keeps the damage here.
//
// IF THIS TEST FAILS, that is good news: svelte/store has started
// clearing the queue on a throw, the amplifier is gone upstream, and
// guardedStore.ts's reason for existing should be re-examined against
// the installed version rather than assumed. Do not "fix" it by
// loosening the assertion.

describe("svelte/store's global subscriber_queue (upstream, unfixed)", () => {
  it("stops notifying every store in the app after one subscriber throws", () => {
    const poisoner = writable(0);
    const bystander = writable(0);

    const seen: number[] = [];
    const stopBystander = bystander.subscribe((v) => seen.push(v));
    seen.length = 0;

    // Healthy to start with: a plain `set` notifies synchronously.
    bystander.set(1);
    expect(seen).toEqual([1]);
    seen.length = 0;

    // Two subscribers, because the queue is only used when there is
    // more than one. And the thrower must not throw on its initial
    // value: `subscribe` calls back directly, outside the queue, so a
    // throw there escapes harmlessly. It takes a throw during a `set`
    // — inside the drain loop — to skip the reset.
    poisoner.subscribe(() => {});
    poisoner.subscribe((v) => {
      if (v !== 0) throw new Error("unguarded subscriber");
    });
    expect(() => poisoner.set(1)).toThrow("unguarded subscriber");

    // And now the bystander — a different store, in a different module,
    // with no relationship to the one that threw — is silent.
    bystander.set(2);
    expect(seen, "the queue should be poisoned; see this file's header").toEqual([]);

    stopBystander();
  });
});
