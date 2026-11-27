package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Animate Land — Instant {G} (Nemesis, S82 OLD batch 679, #5618):
//
//	"Until end of turn, target land becomes a 3/3 creature that's still a
//	 land."
//
// One layer-4 type change (it is a creature in addition to being a land) and
// a base power and toughness of 3/3 (layer 7b), until end of turn, on the
// land the target was as the spell resolved (CR 611.2c); the target is
// re-checked then (CR 608.2b). As a creature it can attack only if its
// controller has controlled it since the turn began (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b2bd58ea-b2c6-47f1-b082-7969f093bbc6",
		Name:         "Animate Land",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerAnimate},
		Targets:      effects.TargetPermanent("target land", effects.Land()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			land := effects.FirstLegalBattlefieldTarget(ctx)
			if land == uuid.Nil {
				return nil
			}
			return effects.BecomeCreatureUntilEOT{Target: land, Types: []string{"Creature"}, SetPower: 3, SetToughness: 3,
				Label: "Animate Land — target land becomes a 3/3 creature until end of turn"}.Apply(ctx)
		},
	})
}
