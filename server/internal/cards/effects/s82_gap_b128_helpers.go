package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b128_helpers.go — S82 set GAP, batch 128 (#5948): the Lord of the
// Rings white rows. Append-only.

// gapB128PreventDamageToThisOnYourTurn is "During your turn, prevent all
// damage that would be dealt to <this creature>": a prevention replacement
// (CR 615.1a) that applies while the source's controller is the active
// player (Gideon Blackblade's, on any creature).
func GapB128PreventDamageToThisOnYourTurn(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:    []game.EventKind{game.EventDealDamage},
		Prevention: true,
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			return ev.Kind == game.RepEventDamage && ev.DamageTarget == src.InstanceID && IsActivePlayer(g, src.Controller)
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.Cancel()
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: label,
	}
}

// gapB128SacrificeAFoodOrPay is the either payment "sacrifice a Food or pay
// <mana>" for MayPay.Pay (Anthropede's discard-or-pay shape with a Food
// branch): one prompt whose answer names the branch, and a branch the
// chooser can't pay is not offered (CR 118.3).
func GapB128SacrificeAFoodOrPay(mana string) UpkeepPayment {
	food := game.PayAction{Kind: game.PayActionSacrifice, Count: 1, Of: game.PermanentQuery{Subtypes: []string{"Food"}}, Words: "sacrifice a Food"}
	pay := game.PayAction{Kind: game.PayActionMana, Mana: mana, Words: "pay " + mana}
	return UpkeepPayment{
		Action: &game.PayAction{Kind: game.PayActionEither, Either: []game.PayAction{food, pay}},
		Words:  "Sacrifice a Food or pay " + mana,
	}
}
