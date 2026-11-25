package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b250_cards_test.go — Amonkhet cards, S82 OLD batch 250 (#5535).
// Every test plays the real catalog entry through the engine.

const (
	b250DesiccatedNaga = "ab172e3e-e71e-4627-9b9a-a9295b891533"
	b250Bontu          = "c2679c32-a23d-4953-bb48-b46580439ce2"
	b250ZenithSeeker   = "a0fdc716-31b9-47f6-8639-de8bb7e3629f"
)

func TestB250DesiccatedNagaDrainsOnlyWithALiliana(t *testing.T) {
	for _, liliana := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		naga := effectstest.B12Push(g, me.ID, "Desiccated Naga", "Creature — Zombie Snake", b250DesiccatedNaga, 3, 2)
		kind := "Jace"
		if liliana {
			kind = "Liliana"
		}
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: kind + " Test", TypeLine: "Legendary Planeswalker — " + kind,
			Counters: map[string]int{game.CounterLoyalty: 4}, Owner: me.ID, Controller: me.ID,
		})
		effectstest.AdvanceToMain(t, g)
		effectstest.FillPool(me, 3)
		effectstest.AddPool(t, g, me.ID, "{B}")
		myLife, theirLife := me.Life, opp.Life
		err := g.ActivateCatalogAbility(me.ID, naga, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
		if !liliana {
			if err == nil {
				t.Error("activated without a Liliana planeswalker")
			}
			continue
		}
		if err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if opp.Life != theirLife-2 || me.Life != myLife+2 {
			t.Errorf("life: opponent %d (was %d), you %d (was %d); want -2 and +2", opp.Life, theirLife, me.Life, myLife)
		}
	}
}

func TestB250BontuNeedsACreatureToHaveDiedAndDrainsOnSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bontu := effectstest.B12Push(g, me.ID, "Bontu the Glorified", "Legendary Creature — God", b250Bontu, 4, 6)
	effectstest.AssertKeywords(t, g, bontu, "menace", "indestructible")
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	effectstest.AdvanceToMain(t, g)
	if c, ok := g.LookupCardForEffect(bontu); !ok || !game.Restricted(&c, game.CantAttack) || !game.Restricted(&c, game.CantBlock) {
		t.Fatal("before anything has died Bontu can neither attack nor block")
	}
	effectstest.FillPool(me, 1)
	effectstest.AddPool(t, g, me.ID, "{B}")
	myLife, theirLife := me.Life, opp.Life
	if err := g.ActivateCatalogAbility(me.ID, bontu, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerScryIfAsked(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	if opp.Life != theirLife-1 || me.Life != myLife+1 {
		t.Errorf("life: opponent %d (was %d), you %d (was %d); want -1 and +1", opp.Life, theirLife, me.Life, myLife)
	}
	if c, ok := g.LookupCardForEffect(bontu); !ok || game.Restricted(&c, game.CantAttack) || game.Restricted(&c, game.CantBlock) {
		t.Error("after a creature died under your control Bontu is free")
	}
}

func TestB250ZenithSeekerGrantsFlyingOnDiscardAndCycle(t *testing.T) {
	for _, cycle := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.B12Push(g, me.ID, "Zenith Seeker", "Creature — Bird Wizard", b250ZenithSeeker, 2, 2)
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		if cycle {
			effectstest.FillPool(me, 2)
			effectstest.CycleFromHand(t, g, "Barkhide Mauler", "Creature — Beast", "84a18b1f-5f9f-4e2b-87cd-c437bdf4d278", "")
		} else {
			effectstest.DrcDiscard(t, g, me, "Spare", "Instant")
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.PickTriggerTarget(t, g, me.ID, bear)
		if effectstest.LatestPickTarget(g, me.ID) != nil {
			t.Errorf("cycle=%v: the ability triggered more than once for one card", cycle)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
			t.Errorf("cycle=%v: the Bear did not gain flying", cycle)
		}
	}
}
