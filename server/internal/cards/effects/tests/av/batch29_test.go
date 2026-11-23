package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// b29Tapped reads a battlefield card's tapped state.
func b29Tapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Tapped
}

// b29AddMana adds coloured mana to a player's pool.
func b29AddMana(p *game.Player, colors ...string) {
	for _, c := range colors {
		p.ManaPool.AddMana(game.ManaToken{Color: c})
	}
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name.
func TestBatch29CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.B29BlackWaltzNo3Oracle:        "Black Waltz No. 3",
		effectstest.B29HeartlessHidetsuguOracle:   "Heartless Hidetsugu",
		effectstest.B29SettleTheWreckageOracle:    "Settle the Wreckage",
		effectstest.B29LegionLieutenantOracle:     "Legion Lieutenant",
		effectstest.B29SazhsChocoboOracle:         "Sazh's Chocobo",
		effectstest.B29YouthfulValkyrieOracle:     "Youthful Valkyrie",
		effectstest.B29ElvishPromenadeOracle:      "Elvish Promenade",
		effectstest.B29WorldMapOracle:             "World Map",
		effectstest.B29DazzlingAngelOracle:        "Dazzling Angel",
		effectstest.B29HammerOfPurphorosOracle:    "Hammer of Purphoros",
		effectstest.B29CorneredByBlackMagesOracle: "Cornered by Black Mages",
		effectstest.B29EmpyreanEagleOracle:        "Empyrean Eagle",
		effectstest.B29MassHysteriaOracle:         "Mass Hysteria",
		effectstest.B29BlowflyInfestationOracle:   "Blowfly Infestation",
		effectstest.B29ZombieApocalypseOracle:     "Zombie Apocalypse",
		effectstest.B29LightPawsOracle:            "Light-Paws, Emperor's Voice",
		effectstest.B29ZiatoraOracle:              "Ziatora, the Incinerator",
		effectstest.B29WitchsOvenOracle:           "Witch's Oven",
		effectstest.B29VedalkenArchmageOracle:     "Vedalken Archmage",
		effectstest.B29ScytheclawRaptorOracle:     "Scytheclaw Raptor",
		effectstest.B29NayaCharmOracle:            "Naya Charm",
		effectstest.B29HallowedHauntingOracle:     "Hallowed Haunting",
		effectstest.B29MountainValleyOracle:       "Mountain Valley",
		effectstest.B29YargleAndMultaniOracle:     "Yargle and Multani",
		effectstest.B29ChitterspitterOracle:       "Chitterspitter",
		effectstest.B29GimliOracle:                "Gimli of the Glittering Caves",
		// #929 shipped the resolution-time choose-a-player prompt and
		// the chosen player's own card pick, which is the pair Gluntch
		// was skipped for. It is a batch-29 card and is now registered.
		"0222dc7c-459b-4909-a037-72b2eb248599": "Gluntch, the Bestower",
	}
	if len(want) != 27 {
		t.Fatalf("the batch registers 27 cards, the table lists %d", len(want))
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
	// The declared skips must stay out until their seam lands:
	// a put-a-card-from-hand-onto-the-battlefield prompt (Walking
	// Atlas, Elvish Piper), and an opponent's three-way choice at
	// resolution (Master of Ceremonies). Power Conduit came off this list
	// in the S82 reprint pass (#5908): Scholar of New Horizons' any-kind
	// counter cost is the seam it was waiting for.
	//
	// Staff of the Storyteller came off this list in slice
	// fra-reprint-b: its counter is spent ("Remove a story counter:
	// Draw a card"), which RemoveCountersFromThis already pays, so it
	// was never the moving-a-counter shape.
	//
	// Ramos, Dragon Engine came off this list in #789, which gave a
	// MANA ability a counter cost — see ramos_dragon_engine.go. It was
	// never one of the batch's own 26, so it is asserted registered
	// here rather than in the table above.
	if _, ok := effects.Lookup("3ed41d2d-211b-4013-8562-8c64d54cc43a"); !ok {
		t.Error("Ramos, Dragon Engine is no longer skipped (#789) but is not registered")
	}
	for oracle, name := range map[string]string{
		"855a4837-7fc8-4b97-afbd-daa88e322c89": "Walking Atlas",
		"872ef617-7cdb-4a7a-85f5-efc9e12bece5": "Elvish Piper",
	} {
		if _, ok := effects.Lookup(oracle); ok {
			t.Errorf("%s is declared skipped on #391 but is registered — update the issue", name)
		}
	}
}

// --- the statics ---------------------------------------------------

func TestB29LegionLieutenantPumpsOtherVampiresYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lt := effectstest.B29Push(g, me.ID, "Legion Lieutenant", "Creature — Vampire Knight", effectstest.B29LegionLieutenantOracle, "{W}{B}", 2, 2, "W", "B")
	mine := effectstest.PushTribalCreature(g, me.ID, "Vampire Nighthawk", "Creature — Vampire Shaman", 2, 3)
	bear := effectstest.PushTribalCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.PushTribalCreature(g, opp.ID, "Their Vampire", "Creature — Vampire", 1, 1)
	if got := effectstest.EffectivePower(t, g, mine); got != 3 {
		t.Errorf("my Vampire: power %d, want 3", got)
	}
	if got := effectstest.EffectivePower(t, g, lt); got != 2 {
		t.Errorf("the Lieutenant itself: power %d, want 2 (\"other\")", got)
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Errorf("a non-Vampire: power %d, want 2", got)
	}
	if got := effectstest.EffectivePower(t, g, theirs); got != 1 {
		t.Errorf("an opponent's Vampire: power %d, want 1 (\"you control\")", got)
	}
}

func TestB29EmpyreanEaglePumpsOtherFliersYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eagle := effectstest.B29Push(g, me.ID, "Empyrean Eagle", "Creature — Bird Spirit", effectstest.B29EmpyreanEagleOracle, "{1}{W}{U}", 2, 3, "W", "U")
	flier := effectstest.PushTribalCreature(g, me.ID, "Bird", "Creature — Bird", 1, 1, "flying")
	ground := effectstest.PushTribalCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.PushTribalCreature(g, opp.ID, "Their Bird", "Creature — Bird", 1, 1, "flying")
	if got := effectstest.EffectivePower(t, g, flier); got != 2 {
		t.Errorf("my flier: power %d, want 2", got)
	}
	if got := effectstest.EffectivePower(t, g, eagle); got != 2 {
		t.Errorf("the Eagle itself: power %d, want 2 (\"other\")", got)
	}
	if got := effectstest.EffectivePower(t, g, ground); got != 2 {
		t.Errorf("a ground creature: power %d, want 2", got)
	}
	if got := effectstest.EffectivePower(t, g, theirs); got != 1 {
		t.Errorf("an opponent's flier: power %d, want 1", got)
	}
	// A creature that GAINS flying from a layer-6 grant is pumped too.
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(ground),
			[]game.Mod{game.AddKeywordsMod("flying")}, g.UntilEndOfTurnDuration(),
			"test — the ground creature gains flying")
	})
	if got := effectstest.EffectivePower(t, g, ground); got != 3 {
		t.Errorf("a creature granted flying: power %d, want 3", got)
	}
}

func TestB29MassHysteriaGivesEveryCreatureHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B29Push(g, me.ID, "Mass Hysteria", "Enchantment", effectstest.B29MassHysteriaOracle, "{R}", 0, 0, "R")
	mine := effectstest.CastAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "", true)
	g.BumpLayerVersionForTest()
	if !effectstest.HasEffectiveKeyword(t, g, mine, "haste") || !effectstest.HasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("all creatures have haste — yours and theirs")
	}
	if effectstest.SummoningSickOf(t, g, mine) {
		t.Error("a hasty creature is not summoning sick")
	}
}

func TestB29HammerOfPurphorosGrantsHasteAndForgesGolems(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hammer := effectstest.B29Push(g, me.ID, "Hammer of Purphoros", "Legendary Enchantment Artifact", effectstest.B29HammerOfPurphorosOracle, "{1}{R}{R}", 0, 0, "R")
	mine := effectstest.CastAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "", true)
	g.BumpLayerVersionForTest()
	if !effectstest.HasEffectiveKeyword(t, g, mine, "haste") {
		t.Error("creatures you control have haste")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("an opponent's creature does not")
	}
	land := effectstest.B29Push(g, me.ID, "Mountain", "Basic Land — Mountain", "", "", 0, 0)
	b29AddMana(me, "R", "R", "R")
	effectstest.B16Activate(t, g, me.ID, hammer, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{land}})
	if g.Battlefield.Contains(land) {
		t.Error("the land is sacrificed to pay")
	}
	golem := effectstest.FindBattlefieldByName(g, "Golem")
	if golem == uuid.Nil {
		t.Fatal("no Golem token")
	}
	c, _ := effectstest.BattlefieldCard(g, golem)
	if !c.IsCreature() || !c.IsArtifact() || !c.IsEnchantment() || c.Power != 3 || c.Toughness != 3 || len(c.Colors) != 0 {
		t.Errorf("the Golem is a 3/3 colorless enchantment artifact creature, got %s %d/%d %v", c.TypeLine, c.Power, c.Toughness, c.Colors)
	}
	if !b29Tapped(t, g, hammer) {
		t.Error("the Hammer taps to activate")
	}
}

func TestB29HallowedHauntingGatesAtSevenAndSizesItsSpirits(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B29Push(g, me.ID, "Hallowed Haunting", "Enchantment", effectstest.B29HallowedHauntingOracle, "{2}{W}{W}", 0, 0, "W")
	bear := effectstest.PushTribalCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for i := 0; i < 5; i++ {
		effectstest.B29Push(g, me.ID, "Anthem", "Enchantment", "", "{W}", 0, 0, "W")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("six enchantments: no flying yet")
	}
	// The seventh enchantment is cast, so the Spirit Cleric follows.
	effectstest.AdvanceToMain(t, g)
	effectstest.B29CastAs(t, g, me, "Seventh Seal", "Enchantment", "", "{W}", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") || !effectstest.HasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Error("seven enchantments: creatures you control have flying and vigilance")
	}
	cleric := effectstest.FindBattlefieldByName(g, "Spirit Cleric")
	if cleric == uuid.Nil {
		t.Fatal("casting an enchantment spell makes a Spirit Cleric")
	}
	if got := effectstest.EffectivePower(t, g, cleric); got != 1 {
		t.Errorf("one Spirit: the Cleric is %d/%d, want 1/1", got, effectstest.EffectiveToughness(t, g, cleric))
	}
	effectstest.PushTribalCreature(g, me.ID, "Spirit", "Creature — Spirit", 1, 1)
	if got := effectstest.EffectivePower(t, g, cleric); got != 2 {
		t.Errorf("two Spirits: the Cleric is %d/%d, want 2/2", got, effectstest.EffectiveToughness(t, g, cleric))
	}
	// A creature spell is not an enchantment spell.
	effectstest.CastAndResolveCreature(t, g, "Elf", "Creature — Elf", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B29CountNamed(g, me.ID, "Spirit Cleric"); got != 1 {
		t.Errorf("a creature spell: %d Clerics, want 1", got)
	}
	if spec, _ := effects.Lookup(effectstest.B29HallowedHauntingOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("the Spirit Cleric sizes itself, so the Haunting is complete")
	}
}

// The Spirit Cleric's size is the TOKEN's own printed ability (ADR 0083),
// so it keeps it when the last Haunting leaves, and a Cleric that gained and
// lost counters is still no 0/0 stand-in.
func TestB29SpiritClericKeepsItsSizeWhenTheHauntingLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	haunting := effectstest.B29Push(g, me.ID, "Hallowed Haunting", "Enchantment", effectstest.B29HallowedHauntingOracle, "{2}{W}{W}", 0, 0, "W")
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, effects.SpiritClericToken(), 2)
	})
	var clerics []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spirit Cleric" {
			clerics = append(clerics, c.InstanceID)
		}
	}
	if len(clerics) != 2 {
		t.Fatalf("want 2 Spirit Clerics, got %d", len(clerics))
	}
	spent, fresh := clerics[0], clerics[1]

	effectstest.GainAndLoseACounter(t, g, spent)
	if !g.Battlefield.Contains(spent) {
		t.Fatal("a Cleric survives losing its last counter while it has a size")
	}
	if got := effectstest.EffectiveToughness(t, g, spent); got != 2 {
		t.Errorf("two Spirits: toughness %d, want 2", got)
	}

	effectstest.B15Destroy(t, g, haunting)
	effectstest.RunStateChecksViaDraw(t, g)
	for _, id := range []uuid.UUID{spent, fresh} {
		if !g.Battlefield.Contains(id) {
			t.Fatal("a Spirit Cleric died when the Haunting left; it carries its own size")
		}
		if got := effectstest.EffectiveToughness(t, g, id); got != 2 {
			t.Errorf("the Haunting is gone: toughness %d, want 2 (two Spirits)", got)
		}
	}
}

func TestB29ChitterspitterEatsATokenEachUpkeepAndGrowsSquirrels(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spitter := effectstest.B29Push(g, me.ID, "Chitterspitter", "Artifact", effectstest.B29ChitterspitterOracle, "{2}{G}", 0, 0, "G")
	squirrel := effectstest.PushTribalCreature(g, me.ID, "Drey Keeper", "Creature — Squirrel", 1, 1)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 green Squirrel"), 1) })
	token := effectstest.FindBattlefieldByName(g, "Squirrel")
	if got := effectstest.EffectivePower(t, g, squirrel); got != 1 {
		t.Fatalf("no acorns: power %d, want 1", got)
	}
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID); prompt == nil || prompt.ChooseMin != 0 || effectstest.HasID(prompt.ChooseCards, squirrel) || !effectstest.HasID(prompt.ChooseCards, token) {
		t.Fatal("only tokens are offered, as the trigger resolves, with a floor of zero for the \"may\"")
	}
	effectstest.AnswerOwnPermanents(t, g, me.ID, token)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(token) {
		t.Error("the chosen token is sacrificed")
	}
	if got := effectstest.CounterCount(g, spitter, "acorn"); got != 1 {
		t.Errorf("%d acorn counters, want 1", got)
	}
	if got := effectstest.EffectivePower(t, g, squirrel); got != 2 {
		t.Errorf("one acorn: Squirrel power %d, want 2", got)
	}
	// {G}, {T}: a Squirrel, which is pumped too.
	effectstest.AdvanceToMain(t, g)
	b29AddMana(me, "G")
	effectstest.B16Activate(t, g, me.ID, spitter, 0, game.ActivateAbilityParams{})
	made := effectstest.FindBattlefieldByName(g, "Squirrel")
	if made == uuid.Nil {
		t.Fatal("the activation makes a Squirrel")
	}
	if got := effectstest.EffectivePower(t, g, made); got != 2 {
		t.Errorf("the new Squirrel: power %d, want 2", got)
	}
}

func TestB29ChitterspitterWithNoTokenIsSilent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B29Push(g, me.ID, "Chitterspitter", "Artifact", effectstest.B29ChitterspitterOracle, "{2}{G}", 0, 0, "G")
	effectstest.PushTribalCreature(g, me.ID, "Drey Keeper", "Creature — Squirrel", 1, 1)
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OwedMays(g, me.ID) > 0 || effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("no token: no prompt at all (CR 603.3d)")
	}
}

// --- the triggers --------------------------------------------------

func TestB29BlackWaltzPingsTheTableOnNoncreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	waltz := effectstest.B29Push(g, me.ID, "Black Waltz No. 3", "Legendary Creature — Wizard", effectstest.B29BlackWaltzNo3Oracle, "{2}{B}{R}", 2, 2, "B", "R")
	if !effectstest.HasEffectiveKeyword(t, g, waltz, "flying") || !effectstest.HasEffectiveKeyword(t, g, waltz, "deathtouch") {
		t.Error("flying and deathtouch are printed")
	}
	before := effectstest.B29Lives(g)
	effectstest.B24CastNoncreature(t, g, "Shock", "Instant", "{R}")
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		want := before[i] - 2
		if i == 0 {
			want = before[i]
		}
		if p.Life != want {
			t.Errorf("seat %d: life %d, want %d", i, p.Life, want)
		}
	}
	mid := effectstest.B29Lives(g)
	effectstest.CastAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.Life != mid[i] {
			t.Errorf("a creature spell: seat %d life %d, want %d", i, p.Life, mid[i])
		}
	}
}

func TestB29SazhsChocoboGrowsOnLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	chocobo := effectstest.B29Push(g, me.ID, "Sazh's Chocobo", "Creature — Bird", effectstest.B29SazhsChocoboOracle, "{G}", 0, 1, "G")
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, chocobo, "+1/+1"); got != 1 {
		t.Errorf("a land drop: %d +1/+1 counters, want 1", got)
	}
	effectstest.CastAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, chocobo, "+1/+1"); got != 1 {
		t.Errorf("a creature entering: %d counters, want still 1", got)
	}
}

func TestB29YouthfulValkyrieGrowsWhenAnotherAngelEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	valkyrie := effectstest.CastAndResolveCreature(t, g, "Youthful Valkyrie", "Creature — Angel", effectstest.B29YouthfulValkyrieOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, valkyrie, "+1/+1"); got != 0 {
		t.Fatalf("its own entry: %d counters, want 0 (\"another\")", got)
	}
	effectstest.CastAndResolveCreature(t, g, "Serra Angel", "Creature — Angel", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, valkyrie, "+1/+1"); got != 1 {
		t.Errorf("another Angel: %d counters, want 1", got)
	}
	effectstest.CastAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, valkyrie, "+1/+1"); got != 1 {
		t.Errorf("a non-Angel: %d counters, want still 1", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, valkyrie, "flying") {
		t.Error("flying is printed")
	}
}

func TestB29DazzlingAngelGainsOnOtherCreaturesEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	start := me.Life
	angel := effectstest.CastAndResolveCreature(t, g, "Dazzling Angel", "Creature — Angel", effectstest.B29DazzlingAngelOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != start {
		t.Fatalf("its own entry: life %d, want %d (\"another\")", me.Life, start)
	}
	effectstest.CastAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != start+1 {
		t.Errorf("a creature entering: life %d, want %d", me.Life, start+1)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 white Spirit with flying"), 2) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != start+3 {
		t.Errorf("two tokens: life %d, want %d", me.Life, start+3)
	}
	if !effectstest.HasEffectiveKeyword(t, g, angel, "flying") {
		t.Error("flying is printed")
	}
}

func TestB29VedalkenArchmageDrawsOnArtifactSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B29Push(g, me.ID, "Vedalken Archmage", "Creature — Vedalken Wizard", effectstest.B29VedalkenArchmageOracle, "{2}{U}{U}", 0, 2, "U")
	hand := me.Hand.Size()
	effectstest.B24CastNoncreature(t, g, "Sol Ring", "Artifact", "{1}")
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("an artifact spell: hand %d → %d, want +1", hand, got)
	}
	hand = me.Hand.Size()
	effectstest.B24CastNoncreature(t, g, "Shock", "Instant", "{R}")
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("an instant: hand %d → %d, want unchanged", hand, got)
	}
}

func TestB29ScytheclawRaptorPunishesOffTurnCasts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B29Push(g, me.ID, "Scytheclaw Raptor", "Creature — Dinosaur", effectstest.B29ScytheclawRaptorOracle, "{2}{R}", 4, 3, "R")
	effectstest.AdvanceToMain(t, g)
	before := opp.Life
	effectstest.B29CastAs(t, g, opp, "Opt", "Instant", "", "{U}", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-4 {
		t.Errorf("an opponent casting on my turn: life %d, want %d", opp.Life, before-4)
	}
	mine := me.Life
	effectstest.B24CastNoncreature(t, g, "Shock", "Instant", "{R}")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != mine {
		t.Errorf("casting on my own turn: life %d, want %d", me.Life, mine)
	}
}

func TestB29GimliGrowsOnLegendsAndMakesTreasureOnConnecting(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gimli := effectstest.B29Push(g, me.ID, "Gimli of the Glittering Caves", "Legendary Creature — Dwarf Warrior", effectstest.B29GimliOracle, "{2}{R}", 1, 1, "R")
	if !effectstest.HasEffectiveKeyword(t, g, gimli, "double strike") {
		t.Error("double strike is printed")
	}
	effectstest.CastAndResolveCreature(t, g, "Legolas", "Legendary Creature — Elf Archer", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, gimli, "+1/+1"); got != 1 {
		t.Errorf("a legendary creature: %d counters, want 1", got)
	}
	effectstest.CastAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, gimli, "+1/+1"); got != 1 {
		t.Errorf("a nonlegendary creature: %d counters, want still 1", got)
	}
	effectstest.AttackWith(t, g, opp.ID, gimli)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B29CountNamed(g, me.ID, "Treasure"); got != 2 {
		t.Errorf("Gimli connecting with double strike: %d Treasures, want 2 (one per damage step)", got)
	}
	if opp.Life != 36 {
		t.Errorf("a 2/2 double striker: opponent life %d, want 36", opp.Life)
	}
}

func TestB29BlowflyInfestationChainsMinusCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B29Push(g, me.ID, "Blowfly Infestation", "Enchantment", effectstest.B29BlowflyInfestationOracle, "{2}{B}", 0, 0, "B")
	infested := effectstest.PushCounterCreature(g, opp.ID, "Infested", game.CounterMinusOne, 1)
	victim := effectstest.PushCounterCreature(g, opp.ID, "Victim", "", 0)
	clean := effectstest.PushCounterCreature(g, me.ID, "Clean", "", 0)
	effectstest.B27Kill(g, infested)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, victim, game.CounterMinusOne); got != 1 {
		t.Errorf("target creature: %d -1/-1 counters, want 1", got)
	}
	// A creature that dies WITHOUT a -1/-1 counter is not a trigger.
	effectstest.B27Kill(g, clean)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a clean death: no trigger")
	}
	if got := effectstest.CounterCount(g, victim, game.CounterMinusOne); got != 1 {
		t.Errorf("still %d counters, want 1", got)
	}
	// The countered creature dying chains: this time onto my own.
	mine := effectstest.PushCounterCreature(g, me.ID, "Mine", "", 0)
	effectstest.B27Kill(g, victim)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, mine, game.CounterMinusOne); got != 1 {
		t.Errorf("the chain: %d -1/-1 counters, want 1", got)
	}
}

func TestB29LightPawsFetchesASmallerAuraOntoItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	paws := effectstest.B29Push(g, me.ID, "Light-Paws, Emperor's Voice", "Legendary Creature — Fox Advisor", effectstest.B29LightPawsOracle, "{1}{W}", 2, 2, "W")
	bear := effectstest.PushTribalCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rancor := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Rancor", TypeLine: effectstest.AuraTypeLine, OracleID: effectstest.RancorOracle, ManaCost: "{G}", Colors: []string{"G"}})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Small Aura", TypeLine: effectstest.AuraTypeLine, ManaCost: "{1}{W}", Colors: []string{"W"}})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Big Aura", TypeLine: effectstest.AuraTypeLine, ManaCost: "{2}{W}{W}", Colors: []string{"W"}})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Ethereal Armor", TypeLine: effectstest.AuraTypeLine, ManaCost: "{W}", Colors: []string{"W"}})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})
	effectstest.AdvanceToMain(t, g)
	effectstest.B29CastAs(t, g, me, "Ethereal Armor", effectstest.AuraTypeLine, "", "{W}{W}", game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Rancor") == uuid.Nil || effectstest.SearchOptionNamed(g, c, "Small Aura") == uuid.Nil {
		t.Error("Rancor and Small Aura (mana value ≤ 2, different names) are offered")
	}
	if effectstest.SearchOptionNamed(g, c, "Big Aura") != uuid.Nil {
		t.Error("an Aura with mana value 4 is not offered")
	}
	if effectstest.SearchOptionNamed(g, c, "Ethereal Armor") != uuid.Nil {
		t.Error("an Aura sharing a name with one you control is not offered")
	}
	if effectstest.SearchOptionNamed(g, c, "Plains") != uuid.Nil {
		t.Error("a land is not an Aura")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, rancor)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rancor) {
		t.Fatal("Rancor is put onto the battlefield")
	}
	if host := effectstest.AttachmentHostOf(t, g, rancor); host.Kind != game.TargetCard || host.ID != paws {
		t.Fatalf("Rancor is attached to %+v, want Light-Paws", host)
	}
	if got := effectstest.EffectivePower(t, g, paws); got != 4 {
		t.Errorf("Light-Paws wearing Rancor: power %d, want 4", got)
	}
	// The fetched Aura was not cast, so it does not chain.
	if effectstest.OwedMays(g, me.ID) > 0 {
		t.Error("a fetched Aura is not \"if you cast it\"")
	}
}

func TestB29ZiatoraFlingsAtTheEndStepThroughAReflexiveTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ziatora := effectstest.B29Push(g, me.ID, "Ziatora, the Incinerator", "Legendary Creature — Demon Dragon", effectstest.B29ZiatoraOracle, "{3}{B}{R}{G}", 6, 6, "B", "R", "G")
	fodder := effectstest.PushCounterCreature(g, me.ID, "Fodder", game.CounterPlusOne, 2)
	if !effectstest.HasEffectiveKeyword(t, g, ziatora, "flying") {
		t.Error("flying is printed")
	}
	before := opp.Life
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if p == nil || p.ChooseMin != 0 {
		t.Fatal("the sacrifice is asked as the trigger resolves, with a floor of zero for the \"may\"")
	}
	for _, id := range p.ChooseCards {
		if id == ziatora {
			t.Error("Ziatora herself is not \"another creature\"")
		}
	}
	effectstest.AnswerOwnPermanents(t, g, me.ID, fodder)
	// The end-step trigger resolves: the sacrifice, then the
	// reflexive trigger asks for its target.
	effectstest.B04WaitForPick(t, g, me.ID)
	if g.Battlefield.Contains(fodder) {
		t.Fatal("the chosen creature is sacrificed before the reflexive trigger fires")
	}
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-6 {
		t.Errorf("a 4/4 with two +1/+1 counters: opponent life %d, want %d (power 6)", opp.Life, before-6)
	}
	if got := effectstest.B29CountNamed(g, me.ID, "Treasure"); got != 3 {
		t.Errorf("%d Treasures, want 3", got)
	}
}

// The damage is the sacrificed creature's last-known power, so an
// anthem counts: a 1/1 under a Glorious Anthem is a 2-power Fling.
func TestB29ZiatoraCountsAnAnthemInTheDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B29Push(g, me.ID, "Ziatora, the Incinerator", "Legendary Creature — Demon Dragon", effectstest.B29ZiatoraOracle, "{3}{B}{R}{G}", 6, 6, "B", "R", "G")
	fodder := effectstest.B16Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1, "R")
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Glorious Anthem", TypeLine: "Enchantment",
		OracleID: effectstest.GloriousAnthemOracle, Owner: me.ID, Controller: me.ID,
	})
	before := opp.Life
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID, fodder)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("a 1/1 under an anthem: opponent life %d, want %d (power 2)", opp.Life, before-2)
	}
}

func TestB29ZiatoraDeclinedOrAloneDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B29Push(g, me.ID, "Ziatora, the Incinerator", "Legendary Creature — Demon Dragon", effectstest.B29ZiatoraOracle, "{3}{B}{R}{G}", 6, 6, "B", "R", "G")
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OwedMays(g, me.ID) > 0 || effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("no other creature: no prompt at all (CR 603.3d)")
	}
	if got := effectstest.B29CountNamed(g, me.ID, "Treasure"); got != 0 {
		t.Errorf("%d Treasures, want 0", got)
	}
}

// --- the activations -----------------------------------------------

func TestB29HeartlessHidetsuguHalvesEveryLifeTotal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hidetsugu := effectstest.B29Push(g, me.ID, "Heartless Hidetsugu", "Legendary Creature — Ogre Shaman", effectstest.B29HeartlessHidetsuguOracle, "{3}{R}{R}", 4, 3, "R")
	g.Seats[0].Life, g.Seats[1].Life, g.Seats[2].Life, g.Seats[3].Life = 40, 21, 7, 1
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, hidetsugu, 0, game.ActivateAbilityParams{})
	if got := effectstest.B29Lives(g); got[0] != 20 || got[1] != 11 || got[2] != 4 || got[3] != 1 {
		t.Errorf("lives %v, want [20 11 4 1]", got)
	}
	if !b29Tapped(t, g, hidetsugu) {
		t.Error("Hidetsugu taps")
	}
}

func TestB29WorldMapTutorsABasicOrAnyLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	worldMap := effectstest.B29Push(g, me.ID, "World Map", "Artifact", effectstest.B29WorldMapOracle, "{1}", 0, 0)
	forest := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Cabal Coffers", TypeLine: "Land"})
	effectstest.AdvanceToMain(t, g)
	b29AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, worldMap, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate {1}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Forest") == uuid.Nil || effectstest.SearchOptionNamed(g, c, "Plains") == uuid.Nil || effectstest.SearchOptionNamed(g, c, "Cabal Coffers") != uuid.Nil {
		t.Error("the {1} mode offers basics only")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, forest)
	if !me.Hand.Contains(forest) {
		t.Error("the Forest is put into hand")
	}
	if g.Battlefield.Contains(worldMap) {
		t.Error("the Map is sacrificed")
	}
	// The {3} mode offers any land.
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	map2 := effectstest.B29Push(g2, me2.ID, "World Map", "Artifact", effectstest.B29WorldMapOracle, "{1}", 0, 0)
	effectstest.PushLibraryCardForTest(me2, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	coffers := effectstest.PushLibraryCardForTest(me2, game.Card{Name: "Cabal Coffers", TypeLine: "Land"})
	effectstest.AdvanceToMain(t, g2)
	b29AddMana(me2, "C", "C", "C")
	if err := g2.ActivateCatalogAbility(me2.ID, map2, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate {3}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if c := effectstest.SearchChoiceFor(g2, me2.ID); c == nil || effectstest.SearchOptionNamed(g2, c, "Cabal Coffers") == uuid.Nil {
		t.Fatal("the {3} mode offers a nonbasic land")
	}
	effectstest.AnswerSearchByID(t, g2, me2.ID, coffers)
	if !me2.Hand.Contains(coffers) {
		t.Error("the {3} mode finds a nonbasic land")
	}
}

func TestB29WitchsOvenBakesOneOrTwoFood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oven := effectstest.B29Push(g, me.ID, "Witch's Oven", "Artifact", effectstest.B29WitchsOvenOracle, "{1}", 0, 0)
	small := effectstest.PushTribalCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	big := effectstest.PushTribalCreature(g, me.ID, "Rhino", "Creature — Rhino", 4, 4)
	grown := effectstest.PushCounterCreature(g, me.ID, "Grown", game.CounterPlusOne, 1) // a 4/4 base pushed as 4/4 — see below
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, oven, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{small}})
	if got := effectstest.B29CountNamed(g, me.ID, "Food"); got != 1 {
		t.Errorf("a 2/2: %d Food, want 1", got)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(oven) })
	effectstest.B16Activate(t, g, me.ID, oven, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{big}})
	if got := effectstest.B29CountNamed(g, me.ID, "Food"); got != 3 {
		t.Errorf("a 4/4: %d Food, want 3 (two more)", got)
	}
	// Counters count: pushCounterCreature is a 4/4 base, so shrink it
	// to a 3/3 base with a +1/+1 counter on top before cooking it.
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == grown {
			g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 3, 3
		}
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(oven) })
	effectstest.B16Activate(t, g, me.ID, oven, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{grown}})
	if got := effectstest.B29CountNamed(g, me.ID, "Food"); got != 5 {
		t.Errorf("a 3/3 with a +1/+1 counter: %d Food, want 5 (two more)", got)
	}
	// A bonus from another permanent counts: a 3/3 under a +1/+1 anthem
	// had toughness 4 as it last stood (CR 603.10, #3172).
	effectstest.PushCatalogPermanent(g, me.ID, "Glorious Anthem", "Enchantment", effectstest.GloriousAnthemOracle, false)
	boosted := effectstest.PushTribalCreature(g, me.ID, "Boosted", "Creature — Bear", 3, 3)
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(oven) })
	before := effectstest.B29CountNamed(g, me.ID, "Food")
	effectstest.B16Activate(t, g, me.ID, oven, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{boosted}})
	if got := effectstest.B29CountNamed(g, me.ID, "Food"); got != before+2 {
		t.Errorf("a 3/3 under an anthem: %d Food, want %d (two more)", got, before+2)
	}
	if spec, _ := effects.Lookup(effectstest.B29WitchsOvenOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("the card is complete: the last-known toughness is post-layer")
	}
}

func TestB29MountainValleyEntersTappedAndFetchesAMountainOrForest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	ground := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Stomping Ground", TypeLine: "Land — Mountain Forest"})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	valley := effectstest.PlayLandFromHand(t, g, "Mountain Valley", effectstest.B29MountainValleyOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if !b29Tapped(t, g, valley) {
		t.Fatal("Mountain Valley enters tapped")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(valley) })
	if err := g.ActivateCatalogAbility(me.ID, valley, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Mountain") == uuid.Nil || effectstest.SearchOptionNamed(g, c, "Forest") == uuid.Nil || effectstest.SearchOptionNamed(g, c, "Stomping Ground") == uuid.Nil {
		t.Error("every Mountain or Forest card is offered, nonbasics included")
	}
	if effectstest.SearchOptionNamed(g, c, "Island") != uuid.Nil {
		t.Error("an Island is not offered")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, ground)
	if !g.Battlefield.Contains(ground) {
		t.Fatal("the chosen land is put onto the battlefield")
	}
	if b29Tapped(t, g, ground) {
		t.Error("a Mirage fetch puts the land in untapped")
	}
	if g.Battlefield.Contains(valley) {
		t.Error("the Valley is sacrificed")
	}
}

// --- the spells ----------------------------------------------------

func TestB29SettleTheWreckageExilesAttackersAndTheyFetchBasics(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.PushVanillaCreature(g, me.ID, "Attacker A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "Attacker B", 3, 3)
	home := effectstest.PushVanillaCreature(g, me.ID, "Stay-at-home", 1, 1)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	f1 := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	f2 := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})
	f3 := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Swamp", TypeLine: "Basic Land — Swamp"})
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.B29CastAs(t, g, opp, "Settle the Wreckage", "Instant", effectstest.B29SettleTheWreckageOracle, "{2}{W}{W}", game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(a) || !g.Exile.Contains(b) {
		t.Fatal("both attackers are exiled")
	}
	if !g.Battlefield.Contains(home) || !g.Battlefield.Contains(theirs) {
		t.Error("a creature that did not attack, and the opponent's, stay")
	}
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the attacked player gets the search prompt")
	}
	if len(c.SearchCards) != 3 {
		t.Errorf("%d basics offered, want 3", len(c.SearchCards))
	}
	effectstest.AnswerSearchByID(t, g, me.ID, f1, f2)
	if !g.Battlefield.Contains(f1) || !g.Battlefield.Contains(f2) {
		t.Fatal("two basics are put onto the battlefield")
	}
	if !b29Tapped(t, g, f1) || !b29Tapped(t, g, f2) {
		t.Error("they enter tapped")
	}
	if g.Battlefield.Contains(f3) {
		t.Error("only as many as were exiled")
	}
}

// TestB29SettleTheWreckageSearchesForWhatLandedInExile pins #866 on
// the card that reads the count. Since ADR 0115 an attacking COMMANDER
// is exiled with the rest of the attackers (CR 903.9a asks its owner
// only afterwards), so it was exiled this way and buys a basic land too,
// whatever its owner then answers.
func TestB29SettleTheWreckageSearchesForWhatLandedInExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Attacker", 2, 2)
	cmd := effectstest.B21Commander(g, me.ID, "Their Commander", "Legendary Creature — Human", 3, 3)
	f1 := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	f2 := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})
	f3 := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	effectstest.DeclareAttack(t, g, opp.ID, bear, cmd)
	effectstest.B29CastAs(t, g, opp, "Settle the Wreckage", "Instant", effectstest.B29SettleTheWreckageOracle, "{2}{W}{W}", game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	})
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Exile.Contains(bear) || !g.Exile.Contains(cmd) {
		t.Fatal("both attackers are exiled, the commander included")
	}
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search for what was exiled")
	}
	if c.SearchMax != 2 {
		t.Errorf("the search is for %d basics, want 2 — the commander was exiled this way too", c.SearchMax)
	}
	effectstest.AnswerSearchByID(t, g, me.ID, f1, f2)
	if !g.Battlefield.Contains(f1) || !g.Battlefield.Contains(f2) {
		t.Fatal("the chosen basics are put onto the battlefield")
	}
	if g.Battlefield.Contains(f3) {
		t.Error("only as many as were exiled")
	}

	effectstest.AnswerCommanderReturn(t, g, me.ID, true)
	if !me.Command.Contains(cmd) {
		t.Error("the commander is in its owner's command zone")
	}
}

func TestB29ElvishPromenadeMakesAnElfWarriorPerElf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushTribalCreature(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	effectstest.PushTribalCreature(g, me.ID, "Elvish Mystic", "Creature — Elf Druid", 1, 1)
	effectstest.PushTribalCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.PushTribalCreature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	effectstest.CastCatalogSpell(t, g, "Elvish Promenade", "Kindred Sorcery — Elf", effectstest.B29ElvishPromenadeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B29CountNamed(g, me.ID, "Elf Warrior"); got != 2 {
		t.Errorf("two Elves: %d Elf Warriors, want 2", got)
	}
	if got := effectstest.B29CountNamed(g, opp.ID, "Elf Warrior"); got != 0 {
		t.Errorf("the opponent gets none, got %d", got)
	}
}

func TestB29CorneredByBlackMagesEdictsAndLeavesAWizard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	keep := effectstest.PushVanillaCreature(g, opp.ID, "Keeper", 3, 3)
	fodder := effectstest.PushVanillaCreature(g, opp.ID, "Fodder", 1, 1)
	effectstest.CastCatalogSpell(t, g, "Cornered by Black Mages", "Sorcery", effectstest.B29CorneredByBlackMagesOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SacrificeChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("the opponent gets a sacrifice prompt")
	}
	if len(effectstest.SacrificeOptions(c)) != 2 {
		t.Errorf("%d options, want the opponent's 2 creatures", len(effectstest.SacrificeOptions(c)))
	}
	effectstest.AnswerSacrifice(t, g, opp.ID, fodder)
	if g.Battlefield.Contains(fodder) || !g.Battlefield.Contains(keep) {
		t.Error("the opponent's choice is sacrificed, the other stays")
	}
	wizard := effectstest.FindBattlefieldByName(g, "Wizard")
	if wizard == uuid.Nil {
		t.Fatal("no Wizard token")
	}
	w, _ := effectstest.BattlefieldCard(g, wizard)
	if w.Controller != me.ID || w.Power != 0 || w.Toughness != 1 || len(w.Colors) != 1 || w.Colors[0] != "B" {
		t.Errorf("the Wizard is my 0/1 black token, got %d/%d %v", w.Power, w.Toughness, w.Colors)
	}
	// The Wizard's own "whenever you cast a noncreature spell" ping
	// ships since ADR 0083 (#1248) — this used to assert the caveat
	// that declared it missing. Its behaviour is exercised in
	// token_abilities_cards_test.go.
	if got := len(game.TriggersForCard(w)); got != 1 {
		t.Errorf("the Wizard has %d triggered abilities, want its printed ping", got)
	}
	if spec, _ := effects.Lookup(effectstest.B29CorneredByBlackMagesOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("nothing about this card is simplified any more")
	}
}

func TestB29ZombieApocalypseRaisesZombiesTappedThenKillsHumans(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zombie := effectstest.PushGraveyardCardTyped(me, "Gravecrawler", "Creature — Zombie")
	zombieHuman := effectstest.PushGraveyardCardTyped(me, "Zombie Villager", "Creature — Zombie Human")
	elf := effectstest.PushGraveyardCardTyped(me, "Dead Elf", "Creature — Elf")
	zombieLand := effectstest.PushGraveyardCardTyped(me, "Zombie Land", "Land — Zombie")
	theirZombie := effectstest.PushGraveyardCardTyped(opp, "Their Zombie", "Creature — Zombie")
	myHuman := effectstest.PushVanillaCreature(g, me.ID, "My Human", 2, 2)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == myHuman {
			g.Battlefield.Cards[i].TypeLine = "Creature — Human Soldier"
		}
	}
	theirHuman := effectstest.PushTribalCreature(g, opp.ID, "Their Human", "Creature — Human Wizard", 1, 1)
	theirElf := effectstest.PushTribalCreature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	effectstest.CastCatalogSpell(t, g, "Zombie Apocalypse", "Sorcery", effectstest.B29ZombieApocalypseOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(zombie) {
		t.Fatal("the Zombie creature card returns")
	}
	if !b29Tapped(t, g, zombie) {
		t.Error("it returns tapped")
	}
	if g.Battlefield.Contains(elf) || g.Battlefield.Contains(zombieLand) || g.Battlefield.Contains(theirZombie) {
		t.Error("a non-Zombie, a non-creature Zombie card and an opponent's Zombie stay put")
	}
	if g.Battlefield.Contains(zombieHuman) {
		t.Error("a Zombie Human returns and is then destroyed")
	}
	if g.Battlefield.Contains(myHuman) || g.Battlefield.Contains(theirHuman) {
		t.Error("every Human is destroyed, mine and theirs")
	}
	if !g.Battlefield.Contains(theirElf) {
		t.Error("a non-Human survives")
	}
}

func TestB29NayaCharmIsChooseOneAndResolvesEachMode(t *testing.T) {
	// Mode 0: 3 damage to target creature.
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	rhino := effectstest.PushVanillaCreature(g, opp.ID, "Rhino", 3, 3)
	effectstest.CastModal(t, g, "Naya Charm", "Instant", effectstest.B29NayaCharmOracle, []int{0}, []game.TargetRef{{Kind: game.TargetCard, ID: rhino}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rhino) {
		t.Error("mode 0: a 3/3 dies to 3 damage")
	}
	// Mode 1: return target card from ANY graveyard to its owner's hand.
	g = effectstest.NewCatalogGame(t)
	opp = g.Seats[1]
	dead := effectstest.PushGraveyardCardTyped(opp, "Their Dead Bear", "Creature — Bear")
	hand := opp.Hand.Size()
	effectstest.CastModal(t, g, "Naya Charm", "Instant", effectstest.B29NayaCharmOracle, []int{1}, []game.TargetRef{{Kind: game.TargetCard, ID: dead}})
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(dead) || opp.Hand.Size() != hand+1 {
		t.Error("mode 1: the card goes to its OWNER's hand")
	}
	// Mode 2: tap all creatures target player controls.
	g = effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.PushVanillaCreature(g, opp.ID, "A", 1, 1)
	b := effectstest.PushVanillaCreature(g, opp.ID, "B", 1, 1)
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 1, 1)
	effectstest.CastModal(t, g, "Naya Charm", "Instant", effectstest.B29NayaCharmOracle, []int{2}, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if !b29Tapped(t, g, a) || !b29Tapped(t, g, b) {
		t.Error("mode 2: every creature the player controls is tapped")
	}
	if b29Tapped(t, g, mine) {
		t.Error("mode 2: mine are not")
	}
}

func TestB29YargleAndMultaniIsVanilla(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.B29YargleAndMultaniOracle)
	if !ok {
		t.Fatal("not registered")
	}
	if spec.Completeness != effects.CompletenessFull || spec.OnResolve != nil || spec.AsEnters != nil ||
		len(spec.Triggered) != 0 || len(spec.Static) != 0 || len(spec.Activated) != 0 || len(spec.PrintedKeywords) != 0 {
		t.Error("Yargle and Multani has no rules text: a bare, complete spec")
	}
	g := effectstest.NewCatalogGame(t)
	id := effectstest.CastAndResolveCreature(t, g, "Yargle and Multani", "Legendary Creature — Frog Spirit Elemental", effectstest.B29YargleAndMultaniOracle)
	if !g.Battlefield.Contains(id) {
		t.Error("it resolves to the battlefield like any other creature")
	}
}
