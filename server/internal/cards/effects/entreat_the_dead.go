package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Entreat the Dead — Sorcery {X}{X}{B}{B}{B} (Commander 2018, #5526):
//
//	"Return X target creature cards from your graveyard to the battlefield.
//	 Miracle {X}{B}{B} (You may cast this card for its miracle cost when you
//	 draw it if it's the first card you drew this turn.)"
//
// Return to the Ranks without the mana value cap: exactly X target creature
// cards in your graveyard (CountFromX), chosen as the spell is cast, and
// every one still legal as it resolves returns under your control as one
// entry (CR 608.2b). The miracle cost is Entreat the Angels': it carries its
// own X, which is announced as a hard cast's is and paid once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "2de6c3d9-1759-40a2-99c6-8cbe17b4bcdd",
		Name:             "Entreat the Dead",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Miracle("{X}{B}{B}")},
		Targets: targetsCountedByX(TargetCardInGraveyard("X target creature cards from your graveyard",
			YouOwn(), Creature())),
		OnResolve: cmdr2S33ReturnTargetedCardsTogether,
	})
}
