package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// block_capacity_eot_cards_test.go — #1715's cards. The engine half
// (the capacity records, the blocksEach requirement, undo and the
// snapshot) is pinned in game/multi_block_followups_test.go.

const (
	coastlineChimeraOracle   = "da95382c-0536-49aa-a0db-3b52926bf559"
	mountedArchersOracle     = "22517690-3ec1-49c9-951f-93eb3114423a"
	giveNoGroundOracle       = "bae18a46-00e0-4b3e-8fdb-44cc61f30f8e"
	valorMadeRealOracle      = "5bd1348c-51cc-4426-b995-8bb6d4a5bd2e"
	actOfHeroismOracle       = "25edb501-11b5-4617-8a65-2eb4869cccd5"
	yareOracle               = "f0f24c6d-80fb-4e99-a74d-6ed9d349ed3d"
	blazeOfGloryOracle       = "b330ac89-790e-4cc9-96a5-532c48252088"
	lairwatchGiantOracle     = "dac0dc4c-acf3-40de-b6b5-963ff4176d3a"
	anuridSwarmsnapperOracle = "fd275ea7-9ccc-4148-9b33-392a612486dd"
	luminousGuardianOracle   = "0d9a31f5-20d0-4266-95ec-38b4fa96b8aa"
)

// bcCast puts an instant into `caster`'s hand and casts it at `target`
// wherever the cursor is, returning the announce error.
func bcCast(g *game.Game, caster *game.Player, name, oracle string, target uuid.UUID) error {
	id := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Instant",
		OracleID: oracle, Owner: caster.ID, Controller: caster.ID,
	})
	return g.CastSpell(caster.ID, id, game.CastSpellParams{Targets: cardRefs(target)})
}

// bcCombat is the shared table: `me` (the active seat) has n 2/2s to
// attack the defender `opp` with. The cursor is at the beginning of
// combat, where every instant here can be cast.
func bcCombat(t *testing.T, n int) (g *game.Game, me, opp *game.Player, atks []uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	for i := 0; i < n; i++ {
		atks = append(atks, pushVanillaCreature(g, me.ID, "Bear", 2, 2))
	}
	advanceTo(t, g, game.StepBeginCombat)
	return g, me, opp, atks
}

// bcAttack declares every attacker at `opp` and moves to the declare
// blockers step.
func bcAttack(t *testing.T, g *game.Game, opp *game.Player, atks []uuid.UUID) {
	t.Helper()
	declareAttack(t, g, opp.ID, atks...)
	advanceTo(t, g, game.StepDeclareBlockers)
}

// TestValorMadeRealBlocksEveryAttackerThisTurnOnly — the defender casts
// it on their own wall at the beginning of combat: the wall blocks all
// three attackers; after cleanup it is back to one, and the record
// rides a restore point.
func TestValorMadeRealBlocksEveryAttackerThisTurnOnly(t *testing.T) {
	g, _, opp, atks := bcCombat(t, 3)
	wall := pushMonster(g, opp.ID, "Wall", "", 0, 6)
	floatMana(t, g, opp, "{W}")
	if err := bcCast(g, opp, "Valor Made Real", valorMadeRealOracle, wall); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if got := mbCapacity(t, g, wall); got != 0 {
		t.Fatalf("capacity = %d, want any number", got)
	}
	if got := mbCapacity(t, restoreRoundTrip(t, g, true), wall); got != 0 {
		t.Fatalf("restored capacity = %d, want any number", got)
	}
	bcAttack(t, g, opp, atks)
	if err := mbBlock(g, wall, atks...); err != nil {
		t.Fatalf("blocking all three: %v", err)
	}
	advancePastCleanupForTest(t, g)
	if got := mbCapacity(t, g, wall); got != 1 {
		t.Errorf("capacity next turn = %d, want 1", got)
	}
}

// TestGiveNoGroundIsOneEffect — +2/+6 and "any number" together, both
// gone at cleanup.
func TestGiveNoGroundIsOneEffect(t *testing.T) {
	g, _, opp, _ := bcCombat(t, 0)
	bear := pushMonster(g, opp.ID, "Bear", "", 2, 2)
	floatMana(t, g, opp, "{C}{C}{C}{W}")
	if err := bcCast(g, opp, "Give No Ground", giveNoGroundOracle, bear); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 4 || tg != 8 {
		t.Errorf("Bear is %d/%d, want 4/8", p, tg)
	}
	if got := mbCapacity(t, g, bear); got != 0 {
		t.Errorf("capacity = %d, want any number", got)
	}
	advancePastCleanupForTest(t, g)
	if p, got := effectivePower(t, g, bear), mbCapacity(t, g, bear); p != 2 || got != 1 {
		t.Errorf("after cleanup: power %d capacity %d, want 2 and 1", p, got)
	}
}

// TestActOfHeroismUntapsPumpsAndAddsABlock.
func TestActOfHeroismUntapsPumpsAndAddsABlock(t *testing.T) {
	g, _, opp, atks := bcCombat(t, 2)
	bear := pushMonster(g, opp.ID, "Bear", "", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	floatMana(t, g, opp, "{C}{W}")
	if err := bcCast(g, opp, "Act of Heroism", actOfHeroismOracle, bear); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if e2Card(t, g, bear).Tapped {
		t.Fatal("the target is still tapped")
	}
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 4 || tg != 4 {
		t.Errorf("Bear is %d/%d, want 4/4", p, tg)
	}
	bcAttack(t, g, opp, atks)
	if err := mbBlock(g, bear, atks...); err != nil {
		t.Fatalf("blocking two: %v", err)
	}
}

// TestYareTargetsOnlyADefendingPlayersCreatureDuringCombat — +3/+0 and
// up to two more blocks, on a creature a defending player controls; no
// target in a main phase, and none among the active player's own.
func TestYareTargetsOnlyADefendingPlayersCreatureDuringCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushMonster(g, opp.ID, "Bear", "", 2, 2)
	mine := pushMonster(g, me.ID, "My Bear", "", 2, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	floatMana(t, g, opp, "{C}{C}{W}")
	if err := bcCast(g, opp, "Yare", yareOracle, bear); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("Yare in a main phase: %v, want ErrIllegalTarget", err)
	}
	advanceTo(t, g, game.StepBeginCombat)
	if err := bcCast(g, opp, "Yare", yareOracle, mine); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("Yare on the active player's creature: %v, want ErrIllegalTarget", err)
	}
	if err := bcCast(g, opp, "Yare", yareOracle, bear); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 5 || tg != 2 {
		t.Errorf("Bear is %d/%d, want 5/2", p, tg)
	}
	if got := mbCapacity(t, g, bear); got != 3 {
		t.Errorf("capacity = %d, want 3", got)
	}
}

// TestBlazeOfGloryTargetMustBlockEveryAttacker — cast at the beginning
// of combat on the defender's wall: passing with an attacker it could
// block left over is refused, and blocking all three is accepted. Cast
// once blockers are being declared, it is refused.
func TestBlazeOfGloryTargetMustBlockEveryAttacker(t *testing.T) {
	g, _, opp, atks := bcCombat(t, 3)
	wall := pushMonster(g, opp.ID, "Wall", "", 0, 6)
	floatMana(t, g, opp, "{W}")
	if err := bcCast(g, opp, "Blaze of Glory", blazeOfGloryOracle, wall); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	bcAttack(t, g, opp, atks)

	if err := mbBlock(g, wall, atks[0]); err != nil {
		t.Fatal(err)
	}
	var br *game.BlockRefusedError
	if err := g.FinishBlocks(opp.ID); !errors.As(err, &br) || br.Reason != game.BlockReasonRequirement {
		t.Fatalf("finishing with two attackers unblocked: %v, want a block requirement refusal", err)
	}
	if br.Blocker != wall {
		t.Errorf("the refusal names %s, want the wall", br.Blocker)
	}
	if got := br.Sentence(opp.ID); got != "Wall must block Bear if able (Blaze of Glory)." {
		t.Errorf("sentence = %q", got)
	}
	if err := mbBlock(g, wall, atks[1:]...); err != nil {
		t.Fatal(err)
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatalf("finishing with every attacker blocked: %v", err)
	}

	// Too late once the declare blockers step has begun.
	floatMana(t, g, opp, "{W}")
	if err := bcCast(g, opp, "Blaze of Glory", blazeOfGloryOracle, wall); err == nil {
		t.Fatal("Blaze of Glory was cast in the declare blockers step")
	}
}

// TestLairwatchGiantTriggersOnlyWhenItBlocksTwo.
func TestLairwatchGiantTriggersOnlyWhenItBlocksTwo(t *testing.T) {
	for _, n := range []int{1, 2} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		giant := pushMonster(g, opp.ID, "Lairwatch Giant", lairwatchGiantOracle, 5, 3)
		if got := mbCapacity(t, g, giant); got != 2 {
			t.Fatalf("capacity = %d, want 2", got)
		}
		atks := mbAttack(t, g, me, opp, n)
		if err := mbBlock(g, giant, atks...); err != nil {
			t.Fatal(err)
		}
		if err := g.FinishBlocks(opp.ID); err != nil {
			t.Fatal(err)
		}
		want := 0
		if n == 2 {
			want = 1
		}
		if got := triggersOnStackFrom(g, giant); got != want {
			t.Fatalf("blocking %d: %d triggers, want %d", n, got, want)
		}
		passPriorityAroundTable(t, g)
		if got := hasString(effectiveAbilities(t, g, giant), "first strike"); got != (n == 2) {
			t.Errorf("blocking %d: first strike = %v", n, got)
		}
	}
}

// TestSelfCapacityActivationsAddUp — Coastline Chimera, Mounted
// Archers, Anurid Swarmsnapper and Luminous Guardian: each activation
// is one more attacker this turn; cleanup takes them all.
func TestSelfCapacityActivationsAddUp(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, cost, kw string
		idx                    int
	}{
		{"Coastline Chimera", coastlineChimeraOracle, "{C}{W}", "flying", 0},
		{"Mounted Archers", mountedArchersOracle, "{W}", "reach", 0},
		{"Anurid Swarmsnapper", anuridSwarmsnapperOracle, "{C}{G}", "reach", 0},
		{"Luminous Guardian", luminousGuardianOracle, "{C}{C}", "", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			id := pushMonster(g, me.ID, tc.name, tc.oracle, 1, 5)
			if tc.kw != "" && !hasString(effectiveAbilities(t, g, id), tc.kw) {
				t.Errorf("no %s", tc.kw)
			}
			advanceTo(t, g, game.StepPrecombatMain)
			for want := 2; want <= 3; want++ {
				floatMana(t, g, me, tc.cost)
				if err := g.ActivateCatalogAbility(me.ID, id, tc.idx, game.ActivateAbilityParams{}); err != nil {
					t.Fatalf("activate: %v", err)
				}
				passPriorityAroundTable(t, g)
				if got := mbCapacity(t, g, id); got != want {
					t.Fatalf("capacity after %d activations = %d, want %d", want-1, got, want)
				}
			}
			advancePastCleanupForTest(t, g)
			if got := mbCapacity(t, g, id); got != 1 {
				t.Errorf("capacity next turn = %d, want 1", got)
			}
		})
	}
}

// TestLuminousGuardianPumpsToughness — its other activation.
func TestLuminousGuardianPumpsToughness(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushMonster(g, me.ID, "Luminous Guardian", luminousGuardianOracle, 1, 4)
	advanceTo(t, g, game.StepPrecombatMain)
	floatMana(t, g, me, "{W}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if p, tg := effectivePower(t, g, id), effectiveToughness(t, g, id); p != 1 || tg != 5 {
		t.Errorf("Guardian is %d/%d, want 1/5", p, tg)
	}
	if got := mbCapacity(t, g, id); got != 1 {
		t.Errorf("the pump changed its capacity to %d", got)
	}
}
