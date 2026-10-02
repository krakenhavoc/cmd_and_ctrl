// dock.ts — the action dock's request store (ADR 0111 §1, Delivery PR 3).
//
// The dock in the bottom-right corner draws whatever the table is asking
// the viewer right now. The things that ask (Game.svelte's combat state,
// PR 3; targeting and the insufficient-mana prompt, PR 4; pending
// choices and sheets in Delivery PRs 5-6) stop drawing their own
// buttons. Each one registers a
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
  // A key that works whatever the shortcut settings say (Enter and
  // Escape belong to the prompt on screen, ADR 0047): its
  // aria-keyshortcuts value, and the small cap drawn on the button
  // ("Esc", "⏎", aria-hidden). "Enter" or "Escape" here is also what
  // the dock's one key handler presses (dockKeyAction below): a key
  // presses the button that advertises it, and no other.
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
  // A toggle's state, drawn as aria-pressed (a vote's own ballot).
  pressed?: boolean;
  // A row button pushed to the row's far end, apart from the options
  // (a vote's "end vote").
  alignEnd?: boolean;
}

// The server refused something the request sent, and the request has
// an answer to offer (ADR 0111 §2, "Refusals"; the attack-tax and
// attack-limit pickers today). Drawn as an alert in the prompt area.
export interface DockRefusal {
  tag: string;
  text: string;
  // gold (the default) for a refusal with an answer to offer; danger
  // for a bare "Not accepted" (an inline choice's refused answer, #624).
  tone?: "gold" | "danger";
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
  // A muted line under the question: what the answer does (PR 5's
  // inline choices carry their modal's hint here). `hintWarn` draws it
  // as a warning (a "may" trigger with no legal target).
  hint?: string;
  hintWarn?: boolean;
  // A row of option buttons under the question, with an optional lead
  // ("Attack all →"). `rowLayout: "stack"` draws one full-width button
  // per line, for option labels that are sentences (option_pick).
  rowLead?: string;
  row?: DockAction[];
  rowLayout?: "wrap" | "stack";
  refusal?: DockRefusal | null;
  // The action bar, for a rank that takes it.
  primary?: DockAction | null;
  secondary?: DockAction[];
  // An inline choice's own body (PR 5: the mana symbols, the loop
  // count), or a sheet's (PR 6). Drawn in the prompt area, under the
  // question.
  body?: Snippet;
  // ADR 0111 §1, keyboard focus: when a choice or a block declaration
  // opens and focus is on the body, the dock moves focus to its
  // primary ("primary", the default) — or, for a question whose
  // primary must not be one stray Enter away (a yes/no, PR 5: "a stray
  // Enter must not accept an optional effect or pay a cost"), to the
  // request's dialog itself ("dialog"), so a screen reader still lands
  // on the question and the keys it names (Y / N) answer it.
  focus?: "primary" | "dialog";
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

// ---- the one Enter / Escape handler (ADR 0111 §1, Delivery PR 4) ------
//
// Enter presses the open request's primary and Escape its cancel or
// refusal. A key presses exactly the button that ADVERTISES it, through
// `keyShortcuts` (its aria-keyshortcuts, and the cap drawn on it): so a
// request opts each button in, and one that must not take Enter (a
// yes/no question, PR 5: a stray Enter must not accept an optional
// effect or pay a cost) simply does not advertise it.
//
// It stands down:
//   - for a step row (the attack row), which does not take the bar —
//     so Enter never presses `next`, and Escape never touches it;
//   - while a modal layer is open (lib/modalLayers.ts): a modal's own
//     keys win, as #1659 fixed for the targeting walk;
//   - while focus is in a text field, select or contenteditable;
//   - for Enter on a focused control (a button, a link, a board card):
//     Enter belongs to the control that has focus. The dock's own
//     buttons press themselves and stop the key there;
//   - when something already handled the key (defaultPrevented), with
//     a modifier held, or mid-composition.

export type DockKey = "Enter" | "Escape";

// The focused element is a text entry: no dock key at all.
export function isTypingTarget(target: EventTarget | null): boolean {
  if (!target || typeof (target as Element).closest !== "function") return false;
  const el = target as HTMLElement;
  if (el.isContentEditable) return true;
  return !!el.closest("input, textarea, select, [contenteditable=''], [contenteditable='true']");
}

// The focused element is a control that answers Enter itself.
export function isControlTarget(target: EventTarget | null): boolean {
  if (!target || typeof (target as Element).closest !== "function") return false;
  return !!(target as Element).closest(
    "button, a[href], summary, [role='button'], [role='menuitem'], [role='option'], [role='checkbox'], [role='radio'], [role='switch'], [role='tab'], [role='link']",
  );
}

// advertises reports whether an action's keyShortcuts names `key`.
function advertises(a: DockAction | null | undefined, key: DockKey): a is DockAction {
  return !!a && !a.disabled && (a.keyShortcuts ?? "").split(/\s+/).includes(key);
}

// dockKeyAction is the action a key presses on `request`, or null.
// Pure, so the rules are pinned without a DOM.
export function dockKeyAction(
  request: DockRequest | null | undefined,
  key: string,
): DockAction | null {
  if (key !== "Enter" && key !== "Escape") return null;
  if (!request || !takesBar(request)) return null;
  const ordered =
    key === "Enter"
      ? [request.primary, ...(request.secondary ?? [])]
      : [...(request.secondary ?? []), request.primary];
  return ordered.find((a) => advertises(a, key)) ?? null;
}

export interface DockKeyContext {
  // modalOpen from lib/modalLayers.ts.
  modalOpen: boolean;
}

// dockKeyFor is the window handler's whole decision: the action this
// keydown presses, or null to leave the key alone.
export function dockKeyFor(
  e: Pick<
    KeyboardEvent,
    | "key"
    | "target"
    | "defaultPrevented"
    | "isComposing"
    | "shiftKey"
    | "ctrlKey"
    | "altKey"
    | "metaKey"
  >,
  request: DockRequest | null | undefined,
  ctx: DockKeyContext,
): DockAction | null {
  if (e.key !== "Enter" && e.key !== "Escape") return null;
  if (e.defaultPrevented || e.isComposing) return null;
  if (e.shiftKey || e.ctrlKey || e.altKey || e.metaKey) return null;
  if (ctx.modalOpen) return null;
  if (isTypingTarget(e.target)) return null;
  if (e.key === "Enter" && isControlTarget(e.target)) return null;
  return dockKeyAction(request, e.key);
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
