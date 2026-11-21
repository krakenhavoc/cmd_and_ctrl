package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Morophon, the Boundless — Legendary Creature — Shapeshifter {7}, 6/6:
//
//	"Changeling (This card is every creature type.)
//	 As Morophon enters, choose a creature type.
//	 Spells of the chosen type you cast cost {W}{U}{B}{R}{G} less to cast.
//	 This effect reduces only the amount of colored mana you pay.
//	 Other creatures you control of the chosen type get +1/+1."
//
// The type is chosen as it enters (Herald's Horn's ChooseCreatureTypeAsEnters,
// CR 614.12) and kept on the permanent. The reduction is Bard Class's
// colours-only reduction with a five-colour unit: each symbol takes one mana
// of its colour off the cost where the cost has it and nothing off generic
// mana or a colour the cost lacks (CR 118.7, UnitColorsOnly), for each
// spell its controller casts that is of the chosen type. The anthem is
// Adaptive Automaton's: other creatures the controller controls of the type
// get +1/+1, and with no type named yet neither applies.
func init() {
	const name = "Morophon, the Boundless"
	unit, err := game.ParseCost("{W}{U}{B}{R}{G}")
	if err != nil {
		panic("effects.Morophon: " + err.Error())
	}
	reduction := effects.CostsLess(1, "Spells of the chosen type you cast cost {W}{U}{B}{R}{G} less to cast. This effect reduces only the amount of colored mana you pay.",
		effects.YourSpell(), effects.SpellOfTheSourcesChosenType())
	reduction.Unit = &unit
	reduction.UnitColorsOnly = true
	effects.Register(effects.Spec{
		OracleID:        "bf418a2f-e639-4e52-9cad-5a0713a1829b",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"changeling"},
		AsEnters:        effects.ChooseCreatureTypeAsEnters(name),
		CostModifiers:   []game.CostModifier{reduction},
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Chosen: true, Others: true, YoursOnly: true}, 1, 1),
		},
	})
}
