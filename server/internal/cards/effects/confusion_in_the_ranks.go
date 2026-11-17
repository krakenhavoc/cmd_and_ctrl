package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Confusion in the Ranks — Enchantment {3}{R}{R} (Mirrodin, #5739):
//
//	"Whenever an artifact, creature, or enchantment enters, its
//	 controller chooses target permanent another player controls that
//	 shares a card type with it. Exchange control of those permanents."
//
// The trigger watches every permanent that enters, whoever controls it,
// Confusion itself included (CR 603.6a, and the 2004 ruling "Confusion
// in the Ranks triggers on itself entering"). It is Confusion's
// controller's ability, but its target is chosen by the ENTERING
// permanent's controller: OfThatPlayersChoice sends the CR 603.3d prompt
// to the player the trigger's own snapshot of that permanent names as its
// controller (ADR 0019's 2026-10-23 amendment, extended to an enters
// event by #5739). "Another player" is relative to that chooser
// (AnotherPlayerControls), so Confusion's controller's own permanents are
// on offer when someone else's permanent entered. Legality is still
// judged for the ability's controller, so hexproof's "your opponents" is
// read from Confusion's controller's side.
//
// "Shares a card type with it" is read off the entering permanent
// (SharesACardTypeWithIt): an artifact land can take a land (the 2004
// ruling), an artifact creature an artifact or a creature. It is checked
// both when the target is chosen and when the ability resolves (the
// first 2004 ruling), because the CR 608.2b re-check runs the same gate
// against the permanent as it is then. With no legal target the ability
// is removed (CR 603.3d); the choice is not optional when there is one
// (the 2004 ruling).
//
// The exchange is ExchangeControl, one layer-2 effect with two halves. It
// happens only while both permanents are on the battlefield (the third
// 2004 ruling, CR 701.12a): an entering permanent that has left, or left
// and come back as a new object (CR 400.7), is not exchanged. Two
// permanents that have come to have the same controller are not changed
// (CR 701.12b).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "e477abb9-9b99-48a8-8c09-07b120b72540",
		Name:         "Confusion in the Ranks",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{Targeting(
			On(game.EventETB, anArtifactCreatureOrEnchantmentEntered, confusionInTheRanksLabel, confusionInTheRanksExchange),
			AnotherPlayerControls(OfThatPlayersChoice(SharesACardTypeWithIt(
				TargetPermanent("target permanent another player controls that shares a card type with it")))),
		)},
	})
}

const confusionInTheRanksLabel = "Confusion in the Ranks — its controller chooses target permanent another player controls that shares a card type with it; exchange control of those permanents"

// anArtifactCreatureOrEnchantmentEntered is "an artifact, creature, or
// enchantment enters", under anyone's control.
func anArtifactCreatureOrEnchantmentEntered(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventETB || ev.CardID == uuid.Nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && (c.HasCardType("artifact") || c.IsCreature() || c.HasCardType("enchantment"))
}

// confusionInTheRanksExchange exchanges control of the chosen permanent,
// if it is still a legal target, and the permanent that entered, if it is
// still on the battlefield as the same object.
func confusionInTheRanksExchange(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	entered, ok := ctx.TriggeringPermanent()
	if !ok || entered.Left {
		return nil
	}
	for _, ref := range ctx.LegalTargets() {
		if ref.Kind == game.TargetCard {
			return ExchangeControl{A: ctx.Trigger().Object.ID, B: ref.ID, Label: "Confusion in the Ranks — exchange control"}.Apply(ctx)
		}
	}
	return nil
}
