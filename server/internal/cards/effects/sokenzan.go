package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sokenzan — Plane — Kamigawa (Commander Masters, #4870):
//
//	"All creatures get +1/+1 and have haste.
//	 Whenever chaos ensues, untap all creatures that attacked this turn.
//	 If it's a main phase, there is an additional combat phase after this
//	 phase, followed by an additional main phase."
//
// A plane (ADR 0216 §6). The statics are a layer-7c +1/+1 and a layer-6
// haste over every creature, whoever controls it. The chaos ability (CR
// 311.7) untaps every creature that attacked this turn (any player's), then,
// only when it resolves during a main phase, adds a combat phase and a main
// phase after that main phase (CR 500.8); resolving in any other phase it
// adds nothing, per the Relentless Assault ruling the primitive follows.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c09d304a-87d0-4a2a-bcb2-e81341b2dbf2",
		Name:         "Sokenzan",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			b16Anthem(allCreatures, 1, 1),
			b16GrantKeywords(allCreatures, "haste"),
		},
		Triggered: []game.TriggeredAbility{
			WheneverChaosEnsues("Sokenzan — untap all creatures that attacked this turn; if it's a main phase, an additional combat phase and main phase follow",
				Do(UntapCreaturesThatAttackedThisTurn{}, ExtraCombatAndMainAfterThisMain())),
		},
	})
}
