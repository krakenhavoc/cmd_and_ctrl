package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Beregond of the Guard — Legendary Creature — Human Soldier {3}{W}, 3/3
// (Tales of Middle-earth Commander):
//
//	"Whenever Beregond or another Human you control enters, creatures you
//	 control get +1/+1 and gain vigilance until end of turn."
//
// One trigger over every Human that enters under your control, Beregond
// included (his own entry counts as a Human entering). The bonus applies to
// the creatures you control as the ability resolves, Flame-Kin Zealot's
// mass pump (CR 611.2c): a creature that enters afterwards gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8ea50101-8392-471e-8523-b12045a7b95b",
		Name:         "Beregond of the Guard",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
				return ok && c.HasSubtype(source.TextSubtype("Human"))
			}, "Beregond of the Guard — creatures you control get +1/+1 and gain vigilance until end of turn", effects.Do(
				effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Toughness: 1, Label: "Beregond of the Guard — +1/+1"},
				effects.GrantKeywordUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Keywords: []string{"vigilance"}, Label: "Beregond of the Guard — vigilance"},
			)),
		},
	})
}
