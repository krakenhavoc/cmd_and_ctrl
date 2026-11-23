package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Portal Manipulator — Creature — Human Wizard {2}{W/U}{W/U}, 3/2 (Ravnica:
// Clue Edition, #5929):
//
//	"Flash
//	 When this creature enters during the declare attackers step, choose
//	 target player and any number of target attacking creatures their
//	 opponents control. Those creatures are now attacking that player."
//
// The trigger is Misleading Signpost's: it fires only when the creature
// enters in the declare attackers step. Its second clause is judged against
// its first (ControlledByOpponentsOfTargetPlayer): each creature must be an
// attacking creature a player other than the chosen one controls, at
// announcement and again on resolution (CR 608.2b), so a creature that has
// left combat or changed controller drops out and the rest stay. The
// creature is not removed from combat and does not attack a second time, and
// it stays blocked if it was (CR 508.7a, 509.1h), which is what moving its
// attack through ReselectAttackTargetForEffect does. A chosen player that
// has become an illegal target gets nothing: an illegal target is not
// affected (CR 608.2b), and a creature whose controller cannot attack that
// player (CR 508.7c) is left where it is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0dd3127e-de13-4785-b5f2-0af996e9f15f",
		Name:            "Portal Manipulator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			portalManipulatorTrigger("Portal Manipulator — choose target player and any number of target attacking creatures their opponents control; those creatures are now attacking that player"),
		},
	})
}

func portalManipulatorTrigger(label string) game.TriggeredAbility {
	t := effects.On(game.EventETB, effects.EntersDuringDeclareAttackers, label, func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		player, ok := ctx.ClauseTarget(0)
		if !ok {
			return nil
		}
		for _, t := range ctx.ClauseTargets(1) {
			if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				continue
			}
			// A creature that cannot attack that player is left alone
			// (CR 508.7c).
			_ = g.ReselectAttackTargetForEffect(t.ID, player.ID)
		}
		return nil
	})
	t.Targets = effects.Clauses(
		effects.TargetPlayer("target player"),
		effects.ControlledByOpponentsOfTargetPlayer(
			effects.TargetCreature("any number of target attacking creatures their opponents control", effects.AttackingCreature()).WithCount(0, 0), 0))
	return t
}
