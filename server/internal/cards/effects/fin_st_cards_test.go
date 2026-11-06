package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fin_st_cards_test.go — Final Fantasy (FIN), cards S–T (#4462). Each
// test plays the real catalog entry and asserts what its text says.

const (
	finSTKatanaOracle       = "9b8b5812-73bd-453e-bdff-aad1fafa9226"
	finSTKnifeOracle        = "4bf589ed-58f8-40ad-bcee-50cb99da02ca"
	finSTSharlayanOracle    = "564bdbdd-8392-4ee1-a132-1a17a67b2110"
	finSTTrenoOracle        = "6d5bb714-1bf4-4f75-9ea8-bdbd18b0425b"
	finSTInterventionOracle = "94ac2990-58a5-43d4-bdac-ec2f6daf0b51"
	finSTShinraOracle       = "6cfe63ac-3581-461b-8ed3-3e231851e3f8"
	finSTChosenOracle       = "df4ef2f8-2591-4be3-b905-4f9ff9c538ea"
	finSTScorpionOracle     = "5e53fe9f-9995-47ff-85e9-f09b70d7af15"
	finSTSnowOracle         = "01cdaf69-f53a-4735-92f2-ef149de4d5c9"
	finSTSeymourOracle      = "7312efc2-8a02-47cd-874b-c3338241435e"
	finSTTifaOracle         = "878df304-9d6d-4ce9-b65b-1439c251f3e0"
	finSTTonberryOracle     = "a26f3cbb-39b5-4688-92f7-26534613a9f3"
)

func TestFinSTEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		finSTKatanaOracle:       "Samurai's Katana",
		finSTKnifeOracle:        "Thief's Knife",
		finSTSharlayanOracle:    "Sharlayan, Nation of Scholars",
		finSTTrenoOracle:        "Treno, Dark City",
		finSTInterventionOracle: "Sephiroth's Intervention",
		finSTShinraOracle:       "Shinra Reinforcements",
		finSTChosenOracle:       "The Crystal's Chosen",
		finSTScorpionOracle:     "Scorpion Sentinel",
		finSTSnowOracle:         "Snow Villiers",
		finSTSeymourOracle:      "Seymour Flux",
		finSTTifaOracle:         "Tifa Lockhart",
		finSTTonberryOracle:     "Tonberry",
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

// finSTCounters reads a battlefield permanent's counters of one kind.
func finSTCounters(t *testing.T, g *game.Game, id uuid.UUID, kind string) int {
	t.Helper()
	c := findBattlefieldCardByID(g, id)
	if c == nil {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Counters[kind]
}

// --- Samurai's Katana ---------------------------------------------------------------

func TestFinSTSamuraisKatanaMakesAHastyTramplingSamuraiHero(t *testing.T) {
	g := newCatalogGame(t)
	katana := castCatalogSpell(t, g, "Samurai's Katana", "Artifact — Equipment", finSTKatanaOracle, nil)
	passPriorityAroundTable(t, g)
	hero := ysBHero(t, g, katana)
	v := s58p6View(t, g, hero)
	if v.CurrentPower() != 3 || v.CurrentToughness() != 3 || !v.HasSubtype("Samurai") || !v.HasSubtype("Hero") {
		t.Errorf("Hero is %d/%d %v, want a 3/3 Hero Samurai", v.CurrentPower(), v.CurrentToughness(), v.Effective().Subtypes)
	}
	for _, kw := range []string{"trample", "haste"} {
		if !hasEffectiveKeyword(t, g, hero, kw) {
			t.Errorf("the Hero has no %s", kw)
		}
	}
	spec, _ := Lookup(finSTKatanaOracle)
	if len(spec.Activated) != 1 || spec.Activated[0].Label != EquipAbility("{5}").Label {
		t.Errorf("activated rows %+v, want one equip {5}", spec.Activated)
	}
}

// --- Thief's Knife ------------------------------------------------------------------

func TestFinSTThiefsKnifeHeroDrawsOnCombatDamageToAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	knife := castCatalogSpell(t, g, "Thief's Knife", "Artifact — Equipment", finSTKnifeOracle, nil)
	passPriorityAroundTable(t, g)
	hero := ysBHero(t, g, knife)
	v := s58p6View(t, g, hero)
	if v.CurrentPower() != 2 || v.CurrentToughness() != 2 || !v.HasSubtype("Rogue") {
		t.Errorf("Hero is %d/%d %v, want a 2/2 Hero Rogue", v.CurrentPower(), v.CurrentToughness(), v.Effective().Subtypes)
	}
	hand := me.Hand.Size()
	dealCombatDamageToPlayer(g, hero, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, me.Hand.Size())
	}

	// Another creature's damage draws nothing.
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand = me.Hand.Size()
	dealCombatDamageToPlayer(g, bear, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("an unequipped creature's damage drew: hand %d → %d", hand, me.Hand.Size())
	}
}

// --- Sharlayan and Treno ------------------------------------------------------------

func TestFinSTTownDualsEnterTappedAndTapForTheirTwoColours(t *testing.T) {
	for _, tc := range []struct{ name, oracle, produced string }{
		{"Sharlayan, Nation of Scholars", finSTSharlayanOracle, "{W|U}"},
		{"Treno, Dark City", finSTTrenoOracle, "{U|B}"},
	} {
		g := newCatalogGame(t)
		land := playLandFromHand(t, g, tc.name, tc.oracle)
		if c := findBattlefieldCardByID(g, land); c == nil || !c.Tapped {
			t.Errorf("%s should enter tapped", tc.name)
		}
		spec, _ := Lookup(tc.oracle)
		if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != tc.produced || !spec.ManaAbilities[0].Cost.Tap {
			t.Errorf("%s mana abilities %+v, want {T}: %s", tc.name, spec.ManaAbilities, tc.produced)
		}
	}
}

// --- Sephiroth's Intervention -------------------------------------------------------

func TestFinSTSephirothsInterventionDestroysAndGainsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	life := me.Life
	castCatalogSpell(t, g, "Sephiroth's Intervention", "Instant", finSTInterventionOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the creature was not destroyed")
	}
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2", life, me.Life)
	}
}

// --- Shinra Reinforcements ----------------------------------------------------------

func TestFinSTShinraReinforcementsMillsThreeAndGainsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life, library, graveyard := me.Life, me.Library.Size(), me.Graveyard.Size()
	castCatalogSpell(t, g, "Shinra Reinforcements", "Creature — Human Soldier", finSTShinraOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want +3", life, me.Life)
	}
	if me.Library.Size() != library-3 || me.Graveyard.Size() != graveyard+3 {
		t.Errorf("library %d → %d, graveyard %d → %d, want three milled",
			library, me.Library.Size(), graveyard, me.Graveyard.Size())
	}
}

// --- The Crystal's Chosen -----------------------------------------------------------

func TestFinSTCrystalsChosenMakesFourHeroesThenACounterOnEachOfYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 2, 2)
	castCatalogSpell(t, g, "The Crystal's Chosen", "Sorcery", finSTChosenOracle, nil)
	passPriorityAroundTable(t, g)
	heroes := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Hero" || c.Controller != me.ID {
			continue
		}
		heroes++
		if v := s58p6View(t, g, c.InstanceID); v.CurrentPower() != 2 || v.CurrentToughness() != 2 {
			t.Errorf("a Hero is %d/%d, want 2/2 with its counter", v.CurrentPower(), v.CurrentToughness())
		}
	}
	if heroes != 4 {
		t.Errorf("%d Heroes, want 4", heroes)
	}
	if n := finSTCounters(t, g, mine, game.CounterPlusOne); n != 1 {
		t.Errorf("your Bear has %d +1/+1 counters, want 1", n)
	}
	if n := finSTCounters(t, g, theirs, game.CounterPlusOne); n != 0 {
		t.Errorf("an opponent's creature has %d +1/+1 counters, want 0", n)
	}
}

// --- Scorpion Sentinel --------------------------------------------------------------

func TestFinSTScorpionSentinelGetsThreeWithSevenLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	scorpion := b12Push(g, me.ID, "Scorpion Sentinel", "Artifact Creature — Robot Scorpion", finSTScorpionOracle, 1, 4)
	for i := 0; i < 6; i++ {
		b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	}
	if v := s58p6View(t, g, scorpion); v.CurrentPower() != 1 || v.CurrentToughness() != 4 {
		t.Fatalf("with six lands it is %d/%d, want 1/4", v.CurrentPower(), v.CurrentToughness())
	}
	b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	if v := s58p6View(t, g, scorpion); v.CurrentPower() != 4 || v.CurrentToughness() != 4 {
		t.Errorf("with seven lands it is %d/%d, want 4/4", v.CurrentPower(), v.CurrentToughness())
	}
}

// --- Snow Villiers ------------------------------------------------------------------

func TestFinSTSnowVilliersPowerIsYourCreatureCount(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	snow := b12Push(g, me.ID, "Snow Villiers", "Legendary Creature — Human Rebel Monk", finSTSnowOracle, 0, 3)
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Wolf", "Creature — Wolf", 2, 2)
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if v := s58p6View(t, g, snow); v.CurrentPower() != 3 || v.CurrentToughness() != 3 {
		t.Errorf("Snow is %d/%d with three creatures of yours, want 3/3", v.CurrentPower(), v.CurrentToughness())
	}
	if !hasEffectiveKeyword(t, g, snow, "vigilance") {
		t.Error("Snow Villiers has no vigilance")
	}
}

// --- Seymour Flux -------------------------------------------------------------------

func TestFinSTSeymourFluxPaysALifeToDrawAndGrow(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		seymour := b12Push(g, me.ID, "Seymour Flux", "Legendary Creature — Spirit Avatar", finSTSeymourOracle, 5, 5)
		b12ToMyNextUpkeep(t, g)
		life, hand := me.Life, me.Hand.Size()
		passPriorityAroundTable(t, g)
		if !hasPayUnlessFor(g, me.ID) {
			t.Fatalf("pay %v: the upkeep trigger did not offer the life payment", pay)
		}
		answerPayUnless(t, g, me.ID, pay)
		wantLife, wantHand, wantCounters := life-1, hand+1, 1
		if !pay {
			wantLife, wantHand, wantCounters = life, hand, 0
		}
		if me.Life != wantLife || me.Hand.Size() != wantHand {
			t.Errorf("pay %v: life %d → %d, hand %d → %d; want life %d, hand %d",
				pay, life, me.Life, hand, me.Hand.Size(), wantLife, wantHand)
		}
		if n := finSTCounters(t, g, seymour, game.CounterPlusOne); n != wantCounters {
			t.Errorf("pay %v: Seymour has %d +1/+1 counters, want %d", pay, n, wantCounters)
		}
	}
}

// --- Tifa Lockhart ------------------------------------------------------------------

func TestFinSTTifaDoublesHerPowerOnEachLandfall(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tifa := b12Push(g, me.ID, "Tifa Lockhart", "Legendary Creature — Human Monk", finSTTifaOracle, 1, 2)
	if !hasEffectiveKeyword(t, g, tifa, "trample") {
		t.Error("Tifa has no trample")
	}
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	if v := s58p6View(t, g, tifa); v.CurrentPower() != 2 || v.CurrentToughness() != 2 {
		t.Fatalf("after one land Tifa is %d/%d, want 2/2", v.CurrentPower(), v.CurrentToughness())
	}
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	if v := s58p6View(t, g, tifa); v.CurrentPower() != 4 || v.CurrentToughness() != 2 {
		t.Errorf("after two lands Tifa is %d/%d, want 4/2", v.CurrentPower(), v.CurrentToughness())
	}
}

// --- Tonberry -----------------------------------------------------------------------

func TestFinSTTonberryEntersStunnedAndStrikesOnlyOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tonberry := castRingCreature(t, g, "Tonberry", "Creature — Salamander Horror", finSTTonberryOracle, 2, 1)
	passPriorityAroundTable(t, g)
	c := findBattlefieldCardByID(g, tonberry)
	if c == nil || !c.Tapped || c.Counters[game.CounterStun] != 1 {
		t.Fatalf("Tonberry entered %+v, want tapped with one stun counter", c)
	}
	if g.Turn.ActiveSeat != 0 || c.Controller != me.ID {
		t.Fatal("expected seat 0's own turn")
	}
	for _, kw := range []string{"first strike", "deathtouch"} {
		if !hasEffectiveKeyword(t, g, tonberry, kw) {
			t.Errorf("on your turn Tonberry has no %s", kw)
		}
	}
	for g.Turn.ActiveSeat == 0 {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	for _, kw := range []string{"first strike", "deathtouch"} {
		if hasEffectiveKeyword(t, g, tonberry, kw) {
			t.Errorf("on an opponent's turn Tonberry has %s", kw)
		}
	}
}
