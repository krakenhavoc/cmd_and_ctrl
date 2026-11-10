package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Constrictor Sage — Creature — Snake Wizard {4}{U}, 4/4 (Tarkir:
// Dragonstorm, #4463):
//
//	"When this creature enters, tap target creature an opponent controls
//	 and put a stun counter on it.
//	 Renew — {2}{U}, Exile this card from your graveyard: Tap target
//	 creature an opponent controls and put a stun counter on it. Activate
//	 only as a sorcery."
//
// Both halves are Shipwreck Patrol's tap-and-stun body (tapTargetAndStunIt);
// the second rides Renew (renew.go), which works from the graveyard and
// exiles the card as its cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "abfb7234-c5cc-4e0b-9a67-d66473f918fc",
		Name:         "Constrictor Sage",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Constrictor Sage — tap target creature an opponent controls and put a stun counter on it", tapTargetAndStunIt),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
		Activated: []ActivatedAbility{
			Renew("{2}{U}", "Tap target creature an opponent controls and put a stun counter on it.",
				TargetCreature("target creature an opponent controls", OpponentControls()), tapTargetAndStunIt),
		},
	})
}
