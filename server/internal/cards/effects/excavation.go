package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Excavation — Enchantment {1}{U}:
//
//	"{1}, Sacrifice a land: Draw a card. Any player may activate this
//	 ability."
//
// An any-player row (CR 602.2, 602.1b). The whole cost is the
// activator's (CR 602.1a): the {1} from their own pool, and the land
// one THEY control, since a player can sacrifice only a permanent they
// control (CR 701.21a). "Draw a card" is "you", the activator (CR
// 109.5, 113.8).
//
// Purpose {Draws: 1}: the effect plainly helps whoever activates it, so
// the bot may use another player's Excavation (ADR 0106 owner decision
// 2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9f7b2ec6-a828-499b-a2ac-766bdaaf30c1",
		Name:         "Excavation",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{1}, Sacrifice a land: Draw a card. Any player may activate this ability.",
			Cost:      Plus(ManaCost("{1}"), b08SacrificeALand()),
			AnyPlayer: true,
			Purpose:   game.ActivationPurpose{Draws: 1},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
