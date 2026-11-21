package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Permission Denied — Instant {W}{U} (Jurassic World Collection, #5948):
//
//	"Counter target noncreature spell. Your opponents can't cast
//	 noncreature spells this turn."
//
// Two instructions in printed order. The ban is a cast ban on each opponent
// for the rest of the turn with a noncreature filter (Calamity's Wake's
// shape). A spell whose only target is gone does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "95aae5ba-7634-4cad-a0be-496c3398fa3c",
		Name:         "Permission Denied",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetSpell("target noncreature spell", Noncreature()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) > 0 && item.Targets[0].Kind == game.TargetCard {
				if err := (CounterTarget{StackID: item.Targets[0].ID}).Apply(ctx); err != nil {
					return err
				}
			}
			for _, p := range ctx.Game.Seats {
				if p == nil || p.ID == ctx.Controller() {
					continue
				}
				if err := (RestrictCasting{
					Player:   p.ID,
					Rule:     game.CastBanRule{Kind: game.CastBanOutright, Filter: game.PermissionFilter{NoncreatureOnly: true}},
					Label:    "Permission Denied — can't cast noncreature spells this turn",
					Duration: DurationUntilEndOfTurn(ctx),
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
