package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thantis, the Warweaver — Legendary Creature — Spider {3}{B}{R}{G}, 5/5
// (Commander 2018, #5526):
//
//	"Reach, vigilance
//	 All creatures attack each combat if able.
//	 Whenever a creature attacks you or a planeswalker you control, put a
//	 +1/+1 counter on Thantis."
//
// The attack requirement is Grand Melee's AttacksEachCombatWhere over every
// creature, any controller's, Thantis included (vigilance keeps it untapped
// on defence). The counter is Revenge of Ravens' trigger: one per creature
// declared attacking you or a planeswalker you control (CR 508.3a); an
// attack on a battle you protect does not count (#4864).
//
// No simplification.
func init() {
	attacks := effects.AttacksEachCombatWhere(func(_ *game.Card, _ *game.Game, _ *game.Card) bool { return true })
	attacks.Label = "All creatures attack each combat if able."
	effects.Register(effects.Spec{
		OracleID:        "be543388-1e4f-412f-8f3b-dd826fd8e03d",
		Name:            "Thantis, the Warweaver",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "vigilance"},
		Static:          []game.StaticAbility{attacks},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Kind == game.EventAttack && ev.Actor != uuid.Nil &&
					effects.AttacksYouOrYourPlaneswalker(g, ev.Target, source.Controller)
			}, "Thantis, the Warweaver — put a +1/+1 counter on Thantis",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
