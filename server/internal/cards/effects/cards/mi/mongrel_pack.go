package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mongrel Pack — {3}{G} Creature — Dog, 4/1 (Tempest, S82 OLD batch 769, #5628):
//
//	"When this creature dies during combat, create four 1/1 green Dog creature
//	 tokens."
//
// Cathedral Membrane's "dies during combat": the death is read as the Pack
// leaves the battlefield, from the combat step the game is in (the intervening
// condition, CR 603.4), so it dies in combat damage or to a spell cast during
// combat, and one that dies in a main phase makes nothing. The tokens are the
// controller's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e1208d39-86e7-4eae-8dd6-b5bfb29def84",
		Name:         "Mongrel Pack",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
				return effects.ThisDied(ev, source, ch, g) && effects.InCombatStep(g.Turn.Step)
			}, "Mongrel Pack — create four 1/1 green Dog creature tokens",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Dog"), N: 4})),
		},
	})
}
