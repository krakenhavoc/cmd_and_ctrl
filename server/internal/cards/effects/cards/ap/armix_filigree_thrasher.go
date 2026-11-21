package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Armix, Filigree Thrasher — Legendary Artifact Creature — Golem {2}{B},
// 3/2:
//
//	"Whenever Armix attacks, you may discard a card. When you do, target
//	 creature defending player controls gets -X/-X until end of turn,
//	 where X is the number of artifacts you control plus the number of
//	 artifact cards in your graveyard.
//	 Partner (You can have two commanders if both have partner.)"
//
// Inti's shape: the discard is the trigger's "you may" (MayPay, so an
// empty hand is not asked) and only a card really discarded raises the
// CR 603.12 reflexive trigger. The defending player rides the reflexive
// trigger's params (Auron's shape); the target is chosen as it goes on
// the stack and X is counted as it resolves, so the discarded card, if it
// was an artifact, is already in the graveyard and counts. Partner is a
// deck-building rule read off the oracle text (CR 702.124, #2874).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a1464164-fbda-43b5-988e-6e1da9bd9eef",
		Name:         "Armix, Filigree Thrasher",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Armix, Filigree Thrasher — you may discard a card", func(g *game.Game, item *game.StackItem) error {
				return effects.MayPay{
					Chooser:  item.Controller,
					Pay:      effects.DiscardPayment(1),
					Question: "Armix — discard a card to give target creature defending player controls -X/-X?",
					Then:     armixDiscarded,
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}

// armixDiscarded is the "when you do": raised only once the card was
// really discarded. Keyed: never renamed, never reused.
var armixDiscarded = effects.PayUnlessBranches("pay-unless/armix-filigree-thrasher-discard",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.ReflexiveTrigger{
			Label:  "Armix, Filigree Thrasher — target creature defending player controls gets -X/-X until end of turn",
			Body:   armixShrinkBody,
			Params: game.EffectParams{Player: effects.B17DefendingPlayer(ctx.Game, ctx.Trigger().Event)},
		}.Apply(ctx)
	}, nil)

var armixShrinkBody = game.ReflexiveBody("armix-filigree-thrasher/shrink",
	effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		target, ok := effects.B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		you := ctx.Controller()
		x := effects.B15ArtifactsControlled(g, you) +
			len(effects.GraveyardCardIDs(ctx, you, func(c game.Card) bool { return c.IsArtifact() }))
		return effects.BoostUntilEOT{
			Target:    target,
			Power:     -x,
			Toughness: -x,
			Label:     "Armix, Filigree Thrasher — -X/-X until end of turn",
		}.Apply(ctx)
	}),
	func(_ uuid.UUID, p game.EffectParams) *game.TargetSpec {
		return effects.TargetCreature("target creature defending player controls", effects.ControlledBy(p.Player))
	})
