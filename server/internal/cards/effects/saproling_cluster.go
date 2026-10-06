package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Saproling Cluster — Enchantment {1}{G}:
//
//	"{1}, Discard a card: Create a 1/1 green Saproling creature token.
//	 Any player may activate this ability."
//
// An any-player row (CR 602.2, 602.1b). The whole cost is the
// activator's (CR 602.1a): the {1} from their own pool and a card from
// their own hand (CR 701.9a). "Create" is "you", so the token enters
// under the activator's control (CR 109.5, 111.2), not the
// enchantment's controller's.
//
// No Purpose: the bot prices another player's row by the cards it
// draws and its controller's life loss alone, and this row has
// neither, so the bot does not reach across the table for it (ADR 0106
// owner decision 2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3b64051e-d320-4991-bc9e-6e1435a23c17",
		Name:         "Saproling Cluster",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{1}, Discard a card: Create a 1/1 green Saproling creature token. Any player may activate this ability.",
			Cost:      Plus(ManaCost("{1}"), DiscardACard()),
			AnyPlayer: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateToken{Controller: item.Controller, Template: b11GreenSaprolingToken(), N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
