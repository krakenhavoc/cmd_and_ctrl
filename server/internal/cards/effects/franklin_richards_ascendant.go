package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Franklin Richards, Ascendant — Legendary Creature — Mutant Hero
// {5}{R}, 6/6:
//
//	"At the beginning of combat on your turn, if you've cast a
//	 noncreature spell this turn, discover 6."
//
// An intervening if (CR 603.4): checked as combat begins, so the
// ability does not trigger at all without a noncreature spell, and
// again as it resolves. Discover is ADR 0099's (game/discover.go).
func init() {
	Register(Spec{
		OracleID:     "af20f452-5cb9-47cc-9a81-71f23204ddd0",
		Name:         "Franklin Richards, Ascendant",
		Completeness: CompletenessFull,
		Discovers:    true,
		Triggered: []game.TriggeredAbility{
			On(game.EventStepBegan, AllOf(StepBegan(game.StepBeginCombat, true), franklinCastNoncreature),
				"Franklin Richards, Ascendant — discover 6",
				func(g *game.Game, item *game.StackItem) error {
					if g.CastTallyFor(item.Controller).Noncreature == 0 {
						return nil
					}
					return Discover{N: 6}.Apply(NewContext(g, item))
				}),
		},
	})
}

// franklinCastNoncreature is "if you've cast a noncreature spell this
// turn".
func franklinCastNoncreature(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return source != nil && g.CastTallyFor(source.Controller).Noncreature > 0
}
