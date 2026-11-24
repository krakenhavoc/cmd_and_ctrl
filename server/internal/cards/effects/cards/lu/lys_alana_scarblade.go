package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lys Alana Scarblade — Creature — Elf Assassin {2}{B}, 1/1 (Kaldheim
// Commander, #5506):
//
//	"{T}, Discard an Elf card: Target creature gets -X/-X until end of turn,
//	 where X is the number of Elves you control."
//
// The discard is a cost paid as the ability is activated (an Elf card is
// read in the hand); the target is chosen with it. X is counted as the
// ability resolves, through the layered view, the Scarblade itself counted
// if it is still an Elf. A target that has gone is not affected.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b7ebd207-43e3-49eb-86ca-b5987ef9e22f",
		Name:         "Lys Alana Scarblade",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Discard an Elf card: Target creature gets -X/-X until end of turn, where X is the number of Elves you control.",
			Purpose: game.Purpose{Answers: game.AnswerRemove, Targets: game.ForTargets(effects.HarmUntilEOT(0, 0, 0))},
			Cost: effects.Plus(effects.TapCost(), effects.DiscardCardsMatching(1, "an Elf card", func(c game.Card) bool {
				return c.HasSubtypeWord("Elf")
			})),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				src, ok := g.LookupCardForEffect(item.SourceCardID)
				if !ok {
					return nil
				}
				x := effects.S82OldB074ElvesYouControl(g, item.Controller, &src)
				for _, t := range ctx.LegalTargets() {
					return effects.BoostUntilEOT{Target: t.ID, Power: -x, Toughness: -x, Label: "Lys Alana Scarblade — -X/-X"}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
