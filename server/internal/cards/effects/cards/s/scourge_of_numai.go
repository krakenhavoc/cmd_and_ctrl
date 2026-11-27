package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scourge of Numai — Creature — Demon Spirit {3}{B}, 4/4 (Betrayers of
// Kamigawa, S82 OLD, #5602):
//
//	"At the beginning of your upkeep, you lose 2 life if you don't control
//	 an Ogre."
//
// An intervening "if" (CR 603.4): checked as the upkeep begins, so with an
// Ogre out nothing goes on the stack, and again as the trigger resolves,
// so an Ogre that arrives in response saves the life. The Scourge is not
// an Ogre. The life loss is not damage (CR 119.3).
//
// No simplification.
func init() {
	// noOgre reads "an Ogre" through the Scourge's own text (CR 612.2): as
	// the trigger fires the source is the permanent; as it resolves, the
	// card if it is still anywhere, and as printed if it has gone.
	noOgre := func(g *game.Game, you uuid.UUID, source *game.Card) bool {
		word := "Ogre"
		if source != nil {
			word = source.TextSubtype("Ogre")
		}
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == you && c.HasSubtypeWord(word) {
				return false
			}
		}
		return true
	}
	upkeep := effects.AtYourUpkeep("Scourge of Numai — you lose 2 life", effects.Do(effects.YouLoseLife(2)))
	when := upkeep.AppliesTo
	upkeep.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return source != nil && noOgre(g, source.Controller, source) && when(ev, source, lki, g)
	}
	upkeep = effects.InterveningIf(effects.CondBoard{Name: "you don't control an Ogre", Test: func(ctx *effects.Context) bool {
		src, _ := ctx.Game.LookupCardForEffect(ctx.Item.SourceCardID)
		if src.InstanceID == uuid.Nil {
			return noOgre(ctx.Game, ctx.Controller(), nil)
		}
		return noOgre(ctx.Game, ctx.Controller(), &src)
	}}, upkeep)
	effects.Register(effects.Spec{
		OracleID:     "540d1878-7afe-48a5-8b40-3dd1b06a024a",
		Name:         "Scourge of Numai",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{upkeep},
	})
}
