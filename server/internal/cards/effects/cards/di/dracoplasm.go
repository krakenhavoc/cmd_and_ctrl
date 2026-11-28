package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dracoplasm — Creature — Shapeshifter {U}{R}, 0/0 (Tempest, S82 OLD batch 768,
// #5628):
//
//	"Flying
//	 As this creature enters, sacrifice any number of creatures. This creature's
//	 power becomes the total power of those creatures and its toughness becomes
//	 their total toughness.
//	 {R}: This creature gets +1/+0 until end of turn."
//
// The sacrifice is Devour's entry choice (any number of creatures, zero allowed,
// made as the creature enters and before anything can respond), and its Then
// reads the sacrificed creatures' last-known power and toughness (CR 608.2h) and
// asks the landing for them as a base size that never ends
// (ReplacementEvent.EntersWithBasePT, added with this card): a layer 7b set
// effect pinned to the permanent before it is seen on the battlefield, which a
// later +1/+0 or counter lands on top of. The pump is the card's own
// activated ability. Sacrificing nothing makes a 0/0 that dies as a state-based
// action (CR 704.5f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3155c833-8170-4166-8f47-c3c1fad1575a",
		Name:            "Dracoplasm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements:    []game.ReplacementEffect{effects.DracoplasmEntersSacrificing("Dracoplasm")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{R}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Dracoplasm — +1/+0 until end of turn"),
		}},
	})
}
