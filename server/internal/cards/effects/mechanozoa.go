package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mechanozoa — Artifact Creature — Robot Jellyfish {4}{U}{U}, 5/5 (Edge of
// Eternities, #4461):
//
//	"When this creature enters, tap target artifact or creature an
//	 opponent controls and put a stun counter on it. (If a permanent with
//	 a stun counter would become untapped, remove one from it instead.)
//	 Warp {2}{U} (You may cast this card from your hand for its warp cost.
//	 Exile this creature at the beginning of the next end step, then you
//	 may cast it from exile on a later turn.)"
//
// Fractal Mascot's entry trigger (tapTargetAndStunIt) over an artifact or
// a creature an opponent controls. The stun counter is the engine's
// CR 122.1d replacement on untapping. Warp is the shared alternative
// cost, so a warped Mechanozoa still locks a blocker down for a turn
// cycle before it is exiled at the end step.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "1aeb43f7-4d8a-4648-9c13-fd101081b448",
		Name:             "Mechanozoa",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Warp("{2}{U}")},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Mechanozoa — tap target artifact or creature an opponent controls and put a stun counter on it", tapTargetAndStunIt),
				TargetPermanent("target artifact or creature an opponent controls", Or(Artifact(), Creature()), OpponentControls())),
		},
	})
}
