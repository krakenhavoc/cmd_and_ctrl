package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Quandrix, the Proof — Legendary Creature — Elder Dragon {4}{G}{U}, 6/6
// (Secrets of Strixhaven, #4456):
//
//	"Flying, trample
//	 Cascade
//	 Instant and sorcery spells you cast from your hand have cascade."
//
// Cascade is the keyword on the card that has it (Cascade(), CR 702.85);
// the granted half is GrantsCascade, which watches every cast of its
// controller and queues the same cascade trigger for an instant or sorcery
// spell. "From your hand" is the spell's own cast zone (CastFromOwnersHand)
// and its owner being the caster, so a flashback cast from the graveyard, a
// spell cast off the top of a library or from exile, or a commander cast
// from the command zone, never cascades.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2cbccc46-bdef-4dfb-90a4-0278c5c8488a",
		Name:            "Quandrix, the Proof",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			Cascade(),
			GrantsCascade("Quandrix, the Proof", func(spell game.Card, source *game.Card, g *game.Game) bool {
				return (spell.IsInstant() || spell.IsSorcery()) &&
					spell.Owner == source.Controller &&
					CastFromOwnersHand()(g, source.Controller, spell)
			}),
		},
	})
}
