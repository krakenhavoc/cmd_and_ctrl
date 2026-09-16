package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch34_test.go — card-level coverage for the card-coverage
// roadmap's batch 34 (#397, `edhrec_rank` 3550–3649): the "no new
// machinery" group. One test per observable behaviour, driven through
// a real cast, activation or attack rather than by calling primitives.

const (
	b34InvigoratingSurgeOracle    = "5daac63a-1534-4194-8cb7-506508e364f4"
	b34ReveillarkOracle           = "1be13ede-98f8-497e-800c-03e5802932b3"
	b34YavimayaElderOracle        = "7fd4c452-07f2-492c-9c78-d1c6362d9eec"
	b34VitalizeOracle             = "cd3226c6-bd47-4246-8ab3-01316c5d1809"
	b34FiendishDuoOracle          = "ab0dfae5-b9d4-417b-8a0d-2525ae3a73b9"
	b34ArmorcraftJudgeOracle      = "d7f49243-a96e-499f-b2d7-8e9842432420"
	b34OmnathLocusOfTheRoilOracle = "ffd34151-457a-4a93-82db-24b18319a06b"
	b34BeetlebackChiefOracle      = "f5c5f64c-6911-430c-a825-b32b96d39c7d"
	b34CoriMountainMonasteryOrcl  = "35c60b66-8c85-432e-90fe-99c19d21ed15"
	b34BristlingBackwoodsOracle   = "9cbc9f83-8979-42a5-a466-a8d89c8e6de8"
	b34SporemoundOracle           = "1be56a3d-a6c0-4b65-ae71-3d90ceefc6c0"
	b34GuildArtisanOracle         = "aceac269-5100-428a-a4a3-bac57031e30b"
	b34LordOfTheUndeadOracle      = "7714af0e-41d9-4609-967f-27233b46055f"
	b34AuroralProcessionOracle    = "511d3de9-076b-4811-bddf-00623f98993f"
	b34KioraTheRisingTideOracle   = "5d748bce-dff8-46fa-a3d1-633863b7bbff"
	b34RegalForceOracle           = "e264ffe3-0252-49d6-b990-dbb3654325a5"
	b34TurbulentFenOracle         = "114dd40d-5ad8-4913-a08f-572b9521eb5b"
	b34OnWingsOfGoldOracle        = "3b6f608d-f27d-44cd-b0a1-0e1e1aa1c98a"
	b34DragonbornChampionOracle   = "9e51ccf9-d55d-4231-8ff5-ef3ca1485c8d"
	b34ArbaazMirOracle            = "a383ef16-1af8-4b3a-956c-c10a93768617"
	b34MaelstromPulseOracle       = "95ce305f-34bc-4d6d-b7ba-ffd4b2a25336"
	b34SeizanOracle               = "04d0d20f-720e-4cb6-a3ad-ea2b57bb7efa"
	b34TormodOracle               = "6dc0150d-7145-41e2-bd3d-2564d9d32301"
	b34MapTheFrontierOracle       = "11a46cb6-ab01-4630-9541-782db2ef3b91"
	b34RapaciousDragonOracle      = "0944ec2e-1dd9-459f-8f1d-667242cf52fe"
	b34CrosswayTroublemakersOrcl  = "1a362e4d-6c02-4b67-ab63-c6622e505195"
	b34DeathsproutOracle          = "793b0c73-601c-41dc-b49a-45fee4970d52"
	b34ProsperityOracle           = "c586312d-d04a-4bfb-bbb2-b41186ca178e"
	b34HobgoblinBanditLordOracle  = "43fea418-db6f-4953-89d1-6873b2ca41a8"
	b34ElanorGardnerOracle        = "ad4c39d6-a8b3-4dd9-816d-09761fc9651d"
	b34VillageBellRingerOracle    = "d7b50175-b473-452a-bb30-19a71c42e649"
	b34PollutedBondsOracle        = "13b33a41-c22b-4e7a-8323-abadf7f80961"
)

// b34Creature seeds a non-catalog creature with a type line and
// colours, able to attack and tap.
func b34Creature(g *game.Game, owner uuid.UUID, name, typeLine string, power, toughness int, colors ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Colors: colors,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// b34Token pushes a creature token under controller, able to attack.
func b34Token(g *game.Game, controller uuid.UUID, name, typeLine string, power, toughness int) uuid.UUID {
	return pushToken(g, controller, game.Card{
		Name: name, TypeLine: typeLine, Power: power, Toughness: toughness,
	})
}

// b34Lands seeds n vanilla lands under owner.
func b34Lands(g *game.Game, owner uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		b12Permanent(g, owner, "Wastes", "Basic Land")
	}
}

// b34Changeling seeds a creature with the changeling keyword — every
// creature type at once.
func b34Changeling(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Shapeshifter",
		Keywords: []string{game.KeywordChangeling},
		Power:    1, Toughness: 1, Owner: owner, Controller: owner,
	})
}

// b34LandEntersByEffect puts a land onto the battlefield under `p`
// through the graveyard-return path — an effect's entry, with a real
// EventETB, rather than a land play.
func b34LandEntersByEffect(t *testing.T, g *game.Game, p *game.Player, name, typeLine string) uuid.UUID {
	t.Helper()
	id := b17GraveyardCard(p, name, typeLine, "")
	g.WithWriteLock(func() {
		if err := g.ReturnFromGraveyardForEffect(id, game.ZoneBattlefield); err != nil {
			t.Fatalf("ReturnFromGraveyardForEffect: %v", err)
		}
	})
	return id
}

// b34ExileFromGraveyard exiles a graveyard card through the effect
// path, leaving whatever it triggered pending.
func b34ExileFromGraveyard(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(id); err != nil {
			t.Fatalf("ExileCardForEffect: %v", err)
		}
	})
}

// b34PlayLandAs seeds a land in `seat`'s hand and plays it from that
// seat's precombat main phase.
func b34PlayLandAs(t *testing.T, g *game.Game, seat int, name, typeLine string) uuid.UUID {
	t.Helper()
	advanceToMainOf(t, g, seat)
	p := g.Seats[seat]
	id := handCard(p, name, typeLine)
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("play %s: %v", name, err)
	}
	return id
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Bloodgift
// Demon was already on main from the life-for-cards work (#507), so
// the table is the issue's 42 minus it and minus the nine declared
// skips.
func TestBatch34CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b34InvigoratingSurgeOracle:    "Invigorating Surge",
		b34ReveillarkOracle:           "Reveillark",
		b34YavimayaElderOracle:        "Yavimaya Elder",
		b34VitalizeOracle:             "Vitalize",
		b34FiendishDuoOracle:          "Fiendish Duo",
		b34ArmorcraftJudgeOracle:      "Armorcraft Judge",
		b34OmnathLocusOfTheRoilOracle: "Omnath, Locus of the Roil",
		b34BeetlebackChiefOracle:      "Beetleback Chief",
		b34CoriMountainMonasteryOrcl:  "Cori Mountain Monastery",
		b34BristlingBackwoodsOracle:   "Bristling Backwoods",
		b34SporemoundOracle:           "Sporemound",
		b34GuildArtisanOracle:         "Guild Artisan",
		b34LordOfTheUndeadOracle:      "Lord of the Undead",
		b34AuroralProcessionOracle:    "Auroral Procession",
		b34KioraTheRisingTideOracle:   "Kiora, the Rising Tide",
		b34RegalForceOracle:           "Regal Force",
		b34TurbulentFenOracle:         "Turbulent Fen",
		b34OnWingsOfGoldOracle:        "On Wings of Gold",
		b34DragonbornChampionOracle:   "Dragonborn Champion",
		b34ArbaazMirOracle:            "Arbaaz Mir",
		b34MaelstromPulseOracle:       "Maelstrom Pulse",
		b34SeizanOracle:               "Seizan, Perverter of Truth",
		b34TormodOracle:               "Tormod, the Desecrator",
		b34MapTheFrontierOracle:       "Map the Frontier",
		b34RapaciousDragonOracle:      "Rapacious Dragon",
		b34CrosswayTroublemakersOrcl:  "Crossway Troublemakers",
		b34DeathsproutOracle:          "Deathsprout",
		b34ProsperityOracle:           "Prosperity",
		b34HobgoblinBanditLordOracle:  "Hobgoblin Bandit Lord",
		b34ElanorGardnerOracle:        "Elanor Gardner",
		b34VillageBellRingerOracle:    "Village Bell-Ringer",
		b34PollutedBondsOracle:        "Polluted Bonds",
	}
	if len(want) != 32 {
		t.Fatalf("the batch ships 32 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
	}
}

// --- spells ---------------------------------------------------------

func TestB34InvigoratingSurgeAddsOneThenDoubles(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, "+1/+1", 3) })
	castCatalogSpell(t, g, "Invigorating Surge", "Instant", b34InvigoratingSurgeOracle, b16TargetCard(bear))
	passPriorityAroundTable(t, g)
	if got := counterCount(g, bear, "+1/+1"); got != 8 {
		t.Errorf("three counters, plus one, doubled: %d, want 8", got)
	}
	if p := currentPower(t, g, bear); p != 10 {
		t.Errorf("the Bear is a %d/x, want 10", p)
	}
	bare := b12Creature(g, me.ID, "Bare Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Invigorating Surge", "Instant", b34InvigoratingSurgeOracle, b16TargetCard(bare))
	passPriorityAroundTable(t, g)
	if got := counterCount(g, bare, "+1/+1"); got != 2 {
		t.Errorf("no counters, plus one, doubled: %d, want 2", got)
	}
	surge := b13HandCard(me, "Invigorating Surge", "Instant", b34InvigoratingSurgeOracle)
	if err := g.CastSpell(me.ID, surge, game.CastSpellParams{Targets: b16TargetCard(theirs)}); err == nil {
		t.Error("an opponent's creature is not \"target creature you control\"")
	}
}

func TestB34VitalizeUntapsOnlyYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	land := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{a, b, land, theirs} {
		b16Tap(g, id)
	}
	castCatalogSpell(t, g, "Vitalize", "Instant", b34VitalizeOracle, nil)
	passPriorityAroundTable(t, g)
	if b16Tapped(t, g, a) || b16Tapped(t, g, b) {
		t.Error("your creatures untap")
	}
	if !b16Tapped(t, g, land) {
		t.Error("a land is not a creature")
	}
	if !b16Tapped(t, g, theirs) {
		t.Error("an opponent's creature stays tapped")
	}
}

func TestB34AuroralProcessionReturnsAnyCardFromYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := b17GraveyardCard(me, "Dead Forest", "Basic Land — Forest", "")
	theirs := b17GraveyardCard(opp, "Their Bolt", "Instant", "{R}")
	castCatalogSpell(t, g, "Auroral Procession", "Instant", b34AuroralProcessionOracle, b16TargetCard(land))
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(land) {
		t.Error("the land card is returned to hand")
	}
	procession := b13HandCard(me, "Auroral Procession", "Instant", b34AuroralProcessionOracle)
	if err := g.CastSpell(me.ID, procession, game.CastSpellParams{Targets: b16TargetCard(theirs)}); err == nil {
		t.Error("an opponent's graveyard is not yours")
	}
}

func TestB34MaelstromPulseDestroysEveryPermanentSharingTheName(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	a := b34Token(g, opp.ID, "Goblin", "Token Creature — Goblin", 1, 1)
	b := b34Token(g, opp.ID, "Goblin", "Token Creature — Goblin", 1, 1)
	c := b34Token(g, other.ID, "Goblin", "Token Creature — Goblin", 1, 1)
	mine := b34Token(g, me.ID, "Goblin", "Token Creature — Goblin", 1, 1)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	land := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	pulse := b13HandCard(me, "Maelstrom Pulse", "Sorcery", b34MaelstromPulseOracle)
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, pulse, game.CastSpellParams{Targets: b16TargetCard(land)}); err == nil {
		t.Fatal("a land is not \"target nonland permanent\"")
	}
	castCatalogSpell(t, g, "Maelstrom Pulse", "Sorcery", b34MaelstromPulseOracle, b16TargetCard(a))
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c, mine} {
		if g.Battlefield.Contains(id) {
			t.Error("every Goblin at the table — the target, its siblings, yours — is destroyed")
		}
	}
	if !g.Battlefield.Contains(bear) || !g.Battlefield.Contains(land) {
		t.Error("nothing with another name is touched")
	}
}

func TestB34MaelstromPulseSparesAnIndestructibleNamesake(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, opp.ID, "Twin", "Creature — Bear", 2, 2)
	tough := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Twin", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Keywords: []string{"indestructible"}, Owner: opp.ID, Controller: opp.ID,
	})
	castCatalogSpell(t, g, "Maelstrom Pulse", "Sorcery", b34MaelstromPulseOracle, b16TargetCard(a))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) {
		t.Error("the target is destroyed")
	}
	if !g.Battlefield.Contains(tough) {
		t.Error("the single-target verb honours indestructible on the namesake")
	}
	_ = me
}

func TestB34DeathsproutDestroysAndRampsATappedBasic(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	ids := seedSearchLibrary(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Snow-Covered Forest", TypeLine: "Basic Snow Land — Forest"},
		game.Card{Name: "Swamp", TypeLine: "Basic Land — Swamp"},
		game.Card{Name: "Overgrown Tomb", TypeLine: "Land — Swamp Forest"},
	)
	castCatalogSpell(t, g, "Deathsprout", "Instant", b34DeathsproutOracle, b16TargetCard(theirs))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Fatal("the creature is destroyed")
	}
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if len(c.SearchCards) != 2 || searchOptionNamed(g, c, "Overgrown Tomb") != uuid.Nil {
		t.Errorf("the two basics are offered, the shockland is not: %d offered", len(c.SearchCards))
	}
	answerSearchByID(t, g, me.ID, ids[2])
	if !g.Battlefield.Contains(ids[2]) || !b16Tapped(t, g, ids[2]) {
		t.Error("the chosen basic enters the battlefield tapped")
	}
}

func TestB34MapTheFrontierFetchesBasicsAndDesertsTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Desert of the Glorified", TypeLine: "Land — Desert"},
		game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Overgrown Tomb", TypeLine: "Land — Swamp Forest"},
	)
	castCatalogSpell(t, g, "Map the Frontier", "Sorcery", b34MapTheFrontierOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if c.SearchMax != 2 || len(c.SearchCards) != 3 {
		t.Errorf("two basics and a Desert offered, up to two taken: %d offered, max %d", len(c.SearchCards), c.SearchMax)
	}
	if searchOptionNamed(g, c, "Overgrown Tomb") != uuid.Nil || searchOptionNamed(g, c, "Bear") != uuid.Nil {
		t.Error("a nonbasic non-Desert and a creature are not offered")
	}
	answerSearchByID(t, g, me.ID, ids[0], ids[1])
	for _, id := range []uuid.UUID{ids[0], ids[1]} {
		if !g.Battlefield.Contains(id) {
			t.Fatal("both chosen lands are put onto the battlefield")
		}
		if !b16Tapped(t, g, id) {
			t.Error("they enter tapped")
		}
	}
	if me.Library.Size() != 3 {
		t.Errorf("library %d, want 3", me.Library.Size())
	}
}

func TestB34ProsperityEachPlayerDrawsX(t *testing.T) {
	g := newCatalogGame(t)
	hands := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		hands[i] = p.Hand.Size()
	}
	castXSpell(t, g, "Prosperity", "Sorcery", b34ProsperityOracle, "{X}{U}", 2, nil)
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.Hand.Size() != hands[i]+2 {
			t.Errorf("seat %d hand %d, want %d", i, p.Hand.Size(), hands[i]+2)
		}
	}
}

// --- creatures ------------------------------------------------------

func TestB34ReveillarkReturnsUpToTwoSmallCreaturesOnAnyLeave(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lark := b12Push(g, me.ID, "Reveillark", "Creature — Elemental", b34ReveillarkOracle, 4, 3)
	small := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: small, Name: "Dead Elf", TypeLine: "Creature — Elf", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	two := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: two, Name: "Dead Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	big := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: big, Name: "Dead Titan", TypeLine: "Creature — Giant", Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID})
	theirs := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: theirs, Name: "Their Elf", TypeLine: "Creature — Elf", Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	if !hasEffectiveKeyword(t, g, lark, "flying") {
		t.Error("printed flying did not reach the effective abilities")
	}
	// Bounce, not death: "leaves the battlefield".
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(lark) })
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if !hasID(p.PickTargetCards, small) || !hasID(p.PickTargetCards, two) {
		t.Error("creature cards with power 2 or less are offered")
	}
	if hasID(p.PickTargetCards, big) || hasID(p.PickTargetCards, theirs) {
		t.Error("a 6/6 and an opponent's card are not offered")
	}
	if p.PickTargetMin != 0 || p.PickTargetMax != 2 {
		t.Errorf("\"up to two\": min %d max %d", p.PickTargetMin, p.PickTargetMax)
	}
	b17PickCards(t, g, me.ID, small, two)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(small) || !g.Battlefield.Contains(two) {
		t.Error("both chosen creatures return to the battlefield")
	}
	if !me.Hand.Contains(lark) {
		t.Error("the bounced Lark is in hand")
	}
}

func TestB34ReveillarkEvokedIsSacrificedAndStillReturns(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	small := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: small, Name: "Dead Elf", TypeLine: "Creature — Elf", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	lark := castWithAltCost(t, g, "Reveillark", "Creature — Elemental", b34ReveillarkOracle, "evoke")
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, small)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(lark) || !me.Graveyard.Contains(lark) {
		t.Error("the evoked Lark is sacrificed on entry")
	}
	if !g.Battlefield.Contains(small) {
		t.Error("the leave trigger returns the chosen creature")
	}
}

func TestB34YavimayaElderSacrificesForLandsThenACard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	elder := b12Push(g, me.ID, "Yavimaya Elder", "Creature — Human Druid", b34YavimayaElderOracle, 2, 1)
	ids := seedSearchLibrary(me,
		game.Card{Name: "Draw Me", TypeLine: "Sorcery"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Snow-Covered Swamp", TypeLine: "Basic Snow Land — Swamp"},
		game.Card{Name: "Overgrown Tomb", TypeLine: "Land — Swamp Forest"},
		game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"},
		game.Card{Name: "Filler", TypeLine: "Sorcery"},
	)
	advanceToMain(t, g)
	hand := me.Hand.Size()
	b06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, elder, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(elder) {
		t.Fatal("the sacrifice is the cost")
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	// The dies trigger sits above the ability: it resolves first and
	// asks which basics; the draw waits underneath.
	for i := 0; i < 8 && searchChoiceFor(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the dies trigger resolves first and asks for the lands")
	}
	if c.SearchMax != 2 || len(c.SearchCards) != 3 || searchOptionNamed(g, c, "Overgrown Tomb") != uuid.Nil {
		t.Errorf("the three basics offered, up to two taken: %d offered, max %d", len(c.SearchCards), c.SearchMax)
	}
	if me.Hand.Size() != hand {
		t.Error("the ability's draw waits below the trigger")
	}
	answerSearchByID(t, g, me.ID, ids[1], ids[2])
	if !me.Hand.Contains(ids[1]) || !me.Hand.Contains(ids[2]) {
		t.Error("both basics go to hand")
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+3 {
		t.Errorf("two lands and a card: hand %d → %d", hand, me.Hand.Size())
	}
}

func TestB34YavimayaElderDeclinedSearchesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	elder := b12Push(g, me.ID, "Yavimaya Elder", "Creature — Human Druid", b34YavimayaElderOracle, 2, 1)
	b27Kill(g, elder)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if searchChoiceFor(g, me.ID) != nil {
		t.Error("declining the trigger searches nothing")
	}
}

func TestB34FiendishDuoDoublesDamageToOpponentsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	duo := b12Push(g, me.ID, "Fiendish Duo", "Creature — Devil", b34FiendishDuoOracle, 5, 5)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 5)
	if !hasEffectiveKeyword(t, g, duo, "first strike") {
		t.Error("printed first strike did not reach the effective abilities")
	}
	// A Bolt at an opponent is doubled; one at you is not.
	life := opp.Life
	b07BoltPlayer(t, g, opp.ID)
	if opp.Life != life-6 {
		t.Errorf("a Bolt at an opponent: %d → %d, want -6", life, opp.Life)
	}
	mine := me.Life
	batch01OpponentCasts(t, g, opp, "Lightning Bolt", lightningBoltOracle, "{R}", b16TargetPlayer(me.ID))
	passPriorityAroundTable(t, g)
	if me.Life != mine-3 {
		t.Errorf("a Bolt at the Duo's controller: %d → %d, want -3", mine, me.Life)
	}
	// Damage to an opponent's creature is untouched.
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetCard(theirs))
	passPriorityAroundTable(t, g)
	if got := damageMarkedOn(g, theirs); got != 3 {
		t.Errorf("a Bolt at a creature marks %d, want 3", got)
	}
	// Combat damage to an opponent is doubled too.
	life = opp.Life
	attackWith(t, g, opp.ID, bear)
	if opp.Life != life-4 {
		t.Errorf("a 2/2 connecting: %d → %d, want -4", life, opp.Life)
	}
}

func TestB34ArmorcraftJudgeDrawsPerCounteredCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Bare Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(a, "+1/+1", 1)
		_ = g.AddCounterForEffect(b, "+1/+1", 3)
		_ = g.AddCounterForEffect(theirs, "+1/+1", 1)
	})
	castCatalogSpell(t, g, "Armorcraft Judge", "Creature — Elf Artificer", b34ArmorcraftJudgeOracle, nil)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("two of your creatures carry counters: drew %d, want 2", me.Hand.Size()-hand)
	}
}

func TestB34RegalForceDrawsPerGreenCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b34Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1, "G")
	b34Creature(g, me.ID, "Gold Elf", "Creature — Elf", 1, 1, "G", "W")
	b34Creature(g, me.ID, "Bird", "Creature — Bird", 1, 1, "U")
	b12Creature(g, me.ID, "Colourless", "Artifact Creature — Golem", 3, 3)
	b34Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1, "G")
	id := handCardFull(me, "Regal Force", "Creature — Elemental", "{4}{G}{G}{G}", b34RegalForceOracle, []string{"G"})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+3 {
		t.Errorf("two green creatures and the Force itself: drew %d, want 3", me.Hand.Size()-hand)
	}
}

func TestB34OmnathEntersForDamagePerElementalAndGrowsOnLandfall(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	other := b12Creature(g, me.ID, "Flamekin", "Creature — Elemental Shaman", 2, 2)
	b34Changeling(g, me.ID, "Changeling")
	theirs := b12Creature(g, opp.ID, "Their Elemental", "Creature — Elemental", 2, 2)
	life := opp.Life
	omnath := castCatalogSpell(t, g, "Omnath, Locus of the Roil", "Legendary Creature — Elemental", b34OmnathLocusOfTheRoilOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("Omnath, the Flamekin and the changeling are three Elementals: %d → %d, want -3", life, opp.Life)
	}
	// Landfall: a counter on target Elemental you control; seven lands
	// is not eight.
	b34Lands(g, me.ID, 6)
	hand := me.Hand.Size()
	playLandFromHand(t, g, "Forest", "")
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if !hasID(p.PickTargetCards, omnath) || !hasID(p.PickTargetCards, other) {
		t.Error("your Elementals are offered")
	}
	if hasID(p.PickTargetCards, theirs) {
		t.Error("an opponent's Elemental is not \"you control\"")
	}
	pickCard(t, g, me.ID, other)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, other, "+1/+1"); got != 1 {
		t.Errorf("the chosen Elemental gets a counter: %d", got)
	}
	if me.Hand.Size() != hand {
		t.Error("seven lands: no card")
	}
	// The eighth land — put onto the battlefield by an effect — draws.
	b34LandEntersByEffect(t, g, me, "Late Forest", "Basic Land — Forest")
	b04WaitForPick(t, g, me.ID)
	hand = me.Hand.Size()
	pickCard(t, g, me.ID, omnath)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, omnath, "+1/+1"); got != 1 {
		t.Errorf("Omnath gets the counter this time: %d", got)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("eight or more lands: drew %d, want 1", me.Hand.Size()-hand)
	}
}

func TestB34BeetlebackChiefBringsTwoGoblins(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Beetleback Chief", "Creature — Goblin Warrior", b34BeetlebackChiefOracle, nil)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Goblin"); n != 2 {
		t.Errorf("%d Goblins, want 2", n)
	}
}

func TestB34RapaciousDragonBringsTwoTreasures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dragon := castCatalogSpell(t, g, "Rapacious Dragon", "Creature — Dragon", b34RapaciousDragonOracle, nil)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Treasure"); n != 2 {
		t.Errorf("%d Treasures, want 2", n)
	}
	if !hasEffectiveKeyword(t, g, dragon, "flying") {
		t.Error("printed flying did not reach the effective abilities")
	}
}

func TestB34SporemoundMakesASaprolingPerLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Sporemound", "Creature — Fungus", b34SporemoundOracle, 3, 3)
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Saproling"); n != 1 {
		t.Fatalf("%d Saprolings after a land drop, want 1", n)
	}
	b34LandEntersByEffect(t, g, me, "Fetched Forest", "Basic Land — Forest")
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Saproling"); n != 2 {
		t.Errorf("%d Saprolings after a second land, want 2", n)
	}
	sap := battlefieldIDNamed(g, me.ID, "Saproling")
	if p := effectivePower(t, g, sap); p != 1 {
		t.Errorf("the Saproling is a %d/x, want 1", p)
	}
}

func TestB34VillageBellRingerFlashesInAndUntapsYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b16Tap(g, a)
	b16Tap(g, theirs)
	ringer := castCatalogSpell(t, g, "Village Bell-Ringer", "Creature — Human Scout", b34VillageBellRingerOracle, nil)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, ringer, "flash") {
		t.Error("printed flash did not reach the effective abilities")
	}
	if b16Tapped(t, g, a) {
		t.Error("your creature untaps")
	}
	if !b16Tapped(t, g, theirs) {
		t.Error("an opponent's creature does not")
	}
}

func TestB34ArbaazMirPingsOnHimselfAndOnNontokenHistoricPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := lifeOfOpponents(g)
	mine := me.Life
	castCatalogSpell(t, g, "Arbaaz Mir", "Legendary Creature — Human Assassin", b34ArbaazMirOracle, nil)
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats[1:] {
		if p.Life != before[i]-1 {
			t.Errorf("Arbaaz entering: opponent %d lost %d, want 1", i+1, before[i]-p.Life)
		}
	}
	if me.Life != mine+1 {
		t.Errorf("Arbaaz entering: life %d → %d, want +1", mine, me.Life)
	}
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if got := lifeOfOpponents(g); got[0] != before[0]-1 {
		t.Error("a Bear is not historic")
	}
	castCatalogSpell(t, g, "Signet", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Legend", "Legendary Creature — Human", "", nil)
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Saga", "Enchantment — Saga", "", nil)
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats[1:] {
		if p.Life != before[i]-4 {
			t.Errorf("an artifact, a legend and a Saga: opponent %d lost %d, want 4", i+1, before[i]-p.Life)
		}
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TreasureToken(), 1) })
	passPriorityAroundTable(t, g)
	if got := lifeOfOpponents(g); got[0] != before[0]-4 {
		t.Error("a Treasure is an artifact but a token: \"nontoken\"")
	}
	if me.Life != mine+4 {
		t.Errorf("life %d → %d, want +4", mine, me.Life)
	}
}

func TestB34DragonbornChampionDrawsOnFiveDamageInOneHit(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	champ := b12Push(g, me.ID, "Dragonborn Champion", "Creature — Dragon Warrior", b34DragonbornChampionOracle, 5, 3)
	small := b12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	if !hasEffectiveKeyword(t, g, champ, "trample") {
		t.Error("printed trample did not reach the effective abilities")
	}
	hand := me.Hand.Size()
	b07BoltPlayer(t, g, opp.ID)
	if me.Hand.Size() != hand {
		t.Error("three damage is not five")
	}
	attackWith(t, g, opp.ID, champ, small)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("the Champion connects for 5, the Bear for 3: drew %d, want 1", me.Hand.Size()-hand)
	}
}

func TestB34DragonbornChampionIgnoresOpponentsSourcesAndCreatureDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	b12Push(g, me.ID, "Dragonborn Champion", "Creature — Dragon Warrior", b34DragonbornChampionOracle, 5, 3)
	big := b12Creature(g, me.ID, "Wurm", "Creature — Wurm", 6, 6)
	theirs := b12Creature(g, opp.ID, "Their Wurm", "Creature — Wurm", 6, 6)
	hand := me.Hand.Size()
	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(theirs, other.ID, 6)
		_ = g.DealDamageToCreatureForEffect(big, theirs, 6)
	})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Error("an opponent's source, and damage to a creature, do not count")
	}
}

func TestB34KioraLootsOnEntryAndMakesTheScionAtThreshold(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kiora := castCatalogSpell(t, g, "Kiora, the Rising Tide", "Legendary Creature — Merfolk Noble", b34KioraTheRisingTideOracle, nil)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("drew %d, want 2", me.Hand.Size()-hand)
	}
	if g.DiscardPending[me.ID] != 2 {
		t.Errorf("owes %d discards, want 2", g.DiscardPending[me.ID])
	}
	delete(g.DiscardPending, me.ID)
	// Six cards in the graveyard: no threshold, no prompt. Kiora
	// attacks on the next turn, once the sickness has worn off.
	for i := 0; i < 6; i++ {
		pushGraveyardCardForTest(me, "Filler")
	}
	advanceToMainOf(t, g, 1)
	advanceToMainOf(t, g, 0)
	declareAttack(t, g, opp.ID, kiora)
	if latestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("six cards is not threshold")
	}
	passPriorityAroundTable(t, g)
	advanceToMainOf(t, g, 0)
	pushGraveyardCardForTest(me, "Seventh")
	declareAttack(t, g, opp.ID, kiora)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	scion := battlefieldIDNamed(g, me.ID, "Scion of the Deep")
	if scion == uuid.Nil {
		t.Fatal("threshold: the Scion is created")
	}
	if p, tt := effectivePower(t, g, scion), effectiveToughness(t, g, scion); p != 8 || tt != 8 {
		t.Errorf("the Scion is %d/%d, want 8/8", p, tt)
	}
	if c, _ := battlefieldCard(g, scion); !isLegendary(&c) {
		t.Error("the Scion is legendary")
	}
}

func TestB34CrosswayTroublemakersAttackingVampiresAndTheLifeDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	makers := b12Push(g, me.ID, "Crossway Troublemakers", "Creature — Vampire", b34CrosswayTroublemakersOrcl, 5, 5)
	vamp := b12Creature(g, me.ID, "Vampire", "Creature — Vampire", 2, 2)
	home := b12Creature(g, me.ID, "Home Vampire", "Creature — Vampire", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	life := me.Life
	declareAttack(t, g, opp.ID, makers, vamp, bear)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{makers, vamp} {
		if !hasEffectiveKeyword(t, g, id, "deathtouch") || !hasEffectiveKeyword(t, g, id, "lifelink") {
			t.Error("an attacking Vampire you control has deathtouch and lifelink")
		}
	}
	for _, id := range []uuid.UUID{home, bear} {
		if hasEffectiveKeyword(t, g, id, "deathtouch") || hasEffectiveKeyword(t, g, id, "lifelink") {
			t.Error("a Vampire that stayed home and a non-Vampire do not")
		}
	}
	advanceTo(t, g, game.StepCombatDamage)
	if me.Life != life+7 {
		t.Errorf("lifelink on 5 and 2: life %d → %d, want +7", life, me.Life)
	}
	// A Vampire dies: pay 2 life, draw a card.
	life, hand := me.Life, me.Hand.Size()
	b27Kill(g, vamp)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Life != life-2 || me.Hand.Size() != hand+1 {
		t.Errorf("paid 2 and drew 1: life %d → %d, hand %d → %d", life, me.Life, hand, me.Hand.Size())
	}
	b27Kill(g, bear)
	if latestTriggerPrompt(g, me.ID) != nil {
		t.Error("a Bear is not a Vampire")
	}
	b27Kill(g, home)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if me.Life != life-2 || me.Hand.Size() != hand+1 {
		t.Error("declining pays and draws nothing")
	}
}

func TestB34HobgoblinBanditLordPumpsGoblinsAndCountsThisTurnsEntries(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lord := b12Push(g, me.ID, "Hobgoblin Bandit Lord", "Creature — Goblin Rogue", b34HobgoblinBanditLordOracle, 2, 3)
	old := b12Creature(g, me.ID, "Old Goblin", "Creature — Goblin", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if p := effectivePower(t, g, old); p != 2 {
		t.Errorf("another Goblin you control is a %d/x, want 2", p)
	}
	if p := effectivePower(t, g, theirs); p != 1 {
		t.Errorf("an opponent's Goblin is a %d/x — \"you control\"", p)
	}
	if p := effectivePower(t, g, lord); p != 2 {
		t.Errorf("the Lord is a %d/x — \"other\"", p)
	}
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("a Bear is a %d/x", p)
	}
	// The seeded Goblins entered before this turn's upkeep; two more
	// enter now, one as a token, and the changeling counts.
	advanceToMain(t, g)
	castCatalogSpell(t, g, "New Goblin", "Creature — Goblin", "", nil)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 1) })
	shifter := handCardFull(me, "Changeling", "Creature — Shapeshifter", "{1}", "", nil)
	for i := range me.Hand.Cards {
		if me.Hand.Cards[i].InstanceID == shifter {
			me.Hand.Cards[i].Keywords = []string{game.KeywordChangeling}
		}
	}
	if err := g.CastSpell(me.ID, shifter, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the changeling: %v", err)
	}
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Elf", "Creature — Elf", "", nil)
	passPriorityAroundTable(t, g)
	life := opp.Life
	b06AddMana(me, "R")
	b16Activate(t, g, me.ID, lord, 0, game.ActivateAbilityParams{Targets: b16TargetPlayer(opp.ID)})
	if opp.Life != life-3 {
		t.Errorf("three Goblins entered this turn: %d → %d, want -3", life, opp.Life)
	}
	if !b16Tapped(t, g, lord) {
		t.Error("the tap is the cost")
	}
}

func TestB34ElanorMakesAFoodAndRampsAfterEatingOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedSearchLibrary(me,
		game.Card{Name: "Filler", TypeLine: "Sorcery"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Deep", TypeLine: "Sorcery"},
	)
	forest := ids[1]
	castCatalogSpell(t, g, "Elanor Gardner", "Legendary Creature — Halfling Scout", b34ElanorGardnerOracle, nil)
	passPriorityAroundTable(t, g)
	food := battlefieldIDNamed(g, me.ID, "Food")
	if food == uuid.Nil {
		t.Fatal("Elanor makes a Food on entry")
	}
	// No Food eaten: the end step asks nothing.
	advanceToEndStepOf(t, g, 0)
	if latestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("no Food sacrificed this turn, no prompt")
	}
	passPriorityAroundTable(t, g)
	// Next turn: eat the Food, then the end step offers the land.
	advanceToMainOf(t, g, 0)
	b06AddMana(me, "C", "C")
	b16Activate(t, g, me.ID, food, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(food) {
		t.Fatal("the Food is sacrificed")
	}
	advanceToEndStepOf(t, g, 0)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if c := searchChoiceFor(g, me.ID); c != nil {
		answerSearchByID(t, g, me.ID, forest)
	}
	if !g.Battlefield.Contains(forest) {
		t.Fatal("the basic is put onto the battlefield")
	}
	if !b16Tapped(t, g, forest) {
		t.Error("it enters tapped")
	}
}

func TestB34SeizanEveryUpkeepCostsTwoLifeAndDrawsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Seizan, Perverter of Truth", "Legendary Creature — Demon Spirit", b34SeizanOracle, 6, 5)
	life, hand := opp.Life, opp.Hand.Size()
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("the opponent's upkeep: life %d → %d, want -2", life, opp.Life)
	}
	if opp.Hand.Size() != hand+2 {
		t.Errorf("the opponent's upkeep: hand %d → %d, want +2", hand, opp.Hand.Size())
	}
	for seat := 2; seat < 4; seat++ {
		advanceToUpkeepOf(t, g, seat)
		passPriorityAroundTable(t, g)
	}
	life, hand = me.Life, me.Hand.Size()
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if me.Life != life-2 || me.Hand.Size() != hand+2 {
		t.Errorf("your own upkeep too: life %d → %d, hand %d → %d", life, me.Life, hand, me.Hand.Size())
	}
}

// --- enchantments and the Background --------------------------------

func TestB34OnWingsOfGoldLiftsZombiesAndTokensAndMakesAZombie(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "On Wings of Gold", "Enchantment", b34OnWingsOfGoldOracle, 0, 0)
	zombie := b12Creature(g, me.ID, "Gravecrawler", "Creature — Zombie", 2, 2)
	token := b34Token(g, me.ID, "Soldier", "Token Creature — Soldier", 1, 1)
	both := b34Token(g, me.ID, "Zombie Token", "Token Creature — Zombie", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Zombie", "Creature — Zombie", 2, 2)
	for _, tc := range []struct {
		id   uuid.UUID
		want int
	}{{zombie, 3}, {token, 2}, {both, 3}} {
		if p := effectivePower(t, g, tc.id); p != tc.want {
			t.Errorf("a Zombie or token you control gets +1/+1: power %d, want %d", p, tc.want)
		}
		if !hasEffectiveKeyword(t, g, tc.id, "flying") {
			t.Error("a Zombie or token you control has flying")
		}
	}
	if p := effectivePower(t, g, bear); p != 2 || hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("a nontoken non-Zombie is untouched")
	}
	if p := effectivePower(t, g, theirs); p != 2 || hasEffectiveKeyword(t, g, theirs, "flying") {
		t.Error("an opponent's Zombie is untouched")
	}
	// Cards leaving the graveyard: one Zombie per batch.
	a := pushGraveyardCardForTest(me, "Dead A")
	b := pushGraveyardCardForTest(me, "Dead B")
	g.WithWriteLock(func() { _ = g.ExileCardsForEffect([]uuid.UUID{a, b}) })
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Zombie"); n != 1 {
		t.Fatalf("\"one or more\": %d Zombie tokens for two cards leaving at once, want 1", n)
	}
	made := battlefieldIDNamed(g, me.ID, "Zombie")
	if p := effectivePower(t, g, made); p != 2 || !hasEffectiveKeyword(t, g, made, "flying") {
		t.Error("the Zombie it makes is a 1/1 white Zombie token, lifted by its own anthem to a 2/2 flier")
	}
	theirsDead := pushGraveyardCardForTest(opp, "Their Dead")
	b34ExileFromGraveyard(t, g, theirsDead)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Zombie"); n != 1 {
		t.Error("an opponent's graveyard is not yours")
	}
}

func TestB34TormodMakesATappedZombiePerBatchLeavingYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Tormod, the Desecrator", "Legendary Creature — Zombie Wizard", b34TormodOracle, 4, 2)
	dead := pushGraveyardCardForTest(me, "Dead")
	b34ExileFromGraveyard(t, g, dead)
	passPriorityAroundTable(t, g)
	zombie := battlefieldIDNamed(g, me.ID, "Zombie")
	if zombie == uuid.Nil {
		t.Fatal("a card leaving your graveyard makes a Zombie")
	}
	if !b16Tapped(t, g, zombie) {
		t.Error("the Zombie enters tapped")
	}
	if p := effectivePower(t, g, zombie); p != 2 {
		t.Errorf("a 2/2: power %d", p)
	}
	back := pushGraveyardCardForTest(me, "Regrowth Target")
	g.WithWriteLock(func() { _ = g.ReturnFromGraveyardForEffect(back, game.ZoneHand) })
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Zombie"); n != 2 {
		t.Errorf("a regrowth is a card leaving the graveyard: %d Zombies, want 2", n)
	}
}

func TestB34PollutedBondsDrainsAnOpponentPerLand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Polluted Bonds", "Enchantment", b34PollutedBondsOracle, 0, 0)
	mine := me.Life
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	if me.Life != mine {
		t.Fatal("your own land is not an opponent's")
	}
	life := opp.Life
	b34PlayLandAs(t, g, 1, "Swamp", "Basic Land — Swamp")
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("the opponent's land: their life %d → %d, want -2", life, opp.Life)
	}
	if me.Life != mine+2 {
		t.Errorf("you gain 2: %d → %d", mine, me.Life)
	}
	// A land put onto the battlefield by an effect counts too.
	b34LandEntersByEffect(t, g, opp, "Fetched Swamp", "Basic Land — Swamp")
	passPriorityAroundTable(t, g)
	if opp.Life != life-4 || me.Life != mine+4 {
		t.Errorf("a fetched land: their %d, yours %d", opp.Life, me.Life)
	}
}

func TestB34GuildArtisanPaysTreasureWhenYourCommanderAttacksTheHealthiestPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	b12Push(g, me.ID, "Guild Artisan", "Legendary Enchantment — Background", b34GuildArtisanOracle, 0, 0)
	commander := b21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 3, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	// Everyone on 40: no opponent has MORE than the attacked player.
	// The trigger resolves in the declare-attackers step, before any
	// damage moves a life total.
	declareAttack(t, g, opp.ID, commander, bear)
	if n := len(g.PendingTriggers) + len(g.StackMeta); n != 1 {
		t.Fatalf("one trigger for the commander, none for the Bear: %d", n)
	}
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Treasure"); n != 2 {
		t.Fatalf("the commander attacking the (joint) healthiest opponent: %d Treasures, want 2", n)
	}
	advanceTo(t, g, game.StepCombatDamage)
	// Now the attacked player is behind another opponent: nothing.
	advanceToMainOf(t, g, 0)
	if other.Life <= opp.Life {
		t.Fatalf("fixture: seat 2 (%d) should be ahead of seat 1 (%d)", other.Life, opp.Life)
	}
	declareAttack(t, g, opp.ID, commander)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Treasure"); n != 2 {
		t.Errorf("another opponent has more life than the attacked player: still 2 Treasures, got %d", n)
	}
	// Attacking the healthiest opponent pays again; a non-commander
	// never does.
	advanceToMainOf(t, g, 0)
	declareAttack(t, g, other.ID, commander)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Treasure"); n != 4 {
		t.Errorf("the healthiest opponent attacked: %d Treasures, want 4", n)
	}
	advanceToMainOf(t, g, 0)
	declareAttack(t, g, other.ID, bear)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Treasure"); n != 4 {
		t.Errorf("a Bear is not your commander: %d Treasures, want 4", n)
	}
}

func TestB34GuildArtisanIgnoresAnOpponentsCommander(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Guild Artisan", "Legendary Enchantment — Background", b34GuildArtisanOracle, 0, 0)
	theirs := b21Commander(g, opp.ID, "Their Commander", "Legendary Creature — Human", 3, 3)
	advanceToMainOf(t, g, 1)
	declareAttack(t, g, g.Seats[2].ID, theirs)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Treasure"); n != 0 {
		t.Errorf("an opponent's commander is not one you own: %d Treasures, want 0", n)
	}
}

func TestB34LordOfTheUndeadPumpsEveryOtherZombieAndRegrowsOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lord := b12Push(g, me.ID, "Lord of the Undead", "Creature — Zombie", b34LordOfTheUndeadOracle, 2, 2)
	mine := b12Creature(g, me.ID, "Zombie", "Creature — Zombie", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Zombie", "Creature — Zombie", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if p := effectivePower(t, g, mine); p != 3 {
		t.Errorf("another Zombie you control is a %d/x, want 3", p)
	}
	if p := effectivePower(t, g, theirs); p != 3 {
		t.Errorf("an opponent's Zombie is a %d/x, want 3 — the printed card has no controller clause", p)
	}
	if p := effectivePower(t, g, lord); p != 2 {
		t.Errorf("the Lord is a %d/x — \"other\"", p)
	}
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("a Bear is a %d/x", p)
	}
	dead := b17GraveyardCard(me, "Dead Zombie", "Creature — Zombie", "{B}")
	kindred := b17GraveyardCard(me, "Zombie Spell", "Kindred Sorcery — Zombie", "{B}")
	elf := b17GraveyardCard(me, "Dead Elf", "Creature — Elf", "{G}")
	advanceToMain(t, g)
	b06AddMana(me, "C", "B")
	if err := g.ActivateCatalogAbility(me.ID, lord, 0, game.ActivateAbilityParams{Targets: b16TargetCard(elf)}); err == nil {
		t.Fatal("an Elf is not a Zombie card")
	}
	b16Activate(t, g, me.ID, lord, 0, game.ActivateAbilityParams{Targets: b16TargetCard(dead)})
	if !me.Hand.Contains(dead) {
		t.Error("the Zombie card returns to hand")
	}
	if !b16Tapped(t, g, lord) {
		t.Error("the tap is the cost")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(lord) })
	b06AddMana(me, "C", "B")
	b16Activate(t, g, me.ID, lord, 0, game.ActivateAbilityParams{Targets: b16TargetCard(kindred)})
	if !me.Hand.Contains(kindred) {
		t.Error("a Kindred Zombie card is a Zombie card")
	}
}

// --- lands ----------------------------------------------------------

func TestB34CoriMountainMonasteryChecksAndImpulses(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tapped := b12PlayFromHand(t, g, "Cori Mountain Monastery", "Land", b34CoriMountainMonasteryOrcl, game.CastSpellParams{})
	if !b16Tapped(t, g, tapped) {
		t.Fatal("with no Plains or Island it enters tapped")
	}
	b12Permanent(g, me.ID, "Hallowed Fountain", "Land — Plains Island")
	advanceToMainOf(t, g, 0)
	monastery := b12PlayFromHand(t, g, "Cori Mountain Monastery", "Land", b34CoriMountainMonasteryOrcl, game.CastSpellParams{})
	if b16Tapped(t, g, monastery) {
		t.Fatal("with a Plains it enters untapped")
	}
	b28TapForMana(t, g, me.ID, monastery, "")
	if got := poolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("tapped for {R}: pool %v", got)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(monastery) })
	ids := seedSearchLibrary(me,
		game.Card{Name: "Top Land", TypeLine: "Basic Land — Mountain"},
		game.Card{Name: "Deep Card", TypeLine: "Sorcery"},
	)
	b06AddMana(me, "C", "C", "C", "R")
	b16Activate(t, g, me.ID, monastery, 0, game.ActivateAbilityParams{})
	if !g.Exile.Contains(ids[0]) || !me.Library.Contains(ids[1]) {
		t.Fatal("the top card is exiled")
	}
	perm := exiledPermission(g, ids[0])
	if perm.Player != me.ID || perm.CastOnly || !perm.Active(me.ID, g.Turn.Number) {
		t.Errorf("%+v: the controller may PLAY it, a land included", perm)
	}
	// Through the opponents' turns and the controller's next turn the
	// grant holds; it lapses after that turn.
	advanceToMainOf(t, g, 2)
	if !exiledPermission(g, ids[0]).Active(me.ID, g.Turn.Number) {
		t.Error("the grant survives the opponents' turns")
	}
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if p := exiledPermission(g, ids[0]); p.UntilTurn != g.Turn.Number || !p.Active(me.ID, g.Turn.Number) {
		t.Errorf("after your upkeep the grant is %+v, want live through this turn only", p)
	}
	advanceToMainOf(t, g, 1)
	if exiledPermission(g, ids[0]).Active(me.ID, g.Turn.Number) {
		t.Error("the grant ends with your next turn")
	}
}

func TestB34BristlingBackwoodsEntersTappedPingsAnOpponentAndTapsForEitherColour(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[2]
	before := opp.Life
	land := b12PlayFromHand(t, g, "Bristling Backwoods", "Land — Desert", b34BristlingBackwoodsOracle, game.CastSpellParams{})
	if !b16Tapped(t, g, land) {
		t.Fatal("the Desert enters tapped")
	}
	if p := latestPickTarget(g, me.ID); p == nil || hasID(p.PickTargetPlayers, me.ID) || !hasID(p.PickTargetPlayers, opp.ID) {
		t.Fatalf("the trigger asks for an opponent, never you: %+v", p)
	}
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("target opponent takes 1: %d → %d", before, opp.Life)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	b28TapForMana(t, g, me.ID, land, "G")
	if got := poolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("tapped for {G}: pool %v", got)
	}
}

func TestB34TurbulentFenEntersUntappedOnlyAgainstEightOpposingLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	b34Lands(g, opp.ID, 4)
	b34Lands(g, other.ID, 3)
	b34Lands(g, me.ID, 5)
	tapped := b12PlayFromHand(t, g, "Turbulent Fen", "Land — Swamp Forest", b34TurbulentFenOracle, game.CastSpellParams{})
	if !b16Tapped(t, g, tapped) {
		t.Fatal("seven opposing lands — your own do not count — is not eight: it enters tapped")
	}
	b34Lands(g, other.ID, 1)
	advanceToMainOf(t, g, 0)
	fen := b12PlayFromHand(t, g, "Turbulent Fen", "Land — Swamp Forest", b34TurbulentFenOracle, game.CastSpellParams{})
	if b16Tapped(t, g, fen) {
		t.Fatal("eight opposing lands between two opponents: it enters untapped")
	}
	b28TapForMana(t, g, me.ID, fen, "B")
	if got := poolColors(me); len(got) != 1 || got[0] != "B" {
		t.Errorf("tapped for {B}: pool %v", got)
	}
}
