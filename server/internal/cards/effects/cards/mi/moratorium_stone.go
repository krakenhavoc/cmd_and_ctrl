package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moratorium Stone — Artifact {1} (Guildpact, S82 OLD batch 573, #5599):
//
//	"{2}, {T}: Exile target card from a graveyard.
//	 {2}{W}{B}, {T}, Sacrifice this artifact: Exile target nonland card from
//	 a graveyard, all other cards from graveyards with the same name as that
//	 card, and all permanents with that name."
//
// The first ability is Beckon Apparition's exile of one graveyard card. The
// second names the target's card names as it lies in its graveyard
// (CR 201.2), then exiles it, every other graveyard card sharing a name with
// it and every permanent sharing a name with it, all together
// (ExileTargetCardAndEveryCardAndPermanentNamedLikeIt). The Stone is
// sacrificed as a cost, so the ability resolves with the Stone gone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bb0d89e7-4ae6-4ac2-a0e6-51ab08226069",
		Name:         "Moratorium Stone",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}, {T}: Exile target card from a graveyard.",
				Purpose: game.Purpose{Answers: game.AnswerRemove},
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
				Targets: effects.TargetCardInGraveyard("target card from a graveyard"),
				Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					return effects.ExileTargetCards(ctx.Game, ctx.Item)
				})),
			},
			{
				Label:   "{2}{W}{B}, {T}, Sacrifice this artifact: Exile target nonland card from a graveyard, all other cards from graveyards with the same name as that card, and all permanents with that name.",
				Purpose: game.Purpose{Answers: game.AnswerRemove},
				Cost:    effects.Plus(effects.ManaCost("{2}{W}{B}"), effects.TapCost(), effects.SacrificeThis()),
				Targets: effects.TargetCardInGraveyard("target nonland card from a graveyard", effects.Nonland()),
				Effect:  effects.Do(effects.Instruction(effects.ExileTargetCardAndEveryCardAndPermanentNamedLikeIt)),
			},
		},
	})
}
