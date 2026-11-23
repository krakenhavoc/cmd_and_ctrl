package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vhal, Candlekeep Researcher — Legendary Creature — Human Wizard {3}{U},
// 2/3 (Commander Legends: Battle for Baldur's Gate, #5948):
//
//	"Vigilance
//	 {T}: Add an amount of {C} equal to Vhal's toughness. This mana can't be
//	 spent to cast spells from your hand.
//	 Choose a Background (You can have a Background as a second commander.)"
//
// Bighorner Rancher's count-driven mana ability (the amount is read as the
// ability resolves, from the layered toughness, so a +1/+1 counter or a
// toughness change counts) carrying Karolina Dean's restriction: the mana can't
// be spent to cast spells from the hand (it can pay for abilities, and for
// spells cast from other zones). Choose a Background is a deck-construction
// rule (CR 702.124k) that internal/deck reads off the oracle text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7253e7d1-b1df-4637-bf64-850be7fa5188",
		Name:            "Vhal, Candlekeep Researcher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		ManaAbilities: []effects.ManaAbility{{
			Cost:           effects.ManaAbilityCost{Tap: true},
			Label:          "{T}: Add an amount of {C} equal to Vhal's toughness. This mana can't be spent to cast spells from your hand.",
			ProducedAmount: vhalColorless,
			Restrictions:   []string{effects.ManaRestrictNotFromHand},
		}},
	})
}

func vhalColorless(g *game.Game, _, source uuid.UUID) game.ManaAmount {
	g.RecomputeLayersIfStaleLocked()
	n := 0
	if c, ok := g.LookupCardForEffect(source); ok {
		n = c.CurrentToughness()
	}
	return game.ManaAmount{Options: []string{"C"}, N: n}
}
