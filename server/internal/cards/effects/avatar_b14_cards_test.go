package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b14_cards_test.go — Avatar: The Last Airbender, batch 14 (#4459).
// Each test plays the real catalog entry through the engine.

const (
	avB14CubOracle        = "2b0afb89-0944-4861-b9c3-e909e2ac215e"
	avB14BoulderOracle    = "d33ee141-a880-4368-a274-a6ae2f2c624b"
	avB14BumiOracle       = "f7493bcf-5c17-4562-a759-37082ec49ff8"
	avB14ZookeeperOracle  = "6a84a8a4-c8b0-45ce-af6f-605fcffc395b"
	avB14GeneralOracle    = "e43c31ee-1bac-4f96-8b74-334b3238ab73"
	avB14RumbleOracle     = "f04a7763-5fba-42e5-a45a-b3dc2daba902"
	avB14AscensionOracle  = "f8966fd5-2926-4ba0-b7d6-cfea40cd4e05"
	avB14FlopsieOracle    = "a8966245-501c-47e4-8ce0-849dcb532eb2"
	avB14VinebenderOracle = "82fefb82-a436-488d-8ecd-594686117e34"
	avB14HaruOracle       = "7488ce54-37c6-475c-b327-59f52f2b1424"
)

func TestAvatarB14EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		avB14CubOracle:        "Badgermole Cub",
		avB14BoulderOracle:    "The Boulder, Ready to Rumble",
		avB14BumiOracle:       "Bumi, King of Three Trials",
		avB14ZookeeperOracle:  "Diligent Zookeeper",
		avB14GeneralOracle:    "Earth Kingdom General",
		avB14RumbleOracle:     "Earth Rumble",
		avB14AscensionOracle:  "Earthbender Ascension",
		avB14FlopsieOracle:    "Flopsie, Bumi's Buddy",
		avB14VinebenderOracle: "Foggy Swamp Vinebender",
		avB14HaruOracle:       "Haru, Hidden Talent",
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

// --- fixtures ------------------------------------------------------------------

// avB14Enter casts a creature from p's hand and resolves the spell, its
// entry triggers left for the caller.
func avB14Enter(t *testing.T, g *game.Game, p *game.Player, c game.Card) uuid.UUID {
	t.Helper()
	id := avB1Cast(t, g, p, c, game.CastSpellParams{})
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	if !g.Battlefield.Contains(id) {
		t.Fatalf("%s never reached the battlefield", c.Name)
	}
	return id
}

// avB14Forest parks an untapped Forest under owner.
func avB14Forest(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushEarthbendLand(g, owner, "Forest", "Basic Land — Forest")
}

// avB14Earthbent asserts a land is an animated land creature with n
// +1/+1 counters.
func avB14Earthbent(t *testing.T, g *game.Game, land uuid.UUID, n int) {
	t.Helper()
	power, toughness, creatureLand, hasty := earthbentBody(t, g, land)
	if !creatureLand || !hasty {
		t.Errorf("land creature %v, haste %v; want an earthbent land", creatureLand, hasty)
	}
	if power != n || toughness != n || countersOn(g, land, game.CounterPlusOne) != n {
		t.Errorf("earthbent land is %d/%d with %d counters, want %d", power, toughness, countersOn(g, land, game.CounterPlusOne), n)
	}
}

// --- Badgermole Cub ------------------------------------------------------------

func TestAvatarB14BadgermoleCubEarthbendsAndDoublesCreatureMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := avB14Forest(g, me.ID)
	plain := avB14Forest(g, me.ID)
	elf := b21Push(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", actLlanowarElvesOracle, 1, 1, "G")

	avB14Enter(t, g, me, game.Card{Name: "Badgermole Cub", TypeLine: "Creature — Badger Mole", OracleID: avB14CubOracle, Power: 2, Toughness: 2})
	pickTriggerTarget(t, g, me.ID, land)
	passPriorityAroundTable(t, g)
	avB14Earthbent(t, g, land, 1)

	g.WithWriteLock(func() { me.ManaPool = nil })
	tapForMana(t, g, me.ID, elf)
	if got := sortedPool(me); len(got) != 2 || got[0] != "G" || got[1] != "G" {
		t.Errorf("tapping a creature for mana: pool %v, want {G}{G}", got)
	}
	g.WithWriteLock(func() { me.ManaPool = nil })
	tapForMana(t, g, me.ID, plain)
	if got := sortedPool(me); len(got) != 1 {
		t.Errorf("tapping a land that isn't a creature: pool %v, want one {G}", got)
	}
	// The earthbent Forest is a creature now, so it counts.
	g.WithWriteLock(func() { me.ManaPool = nil })
	tapForMana(t, g, me.ID, land)
	if got := sortedPool(me); len(got) != 2 {
		t.Errorf("tapping the earthbent land: pool %v, want {G}{G}", got)
	}
}

func TestAvatarB14BadgermoleCubIgnoresAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b21Push(g, me.ID, "Badgermole Cub", "Creature — Badger Mole", avB14CubOracle, 2, 2, "G")
	theirElf := b21Push(g, opp.ID, "Llanowar Elves", "Creature — Elf Druid", actLlanowarElvesOracle, 1, 1, "G")
	tapForMana(t, g, opp.ID, theirElf)
	if got := sortedPool(opp); len(got) != 1 {
		t.Errorf("an opponent tapping their creature: pool %v, want one {G}", got)
	}
}

// --- The Boulder, Ready to Rumble ----------------------------------------------

func TestAvatarB14BoulderEarthbendsByItsBigCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := avB14Forest(g, me.ID)
	boulder := b21Push(g, me.ID, "The Boulder, Ready to Rumble", "Legendary Creature — Human Warrior Performer", avB14BoulderOracle, 4, 4, "G")
	b12Creature(g, me.ID, "Big", "Creature — Beast", 5, 5)
	b12Creature(g, me.ID, "Small", "Creature — Bear", 3, 3)
	b12Creature(g, opp.ID, "Their Big", "Creature — Beast", 6, 6)

	declareAttack(t, g, opp.ID, boulder)
	pickTriggerTarget(t, g, me.ID, land)
	passPriorityAroundTable(t, g)
	// The Boulder and the 5/5: two creatures you control with power 4+.
	if got := countersOn(g, land, game.CounterPlusOne); got != 2 {
		t.Errorf("earthbend X = %d, want 2", got)
	}
}

// --- Bumi, King of Three Trials ------------------------------------------------

func avB14Bumi(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	return avB14Enter(t, g, me, game.Card{Name: "Bumi, King of Three Trials", TypeLine: "Legendary Creature — Human Noble Ally",
		OracleID: avB14BumiOracle, Power: 4, Toughness: 4})
}

func TestAvatarB14BumiWithNoLessonsChoosesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	avB14Forest(g, me.ID)
	bumi := avB14Bumi(t, g, me)
	noModePick(t, g, me.ID, "no Lesson in the graveyard asks nothing")
	passPriorityAroundTable(t, g)
	if n := countersOn(g, bumi, game.CounterPlusOne); n != 0 {
		t.Errorf("%d counters with X = 0", n)
	}
}

func TestAvatarB14BumiChoosesUpToItsLessons(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := avB14Forest(g, me.ID)
	avB4Lesson(g, me)
	avB4Lesson(g, me)
	bumi := avB14Bumi(t, g, me)

	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("two Lessons: no mode prompt")
	}
	if c.ModeMin != 0 || c.ModeMax != 2 {
		t.Errorf("prompt bounds %d..%d, want 0..2", c.ModeMin, c.ModeMax)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0, 1, 2}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("three bullets with two Lessons: %v, want ErrInvalidParam", err)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0, 2}); err != nil {
		t.Fatalf("counters and earthbend: %v", err)
	}
	pickTriggerTarget(t, g, me.ID, land)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, bumi, game.CounterPlusOne); n != 3 {
		t.Errorf("Bumi has %d counters, want 3", n)
	}
	avB14Earthbent(t, g, land, 3)
}

func TestAvatarB14BumiCapsAtThreeBulletsAndScries(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := avB14Forest(g, me.ID)
	for i := 0; i < 5; i++ {
		avB4Lesson(g, me)
	}
	bumi := avB14Bumi(t, g, me)
	c := modePickChoiceFor(g, me.ID)
	if c == nil || c.ModeMax != 3 {
		t.Fatalf("five Lessons cap at the three bullets: %+v", c)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0, 1, 2}); err != nil {
		t.Fatalf("all three: %v", err)
	}
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the scry")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("pick the opponent: %v", err)
	}
	pickTriggerTarget(t, g, me.ID, land)
	passPriorityAroundTable(t, g)
	answerScryKeepAll(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, bumi, game.CounterPlusOne); n != 3 {
		t.Errorf("Bumi has %d counters, want 3", n)
	}
	avB14Earthbent(t, g, land, 3)
}

// --- Diligent Zookeeper --------------------------------------------------------

func TestAvatarB14ZookeeperPumpsByCreatureTypes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b21Push(g, me.ID, "Diligent Zookeeper", "Creature — Human Citizen Ally", avB14ZookeeperOracle, 4, 4, "G")
	beast := b12Creature(g, me.ID, "Two Types", "Creature — Ape Goat", 1, 1)
	human := b12Creature(g, me.ID, "A Human", "Creature — Human Soldier", 1, 1)
	none := b12Creature(g, me.ID, "No Types", "Creature", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Goat", "Creature — Goat", 1, 1)
	many := b12Creature(g, me.ID, "Many", "Creature — Ape Bat Bear Cat Dog Elk Fox Goat Hound Ox Rat Yeti", 0, 0)

	for _, c := range []struct {
		id   uuid.UUID
		want int
	}{{beast, 3}, {human, 1}, {none, 1}, {theirs, 1}, {many, 10}} {
		got := avB1Card(t, g, c.id)
		if got.CurrentPower() != c.want || got.CurrentToughness() != c.want {
			t.Errorf("%s is %d/%d, want %d/%d", got.Name, got.CurrentPower(), got.CurrentToughness(), c.want, c.want)
		}
	}
}

// --- Earth Kingdom General -----------------------------------------------------

func TestAvatarB14EarthKingdomGeneralGainsOnceEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := avB14Forest(g, me.ID)
	theirBear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	life := me.Life

	avB14Enter(t, g, me, game.Card{Name: "Earth Kingdom General", TypeLine: "Creature — Human Soldier Ally", OracleID: avB14GeneralOracle, Power: 2, Toughness: 2})
	pickTriggerTarget(t, g, me.ID, land)
	passPriorityAroundTable(t, g)
	avB14Earthbent(t, g, land, 2)
	// The earthbend's two counters are a placement you made.
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Fatalf("life %d → %d, want +2", life, me.Life)
	}

	// Once this turn: a counter you put on an opponent's creature asks nothing.
	tsPut(t, g, me.ID, theirBear, game.CounterPlusOne, 3)
	passPriorityAroundTable(t, g)
	if owedMays(g, me.ID) > 0 || me.Life != life+2 {
		t.Errorf("asked again or gained again this turn: life %d", me.Life)
	}
}

func TestAvatarB14EarthKingdomGeneralDecliningKeepsTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b21Push(g, me.ID, "Earth Kingdom General", "Creature — Human Soldier Ally", avB14GeneralOracle, 2, 2, "G")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	life := me.Life

	// An opponent putting counters is not you.
	tsPut(t, g, opp.ID, bear, game.CounterPlusOne, 1)
	passPriorityAroundTable(t, g)
	if owedMays(g, me.ID) > 0 {
		t.Fatal("an opponent's placement triggered it")
	}
	tsPut(t, g, me.ID, bear, game.CounterPlusOne, 1)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)

	advanceToStepOf(t, g, 0, game.StepBeginCombat)
	tsPut(t, g, me.ID, bear, game.CounterPlusOne, 2)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d → %d after a declined offer, want +2", life, me.Life)
	}
}

// --- Earth Rumble --------------------------------------------------------------

func TestAvatarB14EarthRumbleEarthbendsThenTheLandFights(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := avB14Forest(g, me.ID)
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	castCatalogSpell(t, g, "Earth Rumble", "Sorcery", avB14RumbleOracle, []game.TargetRef{{Kind: game.TargetCard, ID: land}})
	passPriorityAroundTable(t, g)
	avB14Earthbent(t, g, land, 2)

	b04WaitForPick(t, g, me.ID)
	pick := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: victim}); err == nil {
		t.Fatal("the first clause took an opponent's creature")
	}
	pickTriggerTarget(t, g, me.ID, land)
	pick = latestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
		t.Fatal("the second clause took your own creature")
	}
	pickTriggerTarget(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if onBattlefield(g, victim) {
		t.Error("the 2/2 survived a fight with the 2/2 land")
	}
}

func TestAvatarB14EarthRumbleRefusesAnOpponentsLand(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := avB14Forest(g, opp.ID)
	if err := castCatalogSpellErr(t, g, "Earth Rumble", "Sorcery", avB14RumbleOracle, []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}); err == nil {
		t.Fatal("earthbend targeted a land you don't control")
	}
}

// --- Earthbender Ascension -----------------------------------------------------

func TestAvatarB14EarthbenderAscensionEarthbendsThenFetchesABasic(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := avB14Forest(g, me.ID)
	basic := uuid.New()
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: basic, Name: "Mountain", TypeLine: "Basic Land — Mountain", Owner: me.ID, Controller: me.ID})
	})

	avB1Cast(t, g, me, game.Card{Name: "Earthbender Ascension", TypeLine: "Enchantment", OracleID: avB14AscensionOracle}, game.CastSpellParams{})
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	pickTriggerTarget(t, g, me.ID, land)
	passPriorityAroundTable(t, g)
	answerSearchNamed(t, g, me.ID, "Mountain")
	passPriorityAroundTable(t, g)
	avB14Earthbent(t, g, land, 2)
	c, ok := battlefieldCard(g, basic)
	if !ok {
		t.Fatal("the basic land was not put onto the battlefield")
	}
	if !c.Tapped {
		t.Error("the basic land entered untapped")
	}
}

func TestAvatarB14EarthbenderAscensionGrowsFromTheFourthQuestCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	asc := b21Push(g, me.ID, "Earthbender Ascension", "Enchantment", avB14AscensionOracle, 0, 0)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(asc, "quest", 2) })

	// Third counter: no reflexive trigger.
	b31PlayLand(t, g, "Forest", "Basic Land — Forest", "")
	passPriorityAroundTable(t, g)
	if n := countersOn(g, asc, "quest"); n != 3 {
		t.Fatalf("quest counters %d, want 3", n)
	}
	if latestPickTarget(g, me.ID) != nil || countersOn(g, bear, game.CounterPlusOne) != 0 {
		t.Fatal("the reflexive trigger came at three counters")
	}

	// Fourth: the counter and trample on target creature you control.
	pushEarthbendLand(g, me.ID, "Another Forest", "Basic Land — Forest")
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(asc, "quest", 0) })
	avB14LandEnters(t, g, me)
	b04WaitForPick(t, g, me.ID)
	pickTriggerTarget(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, asc, "quest"); n != 4 {
		t.Fatalf("quest counters %d, want 4", n)
	}
	if n := countersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("the creature has %d counters, want 1", n)
	}
	if !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the creature has no trample")
	}
}

// avB14LandEnters has a land enter under p's control by an effect.
func avB14LandEnters(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Library.PushTop(game.Card{InstanceID: id, Name: "Fetched Forest", TypeLine: "Basic Land — Forest", Owner: p.ID, Controller: p.ID})
	})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneLibrary, Owner: p.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, id); err != nil {
		t.Fatalf("land enters: %v", err)
	}
	g.RunStateChecksForTest()
}

// --- Flopsie, Bumi's Buddy -----------------------------------------------------

func TestAvatarB14FlopsieCountersEveryoneAndBigOnesTakeOneBlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := b12Creature(g, me.ID, "Big", "Creature — Beast", 3, 3)
	small := b12Creature(g, me.ID, "Small", "Creature — Bear", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	flopsie := avB14Enter(t, g, me, game.Card{Name: "Flopsie, Bumi's Buddy", TypeLine: "Legendary Creature — Ape Goat", OracleID: avB14FlopsieOracle, Power: 4, Toughness: 4})
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{big, small, flopsie} {
		if n := countersOn(g, id, game.CounterPlusOne); n != 1 {
			t.Errorf("%s has %d counters, want 1", id, n)
		}
	}
	if countersOn(g, theirs, game.CounterPlusOne) != 0 {
		t.Error("an opponent's creature got a counter")
	}

	w1 := b12Creature(g, opp.ID, "Wall One", "Creature — Wall", 0, 5)
	w2 := b12Creature(g, opp.ID, "Wall Two", "Creature — Wall", 0, 5)
	w3 := b12Creature(g, opp.ID, "Wall Three", "Creature — Wall", 0, 5)
	w4 := b12Creature(g, opp.ID, "Wall Four", "Creature — Wall", 0, 5)
	advanceToNextSeatsTurn(t, g)
	for g.Turn.ActiveSeat != 0 {
		advanceToNextSeatsTurn(t, g)
	}
	declareAttack(t, g, opp.ID, big, small)
	advanceTo(t, g, game.StepDeclareBlockers)
	// Big is 4/4 now: one blocker at most. Small (2/2) can take two.
	if err := g.DeclareBlockers([]game.BlockDeclaration{{Blocker: w1, Attacker: big}, {Blocker: w2, Attacker: big}}); err == nil {
		t.Fatal("two creatures blocked a power 4 creature")
	}
	if err := g.DeclareBlockers([]game.BlockDeclaration{{Blocker: w1, Attacker: big}, {Blocker: w3, Attacker: small}, {Blocker: w4, Attacker: small}}); err != nil {
		t.Fatalf("one on the big creature, two on the small: %v", err)
	}
}

// --- Foggy Swamp Vinebender ----------------------------------------------------

func TestAvatarB14VinebenderEvadesSmallBlockers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vine := b21Push(g, me.ID, "Foggy Swamp Vinebender", "Creature — Human Plant Ally", avB14VinebenderOracle, 4, 3, "G")
	small := b12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	big := b12Creature(g, opp.ID, "Big", "Creature — Bear", 3, 3)
	declareAttack(t, g, opp.ID, vine)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlockers([]game.BlockDeclaration{{Blocker: small, Attacker: vine}}); err == nil {
		t.Fatal("a power 2 creature blocked it")
	}
	if err := g.DeclareBlockers([]game.BlockDeclaration{{Blocker: big, Attacker: vine}}); err != nil {
		t.Fatalf("a power 3 creature can block it: %v", err)
	}
}

func TestAvatarB14VinebenderWaterbendsACounterOnlyOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	vine := b21Push(g, me.ID, "Foggy Swamp Vinebender", "Creature — Human Plant Ally", avB14VinebenderOracle, 4, 3, "G")
	var helpers []uuid.UUID
	for i := 0; i < 3; i++ {
		helpers = append(helpers, b12Creature(g, me.ID, "Helper", "Artifact Creature — Construct", 1, 1))
	}
	// Three taps can't pay {5} alone.
	if err := avB4Activate(g, me.ID, vine, 0, helpers); err == nil {
		t.Fatal("waterbend {5} was paid with three taps and no mana")
	}
	g.WithWriteLock(func() { me.ManaPool = nil })
	avB14AddMana(t, g, me, "{G}{G}")
	if err := avB4Activate(g, me.ID, vine, 0, helpers); err != nil {
		t.Fatalf("waterbend {5} with {G}{G} and three taps: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := countersOn(g, vine, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters, want 1", n)
	}

	advanceToNextSeatsTurn(t, g)
	advanceToMain(t, g)
	avB14AddMana(t, g, me, "{G}{G}{G}{G}{G}")
	if err := avB4Activate(g, me.ID, vine, 0, nil); err == nil {
		t.Error("activated on an opponent's turn")
	}
}

// avB14AddMana adds mana to p's pool.
func avB14AddMana(t *testing.T, g *game.Game, p *game.Player, mana string) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AddManaForEffect(p.ID, uuid.Nil, mana); err != nil {
			t.Fatalf("AddManaForEffect: %v", err)
		}
	})
}

// --- Haru, Hidden Talent -------------------------------------------------------

func TestAvatarB14HaruEarthbendsWhenAnotherAllyEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := avB14Forest(g, me.ID)
	b21Push(g, me.ID, "Haru, Hidden Talent", "Legendary Creature — Human Peasant Ally", avB14HaruOracle, 1, 1, "G")

	avB14Enter(t, g, me, game.Card{Name: "Not An Ally", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("a non-Ally entering triggered Haru")
	}
	avB14Enter(t, g, me, game.Card{Name: "An Ally", TypeLine: "Creature — Human Ally", Power: 2, Toughness: 2})
	pickTriggerTarget(t, g, me.ID, land)
	passPriorityAroundTable(t, g)
	avB14Earthbent(t, g, land, 1)
}
