package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Urborg Mindsucker — Creature — Horror {2}{B}, 2/2 (Multiverse Gift Box,
// #5632):
//
//	"{B}, Sacrifice this creature: Target opponent discards a card at
//	 random. Activate only as a sorcery."
//
// The Horror is sacrificed as the cost while activating (CR 602.2b,
// 601.2h), so the ability resolves without it. The discard is the engine's random one, drawn from the
// keyed random stream, the same Tourach's Canticle's second discard uses;
// an empty hand discards nothing. A target opponent who has left the game
// in response is illegal and the ability does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "19d11ca8-a95b-4e6e-a71f-8f9d06c580fd",
		Name:         "Urborg Mindsucker",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{B}, Sacrifice this creature: Target opponent discards a card at random. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{B}"), SacrificeThis()),
			Targets:      TargetPlayer("target opponent", Opponent()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				for _, t := range NewContext(g, item).LegalTargets() {
					if t.Kind == game.TargetPlayer {
						return g.DiscardRandomForEffect(t.ID, 1)
					}
				}
				return nil
			},
		}},
	})
}
