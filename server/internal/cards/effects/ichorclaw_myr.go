package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ichorclaw Myr — Artifact Creature — Phyrexian Myr {2}, 1/1:
//
//	"Infect
//	 Whenever this creature becomes blocked, it gets +2/+2 until end of turn."
//
// Infect is the engine's damage tail (ADR 0056, #748). The trigger is the
// block declaration's EventBecomesBlocked, emitted once per blocked attacker
// however many creatures block it (CR 509.3c); the pump is applied to this
// permanent only while it is still the one that was blocked.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8462a2fa-944b-4fff-a161-10345690b56f",
		Name:            "Ichorclaw Myr",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"infect"},
		Triggered: []game.TriggeredAbility{
			On(game.EventBecomesBlocked, Self, "Ichorclaw Myr — gets +2/+2 until end of turn",
				Do(Instruction(func(ctx *Context) error {
					if !sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
						return nil
					}
					return BoostUntilEOT{
						Target: ctx.Item.SourceCardID, Power: 2, Toughness: 2,
						Label: "Ichorclaw Myr — +2/+2",
					}.Apply(ctx)
				}))),
		},
	})
}
