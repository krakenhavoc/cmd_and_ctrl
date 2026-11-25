package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const oldB163RecombinerOracle = "ecf09f2d-a886-4ec8-8ff6-fe15d87b1375"

func TestOldB163ScrapyardRecombinerEntersWithTwoCountersAndTutorsAConstruct(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	relic := effectstest.B12Permanent(g, me.ID, "Relic", "Artifact")
	construct := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: construct, Name: "Golem Construct", TypeLine: "Artifact Creature — Construct", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	decoy := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: decoy, Name: "Plain Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	id := effectstest.CastCatalogSpell(t, g, "Scrapyard Recombiner", "Artifact Creature — Construct", oldB163RecombinerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, id, game.CounterPlusOne); got != 2 {
		t.Fatalf("%d +1/+1 counters on entering, want 2", got)
	}
	// It has been under its controller's control since the turn began only in the next test; clear sickness here.
	effectstest.Unsick(g, id)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{relic}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(relic) {
		t.Fatal("the artifact was not sacrificed as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Plain Bear") != uuid.Nil {
		t.Error("a non-Construct card was offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Golem Construct")
	if !me.Hand.Contains(construct) {
		t.Errorf("the Construct is in %q, not the hand", effectstest.ZoneOf(g, construct))
	}
}

func TestOldB163ScrapyardRecombinerNeedsAnArtifactToSacrificeAndNoSummoningSickness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Scrapyard Recombiner", "Artifact Creature — Construct", oldB163RecombinerOracle, true)
	relic := effectstest.B12Permanent(g, me.ID, "Relic", "Artifact")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{relic}}); err == nil {
		t.Fatal("activated {T} while summoning sick")
	}
	effectstest.Unsick(g, id)
	nonArtifact := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{nonArtifact}}); err == nil {
		t.Fatal("a nonartifact paid \"sacrifice an artifact\"")
	}
}
