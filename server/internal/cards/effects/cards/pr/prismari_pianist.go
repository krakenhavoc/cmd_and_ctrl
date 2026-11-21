package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prismari Pianist — Creature — Djinn Bard {1}{R}{R}, 2/1 (Secrets of
// Strixhaven Commander, #4834):
//
//	"Whenever you cast an instant or sorcery spell, create a 1/1 blue and red
//	 Elemental creature token. If that spell's mana value is 5 or greater,
//	 create three of those tokens instead."
//
// A cast trigger (a copy is not cast, CR 707.10, so it does not trigger). The
// spell's mana value is read as the trigger resolves, X as announced
// (CR 202.3b): from the spell if it is still on the stack and from the
// last-known card otherwise, as Deekah, Fractal Theorist does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f58117e4-ba85-41b8-8fd5-4245716a84dc",
		Name:         "Prismari Pianist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.InstantOrSorceryCastByYou(ev, source, g)
			}, "Prismari Pianist — create an Elemental token (three if its mana value is 5 or greater)",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					n := 1
					if effects.SocB4TriggeringSpellManaValue(ctx.Game, ctx.Item) >= 5 {
						n = 3
					}
					return effects.CreateToken{Template: ctx.TextToken(effects.TokenCard("1/1 blue and red Elemental")), N: n}.Apply(ctx)
				}))),
		},
	})
}
