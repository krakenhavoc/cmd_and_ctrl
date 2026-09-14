import { describe, it, expect } from "vitest";
import {
  SHORTCUTS,
  SHORTCUT_IDS,
  GROUP_ORDER,
  GROUP_LABELS,
  RESERVED_CHORDS,
  bindingConflicts,
  canonicalKey,
  chordDefersToFocus,
  chordFrom,
  conflictingWith,
  defaultBindings,
  dispatchShortcut,
  effectiveBindings,
  evaluateBinding,
  formatChord,
  idleContext,
  elementOf,
  isInteractiveFocus,
  isReservedChord,
  isTypingTarget,
  keyOfChord,
  parseChord,
  sanitizeOverrides,
  shortcutAvailability,
  shortcutDef,
  type BindingMap,
  type DispatchInput,
  type ShortcutContext,
  type ShortcutID,
} from "./shortcuts";

// A context in which every game action is live: seated, acting, at a
// table, with priority and a wide board. Tests narrow from here.
function liveContext(over: Partial<ShortcutContext> = {}): ShortcutContext {
  return {
    ...idleContext(),
    atTable: true,
    activePlayer: true,
    passLegal: true,
    undosRemaining: 1,
    attackAllEligible: 3,
    attackAllDefenders: 1,
    ...over,
  };
}

function input(over: Partial<DispatchInput> = {}): DispatchInput {
  return {
    chord: "Space",
    repeat: false,
    typing: false,
    modalOpen: false,
    interactiveFocus: false,
    enabled: true,
    bindings: defaultBindings(),
    ctx: liveContext(),
    ...over,
  };
}

describe("the shortcut table", () => {
  it("has a unique id per row and no duplicate default bindings", () => {
    expect(new Set(SHORTCUT_IDS).size).toBe(SHORTCUTS.length);
    expect(bindingConflicts(defaultBindings()).size).toBe(0);
  });

  it("ships no default on a reserved chord", () => {
    for (const def of SHORTCUTS) {
      expect(isReservedChord(def.defaultBinding)).toBe(false);
    }
  });

  it("ships no default that collides with a browser or OS reservation", () => {
    // Ctrl/Cmd + these is close-tab, new-tab, reload, new-window and
    // find. Nothing in the keymap may sit on one, in either modifier.
    const forbidden = ["w", "t", "r", "n", "f", "p", "s", "q", "l"];
    for (const def of SHORTCUTS) {
      const parsed = parseChord(def.defaultBinding);
      if (!parsed) continue;
      const key = keyOfChord(parsed);
      const hasMod = parsed !== key && !parsed.startsWith("Shift+");
      if (hasMod) expect(forbidden).not.toContain(key);
      // Alt defaults are out entirely: AltGr is Ctrl+Alt on several
      // European layouts, so an Alt binding eats a letter those
      // players need in order to type.
      expect(parsed.includes("Alt+")).toBe(false);
      // No function keys — they belong to the browser and the OS.
      expect(/^F\d+$/.test(key)).toBe(false);
    }
  });

  it("gives every row a group that the overlay renders", () => {
    for (const def of SHORTCUTS) {
      expect(GROUP_ORDER).toContain(def.group);
      expect(GROUP_LABELS[def.group]).toBeTruthy();
    }
  });

  it("keeps the pre-existing comma binding for settings", () => {
    // Settings.svelte has bound "," since S11.5 and the gear button
    // advertises it in its title. Moving it would break muscle memory
    // for the one shortcut the client already had.
    expect(shortcutDef("openSettings")?.defaultBinding).toBe(",");
  });
});

describe("chordFrom", () => {
  it("lowercases letters and names the space bar", () => {
    expect(chordFrom({ key: "A" })).toBe("a");
    expect(chordFrom({ key: "a" })).toBe("a");
    expect(chordFrom({ key: " " })).toBe("Space");
  });

  it("records Shift for letters so Shift+P is distinct from P", () => {
    expect(chordFrom({ key: "P", shiftKey: true })).toBe("Shift+p");
    expect(chordFrom({ key: "p" })).toBe("p");
  });

  it("drops Shift for punctuation, because the character already encodes it", () => {
    // US layout: ? is Shift+/. German: Shift+ß. Some layouts: no
    // shift at all. Taking the produced character at face value is
    // what makes one stored chord work on all three.
    expect(chordFrom({ key: "?", shiftKey: true })).toBe("?");
    expect(chordFrom({ key: "?" })).toBe("?");
    expect(chordFrom({ key: ",", shiftKey: false })).toBe(",");
  });

  it("records Shift for named keys", () => {
    expect(chordFrom({ key: "Tab", shiftKey: true })).toBe("Shift+Tab");
    expect(chordFrom({ key: " ", shiftKey: true })).toBe("Shift+Space");
  });

  it("orders modifiers Ctrl, Alt, Shift, Meta", () => {
    expect(
      chordFrom({ key: "k", ctrlKey: true, altKey: true, shiftKey: true, metaKey: true }),
    ).toBe("Ctrl+Alt+Shift+Meta+k");
    // Same press, whatever order the flags are read in.
    expect(chordFrom({ key: "k", metaKey: true, ctrlKey: true })).toBe("Ctrl+Meta+k");
  });

  it("returns null for a bare modifier, a dead key, and an IME composition", () => {
    expect(chordFrom({ key: "Shift", shiftKey: true })).toBeNull();
    expect(chordFrom({ key: "Control", ctrlKey: true })).toBeNull();
    expect(chordFrom({ key: "Meta", metaKey: true })).toBeNull();
    expect(chordFrom({ key: "Dead" })).toBeNull();
    expect(chordFrom({ key: "Unidentified" })).toBeNull();
    // Mid-composition presses belong to the IME, not to us.
    expect(chordFrom({ key: "a", isComposing: true })).toBeNull();
  });

  it("canonicalKey folds the legacy key names old browsers still emit", () => {
    expect(canonicalKey("Spacebar")).toBe("Space");
    expect(canonicalKey("Esc")).toBe("Escape");
    expect(canonicalKey("Up")).toBe("ArrowUp");
    expect(canonicalKey("PageDown")).toBe("PageDown");
  });
});

describe("parseChord", () => {
  it("round-trips everything chordFrom can produce", () => {
    const samples: Array<Parameters<typeof chordFrom>[0]> = [
      { key: "a" },
      { key: "?" },
      { key: " " },
      { key: "P", shiftKey: true },
      { key: "z", ctrlKey: true },
      { key: "z", metaKey: true },
      { key: "ArrowUp" },
      { key: "+" },
    ];
    for (const s of samples) {
      const chord = chordFrom(s)!;
      expect(parseChord(chord)).toBe(chord);
    }
  });

  it("normalises modifier case and ordering from a hand-edited blob", () => {
    expect(parseChord("meta+CTRL+z")).toBe("Ctrl+Meta+z");
    expect(parseChord("shift+P")).toBe("Shift+p");
  });

  it("strips a Shift that punctuation cannot carry", () => {
    // chordFrom can never produce this, so leaving it would install a
    // binding no keypress can ever match.
    expect(parseChord("Shift+?")).toBe("?");
  });

  it("rejects malformed input", () => {
    expect(parseChord("")).toBeNull();
    expect(parseChord("   ")).toBeNull();
    expect(parseChord("Hyper+k")).toBeNull();
    expect(parseChord("Ctrl+Ctrl+k")).toBeNull();
    expect(parseChord("Shift")).toBeNull();
    expect(parseChord("Ctrl+Meta")).toBeNull();
  });

  it("keeps a literal plus key addressable", () => {
    expect(parseChord("+")).toBe("+");
    expect(parseChord("Ctrl++")).toBe("Ctrl++");
    expect(keyOfChord("Ctrl++")).toBe("+");
  });
});

describe("formatChord", () => {
  it("upper-cases letters and spells named keys out", () => {
    expect(formatChord("a")).toBe("A");
    expect(formatChord("Shift+p")).toBe("Shift+P");
    expect(formatChord("Space")).toBe("Space");
    expect(formatChord("?")).toBe("?");
  });

  it("uses the Mac glyphs on a Mac", () => {
    expect(formatChord("Ctrl+Meta+z", true)).toBe("⌃⌘Z");
    expect(formatChord("Shift+p", true)).toBe("⇧P");
  });

  it("renders nothing for an unbound or invalid chord", () => {
    expect(formatChord("")).toBe("");
    expect(formatChord("Hyper+q")).toBe("");
  });
});

describe("overrides and conflicts", () => {
  it("effectiveBindings layers overrides over today's defaults", () => {
    const map = effectiveBindings({ undo: "z" });
    expect(map.undo).toBe("z");
    expect(map.passPriority).toBe(defaultBindings().passPriority);
  });

  it("stores only what the user changed, so a retuned default still reaches them", () => {
    // The whole reason overrides are persisted instead of a full map.
    const stored = sanitizeOverrides({ undo: "z", passPriority: defaultBindings().passPriority });
    expect(stored).toEqual({ undo: "z" });
    expect(Object.keys(stored)).not.toContain("passPriority");
  });

  it("keeps an explicit unbind, which is not the same as an absent row", () => {
    const stored = sanitizeOverrides({ drawCard: "" });
    expect(stored.drawCard).toBe("");
    expect(effectiveBindings(stored).drawCard).toBe("");
  });

  it("drops unknown ids, unparseable chords and reserved chords", () => {
    const stored = sanitizeOverrides({
      undo: "z",
      notAnAction: "q",
      passTurn: "Hyper+k",
      attackAll: "Escape",
      drawCard: 7 as unknown as string,
    });
    expect(stored).toEqual({ undo: "z" });
  });

  it("survives a hostile or absent blob", () => {
    expect(sanitizeOverrides(null)).toEqual({});
    expect(sanitizeOverrides("nope")).toEqual({});
    expect(sanitizeOverrides([1, 2, 3])).toEqual({});
    expect(sanitizeOverrides(undefined)).toEqual({});
  });

  it("canonicalises what it keeps", () => {
    expect(sanitizeOverrides({ undo: "CTRL+Z" })).toEqual({ undo: "Ctrl+z" });
  });

  it("reports two actions sharing one chord instead of silently dropping one", () => {
    // Silently deleting a conflicting row at load would make the
    // Settings panel lie about what it stored. The user has to see it.
    const stored = sanitizeOverrides({ undo: "l" });
    const map = effectiveBindings(stored);
    const conflicts = bindingConflicts(map);
    expect(conflicts.get("l")).toEqual(["undo", "toggleGameLog"]);
  });

  it("conflictingWith ignores the row being edited and unbound rows", () => {
    const map = effectiveBindings({});
    expect(conflictingWith(map, "undo", map.undo)).toEqual([]);
    expect(conflictingWith(map, "undo", map.toggleGameLog)).toEqual(["toggleGameLog"]);
    expect(conflictingWith(map, "undo", "")).toEqual([]);
  });

  it("evaluateBinding refuses the reserved chords and reports collisions", () => {
    const map = effectiveBindings({});
    expect(evaluateBinding(map, "undo", { key: "Escape" }).problem).toBe("reserved");
    expect(evaluateBinding(map, "undo", { key: "Enter" }).problem).toBe("reserved");
    expect(evaluateBinding(map, "undo", { key: "Tab", shiftKey: true }).problem).toBe("reserved");
    expect(evaluateBinding(map, "undo", { key: "Shift", shiftKey: true }).problem).toBe("invalid");

    const collide = evaluateBinding(map, "undo", { key: "l" });
    expect(collide.problem).toBeUndefined();
    expect(collide.chord).toBe("l");
    expect(collide.conflicts).toEqual(["toggleGameLog"]);
  });

  it("every reserved chord parses, so the refusal can never be bypassed by case", () => {
    for (const c of RESERVED_CHORDS) {
      expect(parseChord(c)).toBe(c);
    }
  });
});

describe("availability", () => {
  it("gates game actions on being at a table", () => {
    const ctx = idleContext();
    expect(shortcutAvailability("passPriority", ctx)).toEqual({
      enabled: false,
      reason: "only at a game table",
    });
    // Global rows keep working from the lobby and the login screen.
    expect(shortcutAvailability("openSettings", ctx).enabled).toBe(true);
    expect(shortcutAvailability("toggleHelp", ctx).enabled).toBe(true);
    expect(shortcutAvailability("toggleMute", ctx).enabled).toBe(true);
  });

  it("takes pass priority from the server's move list, not from a re-derivation", () => {
    expect(shortcutAvailability("passPriority", liveContext({ passLegal: true })).enabled).toBe(
      true,
    );
    const denied = shortcutAvailability("passPriority", liveContext({ passLegal: false }));
    expect(denied.enabled).toBe(false);
    expect(denied.reason).toBe("you don't hold priority");
  });

  it("stays permissive when the frame carried no move list at all", () => {
    // undefined is "no information", not "nothing is legal". A false
    // no costs the player a window they were entitled to; a false yes
    // costs a rejected click. Same asymmetry timing.ts documents.
    expect(
      shortcutAvailability("passPriority", liveContext({ passLegal: undefined })).enabled,
    ).toBe(true);
  });

  it("withholds every action-kind row from a spectator but leaves the view rows", () => {
    const ctx = liveContext({ spectator: true });
    expect(shortcutAvailability("passPriority", ctx).reason).toBe("spectating — read-only");
    expect(shortcutAvailability("undo", ctx).enabled).toBe(false);
    expect(shortcutAvailability("drawCard", ctx).enabled).toBe(false);
    expect(shortcutAvailability("toggleGameLog", ctx).enabled).toBe(true);
    expect(shortcutAvailability("toggleMute", ctx).enabled).toBe(true);
  });

  it("withholds live actions while the dev scrubber shows a past frame", () => {
    // Every control in the command bar is withheld while replaying,
    // for the same reason: they mutate the LIVE game.
    const ctx = liveContext({ replaying: true });
    expect(shortcutAvailability("passPriority", ctx).reason).toBe("viewing a past frame");
    expect(shortcutAvailability("toggleGameLog", ctx).enabled).toBe(true);
  });

  it("withholds actions after elimination and after the game ends", () => {
    expect(shortcutAvailability("passTurn", liveContext({ eliminated: true })).enabled).toBe(false);
    expect(shortcutAvailability("passTurn", liveContext({ gameOver: true })).enabled).toBe(false);
    // Watching the log of a game you just lost is the whole point of
    // still being in the tab.
    expect(shortcutAvailability("toggleGameLog", liveContext({ gameOver: true })).enabled).toBe(
      true,
    );
  });

  it("gates pass turn on being the active player", () => {
    expect(shortcutAvailability("passTurn", liveContext({ activePlayer: false })).reason).toBe(
      "not your turn",
    );
  });

  it("gates undo on the per-turn budget, with admin unlimited", () => {
    expect(shortcutAvailability("undo", liveContext({ undosRemaining: 0 })).reason).toBe(
      "no undos left this turn",
    );
    expect(shortcutAvailability("undo", liveContext({ undosRemaining: null })).enabled).toBe(true);
  });

  it("refuses to guess which opponent an attack-with-all hits", () => {
    expect(shortcutAvailability("attackAll", liveContext({ attackAllEligible: 0 })).reason).toBe(
      "nothing can attack right now",
    );
    expect(shortcutAvailability("attackAll", liveContext({ attackAllDefenders: 0 })).reason).toBe(
      "no opponent to attack",
    );
    const many = shortcutAvailability("attackAll", liveContext({ attackAllDefenders: 3 }));
    expect(many.enabled).toBe(false);
    expect(many.reason).toMatch(/more than one opponent/);
    expect(shortcutAvailability("attackAll", liveContext({ attackAllDefenders: 1 })).enabled).toBe(
      true,
    );
  });
});

describe("dispatchShortcut", () => {
  it("fires the action bound to the chord", () => {
    expect(dispatchShortcut(input({ chord: "Space" })).action).toBe("passPriority");
    expect(dispatchShortcut(input({ chord: "u" })).action).toBe("undo");
    expect(dispatchShortcut(input({ chord: "," })).action).toBe("openSettings");
  });

  it("never fires while the user is typing", () => {
    // The single most common way a shortcut feature becomes
    // infuriating. Checked before the binding lookup and before the
    // master toggle is even consulted.
    for (const chord of ["Space", "u", "a", ",", "?"]) {
      expect(dispatchShortcut(input({ chord, typing: true }))).toEqual({
        action: null,
        skipped: "typing",
      });
    }
  });

  it("stands down entirely while a modal is open", () => {
    // Includes the keys the modal does not itself handle: passing
    // priority through an unanswered targeting prompt is exactly the
    // bug a "check a list of open flags" approach ships.
    expect(dispatchShortcut(input({ chord: "Space", modalOpen: true })).skipped).toBe("modal");
    expect(dispatchShortcut(input({ chord: ",", modalOpen: true })).skipped).toBe("modal");
  });

  it("ranks typing above modal above the master switch", () => {
    expect(dispatchShortcut(input({ typing: true, modalOpen: true, enabled: false })).skipped).toBe(
      "typing",
    );
    expect(dispatchShortcut(input({ modalOpen: true, enabled: false })).skipped).toBe("modal");
    expect(dispatchShortcut(input({ enabled: false })).skipped).toBe("disabled");
  });

  it("ignores key auto-repeat", () => {
    // Holding Space passes one priority window, not thirty.
    expect(dispatchShortcut(input({ chord: "Space", repeat: true })).skipped).toBe("repeat");
  });

  it("ignores a bare modifier press", () => {
    expect(dispatchShortcut(input({ chord: null })).skipped).toBe("no-chord");
  });

  it("yields Space to a focused control so keyboard navigation still works", () => {
    // A global Space that preventDefault()ed unconditionally would
    // make every button in the game unusable from the keyboard.
    expect(dispatchShortcut(input({ chord: "Space", interactiveFocus: true })).skipped).toBe(
      "focus-defer",
    );
    // …but a letter binding has nothing to defer to.
    expect(dispatchShortcut(input({ chord: "u", interactiveFocus: true })).action).toBe("undo");
  });

  it("follows the deferral rule to wherever the user moves Space", () => {
    const bindings: BindingMap = {
      ...defaultBindings(),
      passPriority: "g",
      toggleGameLog: "Space",
    };
    expect(dispatchShortcut(input({ chord: "g", bindings, interactiveFocus: true })).action).toBe(
      "passPriority",
    );
    expect(
      dispatchShortcut(input({ chord: "Space", bindings, interactiveFocus: true })).skipped,
    ).toBe("focus-defer");
  });

  it("does not defer a modified Space, which activates nothing", () => {
    const bindings: BindingMap = { ...defaultBindings(), drawCard: "Ctrl+Space" };
    expect(chordDefersToFocus("Ctrl+Space")).toBe(false);
    expect(
      dispatchShortcut(input({ chord: "Ctrl+Space", bindings, interactiveFocus: true })).action,
    ).toBe("drawCard");
  });

  it("reports an unavailable action with its reason rather than swallowing the press", () => {
    const res = dispatchShortcut(input({ chord: "t", ctx: liveContext({ activePlayer: false }) }));
    expect(res.action).toBeNull();
    expect(res.skipped).toBe("unavailable");
    expect(res.availability?.reason).toBe("not your turn");
  });

  it("does nothing for an unbound chord", () => {
    expect(dispatchShortcut(input({ chord: "j" })).skipped).toBe("unbound");
  });

  it("honours an explicit unbind", () => {
    const bindings: BindingMap = { ...defaultBindings(), drawCard: "" };
    expect(dispatchShortcut(input({ chord: "d", bindings })).skipped).toBe("unbound");
    // An unbound row must not swallow the empty chord either.
    expect(dispatchShortcut(input({ chord: "", bindings })).skipped).toBe("no-chord");
  });

  it("breaks a conflict in declaration order, the same way every time", () => {
    const bindings: BindingMap = { ...defaultBindings(), undo: "l" };
    const first = dispatchShortcut(input({ chord: "l", bindings })).action;
    const second = dispatchShortcut(input({ chord: "l", bindings })).action;
    expect(first).toBe(second);
    // Asserted against the table's own order rather than a hard-coded
    // winner, so reordering SHORTCUTS is caught here rather than
    // silently changing which action a conflicting key fires.
    const expected = SHORTCUT_IDS.find((id: ShortcutID) => bindings[id] === "l");
    expect(first).toBe(expected);
    expect(["undo", "toggleGameLog"]).toContain(first);
  });

  it("keeps the global rows alive from the lobby and the table alike", () => {
    const lobby = input({ chord: ",", ctx: idleContext() });
    expect(dispatchShortcut(lobby).action).toBe("openSettings");
    expect(dispatchShortcut(input({ chord: "?", ctx: idleContext() })).action).toBe("toggleHelp");
    expect(dispatchShortcut(input({ chord: "Space", ctx: idleContext() })).skipped).toBe(
      "unavailable",
    );
  });
});

describe("DOM predicates", () => {
  it("treats every form control as typing, whatever its type", () => {
    // A focused checkbox uses Space, a number spinner uses the arrow
    // keys. Handing either to the keymap breaks a control the player
    // is actively using, so the tag is enough — the type is not
    // consulted.
    expect(isTypingTarget({ tagName: "INPUT" })).toBe(true);
    expect(isTypingTarget({ tagName: "textarea" })).toBe(true);
    expect(isTypingTarget({ tagName: "SELECT" })).toBe(true);
    expect(isTypingTarget({ tagName: "DIV", isContentEditable: true })).toBe(true);
    expect(isTypingTarget({ tagName: "BUTTON" })).toBe(false);
    expect(isTypingTarget({ tagName: "BODY" })).toBe(false);
    expect(isTypingTarget(null)).toBe(false);
  });

  it("elementOf narrows an EventTarget and rejects the window", () => {
    // A keypress can be targeted at the window or the document, and
    // neither has a tagName to read.
    expect(elementOf(null)).toBeNull();
    expect(elementOf(undefined)).toBeNull();
    expect(elementOf({ addEventListener() {} })).toBeNull();
    expect(elementOf("body")).toBeNull();
    expect(elementOf({ tagName: "INPUT" })).toEqual({ tagName: "INPUT" });
  });

  it("catches a keypress deep inside a contenteditable subtree", () => {
    const span = {
      tagName: "SPAN",
      closest: (sel: string) => (sel.includes("contenteditable") ? {} : null),
    };
    expect(isTypingTarget(span)).toBe(true);
  });

  it("catches an ARIA textbox", () => {
    expect(isTypingTarget({ tagName: "DIV", getAttribute: () => "textbox" })).toBe(true);
    expect(isTypingTarget({ tagName: "DIV", getAttribute: () => "searchbox" })).toBe(true);
  });

  it("recognises the controls Space and Enter belong to", () => {
    expect(isInteractiveFocus({ tagName: "BUTTON" })).toBe(true);
    expect(isInteractiveFocus({ tagName: "A" })).toBe(true);
    expect(isInteractiveFocus({ tagName: "DIV", getAttribute: () => "menuitem" })).toBe(true);
    expect(isInteractiveFocus({ tagName: "DIV", getAttribute: () => null })).toBe(false);
    expect(isInteractiveFocus({ tagName: "BODY", getAttribute: () => null })).toBe(false);
    expect(isInteractiveFocus(null)).toBe(false);
  });

  it("treats a deliberately focusable element as interactive but not tabindex=-1", () => {
    // Settings' panel carries tabindex="-1" purely so focus can be
    // moved to it programmatically; that is not a control.
    expect(isInteractiveFocus({ tagName: "DIV", getAttribute: () => "0" })).toBe(true);
    expect(isInteractiveFocus({ tagName: "DIV", getAttribute: () => "-1" })).toBe(false);
  });
});
