package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Triarch Stalker — Artifact Creature — Necron {3}{B}{B}, 4/5:
//
//	"Targeting Relay — At the beginning of combat on your turn, choose an
//	 opponent. Creatures attacking the last chosen player have menace."
//
// The chosen opponent is kept on the permanent (Card.ChosenPlayer), so it is
// "the last chosen player" until the next combat's choice replaces it. The
// menace goes to every creature attacking that player, whoever controls it
// (not their planeswalkers: attacking a player is attacking the player).
var triarchStalkerChosen = effects.ChoosePlayerThen("option-pick/triarch-stalker-opponent",
	func(ctx *effects.Context, chosen uuid.UUID, _ []uuid.UUID) error {
		if chosen == uuid.Nil {
			return nil
		}
		return ctx.Game.SetChosenPlayerForEffect(ctx.Item.SourceCardID, chosen)
	})

func init() {
	menace := effects.B16GrantKeywords(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return source.ChosenPlayer != uuid.Nil && target.IsCreature() && target.AttackingTarget == source.ChosenPlayer
	}, "menace")
	// Whether a creature is attacking is not on the battlefield's shape: say
	// so, or the cached resolution keeps the menace a declaration ago.
	menace.DependsOnAttackingStatus = true
	effects.Register(effects.Spec{
		OracleID:     "1b3ae9c3-1940-4fb8-a47c-2d9e1790f4a3",
		Name:         "Triarch Stalker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfYourCombat("Triarch Stalker — choose an opponent", func(g *game.Game, item *game.StackItem) error {
				return effects.ChoosePlayer{
					Among:    effects.Opponents,
					Question: "Triarch Stalker — choose an opponent: creatures attacking them have menace",
					ThenKey:  triarchStalkerChosen,
				}.Apply(effects.NewContext(g, item))
			}),
		},
		Static: []game.StaticAbility{menace},
	})
}
