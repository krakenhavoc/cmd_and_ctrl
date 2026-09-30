package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zoetic Glyph — Enchantment — Aura {2}{U}:
//
//	"Enchant artifact
//	 Enchanted artifact is a Golem creature with base power and
//	 toughness 5/4 in addition to its other types.
//	 When this Aura is put into a graveyard from the battlefield,
//	 discover 3."
//
// "In addition to its other types" is an ADD in layer 4 — the artifact
// keeps being an artifact, and keeps its subtypes, so an Equipment made
// a creature this way is still an Equipment (and falls off, CR 301.5c,
// which the state-based actions handle) — and the 5/4 is a layer 7b
// base. The graveyard trigger is Rancor's: EventLTB fires only from the
// battlefield, so "landed in a graveyard" is the whole condition.
// Discover is ADR 0099's (game/discover.go).
func init() {
	Register(Spec{
		OracleID:     "1c21efcf-1007-45bb-ba21-ac33c2a8d751",
		Name:         "Zoetic Glyph",
		Completeness: CompletenessFull,
		Discovers:    true,
		Targets:      TargetPermanent("enchant artifact", Artifact()),
		Static: []game.StaticAbility{
			{
				Layer:     game.Layer4Type,
				AppliesTo: AttachedToSource,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					if !zoeticHas(c.Types, "Creature") {
						c.Types = append(c.Types, "Creature")
					}
					if !zoeticHas(c.Subtypes, "Golem") {
						c.Subtypes = append(c.Subtypes, "Golem")
					}
				},
			},
			SetAttachedBasePT(5, 4),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source)
			}, "Zoetic Glyph — discover 3", DiscoverN(3)),
		},
	})
}

func zoeticHas(list []string, want string) bool {
	for _, s := range list {
		if strings.EqualFold(s, want) {
			return true
		}
	}
	return false
}
