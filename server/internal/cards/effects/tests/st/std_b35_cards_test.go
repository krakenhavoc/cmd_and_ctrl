package testsst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsbu "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/bu"
	cardsst "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/st"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b35_cards_test.go — Standard-set leftovers, batches 35 to 38 (#4452).
// Each test plays the real catalog entry through the engine.

const (
	b35FormAPosseOracle          = "8923462f-0540-41e8-931e-97b5f1dadd51"
	b35HarvesterOfMiseryOracle   = "a0d3bb63-2977-41b2-bcc3-2488e6696cb8"
	b35SandstormSalvagerOracle   = "ec19a9bc-6fc1-4150-976b-37a47b165016"
	b35GreedsGambitOracle        = "0868ebbb-2fe9-4f53-9697-acad8cda5ec1"
	b35HostileInvestigatorOracle = "de83b1a4-0ce0-4cc1-aff3-6d0239d92302"
	b35HighNoonOracle            = "3c8cbf0f-0288-4b5e-bd4f-0f208670b6b5"
	b35LegionExtruderOracle      = "c77dc11e-d95d-4e7e-bb24-d971c179e9b0"
	b35AncientCornucopiaOracle   = "f0950c95-a297-489d-b166-74bba6af5376"
	b35LotusRingOracle           = "ca4d3d27-ea3a-44ac-96e9-b52b563129f8"
	b35SwordWealthPowerOracle    = "b3a80545-9a8a-46b2-9d8c-b0d66fb68a5d"
	b35OltecMatterweaverOracle   = "3cb91696-242f-4e91-806b-61b1bb45a20c"
	b35WorldwalkerHelmOracle     = "9573c85a-e574-4b4e-aae0-2165c45fd27e"
	b35CollectorsCageOracle      = "8f5b71fd-469f-4370-a806-9bfefbae03a2"
	b35OmenpathJourneyOracle     = "8b067e44-2001-4b49-aa4a-2c86f8929751"
	b35ErthaJoOracle             = "0a66ce8b-af99-411f-8ecb-52a5d2f6af3d"
	b35JoleneOracle              = "fb8efd6e-b8f8-4da8-ac65-5b745c02d2bb"
	b35BucolicRanchOracle        = "8f5902bf-4bc4-4d0c-84ea-a425307a4eb2"
	b35FrontierSeekerOracle      = "466cf331-2d9c-4ebc-84f6-30c890e1b637"
	b35KellanJoinsUpOracle       = "9b3ddd31-c058-4846-8784-25c7a9bbcb31"
	b35OkoRingleaderOracle       = "4372b729-dc3d-4469-8fc5-811cc787d653"
	b35JaceReawakenedOracle      = "cbd483b5-5554-43ed-a729-535d01b0d5d3"
	b35StepBetweenWorldsOracle   = "48b43d53-9a8a-47a2-8610-a54d276d7be7"
	b35LilahOracle               = "0757c5a7-e51b-4c39-bb4d-b32657c10cb4"
	b35TinybonesOracle           = "7bc4c7e2-6758-4a85-84e7-03ab93981106"
	b35VraskaSilencerOracle      = "e7c5a568-3855-4a20-9edf-01a8a6a4ba97"
	b35RoxanneOracle             = "11765eaf-87a1-48bd-90c1-7a5f21be5918"
	b35DocAurlockOracle          = "3ca32660-bf5f-4398-bef6-a0ec8bc767cf"
	b35BonnyPallOracle           = "fcb811a8-c6e6-4bc4-b64b-8963a7183669"
	b35KambalOracle              = "619e0686-e88c-4238-8364-75395e733533"
	b35DuplicatorOracle          = "0752fe67-9b8a-42c0-90e3-087d426b2f3b"
	b35LootOracle                = "542be13f-b9c0-4b79-b486-be576c531a82"
	b35JasperFlintOracle         = "bd7c5a8b-cf62-41e1-a664-b16e3e5d142b"
	b35GhiredOracle              = "34b36022-0d93-48cc-a2cc-90ca4d65dc31"
	b35KellanKidOracle           = "afacc200-16ff-40f3-9d43-268438131d24"
	b35ErietteOracle             = "710a3499-64f8-4d2c-8c9c-d58950e47ea7"
	b35RakdosOracle              = "e2c141af-925e-42bd-8d83-e1a71f4c9794"
	b35IntimidationCampaignOra   = "5915039b-3dbd-4a49-8bd3-23ff0fd9c90a"
)

func TestB35EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b35FormAPosseOracle:          "Form a Posse",
		b35HarvesterOfMiseryOracle:   "Harvester of Misery",
		b35SandstormSalvagerOracle:   "Sandstorm Salvager",
		b35GreedsGambitOracle:        "Greed's Gambit",
		b35HostileInvestigatorOracle: "Hostile Investigator",
		b35HighNoonOracle:            "High Noon",
		b35LegionExtruderOracle:      "Legion Extruder",
		b35AncientCornucopiaOracle:   "Ancient Cornucopia",
		b35LotusRingOracle:           "Lotus Ring",
		b35SwordWealthPowerOracle:    "Sword of Wealth and Power",
		b35OltecMatterweaverOracle:   "Oltec Matterweaver",
		b35WorldwalkerHelmOracle:     "Worldwalker Helm",
		b35CollectorsCageOracle:      "Collector's Cage",
		b35OmenpathJourneyOracle:     "Omenpath Journey",
		b35ErthaJoOracle:             "Ertha Jo, Frontier Mentor",
		b35JoleneOracle:              "Jolene, Plundering Pugilist",
		b35BucolicRanchOracle:        "Bucolic Ranch",
		b35FrontierSeekerOracle:      "Frontier Seeker",
		b35KellanJoinsUpOracle:       "Kellan Joins Up",
		b35OkoRingleaderOracle:       "Oko, the Ringleader",
		b35JaceReawakenedOracle:      "Jace Reawakened",
		b35StepBetweenWorldsOracle:   "Step Between Worlds",
		b35LilahOracle:               "Lilah, Undefeated Slickshot",
		b35TinybonesOracle:           "Tinybones, the Pickpocket",
		b35VraskaSilencerOracle:      "Vraska, the Silencer",
		b35RoxanneOracle:             "Roxanne, Starfall Savant",
		b35DocAurlockOracle:          "Doc Aurlock, Grizzled Genius",
		b35BonnyPallOracle:           "Bonny Pall, Clearcutter",
		b35KambalOracle:              "Kambal, Profiteering Mayor",
		b35DuplicatorOracle:          "Esoteric Duplicator",
		b35LootOracle:                "Loot, the Key to Everything",
		b35JasperFlintOracle:         "Laughing Jasper Flint",
		b35GhiredOracle:              "Ghired, Mirror of the Wilds",
		b35KellanKidOracle:           "Kellan, the Kid",
		b35ErietteOracle:             "Eriette, the Beguiler",
		b35RakdosOracle:              "Rakdos, the Muscle",
		b35IntimidationCampaignOra:   "Intimidation Campaign",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestB35FormAPosseMakesXMercenaries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12PlayFromHand(t, g, "Form a Posse", "Sorcery", b35FormAPosseOracle, game.CastSpellParams{XValue: 3})
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.OtjMPMercenaries(g, me.ID)); n != 3 {
		t.Fatalf("X=3 made %d Mercenaries, want 3", n)
	}
}

func TestB35HarvesterOfMiseryShrinksOtherCreaturesWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	harvester := effectstest.Msc17Play(t, g, me, game.Card{Name: "Harvester of Misery", TypeLine: "Creature — Spirit",
		ManaCost: "{3}{B}{B}", OracleID: b35HarvesterOfMiseryOracle, Power: 5, Toughness: 4})
	for _, id := range []uuid.UUID{mine, theirs} {
		if p, tt := effectstest.PtOf(t, g, id); p != 1 || tt != 1 {
			t.Errorf("a Bear is %d/%d, want 1/1", p, tt)
		}
	}
	if p, tt := effectstest.PtOf(t, g, harvester); p != 5 || tt != 4 {
		t.Errorf("the Harvester is %d/%d, want 5/4 (other creatures only)", p, tt)
	}
}

func TestB35HarvesterOfMiseryDiscardAbilityShrinksATarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	effectstest.AdvanceToMain(t, g)
	hid := effectstest.PutInHand(me, game.Card{Name: "Harvester of Misery", TypeLine: "Creature — Spirit",
		ManaCost: "{3}{B}{B}", OracleID: b35HarvesterOfMiseryOracle, Power: 5, Toughness: 4, Owner: me.ID, Controller: me.ID})
	effectstest.B06AddMana(me, "B", "C")
	effectstest.B16Activate(t, g, me.ID, hid, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(theirs)})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, theirs); p != 1 || tt != 1 {
		t.Errorf("the target is %d/%d, want 1/1", p, tt)
	}
	if effectstest.HandHasID(me, hid) {
		t.Error("the Harvester was not discarded to pay")
	}
}

func TestB35SandstormSalvagerMakesAGolem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Sandstorm Salvager", TypeLine: "Creature — Human Artificer",
		ManaCost: "{2}{G}", OracleID: b35SandstormSalvagerOracle, Power: 3, Toughness: 3})
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Golem" && c.IsToken() && c.IsCreature() && c.IsArtifact() && len(c.Colors) == 0 {
			n++
			if p, tt := effectstest.PtOf(t, g, c.InstanceID); p != 3 || tt != 3 {
				t.Errorf("the Golem is %d/%d, want 3/3", p, tt)
			}
		}
	}
	if n != 1 {
		t.Fatalf("%d colorless Golem artifact creature tokens, want 1", n)
	}
}

func TestB35SandstormSalvagerGrowsTokensAndGivesTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	salvager := effectstest.B12Push(g, me.ID, "Sandstorm Salvager", "Creature — Human Artificer", b35SandstormSalvagerOracle, 3, 3)
	merc := effectstest.PushToken(g, me.ID, effects.MercenaryToken())
	nontoken := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.B06AddMana(me, "C", "C")
	effectstest.B16Activate(t, g, me.ID, salvager, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, merc); p != 2 || tt != 2 {
		t.Errorf("the Mercenary is %d/%d, want 2/2", p, tt)
	}
	if !effectstest.HasEffectiveKeyword(t, g, merc, "trample") {
		t.Error("the token gains trample")
	}
	if p, tt := effectstest.PtOf(t, g, nontoken); p != 2 || tt != 2 || effectstest.HasEffectiveKeyword(t, g, nontoken, "trample") {
		t.Errorf("a nontoken creature changed: %d/%d", p, tt)
	}
}

func TestB35GreedsGambitEntersDrawsGainsAndMakesBats(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life := effectstest.LifeOf(g, me.ID)
	hand := len(me.Hand.Cards)
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Greed's Gambit", TypeLine: "Enchantment",
		ManaCost: "{3}{B}", OracleID: b35GreedsGambitOracle})
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Bat"); n != 3 {
		t.Fatalf("%d Bats, want 3", n)
	}
	if got := effectstest.LifeOf(g, me.ID); got != life+6 {
		t.Errorf("life %d, want %d", got, life+6)
	}
	if got := len(me.Hand.Cards); got != hand+3 {
		t.Errorf("hand %d, want %d", got, hand+3)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Bat" {
			if p, tt := effectstest.PtOf(t, g, c.InstanceID); p != 2 || tt != 1 {
				t.Errorf("a Bat is %d/%d, want 2/1", p, tt)
			}
		}
	}
}

func TestB35HostileInvestigatorInvestigatesOncePerTurnOnDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	_, err := effectstest.Msc17Cast(t, g, me, game.Card{Name: "Hostile Investigator", TypeLine: "Creature — Ogre Rogue Detective",
		ManaCost: "{3}{B}", OracleID: b35HostileInvestigatorOracle, Power: 4, Toughness: 4}, game.CastSpellParams{})
	if err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, opp.ID, opp.Hand.Cards[0].InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Clue"); n != 1 {
		t.Fatalf("%d Clues after the first discard, want 1", n)
	}
	// A second discard the same turn investigates no more.
	effectstest.B30DiscardOne(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Clue"); n != 1 {
		t.Errorf("%d Clues after the second discard, want still 1", n)
	}
}

func TestB35HighNoonLimitsEachPlayerToOneSpellAndSacrificesForFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	noon := effectstest.B12Push(g, me.ID, "High Noon", "Enchantment", b35HighNoonOracle, 0, 0)
	if err := g.CastSpell(me.ID, effectstest.PutInHand(me, game.Card{Name: "Shock", TypeLine: "Instant", OracleID: effectstest.LightningBoltOracle, Owner: me.ID, Controller: me.ID}),
		game.CastSpellParams{Targets: effectstest.B16TargetPlayer(opp.ID)}); err != nil {
		t.Fatalf("the first spell: %v", err)
	}
	second := effectstest.PutInHand(me, game.Card{Name: "Shock", TypeLine: "Instant", OracleID: effectstest.LightningBoltOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{Targets: effectstest.B16TargetPlayer(opp.ID), Strict: true}); err == nil {
		t.Fatal("a second spell this turn was cast under High Noon")
	}
	effectstest.PassPriorityAroundTable(t, g)
	life := effectstest.LifeOf(g, opp.ID)
	effectstest.B06AddMana(me, "R", "C", "C", "C", "C")
	effectstest.B16Activate(t, g, me.ID, noon, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetPlayer(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != life-5 {
		t.Errorf("life %d, want %d", got, life-5)
	}
	if g.Battlefield.Contains(noon) {
		t.Error("High Noon was not sacrificed")
	}
}

func TestB35LegionExtruderShocksOnEntryAndSacrificesAnArtifactForAGolem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := effectstest.LifeOf(g, opp.ID)
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Legion Extruder", TypeLine: "Artifact", ManaCost: "{1}{R}", OracleID: b35LegionExtruderOracle})
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != life-2 {
		t.Errorf("life %d, want %d", got, life-2)
	}
}

func TestB35LegionExtruderMakesAGolemForAnotherArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ext := effectstest.B12Push(g, me.ID, "Legion Extruder", "Artifact", b35LegionExtruderOracle, 0, 0)
	other := effectstest.B12Permanent(g, me.ID, "Trinket", "Artifact")
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, ext, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{ext}}); err == nil {
		t.Fatal("Legion Extruder sacrificed itself")
	}
	effectstest.B16Activate(t, g, me.ID, ext, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{other}})
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Golem"); n != 1 {
		t.Fatalf("%d Golems, want 1", n)
	}
}

func TestB35GreedsGambitEndStepDiscardsLosesAndSacrifices(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Greed's Gambit", "Enchantment", b35GreedsGambitOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	life, hand := effectstest.LifeOf(g, me.ID), len(me.Hand.Cards)
	for g.Turn.Step != game.StepEnd {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.AnswerSacrifice(t, g, me.ID, bear)
	if got := effectstest.LifeOf(g, me.ID); got != life-2 {
		t.Errorf("life %d, want %d", got, life-2)
	}
	if got := len(me.Hand.Cards); got != hand-1 {
		t.Errorf("hand %d, want %d", got, hand-1)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the creature was not sacrificed")
	}
}

func TestB35AncientCornucopiaGainsLifePerColorOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Ancient Cornucopia", "Artifact", b35AncientCornucopiaOracle, 0, 0)
	life := effectstest.LifeOf(g, me.ID)
	cast := func(colors []string) {
		id := effectstest.PutInHand(me, game.Card{Name: "Pretty Spell", TypeLine: "Instant", OracleID: effectstest.LightningBoltOracle, Colors: colors, Owner: me.ID, Controller: me.ID})
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.B16TargetPlayer(opp.ID)}); err != nil {
			t.Fatal(err)
		}
	}
	cast([]string{"R", "G"})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != life+2 {
		t.Fatalf("life %d, want %d (two colors)", got, life+2)
	}
	cast([]string{"R"})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != life+2 {
		t.Errorf("life %d, want still %d: only once each turn", got, life+2)
	}
}

func TestB35LotusRingEquipsForThreeThreeVigilanceAndAThreeManaSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.GaPut(g, me.ID, "Bear", "Creature — Bear", "")
	ring := effectstest.GaPut(g, me.ID, "Lotus Ring", "Artifact — Equipment", b35LotusRingOracle)
	if !effectstest.HasEffectiveKeyword(t, g, ring, "indestructible") {
		t.Error("Lotus Ring is indestructible")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	effectstest.EquipTo(t, g, me.ID, ring, bear)
	if p, tt := effectstest.PtOf(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("the equipped creature is %d/%d, want 3/3", p, tt)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Error("the equipped creature has vigilance")
	}
	if effectstest.GrantedManaRowsOf(g, bear) != 1 {
		t.Fatal("the equipped creature has the three-mana ability")
	}
}

func TestB35SwordOfWealthAndPowerMakesATreasureAndCopiesTheNextInstant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	sword := effectstest.GaPut(g, me.ID, "Sword of Wealth and Power", "Artifact — Equipment", b35SwordWealthPowerOracle)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	effectstest.EquipTo(t, g, me.ID, sword, bear)
	if p, tt := effectstest.PtOf(t, g, bear); p != 4 || tt != 4 {
		t.Errorf("the equipped creature is %d/%d, want 4/4", p, tt)
	}
	for _, kw := range []string{game.ProtectionFromCardType("Instant"), game.ProtectionFromCardType("Sorcery")} {
		if !effectstest.HasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the equipped creature lacks %q", kw)
		}
	}
	if effectstest.CountTreasures(g, me.ID) != 0 {
		t.Fatal("a Treasure before any damage")
	}
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: bear, Target: opp.ID, Actor: me.ID, Amount: 3, Combat: true})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountTreasures(g, me.ID); got != 1 {
		t.Fatalf("%d Treasures after the combat damage, want 1", got)
	}
	life := effectstest.LifeOf(g, opp.ID)
	effectstest.CastBoltAtAndKeepTheCopysTarget(t, g, me.ID, opp.ID, 1)
	if got := effectstest.LifeOf(g, opp.ID); got != life-6 {
		t.Errorf("life %d after a copied Bolt, want %d", got, life-6)
	}
	effectstest.CastBoltAtAndKeepTheCopysTarget(t, g, me.ID, opp.ID, 0)
	if got := effectstest.LifeOf(g, opp.ID); got != life-9 {
		t.Errorf("life %d after a second Bolt, want %d (only the next spell is copied)", got, life-9)
	}
}

func TestB35OltecMatterweaverMakesAGnomeOrCopiesAnArtifactToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Oltec Matterweaver", "Creature — Human Artificer", b35OltecMatterweaverOracle, 3, 2)
	effectstest.B12PlayFromHand(t, g, "Bear Cub", "Creature — Bear", "", game.CastSpellParams{})
	effectstest.AnswerMode(t, g, effectstest.ModePickChoiceFor(g, me.ID), me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Gnome"); n != 1 {
		t.Fatalf("%d Gnomes, want 1", n)
	}
	effectstest.B12PlayFromHand(t, g, "Bear Cub", "Creature — Bear", "", game.CastSpellParams{})
	effectstest.AnswerMode(t, g, effectstest.ModePickChoiceFor(g, me.ID), me.ID, 1)
	effectstest.PickCard(t, g, me.ID, func() uuid.UUID {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Gnome" {
				return c.InstanceID
			}
		}
		return uuid.Nil
	}())
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Gnome"); n != 2 {
		t.Fatalf("%d Gnomes after the copy mode, want 2", n)
	}
}

func TestB35WorldwalkerHelmAddsAMapToArtifactTokensOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.GaPut(g, me.ID, "Worldwalker Helm", "Artifact", b35WorldwalkerHelmOracle)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TreasureToken(), 2) })
	if got := effectstest.CountTreasures(g, me.ID); got != 2 {
		t.Errorf("%d Treasures, want 2", got)
	}
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Map"); n != 1 {
		t.Errorf("%d Maps for one instruction, want 1", n)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.RedGoblinToken(), 1) })
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Map"); n != 1 {
		t.Errorf("a Goblin-only creation changed the Maps to %d", n)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.TreasureToken(), 1) })
	if n, _ := effectstest.CountTokensNamed(g, opp.ID, "Map"); n != 0 {
		t.Errorf("an opponent's artifact token got %d Maps", n)
	}
}

func TestB35WorldwalkerHelmCopiesAnArtifactToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	helm := effectstest.GaPut(g, me.ID, "Worldwalker Helm", "Artifact", b35WorldwalkerHelmOracle)
	clue := effectstest.PushToken(g, me.ID, effects.ClueToken())
	effectstest.B06AddMana(me, "U", "C")
	effectstest.B16Activate(t, g, me.ID, helm, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(clue)})
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Clue"); n != 2 {
		t.Fatalf("%d Clues, want 2", n)
	}
}

func TestB35CollectorsCagePlaysTheExiledCardOnlyWithThreeDifferentPowers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	top := effectstest.StackTopLibrary(me, "A", "B", "C", "D", "E")
	cage := game.NewCard("Collector's Cage", me.ID)
	cage.OracleID = b35CollectorsCageOracle
	cage.TypeLine = "Artifact"
	effectstest.EnterWithHideaway(t, g, cage, top[1])
	two := effectstest.B12Creature(g, me.ID, "Two", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Another Two", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Four", "Creature — Bear", 4, 4)
	activate := func(target uuid.UUID) {
		t.Helper()
		effectstest.UntapForTest(g, cage.InstanceID)
		effectstest.FloatForTest(g, me, "C")
		if err := g.ActivateCatalogAbility(me.ID, cage.InstanceID, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(target)}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	// 2, 2, 4 with a counter on a Two is 3, 2, 4: three different powers,
	// and it is the counter that makes the third.
	activate(two)
	if p, _ := effectstest.PtOf(t, g, two); p != 3 {
		t.Fatalf("the counter did not land: power %d", p)
	}
	if effectstest.PlayHiddenGrant(g, top[1]) == nil {
		t.Fatal("no grant with powers 3, 2, 4")
	}
}

func TestB35CollectorsCageGrantsNothingWithTwoDifferentPowers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	top := effectstest.StackTopLibrary(me, "A", "B", "C", "D", "E")
	cage := game.NewCard("Collector's Cage", me.ID)
	cage.OracleID = b35CollectorsCageOracle
	cage.TypeLine = "Artifact"
	effectstest.EnterWithHideaway(t, g, cage, top[1])
	four := effectstest.B12Creature(g, me.ID, "Four", "Creature — Bear", 4, 4)
	effectstest.B12Creature(g, me.ID, "Two", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Another Two", "Creature — Bear", 2, 2)
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, cage.InstanceID, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(four)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlayHiddenGrant(g, top[1]) != nil {
		t.Fatal("granted with powers 5, 2, 2")
	}
}

func TestB35OmenpathJourneyExilesDifferentlyNamedLandsAndPutsOneOutEachEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Plains", "Basic Land — Plains"),
		effectstest.SearchTestLand("Plains", "Basic Land — Plains"),
		effectstest.SearchTestLand("Island", "Basic Land — Island"),
		effectstest.SearchTestLand("Swamp", "Basic Land — Swamp"),
		effectstest.SearchTestLand("Mountain", "Basic Land — Mountain"),
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Wastes", "Basic Land"),
	)
	saga := effectstest.CastCatalogSpell(t, g, "Omenpath Journey", "Enchantment", b35OmenpathJourneyOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	// Two cards with the same name are refused as a set.
	var plains []uuid.UUID
	for _, id := range c.SearchCards {
		g.ReadSnapshot(func() {
			if card, ok := g.LookupCardForEffect(id); ok && card.Name == "Plains" {
				plains = append(plains, id)
			}
		})
	}
	if len(plains) != 2 {
		t.Fatalf("%d Plains offered, want 2", len(plains))
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, plains); err == nil {
		t.Fatal("two lands with the same name were accepted")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Plains", "Island", "Swamp", "Mountain", "Forest")
	effectstest.PassPriorityAroundTable(t, g)
	ref := game.PermissionCardRef(effectstest.HobA9SagaRef(t, g, saga))
	var linked int
	g.ReadSnapshot(func() { linked = len(g.ExiledWithForEffect(ref)) })
	if linked != 5 {
		t.Fatalf("%d cards exiled with Omenpath Journey, want 5", linked)
	}
	lands := func() int {
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.Controller == me.ID && c.IsLand() {
				n++
			}
		}
		return n
	}
	before := lands()
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := lands(); got != before+1 {
		t.Fatalf("lands %d, want %d after the end step", got, before+1)
	}
	g.ReadSnapshot(func() { linked = len(g.ExiledWithForEffect(ref)) })
	if linked != 4 {
		t.Errorf("%d cards still exiled with it, want 4", linked)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsLand() && !c.Tapped {
			t.Errorf("%s came in untapped", c.Name)
		}
	}
}

func TestB35ErthaJoMakesAMercenaryAndCopiesAbilitiesThatTargetACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Ertha Jo, Frontier Mentor", TypeLine: "Legendary Creature — Kor Advisor",
		ManaCost: "{2}{R}{W}", OracleID: b35ErthaJoOracle, Power: 2, Toughness: 3})
	effectstest.OtjMPOneMercenary(t, g, me.ID)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	merc := effectstest.PushToken(g, me.ID, effects.MercenaryToken())
	effectstest.B16Activate(t, g, me.ID, merc, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(mine)})
	for i := 0; i < 32; i++ {
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if effectstest.StackFullyEmpty(g) {
			break
		}
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if p, tt := effectstest.PtOf(t, g, mine); p != 4 || tt != 2 {
		t.Errorf("the Bear is %d/%d, want 4/2: the pump and its copy", p, tt)
	}
}

func TestB35JoleneMakesOneTreasureForABigAttackAndPingsForATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Jolene, Plundering Pugilist", "Legendary Creature — Human Mercenary", b35JoleneOracle, 3, 3)
	small := effectstest.PushVanillaCreature(g, me.ID, "Small", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(small, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountTreasures(g, me.ID); got != 0 {
		t.Fatalf("%d Treasures for a power-2 attacker, want 0", got)
	}
}

func TestB35JoleneBigAttackersMakeOneTreasureTotal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Jolene, Plundering Pugilist", "Legendary Creature — Human Mercenary", b35JoleneOracle, 3, 3)
	big1 := effectstest.PushVanillaCreature(g, me.ID, "Big One", 4, 4)
	big2 := effectstest.PushVanillaCreature(g, me.ID, "Big Two", 5, 5)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{big1, big2} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountTreasures(g, me.ID); got != 1 {
		t.Fatalf("%d Treasures for two big attackers, want 1", got)
	}
}

func TestB35JoleneSacrificesATreasureToPing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	jolene := effectstest.B12Push(g, me.ID, "Jolene, Plundering Pugilist", "Legendary Creature — Human Mercenary", b35JoleneOracle, 3, 3)
	treasure := effectstest.PushToken(g, me.ID, effects.TreasureToken())
	life := effectstest.LifeOf(g, opp.ID)
	effectstest.B06AddMana(me, "R", "C")
	effectstest.B16Activate(t, g, me.ID, jolene, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{treasure}, Targets: effectstest.B16TargetPlayer(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != life-1 {
		t.Errorf("life %d, want %d", got, life-1)
	}
	if g.Battlefield.Contains(treasure) {
		t.Error("the Treasure was not sacrificed")
	}
}

func TestB35IntimidationCampaignDrainsAndMayReturnOnACrime(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life, oppLife, hand := effectstest.LifeOf(g, me.ID), effectstest.LifeOf(g, opp.ID), len(me.Hand.Cards)
	camp := effectstest.Msc17Play(t, g, me, game.Card{Name: "Intimidation Campaign", TypeLine: "Enchantment",
		ManaCost: "{1}{U}{B}", OracleID: b35IntimidationCampaignOra})
	if effectstest.LifeOf(g, me.ID) != life+1 || effectstest.LifeOf(g, opp.ID) != oppLife-1 {
		t.Errorf("drain: you %d (from %d), them %d (from %d)", effectstest.LifeOf(g, me.ID), life, effectstest.LifeOf(g, opp.ID), oppLife)
	}
	if len(me.Hand.Cards) != hand+1 { // the card is added then cast (net 0), then the draw
		t.Errorf("hand %d, want %d", len(me.Hand.Cards), hand+1)
	}
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBecomesTarget, CrimeCommitted: true, Actor: me.ID})
	})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(camp) {
		t.Error("the enchantment did not return to hand")
	}
	if !effectstest.HandHasID(me, camp) {
		t.Error("the enchantment is not in its owner's hand")
	}
}

func TestB35FrontierSeekerTakesAMountCreatureOrAPlainsFromTheTopFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	top := effectstest.HornTopFive(me, []struct{ Name, TypeLine string }{
		{"Mountain", "Basic Land — Mountain"},
		{"Rodeo Rider", "Creature — Human Scout"},
		{"Prairie Mount", "Creature — Mount"},
		{"Plains", "Basic Land — Plains"},
		{"Mount Aura", "Enchantment — Mount"},
	})
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Frontier Seeker", TypeLine: "Creature — Human Scout",
		ManaCost: "{1}{W}", OracleID: b35FrontierSeekerOracle, Power: 2, Toughness: 2})
	prompt := effectstest.LatestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("no prompt")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range prompt.ChooseCards {
		offered[id] = true
	}
	if len(offered) != 2 || !offered[top[2]] || !offered[top[3]] {
		t.Fatalf("offered %v, want only the Mount creature and the Plains", prompt.ChooseCards)
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{top[3]}); err != nil {
		t.Fatal(err)
	}
	if !me.Hand.Contains(top[3]) {
		t.Error("the Plains is not in hand")
	}
}

func b35BucolicRanchActivate(t *testing.T, g *game.Game, me *game.Player, topCards ...struct{ Name, TypeLine string }) []uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	ranch := effectstest.B12Push(g, me.ID, "Bucolic Ranch", "Land — Desert", b35BucolicRanchOracle, 0, 0)
	top := effectstest.HornTopFive(me, topCards)
	effectstest.B06AddMana(me, "C", "C", "C")
	effectstest.B16Activate(t, g, me.ID, ranch, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	return top
}

func TestB35BucolicRanchMayTakeAMountCardToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	top := b35BucolicRanchActivate(t, g, me, struct{ Name, TypeLine string }{"Prairie Mount", "Creature — Mount"})
	effectstest.B35AnswerOption(t, g, me.ID, cardsbu.BucolicRanchTake)
	if !me.Hand.Contains(top[0]) {
		t.Error("the Mount card is not in hand")
	}
}

func TestB35BucolicRanchOffersNoTakeForANonMountAndMayBottomIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	top := b35BucolicRanchActivate(t, g, me, struct{ Name, TypeLine string }{"Bear", "Creature — Bear"})
	c := effectstest.LatestOptionPickFor(g, me.ID)
	if c == nil {
		t.Fatal("no option prompt")
	}
	for _, opt := range c.PickOptions {
		if opt.Label == cardsbu.BucolicRanchTake {
			t.Fatal("a non-Mount card was offered for your hand")
		}
	}
	indexOf := func() int {
		for i, c := range me.Library.Cards {
			if c.InstanceID == top[0] {
				return i
			}
		}
		return -1
	}
	topIdx := indexOf()
	effectstest.B35AnswerOption(t, g, me.ID, cardsbu.BucolicRanchBottom)
	if me.Hand.Contains(top[0]) {
		t.Fatal("the card went to hand")
	}
	if got, want := indexOf(), len(me.Library.Cards)-1-topIdx; got != want {
		t.Errorf("the card is at index %d, want the opposite end %d", got, want)
	}
}

func TestB35BucolicRanchTapsForAnyColorOnlyForMountSpells(t *testing.T) {
	spec, ok := effects.Lookup(b35BucolicRanchOracle)
	if !ok || len(spec.ManaAbilities) != 2 {
		t.Fatalf("Bucolic Ranch has %d mana abilities, want 2", len(spec.ManaAbilities))
	}
	if len(spec.ManaAbilities[1].Restrictions) == 0 {
		t.Error("the any-color mana has no Mount restriction")
	}
}

func TestB35KellanJoinsUpPlotsASmallCardAndGrowsOnALegend(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil
	cheap := effectstest.PutInHand(me, game.Card{Name: "Cheap Spell", TypeLine: "Instant", ManaCost: "{1}{R}", Owner: me.ID, Controller: me.ID})
	dear := effectstest.PutInHand(me, game.Card{Name: "Dear Spell", TypeLine: "Sorcery", ManaCost: "{3}{R}", Owner: me.ID, Controller: me.ID})
	land := effectstest.PutInHand(me, game.Card{Name: "Spare Land", TypeLine: "Basic Land — Mountain", Owner: me.ID, Controller: me.ID})
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Kellan Joins Up", TypeLine: "Legendary Enchantment",
		ManaCost: "{G}{W}{U}", OracleID: b35KellanJoinsUpOracle})
	prompt := effectstest.LatestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("no exile prompt")
	}
	if len(prompt.ChooseCards) != 1 || prompt.ChooseCards[0] != cheap {
		t.Fatalf("offered %v, want only the mana value 2 nonland card (not %v or %v)", prompt.ChooseCards, dear, land)
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{cheap}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(cheap) {
		t.Error("the card stayed in hand")
	}
	inExile := false
	g.ReadSnapshot(func() {
		for _, c := range g.Exile.Cards {
			if c.InstanceID == cheap {
				inExile = true
			}
		}
	})
	if !inExile {
		t.Fatal("the card is not in exile")
	}
	// A legendary creature entering puts a counter on each creature.
	other := effectstest.B12Creature(g, me.ID, "Other Bear", "Creature — Bear", 2, 2)
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Famous Bear", TypeLine: "Legendary Creature — Bear",
		ManaCost: "{2}", Power: 2, Toughness: 2})
	if p, _ := effectstest.PtOf(t, g, other); p != 3 {
		t.Errorf("the other creature has power %d, want 3", p)
	}
	// A nonlegendary creature does not.
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Plain Bear", TypeLine: "Creature — Bear", ManaCost: "{2}", Power: 2, Toughness: 2})
	if p, _ := effectstest.PtOf(t, g, other); p != 3 {
		t.Errorf("a nonlegendary creature grew the board: power %d", p)
	}
}

func TestB35JaceReawakenedCantBeCastInYourFirstThreeTurns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PutInHand(me, game.Card{Name: "Jace Reawakened", TypeLine: "Legendary Planeswalker — Jace",
		ManaCost: "{U}{U}", OracleID: b35JaceReawakenedOracle, Owner: me.ID, Controller: me.ID})
	effectstest.B06AddMana(me, "U", "U")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast Jace Reawakened on your first turn")
	}
}

func b35Planeswalker(g *game.Game, me *game.Player, name, typeLine, oracle string, loyalty int) uuid.UUID {
	id := effectstest.PushCatalogPermanent(g, me.ID, name, typeLine, oracle, false)
	effectstest.FindBattlefieldCardForTest(g, id).Counters = map[string]int{game.CounterLoyalty: loyalty}
	return id
}

func TestB35JaceReawakenedPlusOneDrawsAndDiscards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	jace := b35Planeswalker(g, me, "Jace Reawakened", "Legendary Planeswalker — Jace", b35JaceReawakenedOracle, 3)
	hand := len(me.Hand.Cards)
	effectstest.Pr7Activate(t, g, me.ID, jace, 0, game.ActivateAbilityParams{})
	effectstest.DiscardFromHand(t, g, me.ID)
	if got := len(me.Hand.Cards); got != hand {
		t.Errorf("hand %d, want %d after loot", got, hand)
	}
}

func TestB35JaceReawakenedMinusSixCopiesEverySpellThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	jace := b35Planeswalker(g, me, "Jace Reawakened", "Legendary Planeswalker — Jace", b35JaceReawakenedOracle, 7)
	effectstest.Pr7Activate(t, g, me.ID, jace, 2, game.ActivateAbilityParams{})
	life := effectstest.LifeOf(g, opp.ID)
	effectstest.CastBoltAtAndKeepTheCopysTarget(t, g, me.ID, opp.ID, 1)
	effectstest.CastBoltAtAndKeepTheCopysTarget(t, g, me.ID, opp.ID, 1)
	if got := effectstest.LifeOf(g, opp.ID); got != life-12 {
		t.Errorf("life %d after two copied Bolts, want %d", got, life-12)
	}
}

func TestB35OkoRingleaderMakesElkAndLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	oko := b35Planeswalker(g, me, "Oko, the Ringleader", "Legendary Planeswalker — Oko", b35OkoRingleaderOracle, 3)
	effectstest.Pr7Activate(t, g, me.ID, oko, 1, game.ActivateAbilityParams{})
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Elk" && c.Controller == me.ID && c.IsCreature() {
			n++
			if p, tt := effectstest.PtOf(t, g, c.InstanceID); p != 3 || tt != 3 {
				t.Errorf("the Elk is %d/%d", p, tt)
			}
		}
	}
	if n != 1 {
		t.Fatalf("%d Elk, want 1", n)
	}
}

func TestB35OkoRingleaderPlusOneDiscardsTwoWithoutACrime(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	oko := b35Planeswalker(g, me, "Oko, the Ringleader", "Legendary Planeswalker — Oko", b35OkoRingleaderOracle, 3)
	hand := len(me.Hand.Cards)
	effectstest.Pr7Activate(t, g, me.ID, oko, 0, game.ActivateAbilityParams{})
	effectstest.DiscardFromHand(t, g, me.ID)
	if got := len(me.Hand.Cards); got != hand+2-2 {
		t.Errorf("hand %d, want %d: draw two, discard two", got, hand)
	}
}

func TestB35OkoRingleaderPlusOneDiscardsOneAfterACrime(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	oko := b35Planeswalker(g, me, "Oko, the Ringleader", "Legendary Planeswalker — Oko", b35OkoRingleaderOracle, 3)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBecomesTarget, CrimeCommitted: true, Actor: me.ID})
	})
	hand := len(me.Hand.Cards)
	effectstest.Pr7Activate(t, g, me.ID, oko, 0, game.ActivateAbilityParams{})
	effectstest.DiscardFromHand(t, g, me.ID)
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("hand %d, want %d: draw two, discard one", got, hand+1)
	}
}

func TestB35OkoRingleaderMinusFiveCopiesEachOtherNonlandPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	oko := b35Planeswalker(g, me, "Oko, the Ringleader", "Legendary Planeswalker — Oko", b35OkoRingleaderOracle, 6)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.GaPut(g, me.ID, "Forest", "Basic Land — Forest", "")
	effectstest.PushToken(g, me.ID, effects.ClueToken())
	before := len(g.Battlefield.Cards)
	effectstest.Pr7Activate(t, g, me.ID, oko, 2, game.ActivateAbilityParams{})
	if got := len(g.Battlefield.Cards); got != before+2 {
		t.Errorf("%d permanents, want %d: a Bear and a Clue copied, not the land or Oko", got, before+2)
	}
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Bear"); n != 1 {
		t.Errorf("%d Bear tokens, want 1", n)
	}
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Clue"); n != 2 {
		t.Errorf("%d Clues, want 2", n)
	}
}

func TestB35OkoRingleaderBecomesAHexproofCopyAtBeginningOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	oko := b35Planeswalker(g, me, "Oko, the Ringleader", "Legendary Planeswalker — Oko", b35OkoRingleaderOracle, 3)
	beast := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Beast", TypeLine: "Creature — Beast", Power: 4, Toughness: 4, Colors: []string{"G"}})
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, beast)
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := effectstest.FindBattlefieldCardForTest(g, oko)
	if c == nil || !c.IsCreature() || c.IsPlaneswalker() {
		t.Fatal("Oko is not a creature copy of the Beast")
	}
	if p, tt := effectstest.PtOf(t, g, oko); p != 4 || tt != 4 {
		t.Errorf("Oko is %d/%d, want 4/4", p, tt)
	}
	if !effectstest.HasEffectiveKeyword(t, g, oko, "hexproof") {
		t.Error("the copy has no hexproof")
	}
}

func TestB35StepBetweenWorldsAsksEachPlayerAndOnlyThoseWhoShuffleDrawSeven(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hands := make([]int, len(g.Seats))
	effectstest.CastCatalogSpell(t, g, "Step Between Worlds", "Sorcery", b35StepBetweenWorldsOracle, nil)
	for i, p := range g.Seats {
		hands[i] = len(p.Hand.Cards)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// The caster answers first, then each opponent in turn order.
	for i, p := range g.Seats {
		answer := cardsst.StepBetweenWorldsNo
		if i%2 == 0 {
			answer = cardsst.StepBetweenWorldsYes
		}
		effectstest.B35AnswerOption(t, g, p.ID, answer)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		want := hands[i]
		if i%2 == 0 {
			want = 7
		}
		if got := len(p.Hand.Cards); got != want {
			t.Errorf("seat %d has %d cards, want %d", i, got, want)
		}
	}
	exiled := false
	g.ReadSnapshot(func() {
		for _, c := range g.Exile.Cards {
			if c.Name == "Step Between Worlds" {
				exiled = true
			}
		}
	})
	if !exiled {
		t.Error("Step Between Worlds is not exiled")
	}
	_ = me
}

func TestB35StepBetweenWorldsHasPlot(t *testing.T) {
	spec, _ := effects.Lookup(b35StepBetweenWorldsOracle)
	if len(spec.SpecialActions) != 1 || spec.SpecialActions[0].Cost != "{4}{U}{U}" {
		t.Errorf("special actions %+v, want plot {4}{U}{U}", spec.SpecialActions)
	}
}

func TestB35LilahExilesAndPlotsAMulticoloredSpellCastFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Lilah, Undefeated Slickshot", "Legendary Creature — Human Rogue", b35LilahOracle, 3, 3)
	spell, err := effectstest.Msc17Cast(t, g, me, game.Card{Name: "Gold Divination", TypeLine: "Sorcery", ManaCost: "{2}{U}{R}",
		Colors: []string{"U", "R"}, OracleID: effectstest.SpellControlDivinationOracle}, game.CastSpellParams{})
	if err != nil {
		t.Fatal(err)
	}
	effectstest.B35SettleTriggerOrders(t, g, me.ID)
	if me.Graveyard.Contains(spell) {
		t.Fatal("the spell went to the graveyard")
	}
	if effectstest.ExiledCard(g, spell) == nil {
		for _, c := range g.PendingChoices {
			t.Logf("pending %v", c.Kind)
		}
		t.Fatalf("the spell is not in exile (stack %d)", len(g.StackMeta))
	}
	plotted := false
	for _, ev := range g.Events {
		if ev.Kind == game.EventBecomesPlotted && ev.CardID == spell {
			plotted = true
		}
	}
	if !plotted {
		t.Error("the exiled spell did not become plotted")
	}
}

func TestB35LilahIgnoresMonocoloredSpellsAndCountersTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Lilah, Undefeated Slickshot", "Legendary Creature — Human Rogue", b35LilahOracle, 3, 3)
	spell, err := effectstest.Msc17Cast(t, g, me, game.Card{Name: "Mono Divination", TypeLine: "Sorcery", ManaCost: "{2}{U}",
		Colors: []string{"U"}, OracleID: effectstest.SpellControlDivinationOracle}, game.CastSpellParams{})
	if err != nil {
		t.Fatal(err)
	}
	effectstest.B35SettleTriggerOrders(t, g, me.ID)
	if !me.Graveyard.Contains(spell) || effectstest.ExiledCard(g, spell) != nil {
		t.Error("a mono-colored spell should go to the graveyard as usual")
	}
}

func TestB35TinybonesCastsANonlandPermanentFromTheDamagedPlayersGraveyardWithAnyMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	tiny := effectstest.B27Push(g, me.ID, "Tinybones, the Pickpocket", "Legendary Creature — Skeleton Rogue", b35TinybonesOracle, "{B}", 1, 1, "B")
	var theirs, land, elsewhere uuid.UUID
	g.WithWriteLock(func() {
		theirs = uuid.New()
		opp.Graveyard.PushTop(game.Card{InstanceID: theirs, Name: "Their Dead Bear", TypeLine: "Creature — Bear",
			ManaCost: "{1}{G}", Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
		land = effectstest.PushGraveyardCardTyped(opp, "Their Dead Land", "Basic Land — Forest")
		elsewhere = effectstest.PushGraveyardCardTyped(other, "Other Dead Bear", "Creature — Bear")
	})
	effectstest.DeclareAttack(t, g, opp.ID, tiny)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	effectstest.FinishBlockDeclarations(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	for i := 0; i < 4 && effectstest.LatestPickTarget(g, me.ID) == nil && effectstest.TriggerOnStack(g, tiny) == nil; i++ {
		effectstest.PassPriorityAroundTable(t, g)
	}
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	for _, bad := range []uuid.UUID{land, elsewhere} {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Fatalf("an illegal card (%s) was accepted", bad)
		}
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err != nil {
		t.Fatal(err)
	}
	effectstest.B06AddMana(me, "R", "R")
	for i := 0; i < 8 && effectstest.KeywordCastPromptFor(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	prompt := effectstest.KeywordCastPromptFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no cast offer")
	}
	effectstest.AnswerKeywordCast(t, g, prompt, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(theirs) {
		t.Error("the Bear did not enter, though red mana was spent on a green cost")
	}
	if c := effectstest.BattlefieldCardPtr(t, g, theirs); c.Controller != me.ID {
		t.Error("the Bear is not under the caster's control")
	}
}

func TestB35VraskaTheSilencerReturnsAnOpponentsCreatureAsATappedTreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Vraska, the Silencer", "Legendary Creature — Gorgon Assassin", b35VraskaSilencerOracle, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	token := effectstest.PushToken(g, opp.ID, effects.MercenaryToken())
	// Neither my own creature nor a token triggers it.
	effectstest.P1Destroy(t, g, mine)
	effectstest.P1Destroy(t, g, token)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless) != nil {
		t.Fatal("offered for my own creature or a token")
	}
	effectstest.P1Destroy(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	pay := effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless)
	if pay == nil {
		t.Fatal("no offer to pay {1}")
	}
	effectstest.B06AddMana(me, "C")
	if err := g.ResolvePayUnless(pay.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, theirs)
	if !ok {
		t.Fatal("the creature card did not return")
	}
	if c.Controller != me.ID || !c.Tapped {
		t.Errorf("controller %s tapped %v, want mine and tapped", c.Controller, c.Tapped)
	}
	if c.IsCreature() || !c.IsArtifact() || !c.HasSubtype("Treasure") {
		t.Errorf("type line %q: want a Treasure artifact and no creature", c.TypeLine)
	}
	if effectstest.GrantedManaRowsOf(g, theirs) != 1 {
		t.Error("the Treasure has no sacrifice-for-mana ability")
	}
}

func TestB35RoxanneMakesATappedMeteoriteThatShocksAndDoublesTokenMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := effectstest.LifeOf(g, opp.ID)
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Roxanne, Starfall Savant", TypeLine: "Legendary Creature — Cat Druid",
		ManaCost: "{3}{R}{G}", OracleID: b35RoxanneOracle, Power: 4, Toughness: 4})
	// The Roxanne trigger makes the Meteorite, whose own trigger asks for a target.
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != life-2 {
		t.Errorf("life %d, want %d from Meteorite's 2 damage", got, life-2)
	}
	var meteorite uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Meteorite" && c.IsToken() {
			meteorite = c.InstanceID
			if !c.Tapped || !c.IsArtifact() || len(c.Colors) != 0 {
				t.Errorf("Meteorite tapped=%v artifact=%v colors=%v, want a tapped colorless artifact", c.Tapped, c.IsArtifact(), c.Colors)
			}
		}
	}
	if meteorite == uuid.Nil {
		t.Fatal("no Meteorite")
	}
	effectstest.UntapForTest(g, meteorite)
	if err := g.ActivateManaAbility(me.ID, meteorite, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("tap Meteorite for mana: %v", err)
	}
	if got := me.ManaPool.CountColor("R"); got != 2 {
		t.Errorf("%d {R} after tapping Meteorite for red under Roxanne, want 2", got)
	}
}

func TestB35DocAurlockMakesYourPlottingCostTwoLess(t *testing.T) {
	const aloeOracle = "97489ef7-98c3-4700-bbe6-185215d41b25"
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	aloe := effectstest.PlotHandCard(me, "Aloe Alchemist", "Creature — Plant Warlock", "{1}{G}", aloeOracle, 3, 2)
	// Without Doc Aurlock the {1}{G} plot cost is not payable with {G}.
	me.ManaPool.AddMana(game.ManaToken{Color: "G"})
	if err := g.PerformSpecialAction(me.ID, aloe, game.SpecialActionPlot, game.SpecialActionParams{Strict: true}); err == nil {
		t.Fatal("plotted for {G} with no Doc Aurlock")
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Doc Aurlock, Grizzled Genius", "Legendary Creature — Bear Druid", b35DocAurlockOracle, false)
	if err := g.PerformSpecialAction(me.ID, aloe, game.SpecialActionPlot, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("plot with Doc Aurlock out: %v", err)
	}
	if !g.Exile.Contains(aloe) {
		t.Error("the plotted card is not in exile")
	}
}

func TestB35DocAurlockDoesNotReducePlottingForAnOpponent(t *testing.T) {
	const aloeOracle = "97489ef7-98c3-4700-bbe6-185215d41b25"
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.PushCatalogPermanent(g, opp.ID, "Doc Aurlock, Grizzled Genius", "Legendary Creature — Bear Druid", b35DocAurlockOracle, false)
	aloe := effectstest.PlotHandCard(me, "Aloe Alchemist", "Creature — Plant Warlock", "{1}{G}", aloeOracle, 3, 2)
	me.ManaPool.AddMana(game.ManaToken{Color: "G"})
	if err := g.PerformSpecialAction(me.ID, aloe, game.SpecialActionPlot, game.SpecialActionParams{Strict: true}); err == nil {
		t.Fatal("an opponent's Doc Aurlock reduced your plot cost")
	}
}

func TestB35DocAurlockMakesSpellsCastFromYourGraveyardCostTwoLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	looting := effectstest.SeedGraveyardCard(t, g, "Faithless Looting", "Sorcery", "3d6fa57a-aa53-4b5c-b8af-a7612c823117")
	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	if err := g.CastSpell(me.ID, looting, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "flashback"}); err == nil {
		t.Fatal("flashed back Faithless Looting for {R} with no Doc Aurlock")
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Doc Aurlock, Grizzled Genius", "Legendary Creature — Bear Druid", b35DocAurlockOracle, false)
	if err := g.CastSpell(me.ID, looting, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback {2}{R} should cost {R} under Doc Aurlock: %v", err)
	}
}

func TestB35BonnyPallMakesBeauWhoCountsYourLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.SeedLandOnBattlefield(g, me.ID, "Island", "Basic Land — Island")
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Bonny Pall, Clearcutter", TypeLine: "Legendary Creature — Giant Scout",
		ManaCost: "{3}{G}{U}{U}", OracleID: b35BonnyPallOracle, Power: 6, Toughness: 5})
	var beau uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Beau" && c.IsToken() {
			beau = c.InstanceID
			if !c.IsLegendary() || !c.HasSubtype("Ox") || len(c.Colors) != 1 || c.Colors[0] != "U" {
				t.Errorf("Beau is %q colors %v, want a legendary blue Ox", c.TypeLine, c.Colors)
			}
		}
	}
	if beau == uuid.Nil {
		t.Fatal("no Beau")
	}
	if p, tt := effectstest.PtOf(t, g, beau); p != 2 || tt != 2 {
		t.Errorf("Beau is %d/%d with two lands, want 2/2", p, tt)
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Swamp", "Basic Land — Swamp", "", false)
	if p, tt := effectstest.PtOf(t, g, beau); p != 3 || tt != 3 {
		t.Errorf("Beau is %d/%d with three lands, want 3/3", p, tt)
	}
}

func TestB35BonnyPallDrawsOnAttackAndMayPutALandFromHandOrGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Bonny Pall, Clearcutter", "Legendary Creature — Giant Scout", b35BonnyPallOracle, 6, 5)
	runner := effectstest.PushVanillaCreature(g, me.ID, "Runner", 2, 2)
	var gyLand uuid.UUID
	g.WithWriteLock(func() { gyLand = effectstest.PushGraveyardCardTyped(me, "Dead Forest", "Basic Land — Forest") })
	handLand := effectstest.PutInHand(me, game.Card{Name: "Hand Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
	hand := len(me.Hand.Cards)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(runner, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCards(g, me.ID)
	if c == nil {
		t.Fatal("no land prompt")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range c.ChooseCards {
		offered[id] = true
	}
	if !offered[gyLand] || !offered[handLand] {
		t.Fatalf("offered %v, want the graveyard land and the hand land", c.ChooseCards)
	}
	if len(me.Hand.Cards) < hand+1 {
		t.Errorf("hand %d, want at least %d after the draw", len(me.Hand.Cards), hand+1)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{gyLand}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(gyLand) {
		t.Error("the graveyard land did not enter the battlefield")
	}
}

func TestB35KambalCopiesEachOpponentTokenOnceEachTurnAndDrainsForYours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Kambal, Profiteering Mayor", "Legendary Creature — Human Advisor", b35KambalOracle, 2, 3)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.TreasureToken(), 2) })
	effectstest.PassPriorityAroundTable(t, g)
	mine, tapped := effectstest.CountTokensNamed(g, me.ID, "Treasure")
	if mine != 2 || tapped != 2 {
		t.Fatalf("%d Treasures (%d tapped) under you, want 2 tapped copies", mine, tapped)
	}
	if got := effectstest.CountTreasures(g, opp.ID); got != 2 {
		t.Errorf("the opponent has %d Treasures, want their own 2", got)
	}
	// Only once each turn.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.TreasureToken(), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Treasure"); n != 2 {
		t.Errorf("a second batch the same turn made copies: %d Treasures under you", n)
	}
	// Your own tokens drain, once per batch.
	life, oppLife := effectstest.LifeOf(g, me.ID), effectstest.LifeOf(g, opp.ID)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.ClueToken(), 3) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LifeOf(g, me.ID) != life+1 || effectstest.LifeOf(g, opp.ID) != oppLife-1 {
		t.Errorf("drain: you %d (from %d), them %d (from %d)", effectstest.LifeOf(g, me.ID), life, effectstest.LifeOf(g, opp.ID), oppLife)
	}
}

func TestB35EsotericDuplicatorCopiesASacrificedTokenAtTheNextEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Esoteric Duplicator", "Artifact — Clue", b35DuplicatorOracle, 0, 0)
	clue := effectstest.PushToken(g, me.ID, effects.ClueToken())
	effectstest.B06AddMana(me, "C", "C")
	effectstest.B16Activate(t, g, me.ID, clue, 0, game.ActivateAbilityParams{})
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Clue"); n != 0 {
		t.Fatalf("%d Clues after the sacrifice, want 0", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pay := effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless)
	if pay == nil {
		t.Fatal("no offer to pay {2}")
	}
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ResolvePayUnless(pay.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Clue"); n != 0 {
		t.Fatalf("the copy appeared before the end step: %d Clues", n)
	}
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Clue"); n != 1 {
		t.Fatalf("%d Clues after the end step, want the one copy", n)
	}
}

func TestB35EsotericDuplicatorDoesNothingIfYouDontPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Esoteric Duplicator", "Artifact — Clue", b35DuplicatorOracle, 0, 0)
	clue := effectstest.PushToken(g, me.ID, effects.ClueToken())
	effectstest.B06AddMana(me, "C", "C")
	effectstest.B16Activate(t, g, me.ID, clue, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	pay := effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless)
	if pay == nil {
		t.Fatal("no offer to pay {2}")
	}
	if err := g.ResolvePayUnless(pay.ID, me.ID, false); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Clue"); n != 0 {
		t.Errorf("%d Clues, want none", n)
	}
}

func TestB35LootExilesAsManyCardsAsCardTypesAmongYourOtherNonlandPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Loot, the Key to Everything", "Legendary Creature — Beast Noble", b35LootOracle, 3, 3)
	effectstest.B12Push(g, me.ID, "Golem Friend", "Artifact Creature — Golem", "", 2, 2)
	effectstest.B12Push(g, me.ID, "Old Aura", "Enchantment", "", 0, 0)
	effectstest.GaPut(g, me.ID, "Forest", "Basic Land — Forest", "")
	// Another creature and the land add no new type.
	effectstest.B12Creature(g, me.ID, "More Bears", "Creature — Bear", 2, 2)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	before := me.Library.Size()
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := before - me.Library.Size(); got != 3 {
		t.Errorf("exiled %d cards, want 3 (artifact, creature, enchantment)", got)
	}
	if n := len(g.Exile.Cards); n < 3 {
		t.Errorf("%d cards in exile, want at least 3", n)
	}
}

func TestB35LaughingJasperFlintMakesStolenCreaturesMercenariesAndExilesXOffAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Laughing Jasper Flint", "Legendary Creature — Lizard Rogue", b35JasperFlintOracle, 3, 3)
	own := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	stolen := effectstest.B12Creature(g, opp.ID, "Stolen Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		if c := effectstest.FindBattlefieldCardForTest(g, stolen); c != nil {
			c.Controller = me.ID
		}
		g.RecomputeLayersIfStaleLocked()
	})
	effectstest.PushCatalogPermanent(g, me.ID, "Poke", "Creature — Bear", "", false) // forces a layer pass
	if c, ok := effectstest.BattlefieldCard(g, stolen); !ok || !c.HasSubtype("Mercenary") {
		t.Error("the stolen creature is not a Mercenary")
	}
	if c, ok := effectstest.BattlefieldCard(g, own); !ok || c.HasSubtype("Mercenary") {
		t.Error("your own creature became a Mercenary")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	before := len(g.Exile.Cards)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(g.Exile.Cards) - before; got != 2 {
		t.Errorf("exiled %d cards off the opponent's library, want 2 (Jasper Flint and the stolen Mercenary)", got)
	}
}

func TestB35GhiredLetsNontokenCreaturesCopyATokenThatEnteredThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Ghired, Mirror of the Wilds", "Legendary Creature — Human Shaman", b35GhiredOracle, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.ClueToken(), 1) })
	clue := uuid.Nil
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Clue" && c.IsToken() {
			clue = c.InstanceID
		}
	}
	if clue == uuid.Nil {
		t.Fatal("no Clue")
	}
	idx, ref := effectstest.GrantedActivatedIndex(t, g, bear)
	if idx < 0 {
		t.Fatal("the nontoken creature has no granted ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, bear, idx, game.ActivateAbilityParams{Ref: ref, Targets: effectstest.B16TargetCard(clue)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Clue"); n != 2 {
		t.Fatalf("%d Clues, want 2", n)
	}
	// A token does not get the ability.
	if i, _ := effectstest.GrantedActivatedIndex(t, g, clue); i >= 0 {
		t.Error("a token was granted the ability")
	}
}

// b35KellanSetup puts Kellan out, a cheap creature, an expensive one and a
// land in hand, and flashes back Faithless Looting (mana value 1).
func b35KellanSetup(t *testing.T) (g *game.Game, me *game.Player, cheap, dear, land uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.B12Push(g, me.ID, "Kellan, the Kid", "Legendary Creature — Human Faerie Rogue", b35KellanKidOracle, 3, 3)
	me.Hand.Cards = nil
	cheap = effectstest.PutInHand(me, game.Card{Name: "Cheap Bear", TypeLine: "Creature — Bear", ManaCost: "{1}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	dear = effectstest.PutInHand(me, game.Card{Name: "Dear Bear", TypeLine: "Creature — Bear", ManaCost: "{4}", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
	land = effectstest.PutInHand(me, game.Card{Name: "Hand Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	looting := effectstest.SeedGraveyardCard(t, g, "Faithless Looting", "Sorcery", "3d6fa57a-aa53-4b5c-b8af-a7612c823117")
	for i := range me.Graveyard.Cards {
		if me.Graveyard.Cards[i].InstanceID == looting {
			me.Graveyard.Cards[i].ManaCost = "{R}"
		}
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "R"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.CastSpell(me.ID, looting, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	return g, me, cheap, dear, land
}

func TestB35KellanTheKidCastsAnEqualOrLesserPermanentFreeAndThenNoLand(t *testing.T) {
	g, me, cheap, dear, land := b35KellanSetup(t)
	var prompt *game.PendingChoice
	for i := 0; i < 8 && prompt == nil; i++ {
		if prompt = effectstest.KeywordCastPromptFor(g, me.ID); prompt == nil {
			if err := g.PassPriority(); err != nil {
				t.Fatal(err)
			}
		}
	}
	if prompt == nil {
		t.Fatal("no free-cast offer")
	}
	for _, id := range prompt.DiscardOptions {
		if id == dear || id == land {
			t.Fatalf("offered %v: a costlier creature or a land", id)
		}
	}
	effectstest.AnswerKeywordCast(t, g, prompt, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) {
		t.Error("the Bear was not cast for free")
	}
	if g.Battlefield.Contains(land) {
		t.Error("a land entered after a spell was cast")
	}
}

func TestB35KellanTheKidMayPutALandIfYouDontCast(t *testing.T) {
	g, me, cheap, _, land := b35KellanSetup(t)
	var prompt *game.PendingChoice
	for i := 0; i < 8 && prompt == nil; i++ {
		if prompt = effectstest.KeywordCastPromptFor(g, me.ID); prompt == nil {
			if err := g.PassPriority(); err != nil {
				t.Fatal(err)
			}
		}
	}
	if prompt == nil {
		t.Fatal("no free-cast offer")
	}
	effectstest.AnswerKeywordCast(t, g, prompt, false)
	c := effectstest.LatestChooseCards(g, me.ID)
	if c == nil {
		t.Fatal("no land prompt after declining")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{land}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(land) {
		t.Error("the land did not enter the battlefield")
	}
	if g.Battlefield.Contains(cheap) {
		t.Error("the Bear entered though the cast was declined")
	}
}

func TestB35ErietteStealsAPermanentWhileYourAuraIsAttachedToIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Eriette, the Beguiler", "Legendary Creature — Human Warlock", b35ErietteOracle, 3, 3)
	cheap := effectstest.B12Push(g, opp.ID, "Cheap Bear", "Creature — Bear", "", 2, 2)
	dear := effectstest.B12Push(g, opp.ID, "Dear Bear", "Creature — Bear", "", 5, 5)
	for _, id := range []uuid.UUID{cheap, dear} {
		c := effectstest.FindBattlefieldCardForTest(g, id)
		if id == cheap {
			c.ManaCost = "{1}{G}"
		} else {
			c.ManaCost = "{4}{G}"
		}
	}
	aura := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Mind Control Aura",
		TypeLine: "Enchantment — Aura", ManaCost: "{2}{U}", Owner: me.ID, Controller: me.ID})
	// Too expensive a host: no trigger.
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(aura, game.TargetRef{Kind: game.TargetCard, ID: dear}); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, dear); c.Controller != opp.ID {
		t.Fatal("took control of a permanent with greater mana value")
	}
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(aura, game.TargetRef{Kind: game.TargetCard, ID: cheap}); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, cheap); c.Controller != me.ID {
		t.Fatal("did not take control of the cheap Bear")
	}
	// Moving the Aura away returns the Bear.
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(aura, game.TargetRef{Kind: game.TargetCard, ID: dear}); err != nil {
			t.Fatal(err)
		}
		g.RecomputeLayersIfStaleLocked()
	})
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, cheap); c.Controller != opp.ID {
		t.Error("the Bear did not go back when the Aura left it")
	}
}

func TestB35RakdosSacrificeExilesCardsEqualToItsManaValueAndMakesItIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	rakdos := effectstest.B12Push(g, me.ID, "Rakdos, the Muscle", "Legendary Creature — Demon Mercenary", b35RakdosOracle, 6, 5)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 2, 2)
	effectstest.FindBattlefieldCardForTest(g, fodder).ManaCost = "{2}{B}"
	other := effectstest.B12Creature(g, me.ID, "More Fodder", "Creature — Bear", 2, 2)
	before := len(g.Exile.Cards)
	effectstest.B16Activate(t, g, me.ID, rakdos, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}})
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(g.Exile.Cards) - before; got != 3 {
		t.Errorf("exiled %d cards, want 3 (mana value of the sacrificed Bear)", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, rakdos, "indestructible") {
		t.Error("Rakdos has no indestructible")
	}
	if !effectstest.B16Tapped(t, g, rakdos) {
		t.Error("Rakdos is not tapped")
	}
	// Once each turn.
	if err := g.ActivateCatalogAbility(me.ID, rakdos, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{other}}); err == nil {
		t.Error("activated a second time in one turn")
	}
}
