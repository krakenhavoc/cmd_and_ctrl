package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Malignant Growth — Enchantment {3}{G}{U} (Mirage, S82 OLD b805):
//
//	"Cumulative upkeep {1}
//	 At the beginning of your upkeep, put a growth counter on this
//	 enchantment.
//	 At the beginning of each opponent's draw step, that player draws an
//	 additional card for each growth counter on this enchantment, then this
//	 enchantment deals damage to the player equal to the number of cards they
//	 drew this way."
//
// The cumulative upkeep is the keyword's constructor (CR 702.24a) and the
// growth counters are an ordinary counter kind. The draw trigger reads the
// counters as it resolves (last-known if the enchantment has left, CR 608.2h),
// draws that many through the draw primitive and measures what was actually
// drawn by the drawing player's draw tally, so a library that runs dry, or a
// draw a replacement effect took, deals no damage for the missing cards.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b3235d21-ce52-4a99-a2fe-99b0fe7cc8bc",
		Name:         "Malignant Growth",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeep("Malignant Growth — cumulative upkeep {1}", "{1}"),
			effects.AtYourUpkeep("Malignant Growth — put a growth counter on this enchantment",
				effects.Do(effects.CounterOnThis{Kind: "growth", N: 1})),
			effects.On(game.EventBeginDrawStep, effects.ByAnOpponent,
				"Malignant Growth — that player draws an additional card for each growth counter, then this enchantment deals that much damage to them",
				func(g *game.Game, item *game.StackItem) error {
					player := effects.TriggeringActor(item)
					n := effects.TrekB5CountersOnThis(g, item.SourceCardID, "growth")
					if player == uuid.Nil || n <= 0 {
						return nil
					}
					ctx := effects.NewContext(g, item)
					before := g.TurnTallyFor(player).CardsDrawn
					if err := (effects.DrawCards{Player: player, N: n}).Apply(ctx); err != nil {
						return err
					}
					return effects.DealDamage{
						Source: item.SourceCardID,
						Target: player,
						Amount: g.TurnTallyFor(player).CardsDrawn - before,
					}.Apply(ctx)
				}),
		},
	})
}
