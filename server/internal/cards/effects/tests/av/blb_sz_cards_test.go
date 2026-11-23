package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blb_sz_cards_test.go — Bloomburrow cards S–Z (#4467). Each test plays
// the real catalog entry and asserts what its text says.

const (
	blbSZSavorOracle      = "dfc1dc36-0cc2-47cc-b0dd-1cc7c31057aa"
	blbSZSquireOracle     = "e2f57e48-ff5b-4a42-9b7c-1bd23b0d78ed"
	blbSZShoreUpOracle    = "c9d8c983-2845-42e6-b6b8-bf7cf2729b23"
	blbSZShortBowOracle   = "ce5f5ff9-13f7-41eb-ada0-7f73de42acef"
	blbSZShrikeOracle     = "d014b02f-a8b8-4b97-96e0-f588792069f0"
	blbSZSonarOracle      = "d8c582bc-47de-47ec-8bf9-b2c848e27ab8"
	blbSZSunshowerOracle  = "7a70faf5-cc04-402b-92e3-67746788f510"
	blbSZAnglerOracle     = "60ac490c-b489-4ad2-bcc2-3babaabf8ccb"
	blbSZOtterOracle      = "5dee663f-9e2e-4503-8086-fc51664fdf29"
	blbSZRootweaverOracle = "d17d4003-e1c9-42be-ac76-3b5500fad485"
	blbSZVinereapOracle   = "db649e15-24ca-4fd8-9f0b-f1df28c9c57d"
	blbSZElderOracle      = "33fcb05b-246d-4b82-8b0f-44d5f891b612"
)

func TestBlbSZEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		blbSZSavorOracle:      "Savor",
		blbSZSquireOracle:     "Seedpod Squire",
		blbSZShoreUpOracle:    "Shore Up",
		blbSZShortBowOracle:   "Short Bow",
		blbSZShrikeOracle:     "Shrike Force",
		blbSZSonarOracle:      "Sonar Strike",
		blbSZSunshowerOracle:  "Sunshower Druid",
		blbSZAnglerOracle:     "Tempest Angler",
		blbSZOtterOracle:      "Thieving Otter",
		blbSZRootweaverOracle: "Three Tree Rootweaver",
		blbSZVinereapOracle:   "Vinereap Mentor",
		blbSZElderOracle:      "Warren Elder",
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

// --- Savor -------------------------------------------------------------------

func TestBlbSZSavorShrinksAndMakesAFood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	theirs := effectstest.BlbB1Push(g, opp.ID, "Big", "Creature — Beast", "", 4, 4)
	effectstest.CastCatalogSpell(t, g, "Savor", "Instant", blbSZSavorOracle, effectstest.CardTarget(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("the target is %d/%d, want 2/2", p, tough)
	}
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Food"); n != 1 {
		t.Errorf("I have %d Food tokens, want 1", n)
	}
	if n, _ := effectstest.CountTokensNamed(g, opp.ID, "Food"); n != 0 {
		t.Errorf("the opponent has %d Food tokens, want 0", n)
	}
}

func TestBlbSZSavorKillsATwoToughnessCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp, _ := effectstest.BlbB1Seats(g)
	bear := effectstest.BlbB1Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Savor", "Instant", blbSZSavorOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("a 2/2 given -2/-2 should die")
	}
}

// --- Seedpod Squire ----------------------------------------------------------

func TestBlbSZSeedpodSquirePumpsACreatureWithoutFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	squire := effectstest.BlbB1Push(g, me.ID, "Seedpod Squire", "Creature — Bird Scout", blbSZSquireOracle, 3, 3)
	if !effectstest.HasKeywordOnBattlefield(t, g, squire, "flying") {
		t.Error("the Squire should have flying")
	}
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(squire, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	ch := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: squire}); err == nil {
		t.Fatal("the Squire itself was accepted; it has flying")
	}
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the bear is %d/%d, want 3/3", p, tough)
	}
}

// --- Shore Up ----------------------------------------------------------------

func TestBlbSZShoreUpPumpsGrantsHexproofAndUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	mine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Tapped: true, Owner: me.ID, Controller: me.ID,
	})
	theirs := effectstest.BlbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Shore Up", "Instant", blbSZShoreUpOracle, effectstest.CardTarget(theirs)); err == nil {
		t.Fatal("an opponent's creature was accepted; the target is a creature you control")
	}
	effectstest.CastCatalogSpell(t, g, "Shore Up", "Instant", blbSZShoreUpOracle, effectstest.CardTarget(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("the bear is %d/%d, want 3/3", p, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, mine, "hexproof") {
		t.Error("the bear should have hexproof")
	}
	if effectstest.B16Tapped(t, g, mine) {
		t.Error("the bear should be untapped")
	}
}

// --- Short Bow ---------------------------------------------------------------

func TestBlbSZShortBowEquipsForOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _, _ := effectstest.BlbB1Seats(g)
	bow := effectstest.PushCatalogPermanent(g, me.ID, "Short Bow", "Artifact — Equipment", blbSZShortBowOracle, false)
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.B06AddMana(me, "C")
	effectstest.EquipTo(t, g, me.ID, bow, bear)
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the equipped bear is %d/%d, want 3/3", p, tough)
	}
	for _, kw := range []string{"reach", "vigilance"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, bear, kw) {
			t.Errorf("the equipped bear should have %s", kw)
		}
	}
}

// --- Shrike Force ------------------------------------------------------------

func TestBlbSZShrikeForceHasItsThreeKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	shrike := effectstest.BlbB1Push(g, me.ID, "Shrike Force", "Creature — Bird Knight", blbSZShrikeOracle, 1, 3)
	for _, kw := range []string{"flying", "double strike", "vigilance"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, shrike, kw) {
			t.Errorf("Shrike Force should have %s", kw)
		}
	}
}

// --- Sonar Strike ------------------------------------------------------------

func TestBlbSZSonarStrikeHitsATappedCreatureOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	idle := effectstest.BlbB1Push(g, opp.ID, "Idle", "Creature — Beast", "", 5, 5)
	tapped := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tapped", TypeLine: "Creature — Beast", Power: 5, Toughness: 5,
		Tapped: true, Owner: opp.ID, Controller: opp.ID,
	})
	if err := effectstest.CastCatalogSpellErr(t, g, "Sonar Strike", "Instant", blbSZSonarOracle, effectstest.CardTarget(idle)); err == nil {
		t.Fatal("an untapped creature out of combat was accepted")
	}
	life := effectstest.LifeOf(g, me.ID)
	effectstest.CastCatalogSpell(t, g, "Sonar Strike", "Instant", blbSZSonarOracle, effectstest.CardTarget(tapped))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageOn(t, g, tapped); got != 4 {
		t.Errorf("the tapped creature has %d damage, want 4", got)
	}
	if got := effectstest.LifeOf(g, me.ID); got != life {
		t.Errorf("no Bat: I am at %d, want %d", got, life)
	}
}

func TestBlbSZSonarStrikeGainsThreeWithABat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.BlbB1Push(g, me.ID, "Bat", "Creature — Bat Cleric", "", 1, 1)
	tapped := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tapped", TypeLine: "Creature — Beast", Power: 5, Toughness: 5,
		Tapped: true, Owner: opp.ID, Controller: opp.ID,
	})
	life := effectstest.LifeOf(g, me.ID)
	effectstest.CastCatalogSpell(t, g, "Sonar Strike", "Instant", blbSZSonarOracle, effectstest.CardTarget(tapped))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != life+3 {
		t.Errorf("with a Bat I am at %d, want %d", got, life+3)
	}
}

// --- Sunshower Druid ---------------------------------------------------------

func TestBlbSZSunshowerDruidCountersAndGainsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	life := effectstest.LifeOf(g, me.ID)
	effectstest.CastHoldCreature(t, g, "Sunshower Druid", "Creature — Frog Druid", blbSZSunshowerOracle, 0, 2)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, bear); got != 1 {
		t.Errorf("the bear has %d +1/+1 counters, want 1", got)
	}
	if got := effectstest.LifeOf(g, me.ID); got != life+1 {
		t.Errorf("I am at %d, want %d", got, life+1)
	}
}

// --- Tempest Angler ----------------------------------------------------------

func TestBlbSZTempestAnglerGrowsOnNoncreatureSpellsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _, _ := effectstest.BlbB1Seats(g)
	angler := effectstest.PushCatalogPermanent(g, me.ID, "Tempest Angler", "Creature — Otter Wizard", blbSZAnglerOracle, false)
	effectstest.CastCatalogSpell(t, g, "Test Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOnes(g, angler); n != 0 {
		t.Fatalf("%d counters after a creature spell, want 0", n)
	}
	effectstest.CastCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOnes(g, angler); n != 1 {
		t.Errorf("%d counters after a noncreature spell, want 1", n)
	}
}

// --- Thieving Otter ----------------------------------------------------------

func TestBlbSZThievingOtterDrawsWhenItDamagesAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.SeedLibrary(me, "A", "B")
	otter := effectstest.BlbB1Push(g, me.ID, "Thieving Otter", "Creature — Otter", blbSZOtterOracle, 2, 2)
	bear := effectstest.BlbB1Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	before := me.Hand.Size()
	effectstest.DealNoncombatDamage(g, me.ID, otter, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before {
		t.Fatalf("damage to a creature drew: hand %d, want %d", got, before)
	}
	effectstest.DealCombatDamage(g, me.ID, otter, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before {
		t.Fatalf("damage to its controller drew: hand %d, want %d", got, before)
	}
	effectstest.DealCombatDamage(g, me.ID, otter, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d, want %d after damage to an opponent", got, before+1)
	}
}

// --- Three Tree Rootweaver ---------------------------------------------------

func TestBlbSZThreeTreeRootweaverTapsForMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _, _ := effectstest.BlbB1Seats(g)
	sick := effectstest.PushCatalogPermanent(g, me.ID, "Three Tree Rootweaver", "Creature — Mole Druid", blbSZRootweaverOracle, true)
	if err := g.ActivateManaAbility(me.ID, sick, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("a summoning-sick Rootweaver tapped for mana")
	}
	ready := effectstest.PushCatalogPermanent(g, me.ID, "Three Tree Rootweaver", "Creature — Mole Druid", blbSZRootweaverOracle, false)
	if err := g.ActivateManaAbility(me.ID, ready, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if !effectstest.B16Tapped(t, g, ready) {
		t.Error("the Rootweaver should be tapped")
	}
}

// --- Vinereap Mentor ---------------------------------------------------------

func TestBlbSZVinereapMentorMakesAFoodOnEntryAndDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	mentor := effectstest.CastHoldCreature(t, g, "Vinereap Mentor", "Creature — Squirrel Druid", blbSZVinereapOracle, 3, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Food"); n != 1 {
		t.Fatalf("%d Food tokens after it entered, want 1", n)
	}
	effectstest.KillCreature(t, g, me.ID, mentor)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Food"); n != 2 {
		t.Errorf("%d Food tokens after it died, want 2", n)
	}
}

// --- Warren Elder ------------------------------------------------------------

func TestBlbSZWarrenElderPumpsYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	elder := effectstest.BlbB1Push(g, me.ID, "Warren Elder", "Creature — Rabbit Cleric", blbSZElderOracle, 2, 2)
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.BlbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "C", "W")
	effectstest.B16Activate(t, g, me.ID, elder, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{elder, bear} {
		if p, tough := effectstest.PtOf(t, g, id); p != 3 || tough != 3 {
			t.Errorf("my creature is %d/%d, want 3/3", p, tough)
		}
	}
	if p, tough := effectstest.PtOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("the opponent's creature is %d/%d, want 2/2", p, tough)
	}
}
