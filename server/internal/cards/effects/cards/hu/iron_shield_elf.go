package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Iron-Shield Elf — Creature — Elf Warrior {1}{B}, 3/1 (Lorwyn Eclipsed,
// #4458):
//
//	"Discard a card: This creature gains indestructible until end of
//	 turn. Tap it."
//
// Underrealm Lich's "gains indestructible until end of turn. Tap it."
// with Hardened Academic's discard cost, paid as the ability is activated
// (CR 602.2b). The cost has no {T}, so the Elf can activate it while
// tapped or summoning sick (the card's ruling); it still gains
// indestructible, and "tap it" does nothing to a tapped Elf.
// Indestructible (CR 702.12b) ends in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1c77b4da-b580-4ce9-b127-76de674f5694",
		Name:         "Iron-Shield Elf",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Discard a card: This creature gains indestructible until end of turn. Tap it.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.DiscardACard(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.B09SourceStillOnBattlefield(g, item) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				if err := (effects.GrantKeywordUntilEOT{
					Target:   item.SourceCardID,
					Keywords: []string{"indestructible"},
					Label:    "Iron-Shield Elf — indestructible",
				}).Apply(ctx); err != nil {
					return err
				}
				return effects.TapTarget{Target: item.SourceCardID}.Apply(ctx)
			},
		}},
	})
}
