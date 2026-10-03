package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Desperate Gambit — Instant {R}:
//
//	"Choose a source you control and flip a coin. If you win the flip,
//	 the next time that source would deal damage this turn, it deals
//	 double that damage instead. If you lose the flip, the next time it
//	 would deal damage this turn, prevent that damage."
//
// The source is chosen as the spell resolves (CR 609.7a), among the
// sources you control only (ADR 0108 §3's controller filter on
// choose_source): your permanents and spells, and the cards you own in
// the command zone (CR 108.4a). Then you flip.
//
//   - Won: a "next time" multiplier on that source (ADR 0108 §3): its
//     next instance of damage this turn is doubled, however many things
//     it hits at once (CR 615.8's instance).
//   - Lost: ADR 0107 §6's next-damage shield on the same source, to
//     anything.
//
// With no source you control, the coin is still flipped and nothing
// else happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d8328d27-e27f-4c84-8ab5-cabb80a34f54",
		Name:         "Desperate Gambit",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			controller, source, self := ctx.Controller(), ctx.Source(), item.ID
			_, err := ctx.Game.ChooseDamageSourceThenForEffect(game.ChooseSourcePrompt{
				Chooser:    controller,
				Source:     source,
				Controller: controller,
				Question:   "Desperate Gambit — choose a source you control",
				Then: func(g *game.Game, ref game.ObjectRef, zone game.ZoneKind) error {
					NextTimeFlip{Source: ref, SourceZone: zone}.queue(g, nextTimeFlipBy{
						controller: controller, source: source, self: self,
						kind: game.StackItemSpell, name: "Desperate Gambit",
					})
					return nil
				},
			})
			return err
		},
	})
}
