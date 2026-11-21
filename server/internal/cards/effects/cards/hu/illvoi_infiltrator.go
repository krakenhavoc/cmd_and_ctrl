package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Illvoi Infiltrator — Creature — Jellyfish Rogue {2}{U}, 1/3 (Edge of
// Eternities, #4461):
//
//	"This creature can't be blocked if you've cast two or more spells
//	 this turn.
//	 Whenever this creature deals combat damage to a player, draw a card."
//
// Nimble Brigand's shape with Brightspear Zealot's condition: the block
// rule reads your cast tally for the turn as blockers are declared, so
// two spells cast before blocks make it unblockable, and a spell cast
// after blocks changes nothing (CR 509.1b). The draw is the shared
// combat-damage trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3084377f-73fb-4b00-9dd9-3844b1ecafc2",
		Name:         "Illvoi Infiltrator",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedWhile(effects.OnSelf(), func(g *game.Game, you uuid.UUID, _ game.Card) bool {
				return g.CastTallyFor(you).Total >= 2
			}),
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Illvoi Infiltrator — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
