package cardssy

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Talara's Battalion — Creature — Elf Warrior {1}{G} 4/3:
//
//	"Cast this spell only if you've cast another green spell this turn.
//	 Trample"
//
// Illusory Angel's cast condition over the turn's colours: the caster's tally
// of the colours of the spells they cast this turn is read as the cast is
// announced, before this spell is counted, so any green spell in it is another
// one (CR 601.2). A spell cast and then countered still counts.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "08c7a6f5-7e25-4f8d-9e4a-aec7248e45fa",
		Name:            "Talara's Battalion",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		CastCondition: func(g *game.Game, controller uuid.UUID, _ game.Card) bool {
			return slices.Contains(g.CastTallyFor(controller).SpellColorsCast(), "G")
		},
		CastConditionLabel: "Cast this spell only if you've cast another green spell this turn.",
	})
}
