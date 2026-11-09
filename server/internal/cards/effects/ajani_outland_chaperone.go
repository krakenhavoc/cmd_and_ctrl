package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ajani, Outland Chaperone — Legendary Planeswalker — Ajani {1}{W}{W},
// loyalty 3 (Lorwyn Eclipsed, #4458):
//
//	"+1: Create a 1/1 green and white Kithkin creature token.
//	 −2: Ajani deals 4 damage to target tapped creature.
//	 −8: Look at the top X cards of your library, where X is your life
//	     total. You may put any number of nonland permanent cards with mana
//	     value 3 or less from among them onto the battlefield. Then
//	     shuffle."
//
// +1 is a Kithkin token. −2 is Aerial Assault's "target tapped creature"
// (re-checked on resolution) with the damage Ajani's. −8 reads X as the
// ability resolves, looks at that many cards, and offers any number of the
// nonland permanent cards with mana value 3 or less (a card with {X} in its
// cost counts as X = 0); the picks enter as one simultaneous entry, then
// the library is shuffled whether or not anything was put (Undercity's
// put-then-shuffle tail). At 0 life or less nothing is looked at, and the
// library is still shuffled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3ba809ce-3556-415c-9fcc-291cb10a43cd",
		Name:            "Ajani, Outland Chaperone",
		Completeness:    CompletenessFull,
		StartingLoyalty: 3,
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Create a 1/1 green and white Kithkin creature token.",
				Cost:    LoyaltyCost(1),
				Purpose: game.Purpose{Tokens: 1},
				Effect:  Do(CreateToken{Template: TokenCard("1/1 green and white Kithkin"), N: 1}),
			},
			{
				Label:   "−2: Ajani deals 4 damage to target tapped creature.",
				Cost:    LoyaltyCost(-2),
				Targets: TargetCreature("target tapped creature", tappedPermanent()),
				Effect:  DealDamageToTheTarget(4),
			},
			{
				Label:  "−8: Look at the top X cards of your library, where X is your life total. You may put any number of nonland permanent cards with mana value 3 or less from among them onto the battlefield. Then shuffle.",
				Cost:   LoyaltyCost(-8),
				Effect: Do(Instruction(ajaniOutlandChaperoneUltimate)),
			},
		},
	})
}

func ajaniOutlandChaperoneUltimate(ctx *Context) error {
	you := ctx.Controller()
	p := ctx.Game.PlayerByIDForEffect(you)
	if p == nil {
		return nil
	}
	var looked []uuid.UUID
	if p.Life > 0 {
		looked = ctx.Game.LookAtTopOfLibraryForEffect(you, p.Life)
	}
	return PutFromLibraryOntoBattlefield{
		Player:   you,
		Cards:    looked,
		Match:    And(Nonland(), ManaValueLE(3)),
		Optional: true,
		Label:    "Ajani, Outland Chaperone — put any number of nonland permanent cards with mana value 3 or less onto the battlefield",
		ThenTail: ajaniOutlandChaperoneShuffle,
		ThenItem: true,
	}.Apply(ctx)
}

// ajaniOutlandChaperoneShuffle is "Then shuffle", keyed. An on-disk key.
var ajaniOutlandChaperoneShuffle = PutFromLibraryTail("ajani-outland-chaperone/shuffle",
	func(c *Context, res PutFromLibraryResult) error {
		return ShuffleLibrary{Player: res.Player}.Apply(c)
	})
