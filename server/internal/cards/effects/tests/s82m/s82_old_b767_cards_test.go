package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b767_cards_test.go — S82 OLD batch 767 (#5628, Tempest: creatures
// with keywords). Each test plays the real catalog entry.

const (
	oldB767PincherBeetles   = "0fae1c73-fd34-433e-8989-fbd80e3b1c70"
	oldB767HeartwoodTree    = "27c21086-8d19-4379-af37-f251b2ec2502"
	oldB767Scragnoth        = "9e66f2d6-cc51-4747-914d-2c04e439da2b"
	oldB767WallOfDiffusion  = "ae42f803-f4ac-45f3-9ce7-227a59096f19"
	oldB767BenthicBehemoth  = "aea98950-536e-451d-bd25-ae6e3e459ffb"
	oldB767SkyshroudCondor  = "c7c4c67e-c4db-416b-b2e8-f0e59aa80bba"
	oldB767BayouDragonfly   = "eeb58cb4-b3da-400b-8bfb-ee985cd60f4b"
	oldB767CancelOracle     = "7d00fb28-ea6c-49a9-b4af-ffb38860a9a7"
	oldB767ProdigalSorcerer = "5e961d15-5972-4e4b-9385-1cd7cd7c6bbe"
)

var oldB767Cards = []struct {
	name, typeLine, oracle string
	power, toughness       int
	keywords               []string
}{
	{"Pincher Beetles", "Creature — Insect", oldB767PincherBeetles, 3, 1, []string{"shroud"}},
	{"Heartwood Treefolk", "Creature — Treefolk", oldB767HeartwoodTree, 3, 4, []string{"forestwalk"}},
	{"Scragnoth", "Creature — Beast", oldB767Scragnoth, 3, 4, []string{"protection from blue"}},
	{"Wall of Diffusion", "Creature — Wall", oldB767WallOfDiffusion, 0, 5, []string{"defender"}},
	{"Benthic Behemoth", "Creature — Serpent", oldB767BenthicBehemoth, 7, 6, []string{"islandwalk"}},
	{"Skyshroud Condor", "Creature — Bird", oldB767SkyshroudCondor, 2, 2, []string{"flying"}},
	{"Bayou Dragonfly", "Creature — Insect", oldB767BayouDragonfly, 1, 1, []string{"flying", "swampwalk"}},
}

func TestOldB767EveryCardIsRegisteredFullWithItsKeywords(t *testing.T) {
	for _, c := range oldB767Cards {
		spec, ok := effects.Lookup(c.oracle)
		if !ok {
			t.Errorf("%s is not registered", c.name)
			continue
		}
		if spec.Name != c.name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", c.name, spec.Name, spec.Completeness, spec.Caveats)
		}
		g := effectstest.NewCatalogGame(t)
		id := effectstest.PushCatalogPermanent(g, g.Seats[0].ID, c.name, c.typeLine, c.oracle, false)
		for _, kw := range c.keywords {
			if !effectstest.HasKeywordOnBattlefield(t, g, id, kw) {
				t.Errorf("%s does not have %q on the battlefield", c.name, kw)
			}
		}
	}
}

func TestOldB767PincherBeetlesCannotBeTargeted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	beetles := effectstest.PushCatalogPermanent(g, me.ID, "Pincher Beetles", "Creature — Insect", oldB767PincherBeetles, false)
	sorcerer := effectstest.PushCatalogPermanent(g, me.ID, "Prodigal Sorcerer", "Creature — Human Wizard", oldB767ProdigalSorcerer, false)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, sorcerer, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(beetles)}); err == nil {
		t.Error("an ability targeted a creature with shroud")
	}
}

func TestOldB767LandwalkMakesTheAttackerUnblockableOnlyAgainstTheLand(t *testing.T) {
	for _, tc := range []struct {
		card     int
		land     string
		unblocks bool
	}{
		{1, "Forest", true},
		{1, "Island", false},
		{4, "Island", true},
		{4, "Swamp", false},
		{6, "Swamp", true},
		{6, "Forest", false},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		c := oldB767Cards[tc.card]
		effectstest.AdvanceToMain(t, g)
		attacker := effectstest.PushCatalogPermanent(g, me.ID, c.name, c.typeLine, c.oracle, false)
		effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: tc.land, TypeLine: "Basic Land — " + tc.land})
		blocker := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4, Keywords: []string{"reach"}})
		effectstest.DeclareAttack(t, g, opp.ID, attacker)
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		err := g.DeclareBlocker(blocker, attacker)
		if (err != nil) != tc.unblocks {
			t.Errorf("%s against a %s: block error %v, want refused %v", c.name, tc.land, err, tc.unblocks)
		}
	}
}

func TestOldB767ScragnothCannotBeCounteredAndHasProtectionFromBlue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CCCCG")
	scragnoth := effectstest.CastFromHandForTest(t, g, me, "Scragnoth", "Creature — Beast", "{4}{G}", oldB767Scragnoth, game.CastSpellParams{Strict: true})
	effectstest.FloatForTest(g, opp, "CUU")
	effectstest.CastFromHandForTest(t, g, opp, "Cancel", "Instant", "{1}{U}{U}", oldB767CancelOracle, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: scragnoth}}})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(scragnoth) {
		t.Error("Cancel countered Scragnoth")
	}
}

func TestOldB767SkyshroudCondorNeedsAnotherSpellThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CCUUU")
	condor := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: condor, Name: "Skyshroud Condor", TypeLine: "Creature — Bird", ManaCost: "{1}{U}", OracleID: oldB767SkyshroudCondor, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, condor, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("the Condor was cast as the first spell of the turn")
	}
	effectstest.CastFromHandForTest(t, g, me, "Opt", "Instant", "{U}", "", game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, condor, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the Condor was refused after another spell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(condor) {
		t.Error("the Condor did not enter the battlefield")
	}
}

func TestOldB767WallOfDiffusionBlocksCreaturesWithShadowButNotFlyingShadow(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	shade := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Shade", TypeLine: "Creature — Shade", Power: 2, Toughness: 2, Keywords: []string{"shadow"}})
	flyer := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Flying Shade", TypeLine: "Creature — Shade", Power: 2, Toughness: 2, Keywords: []string{"shadow", "flying"}})
	bear := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	wall := effectstest.PushCatalogPermanent(g, opp.ID, "Wall of Diffusion", "Creature — Wall", oldB767WallOfDiffusion, false)
	other := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Other Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4})
	effectstest.DeclareAttack(t, g, opp.ID, shade, flyer, bear)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(other, shade); err == nil {
		t.Error("a creature without the waiver blocked a creature with shadow")
	}
	if err := g.DeclareBlocker(wall, flyer); err == nil {
		t.Error("the Wall blocked a flying creature with shadow")
	}
	if err := g.DeclareBlocker(wall, shade); err != nil {
		t.Errorf("the Wall could not block a creature with shadow: %v", err)
	}
	if effectstest.HasKeywordOnBattlefield(t, g, wall, "shadow") {
		t.Error("the Wall has shadow")
	}
}
