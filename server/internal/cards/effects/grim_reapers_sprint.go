package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grim Reaper's Sprint — Enchantment — Aura {4}{R} (Fallout):
//
//	"Morbid — This spell costs {3} less to cast if a creature died this
//	 turn.
//	 Enchant creature
//	 When this Aura enters, untap each creature you control. If it's your
//	 main phase, there is an additional combat phase after this phase.
//	 Enchanted creature gets +2/+2 and has haste."
//
// The discount is Purple Worm's self cost modifier over the turn's tally of
// creatures that died; the pump and haste are Goblin War Paint's statics.
// The enters trigger untaps every creature I control, then adds a combat
// phase after the current phase (CR 500.8, Combat Celebrant's extra combat,
// no main phase with it) only when the phase now is a main phase of my own
// turn: cast or put onto the battlefield in an opponent's main phase, or in
// a step that is no main phase, it adds nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "99a6d7fe-40a1-4f2f-a3f6-6777884d5d1a",
		Name:         "Grim Reaper's Sprint",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, ExtraCombat: 1},
		SelfCostModifiers: []game.CostModifier{
			CostsLess(3, "Morbid — This spell costs {3} less to cast if a creature died this turn.", ACreatureDiedThisTurnSpell()),
		},
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("haste"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Grim Reaper's Sprint — untap each creature you control; if it's your main phase, there is an additional combat phase after this phase",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (UntapAllCreaturesYouControl{}).Apply(ctx); err != nil {
						return err
					}
					if !IsYourTurn(g, item.Controller) || (g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain) {
						return nil
					}
					return ExtraCombatAfterThisPhase().Apply(ctx)
				}),
		},
	})
}
