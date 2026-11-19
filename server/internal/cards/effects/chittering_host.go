package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chittering Host — Creature — Eldrazi Horror, 5/6. The combined back face
// of Graf Rats and Midnight Scavengers (CR 712.5b); never a card in a
// deck, only the melded permanent the two become (ADR 0145, #2699), here
// for Innistrad Remastered (S82 GAP, #5948):
//
//	"Haste
//	 Menace
//	 When this creature enters, other creatures you control get +1/+0 and
//	 gain menace until end of turn."
//
// Haste and menace are the back face's printed keywords, carried by the
// melded permanent from the deck import. The enters trigger is Junkyard
// Genius's rally: the creatures are the ones you control as it resolves
// (CR 611.2c), never the Host itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6923cf6b-7d3e-4d95-abaf-1df1a04ac7c1",
		Name:            "Chittering Host",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste", "menace"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Chittering Host — other creatures you control get +1/+0 and gain menace until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					self := ctx.Source()
					match := And(Creature(), YouControl(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.InstanceID != self })
					if err := (BoostUntilEOT{Match: match, Power: 1, Label: "Chittering Host — +1/+0"}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{Match: match, Keywords: []string{"menace"}, Label: "Chittering Host — menace"}.Apply(ctx)
				}),
		},
	})
}
