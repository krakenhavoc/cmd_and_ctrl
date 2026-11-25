package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Legion's Landing // Adanto, the First Fort — Legendary Enchantment {W}
// (Ixalan, S82 older sets):
//
//	"When Legion's Landing enters, create a 1/1 white Vampire creature token
//	 with lifelink.
//	 When you attack with three or more creatures, transform Legion's
//	 Landing."
//
//	Adanto, the First Fort — Legendary Land:
//	"{T}: Add {W}.
//	 {2}{W}, {T}: Create a 1/1 white Vampire creature token with lifelink."
//
// The attack trigger fires once for the declaration, when at least three of
// the attackers were declared by you (Armasaur Guide's batch count).
//
// No simplification.
func init() {
	const oracle = "f7d8b91b-6541-4d3e-af51-7e000eac69c1"
	vampire := effects.CreateToken{Template: effects.TokenCard("1/1 white Vampire with lifelink"), N: 1}
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Legion's Landing",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Legion's Landing — create a 1/1 white Vampire creature token with lifelink", effects.Do(vampire)),
			effects.WithAtBatchEnd(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source != nil && effects.AttackDeclaredByYou(ev, source.Controller)
			}, "Legion's Landing — transform it (you attacked with three or more creatures)", effects.Do(effects.TransformThis{})),
				effects.ThatManyEventsAtLeast(3)),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Adanto, the First Fort",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "{T}: Add {W}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}, {T}: Create a 1/1 white Vampire creature token with lifelink.",
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{W}"), effects.TapCost()),
			Effect:  effects.Do(vampire),
		}},
	})
}
