package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tragedy Feaster — Creature — Demon {2}{B}{B}, 7/6 (Secrets of
// Strixhaven Commander, #4834):
//
//	"Trample
//	 Ward—Discard a card.
//	 Infusion — At the beginning of your end step, sacrifice a permanent
//	 unless you gained life this turn."
//
// Ward—Discard a card is Ward(WardDiscard(1)) (ward.go). The infusion
// trigger fires at every one of its controller's end steps whether or not
// they gained life; the ruling is that the check happens as it resolves,
// so it reads the life gained this turn then (the turn tally every gain
// path bumps, lifelink and a drain's gain half included) and only asks
// for a sacrifice when there was none. The controller chooses the
// permanent, any of theirs, the Feaster itself included; with the choice
// open the table waits on it (a mandatory sacrifice of a permanent
// that does not exist does nothing, CR 701.21a).
//
// No simplification.
func init() {
	const label = "Tragedy Feaster — sacrifice a permanent unless you gained life this turn"
	effects.Register(effects.Spec{
		OracleID:        "0c4277ec-aa12-41f5-8896-c83cfc754d6e",
		Name:            "Tragedy Feaster",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardDiscard(1), "Tragedy Feaster — ward, discard a card"),
			effects.AtYourEndStep(label, effects.Do(effects.Instruction(func(ctx *effects.Context) error {
				if effects.B15LifeGainedThisTurn(ctx.Game, ctx.Controller()) > 0 {
					return nil
				}
				return effects.SacrificeChoice{
					Player:     ctx.Controller(),
					Candidates: effects.PermanentsControlledByPlayer(ctx.Game, ctx.Controller()),
					Question:   "Tragedy Feaster — sacrifice a permanent",
				}.Apply(ctx)
			}))),
		},
	})
}
