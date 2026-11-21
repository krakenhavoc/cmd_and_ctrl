package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Chief Warg — Legendary Creature — Wolf {2}{B}{G}, 3/3 (The Hobbit):
//
//	"Menace
//	 Ferocious — Whenever you attack while you control a creature with
//	 power 4 or greater, you draw a card and lose 1 life."
//
// "Whenever you attack" is once per declaration however many creatures
// attack (OncePerBatch, the Sparring Regimen shape). The power check is
// a condition on the event, read as attackers are declared.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5ebe8de1-aa3d-410d-b43d-1685259c7a97",
		Name:            "The Chief Warg",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller) && effects.B24ControlsCreatureWithPowerAtLeast(g, source.Controller, 4)
			}, "The Chief Warg — ferocious: you draw a card and lose 1 life", effects.B36DrawAndLoseOne)),
		},
	})
}
