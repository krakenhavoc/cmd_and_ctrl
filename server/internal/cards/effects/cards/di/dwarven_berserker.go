package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dwarven Berserker — Creature — Dwarf Berserker {1}{R}, 1/1 (Weatherlight, S82 older sets batch 778, #5629):
//
//	"Whenever this creature becomes blocked, it gets +3/+0 and gains trample until end of turn."
//
// Fires once as the creature becomes blocked (CR 509.3d), however many creatures block it. The pump and
// the keyword are granted to the creature as it is now, if it is still the same object (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cc43507a-ca61-45ae-ac0a-4ddb851c27ef",
		Name:         "Dwarven Berserker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked, effects.ThisBecomesBlocked,
				"Dwarven Berserker — it gets +3/+0 and gains trample until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id := item.SourceCardID
					if !effects.SourceIsStillThisPermanent(g, item) {
						return nil
					}
					if err := (effects.BoostUntilEOT{Target: id, Power: 3, Label: "Dwarven Berserker — +3/+0"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "Dwarven Berserker — trample"}.Apply(ctx)
				}),
		},
	})
}
