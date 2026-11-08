package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sandworm — Creature — Worm {4}{R}, 5/4 (Final Fantasy, #4462):
//
//	"Haste
//	 When this creature enters, destroy target land. Its controller may
//	 search their library for a basic land card, put it onto the
//	 battlefield tapped, then shuffle."
//
// The enters trigger is Sundering Eruption's body without its rider:
// the land's controller is read before it dies, the search is an
// optional prompt addressed to them and happens whether or not the land
// actually died (indestructible, regeneration), and the basic enters
// tapped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e92185f8-579c-4645-9cb8-ef5deb9eebc8",
		Name:            "Sandworm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Sandworm — destroy target land; its controller may search for a basic land, tapped",
				func(g *game.Game, item *game.StackItem) error {
					if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
						return nil
					}
					ctx := NewContext(g, item)
					target := item.Targets[0].ID
					victim, ok := controllerOfTarget(ctx, target)
					if !ok {
						return nil
					}
					if err := (DestroyTarget{Target: target}).Apply(ctx); err != nil {
						return err
					}
					return SearchLibrary{
						Player:        victim,
						Predicate:     IsBasicLand,
						Dest:          game.ZoneBattlefield,
						Limit:         1,
						Shuffle:       true,
						TappedOnEntry: true,
						Optional:      true,
						Reason:        "Sandworm — you may search for a basic land",
					}.Apply(ctx)
				}),
				TargetPermanent("target land", Land())),
		},
	})
}
