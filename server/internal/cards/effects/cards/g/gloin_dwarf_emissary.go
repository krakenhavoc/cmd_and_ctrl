package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const gloinLabel = "Glóin, Dwarf Emissary — create a Treasure token"

// Glóin, Dwarf Emissary — Legendary Creature — Dwarf Advisor {2}{R}, 3/3
// (The Lord of the Rings: Tales of Middle-earth Commander):
//
//	"Whenever you cast a historic spell, create a Treasure token. This
//	 ability triggers only once each turn.
//	 {T}, Sacrifice a Treasure: Goad target creature."
//
// Artificer's Assistant's historic cast watch with Nimrodel Watcher's
// once-a-turn limit (the tally counts an ability the moment it is put on
// the stack). The activation is Farid's goad, paid by tapping Glóin and
// sacrificing any Treasure artifact its controller has.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "29375a7a-af4e-46aa-8d1c-e8d22d3c0ea7",
		Name:         "Glóin, Dwarf Emissary",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.YouCastAHistoricSpell(ev, source, lki, g) && !effects.B11TriggeredThisTurn(g, source.InstanceID, gloinLabel)
			}, gloinLabel, effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice a Treasure: Goad target creature.",
			Cost:    effects.Plus(effects.TapCost(), game.AbilityCost{SacrificeOther: effects.SacrificeSpec("a Treasure", effects.HasSubtype("Treasure"))}),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.GoadTarget,
		}},
	})
}
