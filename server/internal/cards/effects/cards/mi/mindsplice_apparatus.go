package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mindsplice Apparatus — Artifact {3}{U}:
//
//	"Flash
//	 At the beginning of your upkeep, put an oil counter on this
//	 artifact.
//	 Instant and sorcery spells you cast cost {1} less to cast for each
//	 oil counter on this artifact."
//
// Animar's scaling discount on Vat of Rebirth's oil counters: the
// amount hook reads the counters on the modifier's own source, so the
// discount grows by one each of your upkeeps. It spends generic mana
// only (CR 601.2f), so no number of counters pays a coloured pip.
// The upkeep counter is mandatory, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4d0760fb-4e04-4f29-a268-14c13c30232d",
		Name:            "Mindsplice Apparatus",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Mindsplice Apparatus — put an oil counter on it", effects.PutACounterOnThis(MindspliceOilCounter)),
		},
		CostModifiers: []game.CostModifier{
			effects.CostsLessEach(effects.CountersOnSource(MindspliceOilCounter),
				"Instant and sorcery spells you cast cost {1} less to cast for each oil counter on this artifact.",
				effects.YourSpell(), effects.InstantOrSorcerySpell()),
		},
	})
}

const MindspliceOilCounter = "oil"
