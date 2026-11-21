package testsds

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const evercoatUrsineOracle = "65c32889-3d70-449f-af76-05385c78c5a9"

func TestEvercoatUrsineIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(evercoatUrsineOracle)
	if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Evercoat Ursine: %+v %v", spec, ok)
	}
}

// evercoatUrsineTable casts the Ursine and answers both hideaway picks: the
// second card of the first three and the second card of the next three.
func evercoatUrsineTable(t *testing.T) (g *game.Game, me, opp *game.Player, ursine uuid.UUID, first, second uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	top := effectstest.StackTopLibrary(me, "A", "B", "C", "D", "E", "F")
	card := game.NewCard("Evercoat Ursine", me.ID)
	card.OracleID = evercoatUrsineOracle
	card.TypeLine = "Creature — Elemental Bear"
	card.ManaCost = "{4}{G}"
	card.Power, card.Toughness = 6, 5
	me.Hand.PushTop(card)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, top[1])
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, top[4])
	effectstest.PassPriorityAroundTable(t, g)
	return g, me, opp, card.InstanceID, top[1], top[4]
}

func TestEvercoatUrsineHidesTwoSeparateCards(t *testing.T) {
	g, _, _, ursine, first, second := evercoatUrsineTable(t)
	hidden := effectstest.HiddenCardsIn(g)
	if len(hidden) != 2 {
		t.Fatalf("hidden cards = %d, want 2 (two hideaway abilities)", len(hidden))
	}
	got := map[uuid.UUID]bool{}
	for _, h := range hidden {
		got[h.InstanceID] = true
		if h.HiddenBy.ID != ursine {
			t.Errorf("hidden card linked to %s, want the Ursine", h.HiddenBy.ID)
		}
	}
	if !got[first] || !got[second] {
		t.Errorf("hidden = %v, want the two picks %s and %s", got, first, second)
	}
}

func TestEvercoatUrsineCombatDamagePlaysOneOfThemAndLeavesTheOther(t *testing.T) {
	g, me, opp, ursine, first, second := evercoatUrsineTable(t)
	// A creature that has been on the battlefield since the start of the turn.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == ursine {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(ursine, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, second)
	if effectstest.PlayHiddenGrant(g, second) == nil {
		t.Error("the chosen card has no free play")
	}
	if effectstest.PlayHiddenGrant(g, first) != nil {
		t.Error("the card that was not chosen was granted a free play")
	}
	if len(effectstest.HiddenCardsIn(g)) != 2 {
		t.Errorf("hidden cards = %d, want both still hidden until one is cast", len(effectstest.HiddenCardsIn(g)))
	}
}

func TestEvercoatUrsineWithNothingExiledDoesNotTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ursine := effectstest.PushCatalogPermanent(g, me.ID, "Evercoat Ursine", "Creature — Elemental Bear", evercoatUrsineOracle, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == ursine {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 6, 5
			}
		}
	})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(ursine, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 {
		t.Errorf("a choice is open with nothing exiled: %+v", g.PendingChoices)
	}
}
