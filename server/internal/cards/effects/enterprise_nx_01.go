package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Enterprise NX-01 — Legendary Artifact — Spacecraft {3}{U}, 4/3 (Star
// Trek, #4973):
//
//	"When Enterprise NX-01 enters, tap target creature an opponent
//	 controls and put a stun counter on it.
//	 Station (Tap another creature you control: Put charge counters equal
//	 to its power on this Spacecraft. Station only as a sorcery. It's an
//	 artifact creature at 6+.)
//	 6+ | Flying, vigilance"
//
// Shipwreck Patrol's tap-and-stun enters trigger on Galvanizing Sawship's
// station frame: Station(), SpacecraftAt(6, 4, 3) and a gated keyword
// grant of flying and vigilance.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "947729a8-f8b0-4e08-9a87-6602af0e3c0a",
		Name:         "Enterprise NX-01",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Enterprise NX-01 — tap target creature an opponent controls and put a stun counter on it", tapTargetAndStunIt),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
		Activated: []ActivatedAbility{Station()},
		Static: append(
			SpacecraftAt(6, 4, 3),
			ThresholdKeywords(6, "flying", "vigilance"),
		),
	})
}
