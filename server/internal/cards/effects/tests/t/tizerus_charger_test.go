package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tizerus_charger_test.go — Tizerus Charger (#6154): "This creature escapes
// with your choice of a +1/+1 counter or a flying counter on it"
// (CR 702.138c).

const tizerusChargerOracle = "8f94754d-fb2a-4634-a1de-049c82fa6f30"

func TestTizerusChargerIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(tizerusChargerOracle)
	if !ok || spec.Name != "Tizerus Charger" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
	offers := game.AlternativeCostsOfferedFromZone(tizerusChargerOracle, game.ZoneGraveyard)
	if len(offers) != 1 || offers[0].Key != "escape" || offers[0].ManaCost != "{4}{B}" ||
		offers[0].ExileFromGraveyard == nil || offers[0].ExileFromGraveyard.Min != 5 || offers[0].EntersWithCounterName != "" {
		t.Fatalf("graveyard offers = %+v, want one escape for {4}{B} and five cards, with no fixed counter", offers)
	}
}

// escapeTizerus escapes the Charger and answers its counter question with
// option `pick`.
func escapeTizerus(t *testing.T, pick int) (*game.Game, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.SeedGraveyardCard(t, g, "Tizerus Charger", "Creature — Pegasus", tizerusChargerOracle)
	// The printed body, as the deck importer stamps it: a 3/2.
	g.WithWriteLock(func() {
		for i := range active.Graveyard.Cards {
			if c := &active.Graveyard.Cards[i]; c.InstanceID == id {
				c.Power, c.Toughness, c.ManaCost = 3, 2, "{2}{B}"
			}
		}
	})
	pay := effectstest.SeedGraveyardFodder(t, g, 5)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "escape", AltCostIDs: pay,
	}); err != nil {
		t.Fatalf("escape cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceOptionPick)
	if ask == nil || ask.Chooser != active.ID || len(ask.PickOptions) != 2 {
		t.Fatalf("the escaped Charger asked %+v, want its controller to choose between two counters", ask)
	}
	if err := g.ResolveOptionPick(ask.ID, active.ID, pick); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the escaped Charger is not on the battlefield")
	}
	return g, id
}

func TestTizerusChargerEscapesWithTheChosenFlyingCounter(t *testing.T) {
	g, id := escapeTizerus(t, 1)
	c := effectstest.FindCardAnywhere(t, g, id)
	if c.Counters[game.CounterFlying] != 1 || c.Counters[game.CounterPlusOne] != 0 {
		t.Errorf("counters = %v, want one flying counter and no +1/+1", c.Counters)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, id, "flying") {
		t.Error("the flying counter did not give the Charger flying")
	}
}

func TestTizerusChargerEscapesWithTheChosenPlusOneCounter(t *testing.T) {
	g, id := escapeTizerus(t, 0)
	c := effectstest.FindCardAnywhere(t, g, id)
	if c.Counters[game.CounterPlusOne] != 1 || c.Counters[game.CounterFlying] != 0 {
		t.Errorf("counters = %v, want one +1/+1 counter and no flying", c.Counters)
	}
	if p, tough := effectstest.PtOf(t, g, id); p != 4 || tough != 3 {
		t.Errorf("escaped Charger is %d/%d, want 4/3", p, tough)
	}
}

// Cast from hand it did not escape: no question, no counter.
func TestTizerusChargerCastFromHandEntersWithNeither(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Tizerus Charger", TypeLine: "Creature — Pegasus",
		OracleID: tizerusChargerOracle, ManaCost: "{2}{B}", Power: 3, Toughness: 2, Owner: active.ID, Controller: active.ID})
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast from hand: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the hand-cast Charger is not on the battlefield")
	}
	if ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceOptionPick); ask != nil {
		t.Fatalf("a hand-cast Charger asked %+v", ask)
	}
	c := effectstest.FindCardAnywhere(t, g, id)
	if c.Counters[game.CounterPlusOne] != 0 || c.Counters[game.CounterFlying] != 0 {
		t.Errorf("a hand-cast Charger has counters %v", c.Counters)
	}
	if effectstest.HasKeywordOnBattlefield(t, g, id, "flying") {
		t.Error("a hand-cast Charger has flying")
	}
}
