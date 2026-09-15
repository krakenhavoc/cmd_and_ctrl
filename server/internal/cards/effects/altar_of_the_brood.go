package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Altar of the Brood — Artifact {1} (EDHREC rank 1418):
//
//	"Whenever another permanent you control enters, each opponent
//	 mills a card."
//
// The one-mana mill engine every token combo ends on. "Another
// permanent you control" is the ETB condition every landfall and
// artifact-count card shares, with the Altar itself excluded — any
// permanent type, tokens included, so a Krenko activation mills the
// table once per Goblin. One event per permanent, one trigger each,
// which is the printed batching.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c3aafcdd-c890-4971-b8a9-5bfcad794c0b",
		Name:         "Altar of the Brood",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := enteredUnderYourControl(ev, source, g, true)
				return ok
			}, "Altar of the Brood — each opponent mills a card", func(g *game.Game, item *game.StackItem) error {
				return b12EachOpponentMills(g, item, 1)
			}),
		},
	})
}
