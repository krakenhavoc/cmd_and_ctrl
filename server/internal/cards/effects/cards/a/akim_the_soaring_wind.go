package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Akim, the Soaring Wind — {2}{U}{R}{W} Legendary Creature — Bird Dinosaur 3/4
// (Commander 2020):
//
//	"Flying
//	 Whenever you create one or more tokens for the first time each turn,
//	 create a 1/1 white Bird creature token with flying.
//	 {3}{U}{R}{W}: Creature tokens you control gain double strike until end of
//	 turn."
//
// The trigger reads the controller's per-turn tally of tokens created
// (EventTokenCreated, bumped before the harvest, one per token): it is 1
// exactly when this is the first token of the turn, so tokens made earlier in
// the turn, before Akim was there, do not count as "the first time", and the
// rest of a batch and the Bird itself do not trigger it again. The grant is
// Ravenous Robots' (creature tokens you control, until end of turn).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2a20530b-d1d6-4dac-8d08-f583bea55e6e",
		Name:            "Akim, the Soaring Wind",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventTokenCreated, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return source != nil && ev.Kind == game.EventTokenCreated && ev.Actor == source.Controller &&
					g.TurnTallyFor(source.Controller).TokensCreated == 1
			}, "Akim, the Soaring Wind — create a 1/1 white Bird creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Bird with flying"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{U}{R}{W}: Creature tokens you control gain double strike until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{3}{U}{R}{W}"),
			Effect: effects.Do(effects.GrantKeywordUntilEOT{
				Match:    effects.And(effects.Creature(), effects.IsTokenPredicate(), effects.YouControl()),
				Keywords: []string{"double strike"},
				Label:    "Akim, the Soaring Wind — double strike",
			}),
		}},
	})
}
