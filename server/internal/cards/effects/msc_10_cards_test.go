package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_10_cards_test.go — Marvel Super Heroes Commander, tenth batch
// (#4455). Each test plays the real catalog entry and asserts what its
// text says.

const (
	msc10CaptainMarvelOracle = "73fdb375-9c71-421d-8027-8636e3ea8f35"
	msc10HawkeyesShotOracle  = "6bbd1316-02d3-46ab-ba1a-0ecab51ec4c4"
	msc10RossOracle          = "57ffb558-3a1c-43ea-a79c-bcfacad22ca4"
	msc10FoggyOracle         = "734fc893-b736-448a-b851-826976031754"
	msc10FlashOracle         = "f536a518-160e-4267-89da-3d34accc6d2c"
	msc10FloraOracle         = "04dce348-be42-443a-a88d-c47436ba2ec4"
	msc10RebirthOracle       = "05a1e1ee-67f0-4158-a012-01612bc9cc70"
	msc10GymOracle           = "850bb6f7-48d3-4d65-9220-b0bec5ee6b64"
	msc10BartonOracle        = "c9efec35-823f-4c78-be36-d1a57d8b6f44"
	msc10HitMonkeyOracle     = "e27d1993-8d83-449d-8a64-8b66086db995"
	msc10BannerOracle        = "1f14d336-71ef-4e3f-b8e0-42585af8db40"
	msc10BrawlerOracle       = "6a953b71-2994-4c08-9602-b2bbce5b5b4f"
	msc10LukeCageOracle      = "3d776a25-c341-4058-9b5f-84e70edb25a4"
	msc10JarvisOracle        = "63e17cbd-6dae-42d2-a074-7924db1477bf"
	msc10RanksOracle         = "2fb76cdf-438b-48f5-aceb-db637381d052"
	msc10DisintegratorOracle = "cf16702d-9a0e-4e4c-82cb-d9b0b1e23f15"
	msc10SwordsOracle        = "b1544f21-7e98-461b-aed5-e748b0168c52"
)

func TestMsc10EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		msc10CaptainMarvelOracle: "Captain Marvel, Shooting Star",
		msc10HawkeyesShotOracle:  "Hawkeye's Shot",
		msc10RossOracle:          "General Thunderbolt Ross",
		msc10FoggyOracle:         "Foggy Nelson, On Retainer",
		msc10FlashOracle:         "Flash Thompson, Spider-Fan",
		msc10FloraOracle:         "Flora Colossus",
		msc10RebirthOracle:       "Infernal Rebirth",
		msc10GymOracle:           "Fogwell's Gym",
		msc10BartonOracle:        "Hawkeye, Clint Barton",
		msc10HitMonkeyOracle:     "Hit-Monkey",
		msc10BannerOracle:        "Hulk, Bruce Banner",
		msc10BrawlerOracle:       "Hulk, Brutal Brawler",
		msc10LukeCageOracle:      "Luke Cage, Hero for Hire",
		msc10JarvisOracle:        "Jarvis, Earth's Mightiest Butler",
		msc10RanksOracle:         "Endless Ranks of HYDRA",
		msc10DisintegratorOracle: "HYDRA Disintegrator",
	}
	for oracle, name := range want {
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

// msc10Cast puts the named spell in `who`'s hand, floats `mana` and casts
// it under the strict mana gate, so a cast that costs more than the pool
// holds is refused.
func msc10Cast(t *testing.T, g *game.Game, who *game.Player, name, typeLine, cost, oracle, mana string, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	id := handCardFull(who, name, typeLine, cost, oracle, nil)
	if mana != "" {
		if err := g.AddManaForEffect(who.ID, uuid.Nil, mana); err != nil {
			t.Fatalf("AddManaForEffect(%s): %v", mana, err)
		}
	}
	params.Strict = true
	return id, g.CastSpell(who.ID, id, params)
}

func msc10InExile(g *game.Game, id uuid.UUID) bool {
	z := g.FindCardZoneForEffect(id)
	return z != nil && z.Kind == game.ZoneExile
}

// --- Captain Marvel, Shooting Star ---

func TestMsc10CaptainMarvelEntersAndExilesATargetThenBothPlayersGain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)
	myLife, oppLife := me.Life, opp.Life
	castHoldCreature(t, g, "Captain Marvel, Shooting Star", "Legendary Creature — Human Kree Hero", msc10CaptainMarvelOracle, 6, 6)
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	if !msc10InExile(g, theirs) {
		t.Fatal("the targeted creature is not exiled")
	}
	if opp.Life != oppLife+4 {
		t.Errorf("the exiled creature's controller has %d life, want %d", opp.Life, oppLife+4)
	}
	if me.Life != myLife+4 {
		t.Errorf("Captain Marvel's controller has %d life, want %d (its power from her second ability)", me.Life, myLife+4)
	}
}

func TestMsc10CaptainMarvelExilesOnAttackToo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	marvel := b12Push(g, me.ID, "Captain Marvel, Shooting Star", "Legendary Creature — Human Kree Hero", msc10CaptainMarvelOracle, 6, 6)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	oppLife := opp.Life
	declareAttack(t, g, opp.ID, marvel)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if !msc10InExile(g, theirs) {
		t.Fatal("the attack trigger did not exile the target")
	}
	if opp.Life != oppLife+2 {
		t.Errorf("the exiled creature's controller has %d life, want %d", opp.Life, oppLife+2)
	}
}

func TestMsc10CaptainMarvelGainsWhenAnyOtherCreatureIsExiled(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b12Push(g, me.ID, "Captain Marvel, Shooting Star", "Legendary Creature — Human Kree Hero", msc10CaptainMarvelOracle, 6, 6)
	theirs := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
	myLife, oppLife := me.Life, opp.Life
	castCatalogSpell(t, g, "Swords to Plowshares", "Instant", msc10SwordsOracle, b16TargetCard(theirs))
	passPriorityAroundTable(t, g)
	if !msc10InExile(g, theirs) {
		t.Fatal("Swords did not exile the creature")
	}
	if me.Life != myLife+5 {
		t.Errorf("Captain Marvel's controller has %d life, want %d", me.Life, myLife+5)
	}
	if opp.Life != oppLife+5 {
		t.Errorf("Swords' payout to the creature's controller: %d life, want %d", opp.Life, oppLife+5)
	}
}

func TestMsc10CaptainMarvelDoesNotGainWhenACreatureDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b12Push(g, me.ID, "Captain Marvel, Shooting Star", "Legendary Creature — Human Kree Hero", msc10CaptainMarvelOracle, 6, 6)
	theirs := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
	myLife := me.Life
	b27Kill(g, theirs)
	passPriorityAroundTable(t, g)
	if me.Life != myLife {
		t.Errorf("a creature that died paid out: %d life, want %d", me.Life, myLife)
	}
}

func TestMsc10CaptainMarvelDoesNotTriggerOnHerOwnExile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	marvel := b12Push(g, me.ID, "Captain Marvel, Shooting Star", "Legendary Creature — Human Kree Hero", msc10CaptainMarvelOracle, 6, 6)
	myLife := me.Life
	castCatalogSpell(t, g, "Swords to Plowshares", "Instant", msc10SwordsOracle, b16TargetCard(marvel))
	passPriorityAroundTable(t, g)
	if me.Life != myLife+6 {
		t.Errorf("life %d, want %d: only Swords' payout, not her own second ability", me.Life, myLife+6)
	}
}

// --- Hawkeye's Shot ---

func TestMsc10HawkeyesShotCostsTwoLessAtATappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tapped := pushPermanent(g, opp.ID, game.Card{Name: "Tired", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Tapped: true})
	fresh := pushPermanent(g, opp.ID, game.Card{Name: "Fresh", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	price := func(refs []game.TargetRef) int {
		return selfPriced(t, g, me, msc10HawkeyesShotOracle, "Sorcery", "{4}{W}", game.ZoneHand, refs).ManaValue()
	}
	if got := price(b16TargetCard(tapped)); got != 3 {
		t.Errorf("at a tapped creature: %d, want 3", got)
	}
	if got := price(b16TargetCard(fresh)); got != 5 {
		t.Errorf("at an untapped creature: %d, want 5", got)
	}
}

func TestMsc10HawkeyesShotExilesATappedCreatureForThreeMana(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	big := pushPermanent(g, opp.ID, game.Card{Name: "Tired Giant", TypeLine: "Creature — Giant", Power: 5, Toughness: 5, Tapped: true})
	if _, err := msc10Cast(t, g, me, "Hawkeye's Shot", "Sorcery", "{4}{W}", msc10HawkeyesShotOracle, "{C}{C}{W}", game.CastSpellParams{Targets: b16TargetCard(big)}); err != nil {
		t.Fatalf("cast for {2}{W}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !msc10InExile(g, big) {
		t.Error("the tapped creature was not exiled")
	}
}

func TestMsc10HawkeyesShotAtAnUntappedCreatureCostsFive(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	big := pushPermanent(g, opp.ID, game.Card{Name: "Fresh Giant", TypeLine: "Creature — Giant", Power: 5, Toughness: 5})
	if _, err := msc10Cast(t, g, me, "Hawkeye's Shot", "Sorcery", "{4}{W}", msc10HawkeyesShotOracle, "{C}{C}{W}", game.CastSpellParams{Targets: b16TargetCard(big)}); err == nil {
		t.Fatal("cast at an untapped creature for {2}{W}")
	}
	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	toMainForCost(t, g2)
	big2 := pushPermanent(g2, opp2.ID, game.Card{Name: "Fresh Giant", TypeLine: "Creature — Giant", Power: 5, Toughness: 5})
	if _, err := msc10Cast(t, g2, me2, "Hawkeye's Shot", "Sorcery", "{4}{W}", msc10HawkeyesShotOracle, "{C}{C}{C}{C}{W}", game.CastSpellParams{Targets: b16TargetCard(big2)}); err != nil {
		t.Fatalf("full price: %v", err)
	}
	passPriorityAroundTable(t, g2)
	if !msc10InExile(g2, big2) {
		t.Error("the untapped creature was not exiled at full price")
	}
}

// --- General Thunderbolt Ross ---

func TestMsc10RossPumpsEveryAttackerWithTwoOthers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	ross := b12Push(g, me.ID, "General Thunderbolt Ross", "Legendary Creature — Human Soldier", msc10RossOracle, 1, 1)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	idle := b12Creature(g, me.ID, "Idle Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, ross, a, b)
	passPriorityAroundTable(t, g)
	for name, c := range map[string]struct {
		id   uuid.UUID
		want int
	}{"Ross": {ross, 2}, "Bear A": {a, 3}, "Bear B": {b, 3}, "the idle Bear": {idle, 2}} {
		if got := effectivePower(t, g, c.id); got != c.want {
			t.Errorf("%s has power %d, want %d", name, got, c.want)
		}
	}
	if !hasEffectiveKeyword(t, g, ross, "first strike") {
		t.Error("Ross has no first strike")
	}
}

func TestMsc10RossDoesNothingWithOnlyOneOtherAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	ross := b12Push(g, me.ID, "General Thunderbolt Ross", "Legendary Creature — Human Soldier", msc10RossOracle, 1, 1)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, ross, a)
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, a); got != 2 {
		t.Errorf("Bear A has power %d, want 2: only one other attacker", got)
	}
}

// --- Foggy Nelson, On Retainer ---

func TestMsc10FoggyNelsonCountersAndShieldsAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	foggy := castHoldCreature(t, g, "Foggy Nelson, On Retainer", "Legendary Creature — Human Advisor", msc10FoggyOracle, 1, 3)
	ch := pickTargetPrompt(t, g)
	if hasID(ch.PickTargetCards, foggy) {
		t.Error("Foggy Nelson offered himself as the \"another\" target")
	}
	answerPickTarget(t, g, mine)
	passPriorityAroundTable(t, g)
	if got := counterOn(g, mine, game.CounterPlusOne); got != 1 {
		t.Errorf("the target has %d +1/+1 counters, want 1", got)
	}
	if !hasEffectiveKeyword(t, g, mine, "hexproof") {
		t.Error("the target has no hexproof")
	}
	if got := counterOn(g, foggy, game.CounterPlusOne); got != 0 {
		t.Errorf("Foggy Nelson put a counter on himself: %d", got)
	}
}

func TestMsc10FoggyNelsonHexproofEndsWithTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castHoldCreature(t, g, "Foggy Nelson, On Retainer", "Legendary Creature — Human Advisor", msc10FoggyOracle, 1, 3)
	answerPickTarget(t, g, mine)
	passPriorityAroundTable(t, g)
	advanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if hasEffectiveKeyword(t, g, mine, "hexproof") {
		t.Error("hexproof outlived the turn")
	}
	if got := counterOn(g, mine, game.CounterPlusOne); got != 1 {
		t.Errorf("the counter is permanent: %d, want 1", got)
	}
}

// --- Flash Thompson, Spider-Fan ---

func TestMsc10FlashThompsonCanTapAndUntapBoth(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(mine) })
	castHoldCreature(t, g, "Flash Thompson, Spider-Fan", "Legendary Creature — Human Citizen", msc10FlashOracle, 2, 2)
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the enter trigger did not ask for modes")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0, 1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, theirs)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, mine)
	passPriorityAroundTable(t, g)
	if !tappedForTest(t, g, theirs) {
		t.Error("Heckle did not tap the first target")
	}
	if tappedForTest(t, g, mine) {
		t.Error("Hero Worship did not untap the second target")
	}
}

func TestMsc10FlashThompsonCanChooseOnlyOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castHoldCreature(t, g, "Flash Thompson, Spider-Fan", "Legendary Creature — Human Citizen", msc10FlashOracle, 2, 2)
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the enter trigger did not ask for modes")
	}
	answerMode(t, g, c, me.ID, 0)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if !tappedForTest(t, g, theirs) {
		t.Error("Heckle did not tap its target")
	}
}

// --- Flora Colossus ---

func TestMsc10FloraColossusIsAsBigAsYourLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	for i := 0; i < 5; i++ {
		pushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	}
	pushCatalogPermanent(g, opp.ID, "Island", "Basic Land — Island", "", false)
	flora := castHoldCreature(t, g, "Flora Colossus", "Creature — Treefolk", msc10FloraOracle, 0, 0)
	if p, tg := effectivePower(t, g, flora), effectiveToughness(t, g, flora); p != 5 || tg != 5 {
		t.Errorf("with five lands Flora Colossus is %d/%d, want 5/5", p, tg)
	}
	if !hasEffectiveKeyword(t, g, flora, "hexproof") {
		t.Error("Flora Colossus has no hexproof")
	}
	pushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	g.RunStateChecksForTest()
	if p := effectivePower(t, g, flora); p != 6 {
		t.Errorf("with six lands Flora Colossus has power %d, want 6", p)
	}
}

// --- Infernal Rebirth ---

func TestMsc10InfernalRebirthReturnsTwoCreatureCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := batch01GraveyardCard(me, "Dead Bear", "Creature — Bear")
	b := batch01GraveyardCard(me, "Dead Wolf", "Creature — Wolf")
	spell := batch01GraveyardCard(me, "Dead Spell", "Sorcery")
	castCatalogSpell(t, g, "Infernal Rebirth", "Instant", msc10RebirthOracle,
		append(b16TargetCard(a), b16TargetCard(b)...))
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) {
		t.Error("the two creature cards did not return to hand")
	}
	if me.Hand.Contains(spell) {
		t.Error("a noncreature card returned")
	}
}

func TestMsc10InfernalRebirthReturnsOneCreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := batch01GraveyardCard(me, "Dead Bear", "Creature — Bear")
	castCatalogSpell(t, g, "Infernal Rebirth", "Instant", msc10RebirthOracle, b16TargetCard(a))
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) {
		t.Error("the creature card did not return to hand")
	}
}

func TestMsc10InfernalRebirthRefusesNoTargetsAndThreeTargets(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := batch01GraveyardCard(me, "Dead Bear", "Creature — Bear")
	b := batch01GraveyardCard(me, "Dead Wolf", "Creature — Wolf")
	c := batch01GraveyardCard(me, "Dead Ox", "Creature — Ox")
	if err := castCatalogSpellErr(t, g, "Infernal Rebirth", "Instant", msc10RebirthOracle, nil); err == nil {
		t.Error("cast with no target")
	}
	three := append(append(b16TargetCard(a), b16TargetCard(b)...), b16TargetCard(c)...)
	if err := castCatalogSpellErr(t, g, "Infernal Rebirth", "Instant", msc10RebirthOracle, three); err == nil {
		t.Error("cast with three targets")
	}
}

func TestMsc10InfernalRebirthBasicLandcycles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Command Tower", "Land"),
	)
	id, _ := cycleFromHand(t, g, "Infernal Rebirth", "Instant", msc10RebirthOracle, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled card is not in the graveyard")
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the landcycling search did not ask")
	}
	if searchOptionNamed(g, c, "Command Tower") != uuid.Nil {
		t.Error("a nonbasic land was offered")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0]}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Hand.Contains(ids[0]) {
		t.Error("the basic land did not reach the hand")
	}
}

// --- Fogwell's Gym ---

func TestMsc10FogwellsGymTapsForRedAndHurtsYou(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	gym := pushCatalogPermanent(g, me.ID, "Fogwell's Gym", "Land", msc10GymOracle, false)
	life := me.Life
	if err := g.ActivateManaAbility(me.ID, gym, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("pool %v, want one {R}", got)
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d: the land deals 1 damage to you", me.Life, life-1)
	}
}

func TestMsc10FogwellsGymLoots(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	gym := pushCatalogPermanent(g, me.ID, "Fogwell's Gym", "Land", msc10GymOracle, false)
	discarded := ccHandCard(me, "Opt", "Instant", "{U}")
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{R}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	handBefore := me.Hand.Size()
	libBefore := me.Library.Size()
	if err := g.ActivateCatalogAbility(me.ID, gym, 0, game.ActivateAbilityParams{
		Strict: true, DiscardIDs: []uuid.UUID{discarded},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Graveyard.Contains(discarded) {
		t.Error("the discard is not paid at announce")
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != handBefore {
		t.Errorf("hand %d, want %d (discard one, draw one)", me.Hand.Size(), handBefore)
	}
	if me.Library.Size() != libBefore-1 {
		t.Errorf("library %d, want %d", me.Library.Size(), libBefore-1)
	}
	if !cardByID(g, gym).Tapped {
		t.Error("the Gym is not tapped")
	}
}

// --- Hawkeye, Clint Barton / Hulk, Bruce Banner / Hit-Monkey ---

func TestMsc10KeywordCreaturesCarryTheirKeywords(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	barton := b12Push(g, me.ID, "Hawkeye, Clint Barton", "Legendary Creature — Human Archer Hero", msc10BartonOracle, 3, 5)
	banner := b12Push(g, me.ID, "Hulk, Bruce Banner", "Legendary Creature — Gamma Berserker Hero", msc10BannerOracle, 7, 4)
	monkey := b12Push(g, me.ID, "Hit-Monkey", "Legendary Creature — Monkey Assassin", msc10HitMonkeyOracle, 3, 3)
	for _, c := range []struct {
		name string
		id   uuid.UUID
		kws  []string
	}{
		{"Hawkeye", barton, []string{"vigilance"}},
		{"Hulk", banner, []string{"trample"}},
		{"Hit-Monkey", monkey, []string{"reach", "vigilance", "deathtouch", "hexproof", "haste"}},
	} {
		for _, kw := range c.kws {
			if !hasEffectiveKeyword(t, g, c.id, kw) {
				t.Errorf("%s has no %s", c.name, kw)
			}
		}
	}
	if !msc10Spec(t, msc10HitMonkeyOracle).CantBeCountered {
		t.Error("Hit-Monkey's spell can be countered")
	}
}

// msc10Spec is Lookup that fails the test on a miss.
func msc10Spec(t *testing.T, oracle string) Spec {
	t.Helper()
	spec, ok := Lookup(oracle)
	if !ok {
		t.Fatalf("%s is not registered", oracle)
	}
	return spec
}

// --- Hulk, Brutal Brawler ---

func TestMsc10HulkBrawlerMustAttackAndCountersTheOthers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	hulk := b12Push(g, me.ID, "Hulk, Brutal Brawler", "Legendary Creature — Gamma Berserker Hero", msc10BrawlerOracle, 4, 4)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, hulk)
	passPriorityAroundTable(t, g)
	if got := counterOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("my other creature has %d counters, want 1", got)
	}
	if got := counterOn(g, hulk, game.CounterPlusOne); got != 0 {
		t.Errorf("Hulk countered himself: %d", got)
	}
	if got := counterOn(g, theirs, game.CounterPlusOne); got != 0 {
		t.Errorf("an opponent's creature got %d counters", got)
	}
}

func TestMsc10HulkBrawlerIsRequiredToAttack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Hulk, Brutal Brawler", "Legendary Creature — Gamma Berserker Hero", msc10BrawlerOracle, 4, 4)
	advanceTo(t, g, game.StepDeclareAttackers)
	err := g.PassPriority()
	if err == nil || !strings.Contains(err.Error(), "must attack") {
		t.Errorf("declaring no attackers with Hulk able to attack: err = %v, want the must-attack refusal", err)
	}
}

// --- Luke Cage, Hero for Hire ---

func TestMsc10LukeCageMakesATreasureAtYourCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	cage := b12Push(g, me.ID, "Luke Cage, Hero for Hire", "Legendary Creature — Human Mercenary Hero", msc10LukeCageOracle, 4, 4)
	if !hasEffectiveKeyword(t, g, cage, "trample") {
		t.Error("Luke Cage has no trample")
	}
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Treasure") == uuid.Nil {
		t.Error("no Treasure token at the beginning of combat")
	}
}

// --- Jarvis, Earth's Mightiest Butler ---

func TestMsc10JarvisDrawsForHeroSpellsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Jarvis, Earth's Mightiest Butler", "Legendary Creature — Human Advisor", msc10JarvisOracle, 1, 4)
	hero := handCardFull(me, "Test Hero", "Creature — Human Hero", "", "", nil)
	bear := handCardFull(me, "Test Bear", "Creature — Bear", "", "", nil)
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast a non-Hero: %v", err)
	}
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	if err := g.CastSpell(me.ID, hero, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast a Hero: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d: casting the Hero spent a card and Jarvis drew one back", me.Hand.Size(), hand)
	}
}

// --- Endless Ranks of HYDRA ---

func TestMsc10EndlessRanksMakesOneVillainPerOpponent(t *testing.T) {
	g := newCatalogGame(t)
	castCatalogSpell(t, g, "Endless Ranks of HYDRA", "Sorcery", msc10RanksOracle, nil)
	passPriorityAroundTable(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == me.ID && c.Name == "Villain" {
			n++
			if c.Power != 2 || c.Toughness != 1 || !game.HasKeyword(&c, "menace") {
				t.Errorf("Villain token is %d/%d, want 2/1 menace", c.Power, c.Toughness)
			}
		}
	}
	if n != 3 {
		t.Errorf("%d Villain tokens, want 3 (one per opponent at a four-player table)", n)
	}
}

// msc10RanksInGraveyard puts Endless Ranks of HYDRA in the active
// player's graveyard.
func msc10RanksInGraveyard(g *game.Game) (*game.Player, uuid.UUID) {
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: "Endless Ranks of HYDRA", TypeLine: "Sorcery", OracleID: msc10RanksOracle,
		ManaCost: "{3}{B}", Owner: me.ID, Controller: me.ID,
	})
	return me, id
}

// msc10CastCommander casts a free commander creature for `me`.
func msc10CastCommander(t *testing.T, g *game.Game, me *game.Player, name string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Legendary Creature — Human", Power: 2, Toughness: 2,
		IsCommander: true, Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

func TestMsc10EndlessRanksReturnsWhenYourCommanderEntersAndYouPay(t *testing.T) {
	g := newCatalogGame(t)
	me, ranks := msc10RanksInGraveyard(g)
	msc10CastCommander(t, g, me, "My Commander")
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("the commander entering did not offer to pay {1}{B}")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{B}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	answerPayUnless(t, g, me.ID, true)
	if !me.Hand.Contains(ranks) || me.Graveyard.Contains(ranks) {
		t.Error("the card did not return to hand")
	}
}

func TestMsc10EndlessRanksStaysWhenYouDeclineOrDontOwnTheCommander(t *testing.T) {
	g := newCatalogGame(t)
	me, ranks := msc10RanksInGraveyard(g)
	msc10CastCommander(t, g, me, "My Commander")
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("the commander entering did not offer to pay")
	}
	answerPayUnless(t, g, me.ID, false)
	if !me.Graveyard.Contains(ranks) {
		t.Error("a declined payment still returned the card")
	}
	// A creature that is not a commander offers nothing.
	castHoldCreature(t, g, "Plain Bear", "Creature — Bear", "", 2, 2)
	if hasPayUnlessFor(g, me.ID) {
		t.Error("a creature that is not a commander offered the payment")
	}
}

func TestMsc10EndlessRanksTriggersWhenYourCommanderAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, ranks := msc10RanksInGraveyard(g)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	cmdr := b12Push(g, me.ID, "My Commander", "Legendary Creature — Human", "", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == cmdr {
				g.Battlefield.Cards[i].IsCommander = true
			}
		}
	})
	declareAttack(t, g, opp.ID, cmdr)
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("the commander attacking did not offer to pay {1}{B}")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{B}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	answerPayUnless(t, g, me.ID, true)
	if !me.Hand.Contains(ranks) {
		t.Error("the card did not return to hand")
	}
}

// --- HYDRA Disintegrator ---

func TestMsc10HydraDisintegratorMakesAVillainAndEquipsIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "HYDRA Disintegrator", "Artifact — Equipment", msc10DisintegratorOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	villain := findBattlefieldByName(g, "Villain")
	if villain == uuid.Nil {
		t.Fatal("no Villain token")
	}
	if p, tg := effectivePower(t, g, villain), effectiveToughness(t, g, villain); p != 5 || tg != 4 {
		t.Errorf("the equipped Villain is %d/%d, want 5/4 (2/1 plus +3/+3)", p, tg)
	}
	if !hasEffectiveKeyword(t, g, villain, "menace") {
		t.Error("the Villain has no menace")
	}
	if c := cardByID(g, villain); c.Controller != me.ID {
		t.Error("the Villain is not yours")
	}
}
