package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rakdos's Return — Sorcery {X}{B}{R} (Return to Ravnica, S82 OLD):
//
//	"Rakdos's Return deals X damage to target opponent or planeswalker.
//	 That player or that planeswalker's controller discards X cards."
//
// Blightning's damage-then-discard with the announced X for both numbers and
// Bedazzle's opponent-or-planeswalker target. The planeswalker's controller is
// read before the damage, so a planeswalker the damage kills still names its
// controller (CR 608.2h). The discard is that player's own choice. X = 0 is
// legal and does nothing. An illegal target makes the spell do nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "52d1d575-3e7f-4df1-8c89-aee0cac03aca",
		Name:         "Rakdos's Return",
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetOpponentOrPlaneswalker(),
		Purpose:      effects.ForTargets(effects.DamageXToTarget(0)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			targets := ctx.LegalTargets()
			if len(targets) == 0 {
				return nil
			}
			t := targets[0]
			victim := t.ID
			if t.Kind == game.TargetCard {
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok {
					return nil
				}
				victim = c.Controller
			}
			x := ctx.X()
			if x <= 0 {
				return nil
			}
			if err := (effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: x}).Apply(ctx); err != nil {
				return err
			}
			if victim == uuid.Nil {
				return nil
			}
			ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
				Player:   victim,
				Source:   item.SourceCardID,
				N:        x,
				Question: "Rakdos's Return — discard X cards",
			})
			return nil
		},
	})
}
