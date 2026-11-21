package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dreadlight Monstrosity — Creature — Crab Horror {4}{U}{U}, 5/5
// (Innistrad: Crimson Vow):
//
//	"Ward {2}
//	 {3}{U}{U}: This creature can't be blocked this turn. Activate only if
//	 you own a card in exile."
//
// The activation condition is read as the ability is activated (CR 602.5b):
// any card its controller owns in exile, face up or down, a token never.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "068b9586-53e3-4791-a26c-a6b38d038f66",
		Name:         "Dreadlight Monstrosity",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.Ward(effects.WardMana("{2}"), "Dreadlight Monstrosity — ward {2}")},
		Activated: []effects.ActivatedAbility{{
			Label:     "{3}{U}{U}: This creature can't be blocked this turn. Activate only if you own a card in exile.",
			Purpose:   game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:      effects.ManaCost("{3}{U}{U}"),
			Condition: youOwnACardInExile,
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked,
					Label: "Dreadlight Monstrosity — can't be blocked this turn"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

// youOwnACardInExile is "Activate only if you own a card in exile".
func youOwnACardInExile(g *game.Game, controller, _ uuid.UUID) bool {
	for _, c := range g.Exile.Cards {
		if c.Owner == controller && !c.IsToken() {
			return true
		}
	}
	return false
}
