// dock.ts — the action dock's request store (ADR 0111 §1, Delivery PR 3).
//
// The dock in the bottom-right corner draws whatever the table is asking
// the viewer right now. The things that ask (Game.svelte's combat state
// today; targeting, payment, pending choices and sheets in Delivery
// PRs 4-6) stop drawing their own buttons. Each one registers a
// REQUEST here instead, the way every modal registers a ModalLayer:
// mount a <DockRequest request={…} /> inside its own `{#if}`, and the
// request lives exactly as long as that block is on screen.
//
// A request that forgets to register is a visible failure (its buttons
// are missing), never an invisible one (a stale flag that hides `next`
// for the rest of the game), which is why this is a registration and
// not a list of open-flags the dock checks.
//
// The dock draws ONE request, picked by precedence (ADR 0111 §1, "The
// action bar's primary is picked by precedence, strongest first"):
//
//   1. choice   — a pending choice you owe, the mulligan, a discard
//   2. flow     — a flow you started: targeting, a cost, a combat
//                 selection
//   3. blocks   — a block declaration you owe
//   4. gameOver — Back to lobby
//   5. step     — a step's row of optional moves (the attack row). It
//                 does NOT take the action bar: `next` and Pass turn
//                 stay, because the step continues with `next`.
//
// Ties go to the most recently opened request, like a stack.
//
// Every rank above `step` takes the action bar: its secondaries go on
// the left, its one primary in the corner, and `next` and Pass turn give
// way until it closes. A bar-taking request may have no primary at all
// (a combat selection is committed by a click on the board), and then
// the corner is empty.

import { get, type Readable } from "svelte/store";
import type { Snippet } from "svelte";

import { guardedDerived, guardedWritable } from "./guardedStore";
import type { IconName } from "./icons";

export type DockRank = "choice" | "flow" | "blocks" | "gameOver" | "step";

// Strongest first. Exported for tests and for the dock's own ordering.
export const DOCK_RANKS: readonly DockRank[] = ["choice", "flow", "blocks", "gameOver", "step"];

// A button the dock draws for a request.
export interface DockAction {
  // Stable within its request; the keyed `#each` uses it.
  id: string;
  // The visible text, and the accessible name unless `ariaLabel` is
  // set. Names are an e2e and tutorial contract (ADR 0111 §10).
  label: string;
  ariaLabel?: string;
  title?: string;
  disabled?: boolean;
  onPress: () => void;
  // The bound chord, e.g. "a" or "Escape". The dock prints it in the
  // tooltip and as aria-keyshortcuts, through the same formatter the
  // rest of the dock uses, so a rebound key updates both.
  chord?: string;
  // A key that works whatever the shortcut settings say (Escape belongs
  // to the prompt on screen, ADR 0047): its aria-keyshortcuts value,
  // and the small cap drawn on the button ("Esc", aria-hidden).
  keyShortcuts?: string;
  cap?: string;
  // A leading icon. With an empty label, set `ariaLabel`.
  icon?: IconName;
  // A seat-colour dot before the label (the per-opponent attack buttons).
  seatColor?: string;
  // Muted text after the label, part of its name ("{1}" attack tax).
  note?: string;
  // The one emphasised button of a prompt row (not the bar's primary).
  emphasis?: boolean;
}

// The server refused something the request sent, and the request has
// an answer to offer (ADR 0111 §2, "Refusals"; the attack-tax and
// attack-limit pickers today). Drawn as an alert in the prompt area.
export interface DockRefusal {
  tag: string;
  text: string;
  detail?: string;
  actions: DockAction[];
  onDismiss?: () => void;
}

export interface DockRequest {
  rank: DockRank;
  // The accessible name of the request's non-modal dialog: the name
  // the control it replaces carried, so e2e locators keep working.
  label: string;
  // A `role="group"` name around the prompt area (question, row,
  // refusal), when the ADR names one ("declare attackers").
  group?: string;
  // The question line: a short tag ("attack"), the question, and a
  // muted detail after it.
  tag?: string;
  tone?: "danger" | "gold" | "plain";
  question?: string;
  detail?: string;
  // The question line is a polite live region (a combat selection's
  // hint is announced, as the strip's was).
  live?: boolean;
  // A row of option buttons under the question, with an optional lead
  // ("Attack all →").
  rowLead?: string;
  row?: DockAction[];
  refusal?: DockRefusal | null;
  // The action bar, for a rank that takes it.
  primary?: DockAction | null;
  secondary?: DockAction[];
  // Reserved for Delivery PRs 5 and 6: an inline choice's own body, or
  // a sheet's. Drawn in the prompt area, under the question.
  body?: Snippet;
}

export interface DockHandle {
  update(request: DockRequest): void;
  close(): void;
}

interface Entry {
  seq: number;
  request: DockRequest;
}

const entries = guardedWritable<ReadonlyMap<number, Entry>>(new Map(), "dockRequests");

let nextID = 1;
let nextSeq = 1;

function rankIndex(r: DockRank): number {
  return DOCK_RANKS.indexOf(r);
}

// orderDockRequests sorts strongest first: by rank, then newest first.
export function orderDockRequests(list: Iterable<Entry>): DockRequest[] {
  return [...list]
    .sort((a, b) => rankIndex(a.request.rank) - rankIndex(b.request.rank) || b.seq - a.seq)
    .map((e) => e.request);
}

// dockRequests is every open request, strongest first.
export const dockRequests: Readable<DockRequest[]> = guardedDerived(
  entries,
  (m) => orderDockRequests(m.values()),
  "dockRequests",
  [],
);

// activeDockRequest is the one the dock draws, or null.
export const activeDockRequest: Readable<DockRequest | null> = guardedDerived(
  dockRequests,
  (list) => list[0] ?? null,
  "activeDockRequest",
  null,
);

// takesBar reports whether a request replaces `next` and Pass turn in
// the action bar. Only a step row leaves them.
export function takesBar(request: DockRequest | null | undefined): boolean {
  return !!request && request.rank !== "step";
}

// pushDockRequest opens a request and returns its handle. `update`
// replaces its content and keeps its place in the order (a live count
// changing must not make an older request jump above a newer one);
// `close` removes it, once — a second call is a no-op.
export function pushDockRequest(request: DockRequest): DockHandle {
  const id = nextID++;
  const seq = nextSeq++;
  entries.update((prev) => new Map(prev).set(id, { seq, request }));
  return {
    update(next: DockRequest): void {
      entries.update((prev) => {
        const e = prev.get(id);
        if (!e) return prev;
        return new Map(prev).set(id, { seq: e.seq, request: next });
      });
    },
    close(): void {
      entries.update((prev) => {
        if (!prev.has(id)) return prev;
        const next = new Map(prev);
        next.delete(id);
        return next;
      });
    },
  };
}

// currentDockRequest is the non-reactive read, for call sites outside a
// Svelte reactive scope.
export function currentDockRequest(): DockRequest | null {
  return orderDockRequests(get(entries).values())[0] ?? null;
}

// _resetForTests is the vitest teardown hook. Not for prod use.
export function _resetForTests(): void {
  entries.set(new Map());
  nextID = 1;
  nextSeq = 1;
}
