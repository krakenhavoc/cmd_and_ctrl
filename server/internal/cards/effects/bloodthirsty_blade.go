package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bloodthirsty Blade — Artifact — Equipment {2}:
//
//	"Equipped creature gets +2/+0 and is goaded. (It attacks each combat
//	 if able and attacks a player other than you if able.)
//	 {1}: Attach this Equipment to target creature an opponent controls.
//	 Activate only as a sorcery."
//
// The goad is GoadAttached, the static goad Shiny Impetus uses: the goader
// is the Blade's controller, and it lasts exactly as long as the Blade is
// attached. The {1} ability is NOT the equip keyword, so it is an ordinary
// sorcery-speed activated ability with its own target clause (a creature an
// opponent controls) and is not marked Equip: Leonin Shikari does not open
// it. The Blade has no equip cost of its own.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b2747459-05ec-42a6-81be-e97e182bc778",
		Name:         "Bloodthirsty Blade",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(2, 0),
			GoadAttached(),
		},
		Activated: []ActivatedAbility{{
			Label:        "{1}: Attach this Equipment to target creature an opponent controls. Activate only as a sorcery.",
			Cost:         ManaCost("{1}"),
			Targets:      TargetCreature("target creature an opponent controls", OpponentControls()),
			SorcerySpeed: true,
			Effect:       AttachSourceToTarget,
		}},
	})
}
