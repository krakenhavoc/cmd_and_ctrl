package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ceremonial Groundbreaker — Artifact — Equipment {1}{G}{W} (Streets of New
// Capenna, #5491):
//
//	"Equipped creature gets +2/+1 and has trample.
//	 Equip Citizen {1}
//	 Equip {3}"
//
// Dúnedain Blade's two equip abilities: the Citizen-only one is listed first
// with its tighter target clause and cheaper cost, the plain {3} one follows,
// and the restriction is checked while activating (CR 702.6c). Trample is a
// grant to the equipped creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5bb05573-7e8f-471f-908d-04bccead79e3",
		Name:         "Ceremonial Groundbreaker",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 1),
			effects.GrantToAttached("trample"),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipOnlyAbility("Equip Citizen {1}", "{1}",
				effects.TargetCreature("target Citizen you control", effects.YouControl(), effects.Subtype("Citizen"))),
			effects.EquipAbility("{3}"),
		},
	})
}
