package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "1c77b4da-b580-4ce9-b127-76de674f5694",
		Name:         "Iron-Shield Elf",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Discard a card: This creature gains indestructible until end of turn. Tap it.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    DiscardACard(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !b09SourceStillOnBattlefield(g, item) {
					return nil
				}
				ctx := NewContext(g, item)
				if err := (GrantKeywordUntilEOT{
					Target:   item.SourceCardID,
					Keywords: []string{"indestructible"},
					Label:    "Iron-Shield Elf — indestructible",
				}).Apply(ctx); err != nil {
					return err
				}
				return TapTarget{Target: item.SourceCardID}.Apply(ctx)
			},
		}},
	})
}
