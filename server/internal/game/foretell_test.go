package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// foretell_test.go — foretell (CR 702.143), #658.
//
// The three models foretell rides are each tested where they live
// (cast_permission_test.go, face_down_test.go, special_action_test.go).
// What is tested here is the wiring and the two rules that are
// foretell's alone: "on a LATER turn" (CR 702.143a) and "was
// foretold" (CR 702.143c).

const foretellOracle = "test-foretell-oracle"

// withForetellCard declares foretell with the given foretell cost on
// one oracle key.
func withForetellCard(t *testing.T, cost string) {
	t.Helper()
	withCatalogSpecialActions(t, func(id string) []SpecialAction {
		if id != foretellOracle {
			return nil
		}
		return []SpecialAction{{
			Kind:     SpecialActionForetell,
			Cost:     ForetellExileCost,
			CastCost: cost,
			Label:    "Foretell {2}",
		}}
	})
}

// foretellIt seeds a foretellable card in the seat's hand, pays for
// the special action and takes it. Returns the card's instance ID.
func foretellIt(t *testing.T, g *Game, p *Player) uuid.UUID {
	t.Helper()
	card := seedHandCard(p, "Saw It Coming", foretellOracle, "Instant", "{1}{U}{U}")
	p.ManaPool.AddMana(ManaToken{Color: "C"}, ManaToken{Color: "C"})
	if err := g.PerformSpecialAction(p.ID, card.InstanceID, SpecialActionForetell, SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("foretell: %v", err)
	}
	return card.InstanceID
}

// The special action itself: the {2} is paid, the card leaves the
// hand for exile, and it lands face down in the CR 702.143b state.
func TestForetellPaysTwoAndExilesFaceDown(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withForetellCard(t, "{1}{U}")
	advanceTo(t, g, StepPrecombatMain)

	id := foretellIt(t, g, me)

	if len(me.ManaPool) != 0 {
		t.Errorf("mana pool: got %d tokens left, want 0 — the {2} was not paid", len(me.ManaPool))
	}
	if handCard(me, id) != nil {
		t.Error("the foretold card is still in hand")
	}
	c := exiledCardByIDLocked(g, id)
	if c == nil {
		t.Fatal("the foretold card is not in exile")
	}
	if !CardIsForetold(*c) {
		t.Errorf("exiled card: FaceDown=%v kind=%q, want a foretold face-down card", c.FaceDown, c.FaceDownKind)
	}
	// No stack item: CR 116.2 is explicit that a special action does
	// not use the stack, and nothing about foretell is responded to.
	if g.Stack != nil && len(g.Stack.Cards) > 0 {
		t.Error("foretelling put something on the stack")
	}
}

// CR 702.143d: the owner may look at their foretold card and no other
// player may. ADR 0069 decision 2's viewers table is the only place
// that rule is written, and this is the row that reads it.
func TestForetoldCardIsVisibleToItsOwnerOnly(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	withForetellCard(t, "{1}{U}")
	advanceTo(t, g, StepPrecombatMain)

	id := foretellIt(t, g, me)
	c := exiledCardByIDLocked(g, id)
	if c == nil {
		t.Fatal("the foretold card is not in exile")
	}
	if !c.IsKnownTo(me.ID) {
		t.Error("the owner cannot see their own foretold card (CR 702.143d)")
	}
	if c.IsKnownTo(them.ID) {
		t.Error("an opponent can read a foretold card (CR 702.143d)")
	}
}

// CR 702.143a: "Cast it on a LATER turn for its foretell cost." Both
// halves — refused on the turn it was foretold, and offered at the
// foretell cost afterwards.
func TestForetoldCardIsCastableNextTurnForTheForetellCost(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withForetellCard(t, "{1}{U}")
	advanceTo(t, g, StepPrecombatMain)

	id := foretellIt(t, g, me)

	if perm := grantOn(g, me.ID, id, ZoneExile); perm != nil {
		t.Error("the foretell permission is live on the turn the card was foretold (CR 702.143a)")
	}
	me.ManaPool.AddMana(ManaToken{Color: "C"}, ManaToken{Color: "U"})
	err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, FromZone: "exile", AlternativeCost: AltCostKeyForetell})
	if !errors.Is(err, ErrNoPlayPermission) {
		t.Fatalf("cast on the turn it was foretold: got %v, want ErrNoPlayPermission", err)
	}

	// A later turn. Turn.Seq changes at every turn boundary — the same
	// floor warp's grant uses.
	g.Turn.Seq++
	perm := grantOn(g, me.ID, id, ZoneExile)
	if perm == nil {
		t.Fatal("no foretell permission on a later turn")
	}
	if perm.AltCostKey != AltCostKeyForetell {
		t.Errorf("permission key: got %q, want %q", perm.AltCostKey, AltCostKeyForetell)
	}
	c := exiledCardByIDLocked(g, id)
	if c == nil {
		t.Fatal("the foretold card left exile")
	}
	offer := perm.AlternativeCostFor(*c)
	if offer == nil || offer.ManaCost != "{1}{U}" {
		t.Fatalf("foretell offer: got %+v, want a {1}{U} cost", offer)
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, FromZone: "exile", AlternativeCost: AltCostKeyForetell}); err != nil {
		t.Fatalf("cast for the foretell cost on a later turn: %v", err)
	}
	if !g.Stack.Contains(id) {
		t.Fatal("the foretold cast did not reach the stack")
	}
	// CR 406.3a: the card turns face up as it is cast. MoveCard's
	// unconditional reset is what does it (ADR 0069 decision 5).
	for i := range g.Stack.Cards {
		if g.Stack.Cards[i].InstanceID == id && g.Stack.Cards[i].FaceDown {
			t.Error("the spell is still face down on the stack (CR 406.3a)")
		}
	}
}

// CR 702.143c: the spell is foretold because the CARD was, not
// because the foretell cost was the one paid. The bit rides the stack
// item so a resolution can read it back.
func TestForetoldSpellIsMarkedOnTheStack(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withForetellCard(t, "{1}{U}")
	advanceTo(t, g, StepPrecombatMain)

	id := foretellIt(t, g, me)
	g.Turn.Seq++
	me.ManaPool.AddMana(ManaToken{Color: "C"}, ManaToken{Color: "U"})
	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, FromZone: "exile", AlternativeCost: AltCostKeyForetell}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	item := g.StackMeta[id]
	if item == nil {
		t.Fatal("no stack item")
	}
	if !item.Foretold {
		t.Error("the spell was cast from a foretold card and is not marked foretold (CR 702.143c)")
	}
}

// The other half of CR 702.143c, and the reason the bit is not just
// `AltCost == "foretell"`: an ordinary hand cast of the same card is
// NOT a foretold spell.
func TestAHandCastIsNotForetold(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withForetellCard(t, "{1}{U}")
	advanceTo(t, g, StepPrecombatMain)
	card := seedHandCard(me, "Saw It Coming", foretellOracle, "Instant", "{1}{U}{U}")
	me.ManaPool.AddMana(ManaToken{Color: "C"}, ManaToken{Color: "U"}, ManaToken{Color: "U"})

	if err := g.CastSpell(me.ID, card.InstanceID, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("hand cast: %v", err)
	}
	if item := g.StackMeta[card.InstanceID]; item == nil || item.Foretold {
		t.Error("a hand cast is marked foretold")
	}
}

// CR 702.143f: all face-down foretold cards are revealed as the game
// ends. The owner-leaves half shipped with ADR 0069; this is the
// game-end half #658 owed.
func TestForetoldCardsAreRevealedWhenTheGameEnds(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	withForetellCard(t, "{1}{U}")
	advanceTo(t, g, StepPrecombatMain)

	id := foretellIt(t, g, me)
	if c := exiledCardByIDLocked(g, id); c == nil || c.IsKnownTo(them.ID) {
		t.Fatal("the opponent already knows the foretold card before the game ends")
	}
	g.End()
	c := exiledCardByIDLocked(g, id)
	if c == nil {
		t.Fatal("the foretold card left exile")
	}
	if !c.IsKnownTo(them.ID) {
		t.Error("the foretold card was not revealed as the game ended (CR 702.143f)")
	}
}

// CR 400.7 through the object epoch: the permission names the OBJECT
// that was foretold. A copy of the same card exiled by Path to Exile
// is a different object and is not castable out of exile — the bug
// the retired card-level `CastableZones: exile` declaration would
// have shipped.
func TestASecondCopyInExileIsNotCastable(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withForetellCard(t, "{1}{U}")
	advanceTo(t, g, StepPrecombatMain)

	foretellIt(t, g, me)
	other := NewCard("Saw It Coming", me.ID)
	other.OracleID = foretellOracle
	other.TypeLine = "Instant"
	other.ManaCost = "{1}{U}{U}"
	g.Exile.PushTop(other)

	g.Turn.Seq++
	if perm := grantOn(g, me.ID, other.InstanceID, ZoneExile); perm != nil {
		t.Error("a second copy in exile is covered by the first one's foretell permission")
	}
	me.ManaPool.AddMana(ManaToken{Color: "C"}, ManaToken{Color: "U"})
	err := g.CastSpell(me.ID, other.InstanceID, CastSpellParams{Strict: true, FromZone: "exile", AlternativeCost: AltCostKeyForetell})
	if !errors.Is(err, ErrNoPlayPermission) {
		t.Fatalf("cast of an unforetold copy from exile: got %v, want ErrNoPlayPermission", err)
	}
}

// Foretell is a built kind: the performer exists, so the verb pays
// for it and the enumerator offers it.
func TestForetellIsABuiltKind(t *testing.T) {
	if !SpecialActionKindBuilt(SpecialActionForetell) {
		t.Error("foretell is declared but has no performer")
	}
}

// The special action is one mutation and an undo takes it back whole:
// the card returns to hand, the {2} is refunded and no permission is
// left pointing at a card that is no longer in exile.
func TestForetellSurvivesAnUndo(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withForetellCard(t, "{1}{U}")
	advanceTo(t, g, StepPrecombatMain)
	card := seedHandCard(me, "Saw It Coming", foretellOracle, "Instant", "{1}{U}{U}")
	me.ManaPool.AddMana(ManaToken{Color: "C"}, ManaToken{Color: "C"})

	before := g.Clone()
	if err := g.PerformSpecialAction(me.ID, card.InstanceID, SpecialActionForetell, SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("foretell: %v", err)
	}
	g.WithWriteLock(func() { g.RestoreFrom(before) })

	me = g.Seats[0]
	if handCard(me, card.InstanceID) == nil {
		t.Error("undo did not return the foretold card to hand")
	}
	if exiledCardByIDLocked(g, card.InstanceID) != nil {
		t.Error("undo left the card in exile")
	}
	if len(me.ManaPool) != 2 {
		t.Errorf("mana pool after undo: got %d tokens, want 2", len(me.ManaPool))
	}
	if len(me.CastPermissions) != 0 {
		t.Errorf("undo left %d cast permissions behind", len(me.CastPermissions))
	}
}

// Both halves of the foretold state are carried state, not rebuilt:
// the face-down kind (ADR 0069 decision 8) and the permission
// (ADR 0066 decision 9). A snapshot that dropped either would hand
// back a card nobody can read and nobody can cast.
func TestForetoldStateSurvivesASnapshotRoundTrip(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withForetellCard(t, "{1}{U}")
	advanceTo(t, g, StepPrecombatMain)
	id := foretellIt(t, g, me)

	_, restored := roundTrip(t, g)
	restored.Turn.Seq++

	c := exiledCardByIDLocked(restored, id)
	if c == nil {
		t.Fatal("the foretold card did not survive the round trip")
	}
	if !CardIsForetold(*c) {
		t.Errorf("restored card: FaceDown=%v kind=%q, want a foretold face-down card", c.FaceDown, c.FaceDownKind)
	}
	if !c.IsKnownTo(me.ID) {
		t.Error("the owner lost sight of their foretold card across the round trip")
	}
	if perm := restored.CastPermissionForLocked(me.ID, *c, ZoneExile); perm == nil {
		t.Error("the foretell permission did not survive the round trip")
	}
}
