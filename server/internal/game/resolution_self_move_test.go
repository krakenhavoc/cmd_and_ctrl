package game

import (
	"testing"

	"github.com/google/uuid"
)

// resolution_self_move_test.go — #489. A spell whose own text moves it
// out of the stack as it resolves.
//
// "Exile Ascend from Avernus", Genesis Ultimatum's "exile Genesis
// Ultimatum", and by the same shape any "shuffle this into your
// library" rider. The instruction is the card's, so it runs in
// `OnResolve`, which is before the resolution frame decides where the
// spell goes next — and the frame's three exits all assume the spell
// is still on the stack.
//
// The bug has had two faces. As filed it was an ERROR:
// `MoveCard(g.Stack, …)` returned "card instance not found in zone",
// `PassPriority` handed it back to the caller, and the post-resolution
// state checks and priority reset never ran. #529 put the stack exit on
// the shared exit primitive, which finds a card's zone by scan, and the
// error became a silent wrong move: the frame found the spell in exile
// and moved it from there into the graveyard. Both faces are one
// assumption, and `spellMovedItselfLocked` is where it is now checked.
//
// The second half of the issue is its own test below: a resolution that
// DOES fail must still leave the game playable.

const selfMoverOracleID = "test-self-mover"

// castSelfMover registers a one-card stub catalog whose OnResolve runs
// `move` on the resolving spell itself, casts a card of `typeLine` and
// passes priority around the table to resolve it. It returns the game,
// the spell's instance ID and the error PassPriority reported.
func castSelfMover(t *testing.T, typeLine string, move func(g *Game, self uuid.UUID) error) (*Game, uuid.UUID, error) {
	t.Helper()
	var self uuid.UUID
	withEffectHooks(t,
		func(g *Game, _ *StackItem, oracleID string) error {
			if oracleID != selfMoverOracleID {
				return nil
			}
			return move(g, self)
		},
		nil,
		func(oracleID string) bool { return oracleID == selfMoverOracleID },
	)

	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	caster := g.Seats[0]
	self = pushTypedCardToHand(caster, "Self Mover", typeLine)
	for i := range caster.Hand.Cards {
		if caster.Hand.Cards[i].InstanceID == self {
			caster.Hand.Cards[i].OracleID = selfMoverOracleID
			break
		}
	}
	if err := g.CastSpell(caster.ID, self, CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	var lastErr error
	for i := 0; i < len(g.Seats); i++ {
		if err := g.PassPriority(); err != nil {
			lastErr = err
		}
	}
	return g, self, lastErr
}

// zoneKindOf names the zone a card is in, for a readable failure.
func zoneKindOf(t *testing.T, g *Game, id uuid.UUID) ZoneKind {
	t.Helper()
	var kind ZoneKind = "<nowhere>"
	g.WithWriteLock(func() {
		if z := g.findCardZoneLocked(id); z != nil {
			kind = z.Kind
		}
	})
	return kind
}

// TestASorceryThatExilesItselfStaysInExile is the issue's own repro.
// CR 608.2n puts the SPELL into its owner's graveyard, and a spell that
// is no longer on the stack is not one.
func TestASorceryThatExilesItselfStaysInExile(t *testing.T) {
	g, self, err := castSelfMover(t, "Sorcery", func(g *Game, self uuid.UUID) error {
		return g.ExileCardForEffect(self)
	})
	if err != nil {
		t.Fatalf("PassPriority: %v — a spell that exiles itself must not wedge the pass", err)
	}
	if got := zoneKindOf(t, g, self); got != ZoneExile {
		t.Errorf("the self-exiled spell is in %s, want exile — the resolution frame must not move "+
			"a spell its own effect has already placed", got)
	}
	if g.Seats[0].Graveyard.Contains(self) {
		t.Error("and it is NOT in its owner's graveyard: CR 608.2n puts the spell ON THE STACK there, " +
			"and there was none by the time the frame looked")
	}
}

// TestASorceryThatTucksItselfStaysInItsLibrary is the other rider the
// issue names — "shuffle this into your library" / "put this on the
// bottom" — on the same one check.
func TestASorceryThatTucksItselfStaysInItsLibrary(t *testing.T) {
	g, self, err := castSelfMover(t, "Sorcery", func(g *Game, self uuid.UUID) error {
		_, err := g.routeCardToZoneLocked(zoneRoute{CardID: self, Dst: ZoneLibrary, ToBottom: true})
		return err
	})
	if err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if got := zoneKindOf(t, g, self); got != ZoneLibrary {
		t.Errorf("the self-tucked spell is in %s, want library", got)
	}
	if lib := g.Seats[0].Library; lib.Size() == 0 || lib.Cards[0].InstanceID != self {
		t.Error("and it is still on the BOTTOM, where the tuck put it — the frame did not re-route it")
	}
}

// TestAPermanentSpellThatMovesItselfDoesNotWedgeTheResolution is the
// branch that still returned the original `card instance not found in
// zone` after #529: the battlefield entry keeps its own
// `MoveCard(g.Stack, …)`. It is the reason the check lives in the
// resolution frame rather than in routeStackCardToGraveyardLocked.
func TestAPermanentSpellThatMovesItselfDoesNotWedgeTheResolution(t *testing.T) {
	g, self, err := castSelfMover(t, "Creature — Test", func(g *Game, self uuid.UUID) error {
		return g.ExileCardForEffect(self)
	})
	if err != nil {
		t.Fatalf("PassPriority: %v — the permanent branch must tolerate a spell that moved itself", err)
	}
	if got := zoneKindOf(t, g, self); got != ZoneExile {
		t.Errorf("the self-exiled permanent spell is in %s, want exile", got)
	}
	if g.Battlefield.Contains(self) {
		t.Error("it never entered the battlefield: the object the entry was about had already left the stack")
	}
}

// TestAResolutionThatMovedItselfStillRunsThePostResolutionChecks — the
// other half of the report. The state-based actions and the CR 117.3b
// priority reset are not the resolution's; they belong to the pass, and
// they run whatever the resolution did.
func TestAResolutionThatMovedItselfStillRunsThePostResolutionChecks(t *testing.T) {
	g, self, err := castSelfMover(t, "Sorcery", func(g *Game, self uuid.UUID) error {
		return g.ExileCardForEffect(self)
	})
	if err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if g.Turn.PriorityHolder != g.Turn.ActiveSeat {
		t.Errorf("priority holder %d, want the active seat %d (CR 117.3b)",
			g.Turn.PriorityHolder, g.Turn.ActiveSeat)
	}
	if len(g.Stack.Cards) != 0 || len(g.StackMeta) != 0 {
		t.Errorf("stack holds %d cards / %d items, want empty — the item was retired",
			len(g.Stack.Cards), len(g.StackMeta))
	}
	if got := zoneKindOf(t, g, self); got != ZoneExile {
		t.Errorf("control check: the spell is in %s, want exile", got)
	}
}

// TestAFailedResolutionStillLeavesTheGamePlayable is #489's second
// sentence, and the half that is not about self-moves at all:
// "PassPriority returns that error to the caller and SKIPS the
// post-resolution state checks and the priority reset, so the game is
// left in a half-resolved state rather than continuing."
//
// The failure has to be injected, because the shape that used to
// produce it is fixed above. The injection is a real one the engine
// can still produce: a replacement rewrites the spell's exit to the
// BATTLEFIELD, which the exit primitive refuses by design — "not an
// exit, the entry path owns these" (routeDestinationLocked) — so
// `ErrZoneNotFound` comes back out of the resolution frame.
//
// What is pinned is the frame's posture rather than that particular
// error: an unhappy resolution must still sweep the board, drain the
// triggers and hand priority back.
func TestAFailedResolutionStillLeavesTheGamePlayable(t *testing.T) {
	var self uuid.UUID
	withEffectHooks(t, nil, nil, func(string) bool { return false })

	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	caster := g.Seats[0]
	self = pushTypedCardToHand(caster, "Doomed Spell", "Sorcery")

	// A creature carrying lethal damage, which the state-based-action
	// sweep owes a graveyard (CR 704.5g). Nothing but the
	// post-resolution sweep can move it, so it is the witness that the
	// sweep ran. (A printed 0 toughness would not do: the sweep skips
	// those as the import placeholder — see stateBasedActionsLocked.)
	lethal := NewCard("Withered", caster.ID)
	lethal.TypeLine = "Creature — Test"
	lethal.Power, lethal.Toughness = 1, 1
	lethal.DamageMarked = 1
	g.Battlefield.PushTop(lethal)

	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventZoneMove},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventMove && ev.CardID == self && ev.OldZone == ZoneStack
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.NewZone = ZoneBattlefield
				return nil
			},
			Label: "instead, put it onto the battlefield",
		})
	})

	if err := g.CastSpell(caster.ID, self, CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	for i := 0; i < len(g.Seats); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority %d: %v — a failed resolution must not be returned as a failed pass", i, err)
		}
	}

	if g.Turn.PriorityHolder != g.Turn.ActiveSeat {
		t.Errorf("priority holder %d, want the active seat %d — the reset was skipped, "+
			"which is the wedge #489 reported", g.Turn.PriorityHolder, g.Turn.ActiveSeat)
	}
	if g.Battlefield.Contains(lethal.InstanceID) {
		t.Error("the zero-toughness creature is still on the battlefield: the post-resolution " +
			"state-based-action sweep did not run (CR 704.5g)")
	}
	found := false
	for _, ev := range g.Events {
		if ev.Kind == EventEffectError && ev.ErrorMsg == ErrZoneNotFound.Error() {
			found = true
		}
	}
	if !found {
		t.Error("the failure is reported as an EventEffectError rather than swallowed — " +
			"the table has to be told the resolution went wrong")
	}
	// Deliberately NOT asserted: where the spell itself ended up. A
	// refused exit moved nothing, so the card is still on the stack
	// with its StackItem retired, and the next resolution's
	// no-StackMeta fallback routes it. Forcing it off with a raw
	// MoveCard would be a second exit path, which is the thing
	// zone_route.go exists to prevent.
}
