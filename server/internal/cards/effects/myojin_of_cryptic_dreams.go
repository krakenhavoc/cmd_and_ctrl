package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Myojin of Cryptic Dreams — Legendary Creature — Spirit {5}{U}{U}{U},
// 3/3 (Kamigawa: Neon Dynasty Commander, #5493):
//
//	"Myojin of Cryptic Dreams enters with an indestructible counter on it
//	 if you cast it from your hand.
//	 Remove an indestructible counter from Myojin of Cryptic Dreams: Copy
//	 target permanent spell you control three times. (The copies become
//	 tokens.)"
//
// The counter is Myojin of Night's Reach's entry clause. The copy is
// Lithoform Engine's, three times over and without the new-targets offer
// (the card prints none); a copy of a permanent spell becomes a token as
// it resolves (CR 608.3f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                   "34da0b06-565f-4964-b8bc-dd106278e086",
		Name:                       "Myojin of Cryptic Dreams",
		Completeness:               CompletenessFull,
		EntersWithCountersFromCast: neoMyojinEntry(),
		Activated: []ActivatedAbility{{
			Label:   "Remove an indestructible counter from Myojin of Cryptic Dreams: Copy target permanent spell you control three times.",
			Cost:    neoRemoveIndestructible(),
			Targets: TargetSpell("target permanent spell you control", Permanent(), YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return CopySpell{StackID: id, Controller: item.Controller, Count: 3}.Apply(ctx)
			},
		}},
	})
}
