package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tazeem — Plane — Zendikar (Planechase Anthology, #4870):
//
//	"Creatures can't block.
//	 Whenever chaos ensues, draw a card for each land you control."
//
// A plane (ADR 0216 §6). The static is Bedlam's block rule, read from the
// command zone while the plane is face up (CR 901.7); the chaos ability
// counts the planar controller's lands as it resolves (CR 311.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "97f708de-9271-4137-a9fe-dda25b922900",
		Name:         "Tazeem",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBlockAttackers(OnMatching(Creature()), OnMatching(Creature()), "Tazeem: creatures can't block"),
		},
		Triggered: []game.TriggeredAbility{
			WheneverChaosEnsues("Tazeem — draw a card for each land you control",
				func(g *game.Game, item *game.StackItem) error {
					return DrawCards{N: b10LandsControlled(g, item.Controller)}.Apply(NewContext(g, item))
				}),
		},
	})
}
