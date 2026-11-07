package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eshki Dragonclaw — Legendary Creature — Human Warrior {1}{G}{U}{R}, 4/4
// (Tarkir: Dragonstorm, #4463):
//
//	"Vigilance, trample, ward {1}
//	 At the beginning of combat on your turn, if you've cast both a
//	 creature spell and a noncreature spell this turn, draw a card and
//	 put two +1/+1 counters on Eshki Dragonclaw."
//
// Ward is the shared trigger. The combat trigger is Alicia Masters,
// Skilled Sculptor's intervening "if" over the turn's cast tally
// (CR 603.4): every spell is a creature spell or a noncreature spell, so
// the creature spells are the total less the noncreature ones. A
// creature spell with another type (an artifact creature) is a creature
// spell and not a noncreature one. Eshki that has left in response still
// draws the card and gets no counters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "85d87087-48ee-4dc8-a1c0-9c9ccb885adc",
		Name:            "Eshki Dragonclaw",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "trample"},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{1}"), "Eshki Dragonclaw — ward {1}"),
			InterveningIf(CondBoard{
				Name: "you've cast both a creature spell and a noncreature spell this turn",
				Test: func(ctx *Context) bool { return eshkiCastBoth(ctx.Game, ctx.Controller()) },
			}, On(game.EventStepBegan, AllOf(StepBegan(game.StepBeginCombat, true),
				func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return eshkiCastBoth(g, source.Controller)
				}), "Eshki Dragonclaw — draw a card and put two +1/+1 counters on it",
				Do(DrawCards{N: 1}, CounterOnThis{Kind: game.CounterPlusOne, N: 2}))),
		},
	})
}

// eshkiCastBoth is "you've cast both a creature spell and a noncreature
// spell this turn".
func eshkiCastBoth(g *game.Game, player uuid.UUID) bool {
	tally := g.CastTallyFor(player)
	return tally.Noncreature > 0 && tally.Total-tally.Noncreature > 0
}
