package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Peri Brown — Legendary Creature — Human {3}{W}, 2/3 (Doctor Who, #5948):
//
//	"The first historic spell you cast each turn has convoke. (Your creatures
//	 can help cast it. Each creature you tap while casting it pays for {1} or
//	 one mana of that creature's color.)
//	 Doctor's companion"
//
// A granted tap cost (SpellsYouCastHaveConvoke) over a historic spell that is
// the first one its caster has cast this turn (gapB91FirstHistoricSpellThisTurn).
// A spell with convoke of its own gains nothing more (CR 702.51d). Doctor's
// companion is a deck-construction rule.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2fadf4aa-1b7e-4e72-ad20-809a59ddccf8",
		Name:         "Peri Brown",
		Completeness: effects.CompletenessFull,
		GrantedTapCosts: []game.GrantedTapCost{effects.SpellsYouCastHaveConvoke(
			"The first historic spell you cast each turn has convoke.",
			func(g *game.Game, caster uuid.UUID, spell game.Card, _ game.ZoneKind) bool {
				return effects.GapB91FirstHistoricSpellThisTurn(g, caster, spell)
			})},
	})
}
