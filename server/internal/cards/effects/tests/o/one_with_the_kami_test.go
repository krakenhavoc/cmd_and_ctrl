package testso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestOneWithTheKamiIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.OneWithTheKamiOracle)
	if !ok || spec.Name != "One with the Kami" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("registration: %+v", spec)
	}
}

func TestOneWithTheKamiMakesSpiritsEqualToTheEnchantedCreaturesPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.BigBear(g, me.ID, 3, false)
	effectstest.AuraCast(t, g, "One with the Kami", effectstest.OneWithTheKamiOracle, bear)
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountTokensControlled(g, me.ID, "Spirit"); n != 3 {
		t.Errorf("Spirits = %d, want 3", n)
	}
}

func TestOneWithTheKamiTriggersForAnotherModifiedCreatureButNotAPlainOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.BigBear(g, me.ID, 1, false)
	effectstest.AuraCast(t, g, "One with the Kami", effectstest.OneWithTheKamiOracle, host)
	plain := effectstest.BigBear(g, me.ID, 5, false)
	effectstest.KillCreature(t, g, me.ID, plain)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountTokensControlled(g, me.ID, "Spirit"); n != 0 {
		t.Fatalf("a creature with no modification made %d Spirits", n)
	}
	modified := effectstest.BigBear(g, me.ID, 2, true) // power 2 base; counter is layer 7 so the power as it died is 3
	want := 0
	if c, ok := g.LookupCardForEffect(modified); ok {
		want = c.CurrentPower()
	}
	effectstest.KillCreature(t, g, me.ID, modified)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountTokensControlled(g, me.ID, "Spirit"); n != want || n == 0 {
		t.Errorf("Spirits = %d, want %d (its power as it died)", n, want)
	}
}

func TestOneWithTheKamiIgnoresAnOpponentsModifiedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.BigBear(g, me.ID, 1, false)
	effectstest.AuraCast(t, g, "One with the Kami", effectstest.OneWithTheKamiOracle, host)
	theirs := effectstest.BigBear(g, opp.ID, 4, true)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountTokensControlled(g, me.ID, "Spirit"); n != 0 {
		t.Errorf("Spirits = %d, want 0", n)
	}
}

func TestOneWithTheKamiMustEnchantACreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	theirs := effectstest.BigBear(g, g.Seats[1].ID, 2, false)
	hand := g.Seats[0].Hand
	id := uuid.New()
	hand.PushTop(game.Card{InstanceID: id, Name: "One with the Kami", OracleID: effectstest.OneWithTheKamiOracle, TypeLine: effectstest.AuraTypeLine,
		ManaCost: "{3}{G}", Owner: g.Seats[0].ID, Controller: g.Seats[0].ID})
	effectstest.ToMainForCost(t, g)
	effectstest.ApaMana(g.Seats[0], "G", "G", "G", "G")
	err := g.CastSpell(g.Seats[0].ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}})
	if err == nil {
		t.Error("cast on an opponent's creature was accepted")
	}
}
