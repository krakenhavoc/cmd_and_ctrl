package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Valeria Richards, Precocious — {3}{U} Legendary Creature — Human
// Scientist Hero 3/3 (Marvel Super Heroes Commander, #4455):
//
//	"Noncreature spells you cast cost {1} less to cast.
//	 Whenever you cast your first noncreature spell each turn, draw a
//	 card."
//
// The reduction only trims generic mana (CR 601.2f). The draw trigger
// counts the turn's casts, so a noncreature spell cast before Valeria
// entered is the turn's first and no later one draws.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "89931cf0-f20b-49b5-88c6-2a92aee8c50c",
		Name:         "Valeria Richards, Precocious",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Noncreature spells you cast cost {1} less to cast.", effects.YourSpell(), effects.NoncreatureSpell()),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCastYourFirstNoncreatureSpellThisTurn,
				"Valeria Richards — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
