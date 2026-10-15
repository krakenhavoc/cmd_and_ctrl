package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scaled Nurturer — Creature — Dragon Druid {1}{G}, 0/2 (EDHREC rank
// 3669):
//
//	"{T}: Add {G}. When you spend this mana to cast a Dragon creature
//	 spell, you gain 2 life."
//
// The Dragon deck's mana dork, and a Dragon itself for the lords and
// the Thrasher-style payoffs. The mana half is a plain tapped {G},
// with summoning sickness applying as it does to every creature's
// tap ability (CR 302.6).
//
// The life gain is a spend rider (#1547): a triggered ability that
// fires when the token pays for a Dragon creature spell and goes on the
// stack above it. The {G} has no restriction, so it still pays for
// anything; only the life gain is Dragon-only, and the auto-tapper
// plans the Nurturer like any Forest — a Dragon it pays for still gains
// the 2 life.
//
// No simplification is left (#3844, ADR 0068 amendment 2026-10-15). A
// waived payment (strict mana off with too little in the pool, or Cast
// anyway) still spends the share of the cost the pool can pay, so this
// mana pays whenever it can and its rider fires. Mana that never
// entered the pool was not this card's mana.
func init() {
	Register(Spec{
		OracleID:       "13b96709-0e88-476b-9485-956e682bb818",
		Name:           "Scaled Nurturer",
		ReadsManaSpent: game.ReadsPoolManaOnly,
		Completeness:   CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
			SpendRiders: []game.ManaSpendRider{WhenManaSpent("Scaled Nurturer", game.ManaSpendTrigger{
				Label:  "Scaled Nurturer — gain 2 life",
				Effect: scaledNurturerGainLife,
			}, ManaRestrictCast, ManaRestrictType("Creature"), ManaRestrictSubtype("Dragon"))},
		}},
	})
}

// scaledNurturerGainLife: "you gain 2 life" — you being the player who
// spent the mana.
func scaledNurturerGainLife(g *game.Game, item *game.StackItem) error {
	return GainLife{Player: item.Controller, Amount: 2}.Apply(NewContext(g, item))
}
