package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scourge of the Nobilis — Enchantment — Aura {2}{R/W}:
//
//	"Enchant creature
//	 As long as enchanted creature is red, it gets +1/+1 and has "{R/W}: This
//	 creature gets +1/+0 until end of turn."
//	 As long as enchanted creature is white, it gets +1/+1 and has lifelink."
//
// Shield of the Oversoul's shape: conditional statics read live off the
// enchanted creature's colours. The granted firebreathing is a bundle
// (Presence of Gond's) whose recipient test is the colour condition, so it
// ends the moment the creature stops being red or the Aura leaves; the
// creature, not the Aura, is the source of the pump.
func init() {
	const grant = "scourge-of-the-nobilis/firebreathing"
	const text = "{R/W}: This creature gets +1/+0 until end of turn."
	red := effects.B43AttachedHasColor("R")
	white := effects.B43AttachedHasColor("W")
	effects.Register(effects.Spec{
		OracleID:     "4a1126a3-f8ce-4790-aa51-1e3f4b63eb18",
		Name:         "Scourge of the Nobilis",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:   text,
				Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
				Cost:    effects.ManaCost("{R/W}"),
				Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Scourge of the Nobilis — +1/+0 until end of turn"),
			}},
			Text: text,
		}},
		Static: []game.StaticAbility{
			effects.B43PumpAttachedWhen(red, 1, 1),
			effects.GrantAbilities(red, grant),
			effects.B43PumpAttachedWhen(white, 1, 1),
			effects.B43GrantToAttachedWhen(white, "lifelink"),
		},
	})
}
