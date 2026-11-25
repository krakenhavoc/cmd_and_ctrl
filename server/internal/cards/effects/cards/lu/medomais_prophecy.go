package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Medomai's Prophecy — Enchantment — Saga {1}{U} (Theros Beyond Death, #5514):
//
//	"I — Scry 2.
//	 II — Choose a card name.
//	 III — When you cast a spell with the chosen name for the first time this turn,
//	 draw two cards.
//	 IV — Look at the top card of each player's library."
//
// Chapter II queues the engine's card-name prompt for the Saga (CR 201.2 free
// text, stored as the permanent's ChosenName). Chapter III is a delayed triggered
// ability (CR 603.7) created as it resolves, from the name the Saga holds then:
// it fires on the first spell with that name the controller casts this turn
// (CastTheNamedSpellForTheFirstTimeCondition) and draws two. Chapter IV shows
// the top card of each library to the controller only, as a private look.
//
// No simplification.
func init() {
	const name = "Medomai's Prophecy"
	effects.Register(effects.Spec{
		OracleID:     "e2ea1901-68ac-4afc-adc2-ac6bc7465311",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 2},
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "scry 2"), effects.Do(effects.Scry{N: 2})),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "choose a card name"),
				func(g *game.Game, item *game.StackItem) error {
					g.QueueCardNameChoiceForEffect(item.Controller, item.SourceCardID, "Medomai's Prophecy — choose a card name")
					return nil
				}),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "when you cast a spell with the chosen name for the first time this turn, draw two cards"),
				func(g *game.Game, item *game.StackItem) error {
					chosen := g.ChosenNameOf(item.SourceCardID)
					if chosen == "" {
						return nil
					}
					return effects.DelayedOnEvent{
						Label:      "Medomai's Prophecy — draw two cards",
						On:         []game.EventKind{game.EventCast},
						Condition:  effects.CastTheNamedSpellForTheFirstTimeCondition,
						CondParams: game.EffectParams{Name: chosen},
						Body:       effects.DrawTwoBody,
					}.Apply(effects.NewContext(g, item))
				}),
			effects.ChapterTrigger(4, effects.SagaChapterLabel(name, 4, "look at the top card of each player's library"), medomaisProphecyLook),
		},
	})
}

// medomaisProphecyLook is chapter IV: the controller looks at the top card of
// each player's library, their own included (CR 701.20; a look is not a reveal,
// so only the controller comes to know each card).
func medomaisProphecyLook(g *game.Game, item *game.StackItem) error {
	for _, seat := range g.Seats {
		if seat == nil || seat.Eliminated {
			continue
		}
		g.LookAtTopOfPlayersLibraryForEffect(item.Controller, seat.ID, 1)
	}
	return nil
}
