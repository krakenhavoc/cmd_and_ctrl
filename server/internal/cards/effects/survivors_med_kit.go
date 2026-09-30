package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// survivorsMedKitLabel is the activated ability's label, and with it
// the key its "hasn't been chosen" memory is kept under.
const survivorsMedKitLabel = "{1}, {T}: Choose one that hasn't been chosen"

// Survivor's Med Kit — Artifact {1}:
//
//	"{1}, {T}: Choose one that hasn't been chosen —
//	 • Stimpak — Draw a card.
//	 • Fancy Lads Snack Cakes — Create a Food token.
//	 • RadAway — Target player loses all rad counters. Sacrifice this
//	   artifact."
//
// An activated ability with ChooseOneNotChosen (ADR 0097): each bullet
// once for this object, announced with the activation (CR 602.2b) and
// recorded once it succeeds, so a used bullet is refused with nothing
// paid and the bot is never offered one. The bullet names are flavour
// words.
//
// "Loses all rad counters" removes every rad counter from that player;
// "sacrifice this artifact" then sacrifices the Med Kit, which is
// usually the last bullet anyone takes. A Med Kit that left and came
// back is a new object and is not sacrificed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fa08400f-b7ca-4c5c-b7ca-a6583b783878",
		Name:         "Survivor's Med Kit",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: survivorsMedKitLabel,
			Cost:  Plus(ManaCost("{1}"), TapCost()),
			Modes: ChooseOneNotChosen(
				ModeDoing("Stimpak — Draw a card.", nil,
					func(item *game.StackItem, ctx *Context, _ int) error {
						return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
					}),
				ModeDoing("Fancy Lads Snack Cakes — Create a Food token.", nil,
					func(item *game.StackItem, ctx *Context, _ int) error {
						return CreateToken{Controller: item.Controller, Template: FoodToken(), N: 1}.Apply(ctx)
					}),
				ModeDoing("RadAway — Target player loses all rad counters. Sacrifice this artifact.",
					TargetPlayer("target player"),
					func(item *game.StackItem, ctx *Context, occ int) error {
						if t, ok := ModeTarget(ctx, occ); ok && t.Kind == game.TargetPlayer {
							if p := ctx.PlayerByID(t.ID); p != nil {
								if n := p.Counters[game.CounterRad]; n > 0 {
									if err := ctx.Game.AddPlayerCounterForEffect(t.ID, game.CounterRad, -n); err != nil {
										return err
									}
								}
							}
						}
						if !onBattlefield(ctx.Game, item.SourceCardID) {
							return nil
						}
						return SacrificePermanent{Target: item.SourceCardID}.Apply(ctx)
					}),
			),
		}},
	})
}
