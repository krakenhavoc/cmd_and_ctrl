package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Citadel Siege — Enchantment {2}{W}{W}:
//
//	"As this enchantment enters, choose Khans or Dragons.
//	 • Khans — At the beginning of combat on your turn, put two +1/+1
//	   counters on target creature you control.
//	 • Dragons — At the beginning of combat on each opponent's turn,
//	   tap target creature that player controls."
//
// A #1572 anchor-word card: each bullet is a beginning-of-combat
// trigger gated on its word (ADR 0071).
//
// "That player" in the Dragons line is the opponent whose turn it is,
// so the target clause is "a creature the ACTIVE player controls" —
// read off the turn cursor when the trigger goes on the stack and
// again as it resolves, both inside that opponent's beginning of
// combat step. An opponent with no creature drops the trigger with no
// prompt (CR 603.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "40f66e21-5b60-4e42-927d-65397e7ad544",
		Name:         "Citadel Siege",
		Completeness: CompletenessFull,
		AsEnters:     ChooseOptionAsEnters("Citadel Siege", "Khans", "Dragons"),
		Triggered: []game.TriggeredAbility{
			WhenChosen("Khans", Targeting(
				AtBeginningOfYourCombat("Citadel Siege — put two +1/+1 counters on target creature you control",
					citadelSiegeCounters),
				TargetCreature("target creature you control", YouControl()))),
			WhenChosen("Dragons", Targeting(
				On(game.EventStepBegan, AllOf(StepBegan(game.StepBeginCombat, false), ByAnOpponent),
					"Citadel Siege — tap target creature that player controls", citadelSiegeTap),
				TargetCreature("target creature that player controls", controlledByTheActivePlayer()))),
		},
	})
}

// controlledByTheActivePlayer is "that player controls" when "that
// player" is the one whose turn it is. Reads the turn cursor directly
// (isActivePlayer) because target predicates run under g.mu.
func controlledByTheActivePlayer() CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool { return isActivePlayer(g, c.Controller) }
}

// citadelSiegeCounters is the Khans body.
func citadelSiegeCounters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		return AddCounter{Target: t.ID, Kind: "+1/+1", N: 2}.Apply(ctx)
	}
	return nil
}

// citadelSiegeTap is the Dragons body.
func citadelSiegeTap(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		return TapTarget{Target: t.ID}.Apply(ctx)
	}
	return nil
}
