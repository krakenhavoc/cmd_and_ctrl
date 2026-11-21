package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Roadside Assistance — Enchantment — Aura {2}{W}:
//
//	"Enchant creature or Vehicle
//	 When this Aura enters, create a 1/1 colorless Pilot creature token
//	 with "This token saddles Mounts and crews Vehicles as though its
//	 power were 2 greater."
//	 Enchanted permanent gets +1/+1 and has lifelink."
//
// Lightwheel Enhancements' enchant clause with Angelic Gift's enters
// trigger: the Pilot is made by a trigger, so it uses the stack and the
// Aura's controller owns the token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "814f30a1-c0d6-4243-839d-0645d7bcbe4c",
		Name:         "Roadside Assistance",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("enchant creature or Vehicle", effects.Or(effects.Creature(), effects.HasSubtype("Vehicle"))),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantToAttached("lifelink"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Roadside Assistance — create a Pilot token",
				effects.Do(effects.CreateToken{Template: effects.PilotToken(), N: 1})),
		},
	})
}
