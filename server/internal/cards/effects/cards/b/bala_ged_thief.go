package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bala Ged Thief — Creature — Human Rogue Ally {3}{B}, 2/2 (Zendikar, S82 OLD
// b494):
//
//	"Whenever this creature or another Ally you control enters, target player
//	 reveals a number of cards from their hand equal to the number of Allies you
//	 control. You choose one of them. That player discards that card."
//
// Klaw, Sonic Subjugator's reveal-and-pick pair (NewRevealFromHandPick) under a
// rally trigger (one per entering Ally) that targets a player. The number is
// the Allies the controller has as the ability resolves, the Thief itself
// counted if it is still there; the revealing player picks which cards to show
// (all of them when the count reaches their hand), the controller picks one,
// and it is discarded.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3a020ae9-279b-4c6c-8d3d-90d7b335fa3c",
		Name:         "Bala Ged Thief",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.S82OldB313Rally("Bala Ged Thief — target player reveals a number of cards from their hand equal to the number of Allies you control; you choose one and they discard it", balaGedThiefTrigger),
				effects.TargetPlayer("target player")),
		},
	})
}

func balaGedThiefTrigger(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	target := effects.TargetedPlayer(ctx)
	if target == uuid.Nil {
		return nil
	}
	return balaGedThiefReveal.Ask(ctx, target, len(effects.Old494AlliesYouControl(ctx, item.Controller)))
}

var balaGedThiefReveal = effects.NewRevealFromHandPick("Bala Ged Thief", "discards",
	"choose-cards/bala-ged-thief-reveal", "choose-cards/bala-ged-thief-discard",
	func(ctx *effects.Context, owner, card uuid.UUID) error {
		return ctx.Game.DiscardCardsKeyedForEffect(owner, []uuid.UUID{card}, ctx.Source(), game.KeyedTail{})
	})
