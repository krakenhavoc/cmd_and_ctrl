package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Drana's Silencer — Creature — Vampire Rogue {5}{B}, 3/2 (Zendikar Rising,
// #5508):
//
//	"When this creature enters, target creature an opponent controls gets
//	 -X/-X until end of turn, where X is the number of creatures in your
//	 party. (Your party consists of up to one each of Cleric, Rogue,
//	 Warrior, and Wizard.)"
//
// The target is chosen as the trigger goes on the stack and re-checked as it
// resolves (CR 608.2b). X is the party (CR 700.8) as it resolves, the
// Silencer itself being the Rogue if nothing else is.
//
// No simplification.
func init() {
	const label = "Drana's Silencer — target creature an opponent controls gets -X/-X until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "8e993c7a-606d-47f0-bdea-6ad197a59426",
		Name:         "Drana's Silencer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters(label,
					effects.Do(effects.Instruction(func(ctx *effects.Context) error {
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						x := effects.FinB9PartySize(ctx.Game, ctx.Controller())
						return effects.BoostUntilEOT{Target: id, Power: -x, Toughness: -x, Label: label}.Apply(ctx)
					}))),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
