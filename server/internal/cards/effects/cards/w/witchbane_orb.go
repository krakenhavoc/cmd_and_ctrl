package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Witchbane Orb — {4} Artifact (Innistrad, S82 OLD b447, #5572):
//
//	"When this artifact enters, destroy all Curses attached to you.
//	 You have hexproof."
//
// The enter trigger destroys the Curses attached to its controller as it
// resolves (Fraying Sanity's attachment to a player, CR 303.4); the hexproof is
// a player keyword, as Sigarda, Heron's Grace has.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "7411a8f4-b736-482f-886e-5fa671e94e0d",
		Name:           "Witchbane Orb",
		Completeness:   effects.CompletenessFull,
		PlayerKeywords: []string{effects.KeywordHexproof},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Witchbane Orb — destroy all Curses attached to you",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					me := item.Controller
					return effects.DestroyAllMatching{Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
						return c.HasSubtype("Curse") && c.AttachedTo.Kind == game.TargetPlayer && c.AttachedTo.ID == me
					}}.Apply(ctx)
				}),
		},
	})
}
