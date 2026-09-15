package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wavebreak Hippocamp — Enchantment Creature — Horse Fish {2}{U}, 2/2
// (EDHREC rank 2398):
//
//	"Whenever you cast your first spell during each opponent's turn,
//	 draw a card."
//
// The flash deck's card-draw engine. "Your first spell this turn" is
// the engine's per-turn cast tally, bumped before EventCast fires,
// so the condition is Total == 1 on the controller's own cast; "an
// opponent's turn" is the active seat not being the controller. A
// spell cast on the controller's own turn never triggers, and the
// second spell on an opponent's turn is silent.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3405c8a9-a8d6-4b45-9b64-94141076603b",
		Name:         "Wavebreak Hippocamp",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b22FirstSpellOnAnOpponentsTurn(ev, source, g)
			}, "Wavebreak Hippocamp — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
