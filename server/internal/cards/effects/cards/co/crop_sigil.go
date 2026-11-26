package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crop Sigil — Enchantment {G} (Eldritch Moon, S82 OLD, #5541):
//
//	"At the beginning of your upkeep, you may mill a card.
//	 Delirium — {2}{G}, Sacrifice this enchantment: Return up to one target
//	 creature card and up to one target land card from your graveyard to
//	 your hand. Activate only if there are four or more card types among
//	 cards in your graveyard."
//
// Patchwork Beastie's optional upkeep mill. The activation is Badlands
// Revival's two "up to one" clauses read by slot, behind the delirium
// count checked as the ability is activated (CR 207.2c, 602.5b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "76ee013d-5293-4432-b425-649a2e82bd74",
		Name:         "Crop Sigil",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.AtYourUpkeep("Crop Sigil — mill a card", effects.Do(effects.MillCards{N: 1})),
				"Crop Sigil — mill a card?"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Delirium — {2}{G}, Sacrifice this enchantment: Return up to one target creature card and up to one target land card from your graveyard to your hand. Activate only if there are four or more card types among cards in your graveyard.",
			Cost:    effects.Plus(effects.ManaCost("{2}{G}"), effects.SacrificeThis()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Targets: effects.Clauses(
				effects.TargetCardInGraveyard("up to one target creature card from your graveyard", effects.YouOwn(), effects.Creature()).WithCount(0, 1),
				effects.TargetCardInGraveyard("up to one target land card from your graveyard", effects.YouOwn(), effects.Land()).WithCount(0, 1),
			),
			Condition: effects.S82OldB275DeliriumActivation,
			Effect: effects.Do(
				effects.Instruction(func(ctx *effects.Context) error {
					return effects.ReturnClauseCardFromGraveyard(ctx, 0, game.ZoneHand)
				}),
				effects.Instruction(func(ctx *effects.Context) error {
					return effects.ReturnClauseCardFromGraveyard(ctx, 1, game.ZoneHand)
				}),
			),
		}},
	})
}
