// dock.ts — the action dock's request store (ADR 0111 §1, Delivery PR 3).
//
// The dock in the bottom-right corner draws whatever the table is asking
// the viewer right now. The things that ask (Game.svelte's combat state,
// PR 3; targeting and the insufficient-mana prompt, PR 4; inline
// pending choices, PR 5; every big picker as a sheet that grows up out
// of the dock, PR 6) stop drawing their own
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
import { L } from "./labels";

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
  // ADR 0111 §3 (Delivery PR 6, owner decision 2): the request is a
  // SHEET. Its `body` needs room (a scry, a card grid, a cost picker,
  // the mulligan hand), so the dock draws it in a panel that grows up
  // out of the dock instead of in the prompt area, with a minimise
  // control on its top edge. Its buttons are still the action bar's.
  // Use components/board/DockSheet.svelte rather than setting this by
  // hand: it also registers the sheet's modal layer.
  sheet?: DockSheetSpec;
}

export interface DockSheetSpec {
  // The sheet's heading, and the restore chip's text while it is
  // minimised. Defaults to the request's label.
  title?: string;
  // The small mono source tag beside the heading ("CR 701.22"). It is
  // aria-hidden, as it was on the modals, so the dialog's name is the
  // heading alone.
  src?: string;
  // The running count under the body ("1 / 2 selected"). Drawn with
  // the `.prompt-count` class the e2e suite reads.
  count?: string;
  // How wide the body wants to be, in px, before the screen caps it
  // (ADR 0111 §3: "up to min(720px, 100% - 24px)"). Never narrower than
  // the dock itself. Default 560.
  width?: number;
  // A new key is a new question: a sheet minimised for the last one
  // comes back up. Defaults to the label.
  key?: string;
  // Moves the sheet's body (rendered in the picker's own component
  // tree, by DockSheet) into `host`, the dock's sheet panel, and returns
  // the undo. The dock calls it while this request is the one it draws.
  // A body rendered by the dock itself, as a snippet, would update in
  // the dock's effect tree before the picker's `{#if}` closes it, and
  // read a prompt that is already gone. Without `attach`, the request's
  // `body` snippet is drawn in the panel instead.
  attach?: (host: HTMLElement) => () => void;
}

// ---- a sheet's buttons (PR 6) ----------------------------------------
//
// A sheet's confirm and cancel, the way every picker spells them, so the
// keys and their caps are the same on every one.
//
// confirmAction is a sheet's primary. It takes Enter (ADR 0111 §1:
// "Enter is for confirms that commit what the player already picked")
// unless `enter: false`: a confirm that is really a decline ("Fail to
// find", "Reveal nothing") or a payment ("Pay {2}") must not be one
// stray Enter away.
export function confirmAction(
  label: string,
  onPress: () => void,
  opts: { id?: string; disabled?: boolean; enter?: boolean; title?: string } = {},
): DockAction {
  const enter = opts.enter ?? true;
  return {
    id: opts.id ?? "confirm",
    label,
    disabled: opts.disabled,
    title: opts.title,
    keyShortcuts: enter ? "Enter" : undefined,
    cap: enter ? "⏎" : undefined,
    onPress,
  };
}

// cancelAction is a flow's Cancel (a cost picker, the auto-tap
// preview, the attack picker). It takes Escape. A pending choice has
// none: the game is waiting on an answer, not on a way out.
export function cancelAction(onPress: () => void, label: string = L.cancel): DockAction {
  return { id: "cancel", label, keyShortcuts: "Escape", cap: "Esc", onPress };
}

// The widest a sheet may ask to be (ADR 0111 §3).
export const SHEET_MAX_WIDTH = 720;
export const SHEET_DEFAULT_WIDTH = 560;

// sheetKey is what a minimised sheet is remembered by.
export function sheetKey(request: DockRequest | null | undefined): string | null {
  if (!request?.sheet) return null;
  return request.sheet.key ?? request.label;
}

// sheetWidth is the width a sheet asks for, capped at the ADR's 720px.
export function sheetWidth(request: DockRequest | null | undefined): number {
  const w = request?.sheet?.width ?? SHEET_DEFAULT_WIDTH;
  return Math.max(0, Math.min(SHEET_MAX_WIDTH, w));
}

// sheetMaxHeight is the tallest a sheet may be over a play area
// `playH` tall with a dock `dockH` tall under it: 60% of the play area
// on a desktop, 70% on a phone (§3, §8), and never taller than the room
// left above the dock. 0 means "not measured" (the CSS fallback holds).
export function sheetMaxHeight(playH: number, dockH: number, phone: boolean): number {
  if (!(playH > 0)) return 0;
  const share = (phone ? 0.7 : 0.6) * playH;
  const room = playH - dockH - 24;
  return Math.max(120, Math.floor(Math.min(share, room)));
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
//     keys win, as #1659 fixed for the targeting walk. The dock's own
//     sheets (PR 6) register a "sheet" layer, which does not count;
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
  // foreignModalOpen from lib/modalLayers.ts: a dialog that is not the
  // dock's own sheet (PR 6) is on screen. A sheet's layer does not
  // count, because its Enter and Escape are this handler's.
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
