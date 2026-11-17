package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tribute to Urborg — Instant {1}{B} (Dominaria United):
//
//	"Kicker {1}{U} (You may pay an additional {1}{U} as you cast this
//	 spell.)
//	 Target creature gets -2/-2 until end of turn. If this spell was
//	 kicked, that creature gets an additional -1/-1 until end of turn
//	 for each instant and sorcery card in your graveyard."
//
// The instants and sorceries are counted once as the spell resolves; it
// is on the stack then, so it does not count itself. The two shrinks are
// one effect of the same duration, applied as a single -(2+N)/-(2+N).
func init() {
	Register(Spec{
		OracleID:      "8355d44d-cf03-4b19-b2dc-0ca3450dbf5c",
		Name:          "Tribute to Urborg",
		Purpose:       ForTargets(HarmUntilEOT(0, 0, 0)),
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Kicker("{1}{U}")},
		Targets:       TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			n := 2
			if ctx.WasKicked() {
				n += instantAndSorceryCardsInGraveyard(ctx.Game, ctx.Controller())
			}
			return BoostUntilEOT{Target: id, Power: -n, Toughness: -n, Label: "Tribute to Urborg — gets -N/-N until end of turn"}.Apply(ctx)
		},
	})
}
