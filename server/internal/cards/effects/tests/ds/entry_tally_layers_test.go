package testsds

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #2940: the per-turn entry tally reads a permanent AS IT ENTERS with
// the layer pass applied, so a creature that has a type only through
// another effect counts. Every test here puts a Maskwood Nexus on the
// battlefield FIRST and then brings in a plain creature, which has none
// of the types printed.

func TestEntryTallyCountsATypeGrantedByMaskwoodNexus(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushNexusFor(g, me.ID)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Plain Soldier", "Creature — Human Soldier", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		for _, st := range []string{"Elf", "Goblin", "Frog", "Rat"} {
			if got := g.EnteredWithSubtypeThisTurn(me.ID, st); got != 1 {
				t.Errorf("%s entered under Maskwood Nexus: tally %d, want 1", st, got)
			}
		}
	})
}

func TestHobgoblinBanditLordCountsACreatureThatEntersUnderMaskwoodNexus(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lord := effectstest.B12Push(g, me.ID, "Hobgoblin Bandit Lord", "Creature — Goblin Rogue", effectstest.B34HobgoblinBanditLordOracle, 2, 3)
	effectstest.PushNexusFor(g, me.ID)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Plain Soldier", "Creature — Human Soldier", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HobgoblinLordDamage(t, g, me, opp, lord); got != 1 {
		t.Errorf("a Soldier that entered under Maskwood Nexus: %d damage, want 1", got)
	}
}

func TestLilypadVillageCountsACreatureThatEntersUnderMaskwoodNexus(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	village := effectstest.PushCatalogPermanent(g, me.ID, "Lilypad Village", "Land", effectstest.AcLilypadVillageOracle, false)
	effectstest.PushNexusFor(g, me.ID)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Plain Soldier", "Creature — Human Soldier", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B06AddMana(me, "U")
	effectstest.B16Activate(t, g, me.ID, village, 0, game.ActivateAbilityParams{})
	if effectstest.SurveilChoiceFor(g, me.ID) == nil {
		t.Error("a Soldier entering under Maskwood Nexus is a Frog: want a surveil 2 prompt")
	}
}

func TestEowynCountsAHumanThatEntersOnlyThroughMaskwoodNexus(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B43Catalog(g, me.ID, "Éowyn, Shieldmaiden", "Legendary Creature — Human Knight", effectstest.B43EowynOracle, 5, 4)
	effectstest.PushNexusFor(g, me.ID)
	for range 4 {
		effectstest.AdvanceToNextSeatsTurn(t, g)
	}
	effectstest.CastCatalogSpell(t, g, "Plain Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B43TokensNamed(g, me.ID, "Human Knight"); n != 2 {
		t.Errorf("a Bear entered as a Human under Maskwood Nexus: %d Knights, want 2", n)
	}
}

func TestGaladrielElvenQueenCountsAnElfThatEntersOnlyThroughMaskwoodNexus(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me := order[0]
	effectstest.B43Catalog(g, me.ID, "Galadriel, Elven-Queen", "Legendary Creature — Elf Noble", effectstest.GaladrielElvenOracle, 4, 5)
	effectstest.PushNexusFor(g, me.ID)
	for range order {
		effectstest.AdvanceToNextSeatsTurn(t, g)
	}
	effectstest.CastCatalogSpell(t, g, "Plain Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BallotFor(g, me.ID) == nil {
		t.Error("a Bear entered as an Elf under Maskwood Nexus: want a vote")
	}
}
