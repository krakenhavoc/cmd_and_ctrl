package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elvish Eulogist — Creature — Elf Shaman {G}, 1/1 (Lorwyn):
//
//	"Sacrifice this creature: You gain 1 life for each Elf card in your
//	 graveyard."
//
// The sacrifice is the whole cost (CR 602.2b), so the Eulogist is already in
// the graveyard when the ability resolves and counts itself (it is an Elf
// card). The Elf cards are counted as the ability resolves (CR 608.2h), by
// the subtype a card has in the graveyard (a changeling card counts).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "21017dc0-9b79-48b8-8eae-561abfad31d5",
		Name:         "Elvish Eulogist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: You gain 1 life for each Elf card in your graveyard.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.SacrificeThis(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				p := ctx.PlayerByID(item.Controller)
				if p == nil || p.Graveyard == nil {
					return nil
				}
				n := 0
				for _, c := range p.Graveyard.Cards {
					if c.HasSubtypeWord("Elf") {
						n++
					}
				}
				return effects.GainLife{Amount: n}.Apply(ctx)
			},
		}},
	})
}
