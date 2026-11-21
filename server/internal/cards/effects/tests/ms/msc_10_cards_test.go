package testsms

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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

func msc10InExile(g *game.Game, id uuid.UUID) bool {
	z := g.FindCardZoneForEffect(id)
	return z != nil && z.Kind == game.ZoneExile
}

// --- Captain Marvel, Shooting Star ---

func TestMsc10CaptainMarvelEntersAndExilesATargetThenBothPlayersGain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)
	myLife, oppLife := me.Life, opp.Life
	effectstest.CastHoldCreature(t, g, "Captain Marvel, Shooting Star", "Legendary Creature — Human Kree Hero", msc10CaptainMarvelOracle, 6, 6)
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	marvel := effectstest.B12Push(g, me.ID, "Captain Marvel, Shooting Star", "Legendary Creature — Human Kree Hero", msc10CaptainMarvelOracle, 6, 6)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	oppLife := opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, marvel)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !msc10InExile(g, theirs) {
		t.Fatal("the attack trigger did not exile the target")
	}
	if opp.Life != oppLife+2 {
		t.Errorf("the exiled creature's controller has %d life, want %d", opp.Life, oppLife+2)
	}
}

func TestMsc10CaptainMarvelGainsWhenAnyOtherCreatureIsExiled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "Captain Marvel, Shooting Star", "Legendary Creature — Human Kree Hero", msc10CaptainMarvelOracle, 6, 6)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
	myLife, oppLife := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "Swords to Plowshares", "Instant", msc10SwordsOracle, effectstest.B16TargetCard(theirs))
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "Captain Marvel, Shooting Star", "Legendary Creature — Human Kree Hero", msc10CaptainMarvelOracle, 6, 6)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
	myLife := me.Life
	effectstest.B27Kill(g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife {
		t.Errorf("a creature that died paid out: %d life, want %d", me.Life, myLife)
	}
}

func TestMsc10CaptainMarvelDoesNotTriggerOnHerOwnExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	marvel := effectstest.B12Push(g, me.ID, "Captain Marvel, Shooting Star", "Legendary Creature — Human Kree Hero", msc10CaptainMarvelOracle, 6, 6)
	myLife := me.Life
	effectstest.CastCatalogSpell(t, g, "Swords to Plowshares", "Instant", msc10SwordsOracle, effectstest.B16TargetCard(marvel))
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife+6 {
		t.Errorf("life %d, want %d: only Swords' payout, not her own second ability", me.Life, myLife+6)
	}
}

// --- Hawkeye's Shot ---

func TestMsc10HawkeyesShotCostsTwoLessAtATappedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tapped := effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Tired", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Tapped: true})
	fresh := effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Fresh", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	price := func(refs []game.TargetRef) int {
		return effectstest.SelfPriced(t, g, me, msc10HawkeyesShotOracle, "Sorcery", "{4}{W}", game.ZoneHand, refs).ManaValue()
	}
	if got := price(effectstest.B16TargetCard(tapped)); got != 3 {
		t.Errorf("at a tapped creature: %d, want 3", got)
	}
	if got := price(effectstest.B16TargetCard(fresh)); got != 5 {
		t.Errorf("at an untapped creature: %d, want 5", got)
	}
}

func TestMsc10HawkeyesShotExilesATappedCreatureForThreeMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	big := effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Tired Giant", TypeLine: "Creature — Giant", Power: 5, Toughness: 5, Tapped: true})
	if _, err := effectstest.Msc10Cast(t, g, me, "Hawkeye's Shot", "Sorcery", "{4}{W}", msc10HawkeyesShotOracle, "{C}{C}{W}", game.CastSpellParams{Targets: effectstest.B16TargetCard(big)}); err != nil {
		t.Fatalf("cast for {2}{W}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !msc10InExile(g, big) {
		t.Error("the tapped creature was not exiled")
	}
}

func TestMsc10HawkeyesShotAtAnUntappedCreatureCostsFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	big := effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Fresh Giant", TypeLine: "Creature — Giant", Power: 5, Toughness: 5})
	if _, err := effectstest.Msc10Cast(t, g, me, "Hawkeye's Shot", "Sorcery", "{4}{W}", msc10HawkeyesShotOracle, "{C}{C}{W}", game.CastSpellParams{Targets: effectstest.B16TargetCard(big)}); err == nil {
		t.Fatal("cast at an untapped creature for {2}{W}")
	}
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	effectstest.ToMainForCost(t, g2)
	big2 := effectstest.PushPermanent(g2, opp2.ID, game.Card{Name: "Fresh Giant", TypeLine: "Creature — Giant", Power: 5, Toughness: 5})
	if _, err := effectstest.Msc10Cast(t, g2, me2, "Hawkeye's Shot", "Sorcery", "{4}{W}", msc10HawkeyesShotOracle, "{C}{C}{C}{C}{W}", game.CastSpellParams{Targets: effectstest.B16TargetCard(big2)}); err != nil {
		t.Fatalf("full price: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if !msc10InExile(g2, big2) {
		t.Error("the untapped creature was not exiled at full price")
	}
}

// --- General Thunderbolt Ross ---

func TestMsc10RossPumpsEveryAttackerWithTwoOthers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	ross := effectstest.B12Push(g, me.ID, "General Thunderbolt Ross", "Legendary Creature — Human Soldier", msc10RossOracle, 1, 1)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Idle Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, ross, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	for name, c := range map[string]struct {
		id   uuid.UUID
		want int
	}{"Ross": {ross, 2}, "Bear A": {a, 3}, "Bear B": {b, 3}, "the idle Bear": {idle, 2}} {
		if got := effectstest.EffectivePower(t, g, c.id); got != c.want {
			t.Errorf("%s has power %d, want %d", name, got, c.want)
		}
	}
	if !effectstest.HasEffectiveKeyword(t, g, ross, "first strike") {
		t.Error("Ross has no first strike")
	}
}

func TestMsc10RossDoesNothingWithOnlyOneOtherAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	ross := effectstest.B12Push(g, me.ID, "General Thunderbolt Ross", "Legendary Creature — Human Soldier", msc10RossOracle, 1, 1)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, ross, a)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, a); got != 2 {
		t.Errorf("Bear A has power %d, want 2: only one other attacker", got)
	}
}

// --- Foggy Nelson, On Retainer ---

func TestMsc10FoggyNelsonCountersAndShieldsAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	foggy := effectstest.CastHoldCreature(t, g, "Foggy Nelson, On Retainer", "Legendary Creature — Human Advisor", msc10FoggyOracle, 1, 3)
	ch := effectstest.PickTargetPrompt(t, g)
	if effectstest.HasID(ch.PickTargetCards, foggy) {
		t.Error("Foggy Nelson offered himself as the \"another\" target")
	}
	effectstest.AnswerPickTarget(t, g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, mine, game.CounterPlusOne); got != 1 {
		t.Errorf("the target has %d +1/+1 counters, want 1", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, mine, "hexproof") {
		t.Error("the target has no hexproof")
	}
	if got := effectstest.CounterOn(g, foggy, game.CounterPlusOne); got != 0 {
		t.Errorf("Foggy Nelson put a counter on himself: %d", got)
	}
}

func TestMsc10FoggyNelsonHexproofEndsWithTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Foggy Nelson, On Retainer", "Legendary Creature — Human Advisor", msc10FoggyOracle, 1, 3)
	effectstest.AnswerPickTarget(t, g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if effectstest.HasEffectiveKeyword(t, g, mine, "hexproof") {
		t.Error("hexproof outlived the turn")
	}
	if got := effectstest.CounterOn(g, mine, game.CounterPlusOne); got != 1 {
		t.Errorf("the counter is permanent: %d, want 1", got)
	}
}

// --- Flash Thompson, Spider-Fan ---

func TestMsc10FlashThompsonCanTapAndUntapBoth(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(mine) })
	effectstest.CastHoldCreature(t, g, "Flash Thompson, Spider-Fan", "Legendary Creature — Human Citizen", msc10FlashOracle, 2, 2)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the enter trigger did not ask for modes")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0, 1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedForTest(t, g, theirs) {
		t.Error("Heckle did not tap the first target")
	}
	if effectstest.TappedForTest(t, g, mine) {
		t.Error("Hero Worship did not untap the second target")
	}
}

func TestMsc10FlashThompsonCanChooseOnlyOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Flash Thompson, Spider-Fan", "Legendary Creature — Human Citizen", msc10FlashOracle, 2, 2)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the enter trigger did not ask for modes")
	}
	effectstest.AnswerMode(t, g, c, me.ID, 0)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedForTest(t, g, theirs) {
		t.Error("Heckle did not tap its target")
	}
}

// --- Flora Colossus ---

func TestMsc10FloraColossusIsAsBigAsYourLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 5; i++ {
		effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	}
	effectstest.PushCatalogPermanent(g, opp.ID, "Island", "Basic Land — Island", "", false)
	flora := effectstest.CastHoldCreature(t, g, "Flora Colossus", "Creature — Treefolk", msc10FloraOracle, 0, 0)
	if p, tg := effectstest.EffectivePower(t, g, flora), effectstest.EffectiveToughness(t, g, flora); p != 5 || tg != 5 {
		t.Errorf("with five lands Flora Colossus is %d/%d, want 5/5", p, tg)
	}
	if !effectstest.HasEffectiveKeyword(t, g, flora, "hexproof") {
		t.Error("Flora Colossus has no hexproof")
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	g.RunStateChecksForTest()
	if p := effectstest.EffectivePower(t, g, flora); p != 6 {
		t.Errorf("with six lands Flora Colossus has power %d, want 6", p)
	}
}

// --- Infernal Rebirth ---

func TestMsc10InfernalRebirthReturnsTwoCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.Batch01GraveyardCard(me, "Dead Bear", "Creature — Bear")
	b := effectstest.Batch01GraveyardCard(me, "Dead Wolf", "Creature — Wolf")
	spell := effectstest.Batch01GraveyardCard(me, "Dead Spell", "Sorcery")
	effectstest.CastCatalogSpell(t, g, "Infernal Rebirth", "Instant", msc10RebirthOracle,
		append(effectstest.B16TargetCard(a), effectstest.B16TargetCard(b)...))
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) {
		t.Error("the two creature cards did not return to hand")
	}
	if me.Hand.Contains(spell) {
		t.Error("a noncreature card returned")
	}
}

func TestMsc10InfernalRebirthReturnsOneCreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.Batch01GraveyardCard(me, "Dead Bear", "Creature — Bear")
	effectstest.CastCatalogSpell(t, g, "Infernal Rebirth", "Instant", msc10RebirthOracle, effectstest.B16TargetCard(a))
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) {
		t.Error("the creature card did not return to hand")
	}
}

func TestMsc10InfernalRebirthRefusesNoTargetsAndThreeTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.Batch01GraveyardCard(me, "Dead Bear", "Creature — Bear")
	b := effectstest.Batch01GraveyardCard(me, "Dead Wolf", "Creature — Wolf")
	c := effectstest.Batch01GraveyardCard(me, "Dead Ox", "Creature — Ox")
	if err := effectstest.CastCatalogSpellErr(t, g, "Infernal Rebirth", "Instant", msc10RebirthOracle, nil); err == nil {
		t.Error("cast with no target")
	}
	three := append(append(effectstest.B16TargetCard(a), effectstest.B16TargetCard(b)...), effectstest.B16TargetCard(c)...)
	if err := effectstest.CastCatalogSpellErr(t, g, "Infernal Rebirth", "Instant", msc10RebirthOracle, three); err == nil {
		t.Error("cast with three targets")
	}
}

func TestMsc10InfernalRebirthBasicLandcycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Command Tower", "Land"),
	)
	id, _ := effectstest.CycleFromHand(t, g, "Infernal Rebirth", "Instant", msc10RebirthOracle, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled card is not in the graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the landcycling search did not ask")
	}
	if effectstest.SearchOptionNamed(g, c, "Command Tower") != uuid.Nil {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	gym := effectstest.PushCatalogPermanent(g, me.ID, "Fogwell's Gym", "Land", msc10GymOracle, false)
	life := me.Life
	if err := g.ActivateManaAbility(me.ID, gym, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("pool %v, want one {R}", got)
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d: the land deals 1 damage to you", me.Life, life-1)
	}
}

func TestMsc10FogwellsGymLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	gym := effectstest.PushCatalogPermanent(g, me.ID, "Fogwell's Gym", "Land", msc10GymOracle, false)
	discarded := effectstest.CcHandCard(me, "Opt", "Instant", "{U}")
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
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != handBefore {
		t.Errorf("hand %d, want %d (discard one, draw one)", me.Hand.Size(), handBefore)
	}
	if me.Library.Size() != libBefore-1 {
		t.Errorf("library %d, want %d", me.Library.Size(), libBefore-1)
	}
	if !effectstest.CardByID(g, gym).Tapped {
		t.Error("the Gym is not tapped")
	}
}

// --- Hawkeye, Clint Barton / Hulk, Bruce Banner / Hit-Monkey ---

func TestMsc10KeywordCreaturesCarryTheirKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	barton := effectstest.B12Push(g, me.ID, "Hawkeye, Clint Barton", "Legendary Creature — Human Archer Hero", msc10BartonOracle, 3, 5)
	banner := effectstest.B12Push(g, me.ID, "Hulk, Bruce Banner", "Legendary Creature — Gamma Berserker Hero", msc10BannerOracle, 7, 4)
	monkey := effectstest.B12Push(g, me.ID, "Hit-Monkey", "Legendary Creature — Monkey Assassin", msc10HitMonkeyOracle, 3, 3)
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
			if !effectstest.HasEffectiveKeyword(t, g, c.id, kw) {
				t.Errorf("%s has no %s", c.name, kw)
			}
		}
	}
	if !msc10Spec(t, msc10HitMonkeyOracle).CantBeCountered {
		t.Error("Hit-Monkey's spell can be countered")
	}
}

// msc10Spec is Lookup that fails the test on a miss.
func msc10Spec(t *testing.T, oracle string) effects.Spec {
	t.Helper()
	spec, ok := effects.Lookup(oracle)
	if !ok {
		t.Fatalf("%s is not registered", oracle)
	}
	return spec
}

// --- Hulk, Brutal Brawler ---

func TestMsc10HulkBrawlerMustAttackAndCountersTheOthers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	hulk := effectstest.B12Push(g, me.ID, "Hulk, Brutal Brawler", "Legendary Creature — Gamma Berserker Hero", msc10BrawlerOracle, 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, hulk)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("my other creature has %d counters, want 1", got)
	}
	if got := effectstest.CounterOn(g, hulk, game.CounterPlusOne); got != 0 {
		t.Errorf("Hulk countered himself: %d", got)
	}
	if got := effectstest.CounterOn(g, theirs, game.CounterPlusOne); got != 0 {
		t.Errorf("an opponent's creature got %d counters", got)
	}
}

func TestMsc10HulkBrawlerIsRequiredToAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Hulk, Brutal Brawler", "Legendary Creature — Gamma Berserker Hero", msc10BrawlerOracle, 4, 4)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	err := g.PassPriority()
	if err == nil || !strings.Contains(err.Error(), "must attack") {
		t.Errorf("declaring no attackers with Hulk able to attack: err = %v, want the must-attack refusal", err)
	}
}

// --- Luke Cage, Hero for Hire ---

func TestMsc10LukeCageMakesATreasureAtYourCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	cage := effectstest.B12Push(g, me.ID, "Luke Cage, Hero for Hire", "Legendary Creature — Human Mercenary Hero", msc10LukeCageOracle, 4, 4)
	if !effectstest.HasEffectiveKeyword(t, g, cage, "trample") {
		t.Error("Luke Cage has no trample")
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldByName(g, "Treasure") == uuid.Nil {
		t.Error("no Treasure token at the beginning of combat")
	}
}

// --- Jarvis, Earth's Mightiest Butler ---

func TestMsc10JarvisDrawsForHeroSpellsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Jarvis, Earth's Mightiest Butler", "Legendary Creature — Human Advisor", msc10JarvisOracle, 1, 4)
	hero := effectstest.HandCardFull(me, "Test Hero", "Creature — Human Hero", "", "", nil)
	bear := effectstest.HandCardFull(me, "Test Bear", "Creature — Bear", "", "", nil)
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast a non-Hero: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	if err := g.CastSpell(me.ID, hero, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast a Hero: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d: casting the Hero spent a card and Jarvis drew one back", me.Hand.Size(), hand)
	}
}

// --- Endless Ranks of HYDRA ---

func TestMsc10EndlessRanksMakesOneVillainPerOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.CastCatalogSpell(t, g, "Endless Ranks of HYDRA", "Sorcery", msc10RanksOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
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

func TestMsc10EndlessRanksReturnsWhenYourCommanderEntersAndYouPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, ranks := msc10RanksInGraveyard(g)
	effectstest.Msc10CastCommander(t, g, me, "My Commander")
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("the commander entering did not offer to pay {1}{B}")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{B}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	if !me.Hand.Contains(ranks) || me.Graveyard.Contains(ranks) {
		t.Error("the card did not return to hand")
	}
}

func TestMsc10EndlessRanksStaysWhenYouDeclineOrDontOwnTheCommander(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, ranks := msc10RanksInGraveyard(g)
	effectstest.Msc10CastCommander(t, g, me, "My Commander")
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("the commander entering did not offer to pay")
	}
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	if !me.Graveyard.Contains(ranks) {
		t.Error("a declined payment still returned the card")
	}
	// A creature that is not a commander offers nothing.
	effectstest.CastHoldCreature(t, g, "Plain Bear", "Creature — Bear", "", 2, 2)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("a creature that is not a commander offered the payment")
	}
}

func TestMsc10EndlessRanksTriggersWhenYourCommanderAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, ranks := msc10RanksInGraveyard(g)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	cmdr := effectstest.B12Push(g, me.ID, "My Commander", "Legendary Creature — Human", "", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == cmdr {
				g.Battlefield.Cards[i].IsCommander = true
			}
		}
	})
	effectstest.DeclareAttack(t, g, opp.ID, cmdr)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("the commander attacking did not offer to pay {1}{B}")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{B}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	if !me.Hand.Contains(ranks) {
		t.Error("the card did not return to hand")
	}
}

// --- HYDRA Disintegrator ---

func TestMsc10HydraDisintegratorMakesAVillainAndEquipsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "HYDRA Disintegrator", "Artifact — Equipment", msc10DisintegratorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	villain := effectstest.FindBattlefieldByName(g, "Villain")
	if villain == uuid.Nil {
		t.Fatal("no Villain token")
	}
	if p, tg := effectstest.EffectivePower(t, g, villain), effectstest.EffectiveToughness(t, g, villain); p != 5 || tg != 4 {
		t.Errorf("the equipped Villain is %d/%d, want 5/4 (2/1 plus +3/+3)", p, tg)
	}
	if !effectstest.HasEffectiveKeyword(t, g, villain, "menace") {
		t.Error("the Villain has no menace")
	}
	if c := effectstest.CardByID(g, villain); c.Controller != me.ID {
		t.Error("the Villain is not yours")
	}
}
