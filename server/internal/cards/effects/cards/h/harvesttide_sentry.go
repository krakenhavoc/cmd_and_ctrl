package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Harvesttide Sentry — Creature — Human Warrior {1}{G}, 3/1 (Midnight
// Hunt, #5497):
//
//	"Coven — At the beginning of combat on your turn, if you control
//	 three or more creatures with different powers, this creature can't be
//	 blocked by creatures with power 2 or less this turn."
//
// Cavern Stomper's block rule (CantBeBlockedThisTurnByPower), pinned to
// this object and swept at the turn's end; a Sentry that left and came
// back is a new object and gets nothing (CR 400.7).
//
// No simplification.
func init() {
	const label = "Harvesttide Sentry — Coven: can't be blocked by creatures with power 2 or less this turn"
	effects.Register(effects.Spec{
		OracleID:     "6cd6b56d-d69e-4144-a6c7-8c066629435b",
		Name:         "Harvesttide Sentry",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.MidB05AtBeginningOfCombatIfCoven(label, func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.CantBeBlockedThisTurnByPower{
					Target: item.SourceCardID,
					AtMost: 2,
					Label:  "Harvesttide Sentry",
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
