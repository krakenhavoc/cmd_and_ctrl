package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Valko Indorian — Legendary Creature — Human Wizard {3}{B}, 2/4 (Mystery
// Booster Commander, #5948):
//
//	"Rulebreaker — A deck with this commander can have Phyrexian cards of any
//	 color identity and any basic land cards.
//	 Phyrexian creatures you control have menace and lifelink."
//
// The deck clause is a deck-construction exception read off the commander's
// own text by internal/deck (rulebreaker.go); in the game it is nothing.
// The lord half is two layer-6 grants over Phyrexian creatures you control
// (Valko himself is a Human Wizard, so he gets neither).
//
// No simplification.
func init() {
	phyrexians := TribeFilter{Tribes: []string{"Phyrexian"}, YoursOnly: true}
	Register(Spec{
		OracleID:     "8023d419-8fd2-4b92-b6ea-f25fd67e6410",
		Name:         "Valko Indorian",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalKeywordGrant(phyrexians, "menace"),
			TribalKeywordGrant(phyrexians, "lifelink"),
		},
	})
}
