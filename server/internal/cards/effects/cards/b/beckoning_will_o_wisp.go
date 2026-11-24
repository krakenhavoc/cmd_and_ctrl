package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Beckoning Will-o'-Wisp — Creature — Spirit {2}{W}, 1/3:
//
//	"Flying
//	 Lure the Unwary — At the beginning of combat on your turn, choose an
//	 opponent.
//	 Creatures attacking the last chosen player get +1/+0."
//
// The chosen opponent is kept on the permanent (Card.ChosenPlayer), so it
// is "the last chosen player" until the next combat's choice replaces it;
// one that has left the battlefield remembers nobody (CR 400.7). The +1/+0
// goes to every creature attacking that player, whoever controls it.
var beckoningWillOWispChosen = effects.ChoosePlayerThen("option-pick/beckoning-will-o-wisp-opponent",
	func(ctx *effects.Context, chosen uuid.UUID, _ []uuid.UUID) error {
		if chosen == uuid.Nil {
			return nil
		}
		return ctx.Game.SetChosenPlayerForEffect(ctx.Item.SourceCardID, chosen)
	})

func init() {
	boost := effects.B16Anthem(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return source.ChosenPlayer != uuid.Nil && target.IsCreature() && target.AttackingTarget == source.ChosenPlayer
	}, 1, 0)
	boost.Label = "Creatures attacking the last chosen player get +1/+0."
	// Whether a creature is attacking is not on the battlefield's shape.
	boost.DependsOnAttackingStatus = true
	effects.Register(effects.Spec{
		OracleID:        "d39df714-bd0b-4de8-b9df-55713bab7f35",
		Name:            "Beckoning Will-o'-Wisp",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfYourCombat("Lure the Unwary — choose an opponent", func(g *game.Game, item *game.StackItem) error {
				return effects.ChoosePlayer{
					Among:    effects.Opponents,
					Question: "Beckoning Will-o'-Wisp — choose an opponent: creatures attacking them get +1/+0",
					ThenKey:  beckoningWillOWispChosen,
				}.Apply(effects.NewContext(g, item))
			}),
		},
		Static: []game.StaticAbility{boost},
	})
}
