package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// block_requirement_leftovers_1684_test.go — the leftover cards from
// #1684 (#1693's builder stopped at twelve to keep that PR small): the
// three plain Provoke creatures (keywords plus Provoke()), Hunt Down (a
// positional two-target BlocksAttackerUntilEOT), three more Lure-on-a-
// creature/Equipment cards, and three pump-and-must-be-blocked
// sorceries. Each test: the card's main behaviour through a real attack
// and the defender's declaration, plus one illegal block declaration
// refused.

const (
	crestedCraghornOracle   = "597c390e-deeb-4452-aeda-7cf9312a8b01"
	lowlandTrackerOracle    = "de06b33f-0793-419a-bf65-1961634af8e7"
	brontotheriumOracle     = "6e3301a5-a99b-4b17-aa40-34ccc5904977"
	huntDownOracle          = "29b0f3bd-d9bd-4c54-b4c3-9c2e01720e34"
	treeshakerChimeraOracle = "8b119b69-4794-443e-aa13-747c125b5b1b"
	breakerOfArmiesOracle   = "0be1ace3-2b3d-4465-b63a-be523f73df36"
	ochranAssassinOracle    = "ef6ecb67-7e49-4d40-84f2-1e7c02960a0d"
	nemesisMaskOracle       = "34062d79-457f-4af3-b9c6-cab2e4d17581"
	compelledDuelOracle     = "256cc907-2a13-4f43-9d37-2e52034cdaf4"
	emergentGrowthOracle    = "2afd64dd-94b3-4bc6-8411-465b3e07f3c9"
	enlargeOracle           = "405cd74a-6b44-4a80-b53c-ff1406d8ff45"
)

// --- Crested Craghorn, Lowland Tracker, Brontotherium: keywords + Provoke ---

// TestCrestedCraghornHasHasteAndProvokes checks the printed keyword and
// that Provoke behaves exactly as it does off Goblin Grappler.
func TestCrestedCraghornHasHasteAndProvokes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seats1684(g)
	craghorn := reqCreature(g, me.ID, "Crested Craghorn", crestedCraghornOracle, 4, 1)
	if !hasEffectiveAbility(t, g, craghorn, "haste") {
		t.Fatal("Crested Craghorn lacks haste")
	}
	other := reqCreature(g, me.ID, "Other Bear", "", 2, 2, "haste")
	victim := reqCreature(g, opp.ID, "Their Bear", "", 2, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(victim) })

	attackAndTrigger(t, g, craghorn, other)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, victim); c.Tapped {
		t.Fatal("Provoke did not untap the target")
	}

	onToBlockers(t, g)
	reqRefusal(t, reqBlock(t, g, other, victim))
	br := reqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); got != "Their Bear must block Crested Craghorn if able." {
		t.Errorf("sentence = %q", got)
	}
	if err := reqBlock(t, g, craghorn, victim); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
}

// TestLowlandTrackerHasFirstStrikeAndProvokes.
func TestLowlandTrackerHasFirstStrikeAndProvokes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seats1684(g)
	tracker := reqCreature(g, me.ID, "Lowland Tracker", lowlandTrackerOracle, 2, 2, "haste")
	if !hasEffectiveAbility(t, g, tracker, "first strike") {
		t.Fatal("Lowland Tracker lacks first strike")
	}
	other := reqCreature(g, me.ID, "Other Bear", "", 2, 2, "haste")
	victim := reqCreature(g, opp.ID, "Their Bear", "", 2, 2)

	attackAndTrigger(t, g, tracker, other)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	onToBlockers(t, g)
	reqRefusal(t, reqBlock(t, g, other, victim))
	br := reqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); got != "Their Bear must block Lowland Tracker if able." {
		t.Errorf("sentence = %q", got)
	}
	if err := reqBlock(t, g, tracker, victim); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
}

// TestBrontotheriumHasTrampleAndProvokes.
func TestBrontotheriumHasTrampleAndProvokes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seats1684(g)
	bronto := reqCreature(g, me.ID, "Brontotherium", brontotheriumOracle, 5, 3, "haste")
	if !hasEffectiveAbility(t, g, bronto, "trample") {
		t.Fatal("Brontotherium lacks trample")
	}
	other := reqCreature(g, me.ID, "Other Bear", "", 2, 2, "haste")
	victim := reqCreature(g, opp.ID, "Their Bear", "", 2, 2)

	attackAndTrigger(t, g, bronto, other)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	onToBlockers(t, g)
	reqRefusal(t, reqBlock(t, g, other, victim))
	br := reqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); got != "Their Bear must block Brontotherium if able." {
		t.Errorf("sentence = %q", got)
	}
	if err := reqBlock(t, g, bronto, victim); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
}

// --- Hunt Down -------------------------------------------------------------

// TestHuntDownMakesFirstTargetBlockSecond — the first target is the
// would-be blocker, the second is what it must block; order matters.
func TestHuntDownMakesFirstTargetBlockSecond(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seats1684(g)
	hunter := reqCreature(g, me.ID, "Hunter Bear", "", 2, 2, "haste")
	other := reqCreature(g, me.ID, "Other Bear", "", 2, 2, "haste")
	victim := reqCreature(g, opp.ID, "Their Bear", "", 2, 2)

	castCatalogSpell(t, g, "Hunt Down", "Sorcery", huntDownOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}, {Kind: game.TargetCard, ID: hunter}})
	passPriorityAroundTable(t, g)
	recs := blocksAttackerRecords(g)
	if len(recs) != 1 || recs[0].Objects[0].ID != hunter {
		t.Fatalf("records = %+v, want one naming the hunter", recs)
	}

	attackAndTrigger(t, g, hunter, other)
	onToBlockers(t, g)
	reqRefusal(t, reqBlock(t, g, other, victim))
	br := reqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); got != "Their Bear must block Hunter Bear if able (Hunt Down)." {
		t.Errorf("sentence = %q", got)
	}
	if err := reqBlock(t, g, hunter, victim); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
}

// TestHuntDownAgainstAnAttackerThatNeverAttacksAsksNothing — the named
// object never attacked this turn, so the requirement is never judged
// and the would-be blocker is free.
func TestHuntDownAgainstAnAttackerThatNeverAttacksAsksNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seats1684(g)
	hunter := reqCreature(g, me.ID, "Hunter Bear", "", 2, 2, "haste")
	other := reqCreature(g, me.ID, "Other Bear", "", 2, 2, "haste")
	victim := reqCreature(g, opp.ID, "Their Bear", "", 2, 2)
	castCatalogSpell(t, g, "Hunt Down", "Sorcery", huntDownOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}, {Kind: game.TargetCard, ID: hunter}})
	passPriorityAroundTable(t, g)

	// Only "other" attacks — the hunter stays home.
	attackAndTrigger(t, g, other)
	onToBlockers(t, g)
	if owed := mustBlockNow(g); owed != nil {
		t.Errorf("must_block = %v, want nothing owed", owed)
	}
	if err := reqBlock(t, g, other, victim); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
}

// --- Treeshaker Chimera, Breaker of Armies, Ochran Assassin: Lure -----------

// TestTreeshakerChimeraLuresAndDrawsWhenItDies — the Lure clause is
// unfiltered (Prized Unicorn shape), and its dies trigger draws three.
func TestTreeshakerChimeraLuresAndDrawsWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seats1684(g)
	chimera := reqCreature(g, me.ID, "Treeshaker Chimera", treeshakerChimeraOracle, 8, 5, "haste")
	a := reqCreature(g, opp.ID, "Blocker A", "", 3, 4)
	b := reqCreature(g, opp.ID, "Blocker B", "", 3, 4)
	reqToBlockers(t, g, chimera)
	br := reqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); got != "Blocker A must block Treeshaker Chimera if able." {
		t.Errorf("sentence = %q", got)
	}
	hand := len(me.Hand.Cards)
	if err := reqBlock(t, g, chimera, a, b); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	advanceToStepInTurn(t, g, game.StepCombatDamage)
	b26AnswerDamageAssignments(t, g)
	passPriorityAroundTable(t, g)
	if _, onField := battlefieldCard(g, chimera); onField {
		t.Fatal("the Chimera survived 6 damage from two blockers")
	}
	if got := len(me.Hand.Cards) - hand; got != 3 {
		t.Errorf("drew %d cards on death, want 3", got)
	}
}

// TestBreakerOfArmiesLuresEveryAbleBlocker.
func TestBreakerOfArmiesLuresEveryAbleBlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seats1684(g)
	breaker := reqCreature(g, me.ID, "Breaker of Armies", breakerOfArmiesOracle, 10, 8, "haste")
	a := reqCreature(g, opp.ID, "Wall A", "", 0, 4)
	b := reqCreature(g, opp.ID, "Wall B", "", 0, 4)
	reqToBlockers(t, g, breaker)
	br := reqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); got != "Wall A must block Breaker of Armies if able." {
		t.Errorf("sentence = %q", got)
	}
	if err := reqBlock(t, g, breaker, a, b); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
}

// TestOchranAssassinLuresAndHasDeathtouch.
func TestOchranAssassinLuresAndHasDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seats1684(g)
	assassin := reqCreature(g, me.ID, "Ochran Assassin", ochranAssassinOracle, 1, 1, "haste")
	if !hasEffectiveAbility(t, g, assassin, "deathtouch") {
		t.Fatal("Ochran Assassin lacks deathtouch")
	}
	a := reqCreature(g, opp.ID, "Wall A", "", 0, 4)
	b := reqCreature(g, opp.ID, "Wall B", "", 0, 4)
	reqToBlockers(t, g, assassin)
	br := reqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); got != "Wall A must block Ochran Assassin if able." {
		t.Errorf("sentence = %q", got)
	}
	if err := reqBlock(t, g, assassin, a, b); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
}

// --- Nemesis Mask ------------------------------------------------------------

// TestNemesisMaskLuresTheEquippedCreature — Lure's own shape, on an
// Equipment: BlockRequirementWhere(lure, AttachedToSource) plus equip.
func TestNemesisMaskLuresTheEquippedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seats1684(g)
	advanceToMain(t, g)
	mask := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Nemesis Mask", TypeLine: equipTypeLine,
		OracleID: nemesisMaskOracle, Owner: me.ID, Controller: me.ID,
	})
	bear := reqCreature(g, me.ID, "Bear", "", 2, 2, "haste")
	other := reqCreature(g, me.ID, "Other Bear", "", 2, 2, "haste")
	a := reqCreature(g, opp.ID, "Wall A", "", 0, 4)
	b := reqCreature(g, opp.ID, "Wall B", "", 0, 4)
	equipTo(t, g, me.ID, mask, bear)

	reqToBlockers(t, g, bear, other)
	reqRefusal(t, reqBlock(t, g, other, a))
	br := reqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); got != "Wall A must block Bear if able (Nemesis Mask)." {
		t.Errorf("sentence = %q", got)
	}
	if err := reqBlock(t, g, bear, a, b); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
}

// --- Compelled Duel, Emergent Growth, Enlarge: pump + must be blocked ------

// TestCompelledDuelPumpsAndMustBeBlocked.
func TestCompelledDuelPumpsAndMustBeBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seats1684(g)
	bear := reqCreature(g, me.ID, "Bear", "", 2, 2, "haste")
	a := reqCreature(g, opp.ID, "Wall A", "", 0, 4)
	castCatalogSpell(t, g, "Compelled Duel", "Sorcery", compelledDuelOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, bear); got != 5 {
		t.Errorf("power = %d, want 5 (+3/+3)", got)
	}
	if got := effectiveToughness(t, g, bear); got != 5 {
		t.Errorf("toughness = %d, want 5", got)
	}
	reqToBlockers(t, g, bear)
	br := reqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); got != "Bear must be blocked if able (Compelled Duel)." {
		t.Errorf("sentence = %q", got)
	}
	if err := reqBlock(t, g, bear, a); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
}

// TestEmergentGrowthPumpsAndMustBeBlocked.
func TestEmergentGrowthPumpsAndMustBeBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seats1684(g)
	bear := reqCreature(g, me.ID, "Bear", "", 2, 2, "haste")
	a := reqCreature(g, opp.ID, "Wall A", "", 0, 4)
	castCatalogSpell(t, g, "Emergent Growth", "Sorcery", emergentGrowthOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, bear); got != 7 {
		t.Errorf("power = %d, want 7 (+5/+5)", got)
	}
	if got := effectiveToughness(t, g, bear); got != 7 {
		t.Errorf("toughness = %d, want 7", got)
	}
	reqToBlockers(t, g, bear)
	br := reqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); got != "Bear must be blocked if able (Emergent Growth)." {
		t.Errorf("sentence = %q", got)
	}
	if err := reqBlock(t, g, bear, a); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
}

// TestEnlargePumpsGrantsTrampleAndMustBeBlocked.
func TestEnlargePumpsGrantsTrampleAndMustBeBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seats1684(g)
	bear := reqCreature(g, me.ID, "Bear", "", 2, 2, "haste")
	a := reqCreature(g, opp.ID, "Wall A", "", 0, 4)
	castCatalogSpell(t, g, "Enlarge", "Sorcery", enlargeOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, bear); got != 9 {
		t.Errorf("power = %d, want 9 (+7/+7)", got)
	}
	if got := effectiveToughness(t, g, bear); got != 9 {
		t.Errorf("toughness = %d, want 9", got)
	}
	if !hasEffectiveAbility(t, g, bear, "trample") {
		t.Error("Enlarge did not grant trample")
	}
	reqToBlockers(t, g, bear)
	br := reqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); got != "Bear must be blocked if able (Enlarge)." {
		t.Errorf("sentence = %q", got)
	}
	if err := reqBlock(t, g, bear, a); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
}
