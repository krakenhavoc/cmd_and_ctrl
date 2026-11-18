package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Saiba Trespassers — Artifact Creature — Moonfolk Rogue {4}{U}, 3/5 (Neon
// Dynasty, #5493):
//
//	"Channel — {3}{U}, Discard this card: Tap up to two target creatures you
//	 don't control. Those creatures don't untap during their controller's
//	 next untap step."
//
// Frost Breath's tap-and-freeze over the creatures the channeler doesn't
// control; zero targets is legal ("up to").
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d5cdfa86-a234-4ddd-bfa0-4a3bc51eebce",
		Name:         "Saiba Trespassers",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Channel — {3}{U}, Discard this card: Tap up to two target creatures you don't control. Those creatures don't untap during their controller's next untap step.",
			Purpose: game.Purpose{Answers: game.AnswerRestrict},
			Cost:    Plus(ManaCost("{3}{U}"), DiscardThis()),
			Zones:   []game.ZoneKind{game.ZoneHand},
			Targets: TargetCreature("up to two target creatures you don't control", Not(YouControl())).WithCount(0, 2),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return b751TapFreezeTargets(NewContext(g, item), uuid.Nil, "Saiba Trespassers")
			},
		}},
	})
}
