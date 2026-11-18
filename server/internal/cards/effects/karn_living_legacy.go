package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "09ab1278-811f-4b97-bd34-7493d6e081e9",
		Name:            "Karn, Living Legacy",
		Completeness:    CompletenessFull,
		StartingLoyalty: 4,
		Emblem: &EmblemSpec{
			Label: "Karn, Living Legacy emblem",
			Text:  "Tap an untapped artifact you control: This emblem deals 1 damage to any target.",
			Activated: []ActivatedAbility{{
				Label:   "Tap an untapped artifact you control: This emblem deals 1 damage to any target.",
				Purpose: ForTargets(DamageToTarget(0, 1)),
				Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
					Count:  1,
					Filter: TargetPermanent("an untapped artifact you control", Artifact()),
					Label:  "an untapped artifact you control",
				}},
				Targets: TargetAny(),
				Effect:  DealDamageToTheTarget(1),
			}},
		},
		Activated: []ActivatedAbility{
			{
				Label:  "+1: Create a tapped Powerstone token.",
				Cost:   LoyaltyCost(1),
				Effect: Do(CreateTokenAdvanced{Spec: Token(PowerstoneToken()).EntersTapped(), N: 1}),
			},
			{
				Label: "−1: Pay any amount of mana. Look at that many cards from the top of your library, then put one of those cards into your hand and the rest on the bottom of your library in a random order.",
				Cost:  LoyaltyCost(-1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return MayPayX{
						Symbol: "{1}",
						Label:  "Karn, Living Legacy",
						Buys:   "look at the top X cards of your library and put one into your hand",
						Goal:   XAsHighAsYouCan,
						Unit:   game.PayAmountCards,
						Then:   karnLivingLegacyPaid,
					}.Apply(NewContext(g, item))
				},
			},
			{
				Label:  "−7: You get an emblem with \"Tap an untapped artifact you control: This emblem deals 1 damage to any target.\"",
				Cost:   LoyaltyCost(-7),
				Effect: Do(CreateEmblem{}),
			},
		},
	})
}

// karnLivingLegacyPaid is the −1 once X is paid: look at the top X,
// one into the hand, the rest to the bottom in a random order.
var karnLivingLegacyPaid = PayUnlessBranches("pay-unless/karn-living-legacy-look",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		if r.Number <= 0 {
			return nil
		}
		player := ctx.Controller()
		return TakeFromLibraryToHand{
			Player: player,
			Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, r.Number),
			Max:    1,
			Label:  "Karn, Living Legacy — put one of those cards into your hand",
			Key:    takeRestBottomRandom,
		}.Apply(ctx)
	}, nil)
