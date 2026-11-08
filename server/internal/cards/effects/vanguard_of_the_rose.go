package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "de4678f8-a441-4e4a-a339-9b62ebfb0100",
		Name:         "Vanguard of the Rose",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice another creature or artifact: This creature gains indestructible until end of turn. Tap it.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive | game.AnswerSacOutlet},
			Cost: Plus(ManaCost("{1}"), game.AbilityCost{
				SacrificeOther: Another(sacrificeSpec("another creature or artifact", Or(Creature(), Artifact()))),
			}),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if !b09SourceStillOnBattlefield(g, item) {
					return nil
				}
				if err := (GrantKeywordUntilEOT{
					Target:   item.SourceCardID,
					Keywords: []string{"indestructible"},
					Label:    "Vanguard of the Rose — indestructible",
				}).Apply(ctx); err != nil {
					return err
				}
				return TapTarget{Target: item.SourceCardID}.Apply(ctx)
			},
		}},
	})
}
