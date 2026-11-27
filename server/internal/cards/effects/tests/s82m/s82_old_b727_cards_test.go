package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b727_cards_test.go — Urza's Saga creatures with keywords, batch 727
// (#5624). Each test casts the real catalog entry through the engine and
// asserts what its text says.

const (
	oldB727SerraZealot    = "989a7353-8d3d-4ea2-ab5e-8535d95dddae"
	oldB727ViashinoRunner = "ac7e317b-387c-403f-bcf1-403f96302f21"
	oldB727SkitteringSkrg = "fae1d5d9-a50e-423e-8692-3e41f39df081"
	oldB727EndlessWurm    = "ce4d714c-c22b-47be-a372-a3139c4e5b1c"
	oldB727VoiceOfGrace   = "7338f964-b7d8-4491-b933-7d8b7d1b56ed"
	oldB727HawkeaterMoth  = "e3f4bc4c-e747-4bc1-bfba-4ecfa0a5d17f"
	oldB727Guma           = "a9cbf2ed-9896-41b5-b17f-e2fcb3cb82fa"
	oldB727VoiceOfLaw     = "535007c5-03a8-495f-a132-37960f179f96"
	oldB727Zephid         = "88aa710c-26ed-490d-9a4b-4a2b48df1733"
)

var oldB727Cards = []struct {
	oracle, name, typeLine string
	power, toughness       int
	keywords               []string
}{
	{oldB727SerraZealot, "Serra Zealot", "Creature — Human Soldier", 1, 1, []string{"first strike"}},
	{oldB727ViashinoRunner, "Viashino Runner", "Creature — Lizard", 3, 2, []string{"menace"}},
	{oldB727SkitteringSkrg, "Skittering Skirge", "Creature — Phyrexian Imp", 3, 2, []string{"flying"}},
	{oldB727EndlessWurm, "Endless Wurm", "Creature — Wurm", 9, 9, []string{"trample"}},
	{oldB727VoiceOfGrace, "Voice of Grace", "Creature — Angel", 2, 2, []string{"flying", "protection from black"}},
	{oldB727HawkeaterMoth, "Hawkeater Moth", "Creature — Insect", 1, 2, []string{"flying", "shroud"}},
	{oldB727Guma, "Guma", "Creature — Cat", 2, 2, []string{"protection from blue"}},
	{oldB727VoiceOfLaw, "Voice of Law", "Creature — Angel", 2, 2, []string{"flying", "protection from red"}},
	{oldB727Zephid, "Zephid", "Creature — Illusion", 3, 4, []string{"flying", "shroud"}},
}

func TestOldB727EveryCardIsRegisteredFullWithItsPrintedKeywords(t *testing.T) {
	for _, tc := range oldB727Cards {
		spec, ok := effects.Lookup(tc.oracle)
		if !ok {
			t.Errorf("%s is not registered", tc.name)
			continue
		}
		if spec.Name != tc.name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", tc.name, spec.Name, spec.Completeness, spec.Caveats)
		}
		if len(spec.PrintedKeywords) != len(tc.keywords) {
			t.Errorf("%s: printed keywords %v, want %v", tc.name, spec.PrintedKeywords, tc.keywords)
			continue
		}
		for i, kw := range tc.keywords {
			if spec.PrintedKeywords[i] != kw {
				t.Errorf("%s: printed keyword %d is %q, want %q", tc.name, i, spec.PrintedKeywords[i], kw)
			}
		}
	}
}

func TestOldB727CastCreatureHasItsKeywords(t *testing.T) {
	for _, tc := range oldB727Cards {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.CastCreatureCard(t, g, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			if !g.Battlefield.Contains(id) {
				t.Fatal("the creature did not resolve onto the battlefield")
			}
			if tc.name == "Skittering Skirge" {
				return
			}
			for _, kw := range tc.keywords {
				if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
					t.Errorf("the creature lacks %q", kw)
				}
			}
		})
	}
}

func TestOldB727SkitteringSkirgeIsSacrificedWhenYouCastACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	skirge := effectstest.CastCreatureCard(t, g, "Skittering Skirge", "Creature — Phyrexian Imp", oldB727SkitteringSkrg, 3, 2)
	if !g.Battlefield.Contains(skirge) {
		t.Fatal("casting the Skirge itself sacrificed it")
	}
	if !effectstest.HasEffectiveKeyword(t, g, skirge, "flying") {
		t.Error("the Skirge lacks flying")
	}
	// A noncreature spell does nothing.
	effectstest.CastCatalogSpell(t, g, "Shock", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(skirge) {
		t.Fatal("a noncreature spell sacrificed the Skirge")
	}
	// An opponent's creature spell does nothing: "When you cast".
	g.Turn.ActiveSeat = 1
	effectstest.CastCreatureCard(t, g, "Their Bear", "Creature — Bear", "", 2, 2)
	g.Turn.ActiveSeat = 0
	if !g.Battlefield.Contains(skirge) {
		t.Fatal("an opponent's creature spell sacrificed the Skirge")
	}
	effectstest.CastCreatureCard(t, g, "My Bear", "Creature — Bear", "", 2, 2)
	if g.Battlefield.Contains(skirge) {
		t.Error("the Skirge was not sacrificed when its controller cast a creature spell")
	}
	if !me.Graveyard.Contains(skirge) {
		t.Error("the Skirge is not in its owner's graveyard")
	}
	_ = opp
}

func oldB727WurmAtUpkeep(t *testing.T, enchantments int) (*game.Game, uuid.UUID, []uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	wurm := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Endless Wurm", TypeLine: "Creature — Wurm", OracleID: oldB727EndlessWurm,
		Power: 9, Toughness: 9, Keywords: []string{"trample"}, Owner: me.ID, Controller: me.ID,
	})
	var ench []uuid.UUID
	for i := 0; i < enchantments; i++ {
		ench = append(ench, effectstest.TmpB1Push(g, me.ID, "Enchantment", "Enchantment", "", 0, 0))
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(wurm) {
		t.Fatal("the Wurm triggered at another player's upkeep")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if enchantments > 0 {
		for i := 0; i < 8 && effectstest.LatestOptionPickFor(g, me.ID) == nil; i++ {
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority: %v", err)
			}
		}
	}
	return g, wurm, ench
}

func TestOldB727EndlessWurmIsSacrificedWithoutAnEnchantment(t *testing.T) {
	g, wurm, _ := oldB727WurmAtUpkeep(t, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wurm) {
		t.Error("the Wurm survived its upkeep with no enchantment to sacrifice")
	}
}

func TestOldB727EndlessWurmSacrificesAnEnchantmentToStay(t *testing.T) {
	g, wurm, ench := oldB727WurmAtUpkeep(t, 2)
	me := g.Seats[0]
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.AnswerChooseCards(t, g, me.ID, ench[0])
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(wurm) {
		t.Error("the Wurm was sacrificed although an enchantment was")
	}
	if g.Battlefield.Contains(ench[0]) {
		t.Error("the chosen enchantment was not sacrificed")
	}
	if !g.Battlefield.Contains(ench[1]) {
		t.Error("the other enchantment was sacrificed")
	}
}

func TestOldB727EndlessWurmMayBeSacrificedInsteadOfAnEnchantment(t *testing.T) {
	g, wurm, ench := oldB727WurmAtUpkeep(t, 1)
	effectstest.AnswerOptionPick(t, g, g.Seats[0].ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wurm) {
		t.Error("the Wurm stayed after the controller chose to sacrifice it")
	}
	if !g.Battlefield.Contains(ench[0]) {
		t.Error("the enchantment was sacrificed anyway")
	}
}

func TestOldB727ShroudCreaturesCannotBeTargeted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	moth := effectstest.CastCreatureCard(t, g, "Hawkeater Moth", "Creature — Insect", oldB727HawkeaterMoth, 1, 2)
	err := effectstest.CastCatalogSpellErr(t, g, "Sluggishness", effectstest.AuraTypeLine, "39eb4062-68a0-4f64-a9cc-79be87fd8173",
		[]game.TargetRef{{Kind: game.TargetCard, ID: moth}})
	if err == nil {
		t.Error("an Aura targeted a creature with shroud")
	}
}
