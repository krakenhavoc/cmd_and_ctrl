package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jet's Brainwashing — Sorcery {R}:
//
//	"Kicker {3} (You may pay an additional {3} as you cast this spell.)
//	 Target creature can't block this turn. If this spell was kicked,
//	 gain control of that creature until end of turn, untap it, and it
//	 gains haste until end of turn.
//	 Create a Clue token. (It's an artifact with "{2}, Sacrifice this
//	 token: Draw a card.")"
//
// Kicker is CR 702.33a; a copy of a kicked spell is kicked too, which
// ctx.WasKicked reads. The threaten is Flayer of Loyalties' order:
// control, untap, then the keyword. Creating the Clue is not
// investigating (the ruling), so it is a plain token creation. With its
// only target illegal the spell does nothing at all, the Clue included
// (CR 608.2b).
func init() {
	Register(Spec{
		OracleID:      "5cca85e0-43e8-4b97-a43b-1b8f23f541d1",
		Name:          "Jet's Brainwashing",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Kicker("{3}")},
		Targets:       TargetCreature("target creature"),
		Purpose:       ForTargets(HarmUntilEOT(0, 0, 0)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if id, ok := b16FirstLegalTargetCard(ctx); ok {
				if err := (RestrictUntilEOT{Target: id, Restrictions: game.CantBlock, Label: "Jet's Brainwashing — can't block this turn"}).Apply(ctx); err != nil {
					return err
				}
				if ctx.WasKicked() {
					if err := threatenAndGrant(ctx, id, "Jet's Brainwashing", "haste"); err != nil {
						return err
					}
				}
			}
			return CreateToken{Template: ClueToken(), N: 1}.Apply(ctx)
		},
	})
}
