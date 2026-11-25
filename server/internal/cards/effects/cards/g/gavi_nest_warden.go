package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gavi, Nest Warden — {2}{U}{R}{W} Legendary Creature — Human Shaman 2/5
// (Commander 2020):
//
//	"You may pay {0} rather than pay the cycling cost of the first card you
//	 cycle each turn.
//	 Whenever you draw your second card each turn, create a 2/2 red and white
//	 Dinosaur Cat creature token."
//
// The first line is the cycling-cost alternative New Perspectives uses
// (Spec.CyclingCostAlternatives, a twin row of each cycling ability, CR 118.9),
// open only while the controller has cycled no card this turn (their
// EventCycle events). The second is Mindless Conscription's per-turn draw
// tally: the third-card trigger's twin, at two.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "716a3bb8-04bd-41f5-a568-131f175e00a6",
		Name:         "Gavi, Nest Warden",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		CyclingCostAlternatives: []game.EquipCostAlternative{{
			Label: "You may pay {0} rather than pay the cycling cost of the first card you cycle each turn.",
			Cost:  "{0}",
			Condition: func(g *game.Game, controller uuid.UUID) bool {
				for _, ev := range g.EventsThisTurn() {
					if ev.Kind == game.EventCycle && ev.Actor == controller {
						return false
					}
				}
				return true
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return source != nil && ev.Actor == source.Controller && g.TurnTallyFor(source.Controller).CardsDrawn == 2
			}, "Gavi, Nest Warden — create a 2/2 red and white Dinosaur Cat creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 red and white Dinosaur Cat"), N: 1})),
		},
	})
}
