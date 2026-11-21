package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nimbus Champion — Creature — Avatar Warrior {5}{U}, 4/4 (Battlebond,
// #5529):
//
//	"Flying
//	 Whenever this creature attacks, you may return target creature to its
//	 owner's hand if that creature's power is less than or equal to the
//	 number of Warriors your team controls."
//
// Flying rides PrintedKeywords. "Your team" is "you" (CR 102.4). The target
// is any creature, chosen as the trigger goes on the stack; the "if" is not
// an intervening one (it does not follow the trigger condition), so it is
// read only as the trigger resolves (CR 608.2h): the target's power then
// against the Warrior permanents its controller controls then, the
// Champion included if it is still there. The "you may" is the trigger's
// own yes/no.
//
// No simplification.
func init() {
	attack := effects.Targeting(effects.WheneverThisAttacks("Nimbus Champion — return target creature to its owner's hand if its power is no greater than the number of Warriors you control",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			id := effects.FirstLegalBattlefieldTarget(ctx)
			if id == uuid.Nil {
				return nil
			}
			c, ok := g.LookupCardForEffect(id)
			if !ok || c.CurrentPower() > effects.B26PermanentsOfSubtypeControlled(g, item.Controller, ctx.TextSubtype("Warrior")) {
				return nil
			}
			return effects.BounceToHand{Target: id}.Apply(ctx)
		}), effects.TargetCreature("target creature"))
	attack.OptionalPrompt = &game.TriggerOptionalPrompt{Question: "Nimbus Champion — return target creature to its owner's hand?"}
	effects.Register(effects.Spec{
		OracleID:        "3a536065-d62a-44c9-886b-84209b6bce57",
		Name:            "Nimbus Champion",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{attack},
	})
}
