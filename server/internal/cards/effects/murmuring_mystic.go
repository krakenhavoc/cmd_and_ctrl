package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Murmuring Mystic — Creature — Human Wizard {3}{U}, 1/5 (EDHREC
// rank 1816):
//
//	"Whenever you cast an instant or sorcery spell, create a 1/1 blue
//	 Bird Illusion creature token with flying."
//
// Talrand on a wall. The condition is b12InstantOrSorceryCastByYou
// and the token is b16BlueBirdIllusionToken, whose flying is real
// (carried on its Keywords). The trigger goes on the stack above the
// spell that caused it, so the Bird is on the battlefield before the
// spell resolves, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dcd4da46-5438-4454-8b1b-43ca51bda1f9",
		Name:         "Murmuring Mystic",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b12InstantOrSorceryCastByYou(ev, source, g)
			}, "Murmuring Mystic — create a 1/1 Bird Illusion with flying", Do(CreateToken{Template: b16BlueBirdIllusionToken(), N: 1})),
		},
	})
}
