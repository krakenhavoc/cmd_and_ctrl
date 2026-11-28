package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ashen Powder — Sorcery {2}{B}{B} (Mirage, S82 OLD b799, #5633):
//
//	"Put target creature card from an opponent's graveyard onto the
//	 battlefield under your control."
//
// Agadeem Occultist's reanimation without its condition: the card must be a
// creature card in a graveyard another player owns, and it enters under the
// caster's control (CR 110.2a), still its owner's card. A card that left the
// graveyard by resolution makes the spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d906965a-0f33-4d75-8abe-66f2cf4817b8",
		Name:         "Ashen Powder",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target creature card from an opponent's graveyard", effects.Creature(), effects.Not(effects.YouOwn())),
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				return effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}.Apply(ctx)
			}
			return nil
		},
	})
}
