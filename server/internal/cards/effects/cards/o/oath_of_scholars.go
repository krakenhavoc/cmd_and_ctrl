package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Oath of Scholars — Enchantment {3}{U} (Exodus, S82 OLD batch 742, #5626):
//
//	"At the beginning of each player's upkeep, that player chooses target
//	 player who has more cards in hand than they do and is their opponent. The
//	 first player may discard their hand and draw three cards."
//
// Oath of Druids' shape: the Oath's controller controls the trigger, but the
// upkeep player chooses its target (OfThatPlayersChoice), judged against the
// upkeep player: an opponent of theirs with more cards in hand
// (WithMoreCardsInHandThanTheChooser, #5626). With no such player the trigger
// is removed as it would go on the stack (CR 603.3d), and a target that no
// longer qualifies at resolution makes it do nothing (CR 608.2b). The target
// is only a condition. On resolution the upkeep player is asked, and on a yes
// discards their hand and draws three.
//
// No simplification.
func init() {
	const name = "Oath of Scholars"
	effects.Register(effects.Spec{
		OracleID:     "da4a5129-b005-4fc6-b175-0416627881db",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{effects.Targeting(
			effects.AtEachUpkeep(name+" — that player may discard their hand and draw three cards", oathOfScholarsAsk),
			effects.WithMoreCardsInHandThanTheChooser(effects.OfThatPlayersChoice(
				effects.TargetPlayer("target player who has more cards in hand than they do and is their opponent"))),
		)},
	})
}

// oathOfScholarsAsk asks the upkeep player whether to discard their hand.
func oathOfScholarsAsk(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	first := item.Trigger.Event.Actor
	if first == uuid.Nil || g.PlayerByIDForEffect(first) == nil {
		return nil
	}
	return effects.MayChoice{
		Player:   first,
		Question: "Oath of Scholars — discard your hand and draw three cards?",
		YesLabel: "Discard and draw three",
		NoLabel:  "Keep your hand",
		Then:     oathOfScholarsDiscardDraw,
	}.Apply(effects.NewContext(g, item))
}

// oathOfScholarsDiscardDraw is the upkeep player's yes. An on-disk key: never
// renamed, never reused.
var oathOfScholarsDiscardDraw = effects.MayYes("confirm/oath-of-scholars-discard-hand-draw-three", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	if _, err := effects.DiscardWholeHand(ctx.Game, r.Chooser); err != nil {
		return err
	}
	return effects.DrawCards{Player: r.Chooser, N: 3}.Apply(ctx)
})
