package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Frost Augur — Snow Creature — Human Wizard {U}, 1/2:
//
//	"{S}, {T}: Look at the top card of your library. If it's a snow card, you
//	 may reveal it and put it into your hand. ({S} can be paid with one mana
//	 from a snow source.)"
//
// The look is private; a snow card (the snow supertype on the card in the
// library, CR 205.4a) may be revealed and taken, and a card that isn't snow, or
// that is declined, stays on top. {S} is paid only by mana a snow source made
// (CR 107.4h, #5986).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "85aedc9f-df35-49bc-88c8-23efda9a4b1b",
		Name:         "Frost Augur",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{S}, {T}: Look at the top card of your library. If it's a snow card, you may reveal it and put it into your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tutors: 1},
			Cost:    effects.Plus(effects.ManaCost("{S}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				player := item.Controller
				return effects.TakeFromLibraryToHand{
					Player: player,
					Cards:  g.LookAtTopOfLibraryForEffect(player, 1),
					Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
						return c.HasSupertype("snow")
					},
					Max:      1,
					Optional: true,
					Label:    "Frost Augur — you may reveal the snow card and put it into your hand",
					Key:      effects.OldB071FrostAugurTake,
				}.Apply(ctx)
			},
		}},
	})
}
