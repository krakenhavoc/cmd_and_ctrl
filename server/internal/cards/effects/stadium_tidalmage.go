package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stadium Tidalmage — Creature — Djinn Sorcerer {2}{U}{R}, 4/4 (Secrets of
// Strixhaven, #4456):
//
//	"Whenever this creature enters or attacks, you may draw a card. If you
//	 do, discard a card."
//
// One ability with two trigger conditions (WhenThisEntersOrAttacks, as
// Agrus Kos). The "you may draw a card. If you do, discard a card" is one
// linked clause (CR 607.2) asked as the ability resolves, the Celestus's
// MayChoice and keyed answer (celestusAnswer); declining the draw declines
// the discard, and a draw a limit stops is not offered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1c45e872-3af8-47a4-9dc7-a99c293338f8",
		Name:         "Stadium Tidalmage",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Stadium Tidalmage — you may draw a card; if you do, discard a card",
				func(g *game.Game, item *game.StackItem) error {
					return MayChoice{
						Question: "Stadium Tidalmage — draw a card, then discard a card?",
						Draws:    1,
						YesLabel: "Draw",
						NoLabel:  "Decline",
						Then:     celestusAnswer,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
