package testss82m

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const oldB162GlacialOracle = "92359752-74aa-499e-b2db-0342d7596691"

// oldB162Top puts a named card on top of the library, keeping what is there.
func oldB162Top(p *game.Player, name, typeLine, manaCost string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost, Owner: p.ID, Controller: p.ID})
	return id
}

func TestOldB162GlacialRevelationTakesChosenSnowPermanentsAndGravesTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	// Pushed bottom first: the last pushed is the top card.
	plain := oldB162Top(me, "Plain Bear", "Creature — Bear", "{1}{G}")
	snowSpell := oldB162Top(me, "Snow Bolt", "Snow Instant", "{R}")
	snowCreature := oldB162Top(me, "Snow Yeti", "Snow Creature — Yeti", "{2}{G}")
	snowLandB := oldB162Top(me, "Snow-Covered Swamp", "Basic Snow Land — Swamp", "")
	plainLand := oldB162Top(me, "Forest", "Basic Land — Forest", "")
	snowLandA := oldB162Top(me, "Snow-Covered Forest", "Basic Snow Land — Forest", "")
	deeper := oldB162Top(me, "Deeper Card", "Creature — Bear", "{1}{G}")
	_ = deeper
	// "Deeper Card" was pushed last, so it is the top: the six revealed are it down to Snow Bolt; Plain Bear stays.
	effectstest.CastCatalogSpell(t, g, "Glacial Revelation", "Sorcery", oldB162GlacialOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, snowLandA, snowCreature)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{snowLandA, snowCreature} {
		if !me.Hand.Contains(id) {
			t.Errorf("a chosen snow permanent is in %q, not the hand", effectstest.ZoneOf(g, id))
		}
	}
	for _, id := range []uuid.UUID{snowLandB, plainLand, snowSpell, deeper} {
		if !me.Graveyard.Contains(id) {
			t.Errorf("an untaken revealed card is in %q, not the graveyard", effectstest.ZoneOf(g, id))
		}
	}
	if !me.Library.Contains(plain) {
		t.Errorf("the seventh card is in %q, not the library", effectstest.ZoneOf(g, plain))
	}
}

func TestOldB162GlacialRevelationOffersOnlySnowPermanentCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	snowSpell := oldB162Top(me, "Snow Bolt", "Snow Instant", "{R}")
	plainLand := oldB162Top(me, "Forest", "Basic Land — Forest", "")
	snowLand := oldB162Top(me, "Snow-Covered Forest", "Basic Snow Land — Forest", "")
	effectstest.CastCatalogSpell(t, g, "Glacial Revelation", "Sorcery", oldB162GlacialOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no choose-cards prompt")
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != snowLand {
		t.Fatalf("offered %v, want only the snow land", pick.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(snowLand) || !me.Graveyard.Contains(snowLand) || !me.Graveyard.Contains(snowSpell) || !me.Graveyard.Contains(plainLand) {
		t.Error("taking none should send every revealed card to the graveyard")
	}
}
