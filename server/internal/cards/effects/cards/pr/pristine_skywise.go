package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pristine Skywise — Creature — Dragon {4}{W}{U}, 6/4 (Commander 2019, #5517):
//
//	"Flying
//	 Whenever you cast a noncreature spell, untap this creature. It gains
//	 protection from the color of your choice until end of turn."
//
// One trigger on any noncreature spell its controller casts (it resolves
// above the spell). It untaps the Skywise, then asks for the colour; both
// halves do nothing once the Skywise has left the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "de80c742-4da0-4d67-9711-2b4e6452e2d0",
		Name:            "Pristine Skywise",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.B10NoncreatureSpellCastByYou, "Pristine Skywise — untap and gain protection",
				effects.Do(effects.Instruction(oldB153SkywiseUntapAndProtect))),
		},
	})
}

// oldB153SkywiseUntapAndProtect is the trigger's body.
func oldB153SkywiseUntapAndProtect(ctx *effects.Context) error {
	if !effects.OnBattlefield(ctx.Game, ctx.Source()) {
		return nil
	}
	if err := (effects.UntapTarget{Target: ctx.Source()}).Apply(ctx); err != nil {
		return err
	}
	effects.ChooseProtectionColorFor(ctx, ctx.Controller(), ctx.Source(), "Pristine Skywise", "Pristine Skywise — choose a color")
	return nil
}
