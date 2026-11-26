package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sage-Eye Avengers — {4}{U}{U} Creature — Djinn Monk 4/5 (Fate Reforged):
//
//	"Prowess
//	 Whenever this creature attacks, you may return target creature to its
//	 owner's hand if its power is less than this creature's power."
//
// Nimbus Champion's shape: the "you may" is the trigger's own yes/no and the
// power comparison is not an intervening "if", so it is read only as the
// trigger resolves (CR 608.2h): the target's power then against this creature's
// power then, or as it last existed if it has gone. Prowess (CR 702.108) is the
// engine's keyword.
func init() {
	attack := effects.Targeting(effects.WheneverThisAttacks("Sage-Eye Avengers — return target creature to its owner's hand if its power is less than this creature's power",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			id := effects.FirstLegalBattlefieldTarget(ctx)
			if id == uuid.Nil {
				return nil
			}
			c, ok := g.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			g.RecomputeLayersIfStaleLocked()
			power := effects.DepartedCreaturePower(g, item.SourceCardID)
			if src, ok := g.LookupCardForEffect(item.SourceCardID); ok && effects.OnBattlefield(g, item.SourceCardID) {
				power = src.CurrentPower()
			}
			if c.CurrentPower() >= power {
				return nil
			}
			return effects.BounceToHand{Target: id}.Apply(ctx)
		}), effects.TargetCreature("target creature"))
	attack.OptionalPrompt = &game.TriggerOptionalPrompt{Question: "Sage-Eye Avengers — return target creature to its owner's hand?"}
	effects.Register(effects.Spec{
		OracleID:        "97a1b49e-d6bc-46ba-88c1-991057e64170",
		Name:            "Sage-Eye Avengers",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Triggered:       []game.TriggeredAbility{attack},
	})
}
