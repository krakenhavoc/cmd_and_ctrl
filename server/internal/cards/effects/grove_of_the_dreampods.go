package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grove of the Dreampods — Plane — Fabacin (Commander Masters, #4870):
//
//	"When you planeswalk to Grove of the Dreampods and at the beginning of
//	 your upkeep, reveal cards from the top of your library until you
//	 reveal a creature card. Put that card onto the battlefield and the
//	 rest on the bottom of your library in a random order.
//	 Whenever chaos ensues, return target creature card from your
//	 graveyard to the battlefield."
//
// A plane (ADR 0216 §6). The first ability is one ability with two trigger
// conditions, mandatory (the ruling), over the planar controller's library:
// with no creature card the whole library is revealed and goes to the bottom
// in a random order, which is not a shuffle (the ruling). The chaos ability
// (CR 311.7) returns a creature card from the planar controller's graveyard
// to the battlefield under their control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3e553881-bac7-4830-b739-3b8c6aa0ded6",
		Name:         "Grove of the Dreampods",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenYouPlaneswalkToThisAndAtYourUpkeep("Grove of the Dreampods — reveal until a creature card, put it onto the battlefield and the rest on the bottom in a random order",
				Do(RevealUntilThenPutOntoBattlefield{
					Match:  game.Card.IsCreature,
					Reason: "Grove of the Dreampods — revealed until a creature card",
				})),
			Targeting(WheneverChaosEnsues("Grove of the Dreampods — return target creature card from your graveyard to the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return ReturnFromGraveyard{Target: t, Dest: game.ZoneBattlefield, Controller: item.Controller}.Apply(ctx)
				}), TargetCardInGraveyard("target creature card from your graveyard", Creature(), YouOwn())),
		},
	})
}
