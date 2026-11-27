package testss82m

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	o595Recluse  = "17ff0155-0006-464a-a65f-b8e1b9aeba0b"
	o595Mass     = "18202185-f9d4-446a-b34e-fae9800240f9"
	o595Kannushi = "39a32629-8b05-4539-9ce7-1a6e467d0064"
	o595Yomiji   = "438be2fa-5f8d-42ff-894a-671ba4221196"
	o595Cohort   = "88f9a7e3-0b35-43e4-b16a-19c54a1e4e2c"
	o595Shoji    = "a7cb1da6-e56a-42b6-8087-cf822783373f"
	o595FrostOgr = "b039997c-2115-4f6d-9f7b-0ad35a4796a6"
	o595Shizuko  = "c4984e18-9caf-4dd0-9c24-e60ccb27e2a0"
	o595Marauder = "d9dafa31-41ea-4f1c-bc63-ec84b01d57fa"
)

func TestO595CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		o595Recluse: "Ogre Recluse", o595Mass: "Gnarled Mass", o595Kannushi: "Empty-Shrine Kannushi", o595Yomiji: "Yomiji, Who Bars the Way",
		o595Cohort: "Goblin Cohort", o595Shoji: "Kami of Tattered Shoji", o595FrostOgr: "Frost Ogre", o595Shizuko: "Shizuko, Caller of Autumn",
		o595Marauder: "Ogre Marauder",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestO595OgreRecluseTapsWhenAnyPlayerCastsASpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ogre := effectstest.B12Push(g, me.ID, "Ogre Recluse", "Creature — Ogre Warrior", o595Recluse, 5, 4)
	effectstest.AdvanceToMain(t, g)
	if effectstest.B16Tapped(t, g, ogre) {
		t.Fatal("the Recluse started tapped")
	}
	effectstest.P1g2Cast(t, g, "Bear", "Creature — Bear", "", "{1}{G}", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, ogre) {
		t.Error("the Recluse did not tap when a spell was cast")
	}
}

func TestO595KannushiHasProtectionFromTheColorsOfYourPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	k := effectstest.B12Push(g, me.ID, "Empty-Shrine Kannushi", "Creature — Human Cleric", o595Kannushi, 1, 1)
	has := func(color string) bool {
		c := effectstest.GaLayered(t, g, k)
		return game.HasKeyword(&c, game.ProtectionFromColor(color))
	}
	if has("R") || has("G") {
		t.Fatal("the Kannushi has protection before any coloured permanent but itself is in play")
	}
	effectstest.Pr7bCreature(g, me.ID, "Red Thing", 1, "{R}", []string{"R"})
	g.RunStateChecksForTest()
	if !has("R") {
		t.Error("no protection from red with a red permanent in play")
	}
	if has("G") {
		t.Error("protection from green with no green permanent")
	}
	effectstest.Pr7bCreature(g, g.Seats[1].ID, "Their Green Thing", 1, "{G}", []string{"G"})
	g.RunStateChecksForTest()
	if has("G") {
		t.Error("an opponent's green permanent gave protection from green")
	}
}

func TestO595YomijiReturnsLegendaryPermanentsThatDie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	yomiji := effectstest.B12Push(g, me.ID, "Yomiji, Who Bars the Way", "Legendary Creature — Spirit", o595Yomiji, 4, 4)
	mine := effectstest.B12Creature(g, me.ID, "My Legend", "Legendary Creature — Human", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Legend", "Legendary Creature — Human", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	effectstest.Destroy(t, g, mine)
	effectstest.Destroy(t, g, theirs)
	effectstest.Destroy(t, g, plain)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HandHas(me, mine) {
		t.Error("my legendary creature was not returned to my hand")
	}
	if !effectstest.HandHas(opp, theirs) {
		t.Error("the opponent's legendary creature was not returned to its owner's hand")
	}
	if effectstest.HandHas(me, plain) || !effectstest.GraveyardHas(me, plain) {
		t.Error("a nonlegendary creature was returned")
	}
	effectstest.Destroy(t, g, yomiji)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HandHas(me, yomiji) {
		t.Error("Yomiji returned itself")
	}
}

func TestO595GoblinCohortAttacksOnlyAfterACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gob := effectstest.B12Push(g, me.ID, "Goblin Cohort", "Creature — Goblin Warrior", o595Cohort, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.P1g2Cast(t, g, "Zap", "Instant", "", "{R}", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(gob, opp.ID); err == nil {
		t.Fatal("the Cohort attacked after only a noncreature spell")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	gob2 := effectstest.B12Push(g2, me2.ID, "Goblin Cohort", "Creature — Goblin Warrior", o595Cohort, 2, 2)
	effectstest.AdvanceToMain(t, g2)
	effectstest.P1g2Cast(t, g2, "Bear", "Creature — Bear", "", "{1}{G}", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g2)
	effectstest.AdvanceTo(t, g2, game.StepDeclareAttackers)
	if err := g2.DeclareAttacker(gob2, opp2.ID); err != nil {
		t.Fatalf("the Cohort could not attack after a creature spell: %v", err)
	}
}

func TestO595KamiOfTatteredShojiGainsFlyingOnASpiritSpell(t *testing.T) {
	for _, spirit := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		kami := effectstest.B12Push(g, me.ID, "Kami of Tattered Shoji", "Creature — Spirit", o595Shoji, 2, 5)
		effectstest.AdvanceToMain(t, g)
		line := "Creature — Bear"
		if spirit {
			line = "Creature — Spirit"
		}
		effectstest.P1g2Cast(t, g, "Spell", line, "", "{1}{W}", game.CastSpellParams{})
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.GaLayered(t, g, kami)
		if got := game.HasKeyword(&c, "flying"); got != spirit {
			t.Errorf("spirit=%v: flying = %v", spirit, got)
		}
	}
}

func TestO595ShizukoGivesEachPlayerKeptGreenMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Shizuko, Caller of Autumn", "Legendary Creature — Snake Shaman", o595Shizuko, 2, 3)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.SortedPool(me); len(got) != 3 || got[0] != "G" {
		t.Fatalf("my pool in my upkeep: %v, want GGG", got)
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.SortedPool(opp); len(got) != 3 || got[0] != "G" {
		t.Errorf("the opponent's pool in their upkeep: %v, want GGG", got)
	}
}

func TestO595OgreMarauderLetsTheDefenderSacrificeOrBeUnblockable(t *testing.T) {
	for _, sacrifice := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		ogre := effectstest.B12Push(g, me.ID, "Ogre Marauder", "Creature — Ogre Warrior", o595Marauder, 3, 1)
		blk := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
		effectstest.DeclareAttack(t, g, opp.ID, ogre)
		effectstest.PassPriorityAroundTable(t, g)
		idx := 0
		if sacrifice {
			idx = 1
		}
		effectstest.AnswerOptionPick(t, g, opp.ID, idx)
		if sacrifice {
			effectstest.AnswerChooseCards(t, g, opp.ID, blk)
			effectstest.PassPriorityAroundTable(t, g)
			if g.Battlefield.Contains(blk) {
				t.Error("the defending player did not sacrifice the creature")
			}
			continue
		}
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		if err := g.DeclareBlocker(blk, ogre); err == nil {
			t.Error("the Marauder was blocked though the defender let it through")
		}
	}
}
