package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// counter_placer_test.go is ADR 0056 PR 1's catalog half: test plan
// items 12 (no existing counter replacement applies to a placement on a
// PLAYER unless it opted in) and 13 (the three cards that read WHO PUT
// the counters read it off the event).
//
// Both are guards against a change that is invisible today. Nothing in
// the engine sets ReplacementEvent.CounterFromCombatDamage yet, and
// until PR 2 branches the damage tail nothing puts counters as a damage
// result — which is exactly why the readers land now. PR 2 is what
// makes the wrong answers reachable, and "Doubling Season starts
// doubling combat-damage counters" is not a thing to discover in the PR
// that branches the tail.

const (
	vorinclexPlacerOracle = "5a3fdf5a-bff8-4896-b288-3f43f9a72d9b"
	laezelPlacerOracle    = "c066f921-e349-45d1-8ec3-0955d10bbf19"
)

// --- item 12: the registry sweep -----------------------------------

// countersOnOptIns are the catalog cards that DELIBERATELY apply to a
// counter placement on a player. ADR 0056 Decision 5 keeps one
// ReplacementEventKind for both targets — the cards that replace
// counters on players are the same cards that replace them on
// permanents — and this list is the price of that: a card that reaches
// a player event has to say so here.
//
// Both entries print the clause. Vorinclex is "counters on a permanent
// OR PLAYER", twice; Lae'zel is "on a creature or planeswalker you
// control OR ON YOURSELF".
var countersOnOptIns = map[string]string{
	"Vorinclex, Monstrous Raider": `"If you would put one or more counters on a permanent
		or player" — both halves of the card name players.`,
	"Lae'zel, Vlaakith's Champion": `"on a creature or planeswalker you control or on
		yourself" — the "or on yourself" clause was dead before ADR 0056
		and shipped as a declared caveat.`,
}

// TestNoCounterReplacementLeaksOntoAPlayerEvent is ADR 0056 test plan
// item 12's structural half.
//
// The safety argument for one event kind is that every counter
// replacement written before this change opens its AppliesTo with
// g.LookupCardForEffect(ev.CounterTarget), which fails for uuid.Nil. It
// is a true statement about today's catalog and NOT a property the type
// system enforces, so it is a test: a future card that forgets the
// check would otherwise silently double somebody's poison.
func TestNoCounterReplacementLeaksOntoAPlayerEvent(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0].ID, g.Seats[1].ID

	for _, spec := range All() {
		if len(spec.Replacements) == 0 {
			continue
		}
		for i, r := range spec.Replacements {
			if r.AppliesTo == nil || !watchesCounters(r.Watches) {
				continue
			}
			// A source the card controls, so a controller-scoped
			// predicate cannot decline for the wrong reason.
			src := &game.Card{
				InstanceID: uuid.New(),
				Name:       spec.Name,
				OracleID:   spec.OracleID,
				Owner:      me,
				Controller: me,
			}
			for _, victim := range []uuid.UUID{me, them} {
				ev := &game.ReplacementEvent{
					Kind:          game.RepEventCounter,
					CounterPlayer: victim,
					CounterName:   game.CounterPoison,
					CounterDelta:  1,
				}
				if !r.AppliesTo(ev, g, src) {
					continue
				}
				if _, ok := countersOnOptIns[spec.Name]; ok {
					continue
				}
				t.Errorf(`%s replacement #%d (%q) applies to a counter placement on a PLAYER.

ADR 0056 Decision 5 keeps ONE ReplacementEventKind for counters on
permanents and counters on players, and the reason that is safe is that
every card predicate opens with a card lookup on ev.CounterTarget,
which fails for uuid.Nil.

Either add the guard — a card predicate should start with

    target, ok := g.LookupCardForEffect(ev.CounterTarget)
    if !ok { return false }

— or, if the card really does name players ("on a permanent or player",
"or on yourself"), add it to countersOnOptIns with the printed clause.`,
					spec.Name, i, r.Label)
			}
		}
	}
}

// TestTheCountersOnPlayersOptInsAreStillRegistered is the reverse
// direction: an opt-in for a card that no longer exists, or no longer
// reaches a player event, is a stale exemption — and a stale exemption
// is the only evidence that a deletion was incomplete.
func TestTheCountersOnPlayersOptInsAreStillRegistered(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	seen := map[string]bool{}
	for _, spec := range All() {
		if _, ok := countersOnOptIns[spec.Name]; !ok {
			continue
		}
		src := &game.Card{
			InstanceID: uuid.New(), Name: spec.Name, OracleID: spec.OracleID,
			Owner: me, Controller: me,
		}
		ev := &game.ReplacementEvent{
			Kind:          game.RepEventCounter,
			CounterPlayer: me,
			CounterPlacer: me,
			CounterName:   game.CounterPoison,
			CounterDelta:  1,
		}
		for _, r := range spec.Replacements {
			if r.AppliesTo != nil && watchesCounters(r.Watches) && r.AppliesTo(ev, g, src) {
				seen[spec.Name] = true
			}
		}
	}
	for name, why := range countersOnOptIns {
		if !seen[name] {
			t.Errorf(`countersOnOptIns exempts %q, but no registered replacement of that
card applies to a counter placement on a player any more.

The exemption reads:
%s

Delete the row, or find out why the card stopped reaching the event.`, name, why)
		}
	}
}

func watchesCounters(watches []game.EventKind) bool {
	for _, w := range watches {
		if w == game.EventCounterPlaced {
			return true
		}
	}
	return false
}

// --- item 13: the placer -------------------------------------------

// TestVorinclexReadsThePlacerNotTheTargetsController is the halving
// half's fix. An OPPONENT putting counters on Vorinclex's controller's
// creature is halved, as printed — the old reading doubled it, because
// it asked who controlled the permanent.
//
// This is the board ADR 0056's PR 2 creates: an opponent's infect
// attacker putting -1/-1 counters on your blocker.
func TestVorinclexReadsThePlacerNotTheTargetsController(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0].ID, g.Seats[1].ID
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Vorinclex, Monstrous Raider",
		OracleID:   vorinclexPlacerOracle,
		TypeLine:   "Legendary Creature — Phyrexian Praetor",
		Power:      6, Toughness: 6,
		Owner: me, Controller: me,
	})
	mine := seedCreature(g, "Mine", me)

	// An OPPONENT places four counters on MY creature. Halved, because
	// an opponent is putting them.
	g.WithWriteLock(func() {
		if err := g.AddCounterByForEffect(them, mine, game.CounterPlusOne, 4); err != nil {
			t.Fatalf("AddCounterByForEffect: %v", err)
		}
	})
	if got := countersOn(g, mine, game.CounterPlusOne); got != 2 {
		t.Errorf(`counters = %d, want 2 (an opponent placed them, so halved).

Vorinclex reads WHO IS PUTTING the counters, not who controls the
permanent. The old reading saw "a permanent you control" and doubled.`, got)
	}

	// And my own placement on the OPPONENT's creature doubles.
	theirs := seedCreature(g, "Theirs", them)
	g.WithWriteLock(func() {
		if err := g.AddCounterByForEffect(me, theirs, game.CounterPlusOne, 3); err != nil {
			t.Fatalf("AddCounterByForEffect: %v", err)
		}
	})
	if got := countersOn(g, theirs, game.CounterPlusOne); got != 6 {
		t.Errorf("counters on the opponent's creature = %d, want 6 (I placed them, so doubled)", got)
	}
}

// TestVorinclexAppliesToCountersOnPlayers is the "or player" half of
// the card, dead until ADR 0056 gave player counters a window.
func TestVorinclexAppliesToCountersOnPlayers(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0].ID, g.Seats[1].ID
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Vorinclex, Monstrous Raider",
		OracleID:   vorinclexPlacerOracle,
		TypeLine:   "Legendary Creature — Phyrexian Praetor",
		Power:      6, Toughness: 6,
		Owner: me, Controller: me,
	})

	// The opponent gives ME three poison. Halved, rounded down.
	g.WithWriteLock(func() {
		if err := g.AddPlayerCounterByForEffect(them, me, game.CounterPoison, 3); err != nil {
			t.Fatalf("AddPlayerCounterByForEffect: %v", err)
		}
	})
	if got := playerCounterFor(g, me, game.CounterPoison); got != 1 {
		t.Errorf("poison = %d, want 1 (3 halved, rounded down)", got)
	}

	// I give THEM two. Doubled.
	g.WithWriteLock(func() {
		if err := g.AddPlayerCounterByForEffect(me, them, game.CounterPoison, 2); err != nil {
			t.Fatalf("AddPlayerCounterByForEffect: %v", err)
		}
	})
	if got := playerCounterFor(g, them, game.CounterPoison); got != 4 {
		t.Errorf("opponent poison = %d, want 4 (2 doubled)", got)
	}
}

// TestVorinclexFallsBackWhenNoPlacerIsNamed pins the declared caveat.
// A placement that names nobody — a paid cost, a permanent entering
// with counters, a sandbox edit — still reads the target's controller,
// which is CR 606.4 and CR 714.3's answer for most of them.
func TestVorinclexFallsBackWhenNoPlacerIsNamed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Vorinclex, Monstrous Raider",
		OracleID:   vorinclexPlacerOracle,
		TypeLine:   "Legendary Creature — Phyrexian Praetor",
		Power:      6, Toughness: 6,
		Owner: me, Controller: me,
	})
	mine := seedCreature(g, "Mine", me)
	if err := g.AddCounter(mine, game.CounterPlusOne, 2); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if got := countersOn(g, mine, game.CounterPlusOne); got != 4 {
		t.Errorf("counters = %d, want 4 — with no placer named the old reading stands", got)
	}
}

// TestLaezelIncreasesCountersPutOnYou is the clause the card shipped a
// caveat about: "or on yourself". An experience counter the controller
// puts on themselves is one more.
func TestLaezelIncreasesCountersPutOnYou(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0].ID, g.Seats[1].ID
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Lae'zel, Vlaakith's Champion",
		OracleID:   laezelPlacerOracle,
		TypeLine:   "Legendary Creature — Gith Warrior",
		Power:      3, Toughness: 3,
		Owner: me, Controller: me,
	})

	g.WithWriteLock(func() {
		if err := g.AddPlayerCounterByForEffect(me, me, game.CounterExperience, 1); err != nil {
			t.Fatalf("AddPlayerCounterByForEffect: %v", err)
		}
	})
	if got := playerCounterFor(g, me, game.CounterExperience); got != 2 {
		t.Errorf("experience = %d, want 2 (1 plus one)", got)
	}

	// An OPPONENT giving Lae'zel's controller poison is not "you put",
	// so it is not increased. This is the ADR 0056 PR 2 board: an
	// opponent's infect creature hitting Lae'zel's controller.
	g.WithWriteLock(func() {
		if err := g.AddPlayerCounterByForEffect(them, me, game.CounterPoison, 2); err != nil {
			t.Fatalf("AddPlayerCounterByForEffect: %v", err)
		}
	})
	if got := playerCounterFor(g, me, game.CounterPoison); got != 2 {
		t.Errorf(`poison = %d, want 2.

"If YOU would put" — an opponent put these. Before ADR 0056 the card
read the last resolution and counted uuid.Nil as "you", which would
make an opponent's infect damage give Lae'zel's controller MORE poison
than printed.`, got)
	}
}

// TestLaezelDoesNotIncreaseAnOpponentsPlacementOnYourCreature is the
// same statement for the card half, and the one the placer fixes
// outright: combat damage happens inside no resolution at all, so the
// old heuristic had nothing to read.
func TestLaezelDoesNotIncreaseAnOpponentsPlacementOnYourCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0].ID, g.Seats[1].ID
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Lae'zel, Vlaakith's Champion",
		OracleID:   laezelPlacerOracle,
		TypeLine:   "Legendary Creature — Gith Warrior",
		Power:      3, Toughness: 3,
		Owner: me, Controller: me,
	})
	mine := seedCreature(g, "Mine", me)

	g.WithWriteLock(func() {
		if err := g.AddCounterByForEffect(them, mine, game.CounterMinusOne, 2); err != nil {
			t.Fatalf("AddCounterByForEffect: %v", err)
		}
	})
	if got := countersOn(g, mine, game.CounterMinusOne); got != 2 {
		t.Errorf("-1/-1 counters = %d, want 2 — an opponent placed them, so Lae'zel adds nothing", got)
	}

	g.WithWriteLock(func() {
		if err := g.AddCounterByForEffect(me, mine, game.CounterPlusOne, 1); err != nil {
			t.Fatalf("AddCounterByForEffect: %v", err)
		}
	})
	if got := countersOn(g, mine, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2 — I placed them, so Lae'zel adds one", got)
	}
}

// TestDoublingSeasonSkipsCombatDamageCounters is the judges' reading of
// "if AN EFFECT would put": combat damage is a turn-based action, so
// the -1/-1 counters a wither or infect attacker puts on a blocker are
// not doubled — while the same keyword on a spell, or a fight, is.
//
// Nothing sets CounterFromCombatDamage until ADR 0056's PR 2 branches
// the damage tail, so this test drives the flag directly. That is the
// point: the reader lands before the writer, so the day the tail
// branches, Doubling Season is already right.
func TestDoublingSeasonSkipsCombatDamageCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	ds := seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", me)
	mine := seedCreature(g, "Mine", me)

	applies := func(ev *game.ReplacementEvent) bool {
		hit := false
		for _, s := range All() {
			if s.OracleID != doublingSeasonOracle {
				continue
			}
			src := &game.Card{InstanceID: ds, Name: s.Name, OracleID: s.OracleID, Owner: me, Controller: me}
			for _, r := range s.Replacements {
				if r.AppliesTo != nil && watchesCounters(r.Watches) && r.AppliesTo(ev, g, src) {
					hit = true
				}
			}
		}
		return hit
	}

	effectPlacement := &game.ReplacementEvent{
		Kind:          game.RepEventCounter,
		CounterTarget: mine,
		CounterName:   game.CounterMinusOne,
		CounterDelta:  2,
	}
	if !applies(effectPlacement) {
		t.Error("Doubling Season does not apply to an ordinary effect's counter placement — the card's whole second half")
	}

	combatPlacement := &game.ReplacementEvent{
		Kind:                    game.RepEventCounter,
		CounterTarget:           mine,
		CounterName:             game.CounterMinusOne,
		CounterDelta:            2,
		CounterFromCombatDamage: true,
	}
	if applies(combatPlacement) {
		t.Error(`Doubling Season doubles counters that are the RESULT OF COMBAT DAMAGE.

"If AN EFFECT would put one or more counters" — combat damage is a
turn-based action, not an effect, so wither and infect COMBAT damage is
not doubled. A wither SPELL, or a fight, still is. ADR 0056 Decision 5
cites the judge rulings.`)
	}
}

// playerCounterFor reads a seat's counter of the named kind.
func playerCounterFor(g *game.Game, playerID uuid.UUID, kind string) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, p := range g.Seats {
			if p.ID == playerID {
				n = p.Counters[kind]
			}
		}
	})
	return n
}
