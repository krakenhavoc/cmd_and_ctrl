package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shrine of Limitless Power — Artifact {3} (New Phyrexia):
//
//	"At the beginning of your upkeep and whenever you cast a black spell,
//	 put a charge counter on this artifact.
//	 {4}, {T}, Sacrifice this artifact: Target player discards a card for each charge counter on this artifact."
//
// The counter clause is ONE triggered ability with two trigger conditions.
// The second ability sacrifices the Shrine as a cost, so the number of
// charge counters is its last-known information (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "587461ee-d104-4eef-a4d8-49b1abd56393",
		Name:         "Shrine of Limitless Power",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OldB460ShrineCounterTrigger("Shrine of Limitless Power", "B"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}, Sacrifice this artifact: Target player discards a card for each charge counter on this artifact.",
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetPlayer("target player"),
			Purpose: effects.ForTargets(game.TargetPurpose{Slot: 0, Discards: 1}),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				p, ok := effects.FirstLegalPlayerTarget(ctx)
				if !ok {
					return nil
				}
				n := effects.OldB460ChargeCountersOnSource(g, item.SourceCardID)
				if n <= 0 {
					return nil
				}
				g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: p, Source: item.SourceCardID, N: n})
				return nil
			},
		}},
	})
}
