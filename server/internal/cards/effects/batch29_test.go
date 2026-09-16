package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch29_test.go — card-level coverage for the card-coverage
// roadmap's batch 29 (#391, `edhrec_rank` 3041–3144): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, attack, damage, land play or step
// change. Helpers from the earlier batch test files are reused by
// name; new ones are b29-prefixed.

const (
	b29BlackWaltzNo3Oracle        = "9c84c681-40c1-4ff3-86e6-48b4ab7491e5"
	b29HeartlessHidetsuguOracle   = "b68e965a-96f4-447a-99be-35dbd27846e9"
	b29SettleTheWreckageOracle    = "2a2ea189-b663-4f4a-bb23-ff7a4af25f71"
	b29LegionLieutenantOracle     = "5fa1b2f0-3ba4-49db-9cb5-b6130e4c255e"
	b29SazhsChocoboOracle         = "c2b082b4-5d9c-4125-b353-d1cf4f334720"
	b29YouthfulValkyrieOracle     = "6d37ba4b-ff56-4eec-9dc2-2d7f357dc9c9"
	b29ElvishPromenadeOracle      = "7df3e379-c217-416e-a1c8-46338608c49e"
	b29WorldMapOracle             = "51d4f95e-aad2-4e5f-9f15-a68bbdf9ab3e"
	b29DazzlingAngelOracle        = "7de464e3-fae3-44cc-8233-776fc727c00a"
	b29HammerOfPurphorosOracle    = "212d058b-69c6-4dc2-8c93-bdfe26dc2ffe"
	b29CorneredByBlackMagesOracle = "68a58e0b-1506-492e-8ca3-016e520c10ba"
	b29EmpyreanEagleOracle        = "270d14b2-07bc-46bc-918f-658102265ccf"
	b29MassHysteriaOracle         = "4500131b-7417-4f30-a1b0-97d51b2e6458"
	b29BlowflyInfestationOracle   = "3859873f-2a0f-463b-8ed9-7f2ab5ed393a"
	b29ZombieApocalypseOracle     = "8241277d-654f-4985-9d49-a22c1e59eec2"
	b29LightPawsOracle            = "1718a442-b878-4690-b608-a013de3d79fc"
	b29ZiatoraOracle              = "d46c3fb6-f1c7-4a96-ae42-5bce17fc7c1d"
	b29WitchsOvenOracle           = "8fa0fe02-2452-4386-8e0c-165757b0f0a3"
	b29VedalkenArchmageOracle     = "568cf486-0261-4634-ac36-a6507101b2d0"
	b29ScytheclawRaptorOracle     = "8c653a0d-e35a-4596-bed6-b5156192955b"
	b29NayaCharmOracle            = "20cc0f3e-de69-4e5f-88b3-ae7e6c6b5996"
	b29HallowedHauntingOracle     = "e310ab58-180e-4840-bc8d-9f9b06f1b478"
	b29MountainValleyOracle       = "0b7393aa-d563-45bc-9946-8e7d1729d498"
	b29YargleAndMultaniOracle     = "980e1721-3717-4425-95a4-938dbf2ac661"
	b29ChitterspitterOracle       = "4a338863-d599-46e7-9c30-e11b898ae1b0"
	b29GimliOracle                = "f6747295-d21d-40d7-bc70-baa4a37ae668"
)

// b29Push seeds a permanent with a mana cost, colours and P/T on the
// battlefield with a layer timestamp and no summoning sickness.
func b29Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle, manaCost string, power, toughness int, colors ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: manaCost,
		Colors: colors, Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// b29Lives reads every seat's life total.
func b29Lives(g *game.Game) []int {
	out := make([]int, 0, len(g.Seats))
	for _, p := range g.Seats {
		out = append(out, p.Life)
	}
	return out
}

// b29CountNamed counts battlefield permanents `controller` controls
// with the given name.
func b29CountNamed(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			n++
		}
	}
	return n
}

// b29Tapped reads a battlefield card's tapped state.
func b29Tapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c, ok := battlefieldCard(g, id)
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

// b29CastAs has `caster` cast a card from hand right now, at
// whatever step the game is at — for an instant cast during combat
// or on another player's turn.
func b29CastAs(t *testing.T, g *game.Game, caster *game.Player, name, typeLine, oracle, manaCost string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	id := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: manaCost,
		Owner: caster.ID, Controller: caster.ID,
	})
	if err := g.CastSpell(caster.ID, id, params); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name.
func TestBatch29CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b29BlackWaltzNo3Oracle:        "Black Waltz No. 3",
		b29HeartlessHidetsuguOracle:   "Heartless Hidetsugu",
		b29SettleTheWreckageOracle:    "Settle the Wreckage",
		b29LegionLieutenantOracle:     "Legion Lieutenant",
		b29SazhsChocoboOracle:         "Sazh's Chocobo",
		b29YouthfulValkyrieOracle:     "Youthful Valkyrie",
		b29ElvishPromenadeOracle:      "Elvish Promenade",
		b29WorldMapOracle:             "World Map",
		b29DazzlingAngelOracle:        "Dazzling Angel",
		b29HammerOfPurphorosOracle:    "Hammer of Purphoros",
		b29CorneredByBlackMagesOracle: "Cornered by Black Mages",
		b29EmpyreanEagleOracle:        "Empyrean Eagle",
		b29MassHysteriaOracle:         "Mass Hysteria",
		b29BlowflyInfestationOracle:   "Blowfly Infestation",
		b29ZombieApocalypseOracle:     "Zombie Apocalypse",
		b29LightPawsOracle:            "Light-Paws, Emperor's Voice",
		b29ZiatoraOracle:              "Ziatora, the Incinerator",
		b29WitchsOvenOracle:           "Witch's Oven",
		b29VedalkenArchmageOracle:     "Vedalken Archmage",
		b29ScytheclawRaptorOracle:     "Scytheclaw Raptor",
		b29NayaCharmOracle:            "Naya Charm",
		b29HallowedHauntingOracle:     "Hallowed Haunting",
		b29MountainValleyOracle:       "Mountain Valley",
		b29YargleAndMultaniOracle:     "Yargle and Multani",
		b29ChitterspitterOracle:       "Chitterspitter",
		b29GimliOracle:                "Gimli of the Glittering Caves",
	}
	if len(want) != 26 {
		t.Fatalf("the batch registers 26 cards, the table lists %d", len(want))
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
	// The seven declared skips must stay out until their seam lands:
	// a put-a-card-from-hand-onto-the-battlefield prompt (Walking
	// Atlas, Elvish Piper), a counter-removal cost component (Power
	// Conduit, Staff of the Storyteller, Ramos), a resolution-time
	// choose-a-player prompt plus another player's choice at
	// resolution (Gluntch), and an opponent's three-way choice at
	// resolution (Master of Ceremonies).
	for oracle, name := range map[string]string{
		"855a4837-7fc8-4b97-afbd-daa88e322c89": "Walking Atlas",
		"872ef617-7cdb-4a7a-85f5-efc9e12bece5": "Elvish Piper",
		"2e358f7e-c282-4b92-8255-1436c99cda49": "Power Conduit",
		"0c4e2c90-c17b-42cc-b4d7-cf75970fbe90": "Staff of the Storyteller",
		"3ed41d2d-211b-4013-8562-8c64d54cc43a": "Ramos, Dragon Engine",
		"0222dc7c-459b-4909-a037-72b2eb248599": "Gluntch, the Bestower",
		"18ed0c8a-db8f-4247-87c4-544b833f01bd": "Master of Ceremonies",
	} {
		if _, ok := Lookup(oracle); ok {
			t.Errorf("%s is declared skipped on #391 but is registered — update the issue", name)
		}
	}
}

// --- the statics ---------------------------------------------------

func TestB29LegionLieutenantPumpsOtherVampiresYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lt := b29Push(g, me.ID, "Legion Lieutenant", "Creature — Vampire Knight", b29LegionLieutenantOracle, "{W}{B}", 2, 2, "W", "B")
	mine := pushTribalCreature(g, me.ID, "Vampire Nighthawk", "Creature — Vampire Shaman", 2, 3)
	bear := pushTribalCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := pushTribalCreature(g, opp.ID, "Their Vampire", "Creature — Vampire", 1, 1)
	if got := effectivePower(t, g, mine); got != 3 {
		t.Errorf("my Vampire: power %d, want 3", got)
	}
	if got := effectivePower(t, g, lt); got != 2 {
		t.Errorf("the Lieutenant itself: power %d, want 2 (\"other\")", got)
	}
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("a non-Vampire: power %d, want 2", got)
	}
	if got := effectivePower(t, g, theirs); got != 1 {
		t.Errorf("an opponent's Vampire: power %d, want 1 (\"you control\")", got)
	}
}

func TestB29EmpyreanEaglePumpsOtherFliersYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eagle := b29Push(g, me.ID, "Empyrean Eagle", "Creature — Bird Spirit", b29EmpyreanEagleOracle, "{1}{W}{U}", 2, 3, "W", "U")
	flier := pushTribalCreature(g, me.ID, "Bird", "Creature — Bird", 1, 1, "flying")
	ground := pushTribalCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := pushTribalCreature(g, opp.ID, "Their Bird", "Creature — Bird", 1, 1, "flying")
	if got := effectivePower(t, g, flier); got != 2 {
		t.Errorf("my flier: power %d, want 2", got)
	}
	if got := effectivePower(t, g, eagle); got != 2 {
		t.Errorf("the Eagle itself: power %d, want 2 (\"other\")", got)
	}
	if got := effectivePower(t, g, ground); got != 2 {
		t.Errorf("a ground creature: power %d, want 2", got)
	}
	if got := effectivePower(t, g, theirs); got != 1 {
		t.Errorf("an opponent's flier: power %d, want 1", got)
	}
	// A creature that GAINS flying from a layer-6 grant is pumped too.
	g.WithWriteLock(func() {
		g.RegisterTurnScopedStaticForEffect(b16GrantKeywords(func(target *game.Card, _ *game.Game, _ *game.Card) bool {
			return target.InstanceID == ground
		}, "flying"), uuid.Nil, "test — the ground creature gains flying")
	})
	if got := effectivePower(t, g, ground); got != 3 {
		t.Errorf("a creature granted flying: power %d, want 3", got)
	}
}

func TestB29MassHysteriaGivesEveryCreatureHaste(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b29Push(g, me.ID, "Mass Hysteria", "Enchantment", b29MassHysteriaOracle, "{R}", 0, 0, "R")
	mine := castAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	theirs := pushCatalogPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "", true)
	g.BumpLayerVersionForTest()
	if !hasEffectiveKeyword(t, g, mine, "haste") || !hasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("all creatures have haste — yours and theirs")
	}
	if summoningSickOf(t, g, mine) {
		t.Error("a hasty creature is not summoning sick")
	}
}

func TestB29HammerOfPurphorosGrantsHasteAndForgesGolems(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hammer := b29Push(g, me.ID, "Hammer of Purphoros", "Legendary Enchantment Artifact", b29HammerOfPurphorosOracle, "{1}{R}{R}", 0, 0, "R")
	mine := castAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	theirs := pushCatalogPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "", true)
	g.BumpLayerVersionForTest()
	if !hasEffectiveKeyword(t, g, mine, "haste") {
		t.Error("creatures you control have haste")
	}
	if hasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("an opponent's creature does not")
	}
	land := b29Push(g, me.ID, "Mountain", "Basic Land — Mountain", "", "", 0, 0)
	b29AddMana(me, "R", "R", "R")
	b16Activate(t, g, me.ID, hammer, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{land}})
	if g.Battlefield.Contains(land) {
		t.Error("the land is sacrificed to pay")
	}
	golem := findBattlefieldByName(g, "Golem")
	if golem == uuid.Nil {
		t.Fatal("no Golem token")
	}
	c, _ := battlefieldCard(g, golem)
	if !c.IsCreature() || !c.IsArtifact() || !c.IsEnchantment() || c.Power != 3 || c.Toughness != 3 || len(c.Colors) != 0 {
		t.Errorf("the Golem is a 3/3 colorless enchantment artifact creature, got %s %d/%d %v", c.TypeLine, c.Power, c.Toughness, c.Colors)
	}
	if !b29Tapped(t, g, hammer) {
		t.Error("the Hammer taps to activate")
	}
}

func TestB29HallowedHauntingGatesAtSevenAndSizesItsSpirits(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b29Push(g, me.ID, "Hallowed Haunting", "Enchantment", b29HallowedHauntingOracle, "{2}{W}{W}", 0, 0, "W")
	bear := pushTribalCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for i := 0; i < 5; i++ {
		b29Push(g, me.ID, "Anthem", "Enchantment", "", "{W}", 0, 0, "W")
	}
	if hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("six enchantments: no flying yet")
	}
	// The seventh enchantment is cast, so the Spirit Cleric follows.
	advanceToMain(t, g)
	b29CastAs(t, g, me, "Seventh Seal", "Enchantment", "", "{W}", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "flying") || !hasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Error("seven enchantments: creatures you control have flying and vigilance")
	}
	cleric := findBattlefieldByName(g, "Spirit Cleric")
	if cleric == uuid.Nil {
		t.Fatal("casting an enchantment spell makes a Spirit Cleric")
	}
	if got := effectivePower(t, g, cleric); got != 1 {
		t.Errorf("one Spirit: the Cleric is %d/%d, want 1/1", got, effectiveToughness(t, g, cleric))
	}
	pushTribalCreature(g, me.ID, "Spirit", "Creature — Spirit", 1, 1)
	if got := effectivePower(t, g, cleric); got != 2 {
		t.Errorf("two Spirits: the Cleric is %d/%d, want 2/2", got, effectiveToughness(t, g, cleric))
	}
	// A creature spell is not an enchantment spell.
	castAndResolveCreature(t, g, "Elf", "Creature — Elf", "")
	passPriorityAroundTable(t, g)
	if got := b29CountNamed(g, me.ID, "Spirit Cleric"); got != 1 {
		t.Errorf("a creature spell: %d Clerics, want 1", got)
	}
	if spec, _ := Lookup(b29HallowedHauntingOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the carried-sizing gap must be declared")
	}
}

func TestB29ChitterspitterEatsATokenEachUpkeepAndGrowsSquirrels(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spitter := b29Push(g, me.ID, "Chitterspitter", "Artifact", b29ChitterspitterOracle, "{2}{G}", 0, 0, "G")
	squirrel := pushTribalCreature(g, me.ID, "Drey Keeper", "Creature — Squirrel", 1, 1)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("1/1 green Squirrel"), 1) })
	token := findBattlefieldByName(g, "Squirrel")
	if got := effectivePower(t, g, squirrel); got != 1 {
		t.Fatalf("no acorns: power %d, want 1", got)
	}
	b12ToMyNextUpkeep(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, token)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(token) {
		t.Error("the chosen token is sacrificed")
	}
	if got := counterCount(g, spitter, "acorn"); got != 1 {
		t.Errorf("%d acorn counters, want 1", got)
	}
	if got := effectivePower(t, g, squirrel); got != 2 {
		t.Errorf("one acorn: Squirrel power %d, want 2", got)
	}
	// {G}, {T}: a Squirrel, which is pumped too.
	advanceToMain(t, g)
	b29AddMana(me, "G")
	b16Activate(t, g, me.ID, spitter, 0, game.ActivateAbilityParams{})
	made := findBattlefieldByName(g, "Squirrel")
	if made == uuid.Nil {
		t.Fatal("the activation makes a Squirrel")
	}
	if got := effectivePower(t, g, made); got != 2 {
		t.Errorf("the new Squirrel: power %d, want 2", got)
	}
}

func TestB29ChitterspitterWithNoTokenIsSilent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b29Push(g, me.ID, "Chitterspitter", "Artifact", b29ChitterspitterOracle, "{2}{G}", 0, 0, "G")
	pushTribalCreature(g, me.ID, "Drey Keeper", "Creature — Squirrel", 1, 1)
	b12ToMyNextUpkeep(t, g)
	passPriorityAroundTable(t, g)
	if latestTriggerPrompt(g, me.ID) != nil || latestPickTarget(g, me.ID) != nil {
		t.Error("no token: no prompt at all (CR 603.3d)")
	}
}

// --- the triggers --------------------------------------------------

func TestB29BlackWaltzPingsTheTableOnNoncreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	waltz := b29Push(g, me.ID, "Black Waltz No. 3", "Legendary Creature — Wizard", b29BlackWaltzNo3Oracle, "{2}{B}{R}", 2, 2, "B", "R")
	if !hasEffectiveKeyword(t, g, waltz, "flying") || !hasEffectiveKeyword(t, g, waltz, "deathtouch") {
		t.Error("flying and deathtouch are printed")
	}
	before := b29Lives(g)
	b24CastNoncreature(t, g, "Shock", "Instant", "{R}")
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		want := before[i] - 2
		if i == 0 {
			want = before[i]
		}
		if p.Life != want {
			t.Errorf("seat %d: life %d, want %d", i, p.Life, want)
		}
	}
	mid := b29Lives(g)
	castAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.Life != mid[i] {
			t.Errorf("a creature spell: seat %d life %d, want %d", i, p.Life, mid[i])
		}
	}
}

func TestB29SazhsChocoboGrowsOnLandfall(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	chocobo := b29Push(g, me.ID, "Sazh's Chocobo", "Creature — Bird", b29SazhsChocoboOracle, "{G}", 0, 1, "G")
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	if got := counterCount(g, chocobo, "+1/+1"); got != 1 {
		t.Errorf("a land drop: %d +1/+1 counters, want 1", got)
	}
	castAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	if got := counterCount(g, chocobo, "+1/+1"); got != 1 {
		t.Errorf("a creature entering: %d counters, want still 1", got)
	}
}

func TestB29YouthfulValkyrieGrowsWhenAnotherAngelEnters(t *testing.T) {
	g := newCatalogGame(t)
	valkyrie := castAndResolveCreature(t, g, "Youthful Valkyrie", "Creature — Angel", b29YouthfulValkyrieOracle)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, valkyrie, "+1/+1"); got != 0 {
		t.Fatalf("its own entry: %d counters, want 0 (\"another\")", got)
	}
	castAndResolveCreature(t, g, "Serra Angel", "Creature — Angel", "")
	passPriorityAroundTable(t, g)
	if got := counterCount(g, valkyrie, "+1/+1"); got != 1 {
		t.Errorf("another Angel: %d counters, want 1", got)
	}
	castAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	if got := counterCount(g, valkyrie, "+1/+1"); got != 1 {
		t.Errorf("a non-Angel: %d counters, want still 1", got)
	}
	if !hasEffectiveKeyword(t, g, valkyrie, "flying") {
		t.Error("flying is printed")
	}
}

func TestB29DazzlingAngelGainsOnOtherCreaturesEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	start := me.Life
	angel := castAndResolveCreature(t, g, "Dazzling Angel", "Creature — Angel", b29DazzlingAngelOracle)
	passPriorityAroundTable(t, g)
	if me.Life != start {
		t.Fatalf("its own entry: life %d, want %d (\"another\")", me.Life, start)
	}
	castAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	if me.Life != start+1 {
		t.Errorf("a creature entering: life %d, want %d", me.Life, start+1)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("1/1 colorless Spirit with flying"), 2) })
	passPriorityAroundTable(t, g)
	if me.Life != start+3 {
		t.Errorf("two tokens: life %d, want %d", me.Life, start+3)
	}
	if !hasEffectiveKeyword(t, g, angel, "flying") {
		t.Error("flying is printed")
	}
}

func TestB29VedalkenArchmageDrawsOnArtifactSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b29Push(g, me.ID, "Vedalken Archmage", "Creature — Vedalken Wizard", b29VedalkenArchmageOracle, "{2}{U}{U}", 0, 2, "U")
	hand := me.Hand.Size()
	b24CastNoncreature(t, g, "Sol Ring", "Artifact", "{1}")
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("an artifact spell: hand %d → %d, want +1", hand, got)
	}
	hand = me.Hand.Size()
	b24CastNoncreature(t, g, "Shock", "Instant", "{R}")
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("an instant: hand %d → %d, want unchanged", hand, got)
	}
}

func TestB29ScytheclawRaptorPunishesOffTurnCasts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b29Push(g, me.ID, "Scytheclaw Raptor", "Creature — Dinosaur", b29ScytheclawRaptorOracle, "{2}{R}", 4, 3, "R")
	advanceToMain(t, g)
	before := opp.Life
	b29CastAs(t, g, opp, "Opt", "Instant", "", "{U}", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if opp.Life != before-4 {
		t.Errorf("an opponent casting on my turn: life %d, want %d", opp.Life, before-4)
	}
	mine := me.Life
	b24CastNoncreature(t, g, "Shock", "Instant", "{R}")
	passPriorityAroundTable(t, g)
	if me.Life != mine {
		t.Errorf("casting on my own turn: life %d, want %d", me.Life, mine)
	}
}

func TestB29GimliGrowsOnLegendsAndMakesTreasureOnConnecting(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gimli := b29Push(g, me.ID, "Gimli of the Glittering Caves", "Legendary Creature — Dwarf Warrior", b29GimliOracle, "{2}{R}", 1, 1, "R")
	if !hasEffectiveKeyword(t, g, gimli, "double strike") {
		t.Error("double strike is printed")
	}
	castAndResolveCreature(t, g, "Legolas", "Legendary Creature — Elf Archer", "")
	passPriorityAroundTable(t, g)
	if got := counterCount(g, gimli, "+1/+1"); got != 1 {
		t.Errorf("a legendary creature: %d counters, want 1", got)
	}
	castAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	if got := counterCount(g, gimli, "+1/+1"); got != 1 {
		t.Errorf("a nonlegendary creature: %d counters, want still 1", got)
	}
	attackWith(t, g, opp.ID, gimli)
	passPriorityAroundTable(t, g)
	if got := b29CountNamed(g, me.ID, "Treasure"); got != 2 {
		t.Errorf("Gimli connecting with double strike: %d Treasures, want 2 (one per damage step)", got)
	}
	if opp.Life != 36 {
		t.Errorf("a 2/2 double striker: opponent life %d, want 36", opp.Life)
	}
}

func TestB29BlowflyInfestationChainsMinusCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b29Push(g, me.ID, "Blowfly Infestation", "Enchantment", b29BlowflyInfestationOracle, "{2}{B}", 0, 0, "B")
	infested := pushCounterCreature(g, opp.ID, "Infested", game.CounterMinusOne, 1)
	victim := pushCounterCreature(g, opp.ID, "Victim", "", 0)
	clean := pushCounterCreature(g, me.ID, "Clean", "", 0)
	b27Kill(g, infested)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, victim, game.CounterMinusOne); got != 1 {
		t.Errorf("target creature: %d -1/-1 counters, want 1", got)
	}
	// A creature that dies WITHOUT a -1/-1 counter is not a trigger.
	b27Kill(g, clean)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a clean death: no trigger")
	}
	if got := counterCount(g, victim, game.CounterMinusOne); got != 1 {
		t.Errorf("still %d counters, want 1", got)
	}
	// The countered creature dying chains: this time onto my own.
	mine := pushCounterCreature(g, me.ID, "Mine", "", 0)
	b27Kill(g, victim)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, mine)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, mine, game.CounterMinusOne); got != 1 {
		t.Errorf("the chain: %d -1/-1 counters, want 1", got)
	}
}

func TestB29LightPawsFetchesASmallerAuraOntoItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	paws := b29Push(g, me.ID, "Light-Paws, Emperor's Voice", "Legendary Creature — Fox Advisor", b29LightPawsOracle, "{1}{W}", 2, 2, "W")
	bear := pushTribalCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rancor := pushLibraryCardForTest(me, game.Card{Name: "Rancor", TypeLine: auraTypeLine, OracleID: rancorOracle, ManaCost: "{G}", Colors: []string{"G"}})
	pushLibraryCardForTest(me, game.Card{Name: "Small Aura", TypeLine: auraTypeLine, ManaCost: "{1}{W}", Colors: []string{"W"}})
	pushLibraryCardForTest(me, game.Card{Name: "Big Aura", TypeLine: auraTypeLine, ManaCost: "{2}{W}{W}", Colors: []string{"W"}})
	pushLibraryCardForTest(me, game.Card{Name: "Ethereal Armor", TypeLine: auraTypeLine, ManaCost: "{W}", Colors: []string{"W"}})
	pushLibraryCardForTest(me, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})
	advanceToMain(t, g)
	b29CastAs(t, g, me, "Ethereal Armor", auraTypeLine, "", "{W}{W}", game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Rancor") == uuid.Nil || searchOptionNamed(g, c, "Small Aura") == uuid.Nil {
		t.Error("Rancor and Small Aura (mana value ≤ 2, different names) are offered")
	}
	if searchOptionNamed(g, c, "Big Aura") != uuid.Nil {
		t.Error("an Aura with mana value 4 is not offered")
	}
	if searchOptionNamed(g, c, "Ethereal Armor") != uuid.Nil {
		t.Error("an Aura sharing a name with one you control is not offered")
	}
	if searchOptionNamed(g, c, "Plains") != uuid.Nil {
		t.Error("a land is not an Aura")
	}
	answerSearchByID(t, g, me.ID, rancor)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rancor) {
		t.Fatal("Rancor is put onto the battlefield")
	}
	if host := attachmentHostOf(t, g, rancor); host.Kind != game.TargetCard || host.ID != paws {
		t.Fatalf("Rancor is attached to %+v, want Light-Paws", host)
	}
	if got := effectivePower(t, g, paws); got != 4 {
		t.Errorf("Light-Paws wearing Rancor: power %d, want 4", got)
	}
	// The fetched Aura was not cast, so it does not chain.
	if latestTriggerPrompt(g, me.ID) != nil {
		t.Error("a fetched Aura is not \"if you cast it\"")
	}
}

func TestB29ZiatoraFlingsAtTheEndStepThroughAReflexiveTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ziatora := b29Push(g, me.ID, "Ziatora, the Incinerator", "Legendary Creature — Demon Dragon", b29ZiatoraOracle, "{3}{B}{R}{G}", 6, 6, "B", "R", "G")
	fodder := pushCounterCreature(g, me.ID, "Fodder", game.CounterPlusOne, 2)
	if !hasEffectiveKeyword(t, g, ziatora, "flying") {
		t.Error("flying is printed")
	}
	before := opp.Life
	advanceToEndStepOf(t, g, 0)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	for _, id := range p.PickTargetCards {
		if id == ziatora {
			t.Error("Ziatora herself is not \"another creature\"")
		}
	}
	pickCard(t, g, me.ID, fodder)
	// The end-step trigger resolves: the sacrifice, then the
	// reflexive trigger asks for its target.
	b04WaitForPick(t, g, me.ID)
	if g.Battlefield.Contains(fodder) {
		t.Fatal("the chosen creature is sacrificed before the reflexive trigger fires")
	}
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != before-6 {
		t.Errorf("a 4/4 with two +1/+1 counters: opponent life %d, want %d (power 6)", opp.Life, before-6)
	}
	if got := b29CountNamed(g, me.ID, "Treasure"); got != 3 {
		t.Errorf("%d Treasures, want 3", got)
	}
}

func TestB29ZiatoraDeclinedOrAloneDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b29Push(g, me.ID, "Ziatora, the Incinerator", "Legendary Creature — Demon Dragon", b29ZiatoraOracle, "{3}{B}{R}{G}", 6, 6, "B", "R", "G")
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if latestTriggerPrompt(g, me.ID) != nil || latestPickTarget(g, me.ID) != nil {
		t.Error("no other creature: no prompt at all (CR 603.3d)")
	}
	if got := b29CountNamed(g, me.ID, "Treasure"); got != 0 {
		t.Errorf("%d Treasures, want 0", got)
	}
}

// --- the activations -----------------------------------------------

func TestB29HeartlessHidetsuguHalvesEveryLifeTotal(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hidetsugu := b29Push(g, me.ID, "Heartless Hidetsugu", "Legendary Creature — Ogre Shaman", b29HeartlessHidetsuguOracle, "{3}{R}{R}", 4, 3, "R")
	g.Seats[0].Life, g.Seats[1].Life, g.Seats[2].Life, g.Seats[3].Life = 40, 21, 7, 1
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, hidetsugu, 0, game.ActivateAbilityParams{})
	if got := b29Lives(g); got[0] != 20 || got[1] != 11 || got[2] != 4 || got[3] != 1 {
		t.Errorf("lives %v, want [20 11 4 1]", got)
	}
	if !b29Tapped(t, g, hidetsugu) {
		t.Error("Hidetsugu taps")
	}
}

func TestB29WorldMapTutorsABasicOrAnyLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	worldMap := b29Push(g, me.ID, "World Map", "Artifact", b29WorldMapOracle, "{1}", 0, 0)
	forest := pushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	pushLibraryCardForTest(me, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})
	pushLibraryCardForTest(me, game.Card{Name: "Cabal Coffers", TypeLine: "Land"})
	advanceToMain(t, g)
	b29AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, worldMap, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate {1}: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Forest") == uuid.Nil || searchOptionNamed(g, c, "Plains") == uuid.Nil || searchOptionNamed(g, c, "Cabal Coffers") != uuid.Nil {
		t.Error("the {1} mode offers basics only")
	}
	answerSearchByID(t, g, me.ID, forest)
	if !me.Hand.Contains(forest) {
		t.Error("the Forest is put into hand")
	}
	if g.Battlefield.Contains(worldMap) {
		t.Error("the Map is sacrificed")
	}
	// The {3} mode offers any land.
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	map2 := b29Push(g2, me2.ID, "World Map", "Artifact", b29WorldMapOracle, "{1}", 0, 0)
	pushLibraryCardForTest(me2, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	coffers := pushLibraryCardForTest(me2, game.Card{Name: "Cabal Coffers", TypeLine: "Land"})
	advanceToMain(t, g2)
	b29AddMana(me2, "C", "C", "C")
	if err := g2.ActivateCatalogAbility(me2.ID, map2, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate {3}: %v", err)
	}
	passPriorityAroundTable(t, g2)
	if c := searchChoiceFor(g2, me2.ID); c == nil || searchOptionNamed(g2, c, "Cabal Coffers") == uuid.Nil {
		t.Fatal("the {3} mode offers a nonbasic land")
	}
	answerSearchByID(t, g2, me2.ID, coffers)
	if !me2.Hand.Contains(coffers) {
		t.Error("the {3} mode finds a nonbasic land")
	}
}

func TestB29WitchsOvenBakesOneOrTwoFood(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	oven := b29Push(g, me.ID, "Witch's Oven", "Artifact", b29WitchsOvenOracle, "{1}", 0, 0)
	small := pushTribalCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	big := pushTribalCreature(g, me.ID, "Rhino", "Creature — Rhino", 4, 4)
	grown := pushCounterCreature(g, me.ID, "Grown", game.CounterPlusOne, 1) // a 4/4 base pushed as 4/4 — see below
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, oven, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{small}})
	if got := b29CountNamed(g, me.ID, "Food"); got != 1 {
		t.Errorf("a 2/2: %d Food, want 1", got)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(oven) })
	b16Activate(t, g, me.ID, oven, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{big}})
	if got := b29CountNamed(g, me.ID, "Food"); got != 3 {
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
	b16Activate(t, g, me.ID, oven, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{grown}})
	if got := b29CountNamed(g, me.ID, "Food"); got != 5 {
		t.Errorf("a 3/3 with a +1/+1 counter: %d Food, want 5 (two more)", got)
	}
	if spec, _ := Lookup(b29WitchsOvenOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the static-bonus gap must be declared")
	}
}

func TestB29MountainValleyEntersTappedAndFetchesAMountainOrForest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushLibraryCardForTest(me, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
	pushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	ground := pushLibraryCardForTest(me, game.Card{Name: "Stomping Ground", TypeLine: "Land — Mountain Forest"})
	pushLibraryCardForTest(me, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	valley := playLandFromHand(t, g, "Mountain Valley", b29MountainValleyOracle)
	passPriorityAroundTable(t, g)
	if !b29Tapped(t, g, valley) {
		t.Fatal("Mountain Valley enters tapped")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(valley) })
	if err := g.ActivateCatalogAbility(me.ID, valley, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Mountain") == uuid.Nil || searchOptionNamed(g, c, "Forest") == uuid.Nil || searchOptionNamed(g, c, "Stomping Ground") == uuid.Nil {
		t.Error("every Mountain or Forest card is offered, nonbasics included")
	}
	if searchOptionNamed(g, c, "Island") != uuid.Nil {
		t.Error("an Island is not offered")
	}
	answerSearchByID(t, g, me.ID, ground)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushVanillaCreature(g, me.ID, "Attacker A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Attacker B", 3, 3)
	home := pushVanillaCreature(g, me.ID, "Stay-at-home", 1, 1)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	f1 := pushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	f2 := pushLibraryCardForTest(me, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})
	f3 := pushLibraryCardForTest(me, game.Card{Name: "Swamp", TypeLine: "Basic Land — Swamp"})
	declareAttack(t, g, opp.ID, a, b)
	b29CastAs(t, g, opp, "Settle the Wreckage", "Instant", b29SettleTheWreckageOracle, "{2}{W}{W}", game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	})
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(a) || !g.Exile.Contains(b) {
		t.Fatal("both attackers are exiled")
	}
	if !g.Battlefield.Contains(home) || !g.Battlefield.Contains(theirs) {
		t.Error("a creature that did not attack, and the opponent's, stay")
	}
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the attacked player gets the search prompt")
	}
	if len(c.SearchCards) != 3 {
		t.Errorf("%d basics offered, want 3", len(c.SearchCards))
	}
	answerSearchByID(t, g, me.ID, f1, f2)
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

func TestB29ElvishPromenadeMakesAnElfWarriorPerElf(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushTribalCreature(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	pushTribalCreature(g, me.ID, "Elvish Mystic", "Creature — Elf Druid", 1, 1)
	pushTribalCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	pushTribalCreature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	castCatalogSpell(t, g, "Elvish Promenade", "Kindred Sorcery — Elf", b29ElvishPromenadeOracle, nil)
	passPriorityAroundTable(t, g)
	if got := b29CountNamed(g, me.ID, "Elf Warrior"); got != 2 {
		t.Errorf("two Elves: %d Elf Warriors, want 2", got)
	}
	if got := b29CountNamed(g, opp.ID, "Elf Warrior"); got != 0 {
		t.Errorf("the opponent gets none, got %d", got)
	}
}

func TestB29CorneredByBlackMagesEdictsAndLeavesAWizard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	keep := pushVanillaCreature(g, opp.ID, "Keeper", 3, 3)
	fodder := pushVanillaCreature(g, opp.ID, "Fodder", 1, 1)
	castCatalogSpell(t, g, "Cornered by Black Mages", "Sorcery", b29CorneredByBlackMagesOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	c := sacrificeChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("the opponent gets a sacrifice prompt")
	}
	if len(c.SacrificeOptions) != 2 {
		t.Errorf("%d options, want the opponent's 2 creatures", len(c.SacrificeOptions))
	}
	answerSacrifice(t, g, opp.ID, fodder)
	if g.Battlefield.Contains(fodder) || !g.Battlefield.Contains(keep) {
		t.Error("the opponent's choice is sacrificed, the other stays")
	}
	wizard := findBattlefieldByName(g, "Wizard")
	if wizard == uuid.Nil {
		t.Fatal("no Wizard token")
	}
	w, _ := battlefieldCard(g, wizard)
	if w.Controller != me.ID || w.Power != 0 || w.Toughness != 1 || len(w.Colors) != 1 || w.Colors[0] != "B" {
		t.Errorf("the Wizard is my 0/1 black token, got %d/%d %v", w.Power, w.Toughness, w.Colors)
	}
	if spec, _ := Lookup(b29CorneredByBlackMagesOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the token's missing ping must be declared")
	}
}

func TestB29ZombieApocalypseRaisesZombiesTappedThenKillsHumans(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zombie := pushGraveyardCardTyped(me, "Gravecrawler", "Creature — Zombie")
	zombieHuman := pushGraveyardCardTyped(me, "Zombie Villager", "Creature — Zombie Human")
	elf := pushGraveyardCardTyped(me, "Dead Elf", "Creature — Elf")
	zombieLand := pushGraveyardCardTyped(me, "Zombie Land", "Land — Zombie")
	theirZombie := pushGraveyardCardTyped(opp, "Their Zombie", "Creature — Zombie")
	myHuman := pushVanillaCreature(g, me.ID, "My Human", 2, 2)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == myHuman {
			g.Battlefield.Cards[i].TypeLine = "Creature — Human Soldier"
		}
	}
	theirHuman := pushTribalCreature(g, opp.ID, "Their Human", "Creature — Human Wizard", 1, 1)
	theirElf := pushTribalCreature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	castCatalogSpell(t, g, "Zombie Apocalypse", "Sorcery", b29ZombieApocalypseOracle, nil)
	passPriorityAroundTable(t, g)
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

// pushGraveyardCardTyped seeds a card with a type line into a
// player's graveyard.
func pushGraveyardCardTyped(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine,
		Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID,
	})
	return id
}

func TestB29NayaCharmIsChooseOneAndResolvesEachMode(t *testing.T) {
	// Mode 0: 3 damage to target creature.
	g := newCatalogGame(t)
	opp := g.Seats[1]
	rhino := pushVanillaCreature(g, opp.ID, "Rhino", 3, 3)
	castModal(t, g, "Naya Charm", "Instant", b29NayaCharmOracle, []int{0}, []game.TargetRef{{Kind: game.TargetCard, ID: rhino}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rhino) {
		t.Error("mode 0: a 3/3 dies to 3 damage")
	}
	// Mode 1: return target card from ANY graveyard to its owner's hand.
	g = newCatalogGame(t)
	opp = g.Seats[1]
	dead := pushGraveyardCardTyped(opp, "Their Dead Bear", "Creature — Bear")
	hand := opp.Hand.Size()
	castModal(t, g, "Naya Charm", "Instant", b29NayaCharmOracle, []int{1}, []game.TargetRef{{Kind: game.TargetCard, ID: dead}})
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(dead) || opp.Hand.Size() != hand+1 {
		t.Error("mode 1: the card goes to its OWNER's hand")
	}
	// Mode 2: tap all creatures target player controls.
	g = newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushVanillaCreature(g, opp.ID, "A", 1, 1)
	b := pushVanillaCreature(g, opp.ID, "B", 1, 1)
	mine := pushVanillaCreature(g, me.ID, "Mine", 1, 1)
	castModal(t, g, "Naya Charm", "Instant", b29NayaCharmOracle, []int{2}, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if !b29Tapped(t, g, a) || !b29Tapped(t, g, b) {
		t.Error("mode 2: every creature the player controls is tapped")
	}
	if b29Tapped(t, g, mine) {
		t.Error("mode 2: mine are not")
	}
}

func TestB29YargleAndMultaniIsVanilla(t *testing.T) {
	spec, ok := Lookup(b29YargleAndMultaniOracle)
	if !ok {
		t.Fatal("not registered")
	}
	if spec.Completeness != CompletenessFull || spec.OnResolve != nil || spec.AsEnters != nil ||
		len(spec.Triggered) != 0 || len(spec.Static) != 0 || len(spec.Activated) != 0 || len(spec.PrintedKeywords) != 0 {
		t.Error("Yargle and Multani has no rules text: a bare, complete spec")
	}
	g := newCatalogGame(t)
	id := castAndResolveCreature(t, g, "Yargle and Multani", "Legendary Creature — Frog Spirit Elemental", b29YargleAndMultaniOracle)
	if !g.Battlefield.Contains(id) {
		t.Error("it resolves to the battlefield like any other creature")
	}
}
