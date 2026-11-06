package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Provisions Merchant — Creature — Beast Peasant {2}{G}{G}, 3/3 (Wilds of
// Eldraine, #4471):
//
//	"When this creature enters, create a Food token. (It's an artifact
//	 with "{2}, {T}, Sacrifice this token: You gain 3 life.")
//	 Whenever this creature attacks, you may sacrifice a Food. If you do,
//	 attacking creatures get +1/+1 and gain trample until end of turn."
//
// The attack trigger is The Goose Mother's: "you may sacrifice a Food" is
// asked as it resolves (maySacrificeOne, CR 608.2c), over the Foods its
// controller controls then, and with no Food there is no question. Only a
// Food that really went pays out. The payout is one until-end-of-turn
// record over the creatures attacking at that moment (CR 611.2c), the
// Merchant among them if it still is: +1/+1 (layer 7c) and trample
// (layer 6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2f796869-7350-4886-b210-cb5b9cb3e78a",
		Name:         "Provisions Merchant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Provisions Merchant — create a Food token", Do(CreateToken{Template: FoodToken(), N: 1})),
			WheneverThisAttacks("Provisions Merchant — you may sacrifice a Food; if you do, attacking creatures get +1/+1 and gain trample",
				func(g *game.Game, item *game.StackItem) error {
					return maySacrificeOne(g, item, "Provisions Merchant — sacrifice a Food to give attacking creatures +1/+1 and trample?",
						HasSubtype("Food"), false, provisionsMerchantPump)
				}),
		},
	})
}

// provisionsMerchantPump is the "if you do" once a Food has really gone,
// keyed (an on-disk identity: never renamed, never reused).
var provisionsMerchantPump = RegisterSacrificedOne("prompt-run/provisions-merchant-sacrifice-pump", func(ctx *Context, _ uuid.UUID) error {
	return untilEndOfTurn(ctx, uuid.Nil, AttackingCreature(), "Provisions Merchant — attacking creatures get +1/+1 and gain trample",
		game.ModifyPTMod(1, 1), game.AddKeywordsMod("trample"))
})
