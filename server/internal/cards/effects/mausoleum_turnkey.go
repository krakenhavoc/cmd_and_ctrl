package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mausoleum Turnkey — Creature — Ogre Rogue {3}{B}, 3/2:
//
//	"When this creature enters, return target creature card of an
//	 opponent's choice from your graveyard to your hand."
//
// The target is an opponent's choice (#2055, ADR 0019's 2026-10-16
// amendment): as the trigger is put on the stack (CR 603.3d) its
// controller chooses an opponent, and that opponent chooses the creature
// card in the controller's graveyard. With no creature card there the
// trigger is removed from the stack. "Your" is the trigger's
// controller's.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "d4a61685-7149-4d14-bc17-85b060945687",
		Name:         "Mausoleum Turnkey",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Mausoleum Turnkey — return a creature card of an opponent's choice from your graveyard to your hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t, ok := ctx.ClauseTarget(0)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}.Apply(ctx)
				}),
				OfAnOpponentsChoice(TargetCardInGraveyard("target creature card of an opponent's choice from your graveyard", Creature(), YouOwn()))),
		},
	})
}
