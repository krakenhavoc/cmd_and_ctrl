package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Trenzalore Clocktower — Legendary Land (Doctor Who):
//
//	"{T}: Add {U}. Put a time counter on Trenzalore Clocktower.
//	 {1}{U}, {T}, Remove twelve time counters from Trenzalore Clocktower and
//	 exile it: Shuffle your graveyard and hand into your library, then draw
//	 seven cards. Activate only if you control a Time Lord."
//
// The counter is a rider of the mana ability (Foreboding Statue's), settled
// as the mana is made. The second ability's cost removes twelve counters and
// exiles the land; its effect tucks the hand and graveyard (Midnight Clock's
// payoff) and, in the tail once the tuck is done, shuffles and draws seven.
// "Activate only if you control a Time Lord" is an activation condition read
// off the battlefield (Time Lord is the two type words Time and Lord).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "69143645-97b6-4c7c-9fa2-844fb3b99822",
		Name:         "Trenzalore Clocktower",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 7},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Rider: func(g *game.Game, _, source uuid.UUID) error {
				_, err := g.AddCounterMustSettleNowForEffect(source, game.CounterTime, 1)
				return err
			},
			Label: "{T}: Add {U}. Put a time counter on Trenzalore Clocktower.",
		}},
		Activated: []ActivatedAbility{{
			Label:     "{1}{U}, {T}, Remove twelve time counters from Trenzalore Clocktower and exile it: Shuffle your graveyard and hand into your library, then draw seven cards. Activate only if you control a Time Lord.",
			Purpose:   game.Purpose{Draws: 7, Answers: game.AnswerValue},
			Cost:      Plus(ManaCost("{1}{U}"), TapCost(), RemoveCountersFromThis(game.CounterTime, 12), ExileThis()),
			Condition: gapB95YouControlATimeLord,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ids := append(allHandCardIDs(g, item.Controller), echoOfEonsGraveyardCardIDs(g, item.Controller)...)
				return g.TuckCardsToLibraryKeyedForEffect(ids, game.TuckOptions{}, NewContext(g, item).ItemTail(trenzaloreClocktowerTucked))
			},
		}},
	})
}

// trenzaloreClocktowerTucked is the rest of the ability once the hand and
// graveyard are in the library: shuffle, then draw seven. An on-disk key.
var trenzaloreClocktowerTucked = CardTail("then/trenzalore-clocktower-shuffle-draw", func(ctx *Context, _ game.TailOutcome) error {
	if err := ctx.Game.ShuffleLibraryForEffect(ctx.Controller()); err != nil {
		return err
	}
	return ctx.Game.DrawNForEffect(ctx.Controller(), 7)
})
