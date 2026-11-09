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
