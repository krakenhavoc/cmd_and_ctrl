package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rat King, Verminister — Legendary Creature — Rat Avatar {1}{B}, 1/1
// (Teenage Mutant Ninja Turtles, #4457):
//
//	"Disappear — At the beginning of your end step, if a permanent left
//	 the battlefield under your control this turn, create a 1/1 black Rat
//	 creature token and put a +1/+1 counter on Rat King.
//	 {T}, Sacrifice three Rats: Return target creature card and all other
//	 cards with the same name as that card from your graveyard to the
//	 battlefield tapped."
//
// Disappear is an ability word (CR 207.2c) for revolt's condition, so
// the first ability is AtYourEndStepIfRevolt: checked as the end step
// begins and again as it resolves (CR 603.4). The token is made even
// when Rat King has left by then; the counter goes on him only while he
// is still on the battlefield.
//
// The cost is SacrificeN over Rats, Rat King himself among them (he is
// a Rat). The return reads the target's name as the ability resolves
// and brings it back with every other card of that name in the
// controller's graveyard as one simultaneous entry
// (ReturnFromGraveyardTogether), each tapped, under its owner's control.
// "All other cards" is every card type: a noncreature card that shares
// the name comes back too when it is a permanent card. A target that has
// left the graveyard makes the ability do nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a5656dab-5ce6-48fd-9eae-2c06969bb8df",
		Name:         "Rat King, Verminister",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourEndStepIfRevolt("Rat King, Verminister — Disappear: create a 1/1 black Rat and put a +1/+1 counter on Rat King",
				Do(CreateToken{Template: TokenCard("1/1 black Rat"), N: 1}, CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice three Rats: Return target creature card and all other cards with the same name as that card from your graveyard to the battlefield tapped.",
			Cost:    Plus(TapCost(), SacrificeN(3, "three Rats", Subtype("Rat"))),
			Targets: TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				ids := legalTargetCardIDs(ctx)
				if len(ids) == 0 {
					return nil
				}
				target, ok := g.LookupCardForEffect(ids[0])
				if !ok || target.Name == "" {
					return nil
				}
				all := append(ids[:1:1], graveyardCardIDs(ctx, item.Controller, func(c game.Card) bool {
					return c.InstanceID != target.InstanceID && strings.EqualFold(c.Name, target.Name)
				})...)
				return ReturnFromGraveyardTogether{Targets: all, Tapped: true}.Apply(ctx)
			},
		}},
	})
}
