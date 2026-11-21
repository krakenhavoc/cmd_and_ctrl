package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ward_poison_test.go — "Ward—Get five poison counters" (The Serpent
// Society, #4613, CR 702.21a, CR 118.12a), and the card's dies trigger.
//
// ward_test.go pins what every ward shares and ward_life_sacrifice_test.go
// the life and sacrifice costs. These pin what this cost changes: the payer
// is offered exactly five poison counters at any poison total, paying gives
// exactly five and lets the spell resolve, declining counters it, and the
// tenth counter loses the game as a state-based action (CR 704.5c).

const serpentSocietyOracle = "3688ad0d-653b-4d5d-a57e-444db679d83e"

const serpentSocietyType = "Legendary Creature — Human Snake Villain"

func serpentSocietyTable(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g, me, opp, society := effectstest.WardTable(t, "The Serpent Society", serpentSocietyType, serpentSocietyOracle)
	return g, me, opp, society
}

func TestPoisonWardIsOfferedToTheCasterWithTheCountDeclared(t *testing.T) {
	g, me, opp, society := serpentSocietyTable(t)
	effectstest.CastAtWardedCreature(t, g, opp, society)

	prompt := effectstest.PassUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt for the spell's controller")
	}
	if effectstest.ConfirmChoiceFor(g, me.ID) != nil {
		t.Error("the ward permanent's controller must not be asked to pay")
	}
	if prompt.PoisonCost != 5 {
		t.Errorf("prompt PoisonCost = %d, want 5", prompt.PoisonCost)
	}
	if prompt.LifeCost != 0 {
		t.Errorf("prompt LifeCost = %d, want 0", prompt.LifeCost)
	}
	if effectstest.PoisonOn(opp) != 0 {
		t.Errorf("the offer alone gave %d poison counters, want 0", effectstest.PoisonOn(opp))
	}
}

func TestPoisonWardDeclinedCountersTheSpellAndGivesNothing(t *testing.T) {
	g, _, opp, society := serpentSocietyTable(t)
	blade := effectstest.CastAtWardedCreature(t, g, opp, society)

	prompt := effectstest.PassUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt for the spell's controller")
	}
	if err := g.ResolveConfirm(prompt.ID, opp.ID, false); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(society) {
		t.Error("a declined ward must counter the spell, so the Society should live")
	}
	if !opp.Graveyard.Contains(blade) {
		t.Error("the countered spell should be in its owner's graveyard")
	}
	if effectstest.PoisonOn(opp) != 0 {
		t.Errorf("declining gave %d poison counters, want 0", effectstest.PoisonOn(opp))
	}
}

func TestPoisonWardPaidGivesExactlyFiveAndTheSpellResolves(t *testing.T) {
	g, _, opp, society := serpentSocietyTable(t)
	effectstest.CastAtWardedCreature(t, g, opp, society)

	prompt := effectstest.PassUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt for the spell's controller")
	}
	if err := g.ResolveConfirm(prompt.ID, opp.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.PoisonOn(opp) != 5 {
		t.Errorf("paying the ward: %d poison counters, want exactly 5", effectstest.PoisonOn(opp))
	}
	if opp.Eliminated {
		t.Error("five poison counters must not lose the game")
	}
	if g.Battlefield.Contains(society) {
		t.Error("the ward was paid, so the removal spell should have resolved")
	}
}

// A payer already at five who accepts gets to ten and loses at the next
// state-based check (CR 704.5c). The offer is still made: a player can always
// get counters, and taking them is a legal, if fatal, choice.
func TestPoisonWardPaidFromFiveReachesTenAndLoses(t *testing.T) {
	g, _, opp, society := serpentSocietyTable(t)
	g.WithWriteLock(func() {
		if err := g.AddPlayerCounterForEffect(opp.ID, game.CounterPoison, 5); err != nil {
			t.Fatalf("AddPlayerCounterForEffect: %v", err)
		}
	})
	effectstest.CastAtWardedCreature(t, g, opp, society)

	prompt := effectstest.PassUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("a payer at five poison must still be offered the payment")
	}
	if err := g.ResolveConfirm(prompt.ID, opp.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.PoisonOn(opp) != 10 {
		t.Errorf("poison = %d, want 10", effectstest.PoisonOn(opp))
	}
	if !opp.Eliminated {
		t.Error("ten poison counters must lose the game (CR 704.5c)")
	}
}

// A player at nine is still offered it, and declining keeps them alive.
func TestPoisonWardAtNineCanStillBeDeclined(t *testing.T) {
	g, _, opp, society := serpentSocietyTable(t)
	g.WithWriteLock(func() {
		if err := g.AddPlayerCounterForEffect(opp.ID, game.CounterPoison, 9); err != nil {
			t.Fatalf("AddPlayerCounterForEffect: %v", err)
		}
	})
	effectstest.CastAtWardedCreature(t, g, opp, society)

	prompt := effectstest.PassUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("a payer at nine poison must still be offered the payment")
	}
	if err := g.ResolveConfirm(prompt.ID, opp.ID, false); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PoisonOn(opp) != 9 || opp.Eliminated {
		t.Errorf("poison %d eliminated %v, want 9 and in the game", effectstest.PoisonOn(opp), opp.Eliminated)
	}
	if !g.Battlefield.Contains(society) {
		t.Error("the declined spell should have been countered")
	}
}

// The Society's own controller targeting it is not an opponent's spell, so
// no ward prompt and no poison.
func TestPoisonWardIgnoresItsControllersOwnSpells(t *testing.T) {
	g, me, _, society := serpentSocietyTable(t)
	effectstest.CastAtWardedCreature(t, g, me, society)
	for i := 0; i < 6 && !effectstest.StackFullyEmpty(g); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if effectstest.ConfirmChoiceFor(g, me.ID) != nil {
		t.Error("a ward prompt was raised for the controller's own spell")
	}
	if effectstest.PoisonOn(me) != 0 {
		t.Errorf("poison = %d, want 0", effectstest.PoisonOn(me))
	}
}

// --- the dies trigger ------------------------------------------------

func TestSerpentSocietyEdictsOnAnotherDeathtouchCreatureOfYoursDying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "The Serpent Society", serpentSocietyType, serpentSocietyOracle, false)
	rats := effectstest.PushCatalogPermanent(g, me.ID, "Typhoid Rats", "Creature — Rat", effectstest.TyphoidRatsOracle, false)
	bearA := effectstest.PushCatalogPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "", false)
	bearB := effectstest.PushCatalogPermanent(g, opp.ID, "Their Other Bear", "Creature — Bear", "", false)
	token := effectstest.PushCatalogPermanent(g, opp.ID, "Their Token", "Token Creature — Bear", "", false)

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(rats) })
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.SacrificeChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("the opponent was not asked to sacrifice")
	}
	offered := map[string]bool{}
	for _, id := range effectstest.SacrificeOptions(c) {
		offered[id.String()] = true
	}
	if offered[token.String()] {
		t.Error("a token was offered; the edict says nontoken")
	}
	if !offered[bearA.String()] || !offered[bearB.String()] {
		t.Errorf("both nontoken creatures should be offered, got %v", effectstest.SacrificeOptions(c))
	}
	effectstest.AnswerSacrifice(t, g, opp.ID, bearB)
	effectstest.PassPriorityAroundTable(t, g)

	if _, ok := effectstest.BattlefieldCard(g, bearB); ok {
		t.Error("the chosen creature survived")
	}
	if _, ok := effectstest.BattlefieldCard(g, bearA); !ok {
		t.Error("the unchosen creature was sacrificed")
	}
	if _, ok := effectstest.BattlefieldCard(g, token); !ok {
		t.Error("the token was sacrificed; the edict ignores tokens")
	}
}

// An opponent whose only creature is a token loses nothing.
func TestSerpentSocietyEdictIgnoresTokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "The Serpent Society", serpentSocietyType, serpentSocietyOracle, false)
	rats := effectstest.PushCatalogPermanent(g, me.ID, "Typhoid Rats", "Creature — Rat", effectstest.TyphoidRatsOracle, false)
	token := effectstest.PushCatalogPermanent(g, opp.ID, "Their Token", "Token Creature — Bear", "", false)

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(rats) })
	effectstest.PassPriorityAroundTable(t, g)

	if c := effectstest.SacrificeChoiceFor(g, opp.ID); c != nil {
		t.Errorf("the opponent was asked with only a token to offer: %v", effectstest.SacrificeOptions(c))
	}
	if _, ok := effectstest.BattlefieldCard(g, token); !ok {
		t.Error("the token was sacrificed")
	}
}

// It fires for another creature YOU control with deathtouch, and for nothing
// else: not the Society's own death, not a creature without deathtouch, not
// an opponent's deathtouch creature.
func TestSerpentSocietyDoesNotFireForTheWrongCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	society := effectstest.PushCatalogPermanent(g, me.ID, "The Serpent Society", serpentSocietyType, serpentSocietyOracle, false)
	bear := effectstest.PushCatalogPermanent(g, me.ID, "My Bear", "Creature — Bear", "", false)
	theirRats := effectstest.PushCatalogPermanent(g, opp.ID, "Their Rats", "Creature — Rat", effectstest.TyphoidRatsOracle, false)
	effectstest.PushCatalogPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "", false)
	effectstest.PushCatalogPermanent(g, opp.ID, "Their Other Bear", "Creature — Bear", "", false)

	g.WithWriteLock(func() {
		_ = g.SacrificePermanentForEffect(bear)
		_ = g.SacrificePermanentForEffect(theirRats)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.SacrificeChoiceFor(g, opp.ID); c != nil {
		t.Errorf("fired for a creature without deathtouch or one the opponent controls: %v", effectstest.SacrificeOptions(c))
	}

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(society) })
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.SacrificeChoiceFor(g, opp.ID); c != nil {
		t.Errorf("the Society's own death fired it (\"another\"): %v", effectstest.SacrificeOptions(c))
	}
}
