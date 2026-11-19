package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Joraga Treespeaker — Creature — Elf Druid {G}, 1/1 (Rise of the Eldrazi,
// #5581):
//
//	"Level up {1}{G} ({1}{G}: Put a level counter on this. Level up only as
//	 a sorcery.)
//	 LEVEL 1-4
//	 1/2
//	 {T}: Add {G}{G}.
//	 LEVEL 5+
//	 1/4
//	 Elves you control have "{T}: Add {G}{G}.""
//
// A leveler (leveler.go, CR 702.87a, 711.2a, 711.2b). The band's mana
// ability is an ADR 0093 bundle the Treespeaker grants itself behind the
// band's level check (Honored Hierarch's gated self-grant), so with no level
// counter it has no mana ability (CR 711.5). From level 5 the same bundle
// goes to every Elf its controller controls, itself included; each Elf taps
// for it as its own ability, so a summoning-sick Elf can't (CR 302.6).
//
// No simplification.
const joragaTreespeakerGrant = "joraga-treespeaker/gg"

func init() {
	Register(Spec{
		OracleID:     "a011aefe-5853-4435-b686-179e931ba897",
		Name:         "Joraga Treespeaker",
		Completeness: CompletenessFull,
		Grants:       []AbilityGrant{TapForManaGrant(joragaTreespeakerGrant, "{G}{G}", "Add {G}{G}", "{T}: Add {G}{G}.")},
		Activated:    []ActivatedAbility{LevelUpKeyword("{1}{G}")},
		Static: LevelBands(
			LevelBand{From: 1, To: 4, Power: 1, Toughness: 2, Statics: []game.StaticAbility{
				Labelled(GrantAbilities(selfOnly, joragaTreespeakerGrant), `This creature has "{T}: Add {G}{G}."`),
			}},
			LevelBand{From: 5, Power: 1, Toughness: 4, Statics: []game.StaticAbility{
				Labelled(TribalAbilityGrant(TribeFilter{Tribes: []string{"Elf"}, YoursOnly: true}, joragaTreespeakerGrant),
					`Elves you control have "{T}: Add {G}{G}."`),
			}},
		),
	})
}
