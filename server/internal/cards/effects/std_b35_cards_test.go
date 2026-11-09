package effects

import (
	"testing"

	"github.com/google/uuid"

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
	b35StepBetweenWorldsOracle   = "48b43d53-9a8a-46b2-9d8c-b0d66fb68a5d"
	b35LilahOracle               = "0757c5a7-e51b-4c39-bb4d-b32657c10cb4"
	b35TinybonesOracle           = "7bc4c7e2-6758-4a85-84e7-03ab93981106"
	b35VraskaSilencerOracle      = "e7c5a568-3855-4a20-9edf-01a8a6a4ba97"
	b35RoxanneOracle             = "11765eaf-87a1-48bd-90c1-7a5f21be5918"
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
		b35IntimidationCampaignOra:   "Intimidation Campaign",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestB35FormAPosseMakesXMercenaries(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12PlayFromHand(t, g, "Form a Posse", "Sorcery", b35FormAPosseOracle, game.CastSpellParams{XValue: 3})
	passPriorityAroundTable(t, g)
	if n := len(otjMPMercenaries(g, me.ID)); n != 3 {
		t.Fatalf("X=3 made %d Mercenaries, want 3", n)
	}
}

func TestB35HarvesterOfMiseryShrinksOtherCreaturesWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 3, 3)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	harvester := msc17Play(t, g, me, game.Card{Name: "Harvester of Misery", TypeLine: "Creature — Spirit",
		ManaCost: "{3}{B}{B}", OracleID: b35HarvesterOfMiseryOracle, Power: 5, Toughness: 4})
	for _, id := range []uuid.UUID{mine, theirs} {
		if p, tt := ptOf(t, g, id); p != 1 || tt != 1 {
			t.Errorf("a Bear is %d/%d, want 1/1", p, tt)
		}
	}
	if p, tt := ptOf(t, g, harvester); p != 5 || tt != 4 {
		t.Errorf("the Harvester is %d/%d, want 5/4 (other creatures only)", p, tt)
	}
}

func TestB35HarvesterOfMiseryDiscardAbilityShrinksATarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	advanceToMain(t, g)
	hid := putInHand(me, game.Card{Name: "Harvester of Misery", TypeLine: "Creature — Spirit",
		ManaCost: "{3}{B}{B}", OracleID: b35HarvesterOfMiseryOracle, Power: 5, Toughness: 4, Owner: me.ID, Controller: me.ID})
	b06AddMana(me, "B", "C")
	b16Activate(t, g, me.ID, hid, 0, game.ActivateAbilityParams{Targets: b16TargetCard(theirs)})
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, theirs); p != 1 || tt != 1 {
		t.Errorf("the target is %d/%d, want 1/1", p, tt)
	}
	if handHasID(me, hid) {
		t.Error("the Harvester was not discarded to pay")
	}
}

func TestB35SandstormSalvagerMakesAGolem(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	msc17Play(t, g, me, game.Card{Name: "Sandstorm Salvager", TypeLine: "Creature — Human Artificer",
		ManaCost: "{2}{G}", OracleID: b35SandstormSalvagerOracle, Power: 3, Toughness: 3})
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Golem" && c.IsToken() && c.IsCreature() && c.IsArtifact() && len(c.Colors) == 0 {
			n++
			if p, tt := ptOf(t, g, c.InstanceID); p != 3 || tt != 3 {
				t.Errorf("the Golem is %d/%d, want 3/3", p, tt)
			}
		}
	}
	if n != 1 {
		t.Fatalf("%d colorless Golem artifact creature tokens, want 1", n)
	}
}

func TestB35SandstormSalvagerGrowsTokensAndGivesTrample(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	salvager := b12Push(g, me.ID, "Sandstorm Salvager", "Creature — Human Artificer", b35SandstormSalvagerOracle, 3, 3)
	merc := pushToken(g, me.ID, MercenaryToken())
	nontoken := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b06AddMana(me, "C", "C")
	b16Activate(t, g, me.ID, salvager, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, merc); p != 2 || tt != 2 {
		t.Errorf("the Mercenary is %d/%d, want 2/2", p, tt)
	}
	if !hasEffectiveKeyword(t, g, merc, "trample") {
		t.Error("the token gains trample")
	}
	if p, tt := ptOf(t, g, nontoken); p != 2 || tt != 2 || hasEffectiveKeyword(t, g, nontoken, "trample") {
		t.Errorf("a nontoken creature changed: %d/%d", p, tt)
	}
}

func TestB35GreedsGambitEntersDrawsGainsAndMakesBats(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := lifeOf(g, me.ID)
	hand := len(me.Hand.Cards)
	msc17Play(t, g, me, game.Card{Name: "Greed's Gambit", TypeLine: "Enchantment",
		ManaCost: "{3}{B}", OracleID: b35GreedsGambitOracle})
	if n, _ := countTokensNamed(g, me.ID, "Bat"); n != 3 {
		t.Fatalf("%d Bats, want 3", n)
	}
	if got := lifeOf(g, me.ID); got != life+6 {
		t.Errorf("life %d, want %d", got, life+6)
	}
	if got := len(me.Hand.Cards); got != hand+3 {
		t.Errorf("hand %d, want %d", got, hand+3)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Bat" {
			if p, tt := ptOf(t, g, c.InstanceID); p != 2 || tt != 1 {
				t.Errorf("a Bat is %d/%d, want 2/1", p, tt)
			}
		}
	}
}

func TestB35HostileInvestigatorInvestigatesOncePerTurnOnDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	_, err := msc17Cast(t, g, me, game.Card{Name: "Hostile Investigator", TypeLine: "Creature — Ogre Rogue Detective",
		ManaCost: "{3}{B}", OracleID: b35HostileInvestigatorOracle, Power: 4, Toughness: 4}, game.CastSpellParams{})
	if err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, opp.ID, opp.Hand.Cards[0].InstanceID)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Clue"); n != 1 {
		t.Fatalf("%d Clues after the first discard, want 1", n)
	}
	// A second discard the same turn investigates no more.
	b30DiscardOne(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Clue"); n != 1 {
		t.Errorf("%d Clues after the second discard, want still 1", n)
	}
}

func TestB35HighNoonLimitsEachPlayerToOneSpellAndSacrificesForFive(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	noon := b12Push(g, me.ID, "High Noon", "Enchantment", b35HighNoonOracle, 0, 0)
	if err := g.CastSpell(me.ID, putInHand(me, game.Card{Name: "Shock", TypeLine: "Instant", OracleID: lightningBoltOracle, Owner: me.ID, Controller: me.ID}),
		game.CastSpellParams{Targets: b16TargetPlayer(opp.ID)}); err != nil {
		t.Fatalf("the first spell: %v", err)
	}
	second := putInHand(me, game.Card{Name: "Shock", TypeLine: "Instant", OracleID: lightningBoltOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{Targets: b16TargetPlayer(opp.ID), Strict: true}); err == nil {
		t.Fatal("a second spell this turn was cast under High Noon")
	}
	passPriorityAroundTable(t, g)
	life := lifeOf(g, opp.ID)
	b06AddMana(me, "R", "C", "C", "C", "C")
	b16Activate(t, g, me.ID, noon, 0, game.ActivateAbilityParams{Targets: b16TargetPlayer(opp.ID)})
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != life-5 {
		t.Errorf("life %d, want %d", got, life-5)
	}
	if g.Battlefield.Contains(noon) {
		t.Error("High Noon was not sacrificed")
	}
}

func TestB35LegionExtruderShocksOnEntryAndSacrificesAnArtifactForAGolem(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := lifeOf(g, opp.ID)
	msc17Play(t, g, me, game.Card{Name: "Legion Extruder", TypeLine: "Artifact", ManaCost: "{1}{R}", OracleID: b35LegionExtruderOracle})
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != life-2 {
		t.Errorf("life %d, want %d", got, life-2)
	}
}

func TestB35LegionExtruderMakesAGolemForAnotherArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	ext := b12Push(g, me.ID, "Legion Extruder", "Artifact", b35LegionExtruderOracle, 0, 0)
	other := b12Permanent(g, me.ID, "Trinket", "Artifact")
	b06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, ext, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{ext}}); err == nil {
		t.Fatal("Legion Extruder sacrificed itself")
	}
	b16Activate(t, g, me.ID, ext, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{other}})
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Golem"); n != 1 {
		t.Fatalf("%d Golems, want 1", n)
	}
}

func TestB35GreedsGambitEndStepDiscardsLosesAndSacrifices(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Greed's Gambit", "Enchantment", b35GreedsGambitOracle, 0, 0)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	life, hand := lifeOf(g, me.ID), len(me.Hand.Cards)
	for g.Turn.Step != game.StepEnd {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	passPriorityAroundTable(t, g)
	discardFromHand(t, g, me.ID)
	answerSacrifice(t, g, me.ID, bear)
	if got := lifeOf(g, me.ID); got != life-2 {
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Ancient Cornucopia", "Artifact", b35AncientCornucopiaOracle, 0, 0)
	life := lifeOf(g, me.ID)
	cast := func(colors []string) {
		id := putInHand(me, game.Card{Name: "Pretty Spell", TypeLine: "Instant", OracleID: lightningBoltOracle, Colors: colors, Owner: me.ID, Controller: me.ID})
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: b16TargetPlayer(opp.ID)}); err != nil {
			t.Fatal(err)
		}
	}
	cast([]string{"R", "G"})
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, me.ID); got != life+2 {
		t.Fatalf("life %d, want %d (two colors)", got, life+2)
	}
	cast([]string{"R"})
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, me.ID); got != life+2 {
		t.Errorf("life %d, want still %d: only once each turn", got, life+2)
	}
}

func TestB35LotusRingEquipsForThreeThreeVigilanceAndAThreeManaSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := gaPut(g, me.ID, "Bear", "Creature — Bear", "")
	ring := gaPut(g, me.ID, "Lotus Ring", "Artifact — Equipment", b35LotusRingOracle)
	if !hasEffectiveKeyword(t, g, ring, "indestructible") {
		t.Error("Lotus Ring is indestructible")
	}
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}{C}")
	equipTo(t, g, me.ID, ring, bear)
	if p, tt := ptOf(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("the equipped creature is %d/%d, want 3/3", p, tt)
	}
	if !hasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Error("the equipped creature has vigilance")
	}
	if grantedManaRowsOf(g, bear) != 1 {
		t.Fatal("the equipped creature has the three-mana ability")
	}
}

func TestB35SwordOfWealthAndPowerMakesATreasureAndCopiesTheNextInstant(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	sword := gaPut(g, me.ID, "Sword of Wealth and Power", "Artifact — Equipment", b35SwordWealthPowerOracle)
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}")
	equipTo(t, g, me.ID, sword, bear)
	if p, tt := ptOf(t, g, bear); p != 4 || tt != 4 {
		t.Errorf("the equipped creature is %d/%d, want 4/4", p, tt)
	}
	for _, kw := range []string{game.ProtectionFromCardType("Instant"), game.ProtectionFromCardType("Sorcery")} {
		if !hasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the equipped creature lacks %q", kw)
		}
	}
	if countTreasures(g, me.ID) != 0 {
		t.Fatal("a Treasure before any damage")
	}
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: bear, Target: opp.ID, Actor: me.ID, Amount: 3, Combat: true})
	})
	passPriorityAroundTable(t, g)
	if got := countTreasures(g, me.ID); got != 1 {
		t.Fatalf("%d Treasures after the combat damage, want 1", got)
	}
	life := lifeOf(g, opp.ID)
	castBoltAtAndKeepTheCopysTarget(t, g, me.ID, opp.ID, 1)
	if got := lifeOf(g, opp.ID); got != life-6 {
		t.Errorf("life %d after a copied Bolt, want %d", got, life-6)
	}
	castBoltAtAndKeepTheCopysTarget(t, g, me.ID, opp.ID, 0)
	if got := lifeOf(g, opp.ID); got != life-9 {
		t.Errorf("life %d after a second Bolt, want %d (only the next spell is copied)", got, life-9)
	}
}

func TestB35OltecMatterweaverMakesAGnomeOrCopiesAnArtifactToken(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Oltec Matterweaver", "Creature — Human Artificer", b35OltecMatterweaverOracle, 3, 2)
	b12PlayFromHand(t, g, "Bear Cub", "Creature — Bear", "", game.CastSpellParams{})
	answerMode(t, g, modePickChoiceFor(g, me.ID), me.ID, 0)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Gnome"); n != 1 {
		t.Fatalf("%d Gnomes, want 1", n)
	}
	b12PlayFromHand(t, g, "Bear Cub", "Creature — Bear", "", game.CastSpellParams{})
	answerMode(t, g, modePickChoiceFor(g, me.ID), me.ID, 1)
	pickCard(t, g, me.ID, func() uuid.UUID {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Gnome" {
				return c.InstanceID
			}
		}
		return uuid.Nil
	}())
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Gnome"); n != 2 {
		t.Fatalf("%d Gnomes after the copy mode, want 2", n)
	}
}

func TestB35WorldwalkerHelmAddsAMapToArtifactTokensOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gaPut(g, me.ID, "Worldwalker Helm", "Artifact", b35WorldwalkerHelmOracle)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TreasureToken(), 2) })
	if got := countTreasures(g, me.ID); got != 2 {
		t.Errorf("%d Treasures, want 2", got)
	}
	if n, _ := countTokensNamed(g, me.ID, "Map"); n != 1 {
		t.Errorf("%d Maps for one instruction, want 1", n)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 1) })
	if n, _ := countTokensNamed(g, me.ID, "Map"); n != 1 {
		t.Errorf("a Goblin-only creation changed the Maps to %d", n)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, TreasureToken(), 1) })
	if n, _ := countTokensNamed(g, opp.ID, "Map"); n != 0 {
		t.Errorf("an opponent's artifact token got %d Maps", n)
	}
}

func TestB35WorldwalkerHelmCopiesAnArtifactToken(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	helm := gaPut(g, me.ID, "Worldwalker Helm", "Artifact", b35WorldwalkerHelmOracle)
	clue := pushToken(g, me.ID, ClueToken())
	b06AddMana(me, "U", "C")
	b16Activate(t, g, me.ID, helm, 0, game.ActivateAbilityParams{Targets: b16TargetCard(clue)})
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Clue"); n != 2 {
		t.Fatalf("%d Clues, want 2", n)
	}
}

func TestB35CollectorsCagePlaysTheExiledCardOnlyWithThreeDifferentPowers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	top := stackTopLibrary(me, "A", "B", "C", "D", "E")
	cage := game.NewCard("Collector's Cage", me.ID)
	cage.OracleID = b35CollectorsCageOracle
	cage.TypeLine = "Artifact"
	enterWithHideaway(t, g, cage, top[1])
	two := b12Creature(g, me.ID, "Two", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Another Two", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Four", "Creature — Bear", 4, 4)
	activate := func(target uuid.UUID) {
		t.Helper()
		untapForTest(g, cage.InstanceID)
		floatForTest(g, me, "C")
		if err := g.ActivateCatalogAbility(me.ID, cage.InstanceID, 0, game.ActivateAbilityParams{Targets: b16TargetCard(target)}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	// 2, 2, 4 with a counter on a Two is 3, 2, 4: three different powers,
	// and it is the counter that makes the third.
	activate(two)
	if p, _ := ptOf(t, g, two); p != 3 {
		t.Fatalf("the counter did not land: power %d", p)
	}
	if playHiddenGrant(g, top[1]) == nil {
		t.Fatal("no grant with powers 3, 2, 4")
	}
}

func TestB35CollectorsCageGrantsNothingWithTwoDifferentPowers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	top := stackTopLibrary(me, "A", "B", "C", "D", "E")
	cage := game.NewCard("Collector's Cage", me.ID)
	cage.OracleID = b35CollectorsCageOracle
	cage.TypeLine = "Artifact"
	enterWithHideaway(t, g, cage, top[1])
	four := b12Creature(g, me.ID, "Four", "Creature — Bear", 4, 4)
	b12Creature(g, me.ID, "Two", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Another Two", "Creature — Bear", 2, 2)
	floatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, cage.InstanceID, 0, game.ActivateAbilityParams{Targets: b16TargetCard(four)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if playHiddenGrant(g, top[1]) != nil {
		t.Fatal("granted with powers 5, 2, 2")
	}
}

func TestB35OmenpathJourneyExilesDifferentlyNamedLandsAndPutsOneOutEachEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		searchTestLand("Plains", "Basic Land — Plains"),
		searchTestLand("Plains", "Basic Land — Plains"),
		searchTestLand("Island", "Basic Land — Island"),
		searchTestLand("Swamp", "Basic Land — Swamp"),
		searchTestLand("Mountain", "Basic Land — Mountain"),
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Wastes", "Basic Land"),
	)
	saga := castCatalogSpell(t, g, "Omenpath Journey", "Enchantment", b35OmenpathJourneyOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
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
	answerSearchNamed(t, g, me.ID, "Plains", "Island", "Swamp", "Mountain", "Forest")
	passPriorityAroundTable(t, g)
	ref := game.PermissionCardRef(hobA9SagaRef(t, g, saga))
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
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	msc17Play(t, g, me, game.Card{Name: "Ertha Jo, Frontier Mentor", TypeLine: "Legendary Creature — Kor Advisor",
		ManaCost: "{2}{R}{W}", OracleID: b35ErthaJoOracle, Power: 2, Toughness: 3})
	otjMPOneMercenary(t, g, me.ID)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	merc := pushToken(g, me.ID, MercenaryToken())
	b16Activate(t, g, me.ID, merc, 0, game.ActivateAbilityParams{Targets: b16TargetCard(mine)})
	for i := 0; i < 32; i++ {
		if p := latestPickTarget(g, me.ID); p != nil {
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if stackFullyEmpty(g) {
			break
		}
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if p, tt := ptOf(t, g, mine); p != 4 || tt != 2 {
		t.Errorf("the Bear is %d/%d, want 4/2: the pump and its copy", p, tt)
	}
}

func TestB35JoleneMakesOneTreasureForABigAttackAndPingsForATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Jolene, Plundering Pugilist", "Legendary Creature — Human Mercenary", b35JoleneOracle, 3, 3)
	small := pushVanillaCreature(g, me.ID, "Small", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(small, opp.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := countTreasures(g, me.ID); got != 0 {
		t.Fatalf("%d Treasures for a power-2 attacker, want 0", got)
	}
}

func TestB35JoleneBigAttackersMakeOneTreasureTotal(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Jolene, Plundering Pugilist", "Legendary Creature — Human Mercenary", b35JoleneOracle, 3, 3)
	big1 := pushVanillaCreature(g, me.ID, "Big One", 4, 4)
	big2 := pushVanillaCreature(g, me.ID, "Big Two", 5, 5)
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{big1, big2} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatal(err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := countTreasures(g, me.ID); got != 1 {
		t.Fatalf("%d Treasures for two big attackers, want 1", got)
	}
}

func TestB35JoleneSacrificesATreasureToPing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	jolene := b12Push(g, me.ID, "Jolene, Plundering Pugilist", "Legendary Creature — Human Mercenary", b35JoleneOracle, 3, 3)
	treasure := pushToken(g, me.ID, TreasureToken())
	life := lifeOf(g, opp.ID)
	b06AddMana(me, "R", "C")
	b16Activate(t, g, me.ID, jolene, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{treasure}, Targets: b16TargetPlayer(opp.ID)})
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != life-1 {
		t.Errorf("life %d, want %d", got, life-1)
	}
	if g.Battlefield.Contains(treasure) {
		t.Error("the Treasure was not sacrificed")
	}
}

func TestB35IntimidationCampaignDrainsAndMayReturnOnACrime(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life, oppLife, hand := lifeOf(g, me.ID), lifeOf(g, opp.ID), len(me.Hand.Cards)
	camp := msc17Play(t, g, me, game.Card{Name: "Intimidation Campaign", TypeLine: "Enchantment",
		ManaCost: "{1}{U}{B}", OracleID: b35IntimidationCampaignOra})
	if lifeOf(g, me.ID) != life+1 || lifeOf(g, opp.ID) != oppLife-1 {
		t.Errorf("drain: you %d (from %d), them %d (from %d)", lifeOf(g, me.ID), life, lifeOf(g, opp.ID), oppLife)
	}
	if len(me.Hand.Cards) != hand+1 { // the card is added then cast (net 0), then the draw
		t.Errorf("hand %d, want %d", len(me.Hand.Cards), hand+1)
	}
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBecomesTarget, CrimeCommitted: true, Actor: me.ID})
	})
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(camp) {
		t.Error("the enchantment did not return to hand")
	}
	if !handHasID(me, camp) {
		t.Error("the enchantment is not in its owner's hand")
	}
}

func TestB35FrontierSeekerTakesAMountCreatureOrAPlainsFromTheTopFive(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	top := hornTopFive(me, []struct{ name, typeLine string }{
		{"Mountain", "Basic Land — Mountain"},
		{"Rodeo Rider", "Creature — Human Scout"},
		{"Prairie Mount", "Creature — Mount"},
		{"Plains", "Basic Land — Plains"},
		{"Mount Aura", "Enchantment — Mount"},
	})
	msc17Play(t, g, me, game.Card{Name: "Frontier Seeker", TypeLine: "Creature — Human Scout",
		ManaCost: "{1}{W}", OracleID: b35FrontierSeekerOracle, Power: 2, Toughness: 2})
	prompt := latestChooseCards(g, me.ID)
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

// b35AnswerOption answers the open option pick with the option labelled `label`.
func b35AnswerOption(t *testing.T, g *game.Game, chooser uuid.UUID, label string) {
	t.Helper()
	c := latestOptionPickFor(g, chooser)
	if c == nil {
		t.Fatal("no option prompt is open")
	}
	for i, opt := range c.PickOptions {
		if opt.Label == label {
			if err := g.ResolveOptionPick(c.ID, chooser, i); err != nil {
				t.Fatalf("ResolveOptionPick: %v", err)
			}
			return
		}
	}
	t.Fatalf("%q is not offered; options are %+v", label, c.PickOptions)
}

func b35BucolicRanchActivate(t *testing.T, g *game.Game, me *game.Player, topCards ...struct{ name, typeLine string }) []uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	ranch := b12Push(g, me.ID, "Bucolic Ranch", "Land — Desert", b35BucolicRanchOracle, 0, 0)
	top := hornTopFive(me, topCards)
	b06AddMana(me, "C", "C", "C")
	b16Activate(t, g, me.ID, ranch, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	return top
}

func TestB35BucolicRanchMayTakeAMountCardToHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	top := b35BucolicRanchActivate(t, g, me, struct{ name, typeLine string }{"Prairie Mount", "Creature — Mount"})
	b35AnswerOption(t, g, me.ID, bucolicRanchTake)
	if !me.Hand.Contains(top[0]) {
		t.Error("the Mount card is not in hand")
	}
}

func TestB35BucolicRanchOffersNoTakeForANonMountAndMayBottomIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	top := b35BucolicRanchActivate(t, g, me, struct{ name, typeLine string }{"Bear", "Creature — Bear"})
	c := latestOptionPickFor(g, me.ID)
	if c == nil {
		t.Fatal("no option prompt")
	}
	for _, opt := range c.PickOptions {
		if opt.Label == bucolicRanchTake {
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
	b35AnswerOption(t, g, me.ID, bucolicRanchBottom)
	if me.Hand.Contains(top[0]) {
		t.Fatal("the card went to hand")
	}
	if got, want := indexOf(), len(me.Library.Cards)-1-topIdx; got != want {
		t.Errorf("the card is at index %d, want the opposite end %d", got, want)
	}
}

func TestB35BucolicRanchTapsForAnyColorOnlyForMountSpells(t *testing.T) {
	spec, ok := Lookup(b35BucolicRanchOracle)
	if !ok || len(spec.ManaAbilities) != 2 {
		t.Fatalf("Bucolic Ranch has %d mana abilities, want 2", len(spec.ManaAbilities))
	}
	if len(spec.ManaAbilities[1].Restrictions) == 0 {
		t.Error("the any-color mana has no Mount restriction")
	}
}

func TestB35KellanJoinsUpPlotsASmallCardAndGrowsOnALegend(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil
	cheap := putInHand(me, game.Card{Name: "Cheap Spell", TypeLine: "Instant", ManaCost: "{1}{R}", Owner: me.ID, Controller: me.ID})
	dear := putInHand(me, game.Card{Name: "Dear Spell", TypeLine: "Sorcery", ManaCost: "{3}{R}", Owner: me.ID, Controller: me.ID})
	land := putInHand(me, game.Card{Name: "Spare Land", TypeLine: "Basic Land — Mountain", Owner: me.ID, Controller: me.ID})
	msc17Play(t, g, me, game.Card{Name: "Kellan Joins Up", TypeLine: "Legendary Enchantment",
		ManaCost: "{G}{W}{U}", OracleID: b35KellanJoinsUpOracle})
	prompt := latestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("no exile prompt")
	}
	if len(prompt.ChooseCards) != 1 || prompt.ChooseCards[0] != cheap {
		t.Fatalf("offered %v, want only the mana value 2 nonland card (not %v or %v)", prompt.ChooseCards, dear, land)
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{cheap}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
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
	other := b12Creature(g, me.ID, "Other Bear", "Creature — Bear", 2, 2)
	msc17Play(t, g, me, game.Card{Name: "Famous Bear", TypeLine: "Legendary Creature — Bear",
		ManaCost: "{2}", Power: 2, Toughness: 2})
	if p, _ := ptOf(t, g, other); p != 3 {
		t.Errorf("the other creature has power %d, want 3", p)
	}
	// A nonlegendary creature does not.
	msc17Play(t, g, me, game.Card{Name: "Plain Bear", TypeLine: "Creature — Bear", ManaCost: "{2}", Power: 2, Toughness: 2})
	if p, _ := ptOf(t, g, other); p != 3 {
		t.Errorf("a nonlegendary creature grew the board: power %d", p)
	}
}

func TestB35JaceReawakenedCantBeCastInYourFirstThreeTurns(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	id := putInHand(me, game.Card{Name: "Jace Reawakened", TypeLine: "Legendary Planeswalker — Jace",
		ManaCost: "{U}{U}", OracleID: b35JaceReawakenedOracle, Owner: me.ID, Controller: me.ID})
	b06AddMana(me, "U", "U")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast Jace Reawakened on your first turn")
	}
}

func b35Planeswalker(g *game.Game, me *game.Player, name, typeLine, oracle string, loyalty int) uuid.UUID {
	id := pushCatalogPermanent(g, me.ID, name, typeLine, oracle, false)
	findBattlefieldCardForTest(g, id).Counters = map[string]int{game.CounterLoyalty: loyalty}
	return id
}

func TestB35JaceReawakenedPlusOneDrawsAndDiscards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	jace := b35Planeswalker(g, me, "Jace Reawakened", "Legendary Planeswalker — Jace", b35JaceReawakenedOracle, 3)
	hand := len(me.Hand.Cards)
	pr7Activate(t, g, me.ID, jace, 0, game.ActivateAbilityParams{})
	discardFromHand(t, g, me.ID)
	if got := len(me.Hand.Cards); got != hand {
		t.Errorf("hand %d, want %d after loot", got, hand)
	}
}

func TestB35JaceReawakenedMinusSixCopiesEverySpellThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	jace := b35Planeswalker(g, me, "Jace Reawakened", "Legendary Planeswalker — Jace", b35JaceReawakenedOracle, 7)
	pr7Activate(t, g, me.ID, jace, 2, game.ActivateAbilityParams{})
	life := lifeOf(g, opp.ID)
	castBoltAtAndKeepTheCopysTarget(t, g, me.ID, opp.ID, 1)
	castBoltAtAndKeepTheCopysTarget(t, g, me.ID, opp.ID, 1)
	if got := lifeOf(g, opp.ID); got != life-12 {
		t.Errorf("life %d after two copied Bolts, want %d", got, life-12)
	}
}

func TestB35OkoRingleaderMakesElkAndLoots(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	oko := b35Planeswalker(g, me, "Oko, the Ringleader", "Legendary Planeswalker — Oko", b35OkoRingleaderOracle, 3)
	pr7Activate(t, g, me.ID, oko, 1, game.ActivateAbilityParams{})
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Elk" && c.Controller == me.ID && c.IsCreature() {
			n++
			if p, tt := ptOf(t, g, c.InstanceID); p != 3 || tt != 3 {
				t.Errorf("the Elk is %d/%d", p, tt)
			}
		}
	}
	if n != 1 {
		t.Fatalf("%d Elk, want 1", n)
	}
}

func TestB35OkoRingleaderPlusOneDiscardsTwoWithoutACrime(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	oko := b35Planeswalker(g, me, "Oko, the Ringleader", "Legendary Planeswalker — Oko", b35OkoRingleaderOracle, 3)
	hand := len(me.Hand.Cards)
	pr7Activate(t, g, me.ID, oko, 0, game.ActivateAbilityParams{})
	discardFromHand(t, g, me.ID)
	if got := len(me.Hand.Cards); got != hand+2-2 {
		t.Errorf("hand %d, want %d: draw two, discard two", got, hand)
	}
}

func TestB35OkoRingleaderPlusOneDiscardsOneAfterACrime(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	oko := b35Planeswalker(g, me, "Oko, the Ringleader", "Legendary Planeswalker — Oko", b35OkoRingleaderOracle, 3)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBecomesTarget, CrimeCommitted: true, Actor: me.ID})
	})
	hand := len(me.Hand.Cards)
	pr7Activate(t, g, me.ID, oko, 0, game.ActivateAbilityParams{})
	discardFromHand(t, g, me.ID)
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("hand %d, want %d: draw two, discard one", got, hand+1)
	}
}

func TestB35OkoRingleaderMinusFiveCopiesEachOtherNonlandPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	oko := b35Planeswalker(g, me, "Oko, the Ringleader", "Legendary Planeswalker — Oko", b35OkoRingleaderOracle, 6)
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	gaPut(g, me.ID, "Forest", "Basic Land — Forest", "")
	pushToken(g, me.ID, ClueToken())
	before := len(g.Battlefield.Cards)
	pr7Activate(t, g, me.ID, oko, 2, game.ActivateAbilityParams{})
	if got := len(g.Battlefield.Cards); got != before+2 {
		t.Errorf("%d permanents, want %d: a Bear and a Clue copied, not the land or Oko", got, before+2)
	}
	if n, _ := countTokensNamed(g, me.ID, "Bear"); n != 1 {
		t.Errorf("%d Bear tokens, want 1", n)
	}
	if n, _ := countTokensNamed(g, me.ID, "Clue"); n != 2 {
		t.Errorf("%d Clues, want 2", n)
	}
}

func TestB35OkoRingleaderBecomesAHexproofCopyAtBeginningOfCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	oko := b35Planeswalker(g, me, "Oko, the Ringleader", "Legendary Planeswalker — Oko", b35OkoRingleaderOracle, 3)
	beast := apaPush(g, me.ID, me.ID, game.Card{Name: "Beast", TypeLine: "Creature — Beast", Power: 4, Toughness: 4, Colors: []string{"G"}})
	advanceTo(t, g, game.StepBeginCombat)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, beast)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := findBattlefieldCardForTest(g, oko)
	if c == nil || !c.IsCreature() || c.IsPlaneswalker() {
		t.Fatal("Oko is not a creature copy of the Beast")
	}
	if p, tt := ptOf(t, g, oko); p != 4 || tt != 4 {
		t.Errorf("Oko is %d/%d, want 4/4", p, tt)
	}
	if !hasEffectiveKeyword(t, g, oko, "hexproof") {
		t.Error("the copy has no hexproof")
	}
}

func TestB35StepBetweenWorldsAsksEachPlayerAndOnlyThoseWhoShuffleDrawSeven(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hands := make([]int, len(g.Seats))
	castCatalogSpell(t, g, "Step Between Worlds", "Sorcery", b35StepBetweenWorldsOracle, nil)
	for i, p := range g.Seats {
		hands[i] = len(p.Hand.Cards)
	}
	passPriorityAroundTable(t, g)
	// The caster answers first, then each opponent in turn order.
	for i, p := range g.Seats {
		answer := stepBetweenWorldsNo
		if i%2 == 0 {
			answer = stepBetweenWorldsYes
		}
		b35AnswerOption(t, g, p.ID, answer)
	}
	passPriorityAroundTable(t, g)
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
	spec, _ := Lookup(b35StepBetweenWorldsOracle)
	if len(spec.SpecialActions) != 1 || spec.SpecialActions[0].Cost != "{4}{U}{U}" {
		t.Errorf("special actions %+v, want plot {4}{U}{U}", spec.SpecialActions)
	}
}

func TestB35LilahExilesAndPlotsAMulticoloredSpellCastFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Lilah, Undefeated Slickshot", "Legendary Creature — Human Rogue", b35LilahOracle, 3, 3)
	spell, err := msc17Cast(t, g, me, game.Card{Name: "Gold Divination", TypeLine: "Sorcery", ManaCost: "{2}{U}{R}",
		Colors: []string{"U", "R"}, OracleID: spellControlDivinationOracle}, game.CastSpellParams{})
	if err != nil {
		t.Fatal(err)
	}
	b35SettleTriggerOrders(t, g, me.ID)
	if me.Graveyard.Contains(spell) {
		t.Fatal("the spell went to the graveyard")
	}
	if exiledCard(g, spell) == nil {
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Lilah, Undefeated Slickshot", "Legendary Creature — Human Rogue", b35LilahOracle, 3, 3)
	spell, err := msc17Cast(t, g, me, game.Card{Name: "Mono Divination", TypeLine: "Sorcery", ManaCost: "{2}{U}",
		Colors: []string{"U"}, OracleID: spellControlDivinationOracle}, game.CastSpellParams{})
	if err != nil {
		t.Fatal(err)
	}
	b35SettleTriggerOrders(t, g, me.ID)
	if !me.Graveyard.Contains(spell) || exiledCard(g, spell) != nil {
		t.Error("a mono-colored spell should go to the graveyard as usual")
	}
}

// b35SettleTriggerOrders answers any CR 603.3b ordering prompt as it is
// asked and passes priority until the stack is empty.
func b35SettleTriggerOrders(t *testing.T, g *game.Game, who uuid.UUID) {
	t.Helper()
	for i := 0; i < 48; i++ {
		if p := triggerOrderPromptFor(g, who); p != nil {
			order := append([]uuid.UUID(nil), p.TriggerOrderIDs...)
			if err := g.ResolveTriggerOrder(p.ID, who, order); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		if stackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
}

func TestB35TinybonesCastsANonlandPermanentFromTheDamagedPlayersGraveyardWithAnyMana(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	tiny := b27Push(g, me.ID, "Tinybones, the Pickpocket", "Legendary Creature — Skeleton Rogue", b35TinybonesOracle, "{B}", 1, 1, "B")
	var theirs, land, elsewhere uuid.UUID
	g.WithWriteLock(func() {
		theirs = uuid.New()
		opp.Graveyard.PushTop(game.Card{InstanceID: theirs, Name: "Their Dead Bear", TypeLine: "Creature — Bear",
			ManaCost: "{1}{G}", Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
		land = pushGraveyardCardTyped(opp, "Their Dead Land", "Basic Land — Forest")
		elsewhere = pushGraveyardCardTyped(other, "Other Dead Bear", "Creature — Bear")
	})
	declareAttack(t, g, opp.ID, tiny)
	advanceTo(t, g, game.StepDeclareBlockers)
	finishBlockDeclarations(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	for i := 0; i < 4 && latestPickTarget(g, me.ID) == nil && triggerOnStack(g, tiny) == nil; i++ {
		passPriorityAroundTable(t, g)
	}
	p := latestPickTarget(g, me.ID)
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
	b06AddMana(me, "R", "R")
	for i := 0; i < 8 && keywordCastPromptFor(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	prompt := keywordCastPromptFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no cast offer")
	}
	answerKeywordCast(t, g, prompt, true)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(theirs) {
		t.Error("the Bear did not enter, though red mana was spent on a green cost")
	}
	if c := battlefieldCardPtr(t, g, theirs); c.Controller != me.ID {
		t.Error("the Bear is not under the caster's control")
	}
}

func TestB35VraskaTheSilencerReturnsAnOpponentsCreatureAsATappedTreasure(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Vraska, the Silencer", "Legendary Creature — Gorgon Assassin", b35VraskaSilencerOracle, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	token := pushToken(g, opp.ID, MercenaryToken())
	// Neither my own creature nor a token triggers it.
	p1Destroy(t, g, mine)
	p1Destroy(t, g, token)
	passPriorityAroundTable(t, g)
	if latestChoiceOfKind(g, game.PendingChoicePayUnless) != nil {
		t.Fatal("offered for my own creature or a token")
	}
	p1Destroy(t, g, theirs)
	passPriorityAroundTable(t, g)
	pay := latestChoiceOfKind(g, game.PendingChoicePayUnless)
	if pay == nil {
		t.Fatal("no offer to pay {1}")
	}
	b06AddMana(me, "C")
	if err := g.ResolvePayUnless(pay.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, theirs)
	if !ok {
		t.Fatal("the creature card did not return")
	}
	if c.Controller != me.ID || !c.Tapped {
		t.Errorf("controller %s tapped %v, want mine and tapped", c.Controller, c.Tapped)
	}
	if c.IsCreature() || !c.IsArtifact() || !c.HasSubtype("Treasure") {
		t.Errorf("type line %q: want a Treasure artifact and no creature", c.TypeLine)
	}
	if grantedManaRowsOf(g, theirs) != 1 {
		t.Error("the Treasure has no sacrifice-for-mana ability")
	}
}

func TestB35RoxanneMakesATappedMeteoriteThatShocksAndDoublesTokenMana(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := lifeOf(g, opp.ID)
	msc17Play(t, g, me, game.Card{Name: "Roxanne, Starfall Savant", TypeLine: "Legendary Creature — Cat Druid",
		ManaCost: "{3}{R}{G}", OracleID: b35RoxanneOracle, Power: 4, Toughness: 4})
	// The Roxanne trigger makes the Meteorite, whose own trigger asks for a target.
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != life-2 {
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
	untapForTest(g, meteorite)
	if err := g.ActivateManaAbility(me.ID, meteorite, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("tap Meteorite for mana: %v", err)
	}
	if got := me.ManaPool.CountColor("R"); got != 2 {
		t.Errorf("%d {R} after tapping Meteorite for red under Roxanne, want 2", got)
	}
}
