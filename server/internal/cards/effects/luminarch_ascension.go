package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Luminarch Ascension — Enchantment {1}{W} (EDHREC rank 3113):
//
//	"At the beginning of each opponent's end step, if you didn't lose
//	 life this turn, you may put a quest counter on this enchantment.
//	 (Damage causes loss of life.)
//	 {1}{W}: Create a 4/4 white Angel creature token with flying.
//	 Activate only if this enchantment has four or more quest counters
//	 on it."
//
// The quest ticks up on each opponent's end step you took no damage
// or life loss in — an intervening-if (CR 603.4), checked when the end
// step begins and again as the trigger resolves, off the turn tally's
// LifeLost (#586). "You may" is the trigger's optional prompt. The
// Angel ability's "four or more quest counters" is its activation
// condition (CR 602.1b, #743), SourceHasCountersAtLeast — an instant-
// speed ability with no tap, so it can be activated as many times as
// the mana allows.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "90076bf5-aa9a-4a6e-9035-9aa97fd5561e",
		Name:         "Luminarch Ascension",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventBeginEndStep, luminarchAscensionOpponentsEndStepUnscathed,
				"Luminarch Ascension — put a quest counter",
				func(g *game.Game, item *game.StackItem) error {
					if !luminarchAscensionUnscathed(g, item.Controller) {
						return nil
					}
					return AddCounter{Target: item.SourceCardID, Kind: "quest", N: 1}.Apply(NewContext(g, item))
				}), "Luminarch Ascension — put a quest counter on it?"),
		},
		Activated: []ActivatedAbility{{
			Label:     "{1}{W}: Create a 4/4 white Angel creature token with flying. Activate only if this enchantment has four or more quest counters on it.",
			Cost:      ManaCost("{1}{W}"),
			Condition: SourceHasCountersAtLeast("quest", 4),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateToken{Controller: item.Controller, Template: TokenCard("4/4 white Angel with flying"), N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}

// luminarchAscensionOpponentsEndStepUnscathed is the trigger
// condition: an opponent's end step began and you have lost no life
// this turn.
func luminarchAscensionOpponentsEndStepUnscathed(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.Actor != uuid.Nil && ev.Actor != source.Controller && luminarchAscensionUnscathed(g, source.Controller)
}

// luminarchAscensionUnscathed is "you didn't lose life this turn".
func luminarchAscensionUnscathed(g *game.Game, you uuid.UUID) bool {
	return b18LifeLostThisTurn(g, you) == 0
}
