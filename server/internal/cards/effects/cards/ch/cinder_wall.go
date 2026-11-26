package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cinder Wall — Creature — Wall {R}, 3/3 (Conspiracy, S82 OLD batch 364,
// #5557):
//
//	"Defender
//	 When this creature blocks, destroy it at end of combat."
//
// Runaway Carriage's block trigger (SelfBlocksOnce: one trigger however many
// attackers it blocks) over a delayed destroy at the beginning of the end of
// combat step (CR 511.2). The destroy is skipped if the Wall has left the
// battlefield or come back as a new object by then (CR 400.7); a Wall that
// regenerates or is indestructible stays.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "16128b90-9b81-476e-8999-3ef498728c39",
		Name:            "Cinder Wall",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.SelfBlocksOnce(ev, source)
			}, "Cinder Wall — destroy it at end of combat",
				effects.S82OldB364DestroyThisAtEndOfCombat("Cinder Wall — destroy it")),
		},
	})
}
