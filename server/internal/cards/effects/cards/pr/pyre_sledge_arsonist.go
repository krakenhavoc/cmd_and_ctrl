package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pyre-Sledge Arsonist — Creature — Lizard Shaman {2}{R}, 2/2 (Streets of
// New Capenna, S82):
//
//	"{1}, {T}: This creature deals X damage to any target, where X is the
//	 number of permanents you've sacrificed this turn."
//
// X is the turn's sacrifice tally for the ability's controller
// (TurnTally.PermanentsSacrificed, Obsessive Pursuit's read): tokens,
// Treasures and creatures all count, an opponent's sacrifices do not, and a
// permanent sacrificed to pay this very ability counts. It is read once as
// the ability resolves (CR 608.2h); at 0 nothing is dealt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7f42a05c-9e6d-4cbc-92f1-4433cd22f149",
		Name:         "Pyre-Sledge Arsonist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: This creature deals X damage to any target, where X is the number of permanents you've sacrificed this turn.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets: effects.TargetAny(),
			Effect: effects.SourceDealsComputedDamageToEachLegalTarget(func(ctx *effects.Context) int {
				return ctx.Game.TurnTallyFor(ctx.Controller()).PermanentsSacrificed
			}),
		}},
	})
}
