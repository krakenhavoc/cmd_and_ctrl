package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// b34Creature seeds a non-catalog creature with a type line and
// colours, able to attack and tap.
func b34Creature(g *game.Game, owner uuid.UUID, name, typeLine string, power, toughness int, colors ...string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Colors: colors,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// b34Token pushes a creature token under controller, able to attack.
func b34Token(g *game.Game, controller uuid.UUID, name, typeLine string, power, toughness int) uuid.UUID {
	return effectstest.PushToken(g, controller, game.Card{
		Name: name, TypeLine: typeLine, Power: power, Toughness: toughness,
	})
}

// b34Lands seeds n vanilla lands under owner.
func b34Lands(g *game.Game, owner uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		effectstest.B12Permanent(g, owner, "Wastes", "Basic Land")
	}
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

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Bloodgift
// Demon was already on main from the life-for-cards work (#507), so
// the table is the issue's 42 minus it and minus the nine declared
// skips.
func TestBatch34CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.B34InvigoratingSurgeOracle:    "Invigorating Surge",
		effectstest.B34ReveillarkOracle:           "Reveillark",
		effectstest.B34YavimayaElderOracle:        "Yavimaya Elder",
		effectstest.B34VitalizeOracle:             "Vitalize",
		effectstest.B34FiendishDuoOracle:          "Fiendish Duo",
		effectstest.B34ArmorcraftJudgeOracle:      "Armorcraft Judge",
		effectstest.B34OmnathLocusOfTheRoilOracle: "Omnath, Locus of the Roil",
		effectstest.B34BeetlebackChiefOracle:      "Beetleback Chief",
		effectstest.B34CoriMountainMonasteryOrcl:  "Cori Mountain Monastery",
		effectstest.B34BristlingBackwoodsOracle:   "Bristling Backwoods",
		effectstest.B34SporemoundOracle:           "Sporemound",
		effectstest.B34GuildArtisanOracle:         "Guild Artisan",
		effectstest.B34LordOfTheUndeadOracle:      "Lord of the Undead",
		effectstest.B34AuroralProcessionOracle:    "Auroral Procession",
		effectstest.B34KioraTheRisingTideOracle:   "Kiora, the Rising Tide",
		effectstest.B34RegalForceOracle:           "Regal Force",
		effectstest.B34TurbulentFenOracle:         "Turbulent Fen",
		effectstest.B34OnWingsOfGoldOracle:        "On Wings of Gold",
		effectstest.B34DragonbornChampionOracle:   "Dragonborn Champion",
		effectstest.B34ArbaazMirOracle:            "Arbaaz Mir",
		effectstest.B34MaelstromPulseOracle:       "Maelstrom Pulse",
		effectstest.B34SeizanOracle:               "Seizan, Perverter of Truth",
		effectstest.B34TormodOracle:               "Tormod, the Desecrator",
		effectstest.B34MapTheFrontierOracle:       "Map the Frontier",
		effectstest.B34RapaciousDragonOracle:      "Rapacious Dragon",
		effectstest.B34CrosswayTroublemakersOrcl:  "Crossway Troublemakers",
		effectstest.B34DeathsproutOracle:          "Deathsprout",
		effectstest.B34ProsperityOracle:           "Prosperity",
		effectstest.B34HobgoblinBanditLordOracle:  "Hobgoblin Bandit Lord",
		effectstest.B34ElanorGardnerOracle:        "Elanor Gardner",
		effectstest.B34VillageBellRingerOracle:    "Village Bell-Ringer",
		effectstest.B34PollutedBondsOracle:        "Polluted Bonds",
	}
	if len(want) != 32 {
		t.Fatalf("the batch ships 32 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, "+1/+1", 3) })
	effectstest.CastCatalogSpell(t, g, "Invigorating Surge", "Instant", effectstest.B34InvigoratingSurgeOracle, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, "+1/+1"); got != 8 {
		t.Errorf("three counters, plus one, doubled: %d, want 8", got)
	}
	if p := effectstest.CurrentPower(t, g, bear); p != 10 {
		t.Errorf("the Bear is a %d/x, want 10", p)
	}
	bare := effectstest.B12Creature(g, me.ID, "Bare Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Invigorating Surge", "Instant", effectstest.B34InvigoratingSurgeOracle, effectstest.B16TargetCard(bare))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bare, "+1/+1"); got != 2 {
		t.Errorf("no counters, plus one, doubled: %d, want 2", got)
	}
	surge := effectstest.B13HandCard(me, "Invigorating Surge", "Instant", effectstest.B34InvigoratingSurgeOracle)
	if err := g.CastSpell(me.ID, surge, game.CastSpellParams{Targets: effectstest.B16TargetCard(theirs)}); err == nil {
		t.Error("an opponent's creature is not \"target creature you control\"")
	}
}

func TestB34VitalizeUntapsOnlyYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{a, b, land, theirs} {
		effectstest.B16Tap(g, id)
	}
	effectstest.CastCatalogSpell(t, g, "Vitalize", "Instant", effectstest.B34VitalizeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16Tapped(t, g, a) || effectstest.B16Tapped(t, g, b) {
		t.Error("your creatures untap")
	}
	if !effectstest.B16Tapped(t, g, land) {
		t.Error("a land is not a creature")
	}
	if !effectstest.B16Tapped(t, g, theirs) {
		t.Error("an opponent's creature stays tapped")
	}
}

func TestB34AuroralProcessionReturnsAnyCardFromYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.B17GraveyardCard(me, "Dead Forest", "Basic Land — Forest", "")
	theirs := effectstest.B17GraveyardCard(opp, "Their Bolt", "Instant", "{R}")
	effectstest.CastCatalogSpell(t, g, "Auroral Procession", "Instant", effectstest.B34AuroralProcessionOracle, effectstest.B16TargetCard(land))
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(land) {
		t.Error("the land card is returned to hand")
	}
	procession := effectstest.B13HandCard(me, "Auroral Procession", "Instant", effectstest.B34AuroralProcessionOracle)
	if err := g.CastSpell(me.ID, procession, game.CastSpellParams{Targets: effectstest.B16TargetCard(theirs)}); err == nil {
		t.Error("an opponent's graveyard is not yours")
	}
}

func TestB34MaelstromPulseDestroysEveryPermanentSharingTheName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	a := b34Token(g, opp.ID, "Goblin", "Token Creature — Goblin", 1, 1)
	b := b34Token(g, opp.ID, "Goblin", "Token Creature — Goblin", 1, 1)
	c := b34Token(g, other.ID, "Goblin", "Token Creature — Goblin", 1, 1)
	mine := b34Token(g, me.ID, "Goblin", "Token Creature — Goblin", 1, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	pulse := effectstest.B13HandCard(me, "Maelstrom Pulse", "Sorcery", effectstest.B34MaelstromPulseOracle)
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, pulse, game.CastSpellParams{Targets: effectstest.B16TargetCard(land)}); err == nil {
		t.Fatal("a land is not \"target nonland permanent\"")
	}
	effectstest.CastCatalogSpell(t, g, "Maelstrom Pulse", "Sorcery", effectstest.B34MaelstromPulseOracle, effectstest.B16TargetCard(a))
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "Twin", "Creature — Bear", 2, 2)
	tough := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Twin", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Keywords: []string{"indestructible"}, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Maelstrom Pulse", "Sorcery", effectstest.B34MaelstromPulseOracle, effectstest.B16TargetCard(a))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) {
		t.Error("the target is destroyed")
	}
	if !g.Battlefield.Contains(tough) {
		t.Error("the single-target verb honours indestructible on the namesake")
	}
	_ = me
}

func TestB34DeathsproutDestroysAndRampsATappedBasic(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Snow-Covered Forest", TypeLine: "Basic Snow Land — Forest"},
		game.Card{Name: "Swamp", TypeLine: "Basic Land — Swamp"},
		game.Card{Name: "Overgrown Tomb", TypeLine: "Land — Swamp Forest"},
	)
	effectstest.CastCatalogSpell(t, g, "Deathsprout", "Instant", effectstest.B34DeathsproutOracle, effectstest.B16TargetCard(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Fatal("the creature is destroyed")
	}
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if len(c.SearchCards) != 2 || effectstest.SearchOptionNamed(g, c, "Overgrown Tomb") != uuid.Nil {
		t.Errorf("the two basics are offered, the shockland is not: %d offered", len(c.SearchCards))
	}
	effectstest.AnswerSearchByID(t, g, me.ID, ids[2])
	if !g.Battlefield.Contains(ids[2]) || !effectstest.B16Tapped(t, g, ids[2]) {
		t.Error("the chosen basic enters the battlefield tapped")
	}
}

func TestB34MapTheFrontierFetchesBasicsAndDesertsTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Desert of the Glorified", TypeLine: "Land — Desert"},
		game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Overgrown Tomb", TypeLine: "Land — Swamp Forest"},
	)
	effectstest.CastCatalogSpell(t, g, "Map the Frontier", "Sorcery", effectstest.B34MapTheFrontierOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if c.SearchMax != 2 || len(c.SearchCards) != 3 {
		t.Errorf("two basics and a Desert offered, up to two taken: %d offered, max %d", len(c.SearchCards), c.SearchMax)
	}
	if effectstest.SearchOptionNamed(g, c, "Overgrown Tomb") != uuid.Nil || effectstest.SearchOptionNamed(g, c, "Bear") != uuid.Nil {
		t.Error("a nonbasic non-Desert and a creature are not offered")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, ids[0], ids[1])
	for _, id := range []uuid.UUID{ids[0], ids[1]} {
		if !g.Battlefield.Contains(id) {
			t.Fatal("both chosen lands are put onto the battlefield")
		}
		if !effectstest.B16Tapped(t, g, id) {
			t.Error("they enter tapped")
		}
	}
	if me.Library.Size() != 3 {
		t.Errorf("library %d, want 3", me.Library.Size())
	}
}

func TestB34ProsperityEachPlayerDrawsX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	hands := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		hands[i] = p.Hand.Size()
	}
	effectstest.CastXSpell(t, g, "Prosperity", "Sorcery", effectstest.B34ProsperityOracle, "{X}{U}", 2, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.Hand.Size() != hands[i]+2 {
			t.Errorf("seat %d hand %d, want %d", i, p.Hand.Size(), hands[i]+2)
		}
	}
}

// --- creatures ------------------------------------------------------

func TestB34ReveillarkReturnsUpToTwoSmallCreaturesOnAnyLeave(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lark := effectstest.B12Push(g, me.ID, "Reveillark", "Creature — Elemental", effectstest.B34ReveillarkOracle, 4, 3)
	small := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: small, Name: "Dead Elf", TypeLine: "Creature — Elf", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	two := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: two, Name: "Dead Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	big := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: big, Name: "Dead Titan", TypeLine: "Creature — Giant", Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID})
	theirs := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: theirs, Name: "Their Elf", TypeLine: "Creature — Elf", Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	if !effectstest.HasEffectiveKeyword(t, g, lark, "flying") {
		t.Error("printed flying did not reach the effective abilities")
	}
	// Bounce, not death: "leaves the battlefield".
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(lark) })
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if !effectstest.HasID(p.PickTargetCards, small) || !effectstest.HasID(p.PickTargetCards, two) {
		t.Error("creature cards with power 2 or less are offered")
	}
	if effectstest.HasID(p.PickTargetCards, big) || effectstest.HasID(p.PickTargetCards, theirs) {
		t.Error("a 6/6 and an opponent's card are not offered")
	}
	if p.PickTargetMin != 0 || p.PickTargetMax != 2 {
		t.Errorf("\"up to two\": min %d max %d", p.PickTargetMin, p.PickTargetMax)
	}
	effectstest.B17PickCards(t, g, me.ID, small, two)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(small) || !g.Battlefield.Contains(two) {
		t.Error("both chosen creatures return to the battlefield")
	}
	if !me.Hand.Contains(lark) {
		t.Error("the bounced Lark is in hand")
	}
}

func TestB34ReveillarkEvokedIsSacrificedAndStillReturns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	small := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: small, Name: "Dead Elf", TypeLine: "Creature — Elf", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	lark := effectstest.CastWithAltCost(t, g, "Reveillark", "Creature — Elemental", effectstest.B34ReveillarkOracle, "evoke")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, small)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(lark) || !me.Graveyard.Contains(lark) {
		t.Error("the evoked Lark is sacrificed on entry")
	}
	if !g.Battlefield.Contains(small) {
		t.Error("the leave trigger returns the chosen creature")
	}
}

func TestB34YavimayaElderSacrificesForLandsThenACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	elder := effectstest.B12Push(g, me.ID, "Yavimaya Elder", "Creature — Human Druid", effectstest.B34YavimayaElderOracle, 2, 1)
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Draw Me", TypeLine: "Sorcery"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Snow-Covered Swamp", TypeLine: "Basic Snow Land — Swamp"},
		game.Card{Name: "Overgrown Tomb", TypeLine: "Land — Swamp Forest"},
		game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"},
		game.Card{Name: "Filler", TypeLine: "Sorcery"},
	)
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, elder, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(elder) {
		t.Fatal("the sacrifice is the cost")
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	// The dies trigger sits above the ability: it resolves first and
	// asks which basics; the draw waits underneath.
	for i := 0; i < 8 && effectstest.SearchChoiceFor(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the dies trigger resolves first and asks for the lands")
	}
	if c.SearchMax != 2 || len(c.SearchCards) != 3 || effectstest.SearchOptionNamed(g, c, "Overgrown Tomb") != uuid.Nil {
		t.Errorf("the three basics offered, up to two taken: %d offered, max %d", len(c.SearchCards), c.SearchMax)
	}
	if me.Hand.Size() != hand {
		t.Error("the ability's draw waits below the trigger")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, ids[1], ids[2])
	if !me.Hand.Contains(ids[1]) || !me.Hand.Contains(ids[2]) {
		t.Error("both basics go to hand")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+3 {
		t.Errorf("two lands and a card: hand %d → %d", hand, me.Hand.Size())
	}
}

func TestB34YavimayaElderDeclinedSearchesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	elder := effectstest.B12Push(g, me.ID, "Yavimaya Elder", "Creature — Human Druid", effectstest.B34YavimayaElderOracle, 2, 1)
	effectstest.B27Kill(g, elder)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Error("declining the trigger searches nothing")
	}
}

func TestB34FiendishDuoDoublesDamageToOpponentsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	duo := effectstest.B12Push(g, me.ID, "Fiendish Duo", "Creature — Devil", effectstest.B34FiendishDuoOracle, 5, 5)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 5)
	if !effectstest.HasEffectiveKeyword(t, g, duo, "first strike") {
		t.Error("printed first strike did not reach the effective abilities")
	}
	// A Bolt at an opponent is doubled; one at you is not.
	life := opp.Life
	b07BoltPlayer(t, g, opp.ID)
	if opp.Life != life-6 {
		t.Errorf("a Bolt at an opponent: %d → %d, want -6", life, opp.Life)
	}
	mine := me.Life
	effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "{R}", effectstest.B16TargetPlayer(me.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != mine-3 {
		t.Errorf("a Bolt at the Duo's controller: %d → %d, want -3", mine, me.Life)
	}
	// Damage to an opponent's creature is untouched.
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetCard(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, theirs); got != 3 {
		t.Errorf("a Bolt at a creature marks %d, want 3", got)
	}
	// Combat damage to an opponent is doubled too.
	life = opp.Life
	effectstest.AttackWith(t, g, opp.ID, bear)
	if opp.Life != life-4 {
		t.Errorf("a 2/2 connecting: %d → %d, want -4", life, opp.Life)
	}
}

func TestB34ArmorcraftJudgeDrawsPerCounteredCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Bare Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(a, "+1/+1", 1)
		_ = g.AddCounterForEffect(b, "+1/+1", 3)
		_ = g.AddCounterForEffect(theirs, "+1/+1", 1)
	})
	effectstest.CastCatalogSpell(t, g, "Armorcraft Judge", "Creature — Elf Artificer", effectstest.B34ArmorcraftJudgeOracle, nil)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("two of your creatures carry counters: drew %d, want 2", me.Hand.Size()-hand)
	}
}

func TestB34RegalForceDrawsPerGreenCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b34Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1, "G")
	b34Creature(g, me.ID, "Gold Elf", "Creature — Elf", 1, 1, "G", "W")
	b34Creature(g, me.ID, "Bird", "Creature — Bird", 1, 1, "U")
	effectstest.B12Creature(g, me.ID, "Colourless", "Artifact Creature — Golem", 3, 3)
	b34Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1, "G")
	id := effectstest.HandCardFull(me, "Regal Force", "Creature — Elemental", "{4}{G}{G}{G}", effectstest.B34RegalForceOracle, []string{"G"})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+3 {
		t.Errorf("two green creatures and the Force itself: drew %d, want 3", me.Hand.Size()-hand)
	}
}

func TestB34OmnathEntersForDamagePerElementalAndGrowsOnLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	other := effectstest.B12Creature(g, me.ID, "Flamekin", "Creature — Elemental Shaman", 2, 2)
	effectstest.B34Changeling(g, me.ID, "Changeling")
	theirs := effectstest.B12Creature(g, opp.ID, "Their Elemental", "Creature — Elemental", 2, 2)
	life := opp.Life
	omnath := effectstest.CastCatalogSpell(t, g, "Omnath, Locus of the Roil", "Legendary Creature — Elemental", effectstest.B34OmnathLocusOfTheRoilOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("Omnath, the Flamekin and the changeling are three Elementals: %d → %d, want -3", life, opp.Life)
	}
	// Landfall: a counter on target Elemental you control; seven lands
	// is not eight.
	b34Lands(g, me.ID, 6)
	hand := me.Hand.Size()
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if !effectstest.HasID(p.PickTargetCards, omnath) || !effectstest.HasID(p.PickTargetCards, other) {
		t.Error("your Elementals are offered")
	}
	if effectstest.HasID(p.PickTargetCards, theirs) {
		t.Error("an opponent's Elemental is not \"you control\"")
	}
	effectstest.PickCard(t, g, me.ID, other)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, other, "+1/+1"); got != 1 {
		t.Errorf("the chosen Elemental gets a counter: %d", got)
	}
	if me.Hand.Size() != hand {
		t.Error("seven lands: no card")
	}
	// The eighth land — put onto the battlefield by an effect — draws.
	effectstest.B34LandEntersByEffect(t, g, me, "Late Forest", "Basic Land — Forest")
	effectstest.B04WaitForPick(t, g, me.ID)
	hand = me.Hand.Size()
	effectstest.PickCard(t, g, me.ID, omnath)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, omnath, "+1/+1"); got != 1 {
		t.Errorf("Omnath gets the counter this time: %d", got)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("eight or more lands: drew %d, want 1", me.Hand.Size()-hand)
	}
}

func TestB34BeetlebackChiefBringsTwoGoblins(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Beetleback Chief", "Creature — Goblin Warrior", effectstest.B34BeetlebackChiefOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Goblin"); n != 2 {
		t.Errorf("%d Goblins, want 2", n)
	}
}

func TestB34RapaciousDragonBringsTwoTreasures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dragon := effectstest.CastCatalogSpell(t, g, "Rapacious Dragon", "Creature — Dragon", effectstest.B34RapaciousDragonOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure"); n != 2 {
		t.Errorf("%d Treasures, want 2", n)
	}
	if !effectstest.HasEffectiveKeyword(t, g, dragon, "flying") {
		t.Error("printed flying did not reach the effective abilities")
	}
}

func TestB34SporemoundMakesASaprolingPerLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Sporemound", "Creature — Fungus", effectstest.B34SporemoundOracle, 3, 3)
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Saproling"); n != 1 {
		t.Fatalf("%d Saprolings after a land drop, want 1", n)
	}
	effectstest.B34LandEntersByEffect(t, g, me, "Fetched Forest", "Basic Land — Forest")
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Saproling"); n != 2 {
		t.Errorf("%d Saprolings after a second land, want 2", n)
	}
	sap := effectstest.BattlefieldIDNamed(g, me.ID, "Saproling")
	if p := effectstest.EffectivePower(t, g, sap); p != 1 {
		t.Errorf("the Saproling is a %d/x, want 1", p)
	}
}

func TestB34VillageBellRingerFlashesInAndUntapsYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.B16Tap(g, a)
	effectstest.B16Tap(g, theirs)
	ringer := effectstest.CastCatalogSpell(t, g, "Village Bell-Ringer", "Creature — Human Scout", effectstest.B34VillageBellRingerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, ringer, "flash") {
		t.Error("printed flash did not reach the effective abilities")
	}
	if effectstest.B16Tapped(t, g, a) {
		t.Error("your creature untaps")
	}
	if !effectstest.B16Tapped(t, g, theirs) {
		t.Error("an opponent's creature does not")
	}
}

func TestB34ArbaazMirPingsOnHimselfAndOnNontokenHistoricPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := effectstest.LifeOfOpponents(g)
	mine := me.Life
	effectstest.CastCatalogSpell(t, g, "Arbaaz Mir", "Legendary Creature — Human Assassin", effectstest.B34ArbaazMirOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats[1:] {
		if p.Life != before[i]-1 {
			t.Errorf("Arbaaz entering: opponent %d lost %d, want 1", i+1, before[i]-p.Life)
		}
	}
	if me.Life != mine+1 {
		t.Errorf("Arbaaz entering: life %d → %d, want +1", mine, me.Life)
	}
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOfOpponents(g); got[0] != before[0]-1 {
		t.Error("a Bear is not historic")
	}
	effectstest.CastCatalogSpell(t, g, "Signet", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Legend", "Legendary Creature — Human", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Saga", "Enchantment — Saga", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats[1:] {
		if p.Life != before[i]-4 {
			t.Errorf("an artifact, a legend and a Saga: opponent %d lost %d, want 4", i+1, before[i]-p.Life)
		}
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TreasureToken(), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOfOpponents(g); got[0] != before[0]-4 {
		t.Error("a Treasure is an artifact but a token: \"nontoken\"")
	}
	if me.Life != mine+4 {
		t.Errorf("life %d → %d, want +4", mine, me.Life)
	}
}

func TestB34DragonbornChampionDrawsOnFiveDamageInOneHit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	champ := effectstest.B12Push(g, me.ID, "Dragonborn Champion", "Creature — Dragon Warrior", effectstest.B34DragonbornChampionOracle, 5, 3)
	small := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	if !effectstest.HasEffectiveKeyword(t, g, champ, "trample") {
		t.Error("printed trample did not reach the effective abilities")
	}
	hand := me.Hand.Size()
	b07BoltPlayer(t, g, opp.ID)
	if me.Hand.Size() != hand {
		t.Error("three damage is not five")
	}
	effectstest.AttackWith(t, g, opp.ID, champ, small)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("the Champion connects for 5, the Bear for 3: drew %d, want 1", me.Hand.Size()-hand)
	}
}

func TestB34DragonbornChampionIgnoresOpponentsSourcesAndCreatureDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.B12Push(g, me.ID, "Dragonborn Champion", "Creature — Dragon Warrior", effectstest.B34DragonbornChampionOracle, 5, 3)
	big := effectstest.B12Creature(g, me.ID, "Wurm", "Creature — Wurm", 6, 6)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Wurm", "Creature — Wurm", 6, 6)
	hand := me.Hand.Size()
	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(theirs, other.ID, 6)
		_ = g.DealDamageToCreatureForEffect(big, theirs, 6)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Error("an opponent's source, and damage to a creature, do not count")
	}
}

func TestB34KioraLootsOnEntryAndMakesTheScionAtThreshold(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kiora := effectstest.CastCatalogSpell(t, g, "Kiora, the Rising Tide", "Legendary Creature — Merfolk Noble", effectstest.B34KioraTheRisingTideOracle, nil)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("drew %d, want 2", me.Hand.Size()-hand)
	}
	if effectstest.DiscardOwed(g, me.ID) != 2 {
		t.Errorf("owes %d discards, want 2", effectstest.DiscardOwed(g, me.ID))
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	// Six cards in the graveyard: no threshold, no prompt. Kiora
	// attacks on the next turn, once the sickness has worn off. The
	// two cards just discarded are already there, so top up to six
	// rather than adding six (#651 — the loot's discard is really
	// paid now, where it used to be dropped on the floor).
	for me.Graveyard.Size() < 6 {
		effectstest.PushGraveyardCardForTest(me, "Filler")
	}
	if got := me.Graveyard.Size(); got != 6 {
		t.Fatalf("graveyard = %d, want exactly 6 for the no-threshold half", got)
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.AdvanceToMainOf(t, g, 0)
	effectstest.DeclareAttack(t, g, opp.ID, kiora)
	if effectstest.OwedMays(g, me.ID) > 0 {
		t.Fatal("six cards is not threshold")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToMainOf(t, g, 0)
	effectstest.PushGraveyardCardForTest(me, "Seventh")
	effectstest.DeclareAttack(t, g, opp.ID, kiora)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	scion := effectstest.BattlefieldIDNamed(g, me.ID, "Scion of the Deep")
	if scion == uuid.Nil {
		t.Fatal("threshold: the Scion is created")
	}
	if p, tt := effectstest.EffectivePower(t, g, scion), effectstest.EffectiveToughness(t, g, scion); p != 8 || tt != 8 {
		t.Errorf("the Scion is %d/%d, want 8/8", p, tt)
	}
	if c, _ := effectstest.BattlefieldCard(g, scion); !effects.IsLegendary(&c) {
		t.Error("the Scion is legendary")
	}
}

func TestB34CrosswayTroublemakersAttackingVampiresAndTheLifeDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	makers := effectstest.B12Push(g, me.ID, "Crossway Troublemakers", "Creature — Vampire", effectstest.B34CrosswayTroublemakersOrcl, 5, 5)
	vamp := effectstest.B12Creature(g, me.ID, "Vampire", "Creature — Vampire", 2, 2)
	home := effectstest.B12Creature(g, me.ID, "Home Vampire", "Creature — Vampire", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	life := me.Life
	effectstest.DeclareAttack(t, g, opp.ID, makers, vamp, bear)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{makers, vamp} {
		if !effectstest.HasEffectiveKeyword(t, g, id, "deathtouch") || !effectstest.HasEffectiveKeyword(t, g, id, "lifelink") {
			t.Error("an attacking Vampire you control has deathtouch and lifelink")
		}
	}
	for _, id := range []uuid.UUID{home, bear} {
		if effectstest.HasEffectiveKeyword(t, g, id, "deathtouch") || effectstest.HasEffectiveKeyword(t, g, id, "lifelink") {
			t.Error("a Vampire that stayed home and a non-Vampire do not")
		}
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	if me.Life != life+7 {
		t.Errorf("lifelink on 5 and 2: life %d → %d, want +7", life, me.Life)
	}
	// A Vampire dies: pay 2 life, draw a card.
	life, hand := me.Life, me.Hand.Size()
	effectstest.B27Kill(g, vamp)
	effectstest.PassPriorityAroundTable(t, g)
	// #796: the question is asked as the ability RESOLVES, not as the
	// trigger goes on the stack, so there is no CR 603.5 prompt.
	if effectstest.OwedMays(g, me.ID) > 0 {
		t.Error("the pay-2-life decision is made at resolution, not on announce")
	}
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-2 || me.Hand.Size() != hand+1 {
		t.Errorf("paid 2 and drew 1: life %d → %d, hand %d → %d", life, me.Life, hand, me.Hand.Size())
	}
	effectstest.B27Kill(g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm) != nil {
		t.Error("a Bear is not a Vampire")
	}
	effectstest.B27Kill(g, home)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-2 || me.Hand.Size() != hand+1 {
		t.Error("declining pays and draws nothing")
	}
}

func TestB34HobgoblinBanditLordPumpsGoblinsAndCountsThisTurnsEntries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lord := effectstest.B12Push(g, me.ID, "Hobgoblin Bandit Lord", "Creature — Goblin Rogue", effectstest.B34HobgoblinBanditLordOracle, 2, 3)
	old := effectstest.B12Creature(g, me.ID, "Old Goblin", "Creature — Goblin", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if p := effectstest.EffectivePower(t, g, old); p != 2 {
		t.Errorf("another Goblin you control is a %d/x, want 2", p)
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 1 {
		t.Errorf("an opponent's Goblin is a %d/x — \"you control\"", p)
	}
	if p := effectstest.EffectivePower(t, g, lord); p != 2 {
		t.Errorf("the Lord is a %d/x — \"other\"", p)
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("a Bear is a %d/x", p)
	}
	// The seeded Goblins entered before this turn's upkeep; two more
	// enter now, one as a token, and the changeling counts.
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "New Goblin", "Creature — Goblin", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.RedGoblinToken(), 1) })
	shifter := effectstest.HandCardFull(me, "Changeling", "Creature — Shapeshifter", "{1}", "", nil)
	for i := range me.Hand.Cards {
		if me.Hand.Cards[i].InstanceID == shifter {
			me.Hand.Cards[i].Keywords = []string{game.KeywordChangeling}
		}
	}
	if err := g.CastSpell(me.ID, shifter, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the changeling: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Elf", "Creature — Elf", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	life := opp.Life
	effectstest.B06AddMana(me, "R")
	effectstest.B16Activate(t, g, me.ID, lord, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetPlayer(opp.ID)})
	if opp.Life != life-3 {
		t.Errorf("three Goblins entered this turn: %d → %d, want -3", life, opp.Life)
	}
	if !effectstest.B16Tapped(t, g, lord) {
		t.Error("the tap is the cost")
	}
}

func TestB34ElanorMakesAFoodAndRampsAfterEatingOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Filler", TypeLine: "Sorcery"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Deep", TypeLine: "Sorcery"},
	)
	forest := ids[1]
	effectstest.CastCatalogSpell(t, g, "Elanor Gardner", "Legendary Creature — Halfling Scout", effectstest.B34ElanorGardnerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	food := effectstest.BattlefieldIDNamed(g, me.ID, "Food")
	if food == uuid.Nil {
		t.Fatal("Elanor makes a Food on entry")
	}
	// No Food eaten: the end step asks nothing.
	effectstest.AdvanceToEndStepOf(t, g, 0)
	if effectstest.OwedMays(g, me.ID) > 0 {
		t.Fatal("no Food sacrificed this turn, no prompt")
	}
	effectstest.PassPriorityAroundTable(t, g)
	// Next turn: eat the Food, then the end step offers the land.
	effectstest.AdvanceToMainOf(t, g, 0)
	effectstest.B06AddMana(me, "C", "C")
	effectstest.B16Activate(t, g, me.ID, food, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(food) {
		t.Fatal("the Food is sacrificed")
	}
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.SearchChoiceFor(g, me.ID); c != nil {
		effectstest.AnswerSearchByID(t, g, me.ID, forest)
	}
	if !g.Battlefield.Contains(forest) {
		t.Fatal("the basic is put onto the battlefield")
	}
	if !effectstest.B16Tapped(t, g, forest) {
		t.Error("it enters tapped")
	}
}

func TestB34SeizanEveryUpkeepCostsTwoLifeAndDrawsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Seizan, Perverter of Truth", "Legendary Creature — Demon Spirit", effectstest.B34SeizanOracle, 6, 5)
	life, hand := opp.Life, opp.Hand.Size()
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("the opponent's upkeep: life %d → %d, want -2", life, opp.Life)
	}
	if opp.Hand.Size() != hand+2 {
		t.Errorf("the opponent's upkeep: hand %d → %d, want +2", hand, opp.Hand.Size())
	}
	for seat := 2; seat < 4; seat++ {
		effectstest.AdvanceToUpkeepOf(t, g, seat)
		effectstest.PassPriorityAroundTable(t, g)
	}
	life, hand = me.Life, me.Hand.Size()
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-2 || me.Hand.Size() != hand+2 {
		t.Errorf("your own upkeep too: life %d → %d, hand %d → %d", life, me.Life, hand, me.Hand.Size())
	}
}

// --- enchantments and the Background --------------------------------

func TestB34OnWingsOfGoldLiftsZombiesAndTokensAndMakesAZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "On Wings of Gold", "Enchantment", effectstest.B34OnWingsOfGoldOracle, 0, 0)
	zombie := effectstest.B12Creature(g, me.ID, "Gravecrawler", "Creature — Zombie", 2, 2)
	token := b34Token(g, me.ID, "Soldier", "Token Creature — Soldier", 1, 1)
	both := b34Token(g, me.ID, "Zombie Token", "Token Creature — Zombie", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Zombie", "Creature — Zombie", 2, 2)
	for _, tc := range []struct {
		id   uuid.UUID
		want int
	}{{zombie, 3}, {token, 2}, {both, 3}} {
		if p := effectstest.EffectivePower(t, g, tc.id); p != tc.want {
			t.Errorf("a Zombie or token you control gets +1/+1: power %d, want %d", p, tc.want)
		}
		if !effectstest.HasEffectiveKeyword(t, g, tc.id, "flying") {
			t.Error("a Zombie or token you control has flying")
		}
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 2 || effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("a nontoken non-Zombie is untouched")
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 2 || effectstest.HasEffectiveKeyword(t, g, theirs, "flying") {
		t.Error("an opponent's Zombie is untouched")
	}
	// Cards leaving the graveyard: one Zombie per batch.
	a := effectstest.PushGraveyardCardForTest(me, "Dead A")
	b := effectstest.PushGraveyardCardForTest(me, "Dead B")
	g.WithWriteLock(func() { _ = g.ExileCardsForEffect([]uuid.UUID{a, b}) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Zombie"); n != 1 {
		t.Fatalf("\"one or more\": %d Zombie tokens for two cards leaving at once, want 1", n)
	}
	made := effectstest.BattlefieldIDNamed(g, me.ID, "Zombie")
	if p := effectstest.EffectivePower(t, g, made); p != 2 || !effectstest.HasEffectiveKeyword(t, g, made, "flying") {
		t.Error("the Zombie it makes is a 1/1 white Zombie token, lifted by its own anthem to a 2/2 flier")
	}
	theirsDead := effectstest.PushGraveyardCardForTest(opp, "Their Dead")
	b34ExileFromGraveyard(t, g, theirsDead)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Zombie"); n != 1 {
		t.Error("an opponent's graveyard is not yours")
	}
}

func TestB34TormodMakesATappedZombiePerBatchLeavingYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Tormod, the Desecrator", "Legendary Creature — Zombie Wizard", effectstest.B34TormodOracle, 4, 2)
	dead := effectstest.PushGraveyardCardForTest(me, "Dead")
	b34ExileFromGraveyard(t, g, dead)
	effectstest.PassPriorityAroundTable(t, g)
	zombie := effectstest.BattlefieldIDNamed(g, me.ID, "Zombie")
	if zombie == uuid.Nil {
		t.Fatal("a card leaving your graveyard makes a Zombie")
	}
	if !effectstest.B16Tapped(t, g, zombie) {
		t.Error("the Zombie enters tapped")
	}
	if p := effectstest.EffectivePower(t, g, zombie); p != 2 {
		t.Errorf("a 2/2: power %d", p)
	}
	back := effectstest.PushGraveyardCardForTest(me, "Regrowth Target")
	g.WithWriteLock(func() { _ = g.ReturnFromGraveyardForEffect(back, game.ZoneHand) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Zombie"); n != 2 {
		t.Errorf("a regrowth is a card leaving the graveyard: %d Zombies, want 2", n)
	}
}

func TestB34PollutedBondsDrainsAnOpponentPerLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Polluted Bonds", "Enchantment", effectstest.B34PollutedBondsOracle, 0, 0)
	mine := me.Life
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != mine {
		t.Fatal("your own land is not an opponent's")
	}
	life := opp.Life
	effectstest.B34PlayLandAs(t, g, 1, "Swamp", "Basic Land — Swamp")
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("the opponent's land: their life %d → %d, want -2", life, opp.Life)
	}
	if me.Life != mine+2 {
		t.Errorf("you gain 2: %d → %d", mine, me.Life)
	}
	// A land put onto the battlefield by an effect counts too.
	effectstest.B34LandEntersByEffect(t, g, opp, "Fetched Swamp", "Basic Land — Swamp")
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-4 || me.Life != mine+4 {
		t.Errorf("a fetched land: their %d, yours %d", opp.Life, me.Life)
	}
}

func TestB34GuildArtisanPaysTreasureWhenYourCommanderAttacksTheHealthiestPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.B12Push(g, me.ID, "Guild Artisan", "Legendary Enchantment — Background", effectstest.B34GuildArtisanOracle, 0, 0)
	commander := effectstest.B21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	// Everyone on 40: no opponent has MORE than the attacked player.
	// The trigger resolves in the declare-attackers step, before any
	// damage moves a life total.
	effectstest.DeclareAttack(t, g, opp.ID, commander, bear)
	if n := len(g.PendingTriggers) + len(g.StackMeta); n != 1 {
		t.Fatalf("one trigger for the commander, none for the Bear: %d", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure"); n != 2 {
		t.Fatalf("the commander attacking the (joint) healthiest opponent: %d Treasures, want 2", n)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	// Now the attacked player is behind another opponent: nothing.
	effectstest.AdvanceToMainOf(t, g, 0)
	if other.Life <= opp.Life {
		t.Fatalf("fixture: seat 2 (%d) should be ahead of seat 1 (%d)", other.Life, opp.Life)
	}
	effectstest.DeclareAttack(t, g, opp.ID, commander)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure"); n != 2 {
		t.Errorf("another opponent has more life than the attacked player: still 2 Treasures, got %d", n)
	}
	// Attacking the healthiest opponent pays again; a non-commander
	// never does.
	effectstest.AdvanceToMainOf(t, g, 0)
	effectstest.DeclareAttack(t, g, other.ID, commander)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure"); n != 4 {
		t.Errorf("the healthiest opponent attacked: %d Treasures, want 4", n)
	}
	effectstest.AdvanceToMainOf(t, g, 0)
	effectstest.DeclareAttack(t, g, other.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure"); n != 4 {
		t.Errorf("a Bear is not your commander: %d Treasures, want 4", n)
	}
}

func TestB34GuildArtisanIgnoresAnOpponentsCommander(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Guild Artisan", "Legendary Enchantment — Background", effectstest.B34GuildArtisanOracle, 0, 0)
	theirs := effectstest.B21Commander(g, opp.ID, "Their Commander", "Legendary Creature — Human", 3, 3)
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.DeclareAttack(t, g, g.Seats[2].ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure"); n != 0 {
		t.Errorf("an opponent's commander is not one you own: %d Treasures, want 0", n)
	}
}

func TestB34LordOfTheUndeadPumpsEveryOtherZombieAndRegrowsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lord := effectstest.B12Push(g, me.ID, "Lord of the Undead", "Creature — Zombie", effectstest.B34LordOfTheUndeadOracle, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Zombie", "Creature — Zombie", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Zombie", "Creature — Zombie", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if p := effectstest.EffectivePower(t, g, mine); p != 3 {
		t.Errorf("another Zombie you control is a %d/x, want 3", p)
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 3 {
		t.Errorf("an opponent's Zombie is a %d/x, want 3 — the printed card has no controller clause", p)
	}
	if p := effectstest.EffectivePower(t, g, lord); p != 2 {
		t.Errorf("the Lord is a %d/x — \"other\"", p)
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("a Bear is a %d/x", p)
	}
	dead := effectstest.B17GraveyardCard(me, "Dead Zombie", "Creature — Zombie", "{B}")
	kindred := effectstest.B17GraveyardCard(me, "Zombie Spell", "Kindred Sorcery — Zombie", "{B}")
	elf := effectstest.B17GraveyardCard(me, "Dead Elf", "Creature — Elf", "{G}")
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "B")
	if err := g.ActivateCatalogAbility(me.ID, lord, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(elf)}); err == nil {
		t.Fatal("an Elf is not a Zombie card")
	}
	effectstest.B16Activate(t, g, me.ID, lord, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(dead)})
	if !me.Hand.Contains(dead) {
		t.Error("the Zombie card returns to hand")
	}
	if !effectstest.B16Tapped(t, g, lord) {
		t.Error("the tap is the cost")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(lord) })
	effectstest.B06AddMana(me, "C", "B")
	effectstest.B16Activate(t, g, me.ID, lord, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(kindred)})
	if !me.Hand.Contains(kindred) {
		t.Error("a Kindred Zombie card is a Zombie card")
	}
}

// --- lands ----------------------------------------------------------

func TestB34CoriMountainMonasteryChecksAndImpulses(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tapped := effectstest.B12PlayFromHand(t, g, "Cori Mountain Monastery", "Land", effectstest.B34CoriMountainMonasteryOrcl, game.CastSpellParams{})
	if !effectstest.B16Tapped(t, g, tapped) {
		t.Fatal("with no Plains or Island it enters tapped")
	}
	effectstest.B12Permanent(g, me.ID, "Hallowed Fountain", "Land — Plains Island")
	effectstest.AdvanceToMainOf(t, g, 0)
	monastery := effectstest.B12PlayFromHand(t, g, "Cori Mountain Monastery", "Land", effectstest.B34CoriMountainMonasteryOrcl, game.CastSpellParams{})
	if effectstest.B16Tapped(t, g, monastery) {
		t.Fatal("with a Plains it enters untapped")
	}
	effectstest.B28TapForMana(t, g, me.ID, monastery, "")
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("tapped for {R}: pool %v", got)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(monastery) })
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Top Land", TypeLine: "Basic Land — Mountain"},
		game.Card{Name: "Deep Card", TypeLine: "Sorcery"},
	)
	effectstest.B06AddMana(me, "C", "C", "C", "R")
	effectstest.B16Activate(t, g, me.ID, monastery, 0, game.ActivateAbilityParams{})
	if !g.Exile.Contains(ids[0]) || !me.Library.Contains(ids[1]) {
		t.Fatal("the top card is exiled")
	}
	perm := effectstest.ExiledPermission(g, ids[0])
	if perm.Player != me.ID || perm.CastOnly || !effectstest.PermissionLive(g, perm, me.ID) {
		t.Errorf("%+v: the controller may PLAY it, a land included", perm)
	}
	// Through the opponents' turns and the controller's next turn the
	// grant holds; it lapses after that turn.
	effectstest.AdvanceToMainOf(t, g, 2)
	if !effectstest.PermissionLive(g, effectstest.ExiledPermission(g, ids[0]), me.ID) {
		t.Error("the grant survives the opponents' turns")
	}
	effectstest.AdvanceToMainOf(t, g, 0)
	if p := effectstest.ExiledPermission(g, ids[0]); !effectstest.PermissionLive(g, p, me.ID) {
		t.Errorf("on your next turn the grant is %+v, want live", p)
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	if effectstest.PermissionLive(g, effectstest.ExiledPermission(g, ids[0]), me.ID) {
		t.Error("the grant ends with your next turn")
	}
}

func TestB34BristlingBackwoodsEntersTappedPingsAnOpponentAndTapsForEitherColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[2]
	before := opp.Life
	land := effectstest.B12PlayFromHand(t, g, "Bristling Backwoods", "Land — Desert", effectstest.B34BristlingBackwoodsOracle, game.CastSpellParams{})
	if !effectstest.B16Tapped(t, g, land) {
		t.Fatal("the Desert enters tapped")
	}
	if p := effectstest.LatestPickTarget(g, me.ID); p == nil || effectstest.HasID(p.PickTargetPlayers, me.ID) || !effectstest.HasID(p.PickTargetPlayers, opp.ID) {
		t.Fatalf("the trigger asks for an opponent, never you: %+v", p)
	}
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("target opponent takes 1: %d → %d", before, opp.Life)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	effectstest.B28TapForMana(t, g, me.ID, land, "G")
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("tapped for {G}: pool %v", got)
	}
}

func TestB34TurbulentFenEntersUntappedOnlyAgainstEightOpposingLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	b34Lands(g, opp.ID, 4)
	b34Lands(g, other.ID, 3)
	b34Lands(g, me.ID, 5)
	tapped := effectstest.B12PlayFromHand(t, g, "Turbulent Fen", "Land — Swamp Forest", effectstest.B34TurbulentFenOracle, game.CastSpellParams{})
	if !effectstest.B16Tapped(t, g, tapped) {
		t.Fatal("seven opposing lands — your own do not count — is not eight: it enters tapped")
	}
	b34Lands(g, other.ID, 1)
	effectstest.AdvanceToMainOf(t, g, 0)
	fen := effectstest.B12PlayFromHand(t, g, "Turbulent Fen", "Land — Swamp Forest", effectstest.B34TurbulentFenOracle, game.CastSpellParams{})
	if effectstest.B16Tapped(t, g, fen) {
		t.Fatal("eight opposing lands between two opponents: it enters untapped")
	}
	effectstest.B28TapForMana(t, g, me.ID, fen, "B")
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "B" {
		t.Errorf("tapped for {B}: pool %v", got)
	}
}
