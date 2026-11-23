package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Flamerush Rider — Creature — Human Warrior {4}{R}, 3/3:
//
//	"Whenever this creature attacks, create a token that's a copy of
//	 another target attacking creature and that's tapped and attacking.
//	 Exile the token at end of combat.
//	 Dash {2}{R}{R}"
//
// Dash is the alternative cost (Ragavan's). The target is chosen as the
// trigger goes on the stack ("another" excludes the Rider by object, and
// "attacking" is read then and again as it resolves). The token copies the
// target's copiable values, from last-known information if it left (CR
// 707.2, 608.2h), and enters tapped and attacking a player or planeswalker
// of the controller's choice (CR 508.4), so it was never declared as an
// attacker and triggers nothing that watches attacks. "Exile the token at
// end of combat" is a delayed trigger at the beginning of that step naming
// the token (CR 603.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "f797a325-ec2a-4142-b0ff-90cac1356da9",
		Name:             "Flamerush Rider",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Dash("{2}{R}{R}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisAttacks("Flamerush Rider — create a tapped and attacking token copy of another target attacking creature; exile it at end of combat",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					tmpl, ok := effects.TokenCopyTemplate(g, id)
					if !ok {
						return nil
					}
					return effects.CreateTokensAttackingYourChoice{
						Template: tmpl,
						Copy:     true,
						N:        1,
						Tapped:   true,
						Prefer:   effects.TriggerAttackTarget(item),
						Label:    "Flamerush Rider",
						Tail:     ctx.ItemTail(flamerushRiderRest),
					}.Apply(ctx)
				}),
				effects.Another(effects.TargetCreature("another target attacking creature", effects.AttackingCreature()))),
		},
	})
}

// flamerushRiderRest is "Exile the token at end of combat." as a registered
// tail told the tokens that landed. An on-disk key: never renamed, never
// reused.
var flamerushRiderRest = effects.AttackingTokensThen("attacking-tokens/flamerush-rider-exile-at-end-of-combat",
	func(ctx *effects.Context, created []uuid.UUID, tail game.KeyedTail) error {
		item := tail.Asked.StackItem()
		if len(created) == 0 || item == nil {
			return nil
		}
		return effects.ScheduleDelayedTrigger{
			At:    game.StepEndCombat,
			Label: "Flamerush Rider — exile the token",
			Cards: created,
			Body:  effects.ExileListedCardsBody,
		}.Apply(effects.NewContext(ctx.Game, item))
	})
