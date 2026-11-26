package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Graveyard Shovel — {2} Artifact (Innistrad, S82 OLD b447, #5572):
//
//	"{2}, {T}: Target player exiles a card from their graveyard. If it's a
//	 creature card, you gain 2 life."
//
// Scrabbling Claws' first ability with a rider: the target player chooses the
// card from their own graveyard as the ability resolves, and the life is
// gained only when the card exiled was a creature card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "868b5315-da38-4918-b8e4-70762811605b",
		Name:         "Graveyard Shovel",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Target player exiles a card from their graveyard. If it's a creature card, you gain 2 life.",
			Purpose: game.Purpose{Answers: game.AnswerGraveyard},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target player"),
			Effect:  effects.OldB447TargetExilesCardGainIfCreature,
		}},
	})
}
