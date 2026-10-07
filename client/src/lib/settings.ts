import { get, type Writable } from "svelte/store";
import { guardedWritable } from "./guardedStore";
import { setMuted, setVolumeMultiplier } from "./sounds";
import { setMusicMuted, setMusicVolumeMultiplier } from "./music";
import { setAnimationConfig } from "./animations";
import { STEP_IDS, NO_PRIORITY_STEPS, hasOwnStop, type StepID } from "./turn";
import { sanitizeOverrides } from "./shortcuts";
import { DEFAULT_STACK_STYLE, isStackStyle, type StackStyle } from "./stackLane";
import { DEFAULT_SKIN, isSkin, normalizeAccent, type Skin } from "./skins";
import { DEFAULT_TABLE_LAYOUT, isTableLayout, type TableLayout } from "./tableLayout";
import { normalizeSeen } from "./hints/seen";
import { DEFAULT_PLAYMATS_MODE, isPlaymatsMode, type PlaymatsMode } from "./playmatMode";

// Settings is the client-wide preferences schema. Every toggle the
// Settings panel surfaces maps to a field here. Persisted to
// localStorage under STORAGE_KEY so preferences survive reloads,
// and broadcast via a Svelte writable so components react live
// without page refresh.
//
// Schema is versioned — bumping SETTINGS_VERSION and adding a
// migration in `migrate` lets us evolve the shape without stranding
// users on old JSON. Current version is 1. Absorbs the pre-S11.5
// `cmdctrl.muted` legacy key so mute state is preserved across the
// introduction of this module.
//
// Nothing in this module touches the DOM. Reactive application of
// settings is handled by consumers ($effect blocks in App.svelte,
// Settings.svelte, etc.), so this module stays testable from plain
// JS and safe to import from SSR contexts.

export interface Settings {
  // Schema version. Bump when fields change shape; add a migration.
  __version: number;

  audio: {
    // Master mute for every SFX. Replaces the pre-S11.5
    // localStorage["cmdctrl.muted"] flag.
    muted: boolean;
    // 0..100 linear; sounds.ts applies as multiplier on HTMLAudioElement.volume.
    masterVolume: number;
    // 0..100, applied to the effects bus (draw, tap, damage, etc).
    effectsVolume: number;
    // Music is scaffolded for a future music track; no-op until then.
    musicVolume: number;
  };

  animations: {
    // Master toggle for transitions + tweens. Auto-off if the OS
    // signals prefers-reduced-motion: reduce at first load; user
    // override persists thereafter.
    enabled: boolean;
    // Multiplier on tween/transition durations. Values >1 slow
    // animations down; <1 speed them up. Typed as one of a fixed
    // set so the UI can use a radio select.
    speed: 0.5 | 1 | 1.5 | 2;
    // Per-effect toggles. Disabled when `enabled === false`
    // regardless of the individual flag.
    cardDraw: boolean;
    cardPlay: boolean;
    cardTap: boolean;
    cardUntap: boolean;
    cardFlip: boolean;
    particlesEtb: boolean;
    damagePopups: boolean;
    // ADR 0121 §7: dice and coins tumble at the roller's seat. Off (or
    // the master switch off, or reduced motion) the result still shows,
    // settled, for the same hold; only the motion goes.
    dice: boolean;
  };

  display: {
    // The skin: a block of colour tokens in app.css (lib/skins.ts).
    // Stored under its pre-skins name, "theme", so the synced account
    // copy keeps working.
    theme: Skin;
    // A custom accent colour (#rrggbb) over the skin's own, or "" for
    // the skin's. The rest of the accent family is derived from it.
    accent: string;
    // Battlefield card size. Applied as a CSS var so re-tuning a
    // preference re-renders without DOM rebuilds.
    cardSize: "small" | "medium" | "large";
    // Hand fan vs. stack — fan is the existing S06 default.
    handLayout: "fan" | "stacked";
    // Table layout (Sept 2026 redesign). "quadrant" (default) keeps
    // the around-the-table seating (next seat bottom-left, the two
    // across-table seats on top); "row" puts every opponent in turn
    // order across the top and gives your panel the full width.
    //
    // #956 note: at THREE players those two are now identical — the
    // viewer needs the whole bottom row there, and there is no third
    // arrangement worth having.
    //
    // "focus" (#2336) is the row arrangement split evenly: your board is
    // the bottom half, and every opponent is a summary in the top half,
    // whatever opponentDetail and expandActivePlayer say, and you
    // hover or click an avatar to see a whole board. See tableLayout.ts.
    tableLayout: TableLayout;
    // #1467, ADR 0119 §1: how the stack is drawn. "pile" (the default
    // since v17) is a pile of large, readable cards on the left of the
    // table; phones and short boards draw "compact" instead, without
    // changing this value. "compact" is the docked card in the
    // top-left attention strip. The other three float a lane over the
    // middle of the table while the stack or pending triggers are
    // live — "fan" (cards with arrows to their targets), "spotlight"
    // (the next item large, the queue beside it) and "ribbon" (a
    // numbered row). The board grid never reflows for any of them; see
    // lib/stackLane.ts.
    stackStyle: StackStyle;
    // How an opponent's board is drawn. "summary" (the default)
    // renders a dense read-out — life, untapped mana by colour,
    // creature pips carrying P/T — and expands that seat to a full
    // board when an interaction needs card-level clicks. "full"
    // keeps every opponent rendered as cards at all times, which is
    // how the table worked before.
    //
    // The summary exists because card size is a share of panel
    // height with a clamp() floor (the #858 ramp): past the floor a
    // small panel stops shrinking its cards and starts clipping them,
    // which is #956. A representation that degrades by CHANGING
    // rather than SCALING has no floor to hit. See seatSummary.ts.
    opponentDetail: "summary" | "full";
    // Expand the active player's panel for the duration of their
    // turn. On by default: it changes on a turn boundary, so it
    // cannot land mid-click the way a "something interesting
    // happened" trigger would. Off means a seat only ever expands
    // from something the viewer did — a targeting prompt, block or
    // attack mode, or clicking an avatar to pin it.
    expandActivePlayer: boolean;
    // TEMPORARY. Which mechanism an expanding panel uses.
    //
    // "reflow" gives the expanded seat a larger grid share and
    // shrinks the summaries, keeping every panel in one plane so
    // CombatArrows can measure both endpoints against boardEl.
    // "overlay" floats the expanded panel over the table, which
    // animates far more easily but puts one arrow endpoint across a
    // z-index boundary.
    //
    // Both ship so they can be compared in a real game. ONE OF THEM
    // IS GOING TO BE DELETED along with this setting once that
    // comparison has an answer — don't build anything else on it.
    expandStyle: "reflow" | "overlay";
    // Hover preview delay in milliseconds, 0..1000. S11 hover
    // preview reads this as its activation threshold.
    hoverDelayMs: number;
    // Show opponent hand count badge on their panel. Existing
    // behaviour is always-on; this lets a viewer who finds it
    // distracting hide it.
    showOpponentHandCount: boolean;
    // #2209 (was #1954's single `artOnlyCards`). Where a card is drawn
    // as an art tile — Scryfall's art_crop with a name strip, its P/T
    // or loyalty, counters, status marks and keyword chips — instead
    // of the full card. The hover zoom always shows the whole card.
    // The stack, prompts, catalog, deck views and every face-down card
    // are untouched. Not in practiceTable's forced list on purpose.
    //   battlefieldArt — every permanent on the battlefield. Default ON
    //                    (owner answer 1, 2026-10-04).
    //   handArt        — the viewer's own hand. Default OFF: the hand is
    //                    where a player reads what a card does.
    battlefieldArt: boolean;
    handArt: boolean;
    // ADR 0128: whose playmats are drawn behind the battlefields.
    // "all" (the default) draws every player's, "mine" only yours,
    // "off" none. Per device: it is about this screen (a low-power
    // phone, a window where someone else's art is a distraction), not
    // about the person. Having a playmat of your own is an account
    // matter and lives on the server (PUT /me/playmat), not here.
    playmats: PlaymatsMode;
  };

  gameplay: {
    // Confirm-before-exit when navigating away from an active game.
    confirmExit: boolean;
    // Pass priority automatically outside the stops grid and the
    // #1307 key windows. Off means every priority window waits for a
    // click. The rules live in autopassDecision.ts.
    autoPassPriority: boolean;
    // Per-step stops (S13). For each priority-granting step, true
    // means "stop here when priority lands on me" and false means
    // "auto-pass through it". Untap and Cleanup are not stoppable
    // (they don't grant priority) and are absent from this map.
    // Defaults seeded by defaultStepStops().
    stepStops: Record<string, boolean>;
    // S15: mana-cost enforcement. When true, the client tags every
    // cast_spell action with `strict: true, auto_tap: true` (ADR
    // 0118 §1; and, since #1296, every catalog activate_ability the
    // same way — manaEnforcement.ts) and the server taps what the
    // pool is missing and gates the cast on the payment (plus
    // commander tax for casts from the command zone). Default true
    // since ADR 0118 (settings v19); false is the sandbox / paper
    // tracking posture. A card the board can't pay for still offers
    // "Cast anyway (don't pay)", which casts with `force_cast: true`.
    strictMana: boolean;
    // #1530: always raise the CR 603.3b "order your triggers" prompt,
    // even for a batch the server would order itself because every item
    // commutes (an all-prowess batch, #1511). The decision is the
    // server's, so the server holds the seat's copy and the client keeps
    // it in step (triggerOrderPref.ts). Default off.
    alwaysAskTriggerOrder: boolean;
    // S13.6: when a stopped step lands on the viewer but the
    // legality engine reports no legal response (no castable hand
    // cards, no battlefield activations, no commander cast),
    // auto-pass anyway. Defaults on — the step-stops grid gets to
    // mean "stop if there's something to consider" instead of
    // "stop every time regardless." Flip off to restore strict
    // pre-S13.6 behaviour where every stop demands a click.
    smartAutoPass: boolean;
    // #1307: what counts as a response for smartAutoPass. Each is a
    // category of the viewer's own legal moves; mana abilities and
    // land plays are never responses. All on by default.
    //   respondCounterspells  — casts / activations that target the stack
    //   respondInstants       — any other instant-speed cast
    //   respondAbilities      — any other non-mana activated ability
    //   respondSpecialActions — foretell, suspend, turning face up
    respondCounterspells: boolean;
    respondInstants: boolean;
    respondAbilities: boolean;
    respondSpecialActions: boolean;
    // #1307: stop for every opponent item on the stack, answer or
    // not — the pre-#1307 behaviour. Off by default: with smart
    // autopass on, a spell you can't respond to now passes.
    alwaysStopOpponentStack: boolean;
    // #1307 bluffing. When smart autopass would pass a window you
    // cannot answer, act as if you could instead, so a pause gives
    // nothing away.
    //   bluffCounterspell — represent a counter: bluff at an
    //                       opponent's item on the stack.
    //   bluffInstant      — represent an instant: bluff there and in
    //                       the other key windows (combat, an
    //                       opponent's end step).
    //   bluffMode         — "timed" holds for a random delay between
    //                       the two bounds then passes; "manual"
    //                       holds until you click next.
    // Both bluffs also need the in-game bluff toggle (bluff.ts),
    // which starts on at game load when either is set.
    bluffCounterspell: boolean;
    bluffInstant: boolean;
    bluffMode: "timed" | "manual";
    bluffDelayMinMs: number;
    bluffDelayMaxMs: number;
    // ADR 0119 §2: an automatic pass on a stack whose top item someone
    // else controls waits until that item has been on screen this long,
    // so a spell nobody can answer is still readable. 0 is off; the
    // Settings choices are 0–3 s, and stackHold.ts clamps a stored
    // value to that range where it reads it. Never holds `next`.
    stackHoldMs: number;
    // #323: when every item on the stack is one the viewer put
    // there, auto-pass instead of asking "Counter or Pass?" about
    // your own spell. Defaults on — casting is already the
    // decision, so the follow-up click is pure friction. A stack
    // that holds ANY opponent item still stops, and the session
    // "hold" toggle (holdPriority.ts, surfaced in the action dock
    // and the stack card) suspends this per-window when you do want
    // to respond to your own spell or trigger. Flip off to restore
    // the pre-#323 "every stack stops" behaviour permanently.
    autoPassOwnStack: boolean;
    // S13.6 autopass-mode safety. When OFF (default), the autopass
    // toggle auto-clears the first time the cursor reaches the
    // viewer's own precombat_main — a safety belt so you don't
    // skip your own turn because you forgot to turn off autopass
    // before it cycled back to you. When ON, autopass stays
    // engaged until manually toggled off. Labelled DANGER in the
    // UI; anyone opting in has decided they'd rather eat the risk
    // of a skipped turn than re-toggle every cycle.
    autopassPersistThroughTurns: boolean;
    // #170: right-click any card for a per-card override menu —
    // move between zones, add / remove counters, mark damage,
    // declare combat by hand. Off by default, so right-click keeps
    // its historic meaning (the mana / activated-ability popover)
    // until a player opts in. NOT gated on the admin role: every
    // action the menu can send is already gated server-side to the
    // caller's own cards (requireCardController), so a seated
    // player who turns it on gains a surface, not authority.
    adminOverrides: boolean;
    // S31 sub-PR 8: surface a bot's stated reasoning in the table
    // feed. Bots always announce an improvisation — that disclosure
    // is mandatory and this toggle does not touch it — but the
    // per-move "why" is debug output and off by default.
    //
    // Worth knowing before turning it on: a bot's reasoning can
    // mention cards in its OWN hand. That is a disadvantage to the
    // bot, not a leak of yours (a policy never sees another seat's
    // hidden state), but it does make the game easier.
    showBotReasoning: boolean;
    // ADR 0105 (#1789): mark what you can do right now — a ring on a
    // castable card or a playable land, "N ready" on a pile — while
    // you owe a decision. Off removes only that positive treatment;
    // the dimming of a card you cannot play is a gate and stays.
    highlightLegalActions: boolean;
  };

  shortcuts: {
    // Master switch for the global keymap. Off leaves every local
    // handler alone — modals still close on Escape, prompts still
    // confirm on Enter, cards are still operable from the keyboard.
    // Only the global layer stands down.
    enabled: boolean;
    // OVERRIDES ONLY. A row the user has never touched is absent
    // here and resolves against shortcuts.ts's current default, so
    // retuning a default reaches everyone who did not deliberately
    // choose otherwise. An explicit unbind is stored as "" — which
    // is exactly why this can't be a full map with holes.
    //
    // Keys are ShortcutID; typed as a loose record so settings.ts
    // does not have to move whenever an action is added, and
    // sanitized on the way in by shortcuts.sanitizeOverrides.
    bindings: Record<string, string>;
  };

  accessibility: {
    // Honours the OS prefers-reduced-motion signal at first load;
    // subsequent flips persist as explicit overrides.
    reduceMotion: boolean;
    // CSS --font-scale multiplier. 1.0 = browser default.
    textScale: 0.9 | 1.0 | 1.2 | 1.5;
    // Colour-blind-friendly seat palette override. Swaps colors.ts
    // default palette for an alternate set.
    colorblindPalette: boolean;
    // Always-visible focus outlines. Overrides :focus-visible
    // suppression on buttons + inputs.
    alwaysShowFocus: boolean;
  };

  // ADR 0125 §4: the first-use hints. Both fields are per person, so a
  // signed-in person's seen hints follow them to every browser, and a
  // guest's stay in this one.
  help: {
    // Hint id → the version the person last dismissed (lib/hints/). A
    // hint is unseen when its id is missing or its stored version is
    // lower than its own, so bumping a hint's version offers it again.
    // Ids this client does not know are KEPT: an older tab must not
    // drop a newer client's hints. Only retired ids (hints/retired.ts)
    // are dropped. Merged by union at sign-in and on a 412, never
    // field-wins (settingsSync.ts).
    seen: Record<string, number>;
    // "Hide tips": no hint is offered until it is switched back on.
    tipsOff: boolean;
  };
}

export const SETTINGS_VERSION = 21;
const STORAGE_KEY = "cmdctrl.settings.v1";
const LEGACY_MUTED_KEY = "cmdctrl.muted";

// defaultStepStops seeds the per-step stops map. The defaults match
// MTG Online's standard "stops" — the active player gets stopped on
// their main phases and combat declarations; everyone else passes
// through routine begin/end-step priority unless they opt in.
// Untap and Cleanup are excluded because they don't grant priority
// (CR 502.4 / 514.3); the server's NoPriority sentinel makes any
// attempt to pass during them a no-op anyway. The first-strike combat
// damage step is excluded too, for a different reason: it shares the
// combat_damage stop (turn.ts `stopKeyFor`), so the map keeps the same
// keys it always had and no stored blob needs migrating.
export function defaultStepStops(): Record<string, boolean> {
  const out: Record<string, boolean> = {};
  const opted: ReadonlySet<StepID> = new Set([
    "precombat_main",
    "declare_attackers",
    "declare_blockers",
    "postcombat_main",
    "end",
  ]);
  for (const id of STEP_IDS) {
    if (!hasOwnStop(id)) continue;
    out[id] = opted.has(id);
  }
  return out;
}

// prefersReducedMotion reads the OS hint without subscribing. Used
// to pick the initial default for accessibility.reduceMotion when
// the user has no saved settings yet. SSR-safe.
function prefersReducedMotion(): boolean {
  if (typeof window === "undefined" || !window.matchMedia) return false;
  return window.matchMedia("(prefers-reduced-motion: reduce)").matches;
}

// defaultSettings assembles a fresh Settings object with reasonable
// defaults. Call-site runs once at first load or on reset.
export function defaultSettings(): Settings {
  const reduced = prefersReducedMotion();
  return {
    __version: SETTINGS_VERSION,
    audio: {
      muted: false,
      masterVolume: 80,
      effectsVolume: 100,
      musicVolume: 60,
    },
    animations: {
      // If the OS asks for reduced motion, we start with animations
      // off — but the user can flip them on without overriding the
      // OS pref for other apps (our storage is per-site).
      enabled: !reduced,
      speed: 1,
      cardDraw: true,
      cardPlay: true,
      cardTap: true,
      cardUntap: true,
      cardFlip: true,
      particlesEtb: true,
      damagePopups: true,
      dice: true,
    },
    display: {
      theme: DEFAULT_SKIN,
      accent: "",
      cardSize: "medium",
      handLayout: "fan",
      // Row from #2433 (owner decision): opponents in one row along the
      // top, the viewer's board under them.
      tableLayout: DEFAULT_TABLE_LAYOUT,
      // v17 default: the pile (ADR 0119 §1, owner answer 1). It was
      // compact from v14, until the owner picked the pile.
      stackStyle: DEFAULT_STACK_STYLE,
      // Full from #2433 (owner decision). It was summary from v11,
      // because the full boards clipped at small panel sizes (#956);
      // #2336 sized boards by relevance and fitted every row to its
      // panel, so a full opponent board no longer clips or scrolls.
      opponentDetail: "full",
      // v11 default: on. Expanding the active player is the one
      // automatic expansion that cannot surprise you mid-click,
      // because it happens on a turn boundary.
      expandActivePlayer: true,
      // Overlay from #2433 (owner decision). It was reflow from v11,
      // which keeps combat arrows in one plane. Temporary; see the
      // field comment.
      expandStyle: "overlay",
      hoverDelayMs: 300,
      showOpponentHandCount: true,
      // v16 defaults (#2209): art tiles on the battlefield, full cards
      // in the hand.
      battlefieldArt: true,
      handArt: false,
      // ADR 0128: every player's playmat is drawn, until the player
      // turns them down.
      playmats: DEFAULT_PLAYMATS_MODE,
    },
    gameplay: {
      confirmExit: true,
      // S13 default: on. Pre-S13 this was off because the only
      // gating was "not on viewer's own turn", which felt too
      // aggressive. The S13 stops grid (defaultStepStops) gives
      // the user fine control, so auto-pass-on is now the right
      // default — stops are the affordance for "stop here".
      autoPassPriority: true,
      stepStops: defaultStepStops(),
      // ADR 0118 §1 default: on. A spell costs what it says, and a
      // click taps the lands for it. Off (the S15 default) is the
      // sandbox / paper-tracking posture, still a supported choice.
      strictMana: true,
      // #1530 default: off. The server orders a commuting batch itself.
      alwaysAskTriggerOrder: false,
      // S13.6 default: on. The step-stops grid is the intent
      // affordance; smartAutoPass lets it mean "stop if I
      // might want to respond" instead of "stop every time."
      smartAutoPass: true,
      // #1307 defaults: every response category counts, and an
      // opponent's spell you can't answer passes.
      respondCounterspells: true,
      respondInstants: true,
      respondAbilities: true,
      respondSpecialActions: true,
      alwaysStopOpponentStack: false,
      // #1307 bluff defaults: off, timed, 1.5–4 s. A bluff slows the
      // table, so nobody gets one they didn't ask for.
      bluffCounterspell: false,
      bluffInstant: false,
      bluffMode: "timed",
      bluffDelayMinMs: 1500,
      bluffDelayMaxMs: 4000,
      // ADR 0119 §2 default: about 2 s (owner answer 2a).
      stackHoldMs: 2000,
      // #323 default: ON. "I cast it" is already the decision; the
      // client shouldn't ask you to confirm it. Opponent items on
      // the stack still stop, and the in-game "hold" toggle is the
      // per-window opt-out.
      autoPassOwnStack: true,
      // S13.6 default: OFF — the autopass toggle clears on the
      // viewer's next precombat_main so a forgotten autopass
      // doesn't skip their turn. Opt-in is a DANGER setting.
      autopassPersistThroughTurns: false,
      // #170 default: OFF. Right-click keeps meaning "show this
      // permanent's abilities" until the player opts in to the
      // override menu.
      adminOverrides: false,
      // S31 default: OFF. The improvisation announcement is shown
      // regardless; this only adds the per-move narration, which is
      // a debug surface and a lot of lines.
      showBotReasoning: false,
      // ADR 0105 default: ON, for everyone (owner decision 4).
      highlightLegalActions: true,
    },
    shortcuts: {
      // v10 default: ON. The defaults are chosen not to collide with
      // a browser or OS reservation and the dispatcher refuses to
      // fire while a text field has focus or a modal is up, so
      // shipping the keymap live is what makes the game feel faster
      // for someone who never opens this panel. The switch exists
      // for the player who wants none of it.
      enabled: true,
      // No overrides until the player makes one. See the field
      // comment on Settings["shortcuts"]["bindings"].
      bindings: {},
    },
    accessibility: {
      reduceMotion: reduced,
      textScale: 1.0,
      colorblindPalette: false,
      alwaysShowFocus: false,
    },
    help: {
      seen: {},
      tipsOff: false,
    },
  };
}

// --- per-person and per-device fields (ADR 0110 §4, owner answer 5) ---
//
// A signed-in person's settings follow them to every browser they sign
// in on, except the ones that belong to the SCREEN or the machine. The
// split is decided here, field by field, and nowhere else:
//
//   "synced"  per person. Uploaded to the account (PUT /me/settings) and
//             applied from it at sign-in. How someone plays and what
//             they like the table to look like.
//   "device"  per device. Never leaves this browser. Volumes (speakers
//             versus headphones), anything that depends on screen size
//             (#956), and the two settings that follow the OS
//             reduced-motion signal at runtime.
//
// The type makes the map exhaustive: a field added to Settings without
// a line here is a compile error, and settingsSyncFields.test.ts fails
// too. Nobody can ship a setting without deciding where it lives.
//
// The practice table (practiceTable.ts) forces four fields. Two are
// synced (strictMana, autoPassPriority) and settingsSync.ts never
// uploads their forced values; two are per device (tableLayout,
// cardSize) and never upload at all.

export type SettingsGroup = Exclude<keyof Settings, "__version">;
export type FieldScope = "synced" | "device";
export type SettingsFieldScopes = {
  [G in SettingsGroup]: { [K in keyof Settings[G]]-?: FieldScope };
};

export const SYNCED_FIELDS: Readonly<SettingsFieldScopes> = Object.freeze({
  audio: {
    muted: "device",
    masterVolume: "device",
    effectsVolume: "device",
    musicVolume: "device",
  },
  animations: {
    // Follows the OS reduced-motion signal at runtime.
    enabled: "device",
    speed: "synced",
    cardDraw: "synced",
    cardPlay: "synced",
    cardTap: "synced",
    cardUntap: "synced",
    cardFlip: "synced",
    particlesEtb: "synced",
    damagePopups: "synced",
    dice: "synced",
  },
  display: {
    theme: "synced",
    accent: "synced",
    cardSize: "device",
    handLayout: "device",
    tableLayout: "device",
    stackStyle: "synced",
    opponentDetail: "device",
    expandActivePlayer: "device",
    expandStyle: "device",
    hoverDelayMs: "synced",
    showOpponentHandCount: "synced",
    // #1954 / #2209: a taste, not a screen size (ADR 0110 owner
    // answer 5).
    battlefieldArt: "synced",
    handArt: "synced",
    // ADR 0128: about this screen, not the person.
    playmats: "device",
  },
  gameplay: {
    confirmExit: "synced",
    autoPassPriority: "synced",
    stepStops: "synced",
    strictMana: "synced",
    alwaysAskTriggerOrder: "synced",
    smartAutoPass: "synced",
    respondCounterspells: "synced",
    respondInstants: "synced",
    respondAbilities: "synced",
    respondSpecialActions: "synced",
    alwaysStopOpponentStack: "synced",
    bluffCounterspell: "synced",
    bluffInstant: "synced",
    bluffMode: "synced",
    bluffDelayMinMs: "synced",
    bluffDelayMaxMs: "synced",
    stackHoldMs: "synced",
    autoPassOwnStack: "synced",
    autopassPersistThroughTurns: "synced",
    adminOverrides: "synced",
    showBotReasoning: "synced",
    highlightLegalActions: "synced",
  },
  shortcuts: {
    enabled: "synced",
    bindings: "synced",
  },
  accessibility: {
    // Follows the OS reduced-motion signal at runtime.
    reduceMotion: "device",
    textScale: "device",
    colorblindPalette: "synced",
    alwaysShowFocus: "synced",
  },
  // ADR 0125 §4: both per person.
  help: {
    seen: "synced",
    tipsOff: "synced",
  },
});

/** SyncedSettings is the account's copy: group → synced field → value. */
export type SyncedSettings = Record<string, Record<string, unknown>>;

/** syncedPaths lists every synced field as [group, key], in map order. */
export function syncedPaths(): Array<[SettingsGroup, string]> {
  const out: Array<[SettingsGroup, string]> = [];
  for (const group of Object.keys(SYNCED_FIELDS) as SettingsGroup[]) {
    const scopes = SYNCED_FIELDS[group] as Record<string, FieldScope>;
    for (const key of Object.keys(scopes)) {
      if (scopes[key] === "synced") out.push([group, key]);
    }
  }
  return out;
}

/**
 * syncedSubset is the part of s that goes to the account: every synced
 * field, deep-copied, grouped as in Settings. Nothing per device.
 */
export function syncedSubset(s: Settings): SyncedSettings {
  const out: SyncedSettings = {};
  for (const [group, key] of syncedPaths()) {
    const value = (s[group] as Record<string, unknown>)[key];
    if (value === undefined) continue;
    (out[group] ??= {})[key] = JSON.parse(JSON.stringify(value));
  }
  return out;
}

/** canonicalJSON is JSON with object keys sorted, for comparing copies. */
export function canonicalJSON(v: unknown): string {
  if (Array.isArray(v)) return `[${v.map(canonicalJSON).join(",")}]`;
  if (v !== null && typeof v === "object") {
    const o = v as Record<string, unknown>;
    return `{${Object.keys(o)
      .sort()
      .filter((k) => o[k] !== undefined)
      .map((k) => `${JSON.stringify(k)}:${canonicalJSON(o[k])}`)
      .join(",")}}`;
  }
  return JSON.stringify(v);
}

/**
 * applySyncedCopy returns base with its synced fields replaced by an
 * account copy written by client version `version`. The copy goes
 * through migrate, the one schema validator, so a copy from an older or
 * newer client, or a hand-edited one, comes out in this client's shape.
 * Per-device fields are base's, always.
 */
export function applySyncedCopy(base: Settings, copy: unknown, version: number): Settings {
  if (!copy || typeof copy !== "object" || Array.isArray(copy)) return base;
  const migrated = migrate({ ...(copy as object), __version: version });
  const next: Settings = {
    ...base,
    audio: { ...base.audio },
    animations: { ...base.animations },
    display: { ...base.display },
    gameplay: { ...base.gameplay },
    shortcuts: { ...base.shortcuts },
    accessibility: { ...base.accessibility },
    help: { ...base.help },
  };
  for (const [group, key] of syncedPaths()) {
    (next[group] as Record<string, unknown>)[key] = (migrated[group] as Record<string, unknown>)[
      key
    ];
  }
  return next;
}

// migrate normalises a stored settings blob into the current schema.
// Accepts anything that parses as JSON and does a shallow field-by-
// field merge against defaults, dropping unknown fields and filling
// in missing ones. Also absorbs legacy per-feature localStorage
// keys (currently just cmdctrl.muted) so pre-S11.5 preferences
// don't get silently discarded on first load after upgrade.
function migrate(raw: unknown): Settings {
  const d = defaultSettings();
  if (!raw || typeof raw !== "object") return absorbLegacy(d);

  const s = raw as Partial<Settings>;
  const merged: Settings = {
    __version: SETTINGS_VERSION,
    audio: { ...d.audio, ...(s.audio ?? {}) },
    animations: { ...d.animations, ...(s.animations ?? {}) },
    display: { ...d.display, ...(s.display ?? {}) },
    gameplay: { ...d.gameplay, ...(s.gameplay ?? {}) },
    shortcuts: { ...d.shortcuts, ...(s.shortcuts ?? {}) },
    accessibility: { ...d.accessibility, ...(s.accessibility ?? {}) },
    help: { ...d.help, ...(s.help ?? {}) },
  };
  // v1 → v2 (S13): the gameplay.stepStops map was scaffolded as `{}`
  // pre-S13. Seed defaults for any user whose stored map is empty so
  // the per-step stops UI has something meaningful on first paint.
  // Existing user-configured maps are preserved untouched. Strip any
  // entries for no-priority steps (Untap / Cleanup) to keep the map
  // canonical. While we're here, flip `autoPassPriority` to true if
  // the user is still on the v1 default (false) — at v1 the toggle
  // was a global "auto-pass on opponents' turns" with no per-step
  // control, so off was the only safe default. With stops, the
  // toggle gates the per-step auto-pass and on is the natural
  // default. Pre-existing v1 users who explicitly turned it on stay
  // on; users who left it off get the new behaviour. We can't
  // distinguish "user explicitly left it off" from "user never
  // touched the default" at v1, but the worst case is "auto-pass
  // through opponents' turns the user wasn't expecting" — which
  // their stops grid (also being seeded here) prevents.
  const storedVersion = typeof s.__version === "number" ? s.__version : 0;
  const fromV1 = storedVersion < 2;
  if (Object.keys(merged.gameplay.stepStops).length === 0) {
    merged.gameplay.stepStops = defaultStepStops();
    if (fromV1) {
      merged.gameplay.autoPassPriority = true;
    }
  } else {
    for (const id of NO_PRIORITY_STEPS) {
      delete merged.gameplay.stepStops[id];
    }
  }
  // v2 → v3 (S13 hotfix): early v2 builds (the previous client
  // commit) shipped autoPassPriority=false through the v1→v2
  // migration even after the new "default true" landed. Anyone who
  // already migrated to v2 with stepStops seeded still has the
  // toggle off and hits the "game halts at draw" trap. v3 flips
  // autoPassPriority on for users coming from v2 whose stops grid
  // matches the seeded default — strong signal they haven't tuned
  // either knob, so re-applying the new pairing is safe. Users who
  // customised stops keep their autoPassPriority value untouched.
  if (storedVersion === 2 && stepStopsMatchDefault(merged.gameplay.stepStops)) {
    merged.gameplay.autoPassPriority = true;
  }
  // v3 → v4 (S15): the gameplay.strictMana toggle is new. The
  // shallow merge above already populated it from defaults
  // (false) for any v3 blob that omits the field; nothing else
  // to do here — calling it out so future migrations have a
  // hook to extend.
  //
  // v4 → v5 (S13.6): gameplay.smartAutoPass. Same shape as the
  // v3→v4 migration — the shallow merge fills it from defaults
  // (true) for any v4 blob that omits the field. No user state
  // to rescue from the old world.
  //
  // v5 → v6 (S13.6): gameplay.autopassPersistThroughTurns. Defaults
  // to false (safe); existing v5 blobs inherit the safe default via
  // the shallow merge. The opt-in has a danger-warning banner in
  // the UI so anyone flipping it knows the risk.
  //
  // v6 → v7 (#170): gameplay.adminOverrides. Same shape again — the
  // shallow merge fills it from defaults (false) for any v6 blob
  // that omits the field. Nothing to rescue: the affordance it
  // replaces (Shift+click for a +1/+1 counter) had no stored state.
  //
  // v7 → v8 (#323): gameplay.autoPassOwnStack. Shallow merge fills
  // it from defaults (true) for any v7 blob that omits the field,
  // so existing users pick up the one-fewer-click behaviour without
  // touching their stops grid. Nothing to rescue — the behaviour it
  // replaces was hard-coded, not stored.
  //
  // v8 → v9 (S31 sub-PR 8): gameplay.showBotReasoning. Same shape
  // again — the shallow merge fills it from defaults (false) for any
  // v8 blob that omits the field. Nothing to rescue: bot seats did
  // not exist at v8, so no stored state can be about them. Note this
  // setting does NOT gate improvisation announcements, which are
  // mandatory disclosure and shown at every version.
  //
  // v9 → v10 (#565): the `shortcuts` section. Two things worth
  // spelling out, because neither is the usual "shallow merge fills
  // it from defaults" story this chain has told eight times.
  //
  // First, there IS pre-existing user state to respect, and it is not
  // in this blob. Settings.svelte has bound "," to the settings panel
  // since S11.5 and the command bar's gear button advertises it, so
  // "," is the shipped default for `openSettings` rather than a new
  // key we picked — nobody's muscle memory moves on upgrade.
  //
  // Second, and this is the discipline the rest of this chain is
  // careful about: we store OVERRIDES, never a materialised map. A
  // v9 user has no bindings at all, so they get today's defaults; a
  // v10 user who rebinds one row stores one key. When a default is
  // retuned later, everyone who never touched that row moves with it
  // and everyone who did keeps their choice — which is the same
  // "don't clobber an explicit decision" rule the v2→v3 stepStops
  // migration had to reconstruct from a structural comparison,
  // except here the storage shape makes it free.
  //
  // sanitizeOverrides drops unknown action ids, unparseable chords
  // and the reserved chords (Escape / Enter / Tab), and canonicalises
  // what survives. It also drops an override that has become equal to
  // the current default, so a row the user "changed" back to the
  // shipped value stops being pinned.
  //
  // v10 → v11 (#956 follow-up): display.opponentDetail,
  // display.expandActivePlayer and display.expandStyle. The shallow
  // merge fills all three from defaults for any v10 blob, which is
  // the usual story — but be clear about what that means here,
  // because it is not what the previous nine migrations did.
  //
  // opponentDetail defaults to "summary", so an existing player's
  // opponents CHANGE APPEARANCE on upgrade without them touching
  // anything. Every earlier migration in this chain changed how the
  // client behaved; this is the first that changes what the table
  // looks like. It is deliberate — the full-card rendering clips at
  // small panel sizes and has nowhere left to shrink (#956) — and
  // "full" restores the old look exactly, but a player who opens
  // Settings after upgrading is looking for this row, so it sits at
  // the top of the Display tab rather than the bottom.
  //
  // expandStyle is temporary and disappears with one of the two
  // expansion mechanisms; a stored value for it is expected to stop
  // being honoured, which is fine because the shallow merge will
  // simply drop an unknown field at v12.
  //
  // v11 → v12 (#1307): gameplay.respondCounterspells,
  // respondInstants, respondAbilities, respondSpecialActions (all
  // true) and alwaysStopOpponentStack (false). The shallow merge
  // fills them. Nothing is stored to rescue, but the upgrade does
  // change behaviour for a player who touched nothing: with smart
  // autopass on, an opponent's spell they cannot answer now passes
  // instead of stopping, and a mana ability no longer counts as a
  // response. alwaysStopOpponentStack puts the old stop back.
  //
  // v12 → v13 (#1307 bluffing): gameplay.bluffCounterspell,
  // bluffInstant (false), bluffMode ("timed"), bluffDelayMinMs (1500)
  // and bluffDelayMaxMs (4000). The shallow merge fills them, and
  // with both bluffs off nothing behaves differently. The delay
  // bounds are clamped where they are read (bluff.ts), so a
  // hand-edited blob cannot stall the table.
  //
  // v13 → v14 (#1467): display.stackStyle. The shallow merge fills it
  // from defaults ("compact"), so nobody's stack moves on upgrade —
  // the floating lanes are opt-in. Unlike the enums before it, the
  // value is also checked: this one picks which component the board
  // mounts, and an unknown string (a style that was tried and
  // removed, a hand-edited blob) must fall back to the docked card
  // rather than to no stack at all. (Since v17 that fallback is the
  // pile, the default; see below.)
  if (!isStackStyle(merged.display.stackStyle)) {
    merged.display.stackStyle = DEFAULT_STACK_STYLE;
  }
  // v14 → v15 (ADR 0105, #1789): gameplay.highlightLegalActions. Not
  // the usual shallow-merge fill: the owner decided the highlights
  // start ON for every player, existing ones included (ADR 0105 owner
  // decision 4), so the upgrade writes `true` whatever the stored blob
  // says — a value left over from a hand edit or a pre-release build
  // does not survive it. From v15 on, the stored choice is honoured.
  if (storedVersion < 15) {
    merged.gameplay.highlightLegalActions = true;
  } else if (typeof merged.gameplay.highlightLegalActions !== "boolean") {
    merged.gameplay.highlightLegalActions = true;
  }
  // v15 → v16 (#2209): display.artOnlyCards splits into
  // display.battlefieldArt (default true) and display.handArt (default
  // false). A stored blob is a MATERIALISED copy of every field, the
  // default included, so `artOnlyCards: false` cannot be told from
  // "never touched it" — and neither can an account copy, which is the
  // same subset written the same way (ADR 0110 §4). Only `true` is
  // certainly a choice: the option was off by default. So `true` keeps
  // art in both places, and anything else takes the new defaults —
  // which is what the owner asked for (owner answer 1, 2026-10-04): the
  // battlefield goes to art for everyone who had not opted in.
  const display = merged.display as Settings["display"] & { artOnlyCards?: unknown };
  if (storedVersion < 16) {
    const hadArt = display.artOnlyCards === true;
    display.battlefieldArt = hadArt || d.display.battlefieldArt;
    display.handArt = hadArt || d.display.handArt;
  }
  if (typeof display.battlefieldArt !== "boolean") {
    display.battlefieldArt = d.display.battlefieldArt;
  }
  if (typeof display.handArt !== "boolean") display.handArt = d.display.handArt;
  delete display.artOnlyCards;
  // v16 → v17 (ADR 0119 §1, #2204): the pile becomes the default. A
  // stored `compact` from before v17 becomes `pile`. `compact` was the
  // default, and the shallow merge above has always written defaults
  // into the stored blob, so an untouched `compact` and a chosen one
  // look the same; the v2 → v3 migration met the same problem and moved
  // the untouched case. A stored fan, spotlight or ribbon is kept:
  // nobody reaches those without choosing them. From v17 on, a stored
  // `compact` is honoured.
  if (storedVersion < 17 && merged.display.stackStyle === "compact") {
    merged.display.stackStyle = "pile";
  }
  // v17 → v18 (ADR 0119 §2, #2204): gameplay.stackHoldMs. The shallow
  // merge fills it from defaults (2000) for any v17 blob, so every
  // player gets the hold on upgrade, as the owner asked: an
  // opponent's spell stays up for about 2 s before auto-pass lets it
  // resolve. Nothing is stored to rescue. The value is clamped where
  // it is read (stackHold.ts), as the bluff bounds are.
  //
  // v18 → v19 (ADR 0118 §1, #2188): strict payment becomes the
  // default, and everyone is moved to it ONCE (owner decision 5). A
  // stored `false` is the old default materialised, so it cannot be
  // told from a choice; the v14 → v15 block met the same problem and
  // wrote the new value for everyone. An account copy goes through
  // this same migrate with the version that wrote it (applySyncedCopy),
  // so a synced `false` from a v18 client moves too. From v19 on, the
  // stored choice is honoured: a player who turns it off stays off.
  // (The ADR planned this as v16; v16–v18 went to #2209 and ADR 0119
  // first, so it is v19.)
  if (storedVersion < 19) {
    merged.gameplay.strictMana = true;
  } else if (typeof merged.gameplay.strictMana !== "boolean") {
    merged.gameplay.strictMana = d.gameplay.strictMana;
  }
  // v19 → v20 (player skins): display.theme now names a skin. Until
  // v20 the Theme select was disabled, so a stored value is almost
  // always the old default "dark", which is not a skin: it falls back
  // to the default skin, as does any unknown value (a removed skin, a
  // hand edit). A stored "light" or "high-contrast" is a skin and is
  // kept. display.accent is new; anything but #rrggbb is "".
  if (!isSkin(merged.display.theme)) {
    merged.display.theme = DEFAULT_SKIN;
  }
  merged.display.accent = normalizeAccent(merged.display.accent);
  // #2336: display.tableLayout gained "focus". A new value of an
  // existing field needs no version bump, but the value is checked
  // from here on: an unknown string (a hand edit, a layout that was
  // tried and removed) falls back to the default (the row, #2433) rather than to a
  // board with no grid template.
  if (!isTableLayout(merged.display.tableLayout)) {
    merged.display.tableLayout = DEFAULT_TABLE_LAYOUT;
  }
  // ADR 0128: display.playmats is new. The shallow merge fills it with
  // "all" for any older blob, and a value that is not one of the three
  // (a hand edit) falls back to the default rather than to a board that
  // draws nothing for a reason nobody can see. No version bump: it is
  // per device, so no account copy carries it.
  if (!isPlaymatsMode(merged.display.playmats)) {
    merged.display.playmats = DEFAULT_PLAYMATS_MODE;
  }
  merged.shortcuts = {
    enabled: merged.shortcuts?.enabled !== false,
    bindings: sanitizeOverrides(merged.shortcuts?.bindings),
  };
  // v20 → v21 (ADR 0125 §4): the `help` group. The shallow merge fills
  // it from defaults (nothing seen, tips on) for any v20 blob, so every
  // existing player is offered each hint once. The map is checked, not
  // trusted: a hand-edited or hostile blob keeps only string ids with a
  // positive integer version, and retired ids are dropped. Unknown ids
  // are kept, because adding a hint does not bump SETTINGS_VERSION and
  // an older tab must not drop a newer client's hints on its next write.
  merged.help = {
    seen: normalizeSeen(merged.help?.seen),
    tipsOff: merged.help?.tipsOff === true,
  };
  return absorbLegacy(merged);
}

// stepStopsMatchDefault reports whether the supplied stepStops map
// is structurally identical to defaultStepStops(). Used by the v2→v3
// migration to detect "user hasn't customised stops" so we can
// safely re-seed autoPassPriority without overwriting an explicit
// off-toggle.
function stepStopsMatchDefault(actual: Record<string, boolean>): boolean {
  const expected = defaultStepStops();
  const actualKeys = Object.keys(actual);
  const expectedKeys = Object.keys(expected);
  if (actualKeys.length !== expectedKeys.length) return false;
  for (const k of expectedKeys) {
    if (actual[k] !== expected[k]) return false;
  }
  return true;
}

// absorbLegacy folds pre-S11.5 single-key localStorage flags into
// the unified schema, then clears the legacy keys so migration is
// one-shot. Callers that still read the old keys (e.g. sounds.ts
// isMuted()) keep working; after this migration their next read
// picks up the unified value through the wrapper.
function absorbLegacy(s: Settings): Settings {
  if (typeof localStorage === "undefined") return s;
  const legacyMuted = localStorage.getItem(LEGACY_MUTED_KEY);
  if (legacyMuted !== null) {
    s.audio.muted = legacyMuted === "1";
    localStorage.removeItem(LEGACY_MUTED_KEY);
  }
  return s;
}

function loadSettings(): Settings {
  if (typeof localStorage === "undefined") return defaultSettings();
  const raw = localStorage.getItem(STORAGE_KEY);
  if (!raw) return absorbLegacy(defaultSettings());
  try {
    return migrate(JSON.parse(raw));
  } catch {
    // Corrupted blob — fall back to defaults. We don't want to
    // strand a user in an un-openable settings panel because of a
    // bad write.
    return defaultSettings();
  }
}

function saveSettings(s: Settings): void {
  if (typeof localStorage === "undefined") return;
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(s));
  } catch {
    // QuotaExceeded or Safari-private-mode — swallow. The live
    // store still has the update; persistence is best-effort.
  }
}

export const settings: Writable<Settings> = guardedWritable(loadSettings(), "settings");

// Persist on every mutation. Subscribe rather than wrapping every
// setter because .update() / .set() fire here too.
settings.subscribe((s) => saveSettings(s));

// Bridge audio.muted into sounds.ts. sounds.ts keeps its own cache
// for cheap isMuted() reads from non-reactive call sites; this
// bridge keeps that cache in sync whenever the settings value
// changes — including the migration-time absorption of the legacy
// cmdctrl.muted key.
let lastMuted: boolean | null = null;
let lastVolume: number | null = null;
let lastMusicVolume: number | null = null;
settings.subscribe((s) => {
  if (s.audio.muted !== lastMuted) {
    lastMuted = s.audio.muted;
    setMuted(s.audio.muted);
    setMusicMuted(s.audio.muted);
  }
  // Master × effects, both 0..100, normalised to a 0..1
  // multiplier. sounds.play() folds this into per-event volume.
  const vol = (s.audio.masterVolume / 100) * (s.audio.effectsVolume / 100);
  if (vol !== lastVolume) {
    lastVolume = vol;
    setVolumeMultiplier(vol);
  }
  // Same shape for music: master × music feeds the ambient track's
  // HTMLAudioElement.volume. music.ts pauses the element when the
  // product is 0, so "master=0" or "music=0" stops decoding rather
  // than silently streaming.
  const musicVol = (s.audio.masterVolume / 100) * (s.audio.musicVolume / 100);
  if (musicVol !== lastMusicVolume) {
    lastMusicVolume = musicVol;
    setMusicVolumeMultiplier(musicVol);
  }
});

// Bridge animations.* into animations.ts. Pushes the entire
// animations group on every change rather than diffing because
// (a) the payload is small (<10 booleans + a number), (b) the
// receiver does shallow Object.assign, and (c) animations are
// next-tick-bound so a redundant push has no observable effect.
settings.subscribe((s) => {
  setAnimationConfig({
    enabled: s.animations.enabled,
    speed: s.animations.speed,
    cardDraw: s.animations.cardDraw,
    cardPlay: s.animations.cardPlay,
    cardTap: s.animations.cardTap,
    cardUntap: s.animations.cardUntap,
    cardFlip: s.animations.cardFlip,
    particlesEtb: s.animations.particlesEtb,
    damagePopups: s.animations.damagePopups,
    dice: s.animations.dice,
  });
});

// Observe the OS reduced-motion hint and propagate flips to the
// store when the user has not explicitly overridden. "Explicitly
// overridden" is tracked by comparing the current value to what
// prefersReducedMotion would return — if they match, we keep
// tracking; if they diverge, the user has opted out of OS sync.
// Installed module-scope so it outlives any single component's
// lifecycle.
if (typeof window !== "undefined" && window.matchMedia) {
  const mq = window.matchMedia("(prefers-reduced-motion: reduce)");
  const onChange = (e: MediaQueryListEvent) => {
    const s = get(settings);
    // Only follow OS flips while the in-app toggle still matches
    // the pre-flip OS value. This heuristic means a user who
    // manually toggled the setting stays in charge.
    if (s.accessibility.reduceMotion !== e.matches) {
      // They're different — either (a) we installed after a manual
      // override, (b) the user flipped OS without touching our
      // setting. In either case, sync towards OS so a user who
      // turns on low-motion at the OS level gets it here too.
      settings.update((prev) => ({
        ...prev,
        accessibility: { ...prev.accessibility, reduceMotion: e.matches },
        animations: { ...prev.animations, enabled: !e.matches },
      }));
    }
  };
  // addEventListener is the modern API; Safari < 14 ships
  // addListener instead. Fall back once.
  if (mq.addEventListener) mq.addEventListener("change", onChange);
  else mq.addListener?.(onChange);
}

// updateSettings mutates a single path in the settings tree. Saves
// a few keystrokes at every call-site over the equivalent
// settings.update(prev => ({ ...prev, group: { ...prev.group, key: v } })).
export function updateSettings<G extends keyof Omit<Settings, "__version">>(
  group: G,
  key: keyof Settings[G],
  value: Settings[G][keyof Settings[G]],
): void {
  settings.update((prev) => ({
    ...prev,
    [group]: { ...prev[group], [key]: value },
  }));
}

// resetSettings restores defaults. Used by the "Reset all" button
// in the Advanced tab. Keeps the __version bump so a reset after a
// schema migration doesn't revert to the old shape.
export function resetSettings(): void {
  settings.set(defaultSettings());
}

// settingsOpen drives the modal visibility. Exported so any surface
// (header gear, keyboard shortcut, chat slash-command later) can
// toggle it without the modal needing prop drilling. Settings.svelte
// subscribes and renders itself when this is true.
export const settingsOpen: Writable<boolean> = guardedWritable(false, "settingsOpen");

// SettingsTab names the panel's sidebar entries. Lives here rather
// than in Settings.svelte so a caller can ask for a specific tab
// without importing the component.
export type SettingsTab =
  | "audio"
  | "animations"
  | "display"
  | "playmat"
  | "gameplay"
  | "shortcuts"
  | "accessibility"
  | "advanced";

// settingsTab is which tab the panel shows on its next open. The
// panel resets this to "audio" when it closes, so "open settings"
// with no argument keeps landing on the first tab the way it always
// has; the shortcuts overlay's "Customise" link is what needs the
// deep link.
export const settingsTab: Writable<SettingsTab> = guardedWritable("audio", "settingsTab");

export function openSettings(tab: SettingsTab = "audio"): void {
  settingsTab.set(tab);
  settingsOpen.set(true);
}

export function closeSettings(): void {
  settingsOpen.set(false);
}

// exportSettings serialises the current settings to a JSON string
// suitable for copying to the clipboard. The Advanced tab uses this
// for the "copy my settings" button; a user on a second device can
// then paste the blob into importSettings to carry their prefs
// across without a server-side sync layer.
export function exportSettings(): string {
  return JSON.stringify(get(settings), null, 2);
}

// ImportResult reports what happened. `changed` signals whether the
// store actually updated — a valid no-op import (JSON matches
// current state) returns ok: true, changed: false so callers can
// flash an appropriate confirmation.
export interface ImportResult {
  ok: boolean;
  error?: string;
  changed?: boolean;
}

// importSettings validates the blob, merges via the same migration
// path used at module load, and writes to the store. Rejects on
// parse failure or on blobs that aren't a plain object. Schema
// mismatches are tolerated — missing fields fall back to defaults
// via migrate(), unknown fields are dropped.
export function importSettings(raw: string): ImportResult {
  let parsed: unknown;
  try {
    parsed = JSON.parse(raw);
  } catch (e) {
    return { ok: false, error: (e as Error).message };
  }
  if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
    return { ok: false, error: "settings JSON must be an object" };
  }
  const before = JSON.stringify(get(settings));
  const next = migrate(parsed);
  const after = JSON.stringify(next);
  settings.set(next);
  return { ok: true, changed: before !== after };
}

// fingerprintSettings returns a short alphanumeric hash of the
// current settings for bug reports. Not cryptographic — FNV-1a
// 32-bit keeps the implementation small and dependency-free, which
// is enough to distinguish "same prefs" from "different prefs" when
// comparing two players' reports.
export function fingerprintSettings(): string {
  const str = JSON.stringify(get(settings));
  let h = 0x811c9dc5;
  for (let i = 0; i < str.length; i++) {
    h ^= str.charCodeAt(i);
    h = (h + ((h << 1) + (h << 4) + (h << 7) + (h << 8) + (h << 24))) >>> 0;
  }
  return h.toString(36).padStart(7, "0");
}
