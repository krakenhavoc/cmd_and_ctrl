package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch37_test.go — card-level coverage for the card-coverage
// roadmap's batch 37 (#400, `edhrec_rank` 3853–3952). One test per
// observable behaviour, driven through a real cast, activation or
// attack rather than by calling primitives.
//
// The batch is the first worked after the 2026-09-18 re-triage, so
// two of the four cards here come from a group the issue filed as
// BLOCKED: Warden of Evos Isle under cost modification (#93, shipped
// S28). Its test is the one that matters — it is the evidence that
// the re-sorted ready group is real and not an arithmetic exercise.

const (
	b37HopToItOracle           = "e8a9350a-07c1-47ed-8c4f-88e4b3b17545"
	b37OpenTheGravesOracle     = "28778958-a1f9-4fea-b551-c193d1257f18"
	b37SkyknightVanguardOracle = "9180f77f-c288-4a24-a35d-a16270b3d737"
	b37WardenOfEvosIsleOracle  = "f424b5e9-8f02-4491-a7d8-c7e088611c6a"
)

func TestB37HopToItMakesThreeRabbits(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b18CastFromHand(t, g, me, "Hop to It", "Sorcery", b37HopToItOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if got := countTokensControlled(g, me.ID, "Rabbit"); got != 3 {
		t.Errorf("Hop to It makes three Rabbits, got %d", got)
	}
}

func TestB37OpenTheGravesMakesAZombieForANontokenDeath(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b18CastFromHand(t, g, me, "Open the Graves", "Enchantment", b37OpenTheGravesOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)

	victim := b20CastCreature(t, g, me, "Bear", "Creature — Bear", "", 2, 2)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victim) })
	passPriorityAroundTable(t, g)
	if got := countTokensControlled(g, me.ID, "Zombie"); got != 1 {
		t.Fatalf("a nontoken creature died: want one Zombie, got %d", got)
	}
}

// The nontoken guard is the card: without it the Zombie feeds itself
// and any sacrifice outlet is an infinite loop.
func TestB37OpenTheGravesIgnoresATokenDeath(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b18CastFromHand(t, g, me, "Open the Graves", "Enchantment", b37OpenTheGravesOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)

	b18CastFromHand(t, g, me, "Hop to It", "Sorcery", b37HopToItOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	rabbits := countTokensControlled(g, me.ID, "Rabbit")
	if rabbits != 3 {
		t.Fatalf("setup: want three Rabbits to kill, got %d", rabbits)
	}
	var one uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.HasSubtype("Rabbit") {
			one = c.InstanceID
			break
		}
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(one) })
	passPriorityAroundTable(t, g)
	if got := countTokensControlled(g, me.ID, "Zombie"); got != 0 {
		t.Errorf("a TOKEN died, so Open the Graves stays silent; got %d Zombies", got)
	}
}

func TestB37SkyknightVanguardMakesATappedAttackingSoldier(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vanguard := b20CastCreature(t, g, me, "Skyknight Vanguard", "Creature — Human Knight", b37SkyknightVanguardOracle, 2, 2)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == vanguard {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})

	declareAttack(t, g, opp.ID, vanguard)
	passPriorityAroundTable(t, g)

	tokens := b33AttackingTokensNamed(g, me.ID, opp.ID, "Soldier")
	if len(tokens) != 1 {
		t.Fatalf("one Soldier, tapped and attacking the same defender; got %d", len(tokens))
	}
	c, ok := g.LookupCardForEffect(tokens[0])
	if !ok {
		t.Fatal("the Soldier is on the battlefield")
	}
	if !c.Tapped {
		t.Error("the Soldier enters TAPPED and attacking")
	}
}

// The re-triage's evidence card. Its batch issue files it under cost
// modification, which shipped as #93 in S28 — so this asserts the
// discount really applies, and that it applies only to what the card
// says.
func TestB37WardenOfEvosIsleDiscountsOnlyCreatureSpellsWithFlying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	warden := b20CastCreature(t, g, me, "Warden of Evos Isle", "Creature — Bird Wizard", b37WardenOfEvosIsleOracle, 2, 2)
	passPriorityAroundTable(t, g)
	if _, ok := g.LookupCardForEffect(warden); !ok {
		t.Fatal("setup: the Warden is on the battlefield")
	}

	flier := game.Card{
		InstanceID: uuid.New(), Name: "Test Flier", TypeLine: "Creature — Bird",
		ManaCost: "{3}{U}", Keywords: []string{"flying"},
		Owner: me.ID, Controller: me.ID,
	}
	ground := flier
	ground.InstanceID = uuid.New()
	ground.Name = "Test Groundling"
	ground.TypeLine = "Creature — Bear"
	ground.Keywords = nil

	bolt := flier
	bolt.InstanceID = uuid.New()
	bolt.Name = "Test Bolt"
	bolt.TypeLine = "Instant"
	bolt.Keywords = []string{"flying"} // nonsense on purpose: not a creature spell

	for _, row := range []struct {
		card game.Card
		want int
		why  string
	}{
		{flier, 2, "a creature spell with flying costs {1} less: {3}{U} prices its generic 3 down to 2"},
		{ground, 3, "a creature spell without flying pays its full generic 3"},
		{bolt, 3, "a noncreature spell pays full price even carrying the keyword"},
	} {
		got := b37GenericAfterModifiers(t, g, me, row.card)
		if got != row.want {
			t.Errorf("%s: generic %d, want %d", row.why, got, row.want)
		}
	}
}

// b37GenericAfterModifiers prices `card` for `p` through the engine's
// own modifier pipeline and reports the generic component, which is
// the only part a reduction may spend against (CR 601.2f).
func b37GenericAfterModifiers(t *testing.T, g *game.Game, p *game.Player, card game.Card) int {
	t.Helper()
	base, err := game.ParseCost(card.ManaCost)
	if err != nil {
		t.Fatalf("ParseCost %s: %v", card.ManaCost, err)
	}
	priced, err := g.ApplyCostModifiersForEffect(base, game.CostQuery{
		Game:       g,
		Card:       card,
		Controller: p.ID,
	})
	if err != nil {
		t.Fatalf("ApplyCostModifiersForEffect %s: %v", card.Name, err)
	}
	return priced.Generic
}
