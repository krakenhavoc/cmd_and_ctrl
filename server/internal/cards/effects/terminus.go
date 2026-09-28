package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Terminus — Sorcery {4}{W}{W} (#1665):
//
//	"Put all creatures on the bottom of their owners' libraries.
//	 Miracle {W} (You may cast this card for its miracle cost when you
//	 draw it if it's the first card you drew this turn.)"
//
// The miracle wrath. Six mana is a fair sweeper; one white mana on the
// first draw of a turn — an opponent's turn included, which is the
// point of the keyword — is not.
//
// Not a destroy: indestructible and regeneration are irrelevant, no
// dies-trigger fires, and tokens cease to exist. Every creature leaves
// as ONE simultaneous exit (TuckCardsToLibraryThenForEffect), and each
// owner then orders their own cards on the bottom (the owner chooses
// the relative order). A commander may go to the command zone instead
// (CR 903.9), and one that does is not among the cards its owner
// orders.
func init() {
	Register(Spec{
		OracleID:         "3dd196b6-a85a-4e3e-bb57-ec34241f8117",
		Name:             "Terminus",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Miracle("{W}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return tuckAllCreaturesToBottomByOwner(ctx, item, "Terminus — put these on the bottom of your library in any order")
		},
	})
}

// tuckAllCreaturesToBottomByOwner is Terminus's body: every creature
// on the battlefield to the bottom of its owner's library at once,
// then one ordering prompt per owner with more than one card to order.
func tuckAllCreaturesToBottomByOwner(ctx *Context, item *game.StackItem, label string) error {
	swept := MatchingBattlefield(ctx, Creature())
	ids := make([]uuid.UUID, 0, len(swept))
	for _, c := range swept {
		ids = append(ids, c.InstanceID)
	}
	return ctx.Game.TuckCardsToLibraryThenForEffect(ids, game.TuckOptions{ToBottom: true}, func(g *game.Game, tucked []uuid.UUID) error {
		// Rebuilt from the live *Game, the contract every continuation
		// follows: undo restores this game's fields in place.
		ctx := NewContext(g, item)
		byOwner := map[uuid.UUID][]uuid.UUID{}
		for _, id := range tucked {
			if c, ok := g.LookupCardForEffect(id); ok {
				byOwner[c.Owner] = append(byOwner[c.Owner], id)
			}
		}
		for _, p := range g.Seats {
			if p == nil || p.Eliminated {
				continue
			}
			own := byOwner[p.ID]
			if len(own) < 2 {
				continue
			}
			if err := (PutInLibraryInAnyOrder{
				Chooser:   p.ID,
				Cards:     own,
				From:      game.ZoneLibrary,
				Placement: game.LibraryPlaceBottom,
				Label:     label,
			}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}
