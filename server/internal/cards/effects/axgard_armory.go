package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Axgard Armory — Land (EDHREC rank 3428):
//
//	"This land enters tapped.
//	 {T}: Add {W}.
//	 {1}{R}{R}{W}, {T}, Sacrifice this land: Search your library for
//	 an Aura card and/or an Equipment card, reveal them, put them
//	 into your hand, then shuffle."
//
// The Voltron deck's utility land. The tapped entry is the CR 614
// self-replacement; the white mana is a plain tap ability; the
// tutor is a CR 602 ability with a mana-tap-sacrifice cost, paid at
// announce. "An Aura card and/or an Equipment card" is two searches
// in printed order — up to one Aura, then up to one Equipment — each
// a real prompt the searcher may decline, so two Auras cannot be
// taken (b32SearchAuraThenEquipmentToHand). The library shuffles
// once, after the second; both finds are revealed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bce30fd0-ed1e-495d-9149-6a4c81c45c7b",
		Name:         "Axgard Armory",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
		Activated: []ActivatedAbility{{
			Label:  "{1}{R}{R}{W}, {T}, Sacrifice this land: Search your library for an Aura card and/or an Equipment card, reveal them, put them into your hand, then shuffle.",
			Cost:   Plus(ManaCost("{1}{R}{R}{W}"), TapCost(), SacrificeThis()),
			Effect: b32SearchAuraThenEquipmentToHand,
		}},
	})
}
