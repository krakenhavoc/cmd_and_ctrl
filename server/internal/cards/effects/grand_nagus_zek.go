package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grand Nagus Zek — Legendary Creature — Ferengi Noble {2}{B}, 1/1
// (Star Trek, #4973 batch 12):
//
//	"Grand Nagus Zek gets +1/+1 for each Treasure you control.
//	 Whenever another creature dies, create a Treasure token. This
//	 ability triggers only once each turn."
//
// The bonus is a layer 7c characteristic-defining read of the board
// (Molten Man's shape), so a Treasure cracked for mana shrinks him at
// once and one created by the trigger grows him as it lands.
//
// "Another creature" is any creature, yours or an opponent's, token or
// not, but never Zek himself. "Only once each turn" is the per-turn
// trigger tally (b11TriggeredThisTurn), which counts the trigger when
// it is put on the stack, so a countered one still used up the turn.
// A board wipe kills many creatures in one event batch; OncePerBatch
// keeps that to the one trigger as well, since the tally only moves
// once the first is queued.
//
// No simplification.
func init() {
	const label = "Grand Nagus Zek — create a Treasure token"
	Register(Spec{
		OracleID:     "654ce8c5-b329-4046-9728-7fa276bc8695",
		Name:         "Grand Nagus Zek",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Grand Nagus Zek gets +1/+1 for each Treasure you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				treasure := source.TextSubtype("Treasure")
				n := permanentsYouControl(g, source, func(o *game.Card) bool { return o.HasSubtype(treasure) })
				c.Power += n
				c.Toughness += n
			},
		}},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.CardID == source.InstanceID {
					return false
				}
				if _, ok := diedCreature(ev, g); !ok {
					return false
				}
				return !b11TriggeredThisTurn(g, source.InstanceID, label)
			}, label, Do(CreateToken{Template: TreasureToken(), N: 1}))),
		},
	})
}
