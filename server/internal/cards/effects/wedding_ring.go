package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wedding Ring — Artifact {2}{W}{W}:
//
//	"When this artifact enters, if it was cast, target opponent creates
//	 a token that's a copy of it.
//	 Whenever an opponent who controls an artifact named Wedding Ring
//	 draws a card during their turn, you draw a card.
//	 Whenever an opponent who controls an artifact named Wedding Ring
//	 gains life during their turn, you gain that much life."
//
// "If it was cast" is a CR 603.4 intervening if read off the
// permanent's own cast record (CR 400.7d), so the token copy — which
// was never cast — does not make a copy in turn, and a Ring that was
// put onto the battlefield without being cast offers no copy at all.
// The token is a real copy with the Ring's oracle ID, so it carries
// both watchers: each of the two rings draws or gains for its own
// controller when the OTHER player's ring-holder does it, which is
// the whole card.
//
// "Who controls an artifact named Wedding Ring" is read when the
// draw or the life gain happens, by effective name, and "during their
// turn" is the drawer's or gainer's own turn. The life gained is the
// amount on the event, so a prevented or replaced gain pays nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0c34e962-99d9-4163-b852-4f61886546aa",
		Name:         "Wedding Ring",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(
				On(game.EventETB, AllOf(Self, weddingRingWasCast),
					"Wedding Ring — target opponent creates a token that's a copy of it",
					weddingRingCopyForOpponent),
				TargetPlayer("target opponent", Opponent())),
				ForTargets(game.TargetPurpose{Slot: 0, Tokens: 1})),
			On(game.EventDrawCard, weddingRingOpponentDrew,
				"Wedding Ring — you draw a card",
				Do(DrawCards{N: 1})),
			On(game.EventChangeLife, weddingRingOpponentGainedLife,
				"Wedding Ring — you gain that much life",
				func(g *game.Game, item *game.StackItem) error {
					return GainLife{Player: item.Controller, Amount: item.Trigger.Event.Amount}.Apply(NewContext(g, item))
				}),
		},
	})
}

// weddingRingWasCast is the enter trigger's intervening if.
func weddingRingWasCast(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return source.Provenance.FromZone != ""
}

// weddingRingCopyForOpponent makes the chosen opponent a token copy
// of the Ring.
func weddingRingCopyForOpponent(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		return CreateTokenCopy{Controller: t.ID, Copy: item.SourceCardID, N: 1}.Apply(ctx)
	}
	return nil
}

// weddingRingOpponentDrew — an opponent who controls a Wedding Ring
// drew a card during their own turn.
func weddingRingOpponentDrew(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.Actor != uuid.Nil && weddingRingHolderOnTheirTurn(g, source, ev.Actor)
}

// weddingRingOpponentGainedLife — an opponent who controls a Wedding
// Ring gained life during their own turn.
func weddingRingOpponentGainedLife(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.Amount > 0 && ev.Target != uuid.Nil && weddingRingHolderOnTheirTurn(g, source, ev.Target)
}

// weddingRingHolderOnTheirTurn is the shared clause: `player` is an
// opponent of the source's controller, controls an artifact named
// Wedding Ring, and it is their turn.
func weddingRingHolderOnTheirTurn(g *game.Game, source *game.Card, player uuid.UUID) bool {
	if player == source.Controller || !IsYourTurn(g, player) {
		return false
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsArtifact() && c.Effective().Name == "Wedding Ring" {
			return true
		}
	}
	return false
}
