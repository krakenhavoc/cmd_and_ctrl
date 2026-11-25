package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// the_celestial_toymaker_test.go — #6175. The attack trigger's two
// piles and the defending player's choice, and the end step that reads
// the turn's pile-and-guess tally.

const celestialToymakerOracle = "388fe92a-0bc0-4c0a-a7c6-910c5e71e342"

// ctmPileGuessCount is the #6175 turn tally as a card reads it. (The
// tally's own tests live in another test package, which is why the
// two-line reader is written out here too.)
func ctmPileGuessCount(g *game.Game) int {
	n := 0
	g.WithWriteLock(func() { n = g.PilesOrGuessesThisTurnForEffect() })
	return n
}

// ctmAttack puts a Toymaker under `me`, seeds the three cards it looks
// at and attacks `defender` with it, leaving the trigger resolved and
// the controller's split prompt open.
func ctmAttack(t *testing.T, g *game.Game, me *game.Player, defender uuid.UUID) (uuid.UUID, []uuid.UUID) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	toymaker := effectstest.B12Push(g, me.ID, "The Celestial Toymaker", "Legendary Creature — Rogue Artificer", celestialToymakerOracle, 2, 4)
	looked := effectstest.SeedLibrary(me, "CTM One", "CTM Two", "CTM Three")
	effectstest.StdB42Attack(t, g, defender, toymaker)
	effectstest.PassPriorityAroundTable(t, g)
	return toymaker, looked
}

func TestCelestialToymakerIsCompleteWithNoCaveat(t *testing.T) {
	spec, ok := effects.Lookup(celestialToymakerOracle)
	if !ok {
		t.Fatal("The Celestial Toymaker is not in the catalog")
	}
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("completeness %v, caveats %v; want Full with none", spec.Completeness, spec.Caveats)
	}
}

// The defending player takes the face-down pile: those cards go to the
// Toymaker's controller's hand and the face-up pile stays in exile.
func TestCelestialToymakerDefenderTakesTheFaceDownPile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	_, looked := ctmAttack(t, g, me, opp.ID)
	hand := me.Hand.Size()

	split := effectstest.LatestRevealPickFor(g, me.ID)
	if split == nil {
		t.Fatalf("the controller was not asked to separate the piles: %+v", g.PendingChoices)
	}
	if len(split.ChooseCards) != 3 {
		t.Fatalf("three cards to separate, got %d", len(split.ChooseCards))
	}
	// Two cards in the face-down pile, one in the face-up pile.
	if err := g.ResolveRevealPick(split.ID, me.ID, looked[:2]); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	for _, id := range looked {
		if z := effectstest.ZoneOfCard(g, id); z != game.ZoneExile {
			t.Fatalf("%s is in %v before the pile is chosen, want exile", id, z)
		}
	}

	pick := effectstest.LatestOptionPickFor(g, opp.ID)
	if pick == nil {
		t.Fatalf("the defending player was not asked to choose a pile: %+v", g.PendingChoices)
	}
	if len(pick.PickOptions) != 2 {
		t.Fatalf("%d piles offered, want 2", len(pick.PickOptions))
	}
	effectstest.AnswerOptionPick(t, g, opp.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)

	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("the chosen pile put %d cards into hand, want 2", got)
	}
	for _, id := range looked[:2] {
		if z := effectstest.ZoneOfCard(g, id); z != game.ZoneHand {
			t.Errorf("face-down pile card %s is in %v, want the hand", id, z)
		}
	}
	if z := effectstest.ZoneOfCard(g, looked[2]); z != game.ZoneExile {
		t.Errorf("the pile they left is in %v, want exile", z)
	}
	if n := ctmPileGuessCount(g); n != 1 {
		t.Errorf("tally %d after the Toymaker's own piles, want 1", n)
	}
}

// The other branch: the defending player takes the face-up pile.
func TestCelestialToymakerDefenderTakesTheFaceUpPile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	_, looked := ctmAttack(t, g, me, opp.ID)
	hand := me.Hand.Size()

	split := effectstest.LatestRevealPickFor(g, me.ID)
	if split == nil {
		t.Fatalf("no split prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveRevealPick(split.ID, me.ID, looked[:1]); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	effectstest.AnswerOptionPick(t, g, opp.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)

	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("the face-up pile put %d cards into hand, want 2", got)
	}
	if z := effectstest.ZoneOfCard(g, looked[0]); z != game.ZoneExile {
		t.Errorf("the face-down pile they left is in %v, want exile", z)
	}
}

// Every card in the face-down pile is one choice: separating none makes
// an empty pile, and the defending player may still take it.
func TestCelestialToymakerAnEmptyPileIsStillAPile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	_, looked := ctmAttack(t, g, me, opp.ID)
	hand := me.Hand.Size()
	split := effectstest.LatestRevealPickFor(g, me.ID)
	if split == nil {
		t.Fatalf("no split prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveRevealPick(split.ID, me.ID, nil); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	effectstest.AnswerOptionPick(t, g, opp.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("the empty pile put %d cards into hand, want 0", got-hand)
	}
	for _, id := range looked {
		if z := effectstest.ZoneOfCard(g, id); z != game.ZoneExile {
			t.Errorf("%s is in %v, want exile", id, z)
		}
	}
}

// The end step: 2 life per spell or ability that made a player guess or
// group objects into a pile this turn, for each opponent.
func TestCelestialToymakerEndStepDrainsTwoPerPileOrGuess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "The Celestial Toymaker", "Legendary Creature — Rogue Artificer", celestialToymakerOracle, 2, 4)
	effectstest.SeedFactOrFictionLibrary(me, "FF One", "FF Two", "FF Three", "FF Four", "FF Five")
	effectstest.CastCatalogSpell(t, g, "Fact or Fiction", "Instant", effectstest.FactOrFictionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	if split := effectstest.LatestRevealPickFor(g, opp.ID); split != nil {
		if err := g.ResolveRevealPick(split.ID, opp.ID, nil); err != nil {
			t.Fatalf("ResolveRevealPick: %v", err)
		}
	}
	if pick := effectstest.LatestOptionPickFor(g, me.ID); pick != nil {
		effectstest.AnswerOptionPick(t, g, me.ID, 0)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := ctmPileGuessCount(g); n != 1 {
		t.Fatalf("setup: tally %d, want 1", n)
	}

	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := lives[p.ID]
		if p.ID != me.ID {
			want -= 2
		}
		if p.Life != want {
			t.Errorf("%s has %d life after the end step, want %d", p.Name, p.Life, want)
		}
	}
}

// A turn with no pile and no guess loses nobody anything.
func TestCelestialToymakerEndStepLosesNothingWithoutAPileOrGuess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "The Celestial Toymaker", "Legendary Creature — Rogue Artificer", celestialToymakerOracle, 2, 4)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if p.Life != lives[p.ID] {
			t.Errorf("%s has %d life, want %d: nothing made a pile or a guess", p.Name, p.Life, lives[p.ID])
		}
	}
}
