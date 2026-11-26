package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Alhammarret, High Arbiter — Legendary Creature — Sphinx {5}{U}{U}, 5/5
// (Magic Origins, S82 older sets):
//
//	"Flying
//	 As Alhammarret enters, each opponent reveals their hand. You choose the
//	 name of a nonland card revealed this way.
//	 Your opponents can't cast spells with the chosen name (as long as this
//	 creature is on the battlefield)."
//
// Anointed Peacekeeper's as-enters choice with a revealed list instead of free
// text: every opponent's hand is revealed to the table, and the controller picks
// among the distinct nonland names revealed (the name is stored on the
// permanent, CR 614.12). The ban is a cast restriction of the permanent, so it
// binds the controller's opponents only, matches any face of the spell (CR
// 201.2b) and lifts when Alhammarret leaves. With no nonland card revealed no
// name is chosen and nothing is barred.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "c2531360-61f8-4edc-886b-408811a0639d",
		Name:             "Alhammarret, High Arbiter",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AsEnters:         effects.OldB322AlhammarretAsEnters,
		CastRestrictions: []game.CastRestriction{effects.OldB322OpponentsCantCastTheChosenName()},
	})
}
