package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lotleth Troll — Creature — Zombie Troll {B}{G}, 2/1 (Double Masters 2022):
//
//	"Trample
//	 Discard a creature card: Put a +1/+1 counter on this creature.
//	 {B}: Regenerate this creature."
//
// The discard is a cost the activator pays from their hand at announce
// (CR 602.2b), so it has no mana and can be done at instant speed as often as
// they have creature cards. The regeneration is a shield (CR 701.19a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "61b1d7e5-6155-4204-b110-35a890551ec8",
		Name:            "Lotleth Troll",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "Discard a creature card: Put a +1/+1 counter on this creature.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.DiscardCardsMatching(1, "a creature card", func(c game.Card) bool { return c.IsCreature() }),
				Effect:  effects.PlusOneCountersOnThis(1),
			},
			{
				Label:   "{B}: Regenerate this creature.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    effects.ManaCost("{B}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.Regenerate{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
