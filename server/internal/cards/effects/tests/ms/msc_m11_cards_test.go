package testsms

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_m11_cards_test.go — Marvel Super Heroes Commander, M–Z batch 11 (#4455).
// Each test plays the real catalog entry and asserts what its text says.

const (
	mscM11MBakuOracle      = "9e4656ba-c375-4ef8-ba1d-5a7b0833242f"
	mscM11MarvelBoyOracle  = "7a315db3-e66c-4648-b7a0-5aef2ce9b580"
	mscM11StrategicOracle  = "b8d019fa-3893-453e-a46d-aa53fc3f2f0e"
	mscM11SteelOracle      = "e36100d6-d9ae-454a-bba1-b50a69d5a28f"
	mscM11FrightfulOracle  = "6d3469dd-d39e-463b-9b41-2f6215a41feb"
	mscM11UltronOracle     = "19fd953c-b6d4-457f-b85f-b681c5b0e341"
	mscM11WhiteWidowOracle = "7aff7cb6-aa56-47ae-8813-75d850917320"
	mscM11WildPackOracle   = "294b2aa7-cb46-46bd-8241-69b26d73aa2d"
	mscM11BuckyOracle      = "874bbeb2-aa15-41e8-afad-122ab9b555fd"
	mscM11X23Oracle        = "2b4dd121-b477-42ff-9dd1-5d0f957a509d"
	mscM11ZardaOracle      = "c7b158b9-19f7-47b6-a97c-74023dfe73b9"
)

func TestMscM11EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mscM11MBakuOracle:      "M'Baku, Jabari Chieftain",
		mscM11MarvelBoyOracle:  "Marvel Boy, Noh-Varr",
		mscM11StrategicOracle:  "Strategic Intervention",
		mscM11SteelOracle:      "Swordsman's Steel",
		mscM11FrightfulOracle:  "The Frightful Four",
		mscM11UltronOracle:     "Ultron, Machine Overlord",
		mscM11WhiteWidowOracle: "White Widow, Yelena Belova",
		mscM11WildPackOracle:   "Wild Pack Squad",
		mscM11BuckyOracle:      "Winter Soldier, Bucky Barnes",
		mscM11X23Oracle:        "X-23, Deadly Weapon",
		mscM11ZardaOracle:      "Zarda, the Power Princess",
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

// --- Winter Soldier, Bucky Barnes ---------------------------------------------

func TestMscM11BuckyEntersTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.CastAndResolveCreature(t, g, "Winter Soldier, Bucky Barnes", "Legendary Creature — Human Soldier Hero", mscM11BuckyOracle)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Bucky did not enter")
	}
	if !effectstest.BattlefieldCardPtr(t, g, id).Tapped {
		t.Error("Winter Soldier entered untapped")
	}
}

func TestMscM11BuckyDoesNotTapOtherCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.PushCatalogPermanent(g, g.Seats[0].ID, "Winter Soldier, Bucky Barnes", "Legendary Creature — Human Soldier Hero", mscM11BuckyOracle, false)
	bear := effectstest.B12PlayFromHand(t, g, "Bear", "Creature — Bear", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldCardPtr(t, g, bear).Tapped {
		t.Error("another creature entered tapped")
	}
}

// --- Wild Pack Squad ---------------------------------------------------------------

func mscM11SquadAtCombat(t *testing.T) (*game.Game, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Wild Pack Squad", "Creature — Human Mercenary", mscM11WildPackOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	return g, bear
}

func TestMscM11WildPackSquadGivesFirstStrikeAndVigilanceToTheTarget(t *testing.T) {
	g, bear := mscM11SquadAtCombat(t)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.EffectiveAbilitiesContain(t, g, bear, "first strike") || !effectstest.EffectiveAbilitiesContain(t, g, bear, "vigilance") {
		t.Errorf("the target has %v, want first strike and vigilance", effectstest.EffectiveAbilities(t, g, bear))
	}
}

func TestMscM11WildPackSquadMayTargetAnyCreatureAndLastsOnlyThisTurn(t *testing.T) {
	g, _ := mscM11SquadAtCombat(t)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.EffectiveAbilitiesContain(t, g, theirs, "first strike") {
		t.Fatal("an opponent's creature is a legal target and gained nothing")
	}
	effectstest.PowerUpPassTo(t, g, 0)
	if effectstest.EffectiveAbilitiesContain(t, g, theirs, "first strike") || effectstest.EffectiveAbilitiesContain(t, g, theirs, "vigilance") {
		t.Error("the grant outlasted the turn")
	}
}

func TestMscM11WildPackSquadMayTargetNothing(t *testing.T) {
	g, bear := mscM11SquadAtCombat(t)
	ch := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTargets(ch.ID, ch.Chooser, nil); err != nil {
		t.Fatalf("choosing no target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EffectiveAbilitiesContain(t, g, bear, "first strike") {
		t.Error("a creature gained first strike with no target chosen")
	}
}

// --- The Frightful Four ------------------------------------------------------------

func TestMscM11FrightfulFourDrainsTheFirstNoncreatureSpellOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	four := effectstest.PushCatalogPermanent(g, me.ID, "The Frightful Four", "Legendary Creature — Villain", mscM11FrightfulOracle, false)
	effectstest.AssertKeywords(t, g, four, "menace")
	life := opp.Life

	effectstest.Batch01OpponentCasts(t, g, opp, "First Spell", "", "{2}{R}", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Fatalf("the caster has %d life, want %d (mana value 3)", opp.Life, life-3)
	}
	effectstest.Batch01OpponentCasts(t, g, opp, "Second Spell", "", "{4}", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("the second noncreature spell cost life: %d, want %d", opp.Life, life-3)
	}
}

func TestMscM11FrightfulFourIgnoresYourOwnSpellsAndOnlyTheCasterLoses(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "The Frightful Four", "Legendary Creature — Villain", mscM11FrightfulOracle, false)
	life, thirdLife := me.Life, third.Life

	effectstest.CastFree(t, g, me, "My Spell", "Instant")
	if me.Life != life {
		t.Errorf("my own spell cost me life: %d -> %d", life, me.Life)
	}
	oppLife := opp.Life
	effectstest.Batch01OpponentCasts(t, g, opp, "Their Spell", "", "{1}", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-1 {
		t.Errorf("the caster has %d life, want %d", opp.Life, oppLife-1)
	}
	if third.Life != thirdLife || me.Life != life {
		t.Errorf("someone other than the caster lost life: me %d, third %d", me.Life, third.Life)
	}
}

// --- White Widow, Yelena Belova ----------------------------------------------------

func TestMscM11WhiteWidowGrowsWhenSheConnects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	widow := effectstest.B12Push(g, me.ID, "White Widow, Yelena Belova", "Legendary Creature — Human Assassin Villain", mscM11WhiteWidowOracle, 1, 2)
	effectstest.AssertKeywords(t, g, widow, "deathtouch")
	effectstest.DeclareAttack(t, g, opp.ID, widow)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, widow, game.CounterPlusOne); got != 1 {
		t.Errorf("White Widow has %d +1/+1 counters, want 1", got)
	}
}

func TestMscM11WhiteWidowPutsTheCounterOnAnotherDeathtouchCreatureNotOnHerself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	widow := effectstest.B12Push(g, me.ID, "White Widow, Yelena Belova", "Legendary Creature — Human Assassin Villain", mscM11WhiteWidowOracle, 1, 2)
	snake := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Deathtouch Snake", TypeLine: "Creature — Snake",
		Power: 1, Toughness: 1, Keywords: []string{"deathtouch"}, Owner: me.ID, Controller: me.ID,
	})
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, snake, bear)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, snake, game.CounterPlusOne); got != 1 {
		t.Errorf("the deathtouch Snake has %d counters, want 1", got)
	}
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 0 {
		t.Errorf("the creature without deathtouch got %d counters, want 0", got)
	}
	if got := effectstest.CountersOn(g, widow, game.CounterPlusOne); got != 0 {
		t.Errorf("White Widow did not attack and has %d counters, want 0", got)
	}
}

func TestMscM11WhiteWidowGetsNothingWhenBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	widow := effectstest.B12Push(g, me.ID, "White Widow, Yelena Belova", "Legendary Creature — Human Assassin Villain", mscM11WhiteWidowOracle, 1, 2)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(widow, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, widow); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, widow, game.CounterPlusOne); got != 0 {
		t.Errorf("a blocked White Widow got %d counters, want 0 (no damage to a player)", got)
	}
}

// --- X-23, Deadly Weapon -----------------------------------------------------------

func TestMscM11X23CountersGoOnTheEnteringMutantAndOnX23(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	x23 := effectstest.PushCatalogPermanent(g, me.ID, "X-23, Deadly Weapon", "Legendary Creature — Mutant Berserker Hero", mscM11X23Oracle, false)
	mutant := effectstest.B12PlayFromHand(t, g, "Test Mutant", "Creature — Mutant", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, mutant, game.CounterPlusOne); got != 1 {
		t.Errorf("the Mutant has %d counters, want 1", got)
	}
	if got := effectstest.CountersOn(g, x23, game.CounterPlusOne); got != 1 {
		t.Errorf("X-23 has %d counters, want 1", got)
	}
}

func TestMscM11X23IgnoresNonMutantsAnOpponentsMutantAndItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	x23 := effectstest.PushCatalogPermanent(g, me.ID, "X-23, Deadly Weapon", "Legendary Creature — Mutant Berserker Hero", mscM11X23Oracle, false)
	effectstest.B12PlayFromHand(t, g, "Plain Bear", "Creature — Bear", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MscM11Token(t, g, opp.ID, "1/1 white Soldier")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, x23, game.CounterPlusOne); got != 0 {
		t.Errorf("X-23 has %d counters after a Bear and an opposing creature, want 0", got)
	}
	// X-23 entering herself is "another": cast her and nothing happens.
	g2 := effectstest.NewCatalogGame(t)
	self := effectstest.CastAndResolveCreature(t, g2, "X-23, Deadly Weapon", "Legendary Creature — Mutant Berserker Hero", mscM11X23Oracle)
	effectstest.PassPriorityAroundTable(t, g2)
	if got := effectstest.CountersOn(g2, self, game.CounterPlusOne); got != 0 {
		t.Errorf("X-23 put a counter on herself as she entered: %d", got)
	}
}

// --- Zarda, the Power Princess -----------------------------------------------------

func mscM11ZardaBoard(t *testing.T) (g *game.Game, zarda, hero1, hero2, bear, theirHero uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zarda = effectstest.B12Push(g, me.ID, "Zarda, the Power Princess", "Legendary Creature — Human Warrior Hero", mscM11ZardaOracle, 2, 4)
	hero1 = effectstest.B12Creature(g, me.ID, "Hero One", "Creature — Human Hero", 2, 2)
	hero2 = effectstest.B12Creature(g, me.ID, "Hero Two", "Creature — Human Hero", 2, 2)
	bear = effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirHero = effectstest.B12Creature(g, opp.ID, "Their Hero", "Creature — Human Hero", 2, 2)
	return
}

func TestMscM11ZardaGivesOtherHeroesYouControlExaltedAndNobodyElse(t *testing.T) {
	g, zarda, hero1, _, bear, theirHero := mscM11ZardaBoard(t)
	effectstest.AssertKeywords(t, g, zarda, "flying")
	if !effectstest.EffectiveAbilitiesContain(t, g, hero1, "exalted") {
		t.Error("another Hero you control lacks exalted")
	}
	if effectstest.EffectiveAbilitiesContain(t, g, zarda, "exalted") {
		t.Error("Zarda has exalted herself, but the text says other Heroes")
	}
	if effectstest.EffectiveAbilitiesContain(t, g, bear, "exalted") {
		t.Error("a non-Hero has exalted")
	}
	if effectstest.EffectiveAbilitiesContain(t, g, theirHero, "exalted") {
		t.Error("an opponent's Hero has exalted")
	}
}

func TestMscM11ZardaLoneAttackerGetsOneBonusPerInstance(t *testing.T) {
	g, _, hero1, _, _, _ := mscM11ZardaBoard(t)
	effectstest.AttackWith(t, g, g.Seats[1].ID, hero1)
	// Two Heroes, two exalted instances among permanents I control.
	if got := effectstest.HobA4Power(t, g, hero1); got != 4 {
		t.Errorf("the lone attacking Hero has power %d, want 4 (2 base + 2 exalted instances)", got)
	}
}

func TestMscM11ZardaGivesNothingWhenTheHeroLeaves(t *testing.T) {
	g, zarda, hero1, _, _, _ := mscM11ZardaBoard(t)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(zarda) })
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(zarda) {
		t.Fatal("Zarda did not leave")
	}
	if effectstest.EffectiveAbilitiesContain(t, g, hero1, "exalted") {
		t.Error("exalted stayed after Zarda left")
	}
}

// --- Swordsman's Steel -------------------------------------------------------------

func TestMscM11SteelDrawsOneCardPerEquipmentYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedEquipment(g, me.ID, "Spare Blade", "")
	effectstest.SeedEquipment(g, opp.ID, "Their Blade", "")
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Swordsman's Steel", "Artifact — Equipment", mscM11SteelOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// The Steel left the hand when cast (-1 for the pushed card, +0), then drew
	// one card for itself and one for the Spare Blade; the opponent's Blade
	// is not "you control".
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("hand grew by %d, want 2 (Steel + Spare Blade, not the opponent's)", got)
	}
}

func TestMscM11SteelPumpsTheEquippedCreatureByTwoPerEquipmentYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	steel := effectstest.SeedEquipment(g, me.ID, "Swordsman's Steel", mscM11SteelOracle)
	effectstest.EquipTo(t, g, me.ID, steel, bear)
	if got := effectstest.EffectivePower(t, g, bear); got != 4 {
		t.Fatalf("with only the Steel the bear has power %d, want 4 (+2/+2 for one Equipment)", got)
	}
	effectstest.SeedEquipment(g, me.ID, "Spare Blade", "")
	if got := effectstest.EffectivePower(t, g, bear); got != 6 {
		t.Errorf("with two Equipment the bear has power %d, want 6", got)
	}
	if got := effectstest.EffectiveToughness(t, g, bear); got != 6 {
		t.Errorf("with two Equipment the bear has toughness %d, want 6", got)
	}
	effectstest.SeedEquipment(g, opp.ID, "Their Blade", "")
	if got := effectstest.EffectivePower(t, g, bear); got != 6 {
		t.Errorf("an opponent's Equipment changed the pump: power %d, want 6", got)
	}
}

// --- Ultron, Machine Overlord ------------------------------------------------------

func TestMscM11UltronPumpsOtherRobotsAndConstructsYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ultron := effectstest.B12Push(g, me.ID, "Ultron, Machine Overlord", "Legendary Artifact Creature — Robot Villain", mscM11UltronOracle, 4, 4)
	effectstest.AssertKeywords(t, g, ultron, "flying")
	robot := effectstest.B12Creature(g, me.ID, "Robot", "Artifact Creature — Robot", 1, 1)
	construct := effectstest.B12Creature(g, me.ID, "Construct", "Artifact Creature — Construct", 1, 1)
	both := effectstest.B12Creature(g, me.ID, "Robot Construct", "Artifact Creature — Robot Construct", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Robot", "Artifact Creature — Robot", 1, 1)

	for name, want := range map[string]struct {
		id   uuid.UUID
		p, t int
	}{
		"Robot": {robot, 3, 3}, "Construct": {construct, 3, 3}, "Robot Construct": {both, 3, 3},
		"Ultron himself": {ultron, 4, 4}, "Bear": {bear, 2, 2}, "an opposing Robot": {theirs, 1, 1},
	} {
		if p, tough := effectstest.HobA4Power(t, g, want.id), effectstest.EffectiveToughness(t, g, want.id); p != want.p || tough != want.t {
			t.Errorf("%s is %d/%d, want %d/%d", name, p, tough, want.p, want.t)
		}
	}
}

// --- Marvel Boy, Noh-Varr ----------------------------------------------------------

func TestMscM11MarvelBoyGrowsWhenAnotherCreatureOfYoursEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	boy := effectstest.PushCatalogPermanent(g, me.ID, "Marvel Boy, Noh-Varr", "Legendary Creature — Kree Soldier Hero", mscM11MarvelBoyOracle, false)
	effectstest.B12PlayFromHand(t, g, "Bear", "Creature — Bear", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, boy, game.CounterPlusOne); got != 1 {
		t.Fatalf("Marvel Boy has %d counters after a creature entered, want 1", got)
	}
	effectstest.MscM11Token(t, g, me.ID, "1/1 white Soldier")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, boy, game.CounterPlusOne); got != 2 {
		t.Errorf("a token entering gave %d counters in all, want 2", got)
	}
	effectstest.MscM11Token(t, g, opp.ID, "1/1 white Soldier")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, boy, game.CounterPlusOne); got != 2 {
		t.Errorf("an opponent's creature entering changed the counters: %d, want 2", got)
	}
}

func TestMscM11MarvelBoyDoesNotGrowWhenHeEntersHimself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	boy := effectstest.CastAndResolveCreature(t, g, "Marvel Boy, Noh-Varr", "Legendary Creature — Kree Soldier Hero", mscM11MarvelBoyOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, boy, game.CounterPlusOne); got != 0 {
		t.Errorf("Marvel Boy put a counter on himself as he entered: %d", got)
	}
}

func TestMscM11MarvelBoyGrowsWhenYouActivateAPowerUpAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.PowerUpMain(t, g)
	me := g.Seats[0]
	boy := effectstest.PushCatalogPermanent(g, me.ID, "Marvel Boy, Noh-Varr", "Legendary Creature — Kree Soldier Hero", mscM11MarvelBoyOracle, false)
	brawler := effectstest.PushPowerUpPermanent(g, me.ID, "Brave Brawler", "Creature — Human Warrior Hero", "{1}{W}", effectstest.BraveBrawlerOracle)
	if got := effectstest.CountersOn(g, boy, game.CounterPlusOne); got != 0 {
		t.Fatalf("Marvel Boy starts with %d counters", got)
	}
	effectstest.FloatMana(t, g, me, "{W}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, brawler, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("power up: %v", err)
	}
	effectstest.ResolveTopOfStack(t, g)
	effectstest.ResolveTopOfStack(t, g)
	if got := effectstest.CountersOn(g, boy, game.CounterPlusOne); got != 1 {
		t.Errorf("Marvel Boy has %d counters after a power-up activation, want 1", got)
	}
	if got := effectstest.PlusOnes(g, brawler); got != 2 {
		t.Errorf("the Brawler has %d counters, want its own 2", got)
	}
}

// --- Strategic Intervention ---------------------------------------------------------

func mscM11StrategicBoard(t *testing.T) (g *game.Game, bear, other, theirs uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Strategic Intervention", "Enchantment", mscM11StrategicOracle, false)
	bear = effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other = effectstest.B12Creature(g, me.ID, "Other Bear", "Creature — Bear", 2, 2)
	theirs = effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	return
}

func TestMscM11StrategicInterventionPumpsTheLoneAttackerAndTapsADefendersCreature(t *testing.T) {
	g, bear, _, theirs := mscM11StrategicBoard(t)
	effectstest.DeclareAttack(t, g, g.Seats[1].ID, bear)
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HobA4Power(t, g, bear); got != 3 {
		t.Errorf("the lone attacker has power %d, want 3", got)
	}
	if !effectstest.BattlefieldCardPtr(t, g, theirs).Tapped {
		t.Error("the targeted defending creature is not tapped")
	}
}

func TestMscM11StrategicInterventionMayTapNothing(t *testing.T) {
	g, bear, _, theirs := mscM11StrategicBoard(t)
	effectstest.DeclareAttack(t, g, g.Seats[1].ID, bear)
	ch := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTargets(ch.ID, ch.Chooser, nil); err != nil {
		t.Fatalf("choosing no target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HobA4Power(t, g, bear); got != 3 {
		t.Errorf("the lone attacker has power %d, want 3 even with no tap target", got)
	}
	if effectstest.BattlefieldCardPtr(t, g, theirs).Tapped {
		t.Error("a creature was tapped though no target was chosen")
	}
}

func TestMscM11StrategicInterventionOnlyTargetsTheDefendingPlayersCreatures(t *testing.T) {
	g, bear, other, theirs := mscM11StrategicBoard(t)
	third := effectstest.B12Creature(g, g.Seats[2].ID, "Third Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, g.Seats[1].ID, bear)
	ch := effectstest.PickTargetPrompt(t, g)
	for _, bad := range []uuid.UUID{other, third} {
		if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Errorf("target %s (not controlled by the defending player) was accepted", bad)
		}
	}
	effectstest.AnswerPickTarget(t, g, theirs)
}

func TestMscM11StrategicInterventionDoesNothingWhenTwoCreaturesAttack(t *testing.T) {
	g, bear, other, theirs := mscM11StrategicBoard(t)
	effectstest.DeclareAttack(t, g, g.Seats[1].ID, bear, other)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HobA4Power(t, g, bear); got != 2 {
		t.Errorf("an attacker that was not alone has power %d, want 2", got)
	}
	if effectstest.BattlefieldCardPtr(t, g, theirs).Tapped {
		t.Error("a creature was tapped though the attack was not alone")
	}
}

// --- M'Baku, Jabari Chieftain ------------------------------------------------------

func mscM11SetMonarch(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { g.Monarch = id })
}

func TestMscM11MBakuCrownsTheChosenOpponentWhenThereIsNoMonarch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "M'Baku, Jabari Chieftain", "Legendary Creature — Human Noble Warrior", mscM11MBakuOracle, false)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	ch := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); err == nil {
		t.Error("M'Baku's controller was accepted as \"target opponent\"")
	}
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Monarch != opp.ID {
		t.Errorf("monarch is %s, want the chosen opponent", g.Monarch)
	}
}

func TestMscM11MBakuDoesNothingWhenSomeoneIsAlreadyMonarch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "M'Baku, Jabari Chieftain", "Legendary Creature — Human Noble Warrior", mscM11MBakuOracle, false)
	mscM11SetMonarch(g, me.ID)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			t.Fatal("the end step trigger asked for a target though there is a monarch")
		}
	}
	if g.Monarch != me.ID {
		t.Errorf("the crown moved to %s though the trigger needs no monarch", g.Monarch)
	}
	_ = opp
}

func mscM11MBakuAttackBoard(t *testing.T) (g *game.Game, bear uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "M'Baku, Jabari Chieftain", "Legendary Creature — Human Noble Warrior", mscM11MBakuOracle, false)
	bear = effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	return
}

func TestMscM11MBakuPumpsACreatureAttackingTheMonarchOpponent(t *testing.T) {
	g, bear := mscM11MBakuAttackBoard(t)
	opp := g.Seats[1]
	mscM11SetMonarch(g, opp.ID)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HobA4Power(t, g, bear); got != 3 {
		t.Errorf("the attacker has power %d, want 3", got)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, bear, "trample") {
		t.Error("the attacker did not gain trample")
	}
}

func TestMscM11MBakuDoesNothingWhenTheDefenderIsNotTheMonarch(t *testing.T) {
	g, bear := mscM11MBakuAttackBoard(t)
	opp, third := g.Seats[1], g.Seats[2]
	mscM11SetMonarch(g, third.ID)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HobA4Power(t, g, bear); got != 2 {
		t.Errorf("attacking a non-monarch gave power %d, want 2", got)
	}
	if effectstest.EffectiveAbilitiesContain(t, g, bear, "trample") {
		t.Error("trample was granted though the defender is not the monarch")
	}
}

// #4839, CR 508.3a: a creature attacking the monarch's planeswalker is not
// attacking the monarch.
func TestMscM11MBakuDoesNothingForAnAttackOnTheMonarchsPlaneswalker(t *testing.T) {
	g, bear := mscM11MBakuAttackBoard(t)
	opp := g.Seats[1]
	walker := effectstest.P1g2Planeswalker(g, opp.ID, 5)
	mscM11SetMonarch(g, opp.ID)
	effectstest.DeclareAttack(t, g, walker, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HobA4Power(t, g, bear); got != 2 {
		t.Errorf("a creature attacking the monarch's planeswalker has power %d, want 2", got)
	}
	if effectstest.EffectiveAbilitiesContain(t, g, bear, "trample") {
		t.Error("trample was granted to a creature attacking a planeswalker")
	}
}

func TestMscM11MBakuDoesNothingWhenAnOpponentAttacksYouWhileYouAreMonarch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "M'Baku, Jabari Chieftain", "Legendary Creature — Human Noble Warrior", mscM11MBakuOracle, false)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mscM11SetMonarch(g, me.ID)
	effectstest.PowerUpPassTo(t, g, 1)
	effectstest.DeclareAttack(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HobA4Power(t, g, theirs); got != 2 {
		t.Errorf("a creature attacking M'Baku's controller got power %d, want 2 (not \"one of your opponents\")", got)
	}
}

func TestMscM11MBakuPumpsAnotherPlayersCreatureAttackingTheMonarchOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.PushCatalogPermanent(g, me.ID, "M'Baku, Jabari Chieftain", "Legendary Creature — Human Noble Warrior", mscM11MBakuOracle, false)
	theirs := effectstest.B12Creature(g, third.ID, "Third Bear", "Creature — Bear", 2, 2)
	mscM11SetMonarch(g, opp.ID)
	effectstest.PowerUpPassTo(t, g, 2)
	effectstest.DeclareAttack(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HobA4Power(t, g, theirs); got != 3 {
		t.Errorf("an opposing creature attacking the monarch opponent has power %d, want 3 (any creature)", got)
	}
}
