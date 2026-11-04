package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// White Orchid Phantom — Creature — Spirit Knight {W}{W}, 2/2
// (Secrets of Strixhaven Commander, #4834):
//
//	"Flying, first strike
//	 When this creature enters, destroy up to one target nonbasic land. Its
//	 controller may search their library for a basic land card, put it onto
//	 the battlefield tapped, then shuffle."
//
// Ghost Quarter's victim-side search with the basic entering TAPPED (the
// printed text says so) and an "up to one" nonbasic-land target. The land's
// controller is read before the destroy, since afterwards the land is in a
// graveyard (CR 400.7); the offer is made whether or not the land was
// destroyed (an indestructible land's controller is still asked), as
// printed. With no target chosen the ability does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6edd2a36-3d65-4bf9-bd82-b6d03a7e8fea",
		Name:            "White Orchid Phantom",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("White Orchid Phantom — destroy up to one target nonbasic land; its controller may search for a basic land",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
							return nil
						}
						target := item.Targets[0].ID
						victim, ok := controllerOfTarget(ctx, target)
						if !ok {
							return nil
						}
						if err := (DestroyTarget{Target: target}).Apply(ctx); err != nil {
							return err
						}
						return b07SearchBasicOntoBattlefield(g, item, victim, true, true,
							"White Orchid Phantom — you may search for a basic land")
					}),
				TargetPermanent("up to one target nonbasic land", Land(), b03Nonbasic()).WithCount(0, 1)),
		},
	})
}
