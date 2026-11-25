package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const oldB162ConjuringOracle = "a0a79e9e-38cb-45d8-a7ec-7cf10ef7e5fc"

func oldB162Push(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: p.ID, Controller: p.ID})
	return id
}

func TestOldB162CollectedConjuringCastsTwoSmallSorceriesAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	deep := oldB162Push(me, "Deep Card", "Creature — Bear", "{1}{G}")
	creature := oldB162Push(me, "Exiled Creature", "Creature — Bear", "{1}{G}")
	instant := oldB162Push(me, "Exiled Instant", "Instant", "{R}")
	big := oldB162Push(me, "Big Sorcery", "Sorcery", "{3}{R}")
	s3 := oldB162Push(me, "Sorcery C", "Sorcery", "{1}{R}")
	s2 := oldB162Push(me, "Sorcery B", "Sorcery", "{1}{U}{R}")
	s1 := oldB162Push(me, "Sorcery A", "Sorcery", "{U}{R}")
	effectstest.CastCatalogSpell(t, g, "Collected Conjuring", "Sorcery", oldB162ConjuringOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	offer := effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, s1, s2, s3)
	for _, id := range []uuid.UUID{big, instant, creature} {
		for _, o := range offer.DiscardOptions {
			if o == id {
				t.Fatalf("an ineligible card %s is offered", id)
			}
		}
	}
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, s1, game.CastSpellParams{}); err != nil {
		t.Fatalf("first cast: %v", err)
	}
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, s2, game.CastSpellParams{}); err != nil {
		t.Fatalf("second cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
		if err := effectstest.CastThroughResolutionCast(t, g, me.ID, s3, game.CastSpellParams{}); err == nil {
			t.Error("a third sorcery was cast")
		}
		effectstest.DeclineResolutionCast(t, g, me.ID)
		effectstest.PassPriorityAroundTable(t, g)
	}
	for _, id := range []uuid.UUID{s1, s2} {
		if me.Library.Contains(id) {
			t.Errorf("a cast spell is in the library")
		}
	}
	for _, id := range []uuid.UUID{s3, big, instant, creature} {
		if !me.Library.Contains(id) {
			t.Errorf("an uncast exiled card is in %q, not the library", effectstest.ZoneOf(g, id))
		}
	}
	if len(g.Exile.Cards) != 0 {
		t.Errorf("%d cards still in exile", len(g.Exile.Cards))
	}
	if !me.Library.Contains(deep) {
		t.Error("the seventh card left the library")
	}
	// The deep card stays on top; the six uncast cards went to the bottom.
	if top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID; top != deep {
		t.Errorf("the top card is %v, want the seventh card", top)
	}
}

func TestOldB162CollectedConjuringWithNothingToCastBottomsEverything(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var six []uuid.UUID
	for i := 0; i < 6; i++ {
		six = append(six, oldB162Push(me, "Exiled Creature", "Creature — Bear", "{1}{G}"))
	}
	deep := oldB162Push(me, "Deep Card", "Creature — Bear", "{1}{G}")
	_ = deep
	effectstest.CastCatalogSpell(t, g, "Collected Conjuring", "Sorcery", oldB162ConjuringOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range six {
		if !me.Library.Contains(id) {
			t.Errorf("an exiled card is in %q, not the library", effectstest.ZoneOf(g, id))
		}
	}
}
