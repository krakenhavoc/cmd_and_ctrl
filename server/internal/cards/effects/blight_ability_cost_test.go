package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blight_ability_cost_test.go — #4684: "Blight N" as part of an
// activated ability's cost (CR 701.68a via CR 602.2b), end to end through
// the cards that print it. The validator's table is
// game/blight_ability_cost_test.go; the enumerator's
// legal/blight_ability_cost_test.go.

const (
	gristleGluttonOracle     = "b7163800-41fc-410a-9650-50ae90842170"
	stingSlingerOracle       = "945a4a4d-fdfd-4c39-9975-5b22d17a4e70"
	evershrikesGiftOracle    = "b5bc4adf-9105-4901-989c-63612314af5c"
	spiralIntoSolitudeOracle = "d8599f06-f030-4575-8497-307a4d7ec182"
)

// bacTable is the active player in their main phase with Gristle Glutton
// (1/3, untapped, not summoning sick) and a 3/3 Bear, and an opponent
// with a creature of their own.
func bacTable(t *testing.T) (g *game.Game, me, opp *game.Player, glutton, bear, theirs uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	glutton = b12Push(g, me.ID, "Gristle Glutton", "Creature — Goblin Scout", gristleGluttonOracle, 1, 3)
	bear = b12Push(g, me.ID, "Bear", "Creature — Bear", "", 3, 3)
	theirs = b12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 3, 3)
	return g, me, opp, glutton, bear, theirs
}

func TestBlightCostCardsAreFull(t *testing.T) {
	for _, oracle := range []string{gristleGluttonOracle, stingSlingerOracle, evershrikesGiftOracle, spiralIntoSolitudeOracle} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Fatalf("%s is not registered", oracle)
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: Completeness = %v, Caveats = %v, want Full and none", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// The activator names any creature they control, not only the source: the
// Bear takes the counter, the Glutton taps, and the loot resolves.
func TestGristleGluttonBlightsACreatureOfYourChoiceAndLoots(t *testing.T) {
	g, me, _, glutton, bear, _ := bacTable(t)
	pitch := handCardForTest(me, "Pitch Me", "Sorcery", "")
	library := me.Library.Size()

	b16Activate(t, g, me.ID, glutton, 0, game.ActivateAbilityParams{Strict: true, BlightIDs: []uuid.UUID{bear}})
	if got := minusCounters(g, bear); got != 1 {
		t.Errorf("-1/-1 counters on the blighted Bear: %d, want 1 (paid at announce)", got)
	}
	if got := minusCounters(g, glutton); got != 0 {
		t.Errorf("-1/-1 counters on the Glutton: %d, want 0", got)
	}
	if !b16Tapped(t, g, glutton) {
		t.Error("the {T} was not paid")
	}
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID, pitch)
	if !me.Graveyard.Contains(pitch) {
		t.Error("the chosen card was not discarded")
	}
	if me.Library.Size() != library-1 {
		t.Errorf("library %d after the loot, want %d (a card drawn)", me.Library.Size(), library-1)
	}
}

// "A creature you control" includes the source.
func TestGristleGluttonCanBlightItself(t *testing.T) {
	g, me, _, glutton, bear, _ := bacTable(t)
	b16Activate(t, g, me.ID, glutton, 0, game.ActivateAbilityParams{Strict: true, BlightIDs: []uuid.UUID{glutton}})
	if got := minusCounters(g, glutton); got != 1 {
		t.Errorf("-1/-1 counters on the Glutton: %d, want 1", got)
	}
	if got := minusCounters(g, bear); got != 0 {
		t.Errorf("-1/-1 counters on the Bear: %d, want 0", got)
	}
}

// With an empty hand nothing is discarded, so nothing is drawn ("if you
// do"), and the blight is still paid.
func TestGristleGluttonWithAnEmptyHandDrawsNothing(t *testing.T) {
	g, me, _, glutton, bear, _ := bacTable(t)
	g.WithWriteLock(func() {
		for me.Hand.Size() > 0 {
			me.Hand.PopTop()
		}
	})
	library := me.Library.Size()
	b16Activate(t, g, me.ID, glutton, 0, game.ActivateAbilityParams{Strict: true, BlightIDs: []uuid.UUID{bear}})
	passPriorityAroundTable(t, g)
	monSettle(t, g)
	if me.Library.Size() != library {
		t.Errorf("library %d, want %d: nothing was discarded, so nothing is drawn", me.Library.Size(), library)
	}
	if got := minusCounters(g, bear); got != 1 {
		t.Errorf("-1/-1 counters on the Bear: %d, want 1", got)
	}
}

// A refused activation costs nothing: no pick, two picks, an opponent's
// creature, a noncreature.
func TestBlightCostRefusalsPayNothing(t *testing.T) {
	g, me, _, glutton, bear, theirs := bacTable(t)
	rock := b12Push(g, me.ID, "Rock", "Artifact", "", 0, 0)
	for _, tc := range []struct {
		name string
		ids  []uuid.UUID
		want error
	}{
		{"no creature named", nil, game.ErrInvalidParam},
		{"two creatures named", []uuid.UUID{bear, glutton}, game.ErrInvalidParam},
		{"an opponent's creature", []uuid.UUID{theirs}, game.ErrCardCallerMismatch},
		{"a noncreature", []uuid.UUID{rock}, game.ErrNotACreature},
	} {
		err := g.ActivateCatalogAbility(me.ID, glutton, 0, game.ActivateAbilityParams{Strict: true, BlightIDs: tc.ids})
		if !errors.Is(err, tc.want) {
			t.Errorf("%s: %v, want %v", tc.name, err, tc.want)
		}
		if b16Tapped(t, g, glutton) {
			t.Fatalf("%s: a refused activation tapped the Glutton", tc.name)
		}
		for _, id := range []uuid.UUID{glutton, bear, theirs} {
			if got := minusCounters(g, id); got != 0 {
				t.Fatalf("%s: a refused activation put %d counters on %v", tc.name, got, id)
			}
		}
	}
}

// The pick is not a target: a hexproof, shrouded creature you control
// pays.
func TestBlightCostCanBlightAHexproofCreature(t *testing.T) {
	g, me, _, glutton, _, _ := bacTable(t)
	hex := b12Push(g, me.ID, "Warded", "Creature — Test", "", 3, 3)
	findBattlefieldCardByID(g, hex).Keywords = []string{"hexproof", "shroud"}
	b16Activate(t, g, me.ID, glutton, 0, game.ActivateAbilityParams{Strict: true, BlightIDs: []uuid.UUID{hex}})
	if got := minusCounters(g, hex); got != 1 {
		t.Errorf("-1/-1 counters on the hexproof creature: %d, want 1", got)
	}
}

// A creature the counters kill still paid: the 1/1 dies at the next
// state-based check and the ability resolves.
func TestBlightCostPaidByACreatureThatDies(t *testing.T) {
	g, me, _, glutton, _, _ := bacTable(t)
	token := b12Push(g, me.ID, "Goblin", "Creature — Goblin", "", 1, 1)
	pitch := handCardForTest(me, "Pitch Me", "Sorcery", "")
	b16Activate(t, g, me.ID, glutton, 0, game.ActivateAbilityParams{Strict: true, BlightIDs: []uuid.UUID{token}})
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID, pitch)
	if findBattlefieldCardByID(g, token) != nil {
		t.Error("the 1/1 survived a -1/-1 counter")
	}
	if !me.Graveyard.Contains(pitch) {
		t.Error("the ability did not resolve after its blight killed the creature that paid it")
	}
}

// CR 614.16: a cost's counters are not put by an effect. Doubling Season
// ("if an effect would put") leaves them alone; Vizier of Remedies ("if
// one or more -1/-1 counters would be put", no effect named) takes one
// away, and the cost is still paid.
func TestBlightCostIsNotAnEffectsPlacement(t *testing.T) {
	for _, tc := range []struct {
		name, card, typeLine, oracle string
		want                         int
	}{
		{"Doubling Season does not double a cost", "Doubling Season", "Enchantment", twDoublingSeasonOracle, 1},
		{"Vizier of Remedies takes the counter away", "Vizier of Remedies", "Creature — Human Monk", ccrVizierOracle, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			advanceToMain(t, g)
			ccrPushReplacement(g, me.ID, tc.card, tc.typeLine, tc.oracle)
			slinger := b12Push(g, me.ID, "Sting-Slinger", "Creature — Goblin Warrior", stingSlingerOracle, 3, 3)
			wall := b12Push(g, me.ID, "Wall", "Creature — Wall", "", 0, 10)
			life := opp.Life
			b06AddMana(me, "C", "R")
			b16Activate(t, g, me.ID, slinger, 0, game.ActivateAbilityParams{Strict: true, BlightIDs: []uuid.UUID{wall}})
			if got := minusCounters(g, wall); got != tc.want {
				t.Errorf("-1/-1 counters: %d, want %d", got, tc.want)
			}
			passPriorityAroundTable(t, g)
			if got := life - opp.Life; got != 2 {
				t.Errorf("damage to an opponent: %d, want 2", got)
			}
		})
	}
}

// Sting-Slinger deals 2 damage to each opponent, and not to its
// controller.
func TestStingSlingerDealsTwoToEachOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	slinger := b12Push(g, me.ID, "Sting-Slinger", "Creature — Goblin Warrior", stingSlingerOracle, 3, 3)
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}
	b06AddMana(me, "C", "R")
	b16Activate(t, g, me.ID, slinger, 0, game.ActivateAbilityParams{Strict: true, BlightIDs: []uuid.UUID{slinger}})
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := 2
		if p.ID == me.ID {
			want = 0
		}
		if got := before[p.ID] - p.Life; got != want {
			t.Errorf("seat %v lost %d, want %d", p.ID, got, want)
		}
	}
	if got := minusCounters(g, slinger); got != 1 {
		t.Errorf("-1/-1 counters on Sting-Slinger: %d, want 1", got)
	}
}

// A graveyard activation blights a creature on the battlefield, and with
// no creature to blight it can't be activated (CR 701.68b).
func TestEvershrikesGiftReturnsFromTheGraveyardByBlightingACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	gift := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: gift, Name: "Evershrike's Gift", TypeLine: "Enchantment — Aura",
		OracleID: evershrikesGiftOracle, Owner: me.ID, Controller: me.ID})
	b06AddMana(me, "C", "W")

	err := g.ActivateCatalogAbility(me.ID, gift, 0, game.ActivateAbilityParams{Strict: true})
	if !errors.Is(err, game.ErrCantPayCounterCost) {
		t.Fatalf("with no creature: %v, want ErrCantPayCounterCost", err)
	}
	if !me.Graveyard.Contains(gift) {
		t.Fatal("a refused activation moved the card")
	}

	wall := b12Push(g, me.ID, "Wall", "Creature — Wall", "", 0, 4)
	b16Activate(t, g, me.ID, gift, 0, game.ActivateAbilityParams{Strict: true, BlightIDs: []uuid.UUID{wall}})
	if got := minusCounters(g, wall); got != 2 {
		t.Errorf("-1/-1 counters on the Wall: %d, want 2", got)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(gift) {
		t.Error("Evershrike's Gift did not return to hand")
	}
}

// Spiral into Solitude: the blight and the sacrifice are both paid, and
// the creature the Aura enchanted is exiled.
func TestSpiralIntoSolitudeBlightsSacrificesAndExiles(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	victim := b12Push(g, opp.ID, "Ogre", "Creature — Ogre", "", 3, 3)
	mine := b12Push(g, me.ID, "Bear", "Creature — Bear", "", 3, 3)
	spiral := pushCatalogPermanent(g, me.ID, "Spiral into Solitude", "Enchantment — Aura", spiralIntoSolitudeOracle, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == spiral {
				g.Battlefield.Cards[i].AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: victim}
			}
		}
	})
	b06AddMana(me, "C", "W")
	b16Activate(t, g, me.ID, spiral, 0, game.ActivateAbilityParams{Strict: true, BlightIDs: []uuid.UUID{mine}})
	if got := minusCounters(g, mine); got != 1 {
		t.Errorf("-1/-1 counters on the Bear: %d, want 1", got)
	}
	if findBattlefieldCardByID(g, spiral) != nil {
		t.Error("the Aura was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(victim) {
		t.Error("the enchanted creature was not exiled")
	}
}

// The announcement's blight pick survives a restore with the ability on
// the stack: the counters are already on, and the ability resolves.
func TestBlightCostAbilityOnTheStackIsARestorePoint(t *testing.T) {
	g, me, _, glutton, bear, _ := bacTable(t)
	handCardForTest(me, "Pitch Me", "Sorcery", "")
	b16Activate(t, g, me.ID, glutton, 0, game.ActivateAbilityParams{Strict: true, BlightIDs: []uuid.UUID{bear}})
	g = restoreRoundTrip(t, g, true)
	if got := minusCounters(g, bear); got != 1 {
		t.Errorf("-1/-1 counters on the Bear after a restore: %d, want 1", got)
	}
	passPriorityAroundTable(t, g)
	discardFromHand(t, g, me.ID)
}
