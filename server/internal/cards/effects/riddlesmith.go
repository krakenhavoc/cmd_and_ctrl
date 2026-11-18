package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Riddlesmith — Creature — Human Artificer {1}{U}, 2/1 (Reprint, #5908):
//
//	"Whenever you cast an artifact spell, you may draw a card. If you do,
//	 discard a card."
//
// The cast watcher is Forensic Gadgeteer's over artifact spells. The "may
// draw a card. If you do, discard a card" is one linked clause asked as the
// ability resolves (Stadium Tidalmage's MayChoice and celestusAnswer):
// declining the draw declines the discard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e3b9c9e3-5a35-4dd2-ab17-d1b6a140b185",
		Name:         "Riddlesmith",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Artifact(), "Riddlesmith — you may draw a card; if you do, discard a card",
				func(g *game.Game, item *game.StackItem) error {
					return MayChoice{
						Question: "Riddlesmith — draw a card, then discard a card?",
						Draws:    1,
						YesLabel: "Draw",
						NoLabel:  "Decline",
						Then:     celestusAnswer,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
