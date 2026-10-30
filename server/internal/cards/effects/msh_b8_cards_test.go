package effects

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msh_b8_cards_test.go — Marvel Super Heroes, half B, eighth batch (#4455).
// Each test plays the real catalog entry and asserts what its text says.

const (
	mshB8SheHulkOracle       = "bbd0a2a4-f81b-4a6b-bc4d-d7e53dc51333"
	mshB8SpeedOracle         = "2fb0f2d2-ed1f-4181-91f2-684a71c34998"
	mshB8SpiderManOracle     = "85dec6cf-6f27-4eb0-834b-5f6fbbe25fc8"
	mshB8ExecutiveOracle     = "7c07970d-063a-44a1-9761-b80aafa90dac"
	mshB8StolenTechOracle    = "51023277-5434-4913-b9dc-df2a3fde1652"
	mshB8IntelligenceOracle  = "d82929f2-57ad-45be-a834-984437046026"
	mshB8StrengthOracle      = "e1b577ad-1488-43b5-83d9-e40bd60c74be"
	mshB8LockupOracle        = "192e150a-96b0-42d6-9d39-03d0bed61f41"
	mshB8SuperSkrullOracle   = "4644013f-c96a-450e-a7bc-ddfa7a563050"
	mshB8SurveillanceOracle  = "a3648376-dc8b-409b-b2d1-c29e326a059c"
	mshB8SheHulkTypeLine     = "Legendary Creature — Gamma Hero"
	mshB8SpiderManTypeLine   = "Legendary Creature — Spider Human Hero"
	mshB8SpeedTypeLine       = "Legendary Creature — Mutant Hero"
	mshB8SuperSkrullTypeLine = "Legendary Creature — Skrull Shapeshifter Villain"
)

func TestMshB8EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mshB8SheHulkOracle:      "She-Hulk, Jade Defender",
		mshB8SpeedOracle:        "Speed, Young Avenger",
		mshB8SpiderManOracle:    "Spider-Man, To the Rescue",
		mshB8ExecutiveOracle:    "Stark Industries Executive",
		mshB8StolenTechOracle:   "Stolen Stark Tech",
		mshB8IntelligenceOracle: "Super Intelligence",
		mshB8StrengthOracle:     "Super Strength",
		mshB8LockupOracle:       "Super Villain Lockup",
		mshB8SuperSkrullOracle:  "Super-Skrull",
		mshB8SurveillanceOracle: "Surveillance Room",
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

// --- She-Hulk, Jade Defender --------------------------------------------------

func mshB8SheHulkActivate(t *testing.T, g *game.Game, me *game.Player, she uuid.UUID, mana string, targets []game.TargetRef) error {
	t.Helper()
	return hobB4Activate(t, g, me.ID, she, 0, mana, targets)
}

func TestMshB8SheHulkPowerUpDestroysAnArtifactOrEnchantmentAndGrowsHer(t *testing.T) {
	for _, tc := range []struct{ name, typeLine string }{
		{"Test Rock", "Artifact"},
		{"Test Rune", "Enchantment"},
	} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		powerUpMain(t, g)
		she := pushPowerUpPermanent(g, me.ID, "She-Hulk, Jade Defender", mshB8SheHulkTypeLine, "{3}{G}", mshB8SheHulkOracle)
		target := b12Permanent(g, opp.ID, tc.name, tc.typeLine)

		if got := chargedPowerUp(t, g, me.ID, she); got != "{4}{G}{G}" {
			t.Fatalf("power-up costs %q, want {4}{G}{G}", got)
		}
		if err := mshB8SheHulkActivate(t, g, me, she, "{C}{C}{C}{C}{G}{G}", cardRefs(target)); err != nil {
			t.Fatalf("%s: activate: %v", tc.typeLine, err)
		}
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(target) {
			t.Errorf("%s: the %s survived", tc.typeLine, tc.name)
		}
		if got := b12Counter(t, g, she, game.CounterPlusOne); got != 1 {
			t.Errorf("%s: +1/+1 counters = %d, want 1", tc.typeLine, got)
		}
		if err := mshB8SheHulkActivate(t, g, me, she, "{C}{C}{C}{C}{G}{G}", nil); err == nil {
			t.Errorf("%s: the power-up was activated a second time", tc.typeLine)
		}
	}
}

func TestMshB8SheHulkCanChooseNoTargetAndStillGetsTheCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	she := pushPowerUpPermanent(g, me.ID, "She-Hulk, Jade Defender", mshB8SheHulkTypeLine, "{3}{G}", mshB8SheHulkOracle)

	if err := mshB8SheHulkActivate(t, g, me, she, "{C}{C}{C}{C}{G}{G}", nil); err != nil {
		t.Fatalf("activate with no target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := b12Counter(t, g, she, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
}

func TestMshB8SheHulkCannotTargetACreatureOrLand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	she := pushPowerUpPermanent(g, me.ID, "She-Hulk, Jade Defender", mshB8SheHulkTypeLine, "{3}{G}", mshB8SheHulkOracle)
	bear := b12Creature(g, opp.ID, "Test Bear", "Creature — Bear", 2, 2)
	land := b12Permanent(g, opp.ID, "Test Land", "Land")

	for _, bad := range []uuid.UUID{bear, land} {
		if err := mshB8SheHulkActivate(t, g, me, she, "{C}{C}{C}{C}{G}{G}", cardRefs(bad)); err == nil {
			t.Fatal("She-Hulk targeted a creature or land")
		}
	}
	if got := b12Counter(t, g, she, game.CounterPlusOne); got != 0 {
		t.Errorf("a refused activation left %d counters", got)
	}
}

func TestMshB8SheHulkPowerUpIsCheaperTheTurnShesCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	she := castPowerUpCreature(t, g, me, "She-Hulk, Jade Defender", mshB8SheHulkTypeLine, "{3}{G}", "{C}{C}{C}{G}", mshB8SheHulkOracle, 4)
	if got := chargedPowerUp(t, g, me.ID, she); got != "{1}{G}" {
		t.Errorf("the turn she entered the power-up costs %q, want {1}{G}", got)
	}
}

// --- Speed, Young Avenger -----------------------------------------------------

func mshB8SpeedBoard(t *testing.T, g *game.Game) (speed, rusher uuid.UUID) {
	t.Helper()
	me := g.Seats[0]
	advanceToMain(t, g)
	speed = pushCatalogPermanent(g, me.ID, "Speed, Young Avenger", mshB8SpeedTypeLine, mshB8SpeedOracle, false)
	rusher = brCreature(g, me.ID, "Test Rusher", "Creature — Goblin", 2, 1, "haste")
	return speed, rusher
}

func TestMshB8SpeedPaysOneSoAHasteCreatureOnlyHasteCanBlock(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	_, rusher := mshB8SpeedBoard(t, g)
	slow := b12Creature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	fast := brCreature(g, opp.ID, "Raging Goblin", "Creature — Goblin", 1, 1, "haste")
	b06AddMana(me, "G")

	castCatalogSpell(t, g, "Test Instant", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, true)
	answerPickTarget(t, g, rusher)
	passPriorityAroundTable(t, g)

	if n := scopedBlockRuleCount(g); n != 1 {
		t.Fatalf("%d scoped block rules, want 1", n)
	}
	if len(poolColors(me)) != 0 {
		t.Errorf("pool = %v after paying {1}", poolColors(me))
	}
	brAttack(t, g, rusher)
	offered := brOffered(t, g, opp.ID)
	if brOffers(offered, slow, rusher) || !brOffers(offered, fast, rusher) {
		t.Errorf("the enumerator disagrees with the rule: %v", offered)
	}
	brRefusal(t, g.DeclareBlocker(slow, rusher), game.BlockReasonCantBeBlockedExceptBy)
	if err := g.DeclareBlocker(fast, rusher); err != nil {
		t.Fatalf("a creature with haste blocks: %v", err)
	}
}

func TestMshB8SpeedOnlyOffersCreaturesWithHaste(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	speed, rusher := mshB8SpeedBoard(t, g)
	bear := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	theirs := brCreature(g, opp.ID, "Their Rusher", "Creature — Goblin", 1, 1, "haste")
	b06AddMana(me, "G")

	castCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, true)
	pick := pickTargetPrompt(t, g)
	for _, want := range []uuid.UUID{speed, rusher, theirs} {
		if !hasID(pick.PickTargetCards, want) {
			t.Errorf("a creature with haste was not offered: %v", pick.PickTargetCards)
		}
	}
	if hasID(pick.PickTargetCards, bear) {
		t.Errorf("a creature without haste was offered: %v", pick.PickTargetCards)
	}
}

func TestMshB8SpeedDecliningTheMayPayDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mshB8SpeedBoard(t, g)
	b06AddMana(me, "G")

	castCatalogSpell(t, g, "Test Instant", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)

	if n := scopedBlockRuleCount(g); n != 0 {
		t.Errorf("%d block rules after declining", n)
	}
	if len(poolColors(me)) != 1 {
		t.Errorf("the {1} was spent anyway: pool = %v", poolColors(me))
	}
}

func TestMshB8SpeedIgnoresCreatureSpellsAndOpponentsSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mshB8SpeedBoard(t, g)
	b06AddMana(me, "G")

	castCatalogSpell(t, g, "Test Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		t.Fatal("a creature spell offered the {1}")
	}
	if n := scopedBlockRuleCount(g); n != 0 {
		t.Errorf("%d block rules after a creature spell", n)
	}
}

func TestMshB8SpeedRuleEndsWithTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	_, rusher := mshB8SpeedBoard(t, g)
	b06AddMana(me, "G")

	castCatalogSpell(t, g, "Test Instant", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, true)
	answerPickTarget(t, g, rusher)
	passPriorityAroundTable(t, g)
	advanceToUpkeepOf(t, g, 1)
	if n := scopedBlockRuleCount(g); n != 0 {
		t.Errorf("%d scoped block rules outlived the turn", n)
	}
}

// --- Spider-Man, To the Rescue ------------------------------------------------

func mshB8CastSpiderMan(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	return castPowerUpCreature(t, g, g.Seats[0], "Spider-Man, To the Rescue", mshB8SpiderManTypeLine, "{2}{G}", "{C}{C}{G}", mshB8SpiderManOracle, 3)
}

func TestMshB8SpiderManTapsToGiveAnotherCreatureIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	ally := b12Creature(g, me.ID, "Test Ally", "Creature — Bear", 2, 2)
	spider := mshB8CastSpiderMan(t, g)

	answerLatestTriggerPrompt(t, g, me.ID, true)
	mshB2PickIfAsked(t, g, me.ID, ally)
	passPriorityAroundTable(t, g)

	if !ringBFCard(t, g, spider).Tapped {
		t.Error("Spider-Man was not tapped")
	}
	if !hasEffectiveKeyword(t, g, ally, "indestructible") {
		t.Error("the other creature did not gain indestructible")
	}
	if hasEffectiveKeyword(t, g, spider, "indestructible") {
		t.Error("Spider-Man gained indestructible: it says another creature")
	}
	advanceToNextSeatsTurn(t, g)
	if hasEffectiveKeyword(t, g, ally, "indestructible") {
		t.Error("indestructible outlived the turn")
	}
}

func TestMshB8SpiderManDecliningToTapDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	ally := b12Creature(g, me.ID, "Test Ally", "Creature — Bear", 2, 2)
	spider := mshB8CastSpiderMan(t, g)

	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)

	if ringBFCard(t, g, spider).Tapped {
		t.Error("Spider-Man tapped after declining")
	}
	if latestPickTarget(g, me.ID) != nil || hasEffectiveKeyword(t, g, ally, "indestructible") {
		t.Error("declining still gave indestructible")
	}
}

func TestMshB8SpiderManNeverTargetsAnAttackerOrAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	attacker := brCreature(g, me.ID, "Test Attacker", "Creature — Goblin", 2, 1, "haste")
	idle := b12Creature(g, me.ID, "Idle Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, attacker).AttackingTarget = opp.ID })
	spider := mshB8CastSpiderMan(t, g)

	answerLatestTriggerPrompt(t, g, me.ID, true)
	mshB2PickIfAsked(t, g, me.ID, idle)
	passPriorityAroundTable(t, g)

	if !hasEffectiveKeyword(t, g, idle, "indestructible") {
		t.Error("the nonattacking creature did not gain indestructible")
	}
	for name, id := range map[string]uuid.UUID{"attacker": attacker, "opponent's creature": theirs, "Spider-Man": spider} {
		if hasEffectiveKeyword(t, g, id, "indestructible") {
			t.Errorf("the %s gained indestructible", name)
		}
	}
}

// --- Stark Industries Executive -----------------------------------------------

func TestMshB8ExecutiveMakesATreasureForTwoAndATap(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	exec := pushCatalogPermanent(g, me.ID, "Stark Industries Executive", "Creature — Human Advisor", mshB8ExecutiveOracle, false)

	if err := hobB4Activate(t, g, me.ID, exec, 0, "{C}{C}", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !ringBFCard(t, g, exec).Tapped {
		t.Error("the {T} cost did not tap the Executive")
	}
	passPriorityAroundTable(t, g)
	if got := mshB1Tokens(g, me.ID, "Treasure"); got != 1 {
		t.Errorf("%d Treasures, want 1", got)
	}
	if err := hobB4Activate(t, g, me.ID, exec, 0, "{C}{C}", nil); err == nil {
		t.Error("activated while tapped")
	}
}

func TestMshB8ExecutiveIsRefusedWithoutTwoMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	exec := pushCatalogPermanent(g, me.ID, "Stark Industries Executive", "Creature — Human Advisor", mshB8ExecutiveOracle, false)
	if err := hobB4Activate(t, g, me.ID, exec, 0, "{C}", nil); err == nil {
		t.Error("activated for one mana")
	}
	if got := mshB1Tokens(g, me.ID, "Treasure"); got != 0 {
		t.Errorf("%d Treasures from a refused activation", got)
	}
}

// --- Stolen Stark Tech --------------------------------------------------------

func TestMshB8StolenTechAttachesOnEntryAndGivesIndestructibleThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	bear := seedBear(g, me.ID)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	tech := castPowerUpCreature(t, g, me, "Stolen Stark Tech", "Artifact — Equipment", "{1}{B}", "{C}{B}", mshB8StolenTechOracle, 0)

	b04WaitForPick(t, g, me.ID)
	if p := latestPickTarget(g, me.ID); hasID(p.PickTargetCards, theirs) || !hasID(p.PickTargetCards, bear) {
		t.Fatalf("the entry trigger targets a creature you control: %+v", p)
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)

	if host := attachmentHostOf(t, g, tech); host.ID != bear {
		t.Fatalf("the Equipment did not attach to the Bear: %+v", host)
	}
	if p, tough := effectivePTOf(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("equipped P/T %d/%d, want 3/2", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("the creature did not gain indestructible this turn")
	}
	advanceToNextSeatsTurn(t, g)
	if hasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("indestructible outlived the turn")
	}
	if p, _ := effectivePTOf(t, g, bear); p != 3 {
		t.Errorf("the +1/+0 ended with the turn: power %d", p)
	}
}

func TestMshB8StolenTechEquipCostsOneAndMovesTheBonus(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	first := seedBear(g, me.ID)
	second := seedBear(g, me.ID)
	tech := b12Push(g, me.ID, "Stolen Stark Tech", "Artifact — Equipment", mshB8StolenTechOracle, 0, 0)

	if err := hobB4Activate(t, g, me.ID, tech, 0, "", cardRefs(first)); err == nil {
		t.Fatal("equipped for free")
	}
	floatMana(t, g, me, "{C}")
	equipTo(t, g, me.ID, tech, first)
	if p, _ := effectivePTOf(t, g, first); p != 3 {
		t.Errorf("equipped power %d, want 3", p)
	}
	if hasEffectiveKeyword(t, g, first, "indestructible") {
		t.Error("equipping grants no indestructible: only the entry trigger does")
	}
	floatMana(t, g, me, "{C}")
	equipTo(t, g, me.ID, tech, second)
	if p, _ := effectivePTOf(t, g, first); p != 2 {
		t.Errorf("the old host kept the bonus: power %d", p)
	}
}

// --- Super Intelligence -------------------------------------------------------

func mshB8CastAura(t *testing.T, g *game.Game, name, oracle string, host uuid.UUID) uuid.UUID {
	t.Helper()
	aura := castCatalogSpell(t, g, name, "Enchantment — Aura", oracle, cardRefs(host))
	passPriorityAroundTable(t, g)
	if h := attachmentHostOf(t, g, aura); h.ID != host {
		t.Fatalf("%s did not attach to its host: %+v", name, h)
	}
	return aura
}

func TestMshB8SuperIntelligenceDrawsForTheEnchantedCreaturesControllerEachUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	seedLibrary(me, "Mine A", "Mine B")
	seedLibrary(opp, "Theirs A", "Theirs B", "Theirs C")
	mshB8CastAura(t, g, "Super Intelligence", mshB8IntelligenceOracle, theirs)

	oppHand, myHand := opp.Hand.Size(), me.Hand.Size()
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if got := opp.Hand.Size(); got != oppHand+1 {
		t.Errorf("the creature's controller drew %d, want 1 in their upkeep", got-oppHand)
	}
	if got := me.Hand.Size(); got != myHand {
		t.Errorf("the Aura's controller drew %d, want 0", got-myHand)
	}
}

func TestMshB8SuperIntelligenceOnYourOwnCreatureDrawsInYourUpkeepOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	seedLibrary(me, "Mine A", "Mine B", "Mine C")
	seedLibrary(opp, "Theirs A", "Theirs B")
	mshB8CastAura(t, g, "Super Intelligence", mshB8IntelligenceOracle, mine)

	oppHand := opp.Hand.Size()
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if got := opp.Hand.Size(); got != oppHand {
		t.Errorf("an opponent's upkeep drew %d cards for them", got-oppHand)
	}
	myHand := me.Hand.Size()
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != myHand+1 {
		t.Errorf("my upkeep drew %d, want 1", got-myHand)
	}
}

// --- Super Strength -----------------------------------------------------------

func TestMshB8SuperStrengthGivesPlusFourTrampleAndWardOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	mshB8CastAura(t, g, "Super Strength", mshB8StrengthOracle, bear)

	if p, tough := effectivePTOf(t, g, bear); p != 6 || tough != 6 {
		t.Errorf("enchanted P/T %d/%d, want 6/6", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the enchanted creature has trample")
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	blade := castAtWardedCreature(t, g, opp, bear)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward payment prompt for the spell's controller")
	}
	answerPayUnless(t, g, opp.ID, false)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("a declined ward counters the spell: the Bear should live")
	}
	if !opp.Graveyard.Contains(blade) {
		t.Error("the countered spell is in its owner's graveyard")
	}
}

func TestMshB8SuperStrengthEndsWhenTheAuraLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	aura := mshB8CastAura(t, g, "Super Strength", mshB8StrengthOracle, bear)
	exileCreature(t, g, aura)
	if p, tough := effectivePTOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("after the Aura left: %d/%d, want 2/2", p, tough)
	}
	if hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("trample outlived the Aura")
	}
}

// --- Super Villain Lockup -----------------------------------------------------

func TestMshB8LockupExilesATappedOpposingCreatureUntilItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	tapped := b12Creature(g, opp.ID, "Tapped Bear", "Creature — Bear", 2, 2)
	untapped := b12Creature(g, opp.ID, "Untapped Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "My Tapped Bear", "Creature — Bear", 2, 2)
	tapForTest(t, g, tapped, mine)

	lockup := castPowerUpCreature(t, g, me, "Super Villain Lockup", "Enchantment", "{1}{W}", "{C}{W}", mshB8LockupOracle, 0)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if !hasID(p.PickTargetCards, tapped) || hasID(p.PickTargetCards, untapped) || hasID(p.PickTargetCards, mine) {
		t.Fatalf("targets must be tapped creatures an opponent controls: %v", p.PickTargetCards)
	}
	pickCard(t, g, me.ID, tapped)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(tapped) {
		t.Fatal("the tapped creature was not exiled")
	}
	exileCreature(t, g, lockup)
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Tapped Bear") == uuid.Nil {
		t.Error("the exiled creature did not return when the enchantment left")
	}
}

func TestMshB8LockupWithoutATappedTargetDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	untapped := b12Creature(g, opp.ID, "Untapped Bear", "Creature — Bear", 2, 2)
	castPowerUpCreature(t, g, me, "Super Villain Lockup", "Enchantment", "{1}{W}", "{C}{W}", mshB8LockupOracle, 0)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("offered a target with no tapped creature")
	}
	if !g.Battlefield.Contains(untapped) {
		t.Error("an untapped creature was exiled")
	}
}

// --- Super-Skrull -------------------------------------------------------------

func mshB8PushSkrull(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b12Push(g, owner, "Super-Skrull", mshB8SuperSkrullTypeLine, mshB8SuperSkrullOracle, 4, 5)
}

func TestMshB8SuperSkrullMakesAZeroFourWallWithDefender(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	skrull := mshB8PushSkrull(g, me.ID)

	if err := hobB4Activate(t, g, me.ID, skrull, 0, "{C}{C}{W}", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	wall := mshB1Token(t, g, me.ID, "Wall")
	if wall.Power != 0 || wall.Toughness != 4 || len(wall.Colors) != 0 || !hasKeywordForTest(wall.Keywords, "defender") {
		t.Errorf("the Wall is %d/%d colors %v keywords %v, want a colorless 0/4 defender", wall.Power, wall.Toughness, wall.Colors, wall.Keywords)
	}
}

func TestMshB8SuperSkrullPumpsItselfFourFourUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	skrull := mshB8PushSkrull(g, me.ID)

	if err := hobB4Activate(t, g, me.ID, skrull, 1, "{C}{C}{C}{G}", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, skrull); p != 8 || tough != 9 {
		t.Errorf("pumped P/T %d/%d, want 8/9", p, tough)
	}
	advanceToNextSeatsTurn(t, g)
	if p, tough := effectivePTOf(t, g, skrull); p != 4 || tough != 5 {
		t.Errorf("after the turn P/T %d/%d, want 4/5", p, tough)
	}
}

func TestMshB8SuperSkrullDealsFourDamageToTargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	skrull := mshB8PushSkrull(g, me.ID)
	small := b12Creature(g, opp.ID, "Small Bear", "Creature — Bear", 4, 4)
	big := b12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 5, 5)

	if err := hobB4Activate(t, g, me.ID, skrull, 2, "{C}{C}{C}{C}{R}", cardRefs(small)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the 4/4 survived 4 damage")
	}
	if err := hobB4Activate(t, g, me.ID, skrull, 2, "{C}{C}{C}{C}{R}", cardRefs(big)); err != nil {
		t.Fatalf("activate again: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(big) {
		t.Error("the 5/5 died to 4 damage")
	}
	if err := hobB4Activate(t, g, me.ID, skrull, 2, "{C}{C}{C}{C}{R}", cardRefs(opp.ID)); err == nil {
		t.Error("the damage ability targeted a player")
	}
}

func TestMshB8SuperSkrullTargetPlayerDrawsFour(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	skrull := mshB8PushSkrull(g, me.ID)
	seedLibrary(opp, "A", "B", "C", "D", "E", "F")
	hand := opp.Hand.Size()

	if err := hobB4Activate(t, g, me.ID, skrull, 3, "{C}{C}{C}{C}{C}{U}", []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := opp.Hand.Size(); got != hand+4 {
		t.Errorf("target player's hand grew by %d, want 4", got-hand)
	}
}

func TestMshB8SuperSkrullAbilitiesNeedTheirColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	skrull := mshB8PushSkrull(g, me.ID)
	for idx, mana := range map[int]string{0: "{C}{C}{C}", 1: "{C}{C}{C}{C}"} {
		if err := hobB4Activate(t, g, me.ID, skrull, idx, mana, nil); err == nil {
			t.Errorf("ability %d activated without its colored mana", idx)
		}
	}
}

// --- Surveillance Room --------------------------------------------------------

func TestMshB8SurveillanceRoomEntersUntappedAndSurveils(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Top", "Next")

	id := playLandFromHand(t, g, "Surveillance Room", mshB8SurveillanceOracle)
	if c, _ := battlefieldCard(g, id); c.Tapped {
		t.Error("Surveillance Room entered tapped")
	}
	if triggerOnStack(g, id) == nil {
		t.Fatal("the surveil trigger did not go on the stack")
	}
	passPriorityAroundTable(t, g)
	if surveilChoiceFor(g, me.ID) == nil {
		t.Error("resolving the trigger queued no surveil prompt")
	}
}

func TestMshB8SurveillanceRoomMakesColorlessOrFiltersForOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := seedPermanentWithOracle(g, me.ID, "Surveillance Room", "Land", mshB8SurveillanceOracle)

	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("{T}: Add {C}: %v", err)
	}
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"C"}) {
		t.Errorf("pool = %v, want {C}", got)
	}

	other := seedPermanentWithOracle(g, me.ID, "Surveillance Room", "Land", mshB8SurveillanceOracle)
	me.ManaPool = nil
	if err := g.ActivateManaAbility(me.ID, other, 1, game.ManaAbilityParams{Colors: []string{"U"}}); err == nil {
		t.Error("the filter ability worked without {1}")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateManaAbility(me.ID, other, 1, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("{1},{T}: any colour: %v", err)
	}
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"R"}) {
		t.Errorf("pool after the filter = %v, want {R} (the {C} paid for it)", got)
	}
}
