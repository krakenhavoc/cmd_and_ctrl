package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spider-Man, Hometown Hero — {4}{R} Legendary Creature — Spider Human
// Hero 5/4 (Marvel Super Heroes Commander, #4455):
//
//	"Reach
//	 When Spider-Man enters, target creature with power 2 or less can't be
//	 blocked this turn."
//
// The power bound is part of the target restriction, checked on the stack
// and again at resolution (CR 608.2b): a creature that grew past power 2 in
// response is no longer a legal target and is left alone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ef40224a-5b49-48b0-ad29-c6ce391548f5",
		Name:            "Spider-Man, Hometown Hero",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(
				Targeting(
					WhenThisEnters("Spider-Man — target creature with power 2 or less can't be blocked this turn",
						func(g *game.Game, item *game.StackItem) error {
							ctx := NewContext(g, item)
							id, ok := b16FirstLegalTargetCard(ctx)
							if !ok {
								return nil
							}
							return RestrictUntilEOT{Target: id, Restrictions: game.CantBeBlocked,
								Label: "Spider-Man, Hometown Hero — can't be blocked"}.Apply(ctx)
						}),
					TargetCreature("target creature with power 2 or less", PowerLE(2))),
				ForTargets(CantBeBlockedUntilEOT(0, 0, 0))),
		},
	})
}
