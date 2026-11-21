package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aberrant Researcher // Perfected Form — {3}{U} Creature — Human Insect
// 3/2 // Creature — Insect Horror 5/4 (Innistrad Remastered, S82 GAP,
// #5948):
//
//	Front: "Flying
//	        At the beginning of your upkeep, mill a card. If an instant or
//	        sorcery card was milled this way, transform this creature."
//	Back:  "Flying"
//
// Ainok Wayfarer's mill with a tail: the tail reads the card that actually
// reached the graveyard (CR 400.7; a card a replacement sent elsewhere was
// not milled this way) and transforms the Researcher in place if it was an
// instant or sorcery. The back face is a plain 5/4 flier.
//
// No simplification.
func init() {
	const oracle = "722ee05f-6815-41b7-9813-5ec3902c1e9f"
	Register(Spec{
		OracleID:        oracle,
		Name:            "Aberrant Researcher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Aberrant Researcher — mill a card; if an instant or sorcery card was milled this way, transform this creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return MillToZone{N: 1, ThenTail: ctx.ItemTail(aberrantResearcherMilled)}.Apply(ctx)
				}),
		},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Perfected Form",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}

// aberrantResearcherMilled is "if an instant or sorcery card was milled
// this way, transform this creature", run once the card has landed. An
// on-disk key: never renamed, never reused.
var aberrantResearcherMilled = CardTail("aberrant-researcher/after-mill", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && (c.HasCardType("instant") || c.HasCardType("sorcery")) {
			return TransformThis{}.Apply(ctx)
		}
	}
	return nil
})
