package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nightbird's Clutches — {1}{R} Sorcery (Innistrad, S82 OLD b448, #5572):
//
//	"Up to two target creatures can't block this turn.
//	 Flashback {3}{R} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// Each target still legal at resolution (CR 608.2b) is barred from blocking until
// end of turn (Sparkmage's Gambit's second half); with no target the spell does
// nothing. Flashback is the shared graveyard cast.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "d12b875c-0056-4a57-a7ff-09c73d00f8cc",
		Name:             "Nightbird's Clutches",
		Completeness:     effects.CompletenessFull,
		Purpose:          effects.ForTargets(effects.CantBlockUntilEOT(0, 0, 0)),
		Targets:          effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{3}{R}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if err := (effects.RestrictUntilEOT{Target: t.ID, Restrictions: game.CantBlock, Label: "Nightbird's Clutches — can't block this turn"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
