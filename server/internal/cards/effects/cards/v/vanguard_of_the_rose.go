package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vanguard of the Rose — Creature — Vampire Knight {1}{W}, 3/1 (The Lost
// Caverns of Ixalan, #4470):
//
//	"{1}, Sacrifice another creature or artifact: This creature gains
//	 indestructible until end of turn. Tap it."
//
// Underrealm Lich's "gains indestructible until end of turn. Tap it." with
// Acolyte of Aclazotz's cost: the sacrifice is paid as the ability is
// activated (CR 602.2b), and "another" is object identity, so the
// Vanguard can't pay with itself. Indestructible (CR 702.12b) ends in the
// cleanup step (CR 514.2). The cost has no tap, so the ability can be
// activated the turn the Vanguard enters, and the Vanguard taps as the
// ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "de4678f8-a441-4e4a-a339-9b62ebfb0100",
		Name:         "Vanguard of the Rose",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice another creature or artifact: This creature gains indestructible until end of turn. Tap it.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive | game.AnswerSacOutlet},
			Cost: effects.Plus(effects.ManaCost("{1}"), game.AbilityCost{
				SacrificeOther: effects.Another(effects.SacrificeSpec("another creature or artifact", effects.Or(effects.Creature(), effects.Artifact()))),
			}),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if !effects.B09SourceStillOnBattlefield(g, item) {
					return nil
				}
				if err := (effects.GrantKeywordUntilEOT{
					Target:   item.SourceCardID,
					Keywords: []string{"indestructible"},
					Label:    "Vanguard of the Rose — indestructible",
				}).Apply(ctx); err != nil {
					return err
				}
				return effects.TapTarget{Target: item.SourceCardID}.Apply(ctx)
			},
		}},
	})
}
