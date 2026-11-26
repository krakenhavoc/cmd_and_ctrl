package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Disciple of Phenax — Creature — Human Cleric {2}{B}{B}, 1/3 (Theros):
//
//	"When this creature enters, target player reveals a number of cards from
//	 their hand equal to your devotion to black. You choose one of them. That
//	 player discards that card. (Each {B} in the mana costs of permanents you
//	 control counts toward your devotion to black.)"
//
// Acquisitions Expert's reveal-N-and-pick flow: devotion is counted as the
// trigger resolves (CR 700.5, the Disciple's own {B}{B} included while it is
// there), the target player chooses which cards to reveal (a hand no larger
// than the count is revealed whole), only those are revealed, and the
// Disciple's controller picks one of them to be discarded.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d7cbe54b-f217-45f5-8583-c254a1c80bb8",
		Name:         "Disciple of Phenax",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Disciple of Phenax — target player reveals cards from their hand equal to your devotion to black; you choose one and they discard it", disciplePhenaxEnters),
				effects.TargetPlayer("target player")),
		},
	})
}

func disciplePhenaxEnters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	return disciplePhenaxReveal.Ask(ctx, effects.TargetedPlayer(ctx), effects.DevotionTo(g, item.Controller, ctx.TextColor("B")))
}

// disciplePhenaxReveal is the reveal-N-and-pick flow. Both keys are on-disk
// identities: never renamed, never reused.
var disciplePhenaxReveal = effects.NewRevealFromHandPick("Disciple of Phenax", "discards",
	"choose-cards/disciple-of-phenax-reveal", "choose-cards/disciple-of-phenax-discard",
	func(ctx *effects.Context, owner, card uuid.UUID) error {
		return ctx.Game.DiscardCardsKeyedForEffect(owner, []uuid.UUID{card}, ctx.Source(), game.KeyedTail{})
	})
