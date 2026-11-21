package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Furor of the Bitten — #4380, S76. Uses the existing catalog seams.
func init() {
	effects.Register(effects.Spec{
		OracleID: "40be0923-5d57-46fd-84ef-7f1a1fd5da42", Name: "Furor of the Bitten", Completeness: effects.CompletenessFull,
		Targets: effects.EnchantCreature(), Purpose: game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{effects.PumpAttached(2, 2), effects.AttacksEachCombatWhere(func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return source.AttachedTo.Kind == game.TargetCard && source.AttachedTo.ID == target.InstanceID
		})},
	})
}
