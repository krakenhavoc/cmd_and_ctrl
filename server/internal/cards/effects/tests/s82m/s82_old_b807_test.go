package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b807_test.go — Mirage creatures with keywords, S82 OLD batch 807,
// played through the real catalog entries.

const (
	oldB807Noble    = "99dd6f15-855f-431f-a3c8-31f1421af88f"
	oldB807Teremko  = "a00cb1fa-a196-4403-9008-532dbd15769e"
	oldB807IronTusk = "a299470a-72ea-4848-833b-dc894db7be15"
	oldB807Dragon   = "a5e9f72b-7628-4265-8b43-592a0bf678b3"
	oldB807Falcon   = "aa17782b-d8e7-423b-9f03-01ccb600f69a"
	oldB807Rhinos   = "b2d7251e-2848-4b85-8d99-7c43be163dfb"
	oldB807Drake    = "b9be76a3-e14f-48dd-8824-17d5708ac7e0"
	oldB807Minotaur = "c5661500-c48f-4f6d-bbe7-c8bd7118d862"
)

func TestOldB807CardsAreFullWithNoCaveats(t *testing.T) {
	for id, name := range map[string]string{
		oldB807Noble: "Noble Elephant", oldB807Teremko: "Teremko Griffin",
		oldB807IronTusk: "Iron Tusk Elephant", oldB807Dragon: "Catacomb Dragon",
		oldB807Falcon: "Windreaper Falcon", oldB807Rhinos: "Crash of Rhinos",
		oldB807Drake: "Hazerider Drake", oldB807Minotaur: "Talruum Minotaur",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB807PrintedKeywordsAreOnTheBattlefieldCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, tc := range []struct {
		name, typeLine, oracle string
		p, tough               int
		have, lack             []string
	}{
		{"Noble Elephant", "Creature — Elephant", oldB807Noble, 2, 2, []string{"trample", game.KeywordBanding}, []string{"flying"}},
		{"Teremko Griffin", "Creature — Griffin", oldB807Teremko, 2, 2, []string{"flying", game.KeywordBanding}, []string{"trample"}},
		{"Iron Tusk Elephant", "Creature — Elephant", oldB807IronTusk, 3, 3, []string{"trample"}, []string{game.KeywordBanding}},
		{"Catacomb Dragon", "Creature — Dragon", oldB807Dragon, 4, 4, []string{"flying"}, []string{"trample"}},
		{"Windreaper Falcon", "Creature — Bird", oldB807Falcon, 1, 1, []string{"flying", game.ProtectionFromColor("U")}, []string{game.ProtectionFromColor("R")}},
		{"Crash of Rhinos", "Creature — Rhino", oldB807Rhinos, 8, 4, []string{"trample"}, []string{"haste"}},
		{"Hazerider Drake", "Creature — Drake", oldB807Drake, 2, 3, []string{"flying", game.ProtectionFromColor("R")}, []string{game.ProtectionFromColor("U")}},
		{"Talruum Minotaur", "Creature — Minotaur Berserker", oldB807Minotaur, 3, 3, []string{"haste"}, []string{"trample"}},
	} {
		id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.p, tc.tough)
		for _, kw := range tc.have {
			if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
				t.Errorf("%s lacks %q", tc.name, kw)
			}
		}
		for _, kw := range tc.lack {
			if effectstest.HasEffectiveKeyword(t, g, id, kw) {
				t.Errorf("%s has %q, which it does not print", tc.name, kw)
			}
		}
	}
}

func TestOldB807CatacombDragonShrinksOnlyNonartifactNonDragonBlockersByHalfTheirPower(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		power, want    int
	}{
		{"Ogre", "Creature — Ogre", 5, 3},
		{"Bear", "Creature — Bear", 2, 1},
		{"Squire", "Creature — Human", 1, 1},
		{"Golem", "Artifact Creature — Golem", 4, 4},
		{"Wyrm", "Creature — Dragon", 4, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			dragon := effectstest.B12Push(g, me.ID, "Catacomb Dragon", "Creature — Dragon", oldB807Dragon, 4, 4)
			blocker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: tc.name, TypeLine: tc.typeLine, Keywords: []string{"reach"},
				Power: tc.power, Toughness: 9, Owner: opp.ID, Controller: opp.ID,
			})
			effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
			if err := g.DeclareAttacker(dragon, opp.ID); err != nil {
				t.Fatal(err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
			if err := g.DeclareBlocker(blocker, dragon); err != nil {
				t.Fatal(err)
			}
			effectstest.FinishBlockDeclarations(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.EffectivePower(t, g, blocker); got != tc.want {
				t.Errorf("%s blocker power = %d, want %d", tc.typeLine, got, tc.want)
			}
		})
	}
}
