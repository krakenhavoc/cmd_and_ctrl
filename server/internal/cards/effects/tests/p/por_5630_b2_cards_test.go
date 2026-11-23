package testsp

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// por_5630_b2_cards_test.go — Portal cards, batch 2 (#5630). Every test
// plays the real catalog entry through the engine.

const (
	porB2BorderGuardOracle       = "1ef5003c-f540-4cdc-913f-7d5280ad9f62"
	porB2CoralEelOracle          = "10706fd1-7847-4316-be8d-59b56143ce45"
	porB2DevotedHeroOracle       = "fff6f2d0-22b5-41d0-99dc-bc817c3bd166"
	porB2FootSoldiersOracle      = "a768ba13-4d1c-4dce-a4a6-86a39c069c3f"
	porB2GiantOctopusOracle      = "c181d2a4-5959-4409-9bd3-ecedf8ec9516"
	porB2GoblinBullyOracle       = "56c5afdc-6777-45c7-8e2c-3cadecb95c5a"
	porB2GorillaWarriorOracle    = "c0e6ae0f-6cf7-48ca-acb8-73d4c38b9005"
	porB2HighlandGiantOracle     = "2f3878b8-61af-4276-b7ab-11b70e44eb62"
	porB2HornedTurtleOracle      = "7a4ae76b-6241-4fa8-b239-5ec941a0e7df"
	porB2KnightErrantOracle      = "920cc1a7-ac6b-44c0-ade7-347f05bba0f2"
	porB2LizardWarriorOracle     = "123337af-f976-4956-949f-c650da812113"
	porB2MinotaurWarriorOracle   = "53fb9a9b-1f5d-48c9-88c0-0864c187f15e"
	porB2MuckRatsOracle          = "bca13a12-6723-4a5e-8f1b-21646a8b3e7e"
	porB2RedwoodTreefolkOracle   = "2cdebbc2-2e0b-4d26-bb7f-4736f0318ca7"
	porB2RegalUnicornOracle      = "2b856a43-ca6b-4f08-b610-9794ee6d7fcf"
	porB2RowanTreefolkOracle     = "ae37dd07-2926-4801-b89d-3ffad9f0f575"
	porB2SkeletalCrocodileOracle = "90f9440a-0bf8-46eb-9cfa-16fb88a217d8"
	porB2SkeletalSnakeOracle     = "7af786a0-0851-4f87-ad0d-36b40527700c"
	porB2SpinedWurmOracle        = "05425f2f-7228-4bf5-8fe1-6fe99107e8e0"
	porB2WhiptailWurmOracle      = "50fa6a63-e031-47cb-8fd5-a6c235203722"
	porB2BogWraithOracle         = "508248d1-09a4-4e41-a4c9-286618e5061e"
	porB2FeralShadowOracle       = "cf26dcb4-e181-4ba5-bc03-b56f57032b85"
	porB2WallOfSwordsOracle      = "eb098958-50d3-4476-ba74-382033703ff9"
	porB2ChargingRhinoOracle     = "26966ecb-15d3-47e5-ab63-e38510c87ecc"
	porB2StalkingTigerOracle     = "e6af1429-276f-41a8-88d0-062661dc0cd4"
	porB2FleetFootedMonkOracle   = "a6a26261-1915-4699-a2e9-cca43b39a3eb"
	porB2SacredKnightOracle      = "fe8f5cc0-5849-437f-8d1f-cb191f2fc638"
	porB2ChargingBanditsOracle   = "0c67133b-be15-47d7-8cf6-79987a42044d"
	porB2ChargingPaladinOracle   = "6cccf44e-c05e-4239-b643-54b559c98552"
	porB2DreadReaperOracle       = "bd73ab86-0ac9-4ce0-be41-f4ad257e74f6"
	porB2SerpentWarriorOracle    = "fe249f27-db6c-4826-836e-a04efb1a3eaa"
	porB2FireImpOracle           = "5bd806e7-3f9b-4bb4-9708-f87a578f531e"
	porB2ThundermareOracle       = "58fa33fd-02a4-4ec9-9226-82594cfec363"
	porB2FireSnakeOracle         = "e96542ed-1931-4da1-9d9e-d10878c4ae6b"
	porB2NoxiousToadOracle       = "2fdc484e-b3c3-4f4e-99a1-26a1134aa1cd"
	porB2BreathOfLifeOracle      = "30d9e200-b944-43ff-89b8-a550a788ae03"
	porB2WickedPactOracle        = "39e21a5a-b278-478a-854c-17695c0f6246"
	porB2FireTempestOracle       = "ede6a352-52de-4ae0-816f-50140033dedf"
	porB2DrySpellOracle          = "980a9957-6b52-45c6-b847-f84974d5a653"
	porB2CloakOfFeathersOracle   = "cb4baf53-51ed-468b-a468-5d7d45a6dc26"
	porB2AngelicBlessingOracle   = "d3758fca-0522-4b5a-a1cc-3b2b3ab299ba"
	porB2RaiseDeadOracle         = "cbc9c731-181a-4f00-a7b0-eb7e56eac2ea"
	porB2TimeEbbOracle           = "30cc8f7b-3c28-40f5-8f8f-157e8212280b"
	porB2NaturalSpringOracle     = "f7571a2e-aaf3-4148-ab76-2a2e35273c70"
	porB2VampiricFeastOracle     = "1980ca2e-a415-4de1-ac30-7055507e82a2"
	porB2UntamedWildsOracle      = "b3b4c21d-f8d7-455f-be46-d5eb909d54df"
	porB2MysticDenialOracle      = "d2bd23a6-4f77-4d6e-bf8f-339cb7a4184d"
	porB2VirtuesRuinOracle       = "fc070e97-73ac-4028-9432-ff2012f2e778"
	porB2NaturesRuinOracle       = "d119ca9a-ee82-42c2-81e6-684205ca5183"
	porB2DevastationOracle       = "052838cb-dcf0-46f5-82b1-c3ed863b42b7"
	porB2NeedleStormOracle       = "1f513d1e-03ca-4e0f-af48-6490d71d5c41"
	porB2AssassinsBladeOracle    = "76da2150-34b9-4483-99df-131e1c5468d5"
	porB2ScorchingWindsOracle    = "db025c94-1182-412e-9665-b5ae89d26616"
	porB2TreetopDefenseOracle    = "f215d0f9-a53e-431a-a70d-9dc4e3caa41e"
)

// porB2Colored seeds a creature of the given colour for `owner`.
func porB2Colored(g *game.Game, owner uuid.UUID, name string, power, toughness int, colors ...string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Test",
		Power: power, Toughness: toughness, Colors: colors, Owner: owner, Controller: owner,
	})
}

// --- every card is in the catalog ----------------------------------------

func TestPor5630B2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		porB2BorderGuardOracle:       "Border Guard",
		porB2CoralEelOracle:          "Coral Eel",
		porB2DevotedHeroOracle:       "Devoted Hero",
		porB2FootSoldiersOracle:      "Foot Soldiers",
		porB2GiantOctopusOracle:      "Giant Octopus",
		porB2GoblinBullyOracle:       "Goblin Bully",
		porB2GorillaWarriorOracle:    "Gorilla Warrior",
		porB2HighlandGiantOracle:     "Highland Giant",
		porB2HornedTurtleOracle:      "Horned Turtle",
		porB2KnightErrantOracle:      "Knight Errant",
		porB2LizardWarriorOracle:     "Lizard Warrior",
		porB2MinotaurWarriorOracle:   "Minotaur Warrior",
		porB2MuckRatsOracle:          "Muck Rats",
		porB2RedwoodTreefolkOracle:   "Redwood Treefolk",
		porB2RegalUnicornOracle:      "Regal Unicorn",
		porB2RowanTreefolkOracle:     "Rowan Treefolk",
		porB2SkeletalCrocodileOracle: "Skeletal Crocodile",
		porB2SkeletalSnakeOracle:     "Skeletal Snake",
		porB2SpinedWurmOracle:        "Spined Wurm",
		porB2WhiptailWurmOracle:      "Whiptail Wurm",
		porB2BogWraithOracle:         "Bog Wraith",
		porB2FeralShadowOracle:       "Feral Shadow",
		porB2WallOfSwordsOracle:      "Wall of Swords",
		porB2ChargingRhinoOracle:     "Charging Rhino",
		porB2StalkingTigerOracle:     "Stalking Tiger",
		porB2FleetFootedMonkOracle:   "Fleet-Footed Monk",
		porB2SacredKnightOracle:      "Sacred Knight",
		porB2ChargingBanditsOracle:   "Charging Bandits",
		porB2ChargingPaladinOracle:   "Charging Paladin",
		porB2DreadReaperOracle:       "Dread Reaper",
		porB2SerpentWarriorOracle:    "Serpent Warrior",
		porB2FireImpOracle:           "Fire Imp",
		porB2ThundermareOracle:       "Thundermare",
		porB2FireSnakeOracle:         "Fire Snake",
		porB2NoxiousToadOracle:       "Noxious Toad",
		porB2BreathOfLifeOracle:      "Breath of Life",
		porB2WickedPactOracle:        "Wicked Pact",
		porB2FireTempestOracle:       "Fire Tempest",
		porB2DrySpellOracle:          "Dry Spell",
		porB2CloakOfFeathersOracle:   "Cloak of Feathers",
		porB2AngelicBlessingOracle:   "Angelic Blessing",
		porB2RaiseDeadOracle:         "Raise Dead",
		porB2TimeEbbOracle:           "Time Ebb",
		porB2NaturalSpringOracle:     "Natural Spring",
		porB2VampiricFeastOracle:     "Vampiric Feast",
		porB2UntamedWildsOracle:      "Untamed Wilds",
		porB2MysticDenialOracle:      "Mystic Denial",
		porB2VirtuesRuinOracle:       "Virtue's Ruin",
		porB2NaturesRuinOracle:       "Nature's Ruin",
		porB2DevastationOracle:       "Devastation",
		porB2NeedleStormOracle:       "Needle Storm",
		porB2AssassinsBladeOracle:    "Assassin's Blade",
		porB2ScorchingWindsOracle:    "Scorching Winds",
		porB2TreetopDefenseOracle:    "Treetop Defense",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v; want Full with none", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- vanilla bodies ------------------------------------------------------------

// A vanilla body has nothing but its printed card: no ability, no
// keyword, and it can attack and block.
func TestPor5630B2VanillaBodiesHaveNoAbilities(t *testing.T) {
	for oracle, name := range map[string]string{
		porB2BorderGuardOracle: "Border Guard", porB2CoralEelOracle: "Coral Eel",
		porB2DevotedHeroOracle: "Devoted Hero", porB2FootSoldiersOracle: "Foot Soldiers",
		porB2GiantOctopusOracle: "Giant Octopus", porB2GoblinBullyOracle: "Goblin Bully",
		porB2GorillaWarriorOracle: "Gorilla Warrior", porB2HighlandGiantOracle: "Highland Giant",
		porB2HornedTurtleOracle: "Horned Turtle", porB2KnightErrantOracle: "Knight Errant",
		porB2LizardWarriorOracle: "Lizard Warrior", porB2MinotaurWarriorOracle: "Minotaur Warrior",
		porB2MuckRatsOracle: "Muck Rats", porB2RedwoodTreefolkOracle: "Redwood Treefolk",
		porB2RegalUnicornOracle: "Regal Unicorn", porB2RowanTreefolkOracle: "Rowan Treefolk",
		porB2SkeletalCrocodileOracle: "Skeletal Crocodile", porB2SkeletalSnakeOracle: "Skeletal Snake",
		porB2SpinedWurmOracle: "Spined Wurm", porB2WhiptailWurmOracle: "Whiptail Wurm",
	} {
		spec, _ := effects.Lookup(oracle)
		if len(spec.Triggered) != 0 || len(spec.Activated) != 0 || len(spec.PrintedKeywords) != 0 ||
			len(spec.BlockRules) != 0 || spec.OnResolve != nil {
			t.Errorf("%s registers an ability; it is a vanilla creature", name)
		}
	}
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wurm := effectstest.B12Push(g, me.ID, "Spined Wurm", "Creature — Wurm", porB2SpinedWurmOracle, 5, 4)
	guard := effectstest.B12Push(g, opp.ID, "Border Guard", "Creature — Human Soldier", porB2BorderGuardOracle, 1, 4)
	if err := effectstest.PorBlock(t, g, wurm, opp.ID, guard); err != nil {
		t.Errorf("Border Guard couldn't block Spined Wurm: %v", err)
	}
}

// --- keyword bodies ------------------------------------------------------------

func TestPor5630B2KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Bog Wraith", "Creature — Wraith", porB2BogWraithOracle, []string{"swampwalk"}},
		{"Feral Shadow", "Creature — Nightstalker", porB2FeralShadowOracle, []string{"flying"}},
		{"Wall of Swords", "Creature — Wall", porB2WallOfSwordsOracle, []string{"defender", "flying"}},
		{"Dread Reaper", "Creature — Horror", porB2DreadReaperOracle, []string{"flying"}},
		{"Thundermare", "Creature — Elemental Horse", porB2ThundermareOracle, []string{"haste"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			for _, kw := range tc.keywords {
				if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

func TestPor5630B2BogWraithCantBeBlockedWhileTheDefenderHasASwamp(t *testing.T) {
	for _, hasSwamp := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		wraith := effectstest.B12Push(g, me.ID, "Bog Wraith", "Creature — Wraith", porB2BogWraithOracle, 3, 3)
		bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
		if hasSwamp {
			effectstest.B12Permanent(g, opp.ID, "Swamp", "Basic Land — Swamp")
		}
		err := effectstest.PorBlock(t, g, wraith, opp.ID, bear)
		if hasSwamp && err == nil {
			t.Error("Bog Wraith was blocked while the defender controls a Swamp")
		}
		if !hasSwamp && err != nil {
			t.Errorf("Bog Wraith couldn't be blocked with no Swamp: %v", err)
		}
	}
}

func TestPor5630B2WallOfSwordsCantAttackButBlocksAFlyer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := effectstest.B12Push(g, me.ID, "Wall of Swords", "Creature — Wall", porB2WallOfSwordsOracle, 3, 5)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(wall, opp.ID); err == nil {
		t.Error("Wall of Swords attacked")
	}
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	bird := effectstest.PorFlyer(g, me.ID, "Bird")
	wall = effectstest.B12Push(g, opp.ID, "Wall of Swords", "Creature — Wall", porB2WallOfSwordsOracle, 3, 5)
	if err := effectstest.PorBlock(t, g, bird, opp.ID, wall); err != nil {
		t.Errorf("Wall of Swords couldn't block a flyer: %v", err)
	}
}

// --- block restrictions --------------------------------------------------------

func TestPor5630B2RhinoAndTigerRefuseASecondBlocker(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Charging Rhino", "Creature — Rhino", porB2ChargingRhinoOracle},
		{"Stalking Tiger", "Creature — Cat", porB2StalkingTigerOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			beast := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 4, 4)
			a := effectstest.B12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
			b := effectstest.B12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
			effectstest.BrAttack(t, g, beast)
			effectstest.BrRefusal(t, g.DeclareBlockers([]game.BlockDeclaration{
				{Blocker: a, Attacker: beast}, {Blocker: b, Attacker: beast},
			}), game.BlockReasonTooManyBlockers)
			if err := g.DeclareBlockers([]game.BlockDeclaration{{Blocker: a, Attacker: beast}}); err != nil {
				t.Errorf("a single blocker was refused: %v", err)
			}
		})
	}
}

func TestPor5630B2FleetFootedMonkEvadesPowerTwoOrGreater(t *testing.T) {
	for _, tc := range []struct {
		power int
		want  bool
	}{{2, false}, {3, false}, {1, true}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		monk := effectstest.B12Push(g, me.ID, "Fleet-Footed Monk", "Creature — Human Monk", porB2FleetFootedMonkOracle, 1, 1)
		blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Test", tc.power, 4)
		err := effectstest.PorBlock(t, g, monk, opp.ID, blocker)
		if tc.want && err != nil {
			t.Errorf("a %d-power creature couldn't block the Monk: %v", tc.power, err)
		}
		if !tc.want && err == nil {
			t.Errorf("a %d-power creature blocked the Monk", tc.power)
		}
	}
}

func TestPor5630B2SacredKnightEvadesBlackAndRedCreatures(t *testing.T) {
	for _, tc := range []struct {
		colors []string
		want   bool
	}{{[]string{"B"}, false}, {[]string{"R"}, false}, {[]string{"B", "R"}, false}, {[]string{"G"}, true}, {[]string{"W", "G"}, true}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		knight := effectstest.B12Push(g, me.ID, "Sacred Knight", "Creature — Human Knight", porB2SacredKnightOracle, 3, 2)
		blocker := porB2Colored(g, opp.ID, "Blocker", 2, 2, tc.colors...)
		err := effectstest.PorBlock(t, g, knight, opp.ID, blocker)
		if tc.want && err != nil {
			t.Errorf("a %v creature couldn't block Sacred Knight: %v", tc.colors, err)
		}
		if !tc.want && err == nil {
			t.Errorf("a %v creature blocked Sacred Knight", tc.colors)
		}
	}
}

// --- attack triggers -----------------------------------------------------------

func TestPor5630B2ChargingCreaturesPumpWhenTheyAttack(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, tough           int
		wantP, wantT           int
	}{
		{"Charging Bandits", "Creature — Human Rogue", porB2ChargingBanditsOracle, 3, 3, 5, 3},
		{"Charging Paladin", "Creature — Human Knight", porB2ChargingPaladinOracle, 2, 2, 2, 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, them := effectstest.SeatsForCombat(g)
			id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.tough)
			effectstest.DeclareAttack(t, g, them.ID, id)
			effectstest.PassPriorityAroundTable(t, g)
			if p, tough := effectstest.PtOf(t, g, id); p != tc.wantP || tough != tc.wantT {
				t.Errorf("%s is %d/%d while attacking, want %d/%d", tc.name, p, tough, tc.wantP, tc.wantT)
			}
		})
	}
}

// --- enters and dies triggers --------------------------------------------------

func TestPor5630B2EntersLoseLife(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, tough, loss     int
	}{
		{"Dread Reaper", "Creature — Horror", porB2DreadReaperOracle, 6, 5, 5},
		{"Serpent Warrior", "Creature — Snake Warrior", porB2SerpentWarriorOracle, 3, 3, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			life := me.Life
			effectstest.CastHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, tc.power, tc.tough)
			if me.Life != life-tc.loss {
				t.Errorf("life %d -> %d, want -%d", life, me.Life, tc.loss)
			}
		})
	}
}

func TestPor5630B2FireImpDealsTwoToTargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Fire Imp", "Creature — Imp", porB2FireImpOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err == nil {
		t.Fatal("a player was accepted as the target")
	}
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the 2/2 survived 2 damage")
	}
}

func TestPor5630B2ThundermareTapsEveryOtherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, them.ID, "Theirs", "Creature — Bear", 2, 2)
	mare := effectstest.CastHoldCreature(t, g, "Thundermare", "Creature — Elemental Horse", porB2ThundermareOracle, 5, 5)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, mine) || !effectstest.B16Tapped(t, g, theirs) {
		t.Error("another creature was left untapped")
	}
	if effectstest.B16Tapped(t, g, mare) {
		t.Error("Thundermare tapped itself")
	}
}

func TestPor5630B2FireSnakeDestroysATargetLandWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	snake := effectstest.B12Push(g, me.ID, "Fire Snake", "Creature — Snake", porB2FireSnakeOracle, 3, 1)
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(snake) })
	g.RunStateChecksForTest()
	effectstest.B04WaitForPick(t, g, me.ID)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
		t.Fatal("a creature was accepted as the target")
	}
	effectstest.PickTriggerTarget(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land survived")
	}
}

func TestPor5630B2NoxiousToadMakesEachOpponentDiscardWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	toad := effectstest.B12Push(g, me.ID, "Noxious Toad", "Creature — Frog", porB2NoxiousToadOracle, 1, 1)
	effectstest.B15Destroy(t, g, toad)
	for i, p := range g.Seats {
		owed := effectstest.DiscardOwed(g, p.ID)
		if i == 0 && owed != 0 {
			t.Errorf("you owe %d discards, want none", owed)
		}
		if i != 0 && owed != 1 {
			t.Errorf("seat %d owes %d discards, want 1", i, owed)
		}
	}
}

// --- spells --------------------------------------------------------------------

func TestPor5630B2BreathOfLifeReanimatesYourCreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.SeedGraveyardCreature(me, "My Zombie", "{4}{B}")
	effectstest.CastCatalogSpell(t, g, "Breath of Life", "Sorcery", porB2BreathOfLifeOracle, effectstest.CardTarget(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if card, ok := effectstest.BattlefieldCard(g, mine); !ok || card.Controller != me.ID {
		t.Error("the creature card did not return to the battlefield under your control")
	}
}

func TestPor5630B2RaiseDeadReturnsACreatureCardToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.PushGraveyardCardTyped(caster, "Bear", "Creature — Bear")
	rock := effectstest.PushGraveyardCardTyped(caster, "Rock", "Artifact")
	if err := effectstest.CastCatalogSpellErr(t, g, "Raise Dead", "Sorcery", porB2RaiseDeadOracle, effectstest.CardTarget(rock)); err == nil {
		t.Fatal("Raise Dead targeted a noncreature card")
	}
	effectstest.CastCatalogSpell(t, g, "Raise Dead", "Sorcery", porB2RaiseDeadOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !caster.Hand.Contains(bear) {
		t.Error("the creature card did not return to your hand")
	}
}

func TestPor5630B2WickedPactDestroysTwoNonblackAndCostsFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := porB2Colored(g, opp.ID, "Green A", 2, 2, "G")
	b := porB2Colored(g, opp.ID, "Green B", 2, 2, "G")
	black := porB2Colored(g, opp.ID, "Black", 2, 2, "B")
	if err := effectstest.CastCatalogSpellErr(t, g, "Wicked Pact", "Sorcery", porB2WickedPactOracle, effectstest.CardTargets(a, black)); err == nil {
		t.Fatal("Wicked Pact targeted a black creature")
	}
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Wicked Pact", "Sorcery", porB2WickedPactOracle, effectstest.CardTargets(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("a target survived")
	}
	if me.Life != life-5 {
		t.Errorf("life %d -> %d, want -5", life, me.Life)
	}
}

func TestPor5630B2SweepsDamageEachCreatureAndEachPlayer(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		amount       int
	}{
		{"Fire Tempest", porB2FireTempestOracle, 6},
		{"Dry Spell", porB2DrySpellOracle, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Test", 1, tc.amount+1)
			theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Test", 1, tc.amount)
			lives := effectstest.B27Lives(g)
			effectstest.CastCatalogSpell(t, g, tc.name, "Sorcery", tc.oracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			for i, l := range effectstest.B27Lives(g) {
				if l != lives[i]-tc.amount {
					t.Errorf("seat %d life %d -> %d, want -%d", i, lives[i], l, tc.amount)
				}
			}
			if g.Battlefield.Contains(theirs) {
				t.Errorf("a %d-toughness creature survived", tc.amount)
			}
			if got := effectstest.CardByID(g, mine).DamageMarked; got != tc.amount {
				t.Errorf("%d damage marked on the survivor, want %d", got, tc.amount)
			}
		})
	}
}

func TestPor5630B2NeedleStormHitsOnlyFlyers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bird := effectstest.PorFlyer(g, opp.ID, "Bird")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	lives := effectstest.B27Lives(g)
	effectstest.CastCatalogSpell(t, g, "Needle Storm", "Sorcery", porB2NeedleStormOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bird) {
		t.Error("the flyer survived 4 damage")
	}
	if got := effectstest.CardByID(g, bear).DamageMarked; got != 0 {
		t.Errorf("the ground creature has %d damage, want 0", got)
	}
	for i, l := range effectstest.B27Lives(g) {
		if l != lives[i] {
			t.Errorf("seat %d life changed: %d -> %d", i, lives[i], l)
		}
	}
	_ = me
}

func TestPor5630B2ColourWraths(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, color string
	}{
		{"Virtue's Ruin", porB2VirtuesRuinOracle, "W"},
		{"Nature's Ruin", porB2NaturesRuinOracle, "G"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			mine := porB2Colored(g, me.ID, "Mine", 2, 2, tc.color)
			theirs := porB2Colored(g, opp.ID, "Theirs", 2, 2, tc.color, "U")
			other := porB2Colored(g, opp.ID, "Other", 2, 2, "R")
			effectstest.CastCatalogSpell(t, g, tc.name, "Sorcery", tc.oracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
				t.Errorf("a %s creature survived", tc.color)
			}
			if !g.Battlefield.Contains(other) {
				t.Error("a creature of another colour was destroyed")
			}
		})
	}
}

func TestPor5630B2DevastationDestroysCreaturesAndLandsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	forest := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Devastation", "Sorcery", porB2DevastationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || g.Battlefield.Contains(forest) {
		t.Error("a creature or land survived")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("a noncreature, nonland permanent was destroyed")
	}
}

func TestPor5630B2CloakOfFeathersGivesFlyingAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Cloak of Feathers", "Sorcery", porB2CloakOfFeathersOracle, effectstest.CardTarget(bear))
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the creature did not gain flying")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, me.Hand.Size())
	}
}

func TestPor5630B2AngelicBlessingPumpsAndGivesFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.B12Creature(g, g.Seats[0].ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Angelic Blessing", "Sorcery", porB2AngelicBlessingOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("the creature is %d/%d, want 5/5", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the creature did not gain flying")
	}
}

func TestPor5630B2TimeEbbPutsACreatureOnTopOfItsOwnersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.SeedLibrary(opp, "Under")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	if err := effectstest.CastCatalogSpellErr(t, g, "Time Ebb", "Sorcery", porB2TimeEbbOracle, effectstest.CardTarget(rock)); err == nil {
		t.Fatal("Time Ebb targeted a noncreature permanent")
	}
	effectstest.CastCatalogSpell(t, g, "Time Ebb", "Sorcery", porB2TimeEbbOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the creature is still on the battlefield")
	}
	if top := effectstest.LibraryTopIDs(opp, 1); top[0] != bear {
		t.Error("the creature is not on top of its owner's library")
	}
}

func TestPor5630B2NaturalSpringGivesTargetPlayerEight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Natural Spring", "Sorcery", porB2NaturalSpringOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life+8 {
		t.Errorf("life %d -> %d, want +8", life, opp.Life)
	}
}

func TestPor5630B2VampiricFeastDealsFourAndGainsFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine, theirs := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "Vampiric Feast", "Sorcery", porB2VampiricFeastOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirs-4 || me.Life != mine+4 {
		t.Errorf("target %d -> %d (want -4), you %d -> %d (want +4)", theirs, opp.Life, mine, me.Life)
	}
}

func TestPor5630B2UntamedWildsPutsABasicLandOntoTheBattlefieldUntapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.CastCatalogSpell(t, g, "Untamed Wilds", "Sorcery", porB2UntamedWildsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	forest := effectstest.FindBattlefieldByName(g, "Forest")
	if forest == uuid.Nil {
		t.Fatal("Untamed Wilds did not fetch the Forest")
	}
	if effectstest.B16Tapped(t, g, forest) {
		t.Error("the land entered tapped")
	}
}

func TestPor5630B2MysticDenialCountersCreatureOrSorceryOnly(t *testing.T) {
	for _, tc := range []struct {
		typeLine string
		want     bool
	}{
		{"Creature — Bear", true},
		{"Sorcery", true},
		{"Instant", false},
		{"Artifact", false},
	} {
		t.Run(tc.typeLine, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			var spell uuid.UUID
			if tc.typeLine == "Instant" {
				spell = effectstest.B13OpponentCasts(t, g, opp, "Their Instant", "Instant", "", "{R}", nil)
			} else {
				spell = effectstest.B13PlayAs(t, g, 1, "Their Spell", tc.typeLine, "")
			}
			err := g.CastSpell(me.ID, effectstest.B13HandCard(me, "Mystic Denial", "Instant", porB2MysticDenialOracle),
				game.CastSpellParams{Targets: effectstest.B16TargetCard(spell)})
			if !tc.want {
				if err == nil {
					t.Errorf("Mystic Denial targeted a %s spell", tc.typeLine)
				}
				return
			}
			if err != nil {
				t.Fatalf("Mystic Denial at a %s spell: %v", tc.typeLine, err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if !opp.Graveyard.Contains(spell) || g.Battlefield.Contains(spell) {
				t.Errorf("the %s spell was not countered", tc.typeLine)
			}
		})
	}
}

// --- "only if you've been attacked this step" ----------------------------------

func TestPor5630B2AttackedSpellsNeedYouAttackedThisStep(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Assassin's Blade", porB2AssassinsBladeOracle},
		{"Scorching Winds", porB2ScorchingWindsOracle},
		{"Treetop Defense", porB2TreetopDefenseOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
			if err := effectstest.CastCatalogSpellErr(t, g, tc.name, "Instant", tc.oracle, nil); err == nil {
				t.Fatal("cast outside the declare attackers step")
			}
			att := effectstest.SfPush(g, me.ID, "Attacker", "Creature — Test", 3, []string{"R"})
			effectstest.SfAttackHold(t, g, third.ID, att)
			card := effectstest.SfInstant(tc.name, tc.oracle)
			card.InstanceID, card.Owner, card.Controller = uuid.New(), opp.ID, opp.ID
			opp.Hand.PushTop(card)
			var targets []game.TargetRef
			if tc.name == "Assassin's Blade" {
				targets = effectstest.CardTarget(att)
			}
			if err := g.CastSpell(opp.ID, card.InstanceID, game.CastSpellParams{Targets: targets}); err == nil {
				t.Fatal("cast by a player nobody attacked")
			}
		})
	}
}

func TestPor5630B2AssassinsBladeDestroysANonblackAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	red := effectstest.SfPush(g, me.ID, "Red Attacker", "Creature — Test", 3, []string{"R"})
	black := effectstest.SfPush(g, me.ID, "Black Attacker", "Creature — Test", 3, []string{"B"})
	home := effectstest.SfPush(g, me.ID, "Home", "Creature — Test", 1, []string{"R"})
	effectstest.SfAttackHold(t, g, opp.ID, red, black)
	card := effectstest.SfInstant("Assassin's Blade", porB2AssassinsBladeOracle)
	card.InstanceID, card.Owner, card.Controller = uuid.New(), opp.ID, opp.ID
	opp.Hand.PushTop(card)
	for _, bad := range []uuid.UUID{black, home} {
		if err := g.CastSpell(opp.ID, card.InstanceID, game.CastSpellParams{Targets: effectstest.CardTarget(bad)}); err == nil {
			t.Fatal("Assassin's Blade targeted a black attacker or a creature that isn't attacking")
		}
	}
	if err := g.CastSpell(opp.ID, card.InstanceID, game.CastSpellParams{Targets: effectstest.CardTarget(red)}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(red) {
		t.Error("the nonblack attacker survived")
	}
	if !g.Battlefield.Contains(black) || !g.Battlefield.Contains(home) {
		t.Error("another creature was destroyed")
	}
}

func TestPor5630B2ScorchingWindsDamagesEachAttackingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	toOpp := effectstest.SfPush(g, me.ID, "At You", "Creature — Test", 3, []string{"R"})
	toThird := effectstest.SfPush(g, me.ID, "At Them", "Creature — Test", 3, []string{"R"})
	home := effectstest.SfPush(g, me.ID, "Home", "Creature — Test", 1, []string{"R"})
	blocker := effectstest.SfPush(g, opp.ID, "Blocker", "Creature — Test", 2, []string{"G"})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(toOpp, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.DeclareAttacker(toThird, third.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	g.RunStateChecksForTest()
	lives := effectstest.B27Lives(g)
	effectstest.SfCast(t, g, opp, effectstest.SfInstant("Scorching Winds", porB2ScorchingWindsOracle), game.CastSpellParams{})
	for _, id := range []uuid.UUID{toOpp, toThird} {
		if got := effectstest.CardByID(g, id).DamageMarked; got != 1 {
			t.Errorf("an attacker has %d damage, want 1", got)
		}
	}
	for _, id := range []uuid.UUID{home, blocker} {
		if got := effectstest.CardByID(g, id).DamageMarked; got != 0 {
			t.Errorf("a creature that isn't attacking has %d damage, want 0", got)
		}
	}
	for i, l := range effectstest.B27Lives(g) {
		if l != lives[i] {
			t.Errorf("seat %d life changed: %d -> %d", i, lives[i], l)
		}
	}
}

func TestPor5630B2TreetopDefenseGivesYourCreaturesReach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bird := effectstest.PorFlyer(g, me.ID, "Bird")
	// The defender casts it, so "you" is opp.
	defenders := effectstest.SfPush(g, opp.ID, "Ground Blocker", "Creature — Test", 2, []string{"G"})
	attackers := effectstest.SfPush(g, me.ID, "Attacker's Other", "Creature — Test", 2, []string{"R"})
	effectstest.SfAttackHold(t, g, opp.ID, bird)
	effectstest.SfCast(t, g, opp, effectstest.SfInstant("Treetop Defense", porB2TreetopDefenseOracle), game.CastSpellParams{})
	if !effectstest.HasEffectiveKeyword(t, g, defenders, "reach") {
		t.Error("the caster's creature did not gain reach")
	}
	if effectstest.HasEffectiveKeyword(t, g, attackers, "reach") {
		t.Error("an opponent's creature gained reach")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(defenders, bird); err != nil {
		t.Errorf("your creature couldn't block the flyer with reach: %v", err)
	}
}
