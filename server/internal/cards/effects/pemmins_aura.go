package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "ada2f380-bca7-4d64-9587-bab11c776ba8",
		Name:         "Pemmin's Aura",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Activated: []ActivatedAbility{
			{
				Label:   "{U}: Untap enchanted creature.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    ManaCost("{U}"),
				Effect:  untapAttachedHost,
			},
			{
				Label:   "{U}: Enchanted creature gains flying until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    ManaCost("{U}"),
				Effect:  scgB1HostGainsUntilEOT("flying", "Pemmin's Aura — flying until end of turn"),
			},
			{
				Label:   "{U}: Enchanted creature gains shroud until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerProtect},
				Cost:    ManaCost("{U}"),
				Effect:  scgB1HostGainsUntilEOT("shroud", "Pemmin's Aura — shroud until end of turn"),
			},
			{
				Label:   "{1}: Enchanted creature gets +1/-1 or -1/+1 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    ManaCost("{1}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return MayChoice{
						Question: "Pemmin's Aura — +1/-1 or -1/+1 until end of turn?",
						YesLabel: "+1/-1",
						NoLabel:  "-1/+1",
						Then:     pemminsAuraPump,
					}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

// pemminsAuraPump is Pemmin's Aura's choice: +1/-1 on a yes, -1/+1 on a
// no, to the creature the Aura enchants as the answer arrives. Keyed
// (#4291).
var pemminsAuraPump = MayYesNo("confirm/pemmins-aura-pump",
	func(ctx *Context, _ game.ConfirmAnswer) error {
		return scgB1HostGetsUntilEOT(ctx, 1, -1, "Pemmin's Aura — +1/-1 until end of turn")
	},
	func(ctx *Context, _ game.ConfirmAnswer) error {
		return scgB1HostGetsUntilEOT(ctx, -1, 1, "Pemmin's Aura — -1/+1 until end of turn")
	})
