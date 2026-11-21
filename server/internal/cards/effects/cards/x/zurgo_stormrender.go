package cardsx

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zurgo Stormrender — Legendary Creature — Orc Warrior {R}{W}{B}, 3/3
// (Tarkir: Dragonstorm Commander):
//
//	"Mobilize 1 (Whenever this creature attacks, create a tapped and
//	 attacking 1/1 red Warrior creature token. Sacrifice it at the beginning
//	 of the next end step.)
//	 Whenever a creature token you control leaves the battlefield, draw a card
//	 if it was attacking. Otherwise, each opponent loses 1 life."
//
// Mobilize is the shared keyword trigger. The second ability is Nadier's
// Nightblade's token exit (EventLTB, controller read as it left, CR 603.10a)
// narrowed to creature tokens, one trigger per token even when Zurgo leaves
// with them (the rulings). "Was attacking" is the leaving permanent's combat
// state carried on the event (Event.AttackingTarget, taken as it left the
// battlefield), so an attacking Warrior sacrificed at end of turn, which is
// no longer attacking after combat, drains and a Warrior that dies in combat
// draws.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0f93d88c-9d2e-416d-a10b-99483360b1fb",
		Name:         "Zurgo Stormrender",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Mobilize("Zurgo Stormrender", 1),
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && effects.IsToken(c) && effects.LeftAsType(ev, c, "Creature") && effects.LeftUnderControlOf(ev, c) == source.Controller
			}, "Zurgo Stormrender — draw a card if it was attacking, otherwise each opponent loses 1 life",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if item.Trigger.Event.AttackingTarget != uuid.Nil {
						return effects.DrawCards{N: 1}.Apply(ctx)
					}
					return effects.EachOpponentLosesLife{N: 1}.Apply(ctx)
				}),
		},
	})
}
