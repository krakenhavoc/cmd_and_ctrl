package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nils, Discipline Enforcer — Legendary Creature — Human Cleric {2}{W},
// 2/2 (Secrets of Strixhaven Commander, #4834):
//
//	"At the beginning of your end step, for each player, put a +1/+1
//	 counter on up to one target creature that player controls.
//	 Each creature with one or more counters on it can't attack you or
//	 planeswalkers you control unless its controller pays {X}, where X is
//	 the number of counters on that creature."
//
// The end-step trigger has one "up to one" clause per seat, Nils's
// controller included, each bound to that player (Molten Primordial's
// TargetsFrom shape, which reads only the seat list, so a waiting
// trigger is a restore point). Every pick is optional (the ruling).
//
// The second ability is Sphere of Safety's attack tax with a per-creature
// count: X is every counter on the attacking creature, of any kind (the
// ruling), read at declaration (CR 508.1h), so a creature with no
// counters attacks for free. It covers attacks on Nils's controller and
// their planeswalkers, not battles.
//
// No simplification.
func init() {
	end := AtYourEndStep("Nils, Discipline Enforcer — for each player, a +1/+1 counter on up to one target creature they control", putPlusOneCounterOnEachLegalTarget)
	end.TargetsFrom = nilsDisciplineEnforcerClauses
	Register(Spec{
		OracleID:     "ff6b52b7-ae3c-465b-b10e-901160e5230e",
		Name:         "Nils, Discipline Enforcer",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{end},
		AttackTaxes: []game.AttackTax{
			ProtectingPlaneswalkers(AttackTaxCounting(
				func(q game.AttackTaxQuery) int { return socB9CountersOn(q.Attacker) },
				"Each creature with one or more counters on it can't attack you or planeswalkers you control unless its controller pays {X}, where X is the number of counters on that creature.",
			)),
		},
	})
}

// nilsDisciplineEnforcerClauses is one "up to one target creature that
// player controls" clause per seat, in seat order. It reads only the
// seat list, which is constant over an active game.
func nilsDisciplineEnforcerClauses(_ game.TriggerContext, _ *game.Card, g *game.Game) *game.TargetSpec {
	var clauses []*game.TargetSpec
	for _, p := range g.Seats {
		if p == nil {
			continue
		}
		clauses = append(clauses, TargetCreature("up to one target creature "+p.Name+" controls",
			ControlledBy(p.ID)).WithCount(0, 1))
	}
	if len(clauses) == 0 {
		return nil
	}
	return Clauses(clauses[0], clauses[1:]...)
}
