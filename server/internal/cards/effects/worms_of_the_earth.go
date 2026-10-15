package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Worms of the Earth — Enchantment {2}{B}{B}{B}:
//
//	"Players can't play lands.
//	 Lands can't enter the battlefield.
//	 At the beginning of each upkeep, any player may sacrifice two lands
//	 of their choice or have this enchantment deal 5 damage to that
//	 player. If a player does either, destroy this enchantment."
//
// Sentence one is the land-play gate (land_play_restriction.go, ADR
// 0109 §4) and sentence two the entry gate (entry_restriction.go, #1985):
// a land can be neither played nor put onto the battlefield by any
// other means, a token included. Sentence three is ADR 0187 slice 8
// (#3576): one offer per upkeep, made to every player in APNAP order
// from the active player, the controller included (AnyPlayerPays). The
// two ways to pay are ONE offer (EitherPayment, §5), so each player is
// asked once and their answer names the branch. The first payer ends the
// offer and destroys the enchantment; with no payer it stays. Destroying
// it is gated on the source still being the same permanent (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9638ec0c-c8bf-4aa5-a7ca-8677e4b9f241",
		Name:         "Worms of the Earth",
		Completeness: CompletenessFull,
		LandPlayRestrictions: []game.LandPlayRestriction{
			PlayersCantPlayLands("Players can't play lands."),
		},
		EntryRestrictions: []game.EntryRestriction{
			LandsCantEnter("Lands can't enter the battlefield."),
		},
		Triggered: []game.TriggeredAbility{
			AtEachUpkeep("Worms of the Earth — any player may sacrifice two lands or take 5 damage to destroy it",
				func(g *game.Game, item *game.StackItem) error {
					return AnyPlayerPays{
						Pay: EitherPayment(
							SacrificePayment(2, "land", "lands", QueryType("land")),
							DamagePayment(5),
						),
						Question: "Worms of the Earth — sacrifice two lands or take 5 damage to destroy it?",
						OnPaid: func(ctx *Context, _ uuid.UUID) error {
							if !sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
								return nil
							}
							return ctx.Game.DestroyPermanentForEffect(ctx.Item.SourceCardID)
						},
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
