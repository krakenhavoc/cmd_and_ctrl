package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stitch Together — Sorcery {B}{B} (EDHREC rank 1630):
//
//	"Return target creature card from your graveyard to your hand.
//	 Threshold — Return that card from your graveyard to the
//	 battlefield instead if there are seven or more cards in your
//	 graveyard."
//
// Raise Dead that becomes a Zombify once the graveyard is full. One
// target clause; the destination is decided at resolution by the
// graveyard count (CR 608.2, CR 608.2h — the count is read once, when
// the effect is applied), which excludes the spell itself: it is on
// the stack, not in the graveyard, while it resolves. Threshold is an
// ability word with no rules meaning of its own (CR 207.2c). "From
// your graveyard" means the card returns under its owner's control,
// which is the caster's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bc3d5911-3580-4132-9daf-2826495b5739",
		Name:         "Stitch Together",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target creature card in your graveyard", Creature(), YouOwn()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			legal := ctx.LegalTargets()
			if len(legal) == 0 || legal[0].Kind != game.TargetCard {
				return nil
			}
			dest := game.ZoneHand
			if p := ctx.PlayerByID(ctx.Controller()); p != nil && p.Graveyard != nil && p.Graveyard.Size() >= 7 {
				dest = game.ZoneBattlefield
			}
			return ReturnFromGraveyard{Target: legal[0].ID, Dest: dest}.Apply(ctx)
		},
	})
}
