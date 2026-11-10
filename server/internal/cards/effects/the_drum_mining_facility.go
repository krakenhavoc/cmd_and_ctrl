package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Drum, Mining Facility — Plane — Earth (Doctor Who, #4870):
//
//	"Whenever you roll the planar die, creatures you control get +1/+1
//	 and gain haste until end of turn.
//	 Whenever chaos ensues, end the turn. (Exile all spells and abilities
//	 from the stack. The player whose turn it is discards down to their
//	 maximum hand size. Damage wears off, and "this turn" and "until end
//	 of turn" effects end.)"
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6). The
// first ability watches the planar die for any face rolled by the planar
// controller (a die roll with a planar face, CR 901.9); the set of creatures
// is locked as it resolves (CR 611.2c). The chaos ability (CR 311.7) ends the
// turn (CR 724.1, EndTheTurn): the stack is exiled, the first ability
// included if it is still waiting, and the rulings' cleanup follows.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9f7e1607-34b5-4332-81e3-391fb49cd6ad",
		Name:         "The Drum, Mining Facility",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventRollDie, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.PlanarFace != "" && ev.Actor == source.Controller
			}, "The Drum, Mining Facility — creatures you control get +1/+1 and gain haste until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					match := And(Creature(), YouControl())
					if err := (BoostUntilEOT{Match: match, Power: 1, Toughness: 1, Label: "The Drum, Mining Facility — +1/+1"}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{Match: match, Keywords: []string{"haste"}, Label: "The Drum, Mining Facility — haste"}.Apply(ctx)
				}),
			WheneverChaosEnsues("The Drum, Mining Facility — end the turn", Do(EndTheTurn{})),
		},
	})
}
