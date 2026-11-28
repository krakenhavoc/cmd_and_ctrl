package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nomad Mythmaker — Creature — Human Nomad Cleric {2}{W}, 2/2 (Judgment,
// #5611):
//
//	"{W}, {T}: Put target Aura card from a graveyard onto the battlefield
//	 under your control attached to a creature you control."
//
// The Aura card is the target; its host is chosen as it enters, among the
// creatures you control its enchant clause admits (CR 303.4f, the Storm
// Herald seam); with none it stays where it was (CR 303.4g).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "591ce7dc-8407-43ed-acfc-e3d8fd547125",
		Name:         "Nomad Mythmaker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, {T}: Put target Aura card from a graveyard onto the battlefield under your control attached to a creature you control.",
			Cost:    game.AbilityCost{Mana: "{W}", Tap: true},
			Targets: effects.TargetCardInGraveyard("target Aura card from a graveyard", effects.HasSubtype("Aura")),
			Effect:  effects.OldB621NomadMythmaker,
		}},
	})
}
