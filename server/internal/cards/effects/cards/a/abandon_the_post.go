package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Abandon the Post — Sorcery {1}{R} (Midnight Hunt, #5497):
//
//	"Up to two target creatures can't block this turn.
//	 Flashback {3}{R} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// Beat a Path's body (up to two creatures, CantBlock until end of turn)
// at sorcery speed with the shared Flashback.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "5404c6a5-399d-4237-b03c-41b6a1b79050",
		Name:             "Abandon the Post",
		Completeness:     effects.CompletenessFull,
		Purpose:          effects.ForTargets(effects.CantBlockUntilEOT(0, 0, 0)),
		Targets:          effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{3}{R}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.RestrictUntilEOT{Target: t.ID, Restrictions: game.CantBlock, Label: "Abandon the Post — can't block"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
