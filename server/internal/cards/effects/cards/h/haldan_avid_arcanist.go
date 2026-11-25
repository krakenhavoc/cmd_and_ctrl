package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Haldan, Avid Arcanist — Legendary Creature — Human Wizard {2}{U}, 1/4
// (Commander 2020, S82 OLD, #5513):
//
//	"Partner with Pako, Arcane Retriever
//	 You may play lands and cast noncreature spells from among cards you
//	 exiled that have fetch counters on them, and you may spend mana as though
//	 it were mana of any color to cast those spells."
//
// A standing permission (ADR 0066) read off the battlefield on every query, so
// it ends when Haldan leaves: Evelyn, the Covetous's shape without the
// once-a-turn budget. "With a fetch counter" is WithCounter, "cards you
// exiled" is ExiledByHolderAbility (the exile stamps the controller of the
// resolving ability, so a card an opponent's Pako exiled does not qualify),
// "lands and noncreature spells" is NoncreatureOnly over a permission that
// covers plays (CR 305.1), and "mana of any color" is AnyColor (CR 609.4b).
// "Partner with" is the CR 702.124j entry search.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6a0b78fb-3838-47d9-8d1a-b6b0c0e8927d",
		Name:         "Haldan, Avid Arcanist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.PartnerWith("Haldan, Avid Arcanist", "Pako, Arcane Retriever"),
		},
		CastPermissions: []game.CastPermission{{
			Zone: game.ZoneExile,
			Filter: game.PermissionFilter{
				WithCounter:           effects.FetchCounter,
				ExiledByHolderAbility: true,
				NoncreatureOnly:       true,
			},
			AnyColor: true,
			Label:    "Play a land or cast a noncreature spell from among cards you exiled with fetch counters (Haldan, Avid Arcanist)",
		}},
	})
}
