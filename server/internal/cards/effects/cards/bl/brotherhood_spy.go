package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brotherhood Spy — Creature — Human Assassin {1}{U}, 1/3 (Assassin's Creed,
// #5948):
//
//	"At the beginning of combat on your turn, if you control a legendary
//	 Assassin, this creature gets +1/+0 until end of turn and can't be blocked
//	 this turn."
//
// An intervening "if" (CR 603.4) like Boundary Lands Ranger's: checked as combat
// begins and again as the ability resolves, so a trigger whose legendary
// Assassin has gone does nothing. The Spy is itself an Assassin but not
// legendary, so it never satisfies its own condition. The pump and the
// restriction are both applied as it resolves; a Spy that has left gets neither.
//
// No simplification.
func init() {
	const label = "Brotherhood Spy — this creature gets +1/+0 until end of turn and can't be blocked this turn"
	effects.Register(effects.Spec{
		OracleID:     "26d800b2-b312-4400-8357-615301971086",
		Name:         "Brotherhood Spy",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{Name: "you control a legendary Assassin", Test: func(ctx *effects.Context) bool {
				legend, assassin := effects.Legendary(), effects.HasSubtype("Assassin")
				for _, c := range ctx.Game.Battlefield.Cards {
					if c.Controller == ctx.Controller() && legend(ctx.Game, ctx.Controller(), c) && assassin(ctx.Game, ctx.Controller(), c) {
						return true
					}
				}
				return false
			}}, effects.AtBeginningOfYourCombat(label, func(g *game.Game, item *game.StackItem) error {
				if !effects.OnBattlefield(g, item.SourceCardID) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				if err := (effects.BoostUntilEOT{Target: item.SourceCardID, Power: 1, Label: label}).Apply(ctx); err != nil {
					return err
				}
				return effects.RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked, Label: label}.Apply(ctx)
			})),
		},
	})
}
