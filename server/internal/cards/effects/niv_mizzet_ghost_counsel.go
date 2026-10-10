package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Niv-Mizzet, Ghost Counsel — Legendary Creature — Spirit Dragon
// {2}{W}{W}{B}{B}, 4/4:
//
//	"Flying
//	 Whenever you gain life, you may pay that much life. If you do,
//	 draw that many cards.
//	 {T}: Each opponent loses 1 life and you gain 1 life."
//
// The "you may pay" is asked as the trigger resolves (MayChoice with a
// LifeCost, CR 608.2d) and re-checked in the Yes branch: a player can
// only pay life they have. The amount is the life gained by the event
// that triggered it, read off the item, so two lifegain events are two
// separate payments. The drain feeds the trigger, as printed.
//
// No simplification.
func init() {
	const label = "Niv-Mizzet, Ghost Counsel — you may pay that much life to draw that many cards"
	Register(Spec{
		OracleID:        "a5350327-e84c-44b6-8359-87882e521048",
		Name:            "Niv-Mizzet, Ghost Counsel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventChangeLife, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Target == source.Controller && ev.Amount > 0
			}, label, func(g *game.Game, item *game.StackItem) error {
				n := 0
				if item.Trigger != nil {
					n = item.Trigger.Event.Amount
				}
				if n <= 0 {
					return nil
				}
				return MayChoice{
					Question: "Niv-Mizzet, Ghost Counsel — pay " + numberWord(n) + " life to draw " + numberWord(n) + " cards?",
					LifeCost: n,
					OnYes: func(ctx *Context) error {
						p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
						if p == nil || p.Eliminated || p.Life < n {
							return nil
						}
						if err := ctx.Game.PayLifeForEffect(ctx.Source(), ctx.Controller(), n); err != nil {
							return err
						}
						return DrawCards{Player: ctx.Controller(), N: n}.Apply(ctx)
					},
				}.Apply(NewContext(g, item))
			}),
		},
		Activated: []ActivatedAbility{{
			Label:  "{T}: Each opponent loses 1 life and you gain 1 life",
			Cost:   TapCost(),
			Effect: drainEachOpponent,
		}},
	})
}
