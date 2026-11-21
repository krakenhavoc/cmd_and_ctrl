package testsms

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// nem_b1_cards_test.go — Nemesis, batch 1 (#5618). Each test plays the
// real catalog entry and asserts what its text says.

const (
	nemB1OraxidOracle        = "7d086ae5-738c-4685-80ae-66bd9379d8f7"
	nemB1VoiceOracle         = "33435856-ea24-4d10-84d5-6872b35b63a3"
	nemB1CommandoOracle      = "d62e9510-dc41-493f-8289-6e11303884fc"
	nemB1CrusherOracle       = "cc18fcfc-bfe8-41d8-9037-5ff05b2a9c44"
	nemB1WallOracle          = "5fa86b45-1d5b-42df-8811-6a10903f3039"
	nemB1MoggSalvageOracle   = "0923ad85-2ea4-4463-8f71-b7d0a162b486"
	nemB1RainOracle          = "c6f69168-71c2-41a2-ab95-cfd7baf0ba68"
	nemB1MassacreOracle      = "ce1866e9-28de-4bcc-9999-11958e6c13a6"
	nemB1MoggAlarmOracle     = "1c2117d5-4600-489c-be5b-dc0ed69b30ca"
	nemB1LawbringerOracle    = "15cc92c0-76ce-4b97-b1a2-a7d58bbfe857"
	nemB1LightbringerOracle  = "8656d9c1-d890-4473-9da8-7ec339d7d546"
	nemB1SerpentOracle       = "25ae99c5-83b7-428d-b049-8ad40e250696"
	nemB1SealOfFireOracle    = "348a345e-4639-41ca-b015-a5d43459eb64"
	nemB1SealOfRemovalOracle = "f0801029-bcf7-4bdb-84bf-e88dcaa9dc03"
)

func TestNemB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		nemB1OraxidOracle:        "Oraxid",
		nemB1VoiceOracle:         "Voice of Truth",
		nemB1CommandoOracle:      "Rootwater Commando",
		nemB1CrusherOracle:       "Flowstone Crusher",
		nemB1WallOracle:          "Flowstone Wall",
		nemB1MoggSalvageOracle:   "Mogg Salvage",
		nemB1RainOracle:          "Refreshing Rain",
		nemB1MassacreOracle:      "Massacre",
		nemB1MoggAlarmOracle:     "Mogg Alarm",
		nemB1LawbringerOracle:    "Lawbringer",
		nemB1LightbringerOracle:  "Lightbringer",
		nemB1SerpentOracle:       "Sliptide Serpent",
		nemB1SealOfFireOracle:    "Seal of Fire",
		nemB1SealOfRemovalOracle: "Seal of Removal",
	} {
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

// nemB1Push seeds a permanent on the battlefield, catalog row or not,
// untapped and not summoning sick.
func nemB1Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int, colors ...string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Colors: colors, Owner: owner, Controller: owner,
	})
}

// nemB1Flyer seeds a vanilla creature with flying.
func nemB1Flyer(g *game.Game, owner uuid.UUID, name string, colors ...string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Keywords: []string{"flying"}, Colors: colors, Owner: owner, Controller: owner,
	})
}

// --- keywords -----------------------------------------------------------------------

func TestNemB1OraxidCantBeBlockedByRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	oraxid := nemB1Push(g, me.ID, "Oraxid", "Creature — Crab Beast", nemB1OraxidOracle, 2, 3, "U")
	if !effectstest.HasEffectiveKeyword(t, g, oraxid, "protection from red") {
		t.Fatal("Oraxid does not have protection from red")
	}
	red := nemB1Push(g, them.ID, "Red blocker", "Creature — Goblin", "", 2, 2, "R")
	green := nemB1Push(g, them.ID, "Green blocker", "Creature — Bear", "", 2, 2, "G")
	effectstest.AttackIntoBlocks(t, g, them.ID, oraxid)
	if effectstest.BlockRefusal(t, g, red, oraxid) == "" {
		t.Error("a red creature was allowed to block Oraxid")
	}
	if got := effectstest.BlockRefusal(t, g, green, oraxid); got != "" {
		t.Errorf("a green creature was refused: %q", got)
	}
}

func TestNemB1VoiceOfTruthIsAFlyerWithProtectionFromWhite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	voice := nemB1Push(g, me.ID, "Voice of Truth", "Creature — Angel", nemB1VoiceOracle, 2, 2, "W")
	for _, kw := range []string{"flying", "protection from white"} {
		if !effectstest.HasEffectiveKeyword(t, g, voice, kw) {
			t.Fatalf("Voice of Truth does not have %s", kw)
		}
	}
	whiteFlyer := nemB1Flyer(g, them.ID, "White flyer", "W")
	blueFlyer := nemB1Flyer(g, them.ID, "Blue flyer", "U")
	ground := nemB1Push(g, them.ID, "Ground blocker", "Creature — Bear", "", 2, 2, "G")
	effectstest.AttackIntoBlocks(t, g, them.ID, voice)
	if effectstest.BlockRefusal(t, g, whiteFlyer, voice) == "" {
		t.Error("a white creature with flying was allowed to block")
	}
	if effectstest.BlockRefusal(t, g, ground, voice) == "" {
		t.Error("a creature without flying or reach was allowed to block")
	}
	if got := effectstest.BlockRefusal(t, g, blueFlyer, voice); got != "" {
		t.Errorf("a blue creature with flying was refused: %q", got)
	}
}

func TestNemB1RootwaterCommandoHasIslandwalk(t *testing.T) {
	for _, tc := range []struct {
		name   string
		island bool
	}{{"defender controls an Island", true}, {"defender controls no Island", false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, them := effectstest.SeatsForCombat(g)
			commando := nemB1Push(g, me.ID, "Rootwater Commando", "Creature — Merfolk", nemB1CommandoOracle, 2, 2, "U")
			bear := nemB1Push(g, them.ID, "Grizzly Bears", "Creature — Bear", "", 2, 2, "G")
			effectstest.SeedLand(g, me.ID, "My Island", "Basic Land — Island", "")
			if tc.island {
				effectstest.SeedLand(g, them.ID, "Island", "Basic Land — Island", "")
			}
			effectstest.AttackIntoBlocks(t, g, them.ID, commando)
			got := effectstest.BlockRefusal(t, g, bear, commando)
			if tc.island && got != game.BlockReasonLandwalk {
				t.Errorf("block against an Island: reason %q, want landwalk", got)
			}
			if !tc.island && got != "" {
				t.Errorf("only the attacker's own Island, yet the block was refused: %q", got)
			}
		})
	}
}

// --- Flowstone pumps -----------------------------------------------------------------------

func TestNemB1FlowstonePumpsGivePlusOneMinusOne(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine string
		power, toughness       int
	}{
		{"Flowstone Crusher", nemB1CrusherOracle, "Creature — Beast", 4, 4},
		{"Flowstone Wall", nemB1WallOracle, "Creature — Wall", 0, 6},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			id := nemB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness, "R")
			for i := 1; i <= 2; i++ {
				effectstest.DcMana(t, g, me.ID, "{R}")
				if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
					t.Fatalf("activation %d: %v", i, err)
				}
				effectstest.PassPriorityAroundTable(t, g)
				if p, tough := effectstest.PtOf(t, g, id); p != tc.power+i || tough != tc.toughness-i {
					t.Fatalf("%s is %d/%d after %d activations, want %d/%d", tc.name, p, tough, i, tc.power+i, tc.toughness-i)
				}
			}
			if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
				t.Error("the pump was activated with an empty mana pool")
			}
		})
	}
}

func TestNemB1FlowstoneWallCantAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	wall := nemB1Push(g, me.ID, "Flowstone Wall", "Creature — Wall", nemB1WallOracle, 0, 6, "R")
	if !effectstest.HasEffectiveKeyword(t, g, wall, "defender") {
		t.Fatal("Flowstone Wall does not have defender")
	}
	effectstest.AdvanceToStepInTurn(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(wall, them.ID); err == nil {
		t.Error("Flowstone Wall was declared as an attacker")
	}
}

// --- the Nemesis free spells -----------------------------------------------------------

// nemB1FreeCase is one of the cycle: the opponent's land, yours, and the
// card's cast.
type nemB1FreeCase struct {
	name, typeLine, oracle string
	theirs, yours          string
}

var nemB1FreeCases = []nemB1FreeCase{
	{"Mogg Salvage", "Instant", nemB1MoggSalvageOracle, "Island", "Mountain"},
	{"Refreshing Rain", "Instant", nemB1RainOracle, "Swamp", "Forest"},
	{"Massacre", "Sorcery", nemB1MassacreOracle, "Plains", "Swamp"},
}

// nemB1FreeTargets seeds what each spell needs to point at and returns
// its targets.
func nemB1FreeTargets(g *game.Game, tc nemB1FreeCase) []game.TargetRef {
	me, them := effectstest.SeatsForCombat(g)
	switch tc.name {
	case "Mogg Salvage":
		return effectstest.CardTarget(nemB1Push(g, them.ID, "Mind Stone", "Artifact", "", 0, 0))
	case "Refreshing Rain":
		return []game.TargetRef{effectstest.PlayerTarget(me.ID)}
	}
	return nil
}

func TestNemB1FreeSpellsNeedTheOpponentsLandAndYours(t *testing.T) {
	for _, tc := range nemB1FreeCases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, them := effectstest.SeatsForCombat(g)
			id := effectstest.HandCardAtMain(t, g, tc.name, tc.typeLine, tc.oracle)
			targets := nemB1FreeTargets(g, tc)
			// Your own copy of their land type doesn't count.
			effectstest.SeedLand(g, me.ID, "My "+tc.theirs, "Basic Land — "+tc.theirs, "")
			effectstest.SeedLand(g, me.ID, tc.yours, "Basic Land — "+tc.yours, "")
			if err := effectstest.CastForAltCost(g, id, "", "free", nil, targets); err == nil {
				t.Fatalf("%s was cast for free with no opponent's %s", tc.name, tc.theirs)
			}
			if !me.Hand.Contains(id) {
				t.Fatal("the refused cast left the hand")
			}
			effectstest.SeedLand(g, them.ID, tc.theirs, "Basic Land — "+tc.theirs, "")
			if err := effectstest.CastForAltCost(g, id, "", "free", nil, targets); err != nil {
				t.Fatalf("free cast with both lands: %v", err)
			}
		})
	}
}

func TestNemB1FreeSpellsRefuseWithoutYourLand(t *testing.T) {
	for _, tc := range nemB1FreeCases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			_, them := effectstest.SeatsForCombat(g)
			id := effectstest.HandCardAtMain(t, g, tc.name, tc.typeLine, tc.oracle)
			targets := nemB1FreeTargets(g, tc)
			effectstest.SeedLand(g, them.ID, tc.theirs, "Basic Land — "+tc.theirs, "")
			effectstest.SeedLand(g, them.ID, "Their "+tc.yours, "Basic Land — "+tc.yours, "")
			if err := effectstest.CastForAltCost(g, id, "", "free", nil, targets); err == nil {
				t.Fatalf("%s was cast for free with no %s of your own", tc.name, tc.yours)
			}
		})
	}
}

func TestNemB1MoggSalvageDestroysTheArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, them := effectstest.SeatsForCombat(g)
	stone := nemB1Push(g, them.ID, "Mind Stone", "Artifact", "", 0, 0)
	bear := nemB1Push(g, them.ID, "Grizzly Bears", "Creature — Bear", "", 2, 2, "G")
	if err := effectstest.CastCatalogSpellErr(t, g, "Mogg Salvage", "Instant", nemB1MoggSalvageOracle, effectstest.CardTarget(bear)); err == nil {
		t.Error("a creature was accepted as the target artifact")
	}
	effectstest.CastCatalogSpell(t, g, "Mogg Salvage", "Instant", nemB1MoggSalvageOracle, effectstest.CardTarget(stone))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(stone) {
		t.Error("the artifact survived")
	}
}

func TestNemB1RefreshingRainGainsSixLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, them := effectstest.SeatsForCombat(g)
	life := them.Life
	effectstest.CastCatalogSpell(t, g, "Refreshing Rain", "Instant", nemB1RainOracle, []game.TargetRef{effectstest.PlayerTarget(them.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := them.Life - life; got != 6 {
		t.Errorf("the target gained %d life, want 6", got)
	}
}

func TestNemB1MassacreGivesEveryCreatureMinusTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	mine := nemB1Push(g, me.ID, "My Giant", "Creature — Giant", "", 3, 3, "B")
	bear := nemB1Push(g, them.ID, "Grizzly Bears", "Creature — Bear", "", 2, 2, "G")
	effectstest.CastCatalogSpell(t, g, "Massacre", "Sorcery", nemB1MassacreOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("a 2/2 survived -2/-2")
	}
	if p, tough := effectstest.PtOf(t, g, mine); p != 1 || tough != 1 {
		t.Errorf("the caster's 3/3 is %d/%d, want 1/1: Massacre hits every creature", p, tough)
	}
}

// --- Mogg Alarm ---------------------------------------------------------------------------

func TestNemB1MoggAlarmSacrificesTwoMountainsForTwoGoblins(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	alarm := effectstest.HandCardAtMain(t, g, "Mogg Alarm", "Sorcery", nemB1MoggAlarmOracle)
	m1 := effectstest.SeedLand(g, me.ID, "Mountain A", "Basic Land — Mountain", "")
	island := effectstest.SeedLand(g, me.ID, "Island", "Basic Land — Island", "")
	if err := effectstest.CastForAltCost(g, alarm, "", "sacrifice", []uuid.UUID{m1, island}, nil); err == nil {
		t.Fatal("a Mountain and an Island paid \"sacrifice two Mountains\"")
	}
	if !g.Battlefield.Contains(m1) || !g.Battlefield.Contains(island) || !me.Hand.Contains(alarm) {
		t.Fatal("the refused cast paid something anyway")
	}
	m2 := effectstest.SeedLand(g, me.ID, "Mountain B", "Basic Land — Mountain", "")
	if err := effectstest.CastForAltCost(g, alarm, "", "sacrifice", []uuid.UUID{m1, m2}, nil); err != nil {
		t.Fatalf("Mogg Alarm for two Mountains: %v", err)
	}
	if g.Battlefield.Contains(m1) || g.Battlefield.Contains(m2) {
		t.Error("a Mountain survived the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Goblin"); n != 2 {
		t.Errorf("%d Goblin tokens, want 2", n)
	}
	if !g.Battlefield.Contains(island) {
		t.Error("the Island was taken too")
	}
}

// --- Lawbringer and Lightbringer -----------------------------------------------------------

func TestNemB1BringersExileACreatureOfTheirColour(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, colour, other string
	}{
		{"Lawbringer", nemB1LawbringerOracle, "R", "B"},
		{"Lightbringer", nemB1LightbringerOracle, "B", "R"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me, them := effectstest.SeatsForCombat(g)
			bringer := nemB1Push(g, me.ID, tc.name, "Creature — Kor Rebel", tc.oracle, 2, 2, "W")
			victim := nemB1Push(g, them.ID, "Victim", "Creature — Test", "", 2, 2, tc.colour)
			wrong := nemB1Push(g, them.ID, "Wrong colour", "Creature — Test", "", 2, 2, tc.other)
			if err := g.ActivateCatalogAbility(me.ID, bringer, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(wrong)}); err == nil {
				t.Fatalf("%s targeted a creature of the wrong colour", tc.name)
			}
			if !g.Battlefield.Contains(bringer) {
				t.Fatal("the refused activation sacrificed the creature anyway")
			}
			if err := g.ActivateCatalogAbility(me.ID, bringer, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(victim)}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			if g.Battlefield.Contains(bringer) {
				t.Errorf("%s was not sacrificed as a cost", tc.name)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if !g.Exile.Contains(victim) {
				t.Error("the target was not exiled")
			}
			if !g.Battlefield.Contains(wrong) {
				t.Error("the other creature was touched")
			}
		})
	}
}

// --- Sliptide Serpent ------------------------------------------------------------------------

func TestNemB1SliptideSerpentReturnsToItsOwnersHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, them := effectstest.SeatsForCombat(g)
	serpent := nemB1Push(g, them.ID, "Sliptide Serpent", "Creature — Serpent", nemB1SerpentOracle, 4, 4, "U")
	effectstest.FindBattlefieldCardForTest(g, serpent).Controller = me.ID
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	effectstest.DcMana(t, g, me.ID, "{U}{U}{U}{U}")
	if err := g.ActivateCatalogAbility(me.ID, serpent, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !them.Hand.Contains(serpent) {
		t.Error("the Serpent did not go to its owner's hand")
	}
	if me.Hand.Contains(serpent) {
		t.Error("the Serpent went to its controller's hand, not its owner's")
	}
}

// --- the Seals ---------------------------------------------------------------------------

func TestNemB1SealOfFireDealsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, them := effectstest.SeatsForCombat(g)
	seal := nemB1Push(g, me.ID, "Seal of Fire", "Enchantment", nemB1SealOfFireOracle, 0, 0, "R")
	life := them.Life
	if err := g.ActivateCatalogAbility(me.ID, seal, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(them.ID)}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Graveyard.Contains(seal) {
		t.Error("the Seal was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - them.Life; got != 2 {
		t.Errorf("the target lost %d life, want 2", got)
	}
}

func TestNemB1SealOfRemovalBouncesTheCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, them := effectstest.SeatsForCombat(g)
	seal := nemB1Push(g, me.ID, "Seal of Removal", "Enchantment", nemB1SealOfRemovalOracle, 0, 0, "U")
	bear := nemB1Push(g, them.ID, "Grizzly Bears", "Creature — Bear", "", 2, 2, "G")
	land := effectstest.SeedLand(g, them.ID, "Forest", "Basic Land — Forest", "")
	if err := g.ActivateCatalogAbility(me.ID, seal, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(land)}); err == nil {
		t.Fatal("a land was accepted as the target creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, seal, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Graveyard.Contains(seal) {
		t.Error("the Seal was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !them.Hand.Contains(bear) {
		t.Error("the creature did not return to its owner's hand")
	}
}
