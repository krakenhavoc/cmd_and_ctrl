package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ego Erasure — Kindred Instant — Shapeshifter {2}{U} (Lorwyn, S82 OLD b548,
// #5593):
//
//	"Changeling (This card is every creature type.)
//	 Creatures target player controls get -2/-0 and lose all creature types
//	 until end of turn."
//
// Shields of Velis Vel's shape: the creatures the player controls as the
// spell resolves, locked then (CR 611.2c), get -2/-0 and lose every creature
// type (layer 4, CR 613.1d; a changeling stops being every type). A creature
// they get afterwards is not affected. A target player who has left stops the
// spell (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a7a47969-b29a-4f99-9500-29c59af43c48",
		Name:            "Ego Erasure",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling},
		Targets:         effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			theirs := effects.And(effects.Creature(), effects.ControlledByPlayer(item.Targets[0].ID))
			if err := (effects.BoostUntilEOT{Match: theirs, Power: -2, Label: "Ego Erasure — -2/-0"}).Apply(ctx); err != nil {
				return err
			}
			return effects.UntilEndOfTurn(ctx, uuid.Nil, theirs, "Ego Erasure — lose all creature types", game.LoseAllCreatureTypesMod())
		},
	})
}
