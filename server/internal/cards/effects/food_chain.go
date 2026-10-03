package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Food Chain — Enchantment {2}{G} (#1600):
//
//	"Exile a creature you control: Add X mana of any one color, where X
//	 is 1 plus the exiled creature's mana value. Spend this mana only to
//	 cast creature spells."
//
// A MANA ability (CR 605.1a: it adds mana and does not target) whose
// cost is the exile-a-permanent component (game.ExilePermanentsCost,
// ADR 0020's 2026-10-03 amendment): the creature is named at
// activation, exiled through the ordinary zone change, and the mana is
// in the pool before any trigger the exile caused reaches the stack
// (CR 605.3a). It is not a sacrifice, so dies and sacrifice triggers
// never see it.
//
// X is read off the paid-cost record: the cost puts the exiled
// permanent on paid.Exiled, and its mana value is the one it had as it
// last existed on the battlefield (game.PermanentInfo.ManaValue) — so
// a token that copies nothing is 0 and X is 1, a face-down creature is
// 0, and a Clone copying a six-drop is 6. "Any one color" is one pick
// minting X tokens (OneColorOfAmount), and the tokens carry "spend
// this mana only to cast creature spells" — the cast purpose AND the
// type, because a type tag alone also admits a creature's activated
// ability (#2059).
//
// A commander named to the cost is asked about the command zone before
// anything is paid; the mana is added either way, because the cost was
// paid.
//
// The auto-tapper never uses it: which creature to exile is a choice.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5c8e5092-962e-49ef-ab82-8434e475e4e7",
		Name:         "Food Chain",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:            ManaAbilityCost{ExilePermanents: ExileACreatureYouControl().ExilePermanents},
			ProducedForPaid: foodChainProduced,
			Restrictions:    []string{ManaRestrictCast, ManaRestrictType("Creature")},
			Label:           "Exile a creature you control: Add X mana of any one color, where X is 1 plus the exiled creature's mana value. Spend this mana only to cast creature spells",
		}},
	})
}

// foodChainProduced is "X mana of any one color, where X is 1 plus the
// exiled creature's mana value", read from what the cost exiled.
//
// With nothing exiled — CR 106.7's "could produce" readers and the
// up-front colour list ask before any payment — the answer is one mana
// of any one colour: what Food Chain could make is the five colours,
// and how many is the payment's business.
func foodChainProduced(g *game.Game, _, _ uuid.UUID, paid game.PaidCost) string {
	if len(paid.Exiled) == 0 {
		return OneColorOfAmount(1)
	}
	mv := 0
	if info, ok := g.LastKnownPermanentForEffect(paid.Exiled[0]); ok {
		mv = info.ManaValue
	}
	return OneColorOfAmount(1 + mv)
}
