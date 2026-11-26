package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Faith's Shield — Instant {W}:
//
//	"Target permanent you control gains protection from the color of your
//	 choice until end of turn.
//	 Fateful hour — If you have 5 or less life, instead you and each
//	 permanent you control gain protection from the color of your choice
//	 until end of turn."
//
// The fateful hour half is the whole effect "instead" (the target is
// unused, and the spell is still cast with one): one colour is chosen as it
// resolves and given to you and to every permanent you control.
func init() {
	protection := effects.ProtectionFromAChosenColorForTheTarget("Faith's Shield")
	effects.Register(effects.Spec{
		OracleID:     "dd3cc9e5-b99a-429c-baeb-663a7af70081",
		Name:         "Faith's Shield",
		Purpose:      game.Purpose{Answers: game.AnswerProtect, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "protection"))},
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target permanent you control", effects.YouControl()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if effects.OldB441FatefulHour(ctx.Game, ctx.Controller()) {
				effects.ChooseColorTail(game.ColorForProtection, ctx.Game, item.Controller, ctx.Source(),
					"Faith's Shield — choose a color", ctx.ItemTail(effects.OldB441FaithsShieldFatefulHour))
				return nil
			}
			return protection(item, ctx)
		},
	})
}
