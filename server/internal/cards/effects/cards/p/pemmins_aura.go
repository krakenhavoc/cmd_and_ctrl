package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pemmin's Aura — Enchantment — Aura {1}{U}{U} (Scourge):
//
//	"Enchant creature
//	 {U}: Untap enchanted creature.
//	 {U}: Enchanted creature gains flying until end of turn.
//	 {U}: Enchanted creature gains shroud until end of turn.
//	 {1}: Enchanted creature gets +1/-1 or -1/+1 until end of turn."
//
// Four abilities of the Aura itself, each finding its host as it resolves
// (Crab Umbra's untapAttachedHost and its siblings), so an Aura that has
// fallen off by then does nothing. The last is Endling's choice made as
// the ability resolves: +1/-1 on a yes, -1/+1 on a no.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ada2f380-bca7-4d64-9587-bab11c776ba8",
		Name:         "Pemmin's Aura",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{U}: Untap enchanted creature.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    effects.ManaCost("{U}"),
				Effect:  effects.UntapAttachedHost,
			},
			{
				Label:   "{U}: Enchanted creature gains flying until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    effects.ManaCost("{U}"),
				Effect:  effects.ScgB1HostGainsUntilEOT("flying", "Pemmin's Aura — flying until end of turn"),
			},
			{
				Label:   "{U}: Enchanted creature gains shroud until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerProtect},
				Cost:    effects.ManaCost("{U}"),
				Effect:  effects.ScgB1HostGainsUntilEOT("shroud", "Pemmin's Aura — shroud until end of turn"),
			},
			{
				Label:   "{1}: Enchanted creature gets +1/-1 or -1/+1 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{1}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.MayChoice{
						Question: "Pemmin's Aura — +1/-1 or -1/+1 until end of turn?",
						YesLabel: "+1/-1",
						NoLabel:  "-1/+1",
						Then:     pemminsAuraPump,
					}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}

// pemminsAuraPump is Pemmin's Aura's choice: +1/-1 on a yes, -1/+1 on a
// no, to the creature the Aura enchants as the answer arrives. Keyed
// (#4291).
var pemminsAuraPump = effects.MayYesNo("confirm/pemmins-aura-pump",
	func(ctx *effects.Context, _ game.ConfirmAnswer) error {
		return effects.ScgB1HostGetsUntilEOT(ctx, 1, -1, "Pemmin's Aura — +1/-1 until end of turn")
	},
	func(ctx *effects.Context, _ game.ConfirmAnswer) error {
		return effects.ScgB1HostGetsUntilEOT(ctx, -1, 1, "Pemmin's Aura — -1/+1 until end of turn")
	})
