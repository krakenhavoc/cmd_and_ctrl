package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "6cd6b56d-d69e-4144-a6c7-8c066629435b",
		Name:         "Harvesttide Sentry",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			midB05AtBeginningOfCombatIfCoven(label, func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				return CantBeBlockedThisTurnByPower{
					Target: item.SourceCardID,
					AtMost: 2,
					Label:  "Harvesttide Sentry",
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
