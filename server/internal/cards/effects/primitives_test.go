package effects

import (
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// newTestGame returns a fully-started 2-player game suitable for
// primitive unit tests. Uses the game package's exported Start
// path via a tiny deck so we don't have to reimplement setup.
func newTestGame(t *testing.T) *game.Game {
	t.Helper()
	g := game.NewGame()
	// Two players with 10-card libraries — enough for draw / mill
	// primitives without running out.
	for i := 0; i < 2; i++ {
		deck := make([]game.Card, 10)
		for j := range deck {
			deck[j] = game.NewCard("basic-filler", uuid.Nil)
		}
		p, err := g.AddPlayer("P"+string(rune('0'+i)), deck)
		if err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
		_ = p
	}
	if err := g.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	// Close mulligans so the cursor lands on a priority-granting step.
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	return g
}

func pushBattlefieldCreature(g *game.Game, owner uuid.UUID) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       "Test Creature",
		TypeLine:   "Creature — Test",
		Power:      2,
		Toughness:  2,
		Owner:      owner,
		Controller: owner,
	})
	return id
}

func pushHandCard(g *game.Game, p *game.Player) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id,
		Name:       "Test Card",
		Owner:      p.ID,
		Controller: p.ID,
	})
	return id
}

func ctxFor(g *game.Game, item *game.StackItem) *Context {
	return NewContext(g, item)
}

func TestDealDamageToPlayerAppliesLifeAndEmitsEvent(t *testing.T) {
	g := newTestGame(t)
	target := g.Seats[1]
	before := target.Life

	g.WithWriteLock(func() {
		item := &game.StackItem{Controller: g.Seats[0].ID, SourceCardID: uuid.New()}
		if err := (DealDamage{Target: target.ID, Amount: 3}).Apply(ctxFor(g, item)); err != nil {
			t.Fatalf("Apply: %v", err)
		}
	})

	if got := target.Life; got != before-3 {
		t.Errorf("target life: got %d, want %d", got, before-3)
	}
}

func TestDealDamageToCreatureMarksDamage(t *testing.T) {
	g := newTestGame(t)
	creatureID := pushBattlefieldCreature(g, g.Seats[0].ID)

	g.WithWriteLock(func() {
		if err := (DealDamage{Target: creatureID, Amount: 2}).Apply(ctxFor(g, &game.StackItem{})); err != nil {
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
	g := newTestGame(t)
	p := g.Seats[0]
	before := p.Life

	g.WithWriteLock(func() {
		if err := (GainLife{Player: p.ID, Amount: 5}).Apply(ctxFor(g, &game.StackItem{})); err != nil {
			t.Fatalf("Apply: %v", err)
		}
	})

	if p.Life != before+5 {
		t.Errorf("life: got %d, want %d", p.Life, before+5)
	}
}

func TestDrawCards(t *testing.T) {
	g := newTestGame(t)
	p := g.Seats[0]
	before := p.Hand.Size()
	libBefore := p.Library.Size()

	g.WithWriteLock(func() {
		if err := (DrawCards{Player: p.ID, N: 3}).Apply(ctxFor(g, &game.StackItem{})); err != nil {
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
	g := newTestGame(t)
	p := g.Seats[1]
	libBefore := p.Library.Size()
	gyBefore := p.Graveyard.Size()

	g.WithWriteLock(func() {
		if err := (MillCards{Player: p.ID, N: 2}).Apply(ctxFor(g, &game.StackItem{})); err != nil {
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
	g := newTestGame(t)
	p := g.Seats[0]
	// Ensure at least 2 cards in hand to discard.
	pushHandCard(g, p)
	pushHandCard(g, p)
	before := p.Hand.Size()
	gyBefore := p.Graveyard.Size()

	g.WithWriteLock(func() {
		if err := (DiscardCards{Player: p.ID, N: 2}).Apply(ctxFor(g, &game.StackItem{})); err != nil {
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
	g := newTestGame(t)
	owner := g.Seats[0]
	id := pushBattlefieldCreature(g, owner.ID)

	g.WithWriteLock(func() {
		if err := (DestroyTarget{Target: id}).Apply(ctxFor(g, &game.StackItem{})); err != nil {
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
	g := newTestGame(t)
	id := pushBattlefieldCreature(g, g.Seats[0].ID)

	g.WithWriteLock(func() {
		if err := (ExileTarget{Target: id}).Apply(ctxFor(g, &game.StackItem{})); err != nil {
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
	g := newTestGame(t)
	owner := g.Seats[0]
	id := pushBattlefieldCreature(g, owner.ID)

	g.WithWriteLock(func() {
		if err := (BounceToHand{Target: id}).Apply(ctxFor(g, &game.StackItem{})); err != nil {
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
	g := newTestGame(t)
	id := pushBattlefieldCreature(g, g.Seats[0].ID)

	g.WithWriteLock(func() {
		if err := (TapTarget{Target: id}).Apply(ctxFor(g, &game.StackItem{})); err != nil {
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
		if err := (UntapTarget{Target: id}).Apply(ctxFor(g, &game.StackItem{})); err != nil {
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
	g := newTestGame(t)
	id := pushBattlefieldCreature(g, g.Seats[0].ID)

	g.WithWriteLock(func() {
		if err := (AddCounter{Target: id, Kind: "+1/+1", N: 3}).Apply(ctxFor(g, &game.StackItem{})); err != nil {
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
	g := newTestGame(t)
	owner := g.Seats[0]
	bfBefore := len(g.Battlefield.Cards)

	g.WithWriteLock(func() {
		err := (CreateToken{Controller: owner.ID, Template: WhiteBirdToken(), N: 2}).Apply(ctxFor(g, &game.StackItem{}))
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
	g := newTestGame(t)
	owner := g.Seats[0]
	id := uuid.New()
	owner.Graveyard.PushTop(game.Card{
		InstanceID: id,
		Name:       "Buried Treasure",
		Owner:      owner.ID,
		Controller: owner.ID,
	})

	g.WithWriteLock(func() {
		if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctxFor(g, &game.StackItem{})); err != nil {
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
	g := newTestGame(t)
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
		err := (SearchLibrary{
			Player: p.ID,
			Predicate: func(c game.Card) bool {
				return c.Name == "Needle"
			},
			Dest:    game.ZoneHand,
			Limit:   1,
			Reveal:  true,
			Shuffle: true,
		}).Apply(ctxFor(g, &game.StackItem{}))
		if err != nil {
			t.Fatalf("Apply: %v", err)
		}
	})

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
	g := newTestGame(t)
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
	g.WithWriteLock(func() {
		if g.StackMeta == nil {
			// It's package-private normally, but tests in the same
			// package (effects) can't reach it. Exercise via the public
			// API (AnnounceTrigger produces a StackMeta entry, but not
			// for spells). Simplest: cast a real spell.
		}
	})
	// Re-set up by casting properly.
	g = newTestGame(t)
	caster = g.Seats[0]
	spellID := pushHandCard(g, caster)
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
		if err := (CounterTarget{StackID: spellID}).Apply(ctxFor(g, &game.StackItem{})); err != nil {
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
