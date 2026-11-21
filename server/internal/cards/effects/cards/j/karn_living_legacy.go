package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Karn, Living Legacy — Legendary Planeswalker — Karn {4}, loyalty 4
// (Dominaria United, #5487):
//
//	"+1: Create a tapped Powerstone token. (It's an artifact with "{T}:
//	 Add {C}. This mana can't be spent to cast a nonartifact spell.")
//	 −1: Pay any amount of mana. Look at that many cards from the top of
//	 your library, then put one of those cards into your hand and the
//	 rest on the bottom of your library in a random order.
//	 −7: You get an emblem with "Tap an untapped artifact you control:
//	 This emblem deals 1 damage to any target.""
//
// The +1 is Argothian Opportunist's tapped Powerstone.
//
// The −1's payment happens as the ability resolves (CR 608.2d): the
// amount is chosen then, from 0 to the most Karn's controller can pay,
// as Shanna, Purifying Blade's MayPayX with Symbol "{1}" — any mana pays
// a generic symbol. Paying 0 is paying "any amount", and looks at no
// cards, so declining the payment and paying nothing are the same
// thing. The look is Fomori Vault's: one card into the hand, which is
// mandatory, and the rest under the library in a random order.
//
// The −7's emblem has an activated ability (#5868, CR 114.4): its owner
// activates it from the command zone, and the emblem is the source of
// the damage. "Tap an untapped artifact you control" is Urza, Lord High
// Artificer's TapOthers cost — not {T}, so an artifact creature that
// arrived this turn may pay it (CR 302.6 is about the {T} symbol).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "09ab1278-811f-4b97-bd34-7493d6e081e9",
		Name:            "Karn, Living Legacy",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Emblem: &effects.EmblemSpec{
			Label: "Karn, Living Legacy emblem",
			Text:  "Tap an untapped artifact you control: This emblem deals 1 damage to any target.",
			Activated: []effects.ActivatedAbility{{
				Label:   "Tap an untapped artifact you control: This emblem deals 1 damage to any target.",
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
				Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
					Count:  1,
					Filter: effects.TargetPermanent("an untapped artifact you control", effects.Artifact()),
					Label:  "an untapped artifact you control",
				}},
				Targets: effects.TargetAny(),
				Effect:  effects.DealDamageToTheTarget(1),
			}},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:  "+1: Create a tapped Powerstone token.",
				Cost:   effects.LoyaltyCost(1),
				Effect: effects.Do(effects.CreateTokenAdvanced{Spec: effects.Token(effects.PowerstoneToken()).EntersTapped(), N: 1}),
			},
			{
				Label: "−1: Pay any amount of mana. Look at that many cards from the top of your library, then put one of those cards into your hand and the rest on the bottom of your library in a random order.",
				Cost:  effects.LoyaltyCost(-1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.MayPayX{
						Symbol: "{1}",
						Label:  "Karn, Living Legacy",
						Buys:   "look at the top X cards of your library and put one into your hand",
						Goal:   effects.XAsHighAsYouCan,
						Unit:   game.PayAmountCards,
						Then:   karnLivingLegacyPaid,
					}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:  "−7: You get an emblem with \"Tap an untapped artifact you control: This emblem deals 1 damage to any target.\"",
				Cost:   effects.LoyaltyCost(-7),
				Effect: effects.Do(effects.CreateEmblem{}),
			},
		},
	})
}

// karnLivingLegacyPaid is the −1 once X is paid: look at the top X,
// one into the hand, the rest to the bottom in a random order.
var karnLivingLegacyPaid = effects.PayUnlessBranches("pay-unless/karn-living-legacy-look",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		if r.Number <= 0 {
			return nil
		}
		player := ctx.Controller()
		return effects.TakeFromLibraryToHand{
			Player: player,
			Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, r.Number),
			Max:    1,
			Label:  "Karn, Living Legacy — put one of those cards into your hand",
			Key:    effects.TakeRestBottomRandom,
		}.Apply(ctx)
	}, nil)
