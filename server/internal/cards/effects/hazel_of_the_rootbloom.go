package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hazel of the Rootbloom — Legendary Creature — Squirrel Druid {2}{B}{G},
// 3/5 (Bloomburrow Commander, #5447):
//
//	"{T}, Pay 2 life, Tap X untapped tokens you control: Add X mana in any
//	 combination of colors.
//	 At the beginning of your end step, create a token that's a copy of
//	 target token you control. If that token is a Squirrel, instead create
//	 two tokens that are copies of it."
//
// The first is a mana ability (CR 605.1a), so it announces no X: X is the
// number of tokens the activator taps (TapXUntappedForMana), from none up
// to every untapped token they control, creature or not, and the output
// reads that count off the payment. Tapping a token for the cost is not
// its {T} symbol (CR 302.6), so a token that arrived this turn pays.
// Hazel pays her own {T} and is not a token, so she is never one of the X.
// Each mana is its own colour pick.
//
// The end step trigger targets as it goes on the stack and does nothing
// if the token is gone or no longer yours (CR 608.2b). Whether the token
// is a Squirrel is read as the ability resolves, and both copies are made
// by one creation, so they enter together.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "845b40b1-dc09-44fa-9910-22544c1424ba",
		Name:         "Hazel of the Rootbloom",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{
				Tap:       true,
				Life:      2,
				TapOthers: TapXUntappedForMana("X untapped tokens you control", IsTokenPredicate()),
			},
			ProducedForPaid: func(_ *game.Game, _, _ uuid.UUID, paid game.PaidCost) game.ManaOutput {
				return game.ManaOutput{Amount: game.ManaAmount{
					Options: game.AllColors,
					N:       len(paid.TappedOthers),
				}}
			},
			Label: "{T}, Pay 2 life, Tap X untapped tokens you control: Add X mana in any combination of colors",
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtYourEndStep("Hazel of the Rootbloom — create a token that's a copy of target token you control, or two if it's a Squirrel",
					hazelCopyTargetToken),
				TargetPermanent("target token you control", IsTokenPredicate(), YouControl())),
		},
	})
}

// hazelCopyTargetToken is Hazel's end step trigger resolving.
func hazelCopyTargetToken(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		n := 1
		if c, ok := g.LookupCardForEffect(t.ID); ok && c.HasSubtype(ctx.TextSubtype("Squirrel")) {
			n = 2
		}
		return CreateTokenCopiesTogether{Controller: ctx.Controller(), Copies: []uuid.UUID{t.ID}, N: n}.Apply(ctx)
	}
	return nil
}
