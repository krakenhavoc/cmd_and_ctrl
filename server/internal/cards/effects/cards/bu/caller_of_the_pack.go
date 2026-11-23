package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Caller of the Pack — Creature — Beast {5}{G}{G}, 8/6 (Commander 2015,
// #5546):
//
//	"Trample
//	 Myriad (Whenever this creature attacks, for each opponent other than
//	 defending player, you may create a token copy that's tapped and
//	 attacking that player or a planeswalker they control. Exile the tokens
//	 at end of combat.)"
//
// Myriad is the shared trigger (myriad.go, CR 702.116); trample is printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3b6cbd2a-02ba-4afb-a081-0f25d830a8ee",
		Name:            "Caller of the Pack",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered:       []game.TriggeredAbility{effects.GrantedMyriad("Caller of the Pack")},
	})
}
