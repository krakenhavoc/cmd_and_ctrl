package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mageta the Lion — Legendary Creature — Human Spellshaper {3}{W}{W}, 3/3 (Prophecy,
// S82 OLD b672, #5617):
//
//	"{2}{W}{W}, {T}, Discard two cards: Destroy all creatures except for Mageta.
//	 Those creatures can't be regenerated."
//
// Latulla's tapped, two-discard Spellshaper cost with Damnation's sweep that
// ignores regeneration shields (CR 701.19c). "Except for Mageta" is the ability's
// source, and a Mageta that has left leaves nothing to spare.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cd170b1c-879b-414e-9f0e-5e7fb099e7bb",
		Name:         "Mageta the Lion",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}{W}, {T}, Discard two cards: Destroy all creatures except for Mageta. Those creatures can't be regenerated.",
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
			Cost:    effects.Plus(effects.ManaCost("{2}{W}{W}"), effects.TapCost(), effects.DiscardN(2, "two cards")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				mageta := item.SourceCardID
				return effects.DestroyAllMatching{
					Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
						return c.IsCreature() && c.InstanceID != mageta
					},
					CantBeRegenerated: true,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
