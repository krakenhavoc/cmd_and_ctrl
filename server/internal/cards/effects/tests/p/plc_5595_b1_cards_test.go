package testsp

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// plc_5595_b1_cards_test.go — Planar Chaos, batch 1 (#5595, S82): forty-nine
// cards built from seams the engine already has. Every test plays the
// real catalog entry through the engine.

const (
	plcB1BloodKnightOracle        = "67fba605-9cfa-499c-83e0-4fbd023bcfa0"
	plcB1NeedlepeakSpiderOracle   = "169429d8-aa98-46a8-8681-521c65dafbb4"
	plcB1SerraSphinxOracle        = "9b344428-ec35-45b9-9714-9ca58a338f5c"
	plcB1CalcidermOracle          = "329d53eb-d443-4376-ae5a-4be7dfb1fc91"
	plcB1WaningWurmOracle         = "4921a314-fce7-4f3c-8546-6cf520a2c104"
	plcB1AvenRiftwatcherOracle    = "a4d22abe-87c5-4f6b-ad14-4a7935883169"
	plcB1KeldonMaraudersOracle    = "e05b850f-24cd-42b9-bc4f-0f05f9aea679"
	plcB1BatteringSliverOracle    = "4adf959b-2e09-47ff-a1b1-b889348735b8"
	plcB1ReflexSliverOracle       = "80e4e857-3f7e-4344-bf41-64d3b32f328f"
	plcB1SpittingSliverOracle     = "8415a482-3957-4c00-9015-99defc049877"
	plcB1SynchronousSliverOracle  = "b90e2fb8-9b09-4d95-a8a2-5287b5f18132"
	plcB1DarkheartSliverOracle    = "88e5b905-fe52-4fd6-8cdf-b7ede5166ebc"
	plcB1PoulticeSliverOracle     = "6be3cf35-f774-4afb-8c1f-5c9b616a2b3f"
	plcB1DormantSliverOracle      = "2bc3d5b2-076f-4e01-87be-28d7eb70a2db"
	plcB1ProdigalPyromancerOracle = "29b3cb3e-aeb1-43eb-8095-9f7881aa34bd"
	plcB1RathiTrapperOracle       = "0920d6aa-eab7-4116-b2f8-b13080d5d7e3"
	plcB1SaltfieldRecluseOracle   = "c87e056a-6067-49e6-b3f7-bba8234777de"
	plcB1SerendibSorcererOracle   = "73a59b03-168e-472e-bbb1-cabd782ea014"
	plcB1MagusOfTheBazaarOracle   = "e08d264d-5731-4e2a-ae6d-5a74ef648575"
	plcB1ReveredDeadOracle        = "96a57ce0-f1ab-4af7-8c18-8e35aad11780"
	plcB1MalachOfTheDawnOracle    = "c351a325-42cb-43ad-8648-c204e2317b50"
	plcB1HedgeTrollOracle         = "500663c1-390c-4598-b614-4168c4f7debe"
	plcB1MireBoaOracle            = "e9a3a466-df78-49a0-add3-975b94c5cd09"
	plcB1GroundbreakerOracle      = "a14dc46f-b790-458e-8821-0c5646188680"
	plcB1GaeasAnthemOracle        = "3754dce0-3e97-406f-8807-a4942a222c41"
	plcB1ManaTitheOracle          = "948217ac-289e-455f-8c7c-519aa75fa345"
	plcB1SunlanceOracle           = "3c8cc7e2-7bff-46e0-bb5b-1e721a56a3ef"
	plcB1CradleToGraveOracle      = "a86f9b53-ff44-4b08-a8e9-beef73595de1"
	plcB1DismalFailureOracle      = "1feeb881-ebe6-4afb-ba31-3d116d8ab7a1"
	plcB1WistfulThinkingOracle    = "a936b84c-c350-4a1f-b4a7-86316f11c67f"
	plcB1SkirkShamanOracle        = "78d5ce34-f160-4925-a4ce-25a3546d97c7"
	plcB1DustCoronaOracle         = "3f9ab49b-81fc-487e-8613-e30afcae164d"
	plcB1GossamerPhantasmOracle   = "5ed23ed2-be2f-4958-affe-429f58e9dca5"
	plcB1VampiricLinkOracle       = "3f4521b4-1efd-4592-a50d-5ee9918df177"
	plcB1UtopiaVowOracle          = "917652c7-ac12-4cd7-ab2d-f2260a620d02"
	plcB1MelancholyOracle         = "58cc7682-f8ce-47cf-a291-0301fea9ddfd"
	plcB1EnslaveOracle            = "879ba7c4-4557-4166-adb0-b52454816e5a"
	plcB1GhostTacticianOracle     = "8d89aba8-20ce-44cd-9d82-3d16fdf6aac8"
	plcB1RidgedKusiteOracle       = "a0c3c26e-e9e8-424d-922e-a692bb479ebd"
	plcB1SophicCentaurOracle      = "5e8e443b-2102-4d91-b50c-1679dc893ef7"
	plcB1CitanulWoodreadersOracle = "a4555bd2-be42-442a-8ead-3398318e91c5"
	plcB1GiantDustwaspOracle      = "5f4e45c1-2bd2-4741-bb97-b6a324c574ba"
	plcB1HeroesRememberedOracle   = "2f30ffd2-9910-4fbd-96f9-d66f1da506e9"
	plcB1ShadeOfTrokairOracle     = "61834b70-1b3e-49e6-aaf4-ee7d72155e51"
	plcB1AuramancersGuiseOracle   = "b761f6d9-1063-40f8-a3f1-25b1f9b46963"
	plcB1MidnightCharmOracle      = "ed228f67-3adf-46d4-ac45-e0278592850d"
	plcB1PiracyCharmOracle        = "dffde884-15bc-4985-90e2-a0888ef4cb48"
	plcB1EvolutionCharmOracle     = "5f73fd4e-0d24-4daa-8d82-73c0cf7f801d"
	plcB1HealingLeavesOracle      = "9c675e7a-1d68-44b2-9c1a-543b73782437"

	plcB1BearType  = "Creature — Bear"
	plcB1Sliver    = "Creature — Sliver"
	plcB1Instant   = "Instant"
	plcB1Sorcery   = "Sorcery"
	plcB1FirstStrk = "first strike"
)

func TestPlcB1EveryCardIsRegisteredFull(t *testing.T) {
	names := map[string]string{
		plcB1BloodKnightOracle:        "Blood Knight",
		plcB1NeedlepeakSpiderOracle:   "Needlepeak Spider",
		plcB1SerraSphinxOracle:        "Serra Sphinx",
		plcB1CalcidermOracle:          "Calciderm",
		plcB1WaningWurmOracle:         "Waning Wurm",
		plcB1AvenRiftwatcherOracle:    "Aven Riftwatcher",
		plcB1KeldonMaraudersOracle:    "Keldon Marauders",
		plcB1BatteringSliverOracle:    "Battering Sliver",
		plcB1ReflexSliverOracle:       "Reflex Sliver",
		plcB1SpittingSliverOracle:     "Spitting Sliver",
		plcB1SynchronousSliverOracle:  "Synchronous Sliver",
		plcB1DarkheartSliverOracle:    "Darkheart Sliver",
		plcB1PoulticeSliverOracle:     "Poultice Sliver",
		plcB1DormantSliverOracle:      "Dormant Sliver",
		plcB1ProdigalPyromancerOracle: "Prodigal Pyromancer",
		plcB1RathiTrapperOracle:       "Rathi Trapper",
		plcB1SaltfieldRecluseOracle:   "Saltfield Recluse",
		plcB1SerendibSorcererOracle:   "Serendib Sorcerer",
		plcB1MagusOfTheBazaarOracle:   "Magus of the Bazaar",
		plcB1ReveredDeadOracle:        "Revered Dead",
		plcB1MalachOfTheDawnOracle:    "Malach of the Dawn",
		plcB1HedgeTrollOracle:         "Hedge Troll",
		plcB1MireBoaOracle:            "Mire Boa",
		plcB1GroundbreakerOracle:      "Groundbreaker",
		plcB1GaeasAnthemOracle:        "Gaea's Anthem",
		plcB1ManaTitheOracle:          "Mana Tithe",
		plcB1SunlanceOracle:           "Sunlance",
		plcB1CradleToGraveOracle:      "Cradle to Grave",
		plcB1DismalFailureOracle:      "Dismal Failure",
		plcB1WistfulThinkingOracle:    "Wistful Thinking",
		plcB1SkirkShamanOracle:        "Skirk Shaman",
		plcB1DustCoronaOracle:         "Dust Corona",
		plcB1GossamerPhantasmOracle:   "Gossamer Phantasm",
		plcB1VampiricLinkOracle:       "Vampiric Link",
		plcB1UtopiaVowOracle:          "Utopia Vow",
		plcB1MelancholyOracle:         "Melancholy",
		plcB1EnslaveOracle:            "Enslave",
		plcB1GhostTacticianOracle:     "Ghost Tactician",
		plcB1RidgedKusiteOracle:       "Ridged Kusite",
		plcB1SophicCentaurOracle:      "Sophic Centaur",
		plcB1CitanulWoodreadersOracle: "Citanul Woodreaders",
		plcB1GiantDustwaspOracle:      "Giant Dustwasp",
		plcB1HeroesRememberedOracle:   "Heroes Remembered",
		plcB1ShadeOfTrokairOracle:     "Shade of Trokair",
		plcB1AuramancersGuiseOracle:   "Auramancer's Guise",
		plcB1MidnightCharmOracle:      "Midnight Charm",
		plcB1PiracyCharmOracle:        "Piracy Charm",
		plcB1EvolutionCharmOracle:     "Evolution Charm",
		plcB1HealingLeavesOracle:      "Healing Leaves",
	}
	if len(names) != 49 {
		t.Fatalf("%d cards in the batch, want 49", len(names))
	}
	for oracle, name := range names {
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

// --- helpers -----------------------------------------------------------------

// plcB1EnteredThisTurn records `id` in the turn's per-object entry tally,
// as a creature that entered the battlefield this turn.
func plcB1EnteredThisTurn(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		if g.TurnTally.Entered == nil {
			g.TurnTally.Entered = map[uuid.UUID]int{}
		}
		g.TurnTally.Entered[id]++
	})
}

// plcB1Regenerates activates row `idx` of `id` with `mana` floating, then
// destroys it: the regeneration shield keeps it on the battlefield.
func plcB1Regenerates(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, idx int, mana ...string) {
	t.Helper()
	effectstest.ApaMana(me, mana...)
	effectstest.ApaActivate(t, g, me, id, idx, game.ActivateAbilityParams{})
	effectstest.B27Kill(g, id)
	if !g.Battlefield.Contains(id) {
		t.Error("the shielded creature was destroyed")
	}
}

// --- printed keywords ----------------------------------------------------------

func TestPlcB1PrintedKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Blood Knight", "Creature — Human Knight", plcB1BloodKnightOracle, []string{plcB1FirstStrk, "protection from white"}},
		{"Needlepeak Spider", "Creature — Spider", plcB1NeedlepeakSpiderOracle, []string{"reach"}},
		{"Serra Sphinx", "Creature — Sphinx", plcB1SerraSphinxOracle, []string{"flying", "vigilance"}},
		{"Calciderm", "Creature — Beast", plcB1CalcidermOracle, []string{"shroud"}},
		{"Aven Riftwatcher", "Creature — Bird Rebel Soldier", plcB1AvenRiftwatcherOracle, []string{"flying"}},
		{"Malach of the Dawn", "Creature — Angel", plcB1MalachOfTheDawnOracle, []string{"flying"}},
		{"Mire Boa", "Creature — Snake", plcB1MireBoaOracle, []string{"swampwalk"}},
		{"Groundbreaker", "Creature — Elemental", plcB1GroundbreakerOracle, []string{"trample", "haste"}},
		{"Gossamer Phantasm", "Creature — Illusion", plcB1GossamerPhantasmOracle, []string{"flying"}},
		{"Giant Dustwasp", "Creature — Insect", plcB1GiantDustwaspOracle, []string{"flying"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			for _, kw := range tc.keywords {
				if !effectstest.HasKeywordOf(g, id, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// --- vanishing ---------------------------------------------------------------

func TestPlcB1VanishingEntersWithTimeCounters(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		n                      int
	}{
		{"Calciderm", "Creature — Beast", plcB1CalcidermOracle, 4},
		{"Waning Wurm", "Creature — Zombie Wurm", plcB1WaningWurmOracle, 2},
		{"Aven Riftwatcher", "Creature — Bird Rebel Soldier", plcB1AvenRiftwatcherOracle, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B38CastCreature(t, g, tc.name, tc.typeLine, tc.oracle, 2, 2)
			if got := plcB1BnsTimeCounters(g, id); got != tc.n {
				t.Errorf("%d time counters, want %d", got, tc.n)
			}
		})
	}
}

func TestPlcB1WaningWurmVanishesAfterTwoUpkeeps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	wurm := effectstest.B38CastCreature(t, g, "Waning Wurm", "Creature — Zombie Wurm", plcB1WaningWurmOracle, 7, 6)
	effectstest.Cmdr2B17NextUpkeepOfSeat0(t, g)
	if got := plcB1BnsTimeCounters(g, wurm); got != 1 {
		t.Fatalf("after an upkeep: %d time counters, want 1", got)
	}
	effectstest.Cmdr2B17NextUpkeepOfSeat0(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if _, here := effectstest.BattlefieldCard(g, wurm); here {
		t.Error("the Wurm survived the removal of its last time counter")
	}
}

func TestPlcB1AvenRiftwatcherGainsOnEnteringAndLeaving(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	aven := effectstest.B38CastCreature(t, g, "Aven Riftwatcher", "Creature — Bird Rebel Soldier", plcB1AvenRiftwatcherOracle, 2, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Fatalf("life %d after it entered, want %d", me.Life, life+2)
	}
	effectstest.B27Kill(g, aven)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("life %d after it left, want %d", me.Life, life+4)
	}
}

func TestPlcB1KeldonMaraudersPingsOnEnteringAndLeaving(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := opp.Life
	marauders := effectstest.B38CastCreature(t, g, "Keldon Marauders", "Creature — Human Warrior", plcB1KeldonMaraudersOracle, 3, 3)
	effectstest.B04WaitForPick(t, g, me.ID)
	plcB1Msc5AnswerPickPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Fatalf("the player is at %d after it entered, want %d", opp.Life, life-1)
	}
	effectstest.B27Kill(g, marauders)
	g.RunStateChecksForTest()
	effectstest.B04WaitForPick(t, g, me.ID)
	plcB1Msc5AnswerPickPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("the player is at %d after it left, want %d", opp.Life, life-2)
	}
}

// --- slivers -----------------------------------------------------------------

func TestPlcB1SliverKeywordGrantsReachEverySliverCreature(t *testing.T) {
	for _, tc := range []struct{ name, oracle, keyword string }{
		{"Battering Sliver", plcB1BatteringSliverOracle, "trample"},
		{"Reflex Sliver", plcB1ReflexSliverOracle, "haste"},
		{"Spitting Sliver", plcB1SpittingSliverOracle, plcB1FirstStrk},
		{"Synchronous Sliver", plcB1SynchronousSliverOracle, "vigilance"},
		{"Dormant Sliver", plcB1DormantSliverOracle, "defender"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			theirs := effectstest.TspB1Sliver(g, opp.ID, "Their Sliver", 1, 1)
			bear := effectstest.B12Creature(g, me.ID, "Bear", plcB1BearType, 2, 2)
			lord := effectstest.B12Push(g, me.ID, tc.name, plcB1Sliver, tc.oracle, 2, 2)
			if !effectstest.HasKeywordOf(g, theirs, tc.keyword) || !effectstest.HasKeywordOf(g, lord, tc.keyword) {
				t.Errorf("want every Sliver, an opponent's and this one included, to have %s", tc.keyword)
			}
			if effectstest.HasKeywordOf(g, bear, tc.keyword) {
				t.Errorf("a Bear has %s", tc.keyword)
			}
		})
	}
}

func TestPlcB1DarkheartSliverSacrificesASliverForThreeLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sliver := effectstest.TspB1Sliver(g, me.ID, "My Sliver", 1, 1)
	effectstest.B12Push(g, me.ID, "Darkheart Sliver", plcB1Sliver, plcB1DarkheartSliverOracle, 2, 2)
	effectstest.AdvanceToMain(t, g)
	life := me.Life
	effectstest.TspB1ActivateGranted(t, g, me.ID, sliver, nil)
	if g.Battlefield.Contains(sliver) {
		t.Error("the Sliver that activated the ability was not sacrificed")
	}
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

func TestPlcB1PoulticeSliverRegeneratesATargetSliver(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	healer := effectstest.TspB1Sliver(g, me.ID, "Healer Sliver", 1, 1)
	other := effectstest.TspB1Sliver(g, me.ID, "Other Sliver", 1, 1)
	effectstest.B12Push(g, me.ID, "Poultice Sliver", plcB1Sliver, plcB1PoulticeSliverOracle, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	effectstest.TspB1ActivateGranted(t, g, me.ID, healer, effectstest.CardRefs(other))
	if !effectstest.B16Tapped(t, g, healer) {
		t.Error("the Sliver that activated the ability did not tap")
	}
	effectstest.B27Kill(g, other)
	if !g.Battlefield.Contains(other) {
		t.Error("the shielded Sliver was destroyed")
	}
}

func TestPlcB1DormantSliverDrawsAsEachSliverEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	dormant := effectstest.CastHoldCreature(t, g, "Dormant Sliver", plcB1Sliver, plcB1DormantSliverOracle, 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dormant) {
		t.Fatal("Dormant Sliver did not resolve")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Fatalf("hand %d after Dormant Sliver entered, want %d", got, hand+1)
	}
	effectstest.CastHoldCreature(t, g, "Muscle Sliver", plcB1Sliver, "", 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d after another Sliver entered, want %d", got, hand+2)
	}
}

// --- activated abilities -------------------------------------------------------

func TestPlcB1ProdigalPyromancerPingsAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	pyro := effectstest.B12Push(g, me.ID, "Prodigal Pyromancer", "Creature — Human Wizard", plcB1ProdigalPyromancerOracle, 1, 1)
	life := opp.Life
	effectstest.ApaActivate(t, g, me, pyro, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	if opp.Life != life-1 {
		t.Errorf("the player is at %d, want %d", opp.Life, life-1)
	}
	if !effectstest.HmlTapped(t, g, pyro) {
		t.Error("the Pyromancer did not tap")
	}
}

func TestPlcB1RathiTrapperTapsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	trapper := effectstest.B12Push(g, me.ID, "Rathi Trapper", "Creature — Human Rebel Rogue", plcB1RathiTrapperOracle, 1, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", plcB1BearType, 2, 2)
	effectstest.ApaMana(me, "B")
	effectstest.ApaActivate(t, g, me, trapper, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	if !effectstest.HmlTapped(t, g, bear) {
		t.Error("the target is untapped")
	}
}

func TestPlcB1SaltfieldRecluseShrinksPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	recluse := effectstest.B12Push(g, me.ID, "Saltfield Recluse", "Creature — Human Rebel Cleric", plcB1SaltfieldRecluseOracle, 1, 2)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.ApaActivate(t, g, me, recluse, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(ogre)})
	if p, tough := effectstest.PtOf(t, g, ogre); p != 1 || tough != 3 {
		t.Errorf("the Ogre is %d/%d, want 1/3", p, tough)
	}
}

func TestPlcB1SerendibSorcererSetsAnotherCreatureToZeroTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	sorcerer := effectstest.B12Push(g, me.ID, "Serendib Sorcerer", "Creature — Human Wizard Sorcerer", plcB1SerendibSorcererOracle, 1, 1)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	if err := g.ActivateCatalogAbility(me.ID, sorcerer, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(sorcerer)}); err == nil {
		t.Fatal("the Sorcerer targeted itself")
	}
	effectstest.ApaActivate(t, g, me, sorcerer, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(ogre)})
	if p, tough := effectstest.PtOf(t, g, ogre); p != 0 || tough != 2 {
		t.Errorf("the Ogre is %d/%d, want 0/2", p, tough)
	}
}

func TestPlcB1MagusOfTheBazaarDrawsTwoThenDiscardsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	magus := effectstest.B12Push(g, me.ID, "Magus of the Bazaar", "Creature — Human Wizard", plcB1MagusOfTheBazaarOracle, 0, 1)
	effectstest.D2dSetHand(me, 3)
	effectstest.ApaActivate(t, g, me, magus, 0, game.ActivateAbilityParams{})
	if got := me.Hand.Size(); got != 5 {
		t.Errorf("hand is %d before the discard, want 5", got)
	}
	if n := effectstest.DiscardOwed(g, me.ID); n != 3 {
		t.Errorf("owes %d discards, want 3", n)
	}
}

func TestPlcB1RegenerateThisCreature(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		mana                   []string
	}{
		{"Revered Dead", "Creature — Spirit Soldier", plcB1ReveredDeadOracle, []string{"W"}},
		{"Malach of the Dawn", "Creature — Angel", plcB1MalachOfTheDawnOracle, []string{"W", "W", "W"}},
		{"Mire Boa", "Creature — Snake", plcB1MireBoaOracle, []string{"G"}},
		{"Hedge Troll", "Creature — Troll Cleric", plcB1HedgeTrollOracle, []string{"W"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := effectstest.Top100ActiveSeat(g)
			id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			plcB1Regenerates(t, g, me, id, 0, tc.mana...)
		})
	}
}

func TestPlcB1HedgeTrollGrowsWithAPlains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	troll := effectstest.B12Push(g, me.ID, "Hedge Troll", "Creature — Troll Cleric", plcB1HedgeTrollOracle, 2, 2)
	if p, tough := effectstest.PtOf(t, g, troll); p != 2 || tough != 2 {
		t.Fatalf("with no Plains the Troll is %d/%d, want 2/2", p, tough)
	}
	effectstest.B12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	if p, tough := effectstest.PtOf(t, g, troll); p != 3 || tough != 3 {
		t.Errorf("with a Plains the Troll is %d/%d, want 3/3", p, tough)
	}
}

func TestPlcB1GhostTacticianPumpsYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	tactician := effectstest.B12Push(g, me.ID, "Ghost Tactician", "Creature — Spirit Spellshaper", plcB1GhostTacticianOracle, 2, 5)
	mine := effectstest.B12Creature(g, me.ID, "Mine", plcB1BearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", plcB1BearType, 2, 2)
	pitched := effectstest.PushHandCard(g, me)
	effectstest.ApaMana(me, "W")
	effectstest.ApaActivate(t, g, me, tactician, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{pitched}})
	if !me.Graveyard.Contains(pitched) {
		t.Error("the discard was not paid")
	}
	if p, _ := effectstest.PtOf(t, g, mine); p != 3 {
		t.Errorf("your creature has power %d, want 3", p)
	}
	if p, _ := effectstest.PtOf(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature has power %d, want 2", p)
	}
}

func TestPlcB1RidgedKusitePumpsAndGrantsFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	kusite := effectstest.B12Push(g, me.ID, "Ridged Kusite", "Creature — Horror Spellshaper", plcB1RidgedKusiteOracle, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", plcB1BearType, 2, 2)
	pitched := effectstest.PushHandCard(g, me)
	effectstest.ApaMana(me, "C", "B")
	effectstest.ApaActivate(t, g, me, kusite, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear), DiscardIDs: []uuid.UUID{pitched}})
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("the Bear is %d/%d, want 3/2", p, tough)
	}
	if !effectstest.HasKeywordOf(g, bear, plcB1FirstStrk) {
		t.Error("the Bear has no first strike")
	}
}

func TestPlcB1SophicCentaurGainsTwoPerCardLeftInHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	centaur := effectstest.B12Push(g, me.ID, "Sophic Centaur", "Creature — Centaur Spellshaper", plcB1SophicCentaurOracle, 1, 1)
	effectstest.D2dSetHand(me, 3)
	pitched := me.Hand.Cards[0].InstanceID
	life := me.Life
	effectstest.ApaMana(me, "G", "G", "C", "C")
	effectstest.ApaActivate(t, g, me, centaur, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{pitched}})
	if me.Life != life+4 {
		t.Errorf("life %d, want %d (two cards left in hand)", me.Life, life+4)
	}
}

func TestPlcB1ShadeOfTrokairPumpsItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	shade := effectstest.B12Push(g, me.ID, "Shade of Trokair", "Creature — Shade", plcB1ShadeOfTrokairOracle, 1, 2)
	effectstest.ApaMana(me, "W")
	effectstest.ApaActivate(t, g, me, shade, 0, game.ActivateAbilityParams{})
	if p, tough := effectstest.PtOf(t, g, shade); p != 2 || tough != 3 {
		t.Errorf("the Shade is %d/%d, want 2/3", p, tough)
	}
}

// --- other creatures ---------------------------------------------------------

func TestPlcB1GroundbreakerIsSacrificedAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.B12Push(g, me.ID, "Groundbreaker", "Creature — Elemental", plcB1GroundbreakerOracle, 6, 1)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) || !me.Graveyard.Contains(id) {
		t.Fatal("Groundbreaker was not sacrificed at the end step")
	}
}

// Only an artifact or red creature can block Skirk Shaman.
func TestPlcB1SkirkShamanIsBlockedOnlyByArtifactOrRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shaman := effectstest.B12Push(g, me.ID, "Skirk Shaman", "Creature — Goblin Shaman", plcB1SkirkShamanOracle, 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", plcB1BearType, 2, 2)
	effectstest.HmlSetColors(g, bear, "G")
	golem := effectstest.B12Creature(g, opp.ID, "Golem", "Artifact Creature — Golem", 2, 2)
	goblin := effectstest.B12Creature(g, opp.ID, "Goblin", "Creature — Goblin", 1, 1)
	effectstest.HmlSetColors(g, goblin, "R")
	effectstest.BrAttack(t, g, shaman)
	if err := g.DeclareBlocker(bear, shaman); err == nil {
		t.Error("a green creature blocked Skirk Shaman")
	}
	if err := g.DeclareBlocker(golem, shaman); err != nil {
		t.Errorf("an artifact creature can't block Skirk Shaman: %v", err)
	}
	if err := g.DeclareBlocker(goblin, shaman); err != nil {
		t.Errorf("a red creature can't block Skirk Shaman: %v", err)
	}
}

func TestPlcB1GossamerPhantasmIsSacrificedWhenTargeted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	pyro := effectstest.B12Push(g, me.ID, "Prodigal Pyromancer", "Creature — Human Wizard", plcB1ProdigalPyromancerOracle, 1, 1)
	phantasm := effectstest.B12Push(g, opp.ID, "Gossamer Phantasm", "Creature — Illusion", plcB1GossamerPhantasmOracle, 2, 1)
	effectstest.ApaActivate(t, g, me, pyro, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(phantasm)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(phantasm) || !opp.Graveyard.Contains(phantasm) {
		t.Error("the Phantasm was not sacrificed when it became a target")
	}
}

func TestPlcB1CitanulWoodreadersDrawsOnlyWhenKicked(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		hand := me.Hand.Size()
		id := plcB1DmuCast(t, g, "Citanul Woodreaders", "Creature — Human Druid", plcB1CitanulWoodreadersOracle, kicked)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(id) {
			t.Fatalf("kicked=%v: the Woodreaders are not on the battlefield", kicked)
		}
		want := hand
		if kicked {
			want += 2
		}
		if got := me.Hand.Size(); got != want {
			t.Errorf("kicked=%v: hand %d, want %d", kicked, got, want)
		}
	}
}

// --- suspend -----------------------------------------------------------------

func TestPlcB1SuspendCardsDeclareTheirPrintedNumbers(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		n            int
		cost         string
	}{
		{"Giant Dustwasp", plcB1GiantDustwaspOracle, 4, "{1}{G}"},
		{"Heroes Remembered", plcB1HeroesRememberedOracle, 10, "{W}"},
		{"Shade of Trokair", plcB1ShadeOfTrokairOracle, 3, "{W}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			effectstest.S40SuspendAction(t, tc.oracle, tc.n, tc.cost)
		})
	}
}

func TestPlcB1HeroesRememberedGainsTwenty(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Heroes Remembered", plcB1Sorcery, plcB1HeroesRememberedOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+20 {
		t.Errorf("life %d, want %d", me.Life, life+20)
	}
}

func TestPlcB1GiantDustwaspArrivesFromSuspendWithHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := plcB1TspB1Suspend(t, g, "Giant Dustwasp", "Creature — Insect", plcB1GiantDustwaspOracle, "{C}{G}", 4)
	plcB1TspB1FreeCastAtUpkeep(t, g, id, game.CastSpellParams{})
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Dustwasp is not on the battlefield")
	}
	if !effectstest.HasKeywordOf(g, id, "haste") || !effectstest.HasKeywordOf(g, id, "flying") {
		t.Error("a creature cast from suspend has haste (CR 702.62a), and this one flies")
	}
}

// --- enchantments ------------------------------------------------------------

func TestPlcB1GaeasAnthemPumpsOnlyYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Gaea's Anthem", "Enchantment", plcB1GaeasAnthemOracle, false)
	mine := effectstest.B12Creature(g, me.ID, "Mine", plcB1BearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", plcB1BearType, 2, 2)
	if p, tough := effectstest.PtOf(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("your creature is %d/%d, want 3/3", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's creature is %d/%d, want 2/2", p, tough)
	}
}

func TestPlcB1DustCoronaPumpsAndKeepsFlyersOff(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", plcB1BearType, 2, 2)
	effectstest.Enchant(t, g, "Dust Corona", plcB1DustCoronaOracle, bear)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 2 {
		t.Fatalf("the enchanted Bear is %d/%d, want 4/2", p, tough)
	}
	flyer := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Flyer", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	walker := effectstest.B12Creature(g, opp.ID, "Walker", plcB1BearType, 2, 2)
	effectstest.BrAttack(t, g, bear)
	if err := g.DeclareBlocker(flyer, bear); err == nil {
		t.Error("a creature with flying blocked the enchanted creature")
	}
	if err := g.DeclareBlocker(walker, bear); err != nil {
		t.Errorf("a creature without flying can't block the enchanted creature: %v", err)
	}
}

func TestPlcB1VampiricLinkGainsTheDamageDealt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.SeedBear(g, me.ID)
	theirs := effectstest.SeedBear(g, opp.ID)
	effectstest.Enchant(t, g, "Vampiric Link", plcB1VampiricLinkOracle, mine)
	life := me.Life
	effectstest.EmitNoncombatDamage(g, mine, theirs, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Life - life; got != 2 {
		t.Errorf("gained %d life, want 2", got)
	}
}

func TestPlcB1UtopiaVowStopsBlocksAndGrantsMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", plcB1BearType, 2, 2)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", plcB1BearType, 2, 2)
	effectstest.Enchant(t, g, "Utopia Vow", plcB1UtopiaVowOracle, theirs)
	if n := effectstest.GrantedManaRowsOf(g, theirs); n != 1 {
		t.Errorf("the enchanted creature has %d granted mana rows, want 1", n)
	}
	if n := effectstest.GrantedManaRowsOf(g, attacker); n != 0 {
		t.Errorf("another creature has %d granted mana rows, want 0", n)
	}
	effectstest.BrAttack(t, g, attacker)
	if err := g.DeclareBlocker(theirs, attacker); err == nil {
		t.Error("the enchanted creature blocked")
	}
}

func TestPlcB1MelancholyHoldsTheCreatureUntilItIsNotPaidFor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", plcB1BearType, 2, 2)
	aura := effectstest.Enchant(t, g, "Melancholy", plcB1MelancholyOracle, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HmlTapped(t, g, theirs) {
		t.Fatal("the enchanted creature was not tapped")
	}
	effectstest.WalkToUpkeepOfSeat0(t, g)
	if !effectstest.HmlTapped(t, g, theirs) {
		t.Fatal("the enchanted creature untapped in its controller's untap step")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	if g.Battlefield.Contains(aura) || !me.Graveyard.Contains(aura) {
		t.Error("declining the {B} did not sacrifice Melancholy")
	}
}

func TestPlcB1EnslaveStealsAndHurtsTheOwnerEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", plcB1BearType, 2, 2)
	effectstest.Enchant(t, g, "Enslave", plcB1EnslaveOracle, theirs)
	if got := effectstest.ControllerOf(t, g, theirs); got != me.ID {
		t.Fatalf("controller %s, want the Aura's controller %s", got, me.ID)
	}
	effectstest.WalkToUpkeepOfSeat0(t, g)
	life, mine := opp.Life, me.Life
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("the owner is at %d, want %d", opp.Life, life-1)
	}
	if me.Life != mine {
		t.Errorf("the creature's controller lost life: %d, want %d", me.Life, mine)
	}
}

func TestPlcB1AuramancersGuiseCountsItsAuras(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", plcB1BearType, 2, 2)
	effectstest.Enchant(t, g, "Auramancer's Guise", plcB1AuramancersGuiseOracle, bear)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 4 {
		t.Fatalf("with one Aura the Bear is %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasKeywordOf(g, bear, "vigilance") {
		t.Error("the enchanted creature has no vigilance")
	}
	effectstest.Enchant(t, g, "Dust Corona", plcB1DustCoronaOracle, bear)
	if p, tough := effectstest.PtOf(t, g, bear); p != 8 || tough != 6 {
		t.Errorf("with two Auras the Bear is %d/%d, want 8/6", p, tough)
	}
}

// --- the spells --------------------------------------------------------------

func TestPlcB1ManaTitheCountersUnlessItsControllerPays(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp, victim := plcB1LegB2CounterSetup(t, g, plcB1Sorcery)
		effectstest.CastFromSeat(t, g, opp, "Mana Tithe", plcB1Instant, plcB1ManaTitheOracle,
			[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
		effectstest.PassUntilTaxed(t, g, me.ID)
		if pay {
			me.ManaPool.AddMana(game.ManaToken{Color: "C"})
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		countered := !g.Stack.Contains(victim) && me.Graveyard.Contains(victim)
		if countered == pay {
			t.Errorf("paid %v: countered %v", pay, countered)
		}
	}
}

func TestPlcB1DismalFailureCountersAndTheControllerDiscards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, victim := plcB1LegB2CounterSetup(t, g, plcB1Sorcery)
	effectstest.D2dSetHand(me, 2)
	effectstest.CastFromSeat(t, g, opp, "Dismal Failure", plcB1Instant, plcB1DismalFailureOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Stack.Contains(victim) || !me.Graveyard.Contains(victim) {
		t.Error("the spell was not countered")
	}
	if n := effectstest.DiscardOwed(g, me.ID); n != 1 {
		t.Errorf("the spell's controller owes %d discards, want 1", n)
	}
	if n := effectstest.DiscardOwed(g, opp.ID); n != 0 {
		t.Errorf("the caster owes %d discards, want 0", n)
	}
}

func TestPlcB1WistfulThinkingDrawsTwoThenDiscardsFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	effectstest.D2dSetHand(opp, 3)
	effectstest.CastCatalogSpell(t, g, "Wistful Thinking", plcB1Sorcery, plcB1WistfulThinkingOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size(); got != 5 {
		t.Errorf("the target's hand is %d before the discard, want 5", got)
	}
	if n := effectstest.DiscardOwed(g, opp.ID); n != 4 {
		t.Errorf("the target owes %d discards, want 4", n)
	}
}

func TestPlcB1SunlanceHitsOnlyANonwhiteCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	knight := effectstest.B12Creature(g, opp.ID, "Knight", "Creature — Human Knight", 2, 2)
	effectstest.HmlSetColors(g, knight, "W")
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	effectstest.HmlSetColors(g, ogre, "R")
	if err := effectstest.CastCatalogSpellErr(t, g, "Sunlance", plcB1Sorcery, plcB1SunlanceOracle, effectstest.CardTarget(knight)); err == nil {
		t.Fatal("Sunlance targeted a white creature")
	}
	effectstest.CastCatalogSpell(t, g, "Sunlance", plcB1Sorcery, plcB1SunlanceOracle, effectstest.CardTarget(ogre))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, ogre).DamageMarked; got != 3 {
		t.Errorf("the Ogre has %d damage, want 3", got)
	}
}

func TestPlcB1CradleToGraveDestroysOnlyANewNonblackCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	old := effectstest.B12Creature(g, opp.ID, "Old", plcB1BearType, 2, 2)
	zombie := effectstest.B12Creature(g, opp.ID, "Zombie", "Creature — Zombie", 2, 2)
	effectstest.HmlSetColors(g, zombie, "B")
	plcB1EnteredThisTurn(g, zombie)
	fresh := effectstest.B12Creature(g, opp.ID, "Fresh", plcB1BearType, 2, 2)
	effectstest.HmlSetColors(g, fresh, "G")
	plcB1EnteredThisTurn(g, fresh)
	if err := effectstest.CastCatalogSpellErr(t, g, "Cradle to Grave", plcB1Instant, plcB1CradleToGraveOracle, effectstest.CardTarget(old)); err == nil {
		t.Fatal("Cradle to Grave targeted a creature that did not enter this turn")
	}
	if err := effectstest.CastCatalogSpellErr(t, g, "Cradle to Grave", plcB1Instant, plcB1CradleToGraveOracle, effectstest.CardTarget(zombie)); err == nil {
		t.Fatal("Cradle to Grave targeted a black creature")
	}
	effectstest.CastCatalogSpell(t, g, "Cradle to Grave", plcB1Instant, plcB1CradleToGraveOracle, effectstest.CardTarget(fresh))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fresh) {
		t.Error("the creature that entered this turn survived")
	}
}

func TestPlcB1MidnightCharmEachMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	life := me.Life
	effectstest.CastModal(t, g, "Midnight Charm", plcB1Instant, plcB1MidnightCharmOracle, []int{0}, effectstest.CardTarget(ogre))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, ogre).DamageMarked; got != 1 {
		t.Errorf("the Ogre has %d damage, want 1", got)
	}
	if me.Life != life+1 {
		t.Errorf("life %d, want %d", me.Life, life+1)
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", plcB1BearType, 2, 2)
	effectstest.CastModal(t, g, "Midnight Charm", plcB1Instant, plcB1MidnightCharmOracle, []int{1}, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOf(g, bear, plcB1FirstStrk) {
		t.Error("the Bear has no first strike")
	}
	effectstest.CastModal(t, g, "Midnight Charm", plcB1Instant, plcB1MidnightCharmOracle, []int{2}, effectstest.CardTarget(ogre))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HmlTapped(t, g, ogre) {
		t.Error("the Ogre is untapped")
	}
}

func TestPlcB1PiracyCharmEachMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", plcB1BearType, 2, 2)
	effectstest.CastModal(t, g, "Piracy Charm", plcB1Instant, plcB1PiracyCharmOracle, []int{0}, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOf(g, bear, "islandwalk") {
		t.Error("the Bear has no islandwalk")
	}
	effectstest.CastModal(t, g, "Piracy Charm", plcB1Instant, plcB1PiracyCharmOracle, []int{1}, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 1 {
		t.Errorf("the Bear is %d/%d, want 4/1", p, tough)
	}
	effectstest.D2dSetHand(opp, 2)
	effectstest.CastModal(t, g, "Piracy Charm", plcB1Instant, plcB1PiracyCharmOracle, []int{2}, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.DiscardOwed(g, opp.ID); n != 1 {
		t.Errorf("the target owes %d discards, want 1", n)
	}
}

func TestPlcB1EvolutionCharmFetchesABasicLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		game.Card{Name: "Bear", TypeLine: plcB1BearType})
	effectstest.CastCatalogSpellWithModes(t, g, "Evolution Charm", plcB1Instant, plcB1EvolutionCharmOracle, []int{0})
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.AnswerSearchNamed(t, g, me.ID, "Forest")
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want +1", hand, me.Hand.Size())
	}
}

func TestPlcB1EvolutionCharmReturnsACreatureCardOrGrantsFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	dead := effectstest.PushGraveyardCardForTest(me, "Dead")
	effectstest.CastModal(t, g, "Evolution Charm", plcB1Instant, plcB1EvolutionCharmOracle, []int{1}, effectstest.CardTarget(dead))
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(dead) || me.Graveyard.Contains(dead) {
		t.Error("the creature card did not return to hand")
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", plcB1BearType, 2, 2)
	effectstest.CastModal(t, g, "Evolution Charm", plcB1Instant, plcB1EvolutionCharmOracle, []int{2}, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOf(g, bear, "flying") {
		t.Error("the Bear has no flying")
	}
}

func TestPlcB1HealingLeavesEachMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	life := me.Life
	effectstest.CastModal(t, g, "Healing Leaves", plcB1Instant, plcB1HealingLeavesOracle, []int{0}, []game.TargetRef{effectstest.PlayerTarget(me.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", plcB1BearType, 4, 4)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	effectstest.CastModal(t, g, "Healing Leaves", plcB1Instant, plcB1HealingLeavesOracle, []int{1}, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Pr6Damage(t, g, ogre, bear, 4)
	if got := effectstest.E2Card(t, g, bear).DamageMarked; got != 1 {
		t.Errorf("the shielded creature has %d damage marked, want 1", got)
	}
}

// Local copies of helpers that live in another test package.

func plcB1BnsTimeCounters(g *game.Game, id uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				n = c.Counters[game.CounterTime]
			}
		}
	})
	return n
}

// plcB1Msc5AnswerPickPlayer answers the open pick_target prompt with a player.
func plcB1Msc5AnswerPickPlayer(t *testing.T, g *game.Game, player uuid.UUID) {
	t.Helper()
	ch := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetPlayer, ID: player}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
}

// plcB1DmuCast casts a kicker creature from the active seat's hand with or without
// its kicker, settles the stack and returns the permanent's id.
func plcB1DmuCast(t *testing.T, g *game.Game, name, typeLine, oracle string, kicked bool) uuid.UUID {
	t.Helper()
	var opt []int
	if kicked {
		opt = []int{0}
	}
	id, err := effectstest.CastWithOptionalCosts(t, g, name, typeLine, oracle, nil, opt, nil)
	_ = id
	if err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// plcB1TspB1Suspend suspends a card from the active seat's hand with `mana`
// floating, then leaves it one time counter from its free cast.
func plcB1TspB1Suspend(t *testing.T, g *game.Game, name, typeLine, oracle, mana string, n int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	id := effectstest.HandCardForTest(me, name, typeLine, oracle)
	for i := range me.Hand.Cards {
		if me.Hand.Cards[i].InstanceID == id {
			me.Hand.Cards[i].Power, me.Hand.Cards[i].Toughness = 4, 4 // read only by a creature
		}
	}
	effectstest.FloatMana(t, g, me, mana)
	if err := g.PerformSpecialAction(me.ID, id, game.SpecialActionSuspend, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("suspend %s: %v", name, err)
	}
	if got := effectstest.CounterCountInExile(g, id, game.CounterTime); got != n {
		t.Fatalf("%s has %d time counters after suspending, want %d", name, got, n)
	}
	if n > 1 {
		g.WithWriteLock(func() {
			if err := g.AddCounterForEffect(id, game.CounterTime, -(n - 1)); err != nil {
				t.Fatalf("remove time counters: %v", err)
			}
		})
	}
	return id
}

// plcB1TspB1FreeCastAtUpkeep walks to the owner's next upkeep, lets the last
// time counter come off and casts the card from exile for free.
func plcB1TspB1FreeCastAtUpkeep(t *testing.T, g *game.Game, id uuid.UUID, params game.CastSpellParams) {
	t.Helper()
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.AdvanceToUpkeepOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKind(g, game.PendingChoiceResolutionCast) == nil {
		t.Fatal("the last time counter coming off offered no cast")
	}
	params.Strict, params.FromZone = true, "exile"
	if err := g.CastSpell(me.ID, id, params); err != nil {
		t.Fatalf("the free cast from suspend: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// plcB1LegB2CounterSetup casts the active seat's spell of `typeLine`, then
// passes priority to the next seat and returns it with the spell's id.
func plcB1LegB2CounterSetup(t *testing.T, g *game.Game, typeLine string) (*game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.ToMainForCost(t, g)
	victim := effectstest.CastCatalogSpell(t, g, "Their Spell", typeLine, "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority to the responder: %v", err)
	}
	return me, opp, victim
}
