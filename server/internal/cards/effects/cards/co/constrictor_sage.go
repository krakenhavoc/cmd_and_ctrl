package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "abfb7234-c5cc-4e0b-9a67-d66473f918fc",
		Name:         "Constrictor Sage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Constrictor Sage — tap target creature an opponent controls and put a stun counter on it", effects.TapTargetAndStunIt),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
		Activated: []effects.ActivatedAbility{
			effects.Renew("{2}{U}", "Tap target creature an opponent controls and put a stun counter on it.",
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()), effects.TapTargetAndStunIt),
		},
	})
}
