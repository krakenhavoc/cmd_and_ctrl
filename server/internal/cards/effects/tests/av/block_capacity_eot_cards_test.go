package testsav

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// bcCast puts an instant into `caster`'s hand and casts it at `target`
// wherever the cursor is, returning the announce error.
func bcCast(g *game.Game, caster *game.Player, name, oracle string, target uuid.UUID) error {
	id := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Instant",
		OracleID: oracle, Owner: caster.ID, Controller: caster.ID,
	})
	return g.CastSpell(caster.ID, id, game.CastSpellParams{Targets: effectstest.CardRefs(target)})
}

// TestValorMadeRealBlocksEveryAttackerThisTurnOnly — the defender casts
// it on their own wall at the beginning of combat: the wall blocks all
// three attackers; after cleanup it is back to one, and the record
// rides a restore point.
func TestValorMadeRealBlocksEveryAttackerThisTurnOnly(t *testing.T) {
	g, _, opp, atks := effectstest.BcCombat(t, 3)
	wall := effectstest.PushMonster(g, opp.ID, "Wall", "", 0, 6)
	effectstest.FloatMana(t, g, opp, "{W}")
	if err := bcCast(g, opp, "Valor Made Real", effectstest.ValorMadeRealOracle, wall); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.MbCapacity(t, g, wall); got != 0 {
		t.Fatalf("capacity = %d, want any number", got)
	}
	if got := effectstest.MbCapacity(t, effectstest.RestoreRoundTrip(t, g, true), wall); got != 0 {
		t.Fatalf("restored capacity = %d, want any number", got)
	}
	effectstest.BcAttack(t, g, opp, atks)
	if err := effectstest.MbBlock(g, wall, atks...); err != nil {
		t.Fatalf("blocking all three: %v", err)
	}
	effectstest.AdvancePastCleanupForTest(t, g)
	if got := effectstest.MbCapacity(t, g, wall); got != 1 {
		t.Errorf("capacity next turn = %d, want 1", got)
	}
}

// TestGiveNoGroundIsOneEffect — +2/+6 and "any number" together, both
// gone at cleanup.
func TestGiveNoGroundIsOneEffect(t *testing.T) {
	g, _, opp, _ := effectstest.BcCombat(t, 0)
	bear := effectstest.PushMonster(g, opp.ID, "Bear", "", 2, 2)
	effectstest.FloatMana(t, g, opp, "{C}{C}{C}{W}")
	if err := bcCast(g, opp, "Give No Ground", effectstest.GiveNoGroundOracle, bear); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 4 || tg != 8 {
		t.Errorf("Bear is %d/%d, want 4/8", p, tg)
	}
	if got := effectstest.MbCapacity(t, g, bear); got != 0 {
		t.Errorf("capacity = %d, want any number", got)
	}
	effectstest.AdvancePastCleanupForTest(t, g)
	if p, got := effectstest.EffectivePower(t, g, bear), effectstest.MbCapacity(t, g, bear); p != 2 || got != 1 {
		t.Errorf("after cleanup: power %d capacity %d, want 2 and 1", p, got)
	}
}

// TestActOfHeroismUntapsPumpsAndAddsABlock.
func TestActOfHeroismUntapsPumpsAndAddsABlock(t *testing.T) {
	g, _, opp, atks := effectstest.BcCombat(t, 2)
	bear := effectstest.PushMonster(g, opp.ID, "Bear", "", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	effectstest.FloatMana(t, g, opp, "{C}{W}")
	if err := bcCast(g, opp, "Act of Heroism", effectstest.ActOfHeroismOracle, bear); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.E2Card(t, g, bear).Tapped {
		t.Fatal("the target is still tapped")
	}
	if p, tg := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 4 || tg != 4 {
		t.Errorf("Bear is %d/%d, want 4/4", p, tg)
	}
	effectstest.BcAttack(t, g, opp, atks)
	if err := effectstest.MbBlock(g, bear, atks...); err != nil {
		t.Fatalf("blocking two: %v", err)
	}
}

// TestYareTargetsOnlyADefendingPlayersCreatureDuringCombat — +3/+0 and
// up to two more blocks, on a creature a defending player controls; no
// target in a main phase, and none among the active player's own.
func TestYareTargetsOnlyADefendingPlayersCreatureDuringCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushMonster(g, opp.ID, "Bear", "", 2, 2)
	mine := effectstest.PushMonster(g, me.ID, "My Bear", "", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.FloatMana(t, g, opp, "{C}{C}{W}")
	if err := bcCast(g, opp, "Yare", effectstest.YareOracle, bear); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("Yare in a main phase: %v, want ErrIllegalTarget", err)
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := bcCast(g, opp, "Yare", effectstest.YareOracle, mine); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("Yare on the active player's creature: %v, want ErrIllegalTarget", err)
	}
	if err := bcCast(g, opp, "Yare", effectstest.YareOracle, bear); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 5 || tg != 2 {
		t.Errorf("Bear is %d/%d, want 5/2", p, tg)
	}
	if got := effectstest.MbCapacity(t, g, bear); got != 3 {
		t.Errorf("capacity = %d, want 3", got)
	}
}

// TestBlazeOfGloryTargetMustBlockEveryAttacker — cast at the beginning
// of combat on the defender's wall: passing with an attacker it could
// block left over is refused, and blocking all three is accepted. Cast
// once blockers are being declared, it is refused.
func TestBlazeOfGloryTargetMustBlockEveryAttacker(t *testing.T) {
	g, _, opp, atks := effectstest.BcCombat(t, 3)
	wall := effectstest.PushMonster(g, opp.ID, "Wall", "", 0, 6)
	effectstest.FloatMana(t, g, opp, "{W}")
	if err := bcCast(g, opp, "Blaze of Glory", effectstest.BlazeOfGloryOracle, wall); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.BcAttack(t, g, opp, atks)

	if err := effectstest.MbBlock(g, wall, atks[0]); err != nil {
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
	if err := effectstest.MbBlock(g, wall, atks[1:]...); err != nil {
		t.Fatal(err)
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatalf("finishing with every attacker blocked: %v", err)
	}

	// Too late once the declare blockers step has begun.
	effectstest.FloatMana(t, g, opp, "{W}")
	if err := bcCast(g, opp, "Blaze of Glory", effectstest.BlazeOfGloryOracle, wall); err == nil {
		t.Fatal("Blaze of Glory was cast in the declare blockers step")
	}
}

// TestLairwatchGiantTriggersOnlyWhenItBlocksTwo.
func TestLairwatchGiantTriggersOnlyWhenItBlocksTwo(t *testing.T) {
	for _, n := range []int{1, 2} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		giant := effectstest.PushMonster(g, opp.ID, "Lairwatch Giant", effectstest.LairwatchGiantOracle, 5, 3)
		if got := effectstest.MbCapacity(t, g, giant); got != 2 {
			t.Fatalf("capacity = %d, want 2", got)
		}
		atks := effectstest.MbAttack(t, g, me, opp, n)
		if err := effectstest.MbBlock(g, giant, atks...); err != nil {
			t.Fatal(err)
		}
		if err := g.FinishBlocks(opp.ID); err != nil {
			t.Fatal(err)
		}
		want := 0
		if n == 2 {
			want = 1
		}
		if got := effectstest.TriggersOnStackFrom(g, giant); got != want {
			t.Fatalf("blocking %d: %d triggers, want %d", n, got, want)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.HasString(effectstest.EffectiveAbilities(t, g, giant), "first strike"); got != (n == 2) {
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
		{"Coastline Chimera", effectstest.CoastlineChimeraOracle, "{C}{W}", "flying", 0},
		{"Mounted Archers", effectstest.MountedArchersOracle, "{W}", "reach", 0},
		{"Anurid Swarmsnapper", effectstest.AnuridSwarmsnapperOracle, "{C}{G}", "reach", 0},
		{"Luminous Guardian", effectstest.LuminousGuardianOracle, "{C}{C}", "", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			id := effectstest.PushMonster(g, me.ID, tc.name, tc.oracle, 1, 5)
			if tc.kw != "" && !effectstest.HasString(effectstest.EffectiveAbilities(t, g, id), tc.kw) {
				t.Errorf("no %s", tc.kw)
			}
			effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
			for want := 2; want <= 3; want++ {
				effectstest.FloatMana(t, g, me, tc.cost)
				if err := g.ActivateCatalogAbility(me.ID, id, tc.idx, game.ActivateAbilityParams{}); err != nil {
					t.Fatalf("activate: %v", err)
				}
				effectstest.PassPriorityAroundTable(t, g)
				if got := effectstest.MbCapacity(t, g, id); got != want {
					t.Fatalf("capacity after %d activations = %d, want %d", want-1, got, want)
				}
			}
			effectstest.AdvancePastCleanupForTest(t, g)
			if got := effectstest.MbCapacity(t, g, id); got != 1 {
				t.Errorf("capacity next turn = %d, want 1", got)
			}
		})
	}
}

// TestLuminousGuardianPumpsToughness — its other activation.
func TestLuminousGuardianPumpsToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushMonster(g, me.ID, "Luminous Guardian", effectstest.LuminousGuardianOracle, 1, 4)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.FloatMana(t, g, me, "{W}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); p != 1 || tg != 5 {
		t.Errorf("Guardian is %d/%d, want 1/5", p, tg)
	}
	if got := effectstest.MbCapacity(t, g, id); got != 1 {
		t.Errorf("the pump changed its capacity to %d", got)
	}
}
