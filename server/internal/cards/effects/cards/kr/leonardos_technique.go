package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leonardo's Technique — Sorcery {3}{W} (TMT, #4457):
//
//	"Sneak {1}{W} (You may cast this spell for {1}{W} if you also return
//	 an unblocked attacker you control to hand during the declare
//	 blockers step.)
//	 Return one or two target creature cards each with mana value 3 or
//	 less from your graveyard to the battlefield."
//
// Sneak is Sneak(cost) (CR 702.190). Infernal Rebirth's one-or-two
// targets with a mana value ceiling, put onto the battlefield together;
// a pick that left the graveyard in response is skipped and the other
// still returns (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "e98c2af9-1393-487d-b8e9-f49a67bb7296",
		Name:             "Leonardo's Technique",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Sneak("{1}{W}")},
		Targets: effects.TargetCardInGraveyard("one or two target creature cards each with mana value 3 or less from your graveyard",
			effects.YouOwn(), effects.Creature(), effects.ManaValueLE(3)).WithCount(1, 2),
		OnResolve: effects.ResolveSteps(effects.Instruction(effects.ReturnLegalGraveyardTargetsToBattlefield)),
	})
}
