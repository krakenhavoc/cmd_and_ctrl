package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rekindled Flame — {2}{R}{R} Sorcery:
//
//	"Rekindled Flame deals 4 damage to any target.
//	 At the beginning of your upkeep, if an opponent has no cards in hand, you
//	 may return this card from your graveyard to your hand."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dd99061d-510d-4935-b44e-54721505515f",
		Name:         "Rekindled Flame",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 4)),
		OnResolve:    effects.DamageToFirstTargetStep(4),
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.InGraveyard(effects.InterveningIfYou("an opponent has no cards in hand", effects.S82OldB524NoOpponentHasCardsInHand,
				effects.AtYourUpkeep("Rekindled Flame — return it from your graveyard to your hand", func(g *game.Game, item *game.StackItem) error {
					return effects.ReturnFromGraveyard{Target: item.SourceCardID, Dest: game.ZoneHand}.Apply(effects.NewContext(g, item))
				}))), "Rekindled Flame — return it to your hand?"),
		},
	})
}
