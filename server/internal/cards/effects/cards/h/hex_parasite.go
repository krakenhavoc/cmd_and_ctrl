package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hex Parasite — Artifact Creature — Phyrexian Insect {1}, 1/1:
//
//	"{X}{B/P}: Remove up to X counters from target permanent. For each
//	 counter removed this way, this creature gets +1/+0 until end of turn.
//	 ({B/P} can be paid with either {B} or 2 life.)"
//
// The player takes counters one kind at a time (HexParasiteRemoveCounters),
// never more than X; the bonus is the number that really came off.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c180f21b-f553-4839-80c7-a3ce2fae9b5a",
		Name:         "Hex Parasite",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}{B/P}: Remove up to X counters from target permanent. For each counter removed this way, this creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerRemove | game.AnswerPump},
			Cost:    effects.ManaCost("{X}{B/P}"),
			Targets: effects.TargetPermanent("target permanent"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.HexParasiteRemoveCounters(g, item, ctx.X())
			},
		}},
	})
}
