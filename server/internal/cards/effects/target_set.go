package effects

import (
	"strconv"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// target_set.go — #1559: the catalog's vocabulary for a rule over the
// chosen SET of a clause's targets (CR 601.2c), attached with
// TargetSpec.EachDifferent:
//
//	TargetCardInGraveyard("any number of target creature cards that each have a different mana value X or less",
//	    Creature(), YouOwn()).WithCount(0, 0).EachDifferent(EachDifferentManaValue()).WithManaValueAtMostX()
//	TargetCreature("two target creatures controlled by different players").
//	    WithCount(2, 2).EachDifferent(EachDifferentController())
//
// Each rule is a KEY no two picks may share, because the enumerator
// and the client's picker have to evaluate it as well as the engine
// (game/target_set.go says why a predicate over the set would not
// do). Add a rule here, never in a card file: a key a card writes by
// hand is one nobody else knows the label of.
//
// The Label completes "those targets must …", which is how the
// announce gate's refusal and the picker's banner both say it.

// EachDifferentManaValue is "that each have a different mana value"
// (Agadeem's Awakening, Long Rest). A card whose cost the engine
// cannot read has no key and collides with nothing.
func EachDifferentManaValue() *game.TargetDifference {
	return &game.TargetDifference{
		Label: "each have a different mana value",
		Key: func(_ *game.Game, c game.Card, _ game.ZoneKind) (string, bool) {
			mv, ok := c.ParsedManaValue()
			if !ok {
				return "", false
			}
			return strconv.Itoa(mv), true
		},
	}
}

// EachDifferentController is "controlled by different players" (Run Away
// Together, Chaos Mutation) — and, for the same reason, the "one per
// opponent" of Windgrace's Judgment, read as "no two of those
// permanents share a controller". It reads the controller as the
// card has it NOW, so the resolution re-check judges who controls
// each creature as the spell resolves.
func EachDifferentController() *game.TargetDifference {
	return &game.TargetDifference{
		Label: "be controlled by different players",
		Key: func(_ *game.Game, c game.Card, _ game.ZoneKind) (string, bool) {
			return c.Controller.String(), true
		},
	}
}

// EachDifferentName is "with different names" (Behold the Sinister Six!).
// The name is the card's current one (Effective), so a Clone copying
// something is judged by the name it copied.
func EachDifferentName() *game.TargetDifference {
	return &game.TargetDifference{
		Label: "each have a different name",
		Key: func(_ *game.Game, c game.Card, _ game.ZoneKind) (string, bool) {
			// Normalised as gifts_ungiven.go's DifferentNames does,
			// which is every other name comparison in the catalog.
			name := strings.ToLower(strings.TrimSpace(c.Effective().Name))
			if name == "" {
				return "", false
			}
			return name, true
		},
	}
}
