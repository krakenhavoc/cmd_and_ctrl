package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prismari, the Inspiration — Legendary Creature — Elder Dragon {5}{U}{R},
// 7/7 (Secrets of Strixhaven, #4456):
//
//	"Flying
//	 Ward—Pay 5 life.
//	 Instant and sorcery spells you cast have storm. (Whenever you cast an
//	 instant or sorcery spell, copy it for each spell cast before it this turn.
//	 You may choose new targets for the copies.)"
//
// Flying is a printed keyword and the ward is ward.go's pay-life ward (CR
// 702.21a). The last line is the other shape of storm (Maelstrom Nexus's shape
// for cascade): a permanent that GIVES the keyword, so the trigger lives on the
// battlefield and watches the cast of an instant or sorcery spell its
// controller casts. Each such cast copies the spell once for each spell cast
// before it this turn (CR 702.40a), every player's spells counting, and the
// copies are the caster's and may have new targets (CR 707.10c). A spell that
// also prints storm triggers a second time (CR 702.40b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "258f5045-3aa7-4b95-a15f-3172577294a8",
		Name:            "Prismari, the Inspiration",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Ward(WardLife(5), "Prismari, the Inspiration — ward, pay 5 life"),
			socB32GrantsStorm("Prismari, the Inspiration", func(spell game.Card, _ *game.Card, _ *game.Game) bool {
				return spell.IsInstant() || spell.IsSorcery()
			}),
		},
	})
}
