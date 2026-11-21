package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// delayed_attacks_unblocked_test.go — #3531's delayed "attacks and isn't
// blocked" triggers (ADR 0187 amendment 2026-10-16): Gaze of Pain, Delif's
// Cone and Delif's Cube.

const (
	dauGazeOfPain = "b3a479d4-3e36-496b-abf8-b83231628d76"
	dauDelifsCone = "8af33199-d3ad-44fe-8ee0-abbe36a2278b"
	dauDelifsCube = "3106ba15-08d7-474c-a384-6848b713ba8d"
)

// dauCombat attacks seat 1 with `attackers`, declares `blocks` (blocker →
// attacker), then answers every prompt the triggers raise — `target` for
// each target pick, `yes` for each "you may" — and walks into the combat
// damage step. It reports how many target picks it answered.
func dauCombat(t *testing.T, g *game.Game, attackers []uuid.UUID, blocks map[uuid.UUID]uuid.UUID, yes bool, target uuid.UUID) int {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.Pr7bAttack(t, g, opp.ID, attackers...)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	for blocker, attacker := range blocks {
		if err := g.DeclareBlocker(blocker, attacker); err != nil {
			t.Fatalf("DeclareBlocker: %v", err)
		}
	}
	effectstest.FinishBlockDeclarations(t, g)
	picks := 0
	for i := 0; i < 48; i++ {
		var open *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil {
				open = c
				break
			}
		}
		if open == nil {
			if len(g.StackMeta) == 0 && len(g.PendingTriggers) == 0 {
				break
			}
			if err := g.PassPriority(); err != nil {
				break
			}
			continue
		}
		switch open.Kind {
		case game.PendingChoicePickTarget:
			picks++
			if err := g.ResolvePickTargets(open.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: target}}); err != nil {
				t.Fatalf("pick target: %v", err)
			}
		case game.PendingChoiceConfirm:
			if err := g.ResolveConfirm(open.ID, me.ID, yes); err != nil {
				t.Fatalf("confirm: %v", err)
			}
		default:
			t.Fatalf("unexpected prompt %s", open.Kind)
		}
	}
	effectstest.Pr7bDamageStep(t, g)
	return picks
}

// Gaze of Pain: each creature you control that attacks and isn't blocked
// triggers once, with itself as "it". A yes deals its power to the
// target and it deals no combat damage; the blocked one never triggers.
func TestGazeOfPainTriggersForEachUnblockedCreature(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		a := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("A", "", 2, 2))
		b := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("B", "", 3, 3))
		c := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("C", "", 1, 5))
		wall := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Wall", "", 0, 9))
		victim := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Victim", "", 0, 20))
		effectstest.CastCatalogSpell(t, g, "Gaze of Pain", "Sorcery", dauGazeOfPain, nil)
		effectstest.PassPriorityAroundTable(t, g)
		life := opp.Life

		picks := dauCombat(t, g, []uuid.UUID{a, b, c}, map[uuid.UUID]uuid.UUID{wall: c}, yes, victim)
		if picks != 2 {
			t.Fatalf("yes=%v: %d targeted instances, want 2 (one for each unblocked creature)", yes, picks)
		}
		dmg, lost := effectstest.Pr6Marked(g, victim), effectstest.NcdLifeLost(g, life)
		if yes && (dmg != 5 || lost != 0 || !effectstest.NcdMarked(g, a) || !effectstest.NcdMarked(g, b)) {
			t.Errorf("yes: victim took %d (want 5), defender lost %d (want 0), marked %v %v",
				dmg, lost, effectstest.NcdMarked(g, a), effectstest.NcdMarked(g, b))
		}
		if !yes && (dmg != 0 || lost != 5 || effectstest.NcdMarked(g, a) || effectstest.NcdMarked(g, b)) {
			t.Errorf("no: victim took %d (want 0), defender lost %d (want 5), marked %v %v",
				dmg, lost, effectstest.NcdMarked(g, a), effectstest.NcdMarked(g, b))
		}
		if effectstest.NcdMarked(g, c) {
			t.Errorf("yes=%v: the blocked creature was marked", yes)
		}
	}
}

// Gaze of Pain lasts until end of turn (CR 603.7b, 514.2).
func TestGazeOfPainEndsWithTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.CastCatalogSpell(t, g, "Gaze of Pain", "Sorcery", dauGazeOfPain, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("%d delayed triggers after Gaze of Pain resolved, want 1", len(g.DelayedTriggers))
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("%d delayed triggers survived the turn, want 0", len(g.DelayedTriggers))
	}
}

// dauActivate activates the artifact's ability `idx` at the creature with
// the given mana, and resolves it.
func dauActivate(t *testing.T, g *game.Game, artifact uuid.UUID, idx int, target uuid.UUID, mana ...string) {
	t.Helper()
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, mana...)
	if err := g.ActivateCatalogAbility(me.ID, artifact, idx, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// Delif's Cone: when the creature attacks and isn't blocked, a yes gains
// life equal to its power and it deals no combat damage; a no does
// neither. Another unblocked creature does not fire it.
func TestDelifsConeGainsLifeInsteadOfDamage(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		cone := effectstest.PushCatalogPermanent(g, me.ID, "Delif's Cone", "Artifact", dauDelifsCone, false)
		a := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("A", "", 4, 4))
		other := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Other", "", 1, 1))
		dauActivate(t, g, cone, 0, a)
		if effects.OnBattlefield(g, cone) {
			t.Fatal("Delif's Cone was not sacrificed")
		}
		mine, theirs := me.Life, opp.Life

		dauCombat(t, g, []uuid.UUID{a, other}, nil, yes, uuid.Nil)
		gained, lost := me.Life-mine, effectstest.NcdLifeLost(g, theirs)
		if yes && (gained != 4 || lost != 1 || !effectstest.NcdMarked(g, a)) {
			t.Errorf("yes: gained %d (want 4), defender lost %d (want 1), marked %v", gained, lost, effectstest.NcdMarked(g, a))
		}
		if !yes && (gained != 0 || lost != 5 || effectstest.NcdMarked(g, a)) {
			t.Errorf("no: gained %d (want 0), defender lost %d (want 5), marked %v", gained, lost, effectstest.NcdMarked(g, a))
		}
		if effectstest.NcdMarked(g, other) {
			t.Error("the other creature was marked")
		}
	}
}

// A blocked creature never fires Delif's Cone's trigger, and the trigger
// is gone at the end of the turn.
func TestDelifsConeBlockedDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cone := effectstest.PushCatalogPermanent(g, me.ID, "Delif's Cone", "Artifact", dauDelifsCone, false)
	a := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("A", "", 4, 4))
	wall := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Wall", "", 0, 9))
	dauActivate(t, g, cone, 0, a)
	mine := me.Life
	dauCombat(t, g, []uuid.UUID{a}, map[uuid.UUID]uuid.UUID{wall: a}, true, uuid.Nil)
	if me.Life != mine || effectstest.NcdMarked(g, a) {
		t.Errorf("blocked: gained %d, marked %v; want neither", me.Life-mine, effectstest.NcdMarked(g, a))
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("%d delayed triggers survived the turn, want 0", len(g.DelayedTriggers))
	}
}

// Delif's Cube: the creature deals no combat damage and the Cube gets a
// cube counter, which then pays for a regeneration shield.
func TestDelifsCubeMarksTheCreatureAndCountsACube(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cube := effectstest.PushCatalogPermanent(g, me.ID, "Delif's Cube", "Artifact", dauDelifsCube, false)
	a := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("A", "", 3, 3))
	dauActivate(t, g, cube, 0, a, "C", "C")
	life := opp.Life

	dauCombat(t, g, []uuid.UUID{a}, nil, true, uuid.Nil)
	if lost := effectstest.NcdLifeLost(g, life); lost != 0 || !effectstest.NcdMarked(g, a) {
		t.Fatalf("defender lost %d (want 0), marked %v (want true)", lost, effectstest.NcdMarked(g, a))
	}
	if n := effectstest.LayeredCard(t, g, cube).Counters[effects.DelifsCubeCounter]; n != 1 {
		t.Fatalf("%d cube counters, want 1", n)
	}

	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, cube, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}},
	}); err != nil {
		t.Fatalf("regenerate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.LayeredCard(t, g, cube).Counters[effects.DelifsCubeCounter]; n != 0 {
		t.Errorf("%d cube counters after paying one, want 0", n)
	}
	if effectstest.LayeredCard(t, g, a).RegenerationShields != 1 {
		t.Errorf("the creature has %d regeneration shields, want 1", effectstest.LayeredCard(t, g, a).RegenerationShields)
	}
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, cube, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}},
	}); err == nil {
		t.Error("the regeneration ability was activated with no cube counter to remove")
	}
}

// A Cube that has left the battlefield before its trigger resolves gets
// no counter, and the creature is still marked.
func TestDelifsCubeGoneStillMarksTheCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cube := effectstest.PushCatalogPermanent(g, me.ID, "Delif's Cube", "Artifact", dauDelifsCube, false)
	a := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("A", "", 3, 3))
	dauActivate(t, g, cube, 0, a, "C", "C")
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, cube); err != nil {
		t.Fatal(err)
	}
	life := opp.Life
	dauCombat(t, g, []uuid.UUID{a}, nil, true, uuid.Nil)
	if lost := effectstest.NcdLifeLost(g, life); lost != 0 || !effectstest.NcdMarked(g, a) {
		t.Errorf("defender lost %d (want 0), marked %v (want true)", lost, effectstest.NcdMarked(g, a))
	}
}
