package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Inspiration from Beyond — Sorcery {2}{U} (Foundations, #4465):
//
//	"Mill three cards, then return an instant or sorcery card from your
//	 graveyard to your hand.
//	 Flashback {5}{U}{U} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// Grapple with the Past's shape: the mill, then a choice made as the spell
// resolves from the graveyard as it is then, so a card just milled is a
// legal pick. Nothing is targeted. The return is not optional, so the
// prompt asks for exactly one card when there is one (Min 1). The spell is
// on the stack while it resolves, so it never picks itself, cast from the
// hand or flashed back. Flashback is the shared alternative cost (CR
// 702.34).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "be87e410-4302-43ee-b880-56eedf2b3648",
		Name:             "Inspiration from Beyond",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{5}{U}{U}")},
		OnResolve: ResolveSteps(
			// The pick waits for a prompt the mill asks (CR 608.2c, #4155).
			MillCards{N: 3},
			Instruction(func(ctx *Context) error {
				return ReturnChosenFromGraveyard{
					Player:   ctx.Controller(),
					Question: "Inspiration from Beyond — return an instant or sorcery card from your graveyard to your hand",
					Match:    func(c game.Card) bool { return c.IsInstant() || c.IsSorcery() },
					Min:      1,
					Max:      1,
					ToHand:   true,
				}.Apply(ctx)
			}),
		),
	})
}
