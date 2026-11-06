package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Appa, Aang's Companion — Legendary Creature — Bison Ally {3}{W}, 2/4
// (Avatar: The Last Airbender Eternal, #5075):
//
//	"Flying
//	 Whenever Appa attacks, another target attacking creature without
//	 flying gains flying until end of turn."
//
// The target is chosen as the trigger goes on the stack and re-checked
// as it resolves (CR 608.2b): a creature that gained flying, stopped
// attacking or left in response is no longer legal and gets nothing.
//
// No simplification.
func init() {
	const label = "Appa, Aang's Companion — another target attacking creature without flying gains flying until end of turn"
	Register(Spec{
		OracleID:        "215215b8-6ddc-4e94-8640-6d324ff65f9c",
		Name:            "Appa, Aang's Companion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks(label, firstTargetGainsUntilEndOfTurn(0, label, "flying")),
				Another(TargetCreature("another target attacking creature without flying",
					AttackingCreature(), WithoutKeyword("flying")))),
		},
	})
}
