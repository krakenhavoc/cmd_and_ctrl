package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Twinflame Travelers — Creature — Elemental Sorcerer {2}{U}{R} (Lorwyn Eclipsed, #4458):
//
//	"Flying
//	 If a triggered ability of another Elemental you control triggers, it
//	 triggers an additional time."
//
// A trigger doubler over abilities whose source is another Elemental you
// control (Harmonic Prodigy's shape). It doubles the triggering only, not
// replacement effects or "as this enters" abilities (the rulings).
func init() {
	d := effects.DoublesAbilitiesOf(nil, effects.DoublesAbilitiesOfOptions{
		SourceMatch: func(_ *game.Game, q game.TriggerDoublingQuery) bool {
			src := effects.CharacteristicCard(q.Source.InstanceID, q.Source, q.SourceLKI)
			return q.Source.InstanceID != q.Doubler.InstanceID && src.HasSubtype(q.Doubler.TextSubtype("Elemental"))
		},
	})
	d.Label = "Twinflame Travelers"
	effects.Register(effects.Spec{
		OracleID:        "341cf69a-246c-4eb5-910c-18c13a4c3dce",
		Name:            "Twinflame Travelers",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		TriggerDoublers: []game.TriggerDoubler{d},
	})
}
