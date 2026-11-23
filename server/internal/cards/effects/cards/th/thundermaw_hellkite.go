package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thundermaw Hellkite — Creature — Dragon {3}{R}{R}, 5/5 (Magic 2013, #5568):
//
//	"Flying
//	 Haste (This creature can attack and {T} as soon as it comes under your
//	 control.)
//	 When this creature enters, it deals 1 damage to each creature with
//	 flying your opponents control. Tap those creatures."
//
// The creatures are the ones with flying your opponents control as the
// trigger resolves, read once before any damage (CR 608.2h); the
// damage is one simultaneous instance (CR 608.2), and then each of
// those creatures still on the battlefield is tapped, the ones the
// damage did not reach included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e958144e-01bc-46ff-8f56-ddf2eb098630",
		Name:            "Thundermaw Hellkite",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(
				effects.WhenThisEnters("Thundermaw Hellkite — 1 damage to each creature with flying your opponents control, then tap them",
					thundermawHellkiteEnters),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, OpponentsOnly: true, Partial: true}}),
		},
	})
}

// thundermawHellkiteEnters deals 1 damage to each creature with flying
// the controller's opponents control, then taps those creatures.
func thundermawHellkiteEnters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var flyers []uuid.UUID
	for _, c := range effects.MatchingBattlefield(ctx, effects.And(effects.Creature(), effects.OpponentControls(), effects.HasKeyword("flying"))) {
		flyers = append(flyers, c.InstanceID)
	}
	if err := g.DamageInstanceForEffect(func() error {
		for _, id := range flyers {
			if err := (effects.DealDamage{Source: item.SourceCardID, Target: id, Amount: 1}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	for _, id := range flyers {
		if !effects.OnBattlefield(g, id) {
			continue
		}
		if err := (effects.TapTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
