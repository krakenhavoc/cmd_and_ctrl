package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chain of Plasma — Instant {1}{R} (Onslaught):
//
//	"Chain of Plasma deals 3 damage to any target. Then that player or
//	 that permanent's controller may discard a card. If the player does,
//	 they may copy this spell and may choose a new target for that copy."
//
// The Chain cycle's self-copy (chain_of_vapor.go, CR 707.10), priced in
// a card. The player asked is the damaged player, or the damaged
// permanent's controller, read as the spell resolves. Their discard is
// an "up to one" pick of their own (CR 701.9b), so they may discard
// nothing, and a player with an empty hand is asked nothing. The copy
// offer rides the discard's keyed continuation and is made only when a
// card really left their hand. The copy is theirs (CR 707.10) and they
// choose its new target (CR 707.10c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "029d08cb-da4d-4f4e-b667-62712af63260",
		Name:         "Chain of Plasma",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || !ctx.IsTargetLegal(t) {
				return nil
			}
			asked := t.ID
			switch t.Kind {
			case game.TargetCard:
				c, found := ctx.Game.LookupCardForEffect(t.ID)
				if !found {
					return nil
				}
				asked = c.Controller
			case game.TargetSelf:
				asked = ctx.Controller()
			}
			if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 3}).Apply(ctx); err != nil {
				return err
			}
			if ctx.Game.PlayerByIDForEffect(asked) == nil {
				return nil
			}
			return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
				Player:   asked,
				Source:   ctx.Source(),
				N:        1,
				UpTo:     true,
				Question: "Chain of Plasma — you may discard a card (if you do, you may copy this spell)",
				Asked:    game.NewAskedItem(item),
			}, chainOfPlasmaDiscarded, nil)
		},
	})
}

// chainOfPlasmaDiscarded offers the copy when the asked player really
// discarded a card. Keyed (#4253); an on-disk key: never renamed, never
// reused.
var chainOfPlasmaDiscarded = game.RegisterDiscardThen("choose-cards/chain-of-plasma-discarded", func(g *game.Game, r game.DiscardsDone) error {
	if r.Item == nil || !r.Discarded.Discarded(r.Player) {
		return nil
	}
	return chainOfferCopy(g, r.Item, r.Player, "Chain of Plasma")
})
