package cardsso

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spreading Plague — Enchantment {4}{B} (Invasion, #5616):
//
//	"Whenever a creature enters, destroy all other creatures that share a
//	 color with it. They can't be regenerated."
//
// A trigger on any creature entering, whoever controls it. As it resolves
// the entered creature is read through CR 608.2h (as it is now while it is on
// the battlefield, as it last existed once it has gone), and every OTHER
// creature that shares a colour with it, read through the layered view, is
// destroyed together without regeneration. A colourless creature shares no
// colour, so it destroys nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6c134ef7-5c2f-4e4c-ae40-ef338fbf41e6",
		Name:         "Spreading Plague",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, spreadingPlagueCreatureEntered,
				"Spreading Plague — destroy all other creatures that share a color with it", spreadingPlagueDestroy),
		},
	})
}

func spreadingPlagueCreatureEntered(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature()
}

func spreadingPlagueDestroy(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	info, ok := ctx.TriggeringPermanent()
	if !ok || len(info.Characteristic.Colors) == 0 {
		return nil
	}
	colors := info.Characteristic.Colors
	var entered uuid.UUID
	if obj := ctx.Trigger().Object; obj != nil {
		entered = obj.Ref().ID
	}
	return effects.DestroyAllMatching{
		Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
			if c.InstanceID == entered || !c.IsCreature() {
				return false
			}
			for _, col := range c.EffectiveColors() {
				if slices.Contains(colors, col) {
					return true
				}
			}
			return false
		},
		CantBeRegenerated: true,
	}.Apply(ctx)
}
