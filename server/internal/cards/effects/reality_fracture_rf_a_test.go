package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_rf_a_test.go — Reality Fracture slice rf-a (#2795,
// #2832): Pyre Rhymer // Molten Tide (#3504, a controller-only delayed
// mana trigger) and Loot, the Anomaly (#3505, negative power assigns
// combat damage as positive), plus Tarmogoyf's review.

const (
	rfaPyreRhymer = "8940f503-fc6d-4fae-be74-22c660cae1a7"
	rfaLoot       = "e0142e7d-7b0b-4290-a56c-5fa950fdccf8"
)

func rfaPyreRhymerCard(owner uuid.UUID) game.Card {
	return preparationCard(owner, rfaPyreRhymer, []string{"Prowess", "Prepared"},
		game.Face{Name: "Pyre Rhymer", TypeLine: "Creature — Elemental Sorcerer", ManaCost: "{1}{R}{R}", Colors: []string{"R"}, Power: 3, Toughness: 3},
		game.Face{Name: "Molten Tide", TypeLine: "Instant", ManaCost: "{R}", Colors: []string{"R"}})
}

// Pyre Rhymer enters prepared with prowess; casting its prepared copy
// sets up the turn's trigger, which adds {R} for YOUR Mountains only.
func TestPyreRhymerMoltenTideAddsForYourMountainsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceTo(t, g, game.StepPrecombatMain)
	rhymer := rfaPyreRhymerCard(me.ID)
	castFromHandAndResolve(t, g, me, rhymer)
	if !g.IsPreparedForEffect(rhymer.InstanceID) {
		t.Fatal("Pyre Rhymer did not enter prepared")
	}
	has := false
	for _, a := range effectiveAbilities(t, g, rhymer.InstanceID) {
		has = has || a == "prowess"
	}
	if !has {
		t.Error("Pyre Rhymer has no prowess")
	}
	if p, tough := effectivePower(t, g, rhymer.InstanceID), effectiveToughness(t, g, rhymer.InstanceID); p != 3 || tough != 3 {
		t.Errorf("Pyre Rhymer is %d/%d, want 3/3", p, tough)
	}

	mine := b12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	theirs := b12Permanent(g, opp.ID, "Mountain", "Basic Land — Mountain")
	plains := b12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	// Before the spell: nothing extra for anyone.
	if err := g.ActivateManaAbility(opp.ID, theirs, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if got := poolCount(opp, "R"); got != 1 {
		t.Fatalf("their {R} before Molten Tide = %d, want 1", got)
	}
	g.WithWriteLock(func() { opp.ManaPool = nil })
	theirs = b12Permanent(g, opp.ID, "Mountain", "Basic Land — Mountain")

	fraBCastCopy(t, g, me, "Molten Tide")
	if g.IsPreparedForEffect(rhymer.InstanceID) {
		t.Error("casting the copy did not unprepare Pyre Rhymer")
	}
	g.WithWriteLock(func() { me.ManaPool = nil })

	if err := g.ActivateManaAbility(me.ID, mine, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if got := poolCount(me, "R"); got != 2 {
		t.Errorf("my {R} after tapping my Mountain = %d, want 2", got)
	}
	if err := g.ActivateManaAbility(opp.ID, theirs, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if got := poolCount(opp, "R"); got != 1 {
		t.Errorf("their {R} after tapping their Mountain = %d, want 1: Molten Tide says \"you\"", got)
	}
	if err := g.ActivateManaAbility(me.ID, plains, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if got := poolCount(me, "R"); got != 2 {
		t.Errorf("a Plains changed my {R} to %d", got)
	}
	if len(g.PendingTriggers) != 0 || len(g.StackMeta) != 0 {
		t.Error("the triggered mana ability used the stack")
	}

	advanceToUpkeepOf(t, g, 1)
	again := b12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	g.WithWriteLock(func() { me.ManaPool = nil })
	if err := g.ActivateManaAbility(me.ID, again, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if got := poolCount(me, "R"); got != 1 {
		t.Errorf("next turn's {R} = %d, want 1: it ended at cleanup", got)
	}
}

// --- Loot, the Anomaly -------------------------------------------------

func rfaLootOnBoard(g *game.Game, owner uuid.UUID, power int) uuid.UUID {
	return b12Push(g, owner, "Loot, the Anomaly", "Legendary Creature — Beast Horror", rfaLoot, power, 4)
}

// A negative-power Loot assigns damage equal to the absolute value of
// his power; every other creature keeps CR 510.1a's clamp.
func TestLootAssignsCombatDamageAsThoughPowerWerePositive(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	plain := pushVanillaCreature(g, me.ID, "Negative Bear", -3, 4)
	before := opp.Life
	attackWith(t, g, opp.ID, plain)
	passPriorityAroundTable(t, g)
	if opp.Life != before {
		t.Fatalf("a -3/4 creature that is not Loot dealt %d, want 0 (CR 510.1a)", before-opp.Life)
	}

	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	loot2 := rfaLootOnBoard(g2, me2.ID, -2)
	before = opp2.Life
	attackWith(t, g2, opp2.ID, loot2)
	passPriorityAroundTable(t, g2)
	if want := before - 2; opp2.Life != want {
		t.Errorf("opponent at %d, want %d (Loot is -2/4 and deals 2)", opp2.Life, want)
	}
}

// Both in one combat: Loot's two and the plain creature's none.
func TestLootAndAPlainNegativeCreatureAttackTogether(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	loot := rfaLootOnBoard(g, me.ID, -3)
	plain := pushVanillaCreature(g, me.ID, "Negative Bear", -3, 4)
	before := opp.Life
	attackWith(t, g, opp.ID, loot, plain)
	passPriorityAroundTable(t, g)
	if want := before - 3; opp.Life != want {
		t.Errorf("opponent at %d, want %d (3 from Loot, 0 from the other)", opp.Life, want)
	}
}

// A blocking Loot deals the same amount, and a positive power is read as
// it always was.
func TestLootBlocksForTheAbsoluteValueOfHisPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	loot := rfaLootOnBoard(g, me.ID, -2)
	advanceToNextSeatsTurn(t, g)
	raider := pushVanillaCreature(g, opp.ID, "Raider", 1, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(raider, me.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(loot, raider); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(raider) {
		t.Error("Loot blocks for 2 (his power as though positive), enough to kill a 1/2")
	}
	if !g.Battlefield.Contains(loot) {
		t.Error("Loot took one damage and should survive")
	}
}

func TestLootWithPositivePowerDealsThatPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	loot := rfaLootOnBoard(g, me.ID, 3)
	before := opp.Life
	attackWith(t, g, opp.ID, loot)
	passPriorityAroundTable(t, g)
	if want := before - 3; opp.Life != want {
		t.Errorf("opponent at %d, want %d", opp.Life, want)
	}
}

func rfaFillGraveyard(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		pushGraveyardCardForTest(p, "Filler")
	}
}

// The threshold ability: sacrifice another creature or planeswalker,
// Loot gets -2/-0, and the combat that follows deals the larger number.
func TestLootThresholdSacrificeMakesHimHitHarder(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	loot := rfaLootOnBoard(g, me.ID, -2)
	fodder := pushVanillaCreature(g, me.ID, "Fodder", 1, 1)
	rfaFillGraveyard(me, 7)
	advanceToMain(t, g)

	b16Activate(t, g, me.ID, loot, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}})
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	if p := effectivePower(t, g, loot); p != -4 {
		t.Fatalf("Loot's power = %d, want -4 after -2/-0", p)
	}
	before := opp.Life
	attackWith(t, g, opp.ID, loot)
	passPriorityAroundTable(t, g)
	if want := before - 4; opp.Life != want {
		t.Errorf("opponent at %d, want %d (Loot is -4/4 and deals 4)", opp.Life, want)
	}
}

// Below seven cards the ability is refused and nothing is paid; Loot
// cannot sacrifice himself ("another").
func TestLootThresholdIsGatedAndSaysAnother(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	loot := rfaLootOnBoard(g, me.ID, -2)
	fodder := pushVanillaCreature(g, me.ID, "Fodder", 1, 1)
	rfaFillGraveyard(me, 6)
	advanceToMain(t, g)

	err := g.ActivateCatalogAbility(me.ID, loot, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}})
	if !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("six cards in the graveyard: err = %v, want ErrConditionNotMet", err)
	}
	if !g.Battlefield.Contains(fodder) {
		t.Error("a refused activation still sacrificed the creature")
	}

	rfaFillGraveyard(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, loot, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{loot}}); err == nil {
		t.Error("Loot sacrificed himself to his own ability")
	}
	if !g.Battlefield.Contains(loot) {
		t.Error("Loot left the battlefield")
	}
}

// --- Tarmogoyf ----------------------------------------------------------

// A split card in a graveyard has the types of both halves (CR 709.4c)
// and an adventurer only its creature face, so Tarmogoyf counts what
// each really is there. Kindred is a card type, and a supertype
// (Legendary, Basic, Snow) is not.
func TestTarmogoyfCountsTypesOfSplitAndKindredCardsInGraveyards(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[0].ID
	goyf := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tarmogoyf", TypeLine: "Creature — Lhurgoyf",
		OracleID: tarmogoyfOracle, Owner: owner, Controller: owner,
	})

	split := game.Card{
		InstanceID: uuid.New(), Name: "Fire // Ice", Owner: owner, Controller: owner,
		Layout: game.LayoutSplit,
		Faces: []game.Face{
			{Name: "Fire", TypeLine: "Instant", ManaCost: "{1}{R}"},
			{Name: "Ice", TypeLine: "Sorcery", ManaCost: "{1}{U}"},
		},
	}
	split.SettleImported()
	g.WithWriteLock(func() {
		g.Seats[0].Graveyard.PushTop(split)
		g.BumpLayerVersionForTest()
	})
	if got := effectivePower(t, g, goyf); got != 2 {
		t.Errorf("after a split Instant // Sorcery: power %d, want 2", got)
	}

	pushGraveyardCardWithTypeLine(g, owner, "Goblin Slash", "Kindred Instant — Goblin")
	if got := effectivePower(t, g, goyf); got != 3 {
		t.Errorf("after a Kindred Instant: power %d, want 3 (Instant, Sorcery, Kindred)", got)
	}
	pushGraveyardCardWithTypeLine(g, owner, "Snow Dude", "Legendary Snow Creature — Elf")
	if got, want := effectivePower(t, g, goyf), 4; got != want {
		t.Errorf("after a Legendary Snow Creature: power %d, want %d (only Creature is new)", got, want)
	}
	if got, want := effectiveToughness(t, g, goyf), 5; got != want {
		t.Errorf("toughness %d, want %d", got, want)
	}

	// A modal double-faced card is its front face alone off the stack and
	// battlefield: an Artifact // Planeswalker adds Artifact, not both.
	mdfc := game.Card{
		InstanceID: uuid.New(), Name: "Front // Back", Owner: owner, Controller: owner,
		Layout: game.LayoutModalDFC,
		Faces: []game.Face{
			{Name: "Front", TypeLine: "Artifact", ManaCost: "{2}"},
			{Name: "Back", TypeLine: "Planeswalker — Test", ManaCost: "{2}"},
		},
	}
	mdfc.SettleImported()
	g.WithWriteLock(func() {
		g.Seats[0].Graveyard.PushTop(mdfc)
		g.BumpLayerVersionForTest()
	})
	if got, want := effectivePower(t, g, goyf), 5; got != want {
		t.Errorf("after an Artifact // Planeswalker MDFC: power %d, want %d (front face only)", got, want)
	}
}

// The count follows a card LEAVING a graveyard too (exiled for value),
// through the engine's own zone move rather than a hand-bumped version.
func TestTarmogoyfShrinksWhenTheLastOfATypeLeavesAGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	goyf := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tarmogoyf", TypeLine: "Creature — Lhurgoyf",
		OracleID: tarmogoyfOracle, Owner: me.ID, Controller: me.ID,
	})
	pushGraveyardCardWithTypeLine(g, me.ID, "Grizzly Bears", "Creature — Bear")
	land := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
		g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: land, OldZone: game.ZoneHand, NewZone: game.ZoneGraveyard})
	})
	if got := effectivePower(t, g, goyf); got != 2 {
		t.Fatalf("power %d, want 2 (Creature, Land)", got)
	}
	g.WithWriteLock(func() {
		_, _ = me.Graveyard.Remove(land)
		g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: land, OldZone: game.ZoneGraveyard, NewZone: game.ZoneExile})
	})
	if got := effectivePower(t, g, goyf); got != 1 {
		t.Errorf("power %d after the Land left, want 1", got)
	}
}
