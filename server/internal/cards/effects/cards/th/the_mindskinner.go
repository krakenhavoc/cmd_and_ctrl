package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Mindskinner — Legendary Enchantment Creature — Nightmare {U}{U}{U}, 10/1:
//
//	"The Mindskinner can't be blocked.
//	 If a source you control would deal damage to an opponent, prevent that damage and each opponent mills that many cards."
//
// ADR 0108 §8 (#1906): a prevention static whose additional effect runs
// once per SOURCE in a damage instance (PreventDamageASourceWouldDeal),
// with "that many" the damage it was applied to — so damage that can't be
// prevented is dealt and each opponent still mills that many (CR 615.12).
// With two Mindskinners, the first prevents the damage and the second
// has nothing left to apply to: each opponent mills once (the ruling).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2e3ec514-fc9e-4964-a493-2a6cc9e63630",
		Name:         "The Mindskinner",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBeBlocked)},
		Replacements: []game.ReplacementEffect{
			effects.PreventDamageASourceWouldDeal(effects.PreventionStatic{
				To:    effects.ToAnOpponent,
				From:  effects.FromASourceYouControl(),
				Then:  mindskinnerMillBody,
				Label: "The Mindskinner — prevent the damage; each opponent mills that many cards",
			}),
		},
	})
}

// The additional effect: each opponent mills "that many".
var mindskinnerMillBody = game.DelayedBody("the-mindskinner/each-opponent-mills", mindskinnerMill)

func mindskinnerMill(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
	n := effects.ThatDamage(item)
	if n <= 0 {
		return nil
	}
	// One opponent's library at a time in APNAP order (#4266).
	return effects.MillEachPlayerInTurn(effects.NewContext(g, item), effects.B27OtherOpponents(g, item.Controller, uuid.Nil), effects.ConstantMill(n), game.KeyedTail{})
}
