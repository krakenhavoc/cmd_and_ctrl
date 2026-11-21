package testsms

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// nowhere_to_run_test.go — #1560. Nowhere to Run and Kaya, Bane of the
// Dead are the proof cards for the two statics in
// game/hexproof_bypass.go: "as though it didn't have hexproof" and
// "ward abilities of those creatures don't trigger".

const (
	nowhereToRunOracle = "da82ca96-a613-4d00-9e6b-1fece4fc23d0"
	kayaBaneOracle     = "29d58aa4-9a7e-4052-9b0d-fc282f1be40e"
)

func pushNowhereToRun(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Nowhere to Run", TypeLine: "Enchantment",
		ManaCost: "{1}{B}", OracleID: nowhereToRunOracle, Owner: owner, Controller: owner,
	})
}

// doomBladeErrAt is copy_grants_test.go's castDoomBladeAt for the
// casts that are supposed to be refused.
func doomBladeErrAt(t *testing.T, g *game.Game, target uuid.UUID) error {
	t.Helper()
	return effectstest.CastCatalogSpellErr(t, g, "Doom Blade", "Instant", effectstest.DoomBladeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: target}})
}

// The printed scope: "creatures your opponents control can be the
// targets of spells and abilities" — ANY player's spells and
// abilities, so the hexproof creature's other opponent may target it
// too. The static's own controller's creatures are not "your
// opponents'" and keep their hexproof against everyone.
func TestNowhereToRunWaivesOpponentsHexproofForEveryone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	theirs := effectstest.PushKeywordCreature(g, opp.ID, "Their Bogle", "hexproof")
	mine := effectstest.PushKeywordCreature(g, me.ID, "My Bogle", "hexproof")

	if effectstest.CreatureTargetableBy(g, me.ID, theirs) || effectstest.CreatureTargetableBy(g, third.ID, theirs) {
		t.Fatal("baseline: an opponent's hexproof creature must not be targetable")
	}

	pushNowhereToRun(g, me.ID)

	if !effectstest.CreatureTargetableBy(g, me.ID, theirs) {
		t.Error("Nowhere to Run's controller should be able to target an opponent's hexproof creature")
	}
	if !effectstest.CreatureTargetableBy(g, third.ID, theirs) {
		t.Error(`"spells and abilities" has no "you control": another opponent may target it too`)
	}
	if effectstest.CreatureTargetableBy(g, opp.ID, mine) || effectstest.CreatureTargetableBy(g, third.ID, mine) {
		t.Error("the controller's own hexproof creature is not an opponent's and keeps its hexproof")
	}

	// Announce and resolve: the CR 608.2b re-check reads the same
	// waiver, so the spell does not fizzle.
	effectstest.CastDoomBladeAt(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the re-check must honour the waiver — Doom Blade should have resolved")
	}
}

// Only hexproof is waived. Shroud stops everyone, and protection is a
// test against the SOURCE (CR 702.16b) that the waiver says nothing
// about — a pro-black hexproof creature is still no target for Doom
// Blade.
func TestNowhereToRunWaivesNeitherShroudNorProtection(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushNowhereToRun(g, me.ID)
	shrouded := effectstest.PushKeywordCreature(g, opp.ID, "Their Ledgewalker", "shroud")
	proBlack := effectstest.PushKeywordCreature(g, opp.ID, "Their Knight", "hexproof", "protection from black")

	if effectstest.CreatureTargetableBy(g, me.ID, shrouded) {
		t.Error("shroud is not hexproof and must not be waived")
	}
	if err := doomBladeErrAt(t, g, shrouded); !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("Doom Blade at a shrouded creature: err = %v, want ErrIllegalTarget", err)
	}
	// A BLACK Doom Blade: protection tests the source's colour, so the
	// spell card has to carry one.
	blade := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: blade, Name: "Doom Blade", TypeLine: "Instant",
		ManaCost: "{1}{B}", Colors: []string{"B"},
		OracleID: effectstest.DoomBladeOracle, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, blade, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: proBlack}},
	}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("Doom Blade at a pro-black creature: err = %v, want ErrIllegalTarget", err)
	}
}

// The first 2024-09-20 ruling: a spell aimed at a hexproof creature
// while Nowhere to Run was out becomes an illegal target when the
// enchantment leaves before it resolves.
func TestNowhereToRunLeavingMakesTheHexproofTargetIllegal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ntr := pushNowhereToRun(g, me.ID)
	theirs := effectstest.PushKeywordCreature(g, opp.ID, "Their Bogle", "hexproof")

	blade := effectstest.CastDoomBladeAt(t, g, theirs)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(ntr) })
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(theirs) {
		t.Error("the enchantment left, so the target became illegal and the spell should have fizzled")
	}
	if !me.Graveyard.Contains(blade) {
		t.Error("the fizzled Doom Blade should be in its owner's graveyard")
	}
}

// Ward abilities of an opponent's creature don't trigger while the
// enchantment is out, and do again once it leaves.
func TestNowhereToRunSuppressesWardUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ntr := pushNowhereToRun(g, me.ID)
	first := effectstest.PushCatalogPermanent(g, opp.ID, "Rimeshield Frost Giant",
		"Creature — Giant Warrior", effectstest.RimeshieldOracle, false)
	second := effectstest.PushCatalogPermanent(g, opp.ID, "Rimeshield Frost Giant",
		"Creature — Giant Warrior", effectstest.RimeshieldOracle, false)

	effectstest.CastDoomBladeAt(t, g, first)
	if n := wardTriggersOnStack(g); n != 0 || effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatalf("ward triggered with Nowhere to Run out (%d on the stack)", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("no ward payment should be asked for while Nowhere to Run is out")
	}
	if g.Battlefield.Contains(first) {
		t.Error("without a ward trigger Doom Blade should have resolved")
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(ntr) })
	effectstest.CastDoomBladeAt(t, g, second)
	if wardTriggersOnStack(g) != 1 {
		t.Fatalf("with Nowhere to Run gone, ward should trigger again (%d on the stack)", wardTriggersOnStack(g))
	}
}

// "Ward abilities of THOSE creatures": the static's controller's own
// warded creature still taxes an opponent.
func TestNowhereToRunLeavesYourOwnWardAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushNowhereToRun(g, me.ID)
	mine := effectstest.PushCatalogPermanent(g, me.ID, "Rimeshield Frost Giant",
		"Creature — Giant Warrior", effectstest.RimeshieldOracle, false)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	effectstest.CastAtWardedCreature(t, g, opp, mine)
	if wardTriggersOnStack(g) != 1 {
		t.Errorf("the Nowhere to Run controller's own creature is not an opponent's — its ward must trigger")
	}
}

// A GRANTED ward is the creature's ward too. Its trigger belongs to
// the Equipment, so the suppression has to be asked of the warded
// creature rather than of the trigger's source — the Boots are not a
// creature and would never match "those creatures".
func TestNowhereToRunSuppressesAGrantedWard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	boots := effectstest.SeedEquipment(g, me.ID, "Lavaspur Boots", effectstest.LavaspurBootsOracle)
	effectstest.EquipTo(t, g, me.ID, boots, bear)
	pushNowhereToRun(g, opp.ID)

	effectstest.CastAtWardedCreature(t, g, opp, bear)
	if n := wardTriggersOnStack(g); n != 0 {
		t.Fatalf("the Boots' ward on an opponent's creature triggered (%d on the stack)", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("with no ward trigger the Doom Blade should have killed the bear")
	}
}

// The enters trigger runs with the static already in force, so it
// reaches a hexproof creature.
func TestNowhereToRunEntersTriggerReachesHexproof(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	theirs := effectstest.PushKeywordCreature(g, opp.ID, "Their Bogle", "hexproof")

	effectstest.CastCatalogSpell(t, g, "Nowhere to Run", "Enchantment", nowhereToRunOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(theirs) {
		t.Error("the 2/2 hexproof creature should have died to -3/-3")
	}
}

// Kaya's waiver is "spells and abilities YOU control", and it has a
// PLAYER half (CR 702.11d).
func TestKayaWaivesHexproofOnlyForHerController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	theirs := effectstest.PushKeywordCreature(g, opp.ID, "Their Bogle", "hexproof")
	opp.Statics = append(opp.Statics, game.PlayerStatic{Keyword: game.KeywordHexproof})

	if effectstest.PlayerTargetableBy(g, me.ID, opp.ID) {
		t.Fatal("baseline: a hexproof player must not be targetable by an opponent")
	}

	kaya := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Kaya, Bane of the Dead",
		TypeLine: "Legendary Planeswalker — Kaya", OracleID: kayaBaneOracle,
		Owner: me.ID, Controller: me.ID, StartingLoyalty: 7,
		Counters: map[string]int{game.CounterLoyalty: 7},
	})

	if !effectstest.CreatureTargetableBy(g, me.ID, theirs) || !effectstest.PlayerTargetableBy(g, me.ID, opp.ID) {
		t.Error("Kaya's controller should ignore an opponent's and an opponent's creature's hexproof")
	}
	if effectstest.CreatureTargetableBy(g, third.ID, theirs) || effectstest.PlayerTargetableBy(g, third.ID, opp.ID) {
		t.Error(`"you control": another opponent must still be refused`)
	}

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, kaya, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
	}); err != nil {
		t.Fatalf("Kaya −3 at a hexproof creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("Kaya's −3 should have exiled the hexproof creature")
	}
}

// wardTriggersOnStack counts ward triggers waiting on or headed for
// the stack.
func wardTriggersOnStack(g *game.Game) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, item := range g.StackMeta {
			if item.Kind != game.StackItemSpell && item.Trigger != nil && item.Trigger.Event.Kind == game.EventBecomesTarget {
				n++
			}
		}
		n += len(g.PendingTriggers)
	})
	return n
}
