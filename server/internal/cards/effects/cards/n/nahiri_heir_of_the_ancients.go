package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nahiri, Heir of the Ancients — Legendary Planeswalker — Nahiri {2}{R}{W},
// loyalty 4 (Zendikar Rising, #5508):
//
//	"+1: Create a 1/1 white Kor Warrior creature token. You may attach an
//	 Equipment you control to it.
//	 −2: Look at the top six cards of your library. You may reveal a Warrior
//	 or Equipment card from among them and put it into your hand. Put the
//	 rest on the bottom of your library in a random order.
//	 −3: Nahiri deals damage to target creature or planeswalker equal to
//	 twice the number of Equipment you control."
//
// The +1 is Nahiri, the Lithomancer's create-then-attach (the Equipment is
// offered once the token exists, CR 701.3b). The −2 is Basic Conjuration's
// look-and-reveal over six cards with a Warrior-or-Equipment filter. The −3
// counts the Equipment you control as it resolves (CR 608.2h) and is dealt
// by Nahiri, so protection and lifelink see her.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "39bbc4cf-53fe-47dd-8ba3-1a59b1fa7dd1",
		Name:            "Nahiri, Heir of the Ancients",
		StartingLoyalty: 4,
		Completeness:    effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Create a 1/1 white Kor Warrior creature token. You may attach an Equipment you control to it.",
				Cost:    effects.LoyaltyCost(1),
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return g.CreateTokensKeyedForEffect(game.TokenCreation{
						Controller: item.Controller,
						Source:     item.SourceCardID,
						Groups: []game.TokenGroup{{
							Template: ctx.TextToken(effects.TokenCard("1/1 white Kor Warrior")),
							Count:    1,
						}},
					}, ctx.Tail(effects.S82OldB097NahiriAttachEquipmentTail))
				},
			},
			{
				Label:   "−2: Look at the top six cards of your library. You may reveal a Warrior or Equipment card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.",
				Cost:    effects.LoyaltyCost(-2),
				Purpose: game.Purpose{Tutors: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					player := ctx.Controller()
					return effects.TakeFromLibraryToHand{
						Player: player,
						Cards:  g.LookAtTopOfLibraryForEffect(player, 6),
						Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
							return c.HasSubtypeWord("Warrior") || c.HasSubtypeWord("Equipment")
						},
						Max:      1,
						Optional: true,
						Label:    "Nahiri, Heir of the Ancients — you may reveal a Warrior or Equipment card from among them and put it into your hand",
						Key:      effects.TakeRevealRestBottomRandom,
					}.Apply(ctx)
				},
			},
			{
				Label:   "−3: Nahiri deals damage to target creature or planeswalker equal to twice the number of Equipment you control.",
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					n := 2 * effects.S82OldB097EquipmentYouControl(g, ctx.Controller())
					return effects.DealDamage{Source: item.SourceCardID, Target: id, Amount: n}.Apply(ctx)
				},
			},
		},
	})
}
