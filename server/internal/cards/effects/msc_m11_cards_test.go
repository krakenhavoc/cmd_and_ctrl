package effects

import (
	"testing"

	"github.com/google/uuid"

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
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// mscM11Token makes a token under `owner`'s control, as an effect would.
func mscM11Token(t *testing.T, g *game.Game, owner uuid.UUID, row string) {
	t.Helper()
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{ID: uuid.New(), Controller: owner, SourceCardID: uuid.Nil})
		if err := (CreateToken{Template: TokenCard(row), N: 1, Controller: owner}).Apply(ctx); err != nil {
			t.Fatalf("token %q: %v", row, err)
		}
	})
}

// --- Winter Soldier, Bucky Barnes ---------------------------------------------

func TestMscM11BuckyEntersTapped(t *testing.T) {
	g := newCatalogGame(t)
	id := castAndResolveCreature(t, g, "Winter Soldier, Bucky Barnes", "Legendary Creature — Human Soldier Hero", mscM11BuckyOracle)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Bucky did not enter")
	}
	if !battlefieldCardPtr(t, g, id).Tapped {
		t.Error("Winter Soldier entered untapped")
	}
}

func TestMscM11BuckyDoesNotTapOtherCreatures(t *testing.T) {
	g := newCatalogGame(t)
	pushCatalogPermanent(g, g.Seats[0].ID, "Winter Soldier, Bucky Barnes", "Legendary Creature — Human Soldier Hero", mscM11BuckyOracle, false)
	bear := b12PlayFromHand(t, g, "Bear", "Creature — Bear", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if battlefieldCardPtr(t, g, bear).Tapped {
		t.Error("another creature entered tapped")
	}
}

// --- Wild Pack Squad ---------------------------------------------------------------

func mscM11SquadAtCombat(t *testing.T) (*game.Game, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Wild Pack Squad", "Creature — Human Mercenary", mscM11WildPackOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepBeginCombat)
	return g, bear
}

func TestMscM11WildPackSquadGivesFirstStrikeAndVigilanceToTheTarget(t *testing.T) {
	g, bear := mscM11SquadAtCombat(t)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if !effectiveAbilitiesContain(t, g, bear, "first strike") || !effectiveAbilitiesContain(t, g, bear, "vigilance") {
		t.Errorf("the target has %v, want first strike and vigilance", effectiveAbilities(t, g, bear))
	}
}

func TestMscM11WildPackSquadMayTargetAnyCreatureAndLastsOnlyThisTurn(t *testing.T) {
	g, _ := mscM11SquadAtCombat(t)
	theirs := b12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	if !effectiveAbilitiesContain(t, g, theirs, "first strike") {
		t.Fatal("an opponent's creature is a legal target and gained nothing")
	}
	powerUpPassTo(t, g, 0)
	if effectiveAbilitiesContain(t, g, theirs, "first strike") || effectiveAbilitiesContain(t, g, theirs, "vigilance") {
		t.Error("the grant outlasted the turn")
	}
}

func TestMscM11WildPackSquadMayTargetNothing(t *testing.T) {
	g, bear := mscM11SquadAtCombat(t)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTargets(ch.ID, ch.Chooser, nil); err != nil {
		t.Fatalf("choosing no target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if effectiveAbilitiesContain(t, g, bear, "first strike") {
		t.Error("a creature gained first strike with no target chosen")
	}
}

// --- The Frightful Four ------------------------------------------------------------

func TestMscM11FrightfulFourDrainsTheFirstNoncreatureSpellOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	four := pushCatalogPermanent(g, me.ID, "The Frightful Four", "Legendary Creature — Villain", mscM11FrightfulOracle, false)
	assertKeywords(t, g, four, "menace")
	life := opp.Life

	batch01OpponentCasts(t, g, opp, "First Spell", "", "{2}{R}", nil)
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Fatalf("the caster has %d life, want %d (mana value 3)", opp.Life, life-3)
	}
	batch01OpponentCasts(t, g, opp, "Second Spell", "", "{4}", nil)
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("the second noncreature spell cost life: %d, want %d", opp.Life, life-3)
	}
}

func TestMscM11FrightfulFourIgnoresYourOwnSpellsAndOnlyTheCasterLoses(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "The Frightful Four", "Legendary Creature — Villain", mscM11FrightfulOracle, false)
	life, thirdLife := me.Life, third.Life

	castFree(t, g, me, "My Spell", "Instant")
	if me.Life != life {
		t.Errorf("my own spell cost me life: %d -> %d", life, me.Life)
	}
	oppLife := opp.Life
	batch01OpponentCasts(t, g, opp, "Their Spell", "", "{1}", nil)
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-1 {
		t.Errorf("the caster has %d life, want %d", opp.Life, oppLife-1)
	}
	if third.Life != thirdLife || me.Life != life {
		t.Errorf("someone other than the caster lost life: me %d, third %d", me.Life, third.Life)
	}
}

// --- White Widow, Yelena Belova ----------------------------------------------------

func TestMscM11WhiteWidowGrowsWhenSheConnects(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	widow := b12Push(g, me.ID, "White Widow, Yelena Belova", "Legendary Creature — Human Assassin Villain", mscM11WhiteWidowOracle, 1, 2)
	assertKeywords(t, g, widow, "deathtouch")
	declareAttack(t, g, opp.ID, widow)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, widow, game.CounterPlusOne); got != 1 {
		t.Errorf("White Widow has %d +1/+1 counters, want 1", got)
	}
}

func TestMscM11WhiteWidowPutsTheCounterOnAnotherDeathtouchCreatureNotOnHerself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	widow := b12Push(g, me.ID, "White Widow, Yelena Belova", "Legendary Creature — Human Assassin Villain", mscM11WhiteWidowOracle, 1, 2)
	snake := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Deathtouch Snake", TypeLine: "Creature — Snake",
		Power: 1, Toughness: 1, Keywords: []string{"deathtouch"}, Owner: me.ID, Controller: me.ID,
	})
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, snake, bear)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, snake, game.CounterPlusOne); got != 1 {
		t.Errorf("the deathtouch Snake has %d counters, want 1", got)
	}
	if got := countersOn(g, bear, game.CounterPlusOne); got != 0 {
		t.Errorf("the creature without deathtouch got %d counters, want 0", got)
	}
	if got := countersOn(g, widow, game.CounterPlusOne); got != 0 {
		t.Errorf("White Widow did not attack and has %d counters, want 0", got)
	}
}

func TestMscM11WhiteWidowGetsNothingWhenBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	widow := b12Push(g, me.ID, "White Widow, Yelena Belova", "Legendary Creature — Human Assassin Villain", mscM11WhiteWidowOracle, 1, 2)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(widow, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, widow); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, widow, game.CounterPlusOne); got != 0 {
		t.Errorf("a blocked White Widow got %d counters, want 0 (no damage to a player)", got)
	}
}

// --- X-23, Deadly Weapon -----------------------------------------------------------

func TestMscM11X23CountersGoOnTheEnteringMutantAndOnX23(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	x23 := pushCatalogPermanent(g, me.ID, "X-23, Deadly Weapon", "Legendary Creature — Mutant Berserker Hero", mscM11X23Oracle, false)
	mutant := b12PlayFromHand(t, g, "Test Mutant", "Creature — Mutant", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if got := countersOn(g, mutant, game.CounterPlusOne); got != 1 {
		t.Errorf("the Mutant has %d counters, want 1", got)
	}
	if got := countersOn(g, x23, game.CounterPlusOne); got != 1 {
		t.Errorf("X-23 has %d counters, want 1", got)
	}
}

func TestMscM11X23IgnoresNonMutantsAnOpponentsMutantAndItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	x23 := pushCatalogPermanent(g, me.ID, "X-23, Deadly Weapon", "Legendary Creature — Mutant Berserker Hero", mscM11X23Oracle, false)
	b12PlayFromHand(t, g, "Plain Bear", "Creature — Bear", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	mscM11Token(t, g, opp.ID, "1/1 white Soldier")
	passPriorityAroundTable(t, g)
	if got := countersOn(g, x23, game.CounterPlusOne); got != 0 {
		t.Errorf("X-23 has %d counters after a Bear and an opposing creature, want 0", got)
	}
	// X-23 entering herself is "another": cast her and nothing happens.
	g2 := newCatalogGame(t)
	self := castAndResolveCreature(t, g2, "X-23, Deadly Weapon", "Legendary Creature — Mutant Berserker Hero", mscM11X23Oracle)
	passPriorityAroundTable(t, g2)
	if got := countersOn(g2, self, game.CounterPlusOne); got != 0 {
		t.Errorf("X-23 put a counter on herself as she entered: %d", got)
	}
}

// --- Zarda, the Power Princess -----------------------------------------------------

func mscM11ZardaBoard(t *testing.T) (g *game.Game, zarda, hero1, hero2, bear, theirHero uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zarda = b12Push(g, me.ID, "Zarda, the Power Princess", "Legendary Creature — Human Warrior Hero", mscM11ZardaOracle, 2, 4)
	hero1 = b12Creature(g, me.ID, "Hero One", "Creature — Human Hero", 2, 2)
	hero2 = b12Creature(g, me.ID, "Hero Two", "Creature — Human Hero", 2, 2)
	bear = b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirHero = b12Creature(g, opp.ID, "Their Hero", "Creature — Human Hero", 2, 2)
	return
}

func TestMscM11ZardaGivesOtherHeroesYouControlExaltedAndNobodyElse(t *testing.T) {
	g, zarda, hero1, _, bear, theirHero := mscM11ZardaBoard(t)
	assertKeywords(t, g, zarda, "flying")
	if !effectiveAbilitiesContain(t, g, hero1, "exalted") {
		t.Error("another Hero you control lacks exalted")
	}
	if effectiveAbilitiesContain(t, g, zarda, "exalted") {
		t.Error("Zarda has exalted herself, but the text says other Heroes")
	}
	if effectiveAbilitiesContain(t, g, bear, "exalted") {
		t.Error("a non-Hero has exalted")
	}
	if effectiveAbilitiesContain(t, g, theirHero, "exalted") {
		t.Error("an opponent's Hero has exalted")
	}
}

func TestMscM11ZardaLoneAttackerGetsOneBonusPerInstance(t *testing.T) {
	g, _, hero1, _, _, _ := mscM11ZardaBoard(t)
	attackWith(t, g, g.Seats[1].ID, hero1)
	// Two Heroes, two exalted instances among permanents I control.
	if got := hobA4Power(t, g, hero1); got != 4 {
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
	if effectiveAbilitiesContain(t, g, hero1, "exalted") {
		t.Error("exalted stayed after Zarda left")
	}
}

// --- Swordsman's Steel -------------------------------------------------------------

func TestMscM11SteelDrawsOneCardPerEquipmentYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	seedEquipment(g, me.ID, "Spare Blade", "")
	seedEquipment(g, opp.ID, "Their Blade", "")
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Swordsman's Steel", "Artifact — Equipment", mscM11SteelOracle, nil)
	passPriorityAroundTable(t, g)
	// The Steel left the hand when cast (-1 for the pushed card, +0), then drew
	// one card for itself and one for the Spare Blade; the opponent's Blade
	// is not "you control".
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("hand grew by %d, want 2 (Steel + Spare Blade, not the opponent's)", got)
	}
}

func TestMscM11SteelPumpsTheEquippedCreatureByTwoPerEquipmentYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	steel := seedEquipment(g, me.ID, "Swordsman's Steel", mscM11SteelOracle)
	equipTo(t, g, me.ID, steel, bear)
	if got := effectivePower(t, g, bear); got != 4 {
		t.Fatalf("with only the Steel the bear has power %d, want 4 (+2/+2 for one Equipment)", got)
	}
	seedEquipment(g, me.ID, "Spare Blade", "")
	if got := effectivePower(t, g, bear); got != 6 {
		t.Errorf("with two Equipment the bear has power %d, want 6", got)
	}
	if got := effectiveToughness(t, g, bear); got != 6 {
		t.Errorf("with two Equipment the bear has toughness %d, want 6", got)
	}
	seedEquipment(g, opp.ID, "Their Blade", "")
	if got := effectivePower(t, g, bear); got != 6 {
		t.Errorf("an opponent's Equipment changed the pump: power %d, want 6", got)
	}
}

// --- Ultron, Machine Overlord ------------------------------------------------------

func TestMscM11UltronPumpsOtherRobotsAndConstructsYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ultron := b12Push(g, me.ID, "Ultron, Machine Overlord", "Legendary Artifact Creature — Robot Villain", mscM11UltronOracle, 4, 4)
	assertKeywords(t, g, ultron, "flying")
	robot := b12Creature(g, me.ID, "Robot", "Artifact Creature — Robot", 1, 1)
	construct := b12Creature(g, me.ID, "Construct", "Artifact Creature — Construct", 1, 1)
	both := b12Creature(g, me.ID, "Robot Construct", "Artifact Creature — Robot Construct", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Robot", "Artifact Creature — Robot", 1, 1)

	for name, want := range map[string]struct {
		id   uuid.UUID
		p, t int
	}{
		"Robot": {robot, 3, 3}, "Construct": {construct, 3, 3}, "Robot Construct": {both, 3, 3},
		"Ultron himself": {ultron, 4, 4}, "Bear": {bear, 2, 2}, "an opposing Robot": {theirs, 1, 1},
	} {
		if p, tough := hobA4Power(t, g, want.id), effectiveToughness(t, g, want.id); p != want.p || tough != want.t {
			t.Errorf("%s is %d/%d, want %d/%d", name, p, tough, want.p, want.t)
		}
	}
}

// --- Marvel Boy, Noh-Varr ----------------------------------------------------------

func TestMscM11MarvelBoyGrowsWhenAnotherCreatureOfYoursEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	boy := pushCatalogPermanent(g, me.ID, "Marvel Boy, Noh-Varr", "Legendary Creature — Kree Soldier Hero", mscM11MarvelBoyOracle, false)
	b12PlayFromHand(t, g, "Bear", "Creature — Bear", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if got := countersOn(g, boy, game.CounterPlusOne); got != 1 {
		t.Fatalf("Marvel Boy has %d counters after a creature entered, want 1", got)
	}
	mscM11Token(t, g, me.ID, "1/1 white Soldier")
	passPriorityAroundTable(t, g)
	if got := countersOn(g, boy, game.CounterPlusOne); got != 2 {
		t.Errorf("a token entering gave %d counters in all, want 2", got)
	}
	mscM11Token(t, g, opp.ID, "1/1 white Soldier")
	passPriorityAroundTable(t, g)
	if got := countersOn(g, boy, game.CounterPlusOne); got != 2 {
		t.Errorf("an opponent's creature entering changed the counters: %d, want 2", got)
	}
}

func TestMscM11MarvelBoyDoesNotGrowWhenHeEntersHimself(t *testing.T) {
	g := newCatalogGame(t)
	boy := castAndResolveCreature(t, g, "Marvel Boy, Noh-Varr", "Legendary Creature — Kree Soldier Hero", mscM11MarvelBoyOracle)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, boy, game.CounterPlusOne); got != 0 {
		t.Errorf("Marvel Boy put a counter on himself as he entered: %d", got)
	}
}

func TestMscM11MarvelBoyGrowsWhenYouActivateAPowerUpAbility(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	boy := pushCatalogPermanent(g, me.ID, "Marvel Boy, Noh-Varr", "Legendary Creature — Kree Soldier Hero", mscM11MarvelBoyOracle, false)
	brawler := pushPowerUpPermanent(g, me.ID, "Brave Brawler", "Creature — Human Warrior Hero", "{1}{W}", braveBrawlerOracle)
	if got := countersOn(g, boy, game.CounterPlusOne); got != 0 {
		t.Fatalf("Marvel Boy starts with %d counters", got)
	}
	floatMana(t, g, me, "{W}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, brawler, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("power up: %v", err)
	}
	resolveTopOfStack(t, g)
	resolveTopOfStack(t, g)
	if got := countersOn(g, boy, game.CounterPlusOne); got != 1 {
		t.Errorf("Marvel Boy has %d counters after a power-up activation, want 1", got)
	}
	if got := plusOnes(g, brawler); got != 2 {
		t.Errorf("the Brawler has %d counters, want its own 2", got)
	}
}

// --- Strategic Intervention ---------------------------------------------------------

func mscM11StrategicBoard(t *testing.T) (g *game.Game, bear, other, theirs uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Strategic Intervention", "Enchantment", mscM11StrategicOracle, false)
	bear = b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other = b12Creature(g, me.ID, "Other Bear", "Creature — Bear", 2, 2)
	theirs = b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	return
}

func TestMscM11StrategicInterventionPumpsTheLoneAttackerAndTapsADefendersCreature(t *testing.T) {
	g, bear, _, theirs := mscM11StrategicBoard(t)
	declareAttack(t, g, g.Seats[1].ID, bear)
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	if got := hobA4Power(t, g, bear); got != 3 {
		t.Errorf("the lone attacker has power %d, want 3", got)
	}
	if !battlefieldCardPtr(t, g, theirs).Tapped {
		t.Error("the targeted defending creature is not tapped")
	}
}

func TestMscM11StrategicInterventionMayTapNothing(t *testing.T) {
	g, bear, _, theirs := mscM11StrategicBoard(t)
	declareAttack(t, g, g.Seats[1].ID, bear)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTargets(ch.ID, ch.Chooser, nil); err != nil {
		t.Fatalf("choosing no target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := hobA4Power(t, g, bear); got != 3 {
		t.Errorf("the lone attacker has power %d, want 3 even with no tap target", got)
	}
	if battlefieldCardPtr(t, g, theirs).Tapped {
		t.Error("a creature was tapped though no target was chosen")
	}
}

func TestMscM11StrategicInterventionOnlyTargetsTheDefendingPlayersCreatures(t *testing.T) {
	g, bear, other, theirs := mscM11StrategicBoard(t)
	third := b12Creature(g, g.Seats[2].ID, "Third Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, g.Seats[1].ID, bear)
	ch := pickTargetPrompt(t, g)
	for _, bad := range []uuid.UUID{other, third} {
		if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Errorf("target %s (not controlled by the defending player) was accepted", bad)
		}
	}
	answerPickTarget(t, g, theirs)
}

func TestMscM11StrategicInterventionDoesNothingWhenTwoCreaturesAttack(t *testing.T) {
	g, bear, other, theirs := mscM11StrategicBoard(t)
	declareAttack(t, g, g.Seats[1].ID, bear, other)
	passPriorityAroundTable(t, g)
	if got := hobA4Power(t, g, bear); got != 2 {
		t.Errorf("an attacker that was not alone has power %d, want 2", got)
	}
	if battlefieldCardPtr(t, g, theirs).Tapped {
		t.Error("a creature was tapped though the attack was not alone")
	}
}

// --- M'Baku, Jabari Chieftain ------------------------------------------------------

func mscM11SetMonarch(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { g.Monarch = id })
}

func TestMscM11MBakuCrownsTheChosenOpponentWhenThereIsNoMonarch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "M'Baku, Jabari Chieftain", "Legendary Creature — Human Noble Warrior", mscM11MBakuOracle, false)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); err == nil {
		t.Error("M'Baku's controller was accepted as \"target opponent\"")
	}
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if g.Monarch != opp.ID {
		t.Errorf("monarch is %s, want the chosen opponent", g.Monarch)
	}
}

func TestMscM11MBakuDoesNothingWhenSomeoneIsAlreadyMonarch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "M'Baku, Jabari Chieftain", "Legendary Creature — Human Noble Warrior", mscM11MBakuOracle, false)
	mscM11SetMonarch(g, me.ID)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
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
	g = newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "M'Baku, Jabari Chieftain", "Legendary Creature — Human Noble Warrior", mscM11MBakuOracle, false)
	bear = b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	return
}

func TestMscM11MBakuPumpsACreatureAttackingTheMonarchOpponent(t *testing.T) {
	g, bear := mscM11MBakuAttackBoard(t)
	opp := g.Seats[1]
	mscM11SetMonarch(g, opp.ID)
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if got := hobA4Power(t, g, bear); got != 3 {
		t.Errorf("the attacker has power %d, want 3", got)
	}
	if !effectiveAbilitiesContain(t, g, bear, "trample") {
		t.Error("the attacker did not gain trample")
	}
}

func TestMscM11MBakuDoesNothingWhenTheDefenderIsNotTheMonarch(t *testing.T) {
	g, bear := mscM11MBakuAttackBoard(t)
	opp, third := g.Seats[1], g.Seats[2]
	mscM11SetMonarch(g, third.ID)
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if got := hobA4Power(t, g, bear); got != 2 {
		t.Errorf("attacking a non-monarch gave power %d, want 2", got)
	}
	if effectiveAbilitiesContain(t, g, bear, "trample") {
		t.Error("trample was granted though the defender is not the monarch")
	}
}

func TestMscM11MBakuDoesNothingWhenAnOpponentAttacksYouWhileYouAreMonarch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "M'Baku, Jabari Chieftain", "Legendary Creature — Human Noble Warrior", mscM11MBakuOracle, false)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mscM11SetMonarch(g, me.ID)
	powerUpPassTo(t, g, 1)
	declareAttack(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if got := hobA4Power(t, g, theirs); got != 2 {
		t.Errorf("a creature attacking M'Baku's controller got power %d, want 2 (not \"one of your opponents\")", got)
	}
}

func TestMscM11MBakuPumpsAnotherPlayersCreatureAttackingTheMonarchOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	pushCatalogPermanent(g, me.ID, "M'Baku, Jabari Chieftain", "Legendary Creature — Human Noble Warrior", mscM11MBakuOracle, false)
	theirs := b12Creature(g, third.ID, "Third Bear", "Creature — Bear", 2, 2)
	mscM11SetMonarch(g, opp.ID)
	powerUpPassTo(t, g, 2)
	declareAttack(t, g, opp.ID, theirs)
	passPriorityAroundTable(t, g)
	if got := hobA4Power(t, g, theirs); got != 3 {
		t.Errorf("an opposing creature attacking the monarch opponent has power %d, want 3 (any creature)", got)
	}
}
