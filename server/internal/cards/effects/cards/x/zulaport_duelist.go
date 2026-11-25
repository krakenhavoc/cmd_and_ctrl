package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zulaport Duelist — Creature — Human Rogue {U}, 1/1 (Zendikar Rising,
// #5508):
//
//	"Flash
//	 When this creature enters, up to one target creature gets -2/-0 until
//	 end of turn. Its controller mills two cards."
//
// With no target chosen nothing happens, and nobody mills: "its controller"
// is the target's. The controller is read as the ability resolves, so a
// target that changed hands is milled by its new controller; a target that
// left the battlefield makes the ability fizzle (CR 608.2b).
//
// No simplification.
func init() {
	const label = "Zulaport Duelist — up to one target creature gets -2/-0 until end of turn; its controller mills two cards"
	effects.Register(effects.Spec{
		OracleID:        "e8237d12-2860-4cc8-85a4-458e4c68e69a",
		Name:            "Zulaport Duelist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters(label,
					effects.Do(effects.Instruction(func(ctx *effects.Context) error {
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						if err := (effects.BoostUntilEOT{Target: id, Power: -2, Label: label}).Apply(ctx); err != nil {
							return err
						}
						c, ok := ctx.Game.LookupCardForEffect(id)
						if !ok {
							return nil
						}
						return effects.MillCards{Player: c.Controller, N: 2}.Apply(ctx)
					}))),
				effects.TargetCreature("up to one target creature").WithCount(0, 1)),
		},
	})
}
