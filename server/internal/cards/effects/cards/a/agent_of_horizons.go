package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Agent of Horizons — Creature — Human Rogue {2}{G}, 3/2 (Theros, #5561):
//
//	"{2}{U}: This creature can't be blocked this turn."
//
// The restriction lasts until end of turn and lands only on the same
// object that activated it (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "668a6a9d-80fc-437b-99b8-6807e92eabca",
		Name:         "Agent of Horizons",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}: This creature can't be blocked this turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{2}{U}"),
			Effect:  agentOfHorizonsUnblockable,
		}},
	})
}

// agentOfHorizonsUnblockable is the ability's body: nothing if the
// source has left or come back as a new object.
func agentOfHorizonsUnblockable(g *game.Game, item *game.StackItem) error {
	if !effects.SourceIsStillThisPermanent(g, item) {
		return nil
	}
	return effects.RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked,
		Label: "Agent of Horizons — can't be blocked this turn"}.Apply(effects.NewContext(g, item))
}
