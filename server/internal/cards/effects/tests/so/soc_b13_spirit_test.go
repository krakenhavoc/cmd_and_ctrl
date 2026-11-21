package testsso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b13_spirit_test.go — Spirit of Resilience (#4834).

func socB13Spirit(g *game.Game, me *game.Player) uuid.UUID {
	return effectstest.B12Push(g, me.ID, "Spirit of Resilience", "Creature — Spirit Warrior", socB13SpiritOracle, 2, 2)
}

func socB13GraveyardCard(p *game.Player, name, typeLine string, power, toughness int) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Power: power, Toughness: toughness,
		Owner: p.ID, Controller: p.ID})
	return id
}

func TestSocB13SpiritOfResilienceGrowsAndMayCopyTheCardThatLeft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spirit := socB13Spirit(g, me)
	cannon := socB13GraveyardCard(me, "Glass Cannon", "Creature — Elemental", 5, 1)
	theirs := socB13GraveyardCard(opp, "Their Card", "Creature — Elemental", 3, 3)

	// A card leaving an opponent's graveyard is not yours.
	socB11Move(t, g, opp.ID, theirs, game.ZoneGraveyard, game.ZoneHand)
	if effectstest.TriggerOnStack(g, spirit) != nil {
		t.Fatal("triggered for a card leaving an opponent's graveyard")
	}

	socB11Move(t, g, me.ID, cannon, game.ZoneGraveyard, game.ZoneHand)
	if effectstest.TriggerOnStack(g, spirit) == nil {
		t.Fatal("no trigger for a card leaving your graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, spirit, game.CounterPlusOne); got != 1 {
		t.Fatalf("+1/+1 counters %d, want 1 before the copy is chosen", got)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, cannon)
	card, _ := effectstest.BattlefieldCard(g, spirit)
	if card.Name != "Glass Cannon" {
		t.Fatalf("the Spirit is %q, want a copy of Glass Cannon", card.Name)
	}
	if got := effectstest.EffectivePowerWithCounters(t, g, spirit); got != 6 {
		t.Errorf("power %d as a 5/1 with a +1/+1 counter, want 6", got)
	}
	// Until end of turn.
	effectstest.EndTurn(t, g)
	card, _ = effectstest.BattlefieldCard(g, spirit)
	if card.Name != "Spirit of Resilience" {
		t.Errorf("the Spirit is still %q after the turn ended", card.Name)
	}
	if got := effectstest.CounterCount(g, spirit, game.CounterPlusOne); got != 1 {
		t.Errorf("the counter did not stay: %d", got)
	}
}

func TestSocB13SpiritOfResilienceMayDeclineTheCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spirit := socB13Spirit(g, me)
	cannon := socB13GraveyardCard(me, "Glass Cannon", "Creature — Elemental", 5, 1)
	socB11Move(t, g, me.ID, cannon, game.ZoneGraveyard, game.ZoneHand)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	card, _ := effectstest.BattlefieldCard(g, spirit)
	if card.Name != "Spirit of Resilience" {
		t.Errorf("the Spirit became %q after declining", card.Name)
	}
	if got := effectstest.CounterCount(g, spirit, game.CounterPlusOne); got != 1 {
		t.Errorf("counters %d, want 1", got)
	}
}

func TestSocB13SpiritOfResilienceOffersNoCopyOfALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spirit := socB13Spirit(g, me)
	forest := socB13GraveyardCard(me, "Forest", "Basic Land — Forest", 0, 0)
	socB11Move(t, g, me.ID, forest, game.ZoneGraveyard, game.ZoneHand)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, spirit, game.CounterPlusOne); got != 1 {
		t.Fatalf("counters %d, want 1", got)
	}
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a copy was offered with only a land card among those that left")
	}
}

func TestSocB13SpiritOfResilienceTriggersOnceForSeveralCardsAndCopiesAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spirit := socB13Spirit(g, me)
	rock := socB13GraveyardCard(me, "Sol Ring", "Artifact", 0, 0)
	forest := socB13GraveyardCard(me, "Forest", "Basic Land — Forest", 0, 0)
	cannon := socB13GraveyardCard(me, "Glass Cannon", "Creature — Elemental", 5, 1)
	g.WithWriteLock(func() {
		if err := g.ExileCardsKeyedForEffect([]uuid.UUID{rock, forest, cannon}, game.KeyedTail{}); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, spirit, game.CounterPlusOne); got != 1 {
		t.Fatalf("counters %d for three cards leaving at once, want 1", got)
	}
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil || len(pick.ChooseCards) != 2 {
		t.Fatalf("prompt %+v, want the artifact and the creature, not the land", pick)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, rock)
	card, _ := effectstest.BattlefieldCard(g, spirit)
	if card.Name != "Sol Ring" || !card.HasCardType("artifact") {
		t.Errorf("the Spirit is %q (%s), want a copy of Sol Ring", card.Name, card.TypeLine)
	}
}
