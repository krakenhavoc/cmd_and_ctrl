package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Storm Sculptor — Creature — Merfolk Wizard {3}{U}, 3/2 (Jumpstart 2022, #5948):
//
//	"This creature can't be blocked.
//	 When this creature enters, return a creature you control to its owner's
//	 hand."
//
// Evasion is RestrictSelf(CantBeBlocked) (Dimir Infiltrator). The enters
// trigger doesn't target: its controller chooses as it resolves, from the
// creatures they control, the Sculptor itself included (Roaring Primadox's
// shared ReturnOneYouControl). The return is mandatory.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e43eb15e-e910-4165-b9df-25cbe7054e0a",
		Name:         "Storm Sculptor",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBeBlocked)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Storm Sculptor — return a creature you control to its owner's hand", effects.Do(effects.ReturnOneYouControl{
				Match:    effects.MatchCreature,
				Question: "Storm Sculptor — return a creature you control to its owner's hand",
			})),
		},
	})
}
