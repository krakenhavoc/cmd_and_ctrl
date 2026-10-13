package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #2940: the per-turn entry tally reads a permanent AS IT ENTERS with
// the layer pass applied, so a creature that has a type only through
// another effect counts. Every test here puts a Maskwood Nexus on the
// battlefield FIRST and then brings in a plain creature, which has none
// of the types printed.

func TestEntryTallyCountsATypeGrantedByMaskwoodNexus(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushNexusFor(g, me.ID)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Plain Soldier", "Creature — Human Soldier", "", nil)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		for _, st := range []string{"Elf", "Goblin", "Frog", "Rat"} {
			if got := g.EnteredWithSubtypeThisTurn(me.ID, st); got != 1 {
				t.Errorf("%s entered under Maskwood Nexus: tally %d, want 1", st, got)
			}
		}
	})
}

func TestHobgoblinBanditLordCountsACreatureThatEntersUnderMaskwoodNexus(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lord := b12Push(g, me.ID, "Hobgoblin Bandit Lord", "Creature — Goblin Rogue", b34HobgoblinBanditLordOracle, 2, 3)
	pushNexusFor(g, me.ID)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Plain Soldier", "Creature — Human Soldier", "", nil)
	passPriorityAroundTable(t, g)
	if got := hobgoblinLordDamage(t, g, me, opp, lord); got != 1 {
		t.Errorf("a Soldier that entered under Maskwood Nexus: %d damage, want 1", got)
	}
}

func TestLilypadVillageCountsACreatureThatEntersUnderMaskwoodNexus(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	village := pushCatalogPermanent(g, me.ID, "Lilypad Village", "Land", acLilypadVillageOracle, false)
	pushNexusFor(g, me.ID)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Plain Soldier", "Creature — Human Soldier", "", nil)
	passPriorityAroundTable(t, g)
	b06AddMana(me, "U")
	b16Activate(t, g, me.ID, village, 0, game.ActivateAbilityParams{})
	if surveilChoiceFor(g, me.ID) == nil {
		t.Error("a Soldier entering under Maskwood Nexus is a Frog: want a surveil 2 prompt")
	}
}

func TestEowynCountsAHumanThatEntersOnlyThroughMaskwoodNexus(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b43Catalog(g, me.ID, "Éowyn, Shieldmaiden", "Legendary Creature — Human Knight", b43EowynOracle, 5, 4)
	pushNexusFor(g, me.ID)
	for range 4 {
		advanceToNextSeatsTurn(t, g)
	}
	castCatalogSpell(t, g, "Plain Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if n := b43TokensNamed(g, me.ID, "Human Knight"); n != 2 {
		t.Errorf("a Bear entered as a Human under Maskwood Nexus: %d Knights, want 2", n)
	}
}

func TestGaladrielElvenQueenCountsAnElfThatEntersOnlyThroughMaskwoodNexus(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me := order[0]
	b43Catalog(g, me.ID, "Galadriel, Elven-Queen", "Legendary Creature — Elf Noble", galadrielElvenOracle, 4, 5)
	pushNexusFor(g, me.ID)
	for range order {
		advanceToNextSeatsTurn(t, g)
	}
	castCatalogSpell(t, g, "Plain Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if ballotFor(g, me.ID) == nil {
		t.Error("a Bear entered as an Elf under Maskwood Nexus: want a vote")
	}
}
