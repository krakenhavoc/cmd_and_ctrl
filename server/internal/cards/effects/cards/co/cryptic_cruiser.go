package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cryptic Cruiser — Creature — Eldrazi Processor {3}{U}, 3/3 (Battle for
// Zendikar, #5547):
//
//	"Devoid (This card has no color.)
//	 {2}{U}, Put a card an opponent owns from exile into that player's
//	 graveyard: Tap target creature."
//
// The cost is Void Attendant's processing (the activator names the card; with
// none the ability can't be activated, CR 118.3). The target is chosen as the
// ability goes on the stack and re-checked as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ae7cfdb8-b0f2-46d7-b847-9a760ddbe028",
		Name:            "Cryptic Cruiser",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}, Put a card an opponent owns from exile into that player's graveyard: Tap target creature.",
			Purpose: game.Purpose{Answers: game.AnswerRestrict},
			Cost:    effects.Plus(effects.ManaCost("{2}{U}"), effects.ProcessCards(1, "a card an opponent owns from exile", nil)),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				t, ok := effects.FirstLegalTarget(ctx)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return effects.TapTarget{Target: t.ID}.Apply(ctx)
			},
		}},
	})
}
