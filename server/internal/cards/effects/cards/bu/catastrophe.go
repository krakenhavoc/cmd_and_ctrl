package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Catastrophe — Sorcery {4}{W}{W} (Urza's Saga, S82 OLD batch 731, #5624):
//
//	"Destroy all lands or all creatures. Creatures destroyed this way can't be
//	 regenerated."
//
// Not a modal spell: its controller chooses between the two as it resolves
// (CR 608.2d), then every permanent of that kind is destroyed as one event
// (CR 700.4). Regeneration shields are ignored (CR 701.19c); that is printed
// for creatures, and a land that is a creature is covered, while the one case it
// would reach beyond the printed text, a noncreature land holding a
// regeneration shield, is one no card in the game can produce.
//
// No simplification.
func init() {
	catastropheChosen = effects.OptionPickThen("option-pick/catastrophe-lands-or-creatures", func(ctx *effects.Context, r game.OptionPicked) error {
		var match effects.CardPredicate
		switch r.Index {
		case 0:
			match = effects.Land()
		case 1:
			match = effects.Creature()
		default:
			return nil
		}
		return effects.DestroyAllMatching{Match: match, CantBeRegenerated: true}.Apply(ctx)
	})
	effects.Register(effects.Spec{
		OracleID:     "ed68160e-5e3d-44f2-ac7a-3749fda10119",
		Name:         "Catastrophe",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.PickOption{
				Question: "Catastrophe — destroy all lands or all creatures?",
				Options: []game.ChoiceOption{
					{Label: "Destroy all lands"},
					{Label: "Destroy all creatures"},
				},
				ThenKey: catastropheChosen,
				Carry:   []uuid.UUID(nil),
			}.Apply(ctx)
		},
	})
}

// catastropheChosen is the controller's answer, keyed. Set in init, with the
// card. An on-disk key.
var catastropheChosen game.OptionPickThen
