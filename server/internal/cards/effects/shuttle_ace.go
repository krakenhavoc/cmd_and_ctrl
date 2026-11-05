package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shuttle Ace — Creature — Kelpien Pilot {4}{W}, 4/4 (Star Trek, #4973):
//
//	"During your turn, this creature has flying.
//	 Whenever this creature becomes tapped, you gain 2 life."
//
// The flying is Cid's layer 6 grant on itself, on only during its
// controller's turn. "Becomes tapped" is any tap, an attack included
// (ThisBecameTapped); a vigilance attack does not tap it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4b5b0864-7ddd-497d-a7d6-a31c49bb9c7b",
		Name:         "Shuttle Ace",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return thisPermanent(target, g, source) && isActivePlayer(g, source.Controller)
			}, "flying"),
		},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped,
				"Shuttle Ace — you gain 2 life", Do(GainLife{Amount: 2})),
		},
	})
}
