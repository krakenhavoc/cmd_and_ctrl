package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Temple of Khonshu — Land:
//
//	"This land enters tapped unless you control a Dimension.
//	 {T}: Add {W} or {B}."
//
// The entry clause is the checkland replacement (SelfEntersTappedUnless)
// with a different question: does the entering player control a
// permanent that is a Dimension, by card type or by subtype, read from
// its effective characteristics as the land enters (CR 614.1c, 305.7).
// No printed card is a Dimension today, so the land enters tapped; the
// day one is printed, or an effect gives a permanent the word, this
// card answers correctly without a change.
//
// The mana ability adds {W} or {B} (one choice of colour, as a dual land does).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "34916c4b-fc1b-4864-bb3f-968977fd533d",
		Name:         "Temple of Khonshu",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{youControlADimensionEntry()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|B}",
			Label:    "Add {W} or {B}",
		}},
	})
}

// youControlADimensionEntry is "enters tapped unless you control a Dimension".
func youControlADimensionEntry() game.ReplacementEffect {
	ab := SelfEntersTappedUnless(func(g *game.Game, controller uuid.UUID) bool {
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == controller && (c.HasCardType("dimension") || c.HasSubtype("Dimension")) {
				return true
			}
		}
		return false
	})
	ab.RowLabel = "Enters tapped unless you control a Dimension."
	return ab
}
