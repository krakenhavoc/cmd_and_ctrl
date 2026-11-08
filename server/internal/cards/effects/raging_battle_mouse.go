package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Raging Battle Mouse — Creature — Mouse {1}{R}, 2/1 (Wilds of Eldraine,
// #4471):
//
//	"The second spell you cast each turn costs {1} less to cast.
//	 Celebration — At the beginning of combat on your turn, if two or more
//	 nonland permanents entered the battlefield under your control this
//	 turn, target creature you control gets +1/+1 until end of turn."
//
// The discount is Uthros Psionicist's (secondSpellYouCastThisTurn) at
// {1}: it applies to your second spell of the turn only, and reduces
// generic mana. Celebration is Belligerent of the Ball's intervening if
// (celebration.go, CR 603.4), checked as combat begins and again as the
// ability resolves. The target is any creature you control, the Mouse
// included; with none, the trigger is removed (CR 603.3d).
//
// No simplification.
func init() {
	const label = "Raging Battle Mouse — target creature you control gets +1/+1 until end of turn"
	Register(Spec{
		OracleID:     "9ffe4d90-b46c-4606-89cb-1b724104f6ea",
		Name:         "Raging Battle Mouse",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "The second spell you cast each turn costs {1} less to cast.",
				YourSpell(), secondSpellYouCastThisTurn()),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(AtBeginningOfYourCombatIfCelebration(label, pumpAndGrantFirstTarget(1, 1, label)),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}
