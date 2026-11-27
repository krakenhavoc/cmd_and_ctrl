package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Exile into Darkness — Sorcery {4}{B} (Saviors of Kamigawa, #5601):
//
//	"Target player sacrifices a creature of their choice with mana value 3 or
//	 less.
//	 At the beginning of your upkeep, if you have more cards in hand than each
//	 opponent, you may return this card from your graveyard to your hand."
//
// An edict over the creatures of mana value 3 or less (a player with none
// sacrifices nothing, CR 701.21a). The upkeep ability is Death of a Thousand
// Stings's: it works from the graveyard (CR 113.6) and its "if" is an
// intervening if (CR 603.4).
//
// No simplification.
func init() {
	const name = "Exile into Darkness"
	effects.Register(effects.Spec{
		OracleID:     "bfa21cbc-8a04-4456-bf02-bae6b514d50a",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.OldB587TargetPlayerSacrificesASmallCreature(name+" — sacrifice a creature with mana value 3 or less")(ctx.Game, item)
		},
		Triggered: []game.TriggeredAbility{effects.OldB587ReturnThisCardToHandUpkeep()},
	})
}
