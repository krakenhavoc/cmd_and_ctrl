package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attack_this_permanent_test.go — #2567: creatures that must attack one
// particular permanent during one particular player's next turn. Gideon
// Jura's +2 (one opponent's creatures, read live), Kytheon, Hero of
// Akros's flip, and Gideon, Battle-Forged's +2 (one creature, "its
// controller" read live). The requirement's own rules are pinned in
// game/attack_permanent_requirement_test.go.

const (
	gideonJuraOracle = "7340dbcb-4b26-4038-be8c-a89e76294221"
	kytheonOracle    = "e49d745c-5130-4b54-bcbb-643b08528456"
)

// attackThisRecords counts the live attack-this-permanent records.
func attackThisRecords(g *game.Game) int {
	n := 0
	for _, e := range g.ScopedEffects {
		for _, m := range e.Mods {
			if m.Kind == game.ModAttackPermanentIfAble {
				n++
			}
		}
	}
	return n
}

// leaveDeclareAttackers ends the active player's declaration through
// AdvanceStep, which runs CR 508.1d's checkpoint.
func leaveDeclareAttackers(g *game.Game) error {
	_, err := g.AdvanceStep()
	return err
}

// gideonJuraPlusTwo seats Gideon Jura for seat 0 in its precombat main
// phase and activates the +2 at `target`.
func gideonJuraPlusTwo(t *testing.T, g *game.Game, target uuid.UUID) uuid.UUID {
	t.Helper()
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	gideon := pushGideon(g, me.ID, "Gideon Jura", gideonJuraOracle, 6)
	b16Activate(t, g, me.ID, gideon, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: target}},
	})
	if got := loyaltyCount(g, gideon); got != 8 {
		t.Fatalf("loyalty after +2 = %d, want 8", got)
	}
	if attackThisRecords(g) != 1 {
		t.Fatal("the +2 registered no requirement")
	}
	return gideon
}

// In a four-player game the +2 at the opponent two seats on binds only
// that opponent's creatures, only in their next turn: the opponent whose
// turn comes first owes nothing, the target's creature — and one they
// gained after the ability resolved (CR 611.2c) — must attack Gideon
// himself (attacking his controller instead is refused), and the next
// turn after theirs asks nothing again.
func TestGideonJuraPlusTwoBindsOnlyThatOpponentsCreaturesInTheirNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, first, target, last := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	firstBear := pushVanillaCreature(g, first.ID, "First Bear", 2, 2)
	targetBear := pushVanillaCreature(g, target.ID, "Target Bear", 2, 2)
	pushVanillaCreature(g, last.ID, "Last Bear", 2, 2)
	gideon := gideonJuraPlusTwo(t, g, target.ID)
	late := pushVanillaCreature(g, target.ID, "Late Bear", 2, 2)

	// The opponent in between: nothing is owed, so the declaration can
	// end with every creature home. Under the whole-opposition scope this
	// was refused (stronger than printed).
	advanceToDeclareAttackersOf(t, g, 1)
	if err := leaveDeclareAttackers(g); err != nil {
		t.Fatalf("the first opponent's empty declaration: %v, want none owed", err)
	}
	_ = firstBear

	advanceToDeclareAttackersOf(t, g, 2)
	var re *game.AttackRequirementError
	if err := leaveDeclareAttackers(g); !errors.As(err, &re) {
		t.Fatalf("the target's empty declaration = %v, want an attack requirement refusal", err)
	}
	if got := re.Sentence(target.ID); got != re.AttackerName+" must attack Gideon Jura if able." {
		t.Errorf("sentence = %q", got)
	}
	if err := g.DeclareAttacker(targetBear, me.ID); !errors.Is(err, game.ErrAttackRequirement) {
		t.Fatalf("attacking Gideon's controller instead = %v, want ErrAttackRequirement", err)
	}
	if err := g.DeclareAttacker(targetBear, gideon); err != nil {
		t.Fatalf("attack Gideon: %v", err)
	}
	if err := leaveDeclareAttackers(g); !errors.Is(err, game.ErrAttackRequirement) {
		t.Fatalf("the late Bear home = %v, want ErrAttackRequirement (the set is read live)", err)
	}
	if err := g.DeclareAttacker(late, gideon); err != nil {
		t.Fatalf("attack Gideon with the late Bear: %v", err)
	}
	if err := leaveDeclareAttackers(g); err != nil {
		t.Fatalf("both Bears at Gideon: %v", err)
	}

	advanceToDeclareAttackersOf(t, g, 3)
	if err := leaveDeclareAttackers(g); err != nil {
		t.Fatalf("the last opponent's empty declaration: %v", err)
	}
	if n := attackThisRecords(g); n != 0 {
		t.Errorf("%d requirement records outlived the target's next turn", n)
	}
	advanceToDeclareAttackersOf(t, g, 2)
	if err := leaveDeclareAttackers(g); err != nil {
		t.Fatalf("the target's turn after next: %v, want nothing owed", err)
	}
}

// A creature that cannot attack Gideon owes nothing: with Gideon gone
// before the target's turn, its creatures may stay home (the ruling).
func TestGideonJuraPlusTwoOwesNothingOnceHeIsGone(t *testing.T) {
	g := newCatalogGame(t)
	target := g.Seats[1]
	pushVanillaCreature(g, target.ID, "Target Bear", 2, 2)
	gideon := gideonJuraPlusTwo(t, g, target.ID)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(gideon); err != nil {
			t.Fatalf("destroy Gideon: %v", err)
		}
	})
	advanceToDeclareAttackersOf(t, g, 1)
	if err := leaveDeclareAttackers(g); err != nil {
		t.Fatalf("with Gideon gone: %v, want nothing owed", err)
	}
}

// The live requirement is data: it survives a strict restore point, and
// the restored table still refuses the target's empty declaration.
func TestGideonJuraRequirementSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	target := g.Seats[1]
	bear := pushVanillaCreature(g, target.ID, "Target Bear", 2, 2)
	gideon := gideonJuraPlusTwo(t, g, target.ID)

	r := restoreRoundTrip(t, g, true)
	if attackThisRecords(r) != 1 {
		t.Fatal("the requirement did not survive the restore")
	}
	advanceToDeclareAttackersOf(t, r, 1)
	if err := leaveDeclareAttackers(r); !errors.Is(err, game.ErrAttackRequirement) {
		t.Fatalf("restored empty declaration = %v, want ErrAttackRequirement", err)
	}
	if err := r.DeclareAttacker(bear, gideon); err != nil {
		t.Fatalf("attack Gideon after the restore: %v", err)
	}
	if err := leaveDeclareAttackers(r); err != nil {
		t.Fatalf("declaration after the restore: %v", err)
	}
}

// kytheonCard is the double-faced card as the deck importer builds it,
// front face up.
func kytheonCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   kytheonOracle,
		Layout:     game.LayoutTransform,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{Name: "Kytheon, Hero of Akros", TypeLine: "Legendary Creature — Human Soldier", ManaCost: "{W}",
				Colors: []string{"W"}, Power: 2, Toughness: 1},
			{Name: "Gideon, Battle-Forged", TypeLine: "Legendary Planeswalker — Gideon",
				Colors: []string{"W"}, StartingLoyalty: 3},
		},
	}
	c.SetFace(0)
	return c
}

// battleForged finds the transformed Kytheon, or nil.
func battleForged(g *game.Game) *game.Card {
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.Name == "Gideon, Battle-Forged" {
			return c
		}
	}
	return nil
}

// kytheonCombat seats Kytheon and `others` other creatures for seat 0,
// attacks seat 1 with all of them, runs `mid` in the declare blockers
// step (after the declaration), and resolves whatever triggers at end of
// combat.
func kytheonCombat(t *testing.T, g *game.Game, others int, mid func(attackers []uuid.UUID)) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	kytheon := pushBattlefieldCardWithTimestamp(g, kytheonCard(me.ID))
	attackers := []uuid.UUID{kytheon}
	for i := 0; i < others; i++ {
		attackers = append(attackers, pushVanillaCreature(g, me.ID, "Soldier", 1, 1))
	}
	advanceToDeclareAttackersOf(t, g, 0)
	for _, a := range attackers {
		if err := g.DeclareAttacker(a, opp.ID); err != nil {
			t.Fatalf("declare: %v", err)
		}
	}
	if mid != nil {
		advanceTo(t, g, game.StepDeclareBlockers)
		mid(attackers)
	}
	advanceTo(t, g, game.StepEndCombat)
	passPriorityAroundTable(t, g)
}

// Kytheon and two others attacked: at end of combat he is exiled and
// returns transformed, a planeswalker with his printed loyalty 3.
func TestKytheonTransformsAfterAttackingWithTwoOthers(t *testing.T) {
	g := newCatalogGame(t)
	kytheonCombat(t, g, 2, nil)
	w := battleForged(g)
	if w == nil {
		t.Fatal("Kytheon did not return transformed")
	}
	if got := w.Counters[game.CounterLoyalty]; got != 3 {
		t.Errorf("loyalty = %d, want 3 (CR 306.5b)", got)
	}
	if w.Controller != g.Seats[0].ID {
		t.Error("he returns under his owner's control")
	}
}

// One other attacker is not enough: no flip.
func TestKytheonDoesNotTransformWithOnlyOneOtherAttacker(t *testing.T) {
	g := newCatalogGame(t)
	kytheonCombat(t, g, 1, nil)
	if battleForged(g) != nil {
		t.Fatal("Kytheon transformed with only one other attacker")
	}
}

// A creature that attacked and then left the battlefield still attacked
// this combat (the ruling): the flip happens.
func TestKytheonCountsAnAttackerThatDied(t *testing.T) {
	g := newCatalogGame(t)
	kytheonCombat(t, g, 2, func(attackers []uuid.UUID) {
		g.WithWriteLock(func() {
			if err := g.DestroyPermanentForEffect(attackers[2]); err != nil {
				t.Fatalf("destroy: %v", err)
			}
		})
	})
	if battleForged(g) == nil {
		t.Fatal("an attacker that died no longer counted")
	}
}

// gideonBattleForgedPlusTwo seats Gideon, Battle-Forged for seat 0 in
// its precombat main phase and activates the +2 at `creature`.
func gideonBattleForgedPlusTwo(t *testing.T, g *game.Game, creature uuid.UUID) uuid.UUID {
	t.Helper()
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	gideon := pushWalkerForTest(g, me.ID, "Gideon, Battle-Forged", kytheonOracle+"#1", 3)
	b16Activate(t, g, me.ID, gideon, 0, game.ActivateAbilityParams{Targets: cardRefs(creature)})
	if got := loyaltyCount(g, gideon); got != 5 {
		t.Fatalf("loyalty after +2 = %d, want 5", got)
	}
	return gideon
}

// The +2 binds the one creature in its controller's next turn, and no
// other creature of theirs; the turn after asks nothing.
func TestGideonBattleForgedPlusTwoBindsThatCreatureInItsControllersNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bound := pushVanillaCreature(g, opp.ID, "Bound Bear", 2, 2)
	pushVanillaCreature(g, opp.ID, "Free Bear", 2, 2)
	gideon := gideonBattleForgedPlusTwo(t, g, bound)

	advanceToDeclareAttackersOf(t, g, 1)
	var re *game.AttackRequirementError
	if err := leaveDeclareAttackers(g); !errors.As(err, &re) || re.Attacker != bound {
		t.Fatalf("empty declaration = %v, want a refusal naming the targeted creature", err)
	}
	if got := re.Sentence(opp.ID); got != "Bound Bear must attack Gideon, Battle-Forged if able." {
		t.Errorf("sentence = %q", got)
	}
	if err := g.DeclareAttacker(bound, gideon); err != nil {
		t.Fatalf("attack Gideon: %v", err)
	}
	if err := leaveDeclareAttackers(g); err != nil {
		t.Fatalf("the other Bear home: %v, want nothing owed", err)
	}
	advanceToUpkeepOf(t, g, 2)
	if n := attackThisRecords(g); n != 0 {
		t.Errorf("%d records outlived its controller's next turn", n)
	}
}

// "Its controller" is read live (the ruling): a creature that changes
// controllers before its chance to attack is bound during its new
// controller's next turn, and its old controller's turn asks nothing.
func TestGideonBattleForgedPlusTwoFollowsANewController(t *testing.T) {
	g := newCatalogGame(t)
	oldCtl, newCtl := g.Seats[1], g.Seats[2]
	bound := pushVanillaCreature(g, oldCtl.ID, "Bound Bear", 2, 2)
	gideon := gideonBattleForgedPlusTwo(t, g, bound)
	g.WithWriteLock(func() {
		if !g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(bound),
			[]game.Mod{game.SetControllerMod(newCtl.ID)}, game.IndefiniteDuration(), "test — steal") {
			t.Fatal("control change registered nothing")
		}
		g.RecomputeLayersIfStaleLocked()
	})
	// Summoning sickness is about continuous control since the turn
	// began, which a creature stolen before that turn has.
	advanceToDeclareAttackersOf(t, g, 1)
	if err := leaveDeclareAttackers(g); err != nil {
		t.Fatalf("old controller's turn: %v, want nothing owed", err)
	}
	if attackThisRecords(g) != 1 {
		t.Fatal("the old controller's turn ended the requirement")
	}
	advanceToDeclareAttackersOf(t, g, 2)
	if err := leaveDeclareAttackers(g); !errors.Is(err, game.ErrAttackRequirement) {
		t.Fatalf("new controller's empty declaration = %v, want ErrAttackRequirement", err)
	}
	if err := g.DeclareAttacker(bound, gideon); err != nil {
		t.Fatalf("attack Gideon: %v", err)
	}
	if err := leaveDeclareAttackers(g); err != nil {
		t.Fatal(err)
	}
	advanceToUpkeepOf(t, g, 3)
	if n := attackThisRecords(g); n != 0 {
		t.Errorf("%d records outlived its new controller's next turn", n)
	}
}

// The duration kind and the requirement both survive a strict restore.
func TestGideonBattleForgedRequirementSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bound := pushVanillaCreature(g, opp.ID, "Bound Bear", 2, 2)
	gideonBattleForgedPlusTwo(t, g, bound)
	r := restoreRoundTrip(t, g, true)
	advanceToDeclareAttackersOf(t, r, 1)
	if err := leaveDeclareAttackers(r); !errors.Is(err, game.ErrAttackRequirement) {
		t.Fatalf("restored empty declaration = %v, want ErrAttackRequirement", err)
	}
}

// The +1: indestructible until your next turn, and untapped.
func TestGideonBattleForgedPlusOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMain(t, g)
	gideon := pushWalkerForTest(g, me.ID, "Gideon, Battle-Forged", kytheonOracle+"#1", 3)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if err := g.TapCard(bear, true); err != nil {
		t.Fatal(err)
	}
	b16Activate(t, g, me.ID, gideon, 1, game.ActivateAbilityParams{Targets: cardRefs(bear)})
	if b16Tapped(t, g, bear) {
		t.Error("the creature was not untapped")
	}
	if !hasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Fatal("no indestructible")
	}
	advanceToUpkeepOf(t, g, 1)
	if !hasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("indestructible ended before your next turn")
	}
	advanceToUpkeepOf(t, g, 0)
	if hasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("indestructible outlived the start of your next turn")
	}
}
