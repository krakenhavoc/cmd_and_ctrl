package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Font of Magic — Enchantment {3}{U} (Commander Legends: Battle for Baldur's
// Gate, #5948):
//
//	"Instant and sorcery spells you cast cost {1} less to cast for each time
//	 you've cast a commander from the command zone this game."
//
// Henzie "Toolbox" Torre's reduction over the commander-cast tally
// (SocB3CommanderCastsFromCommandZone), read as each cost is totalled (CR
// 601.2f) and applied to generic mana (CR 118.7a); the spell is yours and an
// instant or a sorcery.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "755230c3-0b89-4123-bed4-2f7729daa859",
		Name:         "Font of Magic",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLessEach(func(q game.CostQuery) int {
				return effects.SocB3CommanderCastsFromCommandZone(q.Game, q.Source.Controller)
			}, "Instant and sorcery spells you cast cost {1} less to cast for each time you've cast a commander from the command zone this game.",
				effects.YourSpell(), effects.InstantOrSorcerySpell()),
		},
	})
}
