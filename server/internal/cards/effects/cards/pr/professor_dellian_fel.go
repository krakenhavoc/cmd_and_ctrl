package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Professor Dellian Fel — Legendary Planeswalker — Dellian {2}{B}{G},
// loyalty 5 (Secrets of Strixhaven Commander, #4834):
//
//	"+2: You gain 3 life.
//	 0: You draw a card and lose 1 life.
//	 −3: Destroy target creature.
//	 −6: You get an emblem with "Whenever you gain life, target opponent
//	 loses that much life.""
//
// The 0 draws, then loses the life (CR 608.2c): the loss waits for any
// prompt the draw asked. The −3 targets any creature and does nothing if
// the target is gone or illegal by resolution (CR 608.2b).
//
// The −6 makes an emblem whose one ability is Sanguine Bond's trigger
// (WheneverYouGainLifeTargetOpponentLosesThatMuch): it watches every
// positive life change of the emblem's owner (CR 114.2), carries the
// amount gained as the trigger is put on the stack, and asks for the
// target opponent then. The emblem's ability works from the command zone
// (CR 114.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8c5fe985-69e1-4a03-99fc-a6fd2314a9a7",
		Name:         "Professor Dellian Fel",
		Completeness: effects.CompletenessFull,
		Emblem: &effects.EmblemSpec{
			Label: "Professor Dellian Fel emblem",
			Text:  "Whenever you gain life, target opponent loses that much life.",
			Triggered: []game.TriggeredAbility{
				effects.WheneverYouGainLifeTargetOpponentLosesThatMuch("Professor Dellian Fel emblem — target opponent loses that much life"),
			},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:  "+2: You gain 3 life.",
				Cost:   effects.LoyaltyCost(2),
				Effect: effects.Do(effects.GainLife{Amount: 3}),
			},
			{
				Label: "0: You draw a card and lose 1 life.",
				Cost:  effects.LoyaltyCost(0),
				Effect: effects.Do(
					effects.DrawCards{N: 1},
					effects.GainLife{Amount: -1},
				),
			},
			{
				Label:   "−3: Destroy target creature.",
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.DestroyEachLegalTarget(item, effects.NewContext(g, item))
				},
			},
			{
				Label: "−6: You get an emblem with \"Whenever you gain life, target opponent loses that much life.\"",
				Cost:  effects.LoyaltyCost(-6),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateEmblem{}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
