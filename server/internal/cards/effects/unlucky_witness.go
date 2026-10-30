package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unlucky Witness — Creature — Human Citizen {R}, 1/1 (Marvel Super Heroes
// Commander, #4455):
//
//	"When this creature dies, exile the top two cards of your library.
//	 Until your next end step, you may play one of those cards."
//
// The two cards are exiled and held by one play permission (CR 611.2c
// locks the set) with a budget of ONE (CastPermission.CastsLeft, Kotose's
// shape): the first cast or land play out of them spends it and the other
// card stays exiled, unplayable. Play, not cast, so a land among them can
// be played with the turn's land drop. The window is Ob Nixilis'
// UntilYourNextEndStep: from your own turn before its end step it closes
// as this turn's end step begins, otherwise as your next turn's does. The
// controller is the one who controlled the Witness as it died.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8ff7c8af-de9a-45f8-a9ba-98209a746891",
		Name:         "Unlucky Witness",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Unlucky Witness — exile the top two cards of your library; until your next end step you may play one of them",
				func(g *game.Game, item *game.StackItem) error {
					_, err := g.ExileTopWithPermissionForEffect(item.Controller, item.Controller, 2, game.CastPermission{
						CastsLeft:  1,
						Duration:   g.UntilYourNextEndStepDuration(item.Controller),
						Source:     item.SourceCardID,
						SourceName: "Unlucky Witness",
						Label:      "Play one of the cards exiled by Unlucky Witness",
					})
					return err
				}),
		},
	})
}
