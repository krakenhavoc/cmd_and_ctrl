package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Smelting Vat — Artifact {4} (The Brothers' War Commander, S82):
//
//	"{1}, {T}, Sacrifice another artifact: Reveal the top eight cards of your
//	 library. Put up to two noncreature artifact cards with total mana value
//	 less than or equal to the sacrificed artifact's mana value from among
//	 them onto the battlefield and the rest on the bottom of your library in
//	 a random order."
//
// The sacrificed artifact's mana value is read from its last-known record on
// the cost (CR 608.2h, Context.SacrificedPermanent); a token that is no copy
// is zero. All eight cards are revealed, so everyone sees them. The ceiling
// is a registered set rule picked for that number (s82BroTotalValueRule).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "71d2a0fe-5426-4fa3-85a4-564c37184996",
		Name:         "Smelting Vat",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}, Sacrifice another artifact: Reveal the top eight cards of your library. Put up to two noncreature artifact cards with total mana value less than or equal to the sacrificed artifact's mana value from among them onto the battlefield and the rest on the bottom of your library in a random order.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}"), TapCost(), SacrificeAnotherN(1, "another artifact", Artifact())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				player := item.Controller
				limit := 0
				if info, ok := ctx.SacrificedPermanent(); ok {
					limit = info.ManaValue
				}
				revealed := g.RevealTopOfLibraryForEffect(player, ctx.Source(), 8, "Smelting Vat — reveal the top eight cards of your library")
				return PutFromLibraryOntoBattlefield{
					Player:   player,
					Cards:    revealed,
					Match:    And(Artifact(), Noncreature()),
					Max:      2,
					Optional: true,
					Rule:     s82BroTotalValueRule(g, revealed, limit, 2),
					Label:    "Smelting Vat — put up to two noncreature artifact cards with total mana value no greater than the sacrificed artifact's onto the battlefield",
					ThenTail: PutRestOnBottomInRandomOrderTail,
				}.Apply(ctx)
			},
		}},
	})
}
