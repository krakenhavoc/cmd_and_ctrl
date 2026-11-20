package testsp

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestDealDamageToPlayerAppliesLifeAndEmitsEvent(t *testing.T) {
	g := effectstest.NewTestGame(t)
	target := g.Seats[1]
	before := target.Life

	g.WithWriteLock(func() {
		item := &game.StackItem{Controller: g.Seats[0].ID, SourceCardID: uuid.New()}
		if err := (effects.DealDamage{Target: target.ID, Amount: 3}).Apply(effectstest.CtxFor(g, item)); err != nil {
			t.Fatalf("Apply: %v", err)
		}
	})

	if got := target.Life; got != before-3 {
		t.Errorf("target life: got %d, want %d", got, before-3)
	}
}

func TestDealDamageToCreatureMarksDamage(t *testing.T) {
	g := effectstest.NewTestGame(t)
	creatureID := effectstest.PushBattlefieldCreature(g, g.Seats[0].ID)

	g.WithWriteLock(func() {
		if err := (effects.DealDamage{Target: creatureID, Amount: 2}).Apply(effectstest.CtxFor(g, &game.StackItem{})); err != nil {
			t.Fatalf("Apply: %v", err)
		}
	})

	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == creatureID && c.DamageMarked != 2 {
			t.Errorf("DamageMarked: got %d, want 2", c.DamageMarked)
		}
	}
}

func TestGainLife(t *testing.T) {
	g := effectstest.NewTestGame(t)
	p := g.Seats[0]
	before := p.Life

	g.WithWriteLock(func() {
		if err := (effects.GainLife{Player: p.ID, Amount: 5}).Apply(effectstest.CtxFor(g, &game.StackItem{})); err != nil {
			t.Fatalf("Apply: %v", err)
		}
	})

	if p.Life != before+5 {
		t.Errorf("life: got %d, want %d", p.Life, before+5)
	}
}

func TestDrawCards(t *testing.T) {
	g := effectstest.NewTestGame(t)
	p := g.Seats[0]
	before := p.Hand.Size()
	libBefore := p.Library.Size()

	g.WithWriteLock(func() {
		if err := (effects.DrawCards{Player: p.ID, N: 3}).Apply(effectstest.CtxFor(g, &game.StackItem{})); err != nil {
			t.Fatalf("Apply: %v", err)
		}
	})

	if got := p.Hand.Size() - before; got != 3 {
		t.Errorf("hand delta: got %d, want 3", got)
	}
	if got := libBefore - p.Library.Size(); got != 3 {
		t.Errorf("library delta: got %d, want 3", got)
	}
}

func TestMillCards(t *testing.T) {
	g := effectstest.NewTestGame(t)
	p := g.Seats[1]
	libBefore := p.Library.Size()
	gyBefore := p.Graveyard.Size()

	g.WithWriteLock(func() {
		if err := (effects.MillCards{Player: p.ID, N: 2}).Apply(effectstest.CtxFor(g, &game.StackItem{})); err != nil {
			t.Fatalf("Apply: %v", err)
		}
	})

	if got := libBefore - p.Library.Size(); got != 2 {
		t.Errorf("library delta: got %d, want 2", got)
	}
	if got := p.Graveyard.Size() - gyBefore; got != 2 {
		t.Errorf("graveyard delta: got %d, want 2", got)
	}
}

func TestDiscardCards(t *testing.T) {
	g := effectstest.NewTestGame(t)
	p := g.Seats[0]
	// Ensure at least 2 cards in hand to discard.
	effectstest.PushHandCard(g, p)
	effectstest.PushHandCard(g, p)
	before := p.Hand.Size()
	gyBefore := p.Graveyard.Size()

	g.WithWriteLock(func() {
		if err := (effects.DiscardCards{Player: p.ID, N: 2}).Apply(effectstest.CtxFor(g, &game.StackItem{})); err != nil {
			t.Fatalf("Apply: %v", err)
		}
	})

	if got := before - p.Hand.Size(); got != 2 {
		t.Errorf("hand delta: got %d, want 2", got)
	}
	if got := p.Graveyard.Size() - gyBefore; got != 2 {
		t.Errorf("graveyard delta: got %d, want 2", got)
	}
}

func TestDestroyTarget(t *testing.T) {
	g := effectstest.NewTestGame(t)
	owner := g.Seats[0]
	id := effectstest.PushBattlefieldCreature(g, owner.ID)

	g.WithWriteLock(func() {
		if err := (effects.DestroyTarget{Target: id}).Apply(effectstest.CtxFor(g, &game.StackItem{})); err != nil {
			t.Fatalf("Apply: %v", err)
		}
	})

	if g.Battlefield.Contains(id) {
		t.Errorf("creature still on battlefield")
	}
	if !owner.Graveyard.Contains(id) {
		t.Errorf("creature not in owner graveyard")
	}
}

func TestExileTarget(t *testing.T) {
	g := effectstest.NewTestGame(t)
	id := effectstest.PushBattlefieldCreature(g, g.Seats[0].ID)

	g.WithWriteLock(func() {
		if err := (effects.ExileTarget{Target: id}).Apply(effectstest.CtxFor(g, &game.StackItem{})); err != nil {
			t.Fatalf("Apply: %v", err)
		}
	})

	if g.Battlefield.Contains(id) {
		t.Errorf("card still on battlefield")
	}
	if !g.Exile.Contains(id) {
		t.Errorf("card not in exile")
	}
}

func TestBounceToHand(t *testing.T) {
	g := effectstest.NewTestGame(t)
	owner := g.Seats[0]
	id := effectstest.PushBattlefieldCreature(g, owner.ID)

	g.WithWriteLock(func() {
		if err := (effects.BounceToHand{Target: id}).Apply(effectstest.CtxFor(g, &game.StackItem{})); err != nil {
			t.Fatalf("Apply: %v", err)
		}
	})

	if g.Battlefield.Contains(id) {
		t.Errorf("card still on battlefield")
	}
	if !owner.Hand.Contains(id) {
		t.Errorf("card not in owner hand")
	}
}

func TestTapUntap(t *testing.T) {
	g := effectstest.NewTestGame(t)
	id := effectstest.PushBattlefieldCreature(g, g.Seats[0].ID)

	g.WithWriteLock(func() {
		if err := (effects.TapTarget{Target: id}).Apply(effectstest.CtxFor(g, &game.StackItem{})); err != nil {
			t.Fatalf("Tap: %v", err)
		}
	})
	foundTapped := false
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			foundTapped = c.Tapped
		}
	}
	if !foundTapped {
		t.Errorf("TapTarget did not tap")
	}

	g.WithWriteLock(func() {
		if err := (effects.UntapTarget{Target: id}).Apply(effectstest.CtxFor(g, &game.StackItem{})); err != nil {
			t.Fatalf("Untap: %v", err)
		}
	})
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id && c.Tapped {
			t.Errorf("UntapTarget left card tapped")
		}
	}
}

func TestAddCounter(t *testing.T) {
	g := effectstest.NewTestGame(t)
	id := effectstest.PushBattlefieldCreature(g, g.Seats[0].ID)

	g.WithWriteLock(func() {
		if err := (effects.AddCounter{Target: id, Kind: "+1/+1", N: 3}).Apply(effectstest.CtxFor(g, &game.StackItem{})); err != nil {
			t.Fatalf("Apply: %v", err)
		}
	})

	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id && c.Counters["+1/+1"] != 3 {
			t.Errorf("+1/+1 count: got %d, want 3", c.Counters["+1/+1"])
		}
	}
}

func TestCreateToken(t *testing.T) {
	g := effectstest.NewTestGame(t)
	owner := g.Seats[0]
	bfBefore := len(g.Battlefield.Cards)

	g.WithWriteLock(func() {
		err := (effects.CreateToken{Controller: owner.ID, Template: effects.TokenCard("2/2 blue Bird with flying"), N: 2}).Apply(effectstest.CtxFor(g, &game.StackItem{}))
		if err != nil {
			t.Fatalf("Apply: %v", err)
		}
	})

	if got := len(g.Battlefield.Cards) - bfBefore; got != 2 {
		t.Errorf("battlefield delta: got %d, want 2", got)
	}
	// Both tokens should be controlled by owner and known to all seats.
	tokens := g.Battlefield.Cards[bfBefore:]
	for _, tok := range tokens {
		if tok.Controller != owner.ID {
			t.Errorf("token Controller: got %v, want %v", tok.Controller, owner.ID)
		}
		for _, seat := range g.Seats {
			if !tok.IsKnownTo(seat.ID) {
				t.Errorf("seat %s is not a knower of the token", seat.Name)
			}
		}
	}
}

func TestReturnFromGraveyard(t *testing.T) {
	g := effectstest.NewTestGame(t)
	owner := g.Seats[0]
	id := uuid.New()
	owner.Graveyard.PushTop(game.Card{
		InstanceID: id,
		Name:       "Buried Treasure",
		Owner:      owner.ID,
		Controller: owner.ID,
	})

	g.WithWriteLock(func() {
		if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(effectstest.CtxFor(g, &game.StackItem{})); err != nil {
			t.Fatalf("Apply: %v", err)
		}
	})

	if owner.Graveyard.Contains(id) {
		t.Errorf("card still in graveyard")
	}
	if !owner.Hand.Contains(id) {
		t.Errorf("card not in hand")
	}
}

func TestSearchLibraryFindsAndMoves(t *testing.T) {
	g := effectstest.NewTestGame(t)
	p := g.Seats[1]
	// Seed a recognisable card deep in the library.
	needleID := uuid.New()
	p.Library.PushTop(game.Card{
		InstanceID: needleID,
		Name:       "Needle",
		TypeLine:   "Basic Land — Forest",
		Owner:      p.ID,
		Controller: p.ID,
	})

	g.WithWriteLock(func() {
		err := (effects.SearchLibrary{
			Player: p.ID,
			Predicate: func(c game.Card) bool {
				return c.Name == "Needle"
			},
			Dest:    game.ZoneHand,
			Limit:   1,
			Reveal:  true,
			Shuffle: true,
		}).Apply(effectstest.CtxFor(g, &game.StackItem{}))
		if err != nil {
			t.Fatalf("Apply: %v", err)
		}
	})
	// A search for a stated quality asks even against one match (CR
	// 701.23b, #3053).
	effectstest.TakeSearchPrompts(t, g)

	if !p.Hand.Contains(needleID) {
		t.Errorf("Needle not in hand after search")
	}
	if p.Library.Contains(needleID) {
		t.Errorf("Needle still in library after search")
	}
}

// TestCounterTargetSpell wires the stack + CounterTarget primitive
// end-to-end. We place a spell card on the stack + StackMeta entry,
// call CounterTarget, verify the spell left the stack into the
// owner's graveyard.
func TestCounterTargetSpell(t *testing.T) {
	g := effectstest.NewTestGame(t)
	caster := g.Seats[0]
	// Seed a spell on the stack by hand (no CastSpell — avoids
	// sorcery-speed plumbing).
	cardID := uuid.New()
	g.Stack.PushTop(game.Card{
		InstanceID: cardID,
		Name:       "Stack Spell",
		TypeLine:   "Sorcery",
		Owner:      caster.ID,
		Controller: caster.ID,
	})
	// Re-set up by casting properly.
	g = effectstest.NewTestGame(t)
	caster = g.Seats[0]
	spellID := effectstest.PushHandCard(g, caster)
	// Mutate the hand card in place to make it a proper instant
	// (sorcery-speed gate otherwise rejects outside main phase).
	for i := range caster.Hand.Cards {
		if caster.Hand.Cards[i].InstanceID == spellID {
			caster.Hand.Cards[i].TypeLine = "Instant"
			break
		}
	}
	if err := g.CastSpell(caster.ID, spellID, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}

	g.WithWriteLock(func() {
		if err := (effects.CounterTarget{StackID: spellID}).Apply(effectstest.CtxFor(g, &game.StackItem{})); err != nil {
			t.Fatalf("Apply: %v", err)
		}
	})

	if g.Stack.Contains(spellID) {
		t.Errorf("spell still on stack")
	}
	if !caster.Graveyard.Contains(spellID) {
		t.Errorf("countered spell not in caster's graveyard")
	}
}
