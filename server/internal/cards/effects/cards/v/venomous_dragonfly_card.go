package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Venomous Dragonfly — Creature — Insect {3}{G}, 1/1 (Mercadian Masques, S82
// OLD b684, #5619):
//
//	"Flying
//	 Whenever this creature blocks or becomes blocked by a creature, destroy
//	 that creature at end of combat."
//
// Flying is a printed keyword. Ashmouth Hound's trigger on the block pair that
// names the Dragonfly, whichever side it is on, with Deathgazer's delayed
// destruction at the end of combat step (CR 511.2) of the other creature of
// the pair, as the object it was then (CR 400.7).
//
// The file is named venomous_dragonfly_card.go because Go reads a file name ending
// in _dragonfly.go as a build constraint for the DragonFly BSD port and would
// leave the card out of every other build.
//
// No simplification.
func init() {
	const label = "Venomous Dragonfly — destroy that creature at end of combat"
	effects.Register(effects.Spec{
		OracleID:        "8e3e5264-d472-41ce-ad65-a53f8b24fc40",
		Name:            "Venomous Dragonfly",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.SkewerSlingerFoe(ev, source.InstanceID) != uuid.Nil
			}, label, effects.S82OldB614DestroyTheFoeAtEndOfCombat(label)),
		},
	})
}
