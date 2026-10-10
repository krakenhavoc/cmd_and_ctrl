package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nardole, Resourceful Cyborg — Legendary Artifact Creature —
// Scientist {1}{U}, 1/2:
//
//	"Undying
//	 {T}: Add {U} for each counter on Nardole. Spend this mana only
//	 to cast noncreature spells.
//	 Doctor's companion"
//
// Undying is PrintedKeywords (#2075). The amount counts counters of
// EVERY kind on Nardole (the +1/+1 undying leaves, a charge counter
// from elsewhere), read when the ability is activated; with none it
// adds nothing. The spend clause is ManaRestrictCast +
// ManaRestrictNotType("Creature") (#2136): it refuses a creature spell
// AND any activated ability, since "only to cast" is a clause about a
// spell.
//
// Doctor's companion is a deck-construction rule ("you can have two
// commanders if the other is the Doctor") and needs nothing on the
// card, as Partner needs nothing on Thrasios: internal/deck reads it
// off the oracle text and pairs Nardole with a lone Time Lord Doctor
// (CR 702.124m, #2874).
func init() {
	Register(Spec{
		OracleID:        "81c515b7-9174-4b34-ad07-4144e3dbaaac",
		Name:            "Nardole, Resourceful Cyborg",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordUndying},
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			ProducedFunc: nardoleProduced,
			Restrictions: []string{ManaRestrictCast, ManaRestrictNotType("Creature")},
			Label:        "Add {U} for each counter on Nardole. Spend this mana only to cast noncreature spells",
		}},
	})
}

// nardoleProduced is one {U} per counter of any kind on the source.
func nardoleProduced(g *game.Game, _, source uuid.UUID) string {
	c, ok := g.LookupCardForEffect(source)
	if !ok {
		return ""
	}
	n := 0
	for _, v := range c.Counters {
		n += v
	}
	return strings.Repeat("{U}", n)
}
