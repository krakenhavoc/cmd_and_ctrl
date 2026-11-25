package testss82m

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

// War of the Spark, batch 173 (#5520).

const (
	oldB173GatewayPlaza = "a6543f71-0326-4e1f-b58f-9ce325d5d036"
	oldB173Rioters      = "d750f1d1-a3f9-4180-a903-2bd3f4cb9674"
)

func TestOldB173CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{oldB173GatewayPlaza: "Gateway Plaza", oldB173Rioters: "Rubblebelt Rioters"} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB173GatewayPlazaIsPaidForOrSacrificed(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.AdvanceToMain(t, g)
		land := effectstest.PlayLandFromHand(t, g, "Gateway Plaza", oldB173GatewayPlaza)
		effectstest.PassPriorityAroundTable(t, g)
		if c, ok := effectstest.BattlefieldCard(g, land); !ok || !c.Tapped {
			t.Fatal("the Plaza did not enter tapped")
		}
		if pay {
			effectstest.DcMana(t, g, me.ID, "{C}")
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(land) != pay {
			t.Errorf("paid=%v: the Plaza on the battlefield = %v", pay, g.Battlefield.Contains(land))
		}
	}
}

func TestOldB173GatewayPlazaMakesManaOfAnyColor(t *testing.T) {
	spec, _ := effects.Lookup(oldB173GatewayPlaza)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{W|U|B|R|G}" || !spec.ManaAbilities[0].Cost.Tap {
		t.Errorf("mana abilities %+v, want one {T}: add one mana of any color", spec.ManaAbilities)
	}
}

func TestOldB173RioterAttackPumpsByTheGreatestPowerYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rioters := effectstest.B12Push(g, me.ID, "Rubblebelt Rioters", "Creature — Human Berserker", oldB173Rioters, 0, 4)
	effectstest.B12Creature(g, me.ID, "Big", "Creature — Ogre", 5, 5)
	effectstest.B12Creature(g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID, "Their Giant", "Creature — Giant", 9, 9)
	if !effectstest.HasEffectiveKeyword(t, g, rioters, "haste") {
		t.Error("no haste")
	}
	effectstest.BrAttack(t, g, rioters)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, rioters); got != 5 {
		t.Errorf("power %d, want 5 (the greatest power among your creatures; an opponent's 9 doesn't count)", got)
	}
}

func TestOldB173RioterAloneGetsNothingFromItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rioters := effectstest.B12Push(g, me.ID, "Rubblebelt Rioters", "Creature — Human Berserker", oldB173Rioters, 0, 4)
	effectstest.BrAttack(t, g, rioters)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, rioters); got != 0 {
		t.Errorf("power %d, want 0 (X is its own power, which is 0)", got)
	}
}
