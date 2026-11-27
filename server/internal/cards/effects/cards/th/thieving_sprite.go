package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thieving Sprite — Creature — Faerie Rogue {2}{B}, 1/1 (Lorwyn):
//
//	"Flying
//	 When this creature enters, target player reveals X cards from their
//	 hand, where X is the number of Faeries you control. You choose one of
//	 those cards. That player discards that card."
//
// Bala Ged Thief's reveal-and-pick pair (NewRevealFromHandPick) under an enters
// trigger that targets a player. X is the Faeries the controller has as the
// ability resolves, the Sprite itself counted if it is still there (it is a
// Faerie) and any Faerie permanent, creature or not; the revealing player
// picks which cards to show (all of them when X reaches their hand), the
// controller picks one and it is discarded.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4c5b2be9-5d48-4e62-ae67-d61f3708bb4f",
		Name:            "Thieving Sprite",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Thieving Sprite — target player reveals X cards from their hand, X the number of Faeries you control; you choose one and they discard it", thievingSpriteTrigger),
				effects.TargetPlayer("target player")),
		},
	})
}

func thievingSpriteTrigger(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	target := effects.TargetedPlayer(ctx)
	if target == uuid.Nil {
		return nil
	}
	faerie := ctx.TextSubtype("Faerie") // a word of the source's text (CR 612.2)
	x := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.HasSubtype(faerie) {
			x++
		}
	}
	return thievingSpriteReveal.Ask(ctx, target, x)
}

var thievingSpriteReveal = effects.NewRevealFromHandPick("Thieving Sprite", "discards",
	"choose-cards/thieving-sprite-reveal", "choose-cards/thieving-sprite-discard",
	func(ctx *effects.Context, owner, card uuid.UUID) error {
		return ctx.Game.DiscardCardsKeyedForEffect(owner, []uuid.UUID{card}, ctx.Source(), game.KeyedTail{})
	})
