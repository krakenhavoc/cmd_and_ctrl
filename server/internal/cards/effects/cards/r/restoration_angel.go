package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Restoration Angel — Creature {3}{W}, 3/4:
//
//	"Flash
//	 Flying
//	 When this creature enters, you may exile target non-Angel creature you
//	 control, then return that card to the battlefield under your control."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dfbd3afc-9905-4cff-a4f4-df08a4d0a7fa",
		Name:            "Restoration Angel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(
				effects.Targeting(
					effects.WhenThisEnters("Restoration Angel — exile target non-Angel creature you control, then return it to the battlefield",
						effects.FlickerFirstLegalTarget),
					effects.TargetCreature("target non-Angel creature you control", effects.YouControl(), effects.Not(effects.OfCreatureType("Angel")))),
				"Restoration Angel — exile a non-Angel creature you control, then return it to the battlefield?"),
		},
	})
}
