package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blades of Velis Vel — Kindred Instant — Shapeshifter {1}{R} (Lorwyn, S82
// OLD b548, #5593):
//
//	"Changeling (This card is every creature type.)
//	 Up to two target creatures each get +2/+0 and gain all creature types
//	 until end of turn."
//
// "Up to two" lets it be cast with one target or none. Each creature still
// legal as the spell resolves gets the pump and the type grant, locked then
// (CR 611.2c); an illegal one is skipped, and with every target illegal the
// spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9955885b-fad7-4ee6-bd62-48d8a7731044",
		Name:            "Blades of Velis Vel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling},
		Purpose:         effects.ForTargets(game.TargetPurpose{Slot: 0, UntilEndOfTurn: game.TargetBenefit}),
		Targets:         effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: 2, Label: "Blades of Velis Vel — +2/+0"}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.GrantAllCreatureTypesUntilEOT{Target: t.ID, Label: "Blades of Velis Vel — all creature types"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
