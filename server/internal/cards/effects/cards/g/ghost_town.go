package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghost Town — Land (Tempest, S82 OLD, #5628):
//
//	"{T}: Add {C}.
//	 {0}: Return this land to its owner's hand. Activate only if it's not your
//	 turn."
//
// A colourless mana ability and a free self-bounce, which uses the stack and so
// can be answered (Bramble Familiar's bounce). "Not your turn" is the active
// player not being the activator: any turn but its controller's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "f2c861d3-b302-4e84-b647-099551007269",
		Name:          "Ghost Town",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{effects.PainlessColorless()},
		Activated: []effects.ActivatedAbility{{
			Label:     "{0}: Return this land to its owner's hand. Activate only if it's not your turn.",
			Purpose:   game.Purpose{Answers: game.AnswerSurvive},
			Cost:      effects.ManaCost("{0}"),
			Condition: effects.DuringAnOpponentsTurn(),
			Effect:    effects.ReturnThisPermanentToOwnersHand,
		}},
	})
}
