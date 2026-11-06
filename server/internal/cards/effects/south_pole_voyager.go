package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "c5173332-68ad-456e-ad85-dd7475464d94",
		Name:         "South Pole Voyager",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID == source.InstanceID ||
					b10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, source.TextSubtype("Ally"))
			}, southPoleVoyagerLabel, southPoleVoyagerBody),
		},
	})
}

const southPoleVoyagerLabel = "South Pole Voyager — you gain 1 life, then draw a card the second time this turn"

func southPoleVoyagerBody(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (GainLife{Amount: 1}).Apply(ctx); err != nil {
		return err
	}
	if g.ResolvedThisTurn(item.SourceCardID, southPoleVoyagerLabel) != 2 {
		return nil
	}
	return DrawCards{N: 1}.Apply(ctx)
}
