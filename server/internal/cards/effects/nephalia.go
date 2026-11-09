package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nephalia — Plane — Innistrad (Planechase Anthology, #4870):
//
//	"At the beginning of your end step, mill seven cards. Then return a
//	 card at random from your graveyard to your hand.
//	 Whenever chaos ensues, return target card from your graveyard to
//	 your hand."
//
// A plane (ADR 0216 §6), controlled by the planar controller. The end step
// ability is one resolution: mill, then a random pick over the whole
// graveyard (the milled cards included, no priority between, the ruling),
// drawn from the keyed random stream. It targets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "72af7529-75bb-488f-801d-7148631b9e9b",
		Name:         "Nephalia",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Nephalia — mill seven cards, then return a card at random from your graveyard to your hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (MillCards{N: 7}).Apply(ctx); err != nil {
						return err
					}
					pick := randomPick(ctx, graveyardIDs(g, item.Controller, nil), 1)
					if len(pick) == 0 {
						return nil
					}
					return g.ReturnFromGraveyardForEffect(pick[0], game.ZoneHand)
				}),
			Targeting(WheneverChaosEnsues("Nephalia — return target card from your graveyard to your hand", returnTargetCardToHand),
				TargetCardInGraveyard("target card from your graveyard", YouOwn())),
		},
	})
}
