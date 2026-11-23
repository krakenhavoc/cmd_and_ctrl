package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tomb of Horrors Adventurer — Creature — Elf Monk {5}{U}, 4/4 (Commander
// Legends: Battle for Baldur's Gate, #5948):
//
//	"When this creature enters, you take the initiative.
//	 Whenever you cast your second spell each turn, copy it. If you've completed
//	 a dungeon, copy that spell twice instead. You may choose new targets for
//	 the copies. (A copy of a permanent spell becomes a token.)"
//
// The initiative is the engine's (CR 725). The second spell is Monk of the
// Open Hand's clock (a cast tally of exactly two), copied the way Double
// Vision copies (CopySpell from last-known information, the new-targets choice
// owed, CR 707.10c). "If you've completed a dungeon" is part of the effect, so
// it is read as the trigger resolves (CR 608.2): one copy, or two.
//
// No simplification.
func init() {
	const label = "Tomb of Horrors Adventurer — copy that spell; if you've completed a dungeon, copy it twice instead"
	effects.Register(effects.Spec{
		OracleID:     "4f429b99-6aca-409f-9d2e-5743d81f1f22",
		Name:         "Tomb of Horrors Adventurer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouTakeTheInitiative("Tomb of Horrors Adventurer"),
			effects.On(game.EventCast, effects.YouCastYourSecondSpellEachTurn, label,
				func(g *game.Game, item *game.StackItem) error {
					n := 1
					if effects.YouCompletedADungeon(g, item.Controller) {
						n = 2
					}
					return effects.CopySpell{
						StackID:          item.Trigger.Event.CardID,
						Controller:       item.Controller,
						Count:            n,
						ChooseNewTargets: true,
						FromLastKnown:    true,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
