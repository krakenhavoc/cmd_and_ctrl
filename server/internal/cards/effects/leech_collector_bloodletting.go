package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leech Collector // Bloodletting — a preparation card (CR 722, ADR 0090),
// Secrets of Strixhaven (#4456):
//
//	Leech Collector {1}{B}, Creature — Human Warlock 2/2:
//	"Whenever you gain life for the first time each turn, this creature
//	 becomes prepared. (While it's prepared, you may cast a copy of its
//	 spell. Doing so unprepares it.)"
//	Bloodletting {B}, Sorcery:
//	"Each opponent loses 2 life."
//
// "For the first time each turn" is read off the turn's log
// (socB28FirstLifeGainThisTurn): the first positive life change of the
// controller's this turn, whether or not the Collector was out for any
// earlier one. It triggers once however much life was gained, and a gain on
// an opponent's turn is that turn's first. A creature that is already
// prepared gains nothing (CR 722.3a). Bloodletting is a plain drain.
//
// No simplification.
func init() {
	const id = "84bb960c-c34e-4979-8c0f-893c85cee27d"
	Register(Spec{
		OracleID:     id,
		Name:         "Leech Collector",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventChangeLife, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return socB28FirstLifeGainThisTurn(ev, source, g)
			}, "Leech Collector — becomes prepared",
				func(g *game.Game, item *game.StackItem) error {
					return BecomePrepared{Target: item.SourceCardID}.Apply(NewContext(g, item))
				}),
		},
	})
	Register(Spec{
		OracleID:     id + "#1",
		Name:         "Bloodletting",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return EachOpponentLosesLife{N: 2}.Apply(ctx)
		},
	})
}
