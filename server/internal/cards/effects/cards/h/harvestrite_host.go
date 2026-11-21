package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Harvestrite Host — Creature — Rabbit Citizen {2}{W}, 3/3 (Bloomburrow,
// #4467):
//
//	"Whenever this creature or another Rabbit you control enters, target
//	 creature you control gets +1/+0 until end of turn. Then draw a card
//	 if this is the second time this ability has resolved this turn."
//
// South Pole Voyager's shape with a Rabbit in place of an Ally and a
// targeted +1/+0 in place of the life: the pump on each resolution, the
// draw only on the ability's second resolution this turn, counted per
// object (Game.ResolvedThisTurn, CR 400.7), so the third and later draw
// nothing. "Rabbit" is a word of the Host's text (CR 612.2). A target
// that left in response removes the ability (CR 608.2b), so it neither
// pumps nor counts as a resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "56208bdc-2eed-4ccc-8c34-c0eba9db0aa9",
		Name:         "Harvestrite Host",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID == source.InstanceID ||
					effects.B10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, source.TextSubtype("Rabbit"))
			}, harvestriteHostLabel, harvestriteHostBody),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}

const harvestriteHostLabel = "Harvestrite Host — target creature you control gets +1/+0, then draw a card the second time this turn"

func harvestriteHostBody(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if id := effects.FirstLegalBattlefieldTarget(ctx); id != uuid.Nil {
		if err := (effects.BoostUntilEOT{Target: id, Power: 1, Label: "Harvestrite Host — +1/+0"}).Apply(ctx); err != nil {
			return err
		}
	}
	if g.ResolvedThisTurn(item.SourceCardID, harvestriteHostLabel) != 2 {
		return nil
	}
	return effects.DrawCards{N: 1}.Apply(ctx)
}
