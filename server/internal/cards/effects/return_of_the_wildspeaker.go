package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Return of the Wildspeaker — Instant {4}{G} (EDHREC rank 178):
//
//	"Choose one —
//	 • Draw cards equal to the greatest power among non-Human
//	   creatures you control.
//	 • Non-Human creatures you control get +3/+3 until end of turn."
//
// Green's instant-speed choice between a refill and an alpha strike.
// The roadmap filed it under "until end of turn" as a blocked card;
// #314 shipped the turn-scoped statics and it is now two primitives
// on two modes:
//
//   - The draw reads CurrentPower — effective power plus counters —
//     so a creature pumped earlier in the turn, or carrying +1/+1
//     counters, draws the bigger number, as printed.
//   - The pump is BoostUntilEOT with a Match predicate, snapshotted
//     at resolution (CR 611.2c): a creature that enters after the
//     spell resolves is not pumped, and a creature that becomes
//     Human afterwards keeps the bonus.
//
// "Non-Human" reads the post-layer subtype list, so a type-changing
// effect composes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "2b76f9e9-cd28-4eaf-8674-215c34263f96",
		Name:     "Return of the Wildspeaker",
		Modes: ChooseOne(
			Mode("Draw cards equal to the greatest power among non-Human creatures you control."),
			Mode("Non-Human creatures you control get +3/+3 until end of turn."),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if ctx.HasMode(0) {
				best := 0
				for _, c := range ctx.Game.BattlefieldCardsForEffect() {
					if !nonHumanCreatureYouControl(ctx.Game, ctx.Controller(), c) {
						continue
					}
					if p := c.CurrentPower(); p > best {
						best = p
					}
				}
				if err := (DrawCards{Player: ctx.Controller(), N: best}).Apply(ctx); err != nil {
					return err
				}
			}
			if ctx.HasMode(1) {
				return BoostUntilEOT{
					Match:     nonHumanCreatureYouControl,
					Power:     3,
					Toughness: 3,
					Label:     "Return of the Wildspeaker — non-Human creatures you control get +3/+3",
				}.Apply(ctx)
			}
			return nil
		},
	})
}

// nonHumanCreatureYouControl is the clause both modes share.
func nonHumanCreatureYouControl(_ *game.Game, caster uuid.UUID, c game.Card) bool {
	return c.IsCreature() && c.Controller == caster && !hasSubtype(c, "Human")
}
