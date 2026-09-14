// Keyboard shortcuts — the table, the chord grammar, and the whole
// of the dispatch decision.
//
// Everything a keypress has to be checked against lives here as a
// pure function over plain data: no DOM, no stores, no Svelte. The
// component layer (ShortcutLayer.svelte) does three things and no
// more — read the event, ask `dispatchShortcut` what to do, call the
// handler. That keeps the rules that are easy to get subtly wrong
// (am I typing? is a modal up? is this move even legal?) under unit
// test, which is the same split gameLog.ts / attackAll.ts / timing.ts
// already use.
//
// Three decisions are load-bearing enough to state up front.
//
// 1. WE MATCH ON `event.key`, NOT `event.code`.
//
//    `event.code` is the physical key position: the key labelled Q on
//    a US board reports `KeyQ` on AZERTY too, where the keycap says A.
//    That is the right answer for WASD movement and the wrong answer
//    for a mnemonic keymap — our shortcuts mean "L for log", "U for
//    undo", and a help overlay that says "L" has to match the letter
//    printed on the key the player presses. `event.key` is the
//    character the layout actually produced, so it does.
//
//    The cost of `key` is the shift problem: on a US layout `?` is
//    Shift+/, on a German layout it is Shift+ß, and on some layouts
//    it needs no shift at all. We take the character at face value
//    and record Shift ONLY for letters (see `chordFrom`), so `?` is
//    the chord `?` everywhere and `Shift+a` stays distinct from `a`.
//
// 2. A SHORTCUT IS ENABLED EXACTLY WHEN THE MOVE IS LEGAL, and the
//    server is what says so. `ShortcutContext.passLegal` is a lookup
//    over `GameView.legal_moves` (ADR 0033 §1), not a re-derivation
//    of priority rules in TypeScript — S31 sub-PR 2 deleted a pile of
//    exactly that and we are not growing it back. An absent move list
//    means "no information" and stays permissive, same asymmetry
//    timing.ts documents: a false yes costs a rejected click, a false
//    no costs the player a window they were entitled to.
//
// 3. ONLY OVERRIDES ARE PERSISTED. `settings.shortcuts.bindings` holds
//    the chords the user deliberately changed, never the full map, so
//    a default we retune later reaches everyone who never touched that
//    row. An explicit unbind is stored as the empty string, which is
//    how "the user turned this off" survives a default change too.

// ---------------------------------------------------------------- //
// Actions                                                          //
// ---------------------------------------------------------------- //

export type ShortcutID =
  | "passPriority"
  | "passTurn"
  | "undo"
  | "attackAll"
  | "holdPriority"
  | "toggleAutopass"
  | "toggleGameLog"
  | "drawCard"
  | "toggleMute"
  | "openSettings"
  | "toggleHelp";

export type ShortcutGroup = "priority" | "combat" | "table" | "interface";

export interface ShortcutDef {
  id: ShortcutID;
  // Short imperative name. Also the overlay row title.
  label: string;
  // One line of "what this actually does", for the overlay and the
  // rebinding row. Written to be readable by someone who has never
  // opened Settings.
  hint: string;
  group: ShortcutGroup;
  // Canonical chord, or "" for an action we ship unbound.
  defaultBinding: string;
  // "action" mutates the game and is gated on the seat being able to
  // act at all (not a spectator, not eliminated, game not over, not
  // staring at a replayed frame). "view" only moves client-side UI
  // and stays available to a spectator watching a finished game.
  kind: "action" | "view";
  // "game" actions need a table; "global" ones work from the lobby,
  // the login screen and the catalog too.
  scope: "game" | "global";
}

// SHORTCUTS is the single source of truth. The help overlay, the
// rebinding UI and the dispatcher all read this array, so a row
// cannot exist in the overlay without being dispatchable, and a key
// cannot be dispatchable without appearing in the overlay.
//
// Default-binding policy:
//   - nothing collides with a browser or OS reservation. No Ctrl/Cmd
//     +W/T/R/N, no Ctrl/Cmd+F, no F-keys, no Alt-anything (AltGr on
//     several European layouts is Ctrl+Alt, so an Alt default would
//     eat a letter those players need for typing).
//   - unmodified letters, which is what makes the keymap feel fast.
//     Safe precisely because the dispatcher refuses to fire while a
//     text field has focus.
//   - Escape and Enter are NOT bound here. Both already mean
//     something at every modal and prompt in the game (Escape closes,
//     Enter confirms) and a global handler that swallowed them would
//     break all of it. They are in RESERVED_CHORDS.
//   - Space IS bound, to the single most-pressed action in Magic, but
//     it defers to a focused control — see chordDefersToFocus.
export const SHORTCUTS: readonly ShortcutDef[] = [
  {
    id: "passPriority",
    label: "Pass priority",
    hint: "Yield the current priority window — the same as the phase widget's “next”.",
    group: "priority",
    defaultBinding: "Space",
    kind: "action",
    scope: "game",
  },
  {
    id: "passTurn",
    label: "Pass turn",
    hint: "Skip the rest of your turn. Only while you are the active player.",
    group: "priority",
    defaultBinding: "t",
    kind: "action",
    scope: "game",
  },
  {
    id: "holdPriority",
    label: "Hold priority",
    hint: "Arm the hold toggle so your own spells and triggers keep the cursor.",
    group: "priority",
    defaultBinding: "h",
    kind: "view",
    scope: "game",
  },
  {
    id: "toggleAutopass",
    label: "Toggle autopass",
    hint: "Pass every priority window until you turn it off.",
    group: "priority",
    defaultBinding: "Shift+p",
    kind: "view",
    scope: "game",
  },
  {
    id: "attackAll",
    label: "Attack with all",
    hint: "Declare every eligible creature. Inert when more than one opponent could be attacked — that choice is never guessed.",
    group: "combat",
    defaultBinding: "a",
    kind: "action",
    scope: "game",
  },
  {
    id: "undo",
    label: "Undo",
    hint: "Rewind your most recent action, against your per-turn undo budget.",
    group: "table",
    defaultBinding: "u",
    kind: "action",
    scope: "game",
  },
  {
    id: "drawCard",
    label: "Draw a card",
    hint: "The sandbox draw verb from the ⋯ menu.",
    group: "table",
    defaultBinding: "d",
    kind: "action",
    scope: "game",
  },
  {
    id: "toggleGameLog",
    label: "Game log",
    hint: "Open or close the log drawer.",
    group: "interface",
    defaultBinding: "l",
    kind: "view",
    scope: "game",
  },
  {
    id: "toggleMute",
    label: "Mute sound",
    hint: "Mute or unmute every sound effect and the music bed.",
    group: "interface",
    defaultBinding: "m",
    kind: "view",
    scope: "global",
  },
  {
    id: "openSettings",
    label: "Settings",
    hint: "Open or close the settings panel.",
    group: "interface",
    // Pre-dates this module: Settings.svelte has bound "," since
    // S11.5 and the command bar's gear button advertises it. Kept
    // byte-identical so nobody's muscle memory breaks.
    defaultBinding: ",",
    kind: "view",
    scope: "global",
  },
  {
    id: "toggleHelp",
    label: "Keyboard shortcuts",
    hint: "Show this list.",
    group: "interface",
    defaultBinding: "?",
    kind: "view",
    scope: "global",
  },
];

export const SHORTCUT_IDS: readonly ShortcutID[] = SHORTCUTS.map((s) => s.id);

const BY_ID = new Map<ShortcutID, ShortcutDef>(SHORTCUTS.map((s) => [s.id, s]));

export function shortcutDef(id: ShortcutID): ShortcutDef | undefined {
  return BY_ID.get(id);
}

export const GROUP_LABELS: Record<ShortcutGroup, string> = {
  priority: "Priority",
  combat: "Combat",
  table: "Table",
  interface: "Interface",
};

// GROUP_ORDER fixes the overlay's section order. Declared separately
// from GROUP_LABELS so adding a label can't silently reorder the UI.
export const GROUP_ORDER: readonly ShortcutGroup[] = ["priority", "combat", "table", "interface"];

// ---------------------------------------------------------------- //
// Chords                                                           //
// ---------------------------------------------------------------- //

// A chord is modifiers then key, joined with "+", in a fixed order:
// Ctrl+Alt+Shift+Meta+<key>. Fixed order matters — it is what makes
// two chords comparable with ===, which is all the dispatcher and the
// conflict detector ever do with them.
const MODIFIER_ORDER = ["Ctrl", "Alt", "Shift", "Meta"] as const;

// Keys that are only ever a modifier being held, plus the two the
// browser emits when it does not know what happened. None of them can
// be the key half of a chord.
const NON_KEYS = new Set([
  "Shift",
  "Control",
  "Alt",
  "Meta",
  "AltGraph",
  "CapsLock",
  "NumLock",
  "ScrollLock",
  "Dead",
  "Unidentified",
  "Process",
]);

// Named keys we canonicalise. Everything else with a multi-character
// `key` (ArrowUp, F5, Home, PageDown, …) passes through unchanged —
// those names are already stable across browsers.
const KEY_ALIASES: Record<string, string> = {
  " ": "Space",
  Spacebar: "Space",
  Esc: "Escape",
  Del: "Delete",
  Up: "ArrowUp",
  Down: "ArrowDown",
  Left: "ArrowLeft",
  Right: "ArrowRight",
};

// canonicalKey normalises the key half of a chord. Single characters
// are lowercased (so the letter A and the letter a are one key, and
// punctuation is untouched — "?".toLowerCase() is "?"); named keys go
// through the alias table.
export function canonicalKey(key: string): string {
  const aliased = KEY_ALIASES[key];
  if (aliased) return aliased;
  if (key.length === 1) return key.toLowerCase();
  return key;
}

// KeyChord is the subset of KeyboardEvent this module reads. Declared
// rather than taking KeyboardEvent so the tests can pass object
// literals and the module stays importable without a DOM.
export interface KeyChord {
  key: string;
  ctrlKey?: boolean;
  altKey?: boolean;
  shiftKey?: boolean;
  metaKey?: boolean;
  isComposing?: boolean;
}

// chordFrom builds the canonical chord string for a keypress, or null
// when the press cannot be one (a bare modifier, a dead key, an IME
// composition in flight).
//
// The Shift rule is the interesting half. Shift is recorded only when
// the key is an ASCII letter or a named key; for any other printable
// character it is dropped, because the character the layout produced
// ALREADY encodes the shift. On a US board Shift+/ produces "?" and
// we store the chord "?"; on a layout where ? is unshifted we store
// "?" as well, and the binding works on both. Recording "Shift+?"
// would have bound a chord only one of those two players can type.
export function chordFrom(e: KeyChord): string | null {
  if (e.isComposing) return null;
  if (!e.key || NON_KEYS.has(e.key)) return null;
  const key = canonicalKey(e.key);
  const isLetter = key.length === 1 && key >= "a" && key <= "z";
  const named = key.length > 1;
  const parts: string[] = [];
  if (e.ctrlKey) parts.push("Ctrl");
  if (e.altKey) parts.push("Alt");
  if (e.shiftKey && (isLetter || named)) parts.push("Shift");
  if (e.metaKey) parts.push("Meta");
  parts.push(key);
  return parts.join("+");
}

// parseChord validates and re-canonicalises a stored chord string.
// Returns null for anything malformed — an unknown modifier, an empty
// key, a duplicated modifier — so a hand-edited or hostile settings
// blob can never install a binding the UI cannot display.
export function parseChord(raw: string): string | null {
  if (typeof raw !== "string") return null;
  const trimmed = raw.trim();
  if (!trimmed) return null;
  // Split on "+" but keep a literal "+" key working: a trailing "+"
  // after a modifier run is the key itself.
  const parts = trimmed.split("+");
  if (parts.length > 1 && parts[parts.length - 1] === "") {
    parts.pop();
    parts[parts.length - 1] = "+";
  }
  const key = parts.pop();
  if (!key) return null;
  const mods = new Set<string>();
  for (const p of parts) {
    const m = MODIFIER_ORDER.find((x) => x.toLowerCase() === p.toLowerCase());
    if (!m) return null;
    if (mods.has(m)) return null;
    mods.add(m);
  }
  const canonical = canonicalKey(key);
  if (NON_KEYS.has(canonical)) return null;
  const isLetter = canonical.length === 1 && canonical >= "a" && canonical <= "z";
  const named = canonical.length > 1;
  // Same Shift rule as chordFrom, applied to stored data: a chord
  // that says Shift on a punctuation key could never be produced by
  // chordFrom, so it would be dead weight. Normalise it away.
  if (mods.has("Shift") && !isLetter && !named) mods.delete("Shift");
  const out: string[] = MODIFIER_ORDER.filter((m) => mods.has(m));
  out.push(canonical);
  return out.join("+");
}

// keyOfChord returns the key half.
export function keyOfChord(chord: string): string {
  const parsed = parseChord(chord);
  if (!parsed) return "";
  const parts = parsed.split("+");
  const last = parts[parts.length - 1];
  return last === "" ? "+" : last;
}

// RESERVED_CHORDS can never be bound, by default or by the user.
//
// Escape and Enter are owned by whatever prompt is on screen — every
// modal in the game closes on one and confirms on the other, and
// Game.svelte cancels targeting on Escape and confirms a multi-target
// pick on Enter. Tab is keyboard navigation and is not ours to take.
export const RESERVED_CHORDS: readonly string[] = [
  "Escape",
  "Enter",
  "Tab",
  "Shift+Tab",
  "Shift+Enter",
];

export function isReservedChord(chord: string): boolean {
  const parsed = parseChord(chord);
  if (!parsed) return false;
  return RESERVED_CHORDS.includes(parsed);
}

// chordDefersToFocus reports whether a chord must yield to whatever
// control currently has keyboard focus.
//
// Space and Enter activate the focused button; Space also scrolls. A
// global handler that preventDefault()ed them unconditionally would
// make every card, every menu item and every button in the game
// unusable from the keyboard — the exact accessibility regression a
// shortcut layer is supposed to avoid. So a binding on Space or Enter
// fires only while focus is somewhere inert (the body, the board
// background), and the moment the player has tabbed to a control the
// control wins. Derived from the chord rather than declared per
// action, so the rule follows a user who rebinds something else onto
// Space.
export function chordDefersToFocus(chord: string): boolean {
  const parsed = parseChord(chord);
  if (!parsed) return false;
  const key = keyOfChord(parsed);
  if (key !== "Space" && key !== "Enter") return false;
  // A modified chord (Ctrl+Space) does not activate a button, so it
  // has nothing to defer to.
  return parsed === key || parsed === `Shift+${key}`;
}

// ---------------------------------------------------------------- //
// Bindings                                                         //
// ---------------------------------------------------------------- //

// BindingMap is a complete action → chord map. "" means unbound.
export type BindingMap = Record<ShortcutID, string>;

// BindingOverrides is what gets persisted: only the rows the user
// changed. Absent id = "use whatever the default is today".
//
// Typed as an open record rather than Partial<Record<ShortcutID,…>>
// on purpose — it is stored verbatim in settings.shortcuts.bindings
// and round-trips through JSON, so the reader has to cope with keys
// that are not ShortcutIDs anyway. sanitizeOverrides is where the
// narrowing actually happens.
export type BindingOverrides = Record<string, string>;

export function defaultBindings(): BindingMap {
  const out = {} as BindingMap;
  for (const s of SHORTCUTS) out[s.id] = s.defaultBinding;
  return out;
}

// sanitizeOverrides scrubs a stored override map. Drops unknown
// action ids, unparseable chords and reserved chords; canonicalises
// what is left; and drops any override that has become equal to the
// current default, so the stored blob does not pin a row to a value
// that is no longer a deliberate choice.
//
// Deliberately does NOT drop a conflicting pair. Two actions on one
// chord is a state the user can reach and has to be able to see and
// fix, and silently deleting one of them at load would make the
// Settings panel lie about what it stored. The dispatcher resolves
// the conflict deterministically (declaration order) and the UI flags
// it instead.
export function sanitizeOverrides(raw: unknown): BindingOverrides {
  const out: BindingOverrides = {};
  if (!raw || typeof raw !== "object" || Array.isArray(raw)) return out;
  const src = raw as Record<string, unknown>;
  for (const def of SHORTCUTS) {
    if (!(def.id in src)) continue;
    const v = src[def.id];
    if (typeof v !== "string") continue;
    // The empty string is meaningful: an explicit unbind. Keep it
    // even when the default is also "" — it costs one key in
    // localStorage and it preserves intent if the default changes.
    if (v === "") {
      out[def.id] = "";
      continue;
    }
    const chord = parseChord(v);
    if (!chord) continue;
    if (RESERVED_CHORDS.includes(chord)) continue;
    if (chord === def.defaultBinding) continue;
    out[def.id] = chord;
  }
  return out;
}

// effectiveBindings layers overrides over today's defaults.
export function effectiveBindings(overrides: BindingOverrides | undefined): BindingMap {
  const out = defaultBindings();
  if (!overrides) return out;
  for (const id of SHORTCUT_IDS) {
    const v = overrides[id];
    if (typeof v === "string") out[id] = v;
  }
  return out;
}

// bindingConflicts reports every chord bound to more than one action,
// in declaration order. Unbound rows never conflict.
export function bindingConflicts(map: BindingMap): Map<string, ShortcutID[]> {
  const byChord = new Map<string, ShortcutID[]>();
  for (const def of SHORTCUTS) {
    const chord = map[def.id];
    if (!chord) continue;
    const list = byChord.get(chord);
    if (list) list.push(def.id);
    else byChord.set(chord, [def.id]);
  }
  const conflicts = new Map<string, ShortcutID[]>();
  for (const [chord, ids] of byChord) {
    if (ids.length > 1) conflicts.set(chord, ids);
  }
  return conflicts;
}

// conflictingWith returns the actions a candidate chord would collide
// with if assigned to `id`. The rebinding UI calls this on every
// capture so it can warn before the assignment is committed.
export function conflictingWith(map: BindingMap, id: ShortcutID, chord: string): ShortcutID[] {
  if (!chord) return [];
  return SHORTCUT_IDS.filter((other) => other !== id && map[other] === chord);
}

// BindingProblem is why a captured chord cannot be assigned.
export type BindingProblem = "invalid" | "reserved";

export interface BindingCandidate {
  chord: string;
  problem?: BindingProblem;
  // Actions already holding this chord. Non-empty is a warning, not a
  // refusal — the user may be deliberately moving a key and will fix
  // the other row next.
  conflicts: ShortcutID[];
}

// evaluateBinding is the rebinding input's whole decision: canonicalise
// what was captured, refuse the reserved chords, and report who else
// already holds it.
export function evaluateBinding(
  map: BindingMap,
  id: ShortcutID,
  captured: KeyChord,
): BindingCandidate {
  const chord = chordFrom(captured);
  if (!chord) return { chord: "", problem: "invalid", conflicts: [] };
  if (RESERVED_CHORDS.includes(chord)) return { chord, problem: "reserved", conflicts: [] };
  return { chord, conflicts: conflictingWith(map, id, chord) };
}

// ---------------------------------------------------------------- //
// Availability                                                     //
// ---------------------------------------------------------------- //

// ShortcutContext is the seat's situation, flattened to plain data by
// the component that owns the snapshot. Everything here is a read of
// a field the server stamped or of local UI state — no rules are
// derived in this file.
export interface ShortcutContext {
  // A game route is mounted and a snapshot has arrived.
  atTable: boolean;
  // Read-only session: the server rejects every action frame.
  spectator: boolean;
  // The dev replay scrubber is showing a past frame. Every control
  // that mutates the LIVE game is withheld while it is.
  replaying: boolean;
  gameOver: boolean;
  eliminated: boolean;
  activePlayer: boolean;
  // Does `GameView.legal_moves` contain a `pass` move? `undefined`
  // means the server shipped no move list on this frame, which is not
  // the same as "nothing is legal" — stay permissive.
  passLegal: boolean | undefined;
  // Undos the seat has left this turn. `null` means unlimited (admin).
  undosRemaining: number | null;
  // What an attack-with-all would declare right now, and how many
  // opponents it could be pointed at.
  attackAllEligible: number;
  attackAllDefenders: number;
}

export function idleContext(): ShortcutContext {
  return {
    atTable: false,
    spectator: false,
    replaying: false,
    gameOver: false,
    eliminated: false,
    activePlayer: false,
    passLegal: undefined,
    undosRemaining: null,
    attackAllEligible: 0,
    attackAllDefenders: 0,
  };
}

export interface Availability {
  enabled: boolean;
  // Why not, in tooltip English. Absent when enabled.
  reason?: string;
}

const AVAILABLE: Availability = { enabled: true };

function unavailable(reason: string): Availability {
  return { enabled: false, reason };
}

// shortcutAvailability answers "would this key do anything right
// now?" for one action. The overlay greys a row with the reason; the
// dispatcher refuses to fire.
export function shortcutAvailability(id: ShortcutID, ctx: ShortcutContext): Availability {
  const def = BY_ID.get(id);
  if (!def) return unavailable("unknown action");

  if (def.scope === "game" && !ctx.atTable) return unavailable("only at a game table");

  // Gates that apply to anything that would send an action frame.
  // View-only actions (log drawer, mute, settings) stay available to
  // a spectator watching a finished game — that is the whole point of
  // watching.
  if (def.kind === "action") {
    if (ctx.spectator) return unavailable("spectating — read-only");
    if (ctx.replaying) return unavailable("viewing a past frame");
    if (ctx.gameOver) return unavailable("the game has ended");
    if (ctx.eliminated) return unavailable("you have been eliminated");
  }

  switch (id) {
    case "passPriority":
      // The server's answer, not ours. `undefined` is "no move list on
      // this frame" — permissive, per the asymmetry in ADR 0009 §3.
      if (ctx.passLegal === false) return unavailable("you don't hold priority");
      return AVAILABLE;
    case "passTurn":
      if (!ctx.activePlayer) return unavailable("not your turn");
      return AVAILABLE;
    case "undo":
      if (ctx.undosRemaining !== null && ctx.undosRemaining <= 0) {
        return unavailable("no undos left this turn");
      }
      return AVAILABLE;
    case "attackAll":
      if (ctx.attackAllEligible <= 0) return unavailable("nothing can attack right now");
      if (ctx.attackAllDefenders <= 0) return unavailable("no opponent to attack");
      // Commander is 2–4 players and attack-all has no single meaning
      // with more than one opponent on the table. attackAll.ts refuses
      // to guess and so does this key: the on-screen cluster already
      // renders one labelled button per seat, and the shortcut says so
      // instead of picking a victim.
      if (ctx.attackAllDefenders > 1) {
        return unavailable("more than one opponent — use the labelled attack buttons");
      }
      return AVAILABLE;
    default:
      return AVAILABLE;
  }
}

// ---------------------------------------------------------------- //
// Dispatch                                                         //
// ---------------------------------------------------------------- //

// Why a keypress did nothing. Returned rather than logged so the
// tests can assert on the reason and not just the absence of an
// action.
export type SkipReason =
  | "no-chord"
  | "repeat"
  | "typing"
  | "modal"
  | "disabled"
  | "unbound"
  | "focus-defer"
  | "unavailable";

export interface DispatchInput {
  // The canonical chord, from chordFrom. null for a bare modifier.
  chord: string | null;
  // Key auto-repeat. Never fires an action: holding Space down should
  // pass one priority window, not thirty.
  repeat: boolean;
  // The event target is a text field. See isTypingTarget.
  typing: boolean;
  // At least one modal / popover layer is registered. Modals own the
  // keyboard while they are up.
  modalOpen: boolean;
  // Focus is on a control that has its own meaning for Space / Enter.
  interactiveFocus: boolean;
  // settings.shortcuts.enabled — the master off switch.
  enabled: boolean;
  bindings: BindingMap;
  ctx: ShortcutContext;
}

export interface DispatchResult {
  action: ShortcutID | null;
  skipped?: SkipReason;
  // Populated when `skipped === "unavailable"`, so the caller can
  // surface the reason rather than swallowing the press silently.
  availability?: Availability;
}

const skip = (r: SkipReason): DispatchResult => ({ action: null, skipped: r });

// dispatchShortcut is the entire decision. Order matters and is the
// order the checks are written in:
//
//   typing beats everything — a shortcut that fires while you are
//   naming a game or pasting a decklist is the single most common way
//   this feature becomes infuriating, so it is checked before the
//   binding lookup and before the master toggle even matters.
//
//   modals beat the keymap — while a prompt is up it owns the
//   keyboard, including the keys it does not itself handle. Anything
//   else lets you pass priority through a targeting prompt you have
//   not answered.
export function dispatchShortcut(input: DispatchInput): DispatchResult {
  if (!input.chord) return skip("no-chord");
  if (input.repeat) return skip("repeat");
  if (input.typing) return skip("typing");
  if (input.modalOpen) return skip("modal");
  if (!input.enabled) return skip("disabled");

  // Declaration order breaks a conflict deterministically. The
  // Settings panel flags the pair rather than letting it be a
  // mystery, but a conflicting keymap still has to behave the same
  // way on every press.
  const id = SHORTCUT_IDS.find((sid) => input.bindings[sid] === input.chord);
  if (!id) return skip("unbound");

  if (input.interactiveFocus && chordDefersToFocus(input.chord)) return skip("focus-defer");

  const availability = shortcutAvailability(id, input.ctx);
  if (!availability.enabled) return { action: null, skipped: "unavailable", availability };
  return { action: id };
}

// ---------------------------------------------------------------- //
// DOM predicates                                                   //
// ---------------------------------------------------------------- //

// ElementLike is the structural slice of Element these predicates
// read. Keeps them callable from the tests with object literals and
// keeps this module importable in a node vitest run with no jsdom.
export interface ElementLike {
  tagName?: string;
  isContentEditable?: boolean;
  getAttribute?: (name: string) => string | null;
  closest?: (selector: string) => unknown;
}

// elementOf narrows an EventTarget (which may be the window, a
// document, or an XHR) to something the predicates below can read.
// Structural rather than `instanceof Element` so it works in a node
// vitest run with no DOM globals.
export function elementOf(t: unknown): ElementLike | null {
  if (!t || typeof t !== "object") return null;
  return typeof (t as ElementLike).tagName === "string" ? (t as ElementLike) : null;
}

const TYPING_TAGS = new Set(["INPUT", "TEXTAREA", "SELECT"]);

// isTypingTarget reports whether a keypress belongs to a text field.
//
// Every input counts, not just the text-ish ones: a focused checkbox
// or radio uses Space natively, and a number spinner uses the arrow
// keys, so handing any of them to the keymap would break a control
// the player is actively using. The deck-paste box, the bug-report
// form, the mulligan-size spinner and the undo-limit spinner are all
// covered by the tag check; `contenteditable` and `role="textbox"`
// cover a rich editor if one ever lands.
export function isTypingTarget(el: ElementLike | null | undefined): boolean {
  if (!el) return false;
  const tag = (el.tagName ?? "").toUpperCase();
  if (TYPING_TAGS.has(tag)) return true;
  if (el.isContentEditable) return true;
  const role = el.getAttribute?.("role");
  if (role === "textbox" || role === "searchbox" || role === "combobox") return true;
  // A keypress inside a contenteditable subtree targets the deepest
  // element, which may itself be a plain <span>.
  if (el.closest?.('[contenteditable]:not([contenteditable="false"])')) return true;
  return false;
}

const INTERACTIVE_TAGS = new Set([
  "BUTTON",
  "A",
  "INPUT",
  "TEXTAREA",
  "SELECT",
  "SUMMARY",
  "OPTION",
]);

const INTERACTIVE_ROLES = new Set([
  "button",
  "link",
  "checkbox",
  "radio",
  "menuitem",
  "menuitemcheckbox",
  "menuitemradio",
  "option",
  "tab",
  "switch",
  "slider",
  "spinbutton",
]);

// isInteractiveFocus reports whether the focused element has its own
// meaning for Space / Enter. Used only by chordDefersToFocus's rule;
// nothing else in the dispatcher cares where focus is.
export function isInteractiveFocus(el: ElementLike | null | undefined): boolean {
  if (!el) return false;
  const tag = (el.tagName ?? "").toUpperCase();
  if (INTERACTIVE_TAGS.has(tag)) return true;
  const role = el.getAttribute?.("role");
  if (role && INTERACTIVE_ROLES.has(role)) return true;
  // Svelte components render focusable divs with tabindex for the
  // card grid; treat anything deliberately focusable as interactive.
  const ti = el.getAttribute?.("tabindex");
  if (ti !== null && ti !== undefined && ti !== "-1") return true;
  return false;
}

// ---------------------------------------------------------------- //
// Display                                                          //
// ---------------------------------------------------------------- //

const MAC_SYMBOLS: Record<string, string> = {
  Ctrl: "⌃",
  Alt: "⌥",
  Shift: "⇧",
  Meta: "⌘",
};

const KEY_LABELS: Record<string, string> = {
  Space: "Space",
  ArrowUp: "↑",
  ArrowDown: "↓",
  ArrowLeft: "←",
  ArrowRight: "→",
};

// formatChord renders a chord for a <kbd>. Letters are upper-cased
// because that is what is printed on the keycap; on macOS the
// modifiers become their glyphs and lose the separators, matching
// every other Mac app.
export function formatChord(chord: string, mac = false): string {
  const parsed = parseChord(chord);
  if (!parsed) return "";
  const key = keyOfChord(parsed);
  // Slice the modifiers off by length rather than splitting on "+",
  // so a binding on the literal + key doesn't lose its own key.
  const head = parsed.slice(0, parsed.length - key.length);
  const mods = head ? head.slice(0, -1).split("+") : [];
  const label = KEY_LABELS[key] ?? (key.length === 1 ? key.toUpperCase() : key);
  if (mac) return mods.map((m) => MAC_SYMBOLS[m] ?? m).join("") + label;
  return [...mods, label].join("+");
}

// isMacLike is the one place this module looks at the host. Separated
// so callers can pass the answer into formatChord and keep the
// formatter pure.
export function isMacLike(): boolean {
  if (typeof navigator === "undefined") return false;
  const ua = navigator.userAgent ?? "";
  return /Mac|iPhone|iPad|iPod/.test(ua);
}
