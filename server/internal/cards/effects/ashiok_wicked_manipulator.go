package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ashiok, Wicked Manipulator — Legendary Planeswalker — Ashiok {3}{B}{B},
// loyalty 5 (Wilds of Eldraine, #4471):
//
//	"If you would pay life while your library has at least that many cards
//	 in it, exile that many cards from the top of your library instead.
//	 +1: Look at the top two cards of your library. Exile one of them and
//	 put the other into your hand.
//	 −2: Create two 1/1 black Nightmare creature tokens with "At the
//	 beginning of combat on your turn, if a card was put into exile this
//	 turn, put a +1/+1 counter on this token."
//	 −7: Target player exiles the top X cards of their library, where X is
//	 the total mana value of cards you own in exile."
//
// The static is a CR 614 replacement on a life PAYMENT (the new
// ReplacementEvent.LifePayment, set by the one cost-shaped life path, so a
// Phyrexian symbol, a ward or a shockland all read it): see
// youExileFromLibraryInsteadOfPayingLife. The +1 is the player's pick of
// which looked-at card goes to hand, the other is exiled (the same choice
// as picking the card to exile); a one-card library has nothing to put in
// hand and exiles it. The −2 makes the Nightmare of std_b53_ashiok_helpers.go.
// The −7 sums the mana value of the cards the controller owns in exile as it
// resolves, and the targeted player exiles that many from the top (all of
// them when fewer).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "67d66678-a8e6-4f80-bade-1f1daf4b0610",
		Name:            "Ashiok, Wicked Manipulator",
		Completeness:    CompletenessFull,
		StartingLoyalty: 5,
		Replacements: []game.ReplacementEffect{youExileFromLibraryInsteadOfPayingLife(
			"Ashiok, Wicked Manipulator — exile that many cards from the top of your library instead of paying life")},
		Activated: []ActivatedAbility{
			{
				Label: "+1: Look at the top two cards of your library. Exile one of them and put the other into your hand.",
				Cost:  LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					player := ctx.Controller()
					looked := g.LookAtTopOfLibraryForEffect(player, 2)
					switch len(looked) {
					case 0:
						return nil
					case 1:
						return g.ExileCardForEffect(looked[0])
					}
					return TakeFromLibraryToHand{
						Player: player,
						Cards:  looked,
						Max:    1,
						Label:  "Ashiok, Wicked Manipulator — put one of them into your hand and exile the other",
						Key:    ashiokTakeRestIntoExile,
					}.Apply(ctx)
				},
			},
			{
				Label:   "−2: Create two 1/1 black Nightmare creature tokens.",
				Cost:    LoyaltyCost(-2),
				Purpose: game.Purpose{Tokens: 2, CreatureTokens: 2},
				Effect:  Do(CreateToken{Template: Std53NightmareToken(), N: 2}),
			},
			{
				Label:   "−7: Target player exiles the top X cards of their library, where X is the total mana value of cards you own in exile.",
				Cost:    LoyaltyCost(-7),
				Targets: TargetPlayer("target player"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					x := ashiokTotalManaValueOwnedInExile(g, ctx.Controller())
					if x <= 0 {
						return nil
					}
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetPlayer {
							continue
						}
						_, err := g.MillToZoneForEffect(t.ID, x, game.ZoneExile)
						return err
					}
					return nil
				},
			},
		},
	})
}

// ashiokTotalManaValueOwnedInExile is "the total mana value of cards you
// own in exile", read as the ability resolves.
func ashiokTotalManaValueOwnedInExile(g *game.Game, owner uuid.UUID) int {
	total := 0
	for _, c := range g.Exile.Cards {
		if c.Owner == owner {
			total += c.ManaValue()
		}
	}
	return total
}
