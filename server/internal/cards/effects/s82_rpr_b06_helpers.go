package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_rpr_b06_helpers.go — shared helpers for the S82 reprint pass,
// batch 6 (#5908).

// rprB06HasteAndEndStepTail is "It gains haste. <Sacrifice|Exile> it at the
// beginning of the next end step." for the token copies a creation made,
// as the body of a keyed tail: the haste is an indefinite grant on each
// token and the removal is one delayed trigger naming them (CR 603.7).
func RprB06HasteAndEndStepTail(ctx *Context, o game.TailOutcome, label string, body game.BodyRef) error {
	if len(o.IDs) == 0 {
		return nil
	}
	for _, id := range o.IDs {
		if err := (ScopedEffectFor{Target: id, Mods: []game.Mod{game.AddKeywordsMod("haste")},
			Duration: game.IndefiniteDuration(), Label: label + " — the token gains haste"}).Apply(ctx); err != nil {
			return err
		}
	}
	return ScheduleDelayedTrigger{
		At:         game.StepEnd,
		Controller: ctx.Item.Controller,
		Label:      label + " — remove the token",
		Cards:      o.IDs,
		Body:       body,
	}.Apply(ctx)
}
