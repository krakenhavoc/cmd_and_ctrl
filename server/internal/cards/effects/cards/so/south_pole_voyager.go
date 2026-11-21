package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// South Pole Voyager — Creature — Human Scout Ally {1}{W}, 2/2:
//
//	"Whenever this creature or another Ally you control enters, you gain
//	 1 life. If this is the second time this ability has resolved this
//	 turn, draw a card."
//
// Jake Sisko's shape: the life always, then the draw only on the
// ability's second resolution this turn, counted per object
// (Game.ResolvedThisTurn, CR 400.7), so the third and later resolutions
// gain life and draw nothing.
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c5173332-68ad-456e-ad85-dd7475464d94",
		Name:         "South Pole Voyager",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID == source.InstanceID ||
					effects.B10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, source.TextSubtype("Ally"))
			}, southPoleVoyagerLabel, southPoleVoyagerBody),
		},
	})
}

const southPoleVoyagerLabel = "South Pole Voyager — you gain 1 life, then draw a card the second time this turn"

func southPoleVoyagerBody(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if err := (effects.GainLife{Amount: 1}).Apply(ctx); err != nil {
		return err
	}
	if g.ResolvedThisTurn(item.SourceCardID, southPoleVoyagerLabel) != 2 {
		return nil
	}
	return effects.DrawCards{N: 1}.Apply(ctx)
}
