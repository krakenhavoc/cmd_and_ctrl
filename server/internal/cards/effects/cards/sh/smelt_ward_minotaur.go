package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Smelt-Ward Minotaur — Creature — Minotaur Warrior {2}{R}, 2/3 (Guilds of
// Ravnica, #5525):
//
//	"Whenever you cast an instant or sorcery spell, target creature an
//	 opponent controls can't block this turn."
//
// The trigger fires on the controller's own instant and sorcery spells and
// targets as it goes on the stack (CR 603.3d); on resolution the pick, if still
// legal, can't block until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ee25680a-d01a-4ddb-908b-27a9bea0ceae",
		Name:         "Smelt-Ward Minotaur",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.InstantOrSorceryCastByYou(ev, source, g)
			}, "Smelt-Ward Minotaur — target creature an opponent controls can't block this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							return effects.RestrictUntilEOT{Target: t.ID, Restrictions: game.CantBlock,
								Label: "Smelt-Ward Minotaur — can't block"}.Apply(ctx)
						}
					}
					return nil
				}),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
