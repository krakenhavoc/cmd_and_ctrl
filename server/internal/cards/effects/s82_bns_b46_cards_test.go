package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bns_b46_cards_test.go — Foundations Jumpstart, S82 BNS batch 46 (#5929).
// Each test plays the real catalog entries through the engine.

const (
	bnsB46Venom     = "cbc808dd-8d50-4f42-b0f1-d4e45d0462a2"
	bnsB46Wisps     = "e699d140-eb13-4f15-aecb-85c76e8b1b14"
	bnsB46Bargain   = "2f872417-3dbe-40ec-ac91-e3132c0b7514"
	bnsB46Onslaught = "ce29f71a-6662-4be4-9915-5a4be87288e0"
	bnsB46Unity     = "dc665ddc-4326-4860-94ba-c40ced8b0af6"
	bnsB46Campus    = "274c2781-f40a-49b4-a039-669739beaab6"
	bnsB46Strike    = "f8c2d9c5-6fe6-4698-9972-83cdd625d94b"
	bnsB46Discovery = "705cc928-d6d4-4294-979b-a385d1b07dc6"
	bnsB46Lunge     = "87644a09-241a-44e5-99f2-16fc9a7d8b93"
	bnsB46Arrow     = "6add505a-9c0f-42e6-898d-be3194c3df33"
)

func TestBnsB46EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		bnsB46Venom: "Coat with Venom", bnsB46Wisps: "Crimson Wisps", bnsB46Bargain: "Dark Bargain",
		bnsB46Onslaught: "Dauntless Onslaught", bnsB46Unity: "Dauntless Unity",
		bnsB46Campus: "Defend the Campus", bnsB46Strike: "Defiant Strike",
		bnsB46Discovery: "Delightful Discovery", bnsB46Lunge: "Desperate Lunge", bnsB46Arrow: "Divine Arrow",
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

func TestBnsB46SingleTargetPumps(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		power, tough int
		keyword      string
		life, draw   int
	}{
		{"Coat with Venom", bnsB46Venom, 1, 2, "deathtouch", 0, 0},
		{"Defiant Strike", bnsB46Strike, 1, 0, "", 0, 1},
		{"Desperate Lunge", bnsB46Lunge, 2, 2, "flying", 2, 0},
		{"Crimson Wisps", bnsB46Wisps, 0, 0, "haste", 0, 1},
	} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		bear := b12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
		life, hand := me.Life, me.Hand.Size()
		castCatalogSpell(t, g, tc.name, "Instant", tc.oracle, cardTarget(bear))
		passPriorityAroundTable(t, g)
		if p, th := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 2+tc.power || th != 2+tc.tough {
			t.Errorf("%s: %d/%d", tc.name, p, th)
		}
		if tc.keyword != "" && !hasKeywordOnBattlefield(t, g, bear, tc.keyword) {
			t.Errorf("%s: no %s", tc.name, tc.keyword)
		}
		if me.Life != life+tc.life {
			t.Errorf("%s: life %d, want %d", tc.name, me.Life, life+tc.life)
		}
		if me.Hand.Size() != hand+tc.draw {
			t.Errorf("%s: hand %d, want %d", tc.name, me.Hand.Size(), hand+tc.draw)
		}
	}
}

func TestBnsB46SingleTargetSpellsFizzleOnAGoneTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Defiant Strike", "Instant", bnsB46Strike, cardTarget(bear))
	g.WithWriteLock(func() { g.Battlefield.Remove(bear) })
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("a fizzled Defiant Strike still drew: hand %d → %d", hand, me.Hand.Size())
	}
}

func TestBnsB46CrimsonWispsReplacesColours(t *testing.T) {
	g := newCatalogGame(t)
	elf := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Elf", TypeLine: "Creature — Elf", Colors: []string{"G"},
		Power: 1, Toughness: 1, Owner: g.Seats[1].ID, Controller: g.Seats[1].ID,
	})
	castCatalogSpell(t, g, "Crimson Wisps", "Instant", bnsB46Wisps, cardTarget(elf))
	passPriorityAroundTable(t, g)
	var colors []string
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == elf {
				colors = c.Effective().Colors
			}
		}
	})
	if len(colors) != 1 || colors[0] != "R" {
		t.Errorf("\"becomes red\" replaces the colours: %v", colors)
	}
}

func TestBnsB46DauntlessOnslaughtBoostsUpToTwo(t *testing.T) {
	g := newCatalogGame(t)
	a := b12Creature(g, g.Seats[0].ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, g.Seats[1].ID, "B", "Creature — Bear", 3, 3)
	c := b12Creature(g, g.Seats[1].ID, "C", "Creature — Bear", 3, 3)
	castCatalogSpell(t, g, "Dauntless Onslaught", "Instant", bnsB46Onslaught, cardTargets(a, b))
	passPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{a: 4, b: 5, c: 3} {
		if got := effectivePower(t, g, id); got != want {
			t.Errorf("power %d, want %d", got, want)
		}
	}
	if got := effectiveToughness(t, g, a); got != 4 {
		t.Errorf("toughness %d, want 4", got)
	}
	if err := castCatalogSpellErr(t, g, "Dauntless Onslaught", "Instant", bnsB46Onslaught, cardTargets(a, b, c)); err == nil {
		t.Error("three targets were accepted")
	}
}

func bnsB46Cast(t *testing.T, g *game.Game, me *game.Player, c game.Card, params game.CastSpellParams) error {
	t.Helper()
	advanceToMain(t, g)
	id := avB1Hand(me, c)
	params.Strict, params.AutoTap = true, true
	return g.CastSpell(me.ID, id, params)
}

func TestBnsB46DauntlessUnityUnkickedAndKicked(t *testing.T) {
	card := game.Card{Name: "Dauntless Unity", TypeLine: "Instant", ManaCost: "{1}{W}", OracleID: bnsB46Unity}
	for _, kicked := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
		theirs := b12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
		avB1Plains(g, me.ID, 2)
		params := game.CastSpellParams{}
		wantP := 3
		if kicked {
			avB1Plains(g, me.ID, 2)
			params.OptionalCosts = []int{0}
			wantP = 4
		}
		if err := bnsB46Cast(t, g, me, card, params); err != nil {
			t.Fatalf("kicked=%v: %v", kicked, err)
		}
		passPriorityAroundTable(t, g)
		if p, th := effectivePower(t, g, mine), effectiveToughness(t, g, mine); p != wantP || th != 3 {
			t.Errorf("kicked=%v: %d/%d, want %d/3", kicked, p, th, wantP)
		}
		if p := effectivePower(t, g, theirs); p != 2 {
			t.Errorf("kicked=%v: an opponent's creature got a boost: power %d", kicked, p)
		}
	}
	// Two mana cannot pay the kicker.
	g := newCatalogGame(t)
	me := g.Seats[0]
	avB1Plains(g, me.ID, 2)
	if err := bnsB46Cast(t, g, me, card, game.CastSpellParams{OptionalCosts: []int{0}}); err == nil {
		t.Error("a kicked cast was paid with two mana")
	}
}

func TestBnsB46DefendTheCampusModes(t *testing.T) {
	g := newCatalogGame(t)
	mine := b12Creature(g, g.Seats[0].ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
	castModal(t, g, "Defend the Campus", "Instant", bnsB46Campus, []int{0}, nil)
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, mine); p != 3 {
		t.Errorf("mine power %d, want 3", p)
	}
	if p := effectivePower(t, g, theirs); p != 2 {
		t.Errorf("theirs power %d, want 2", p)
	}
}

func TestBnsB46DefendTheCampusDestroysOnlyPowerFourOrGreater(t *testing.T) {
	g := newCatalogGame(t)
	small := b12Creature(g, g.Seats[1].ID, "Small", "Creature — Bear", 3, 3)
	big := b12Creature(g, g.Seats[1].ID, "Big", "Creature — Wurm", 4, 4)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	id := avB1Hand(me, game.Card{Name: "Defend the Campus", TypeLine: "Instant", OracleID: bnsB46Campus})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{1}, Targets: cardTarget(small)}); err == nil {
		t.Fatal("a 3-power creature was a legal target")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{1}, Targets: cardTarget(big)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) || !g.Battlefield.Contains(small) {
		t.Error("only the 4/4 should have been destroyed")
	}
}

func TestBnsB46DivineArrowHitsOnlyAttackersAndBlockers(t *testing.T) {
	g := newCatalogGame(t)
	me, them := seatsForCombat(g)
	attacker := b12Creature(g, me.ID, "Ogre", "Creature — Ogre", 5, 5)
	idle := b12Creature(g, them.ID, "Bears", "Creature — Bear", 2, 2)
	attackIntoBlocks(t, g, them.ID, attacker)
	hand := avB1Hand(me, game.Card{Name: "Divine Arrow", TypeLine: "Instant", OracleID: bnsB46Arrow})
	if err := g.CastSpell(me.ID, hand, game.CastSpellParams{Strict: true, Targets: cardTarget(idle)}); err == nil {
		t.Fatal("a creature that is neither attacking nor blocking was a legal target")
	}
	if err := g.CastSpell(me.ID, hand, game.CastSpellParams{Strict: true, Targets: cardTarget(attacker)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := damageOn(t, g, attacker); n != 4 {
		t.Errorf("the attacker has %d damage, want 4", n)
	}
}

func TestBnsB46DarkBargainTakesTwoBinsOneAndHurts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedLibrary(me, "One", "Two", "Three", "Four")
	life := me.Life
	castCatalogSpell(t, g, "Dark Bargain", "Instant", bnsB46Bargain, nil)
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick among the top three cards")
	}
	if len(pick.ChooseCards) != 3 || pick.ChooseMin != 2 || pick.ChooseMax != 2 {
		t.Fatalf("offered %d cards in [%d,%d], want exactly two of the top three", len(pick.ChooseCards), pick.ChooseMin, pick.ChooseMax)
	}
	if slices.Contains(pick.ChooseCards, ids[3]) {
		t.Error("the fourth card was offered")
	}
	if me.Life != life {
		t.Errorf("life %d before the pick; the damage comes after the cards move", me.Life)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[0], ids[2]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(ids[0]) || !me.Hand.Contains(ids[2]) {
		t.Error("the chosen cards are not in hand")
	}
	if !me.Graveyard.Contains(ids[1]) {
		t.Error("the card not chosen is not in the graveyard")
	}
	if !me.Library.Contains(ids[3]) {
		t.Error("the fourth card left the library")
	}
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
}

func TestBnsB46DelightfulDiscoveryCostsLessPerOpponentSpell(t *testing.T) {
	card := game.Card{Name: "Delightful Discovery", TypeLine: "Instant", ManaCost: "{4}{U}", OracleID: bnsB46Discovery}
	setup := func(lands int, oppSpells, ownSpells int) (*game.Game, *game.Player) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		for i := 0; i < lands; i++ {
			pushLandFor(g, me.ID, "Island", "Basic Land — Island")
		}
		g.WithWriteLock(func() {
			if g.SpellsCastThisTurn == nil {
				g.SpellsCastThisTurn = map[uuid.UUID]game.CastTally{}
			}
			g.SpellsCastThisTurn[g.Seats[1].ID] = game.CastTally{Total: oppSpells}
			g.SpellsCastThisTurn[me.ID] = game.CastTally{Total: ownSpells}
		})
		return g, me
	}
	// Four Islands cannot pay {4}{U} with no discount, nor with only own spells cast.
	g, me := setup(4, 0, 3)
	if err := bnsB46Cast(t, g, me, card, game.CastSpellParams{}); err == nil {
		t.Error("cast for four mana with no opponent spell this turn")
	}
	// One opponent spell makes it {3}{U}.
	g, me = setup(4, 1, 0)
	if err := bnsB46Cast(t, g, me, card, game.CastSpellParams{}); err != nil {
		t.Errorf("one opponent spell: %v", err)
	}
	// The discount never touches the {U}: four opponent spells leave {U}.
	g, me = setup(1, 4, 0)
	library := seedLibrary(me, "A", "B", "C", "D")
	hand := me.Hand.Size()
	if err := bnsB46Cast(t, g, me, card, game.CastSpellParams{}); err != nil {
		t.Fatalf("four opponent spells: %v", err)
	}
	passPriorityAroundTable(t, g)
	scry := latestChoiceOfKind(g, game.PendingChoiceScry)
	if scry == nil {
		t.Fatal("no scry 2 was asked")
	}
	if err := g.ResolveScry(scry.ID, me.ID, nil, scry.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(library[0]) || !me.Hand.Contains(library[1]) {
		t.Error("the top two cards should have been drawn")
	}
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+2)
	}
}
