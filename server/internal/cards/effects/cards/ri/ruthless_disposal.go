package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ruthless Disposal — Sorcery {4}{B} (Jumpstart 2022, #5948):
//
//	"As an additional cost to cast this spell, discard a card and
//	 sacrifice a creature.
//	 Two target creatures each get -13/-13 until end of turn."
//
// One additional cost with both parts, paid while the spell is on the
// stack: the discarded card and the sacrificed creature are named at
// announce. The two targets are chosen as the spell is cast, each shrunk
// by 13 as it resolves; a target that is gone is skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dcd12cb6-6920-4bff-af73-104ad32f7d59",
		Name:         "Ruthless Disposal",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		AdditionalCost: &game.AdditionalCost{
			DiscardCards: 1,
			Sacrifice:    effects.SacrificeSpec("a creature", effects.Creature()),
			Label:        "Discard a card and sacrifice a creature",
		},
		Targets: effects.TargetCreature("two target creatures").WithCount(2, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: -13, Toughness: -13, Label: "Ruthless Disposal — -13/-13"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
