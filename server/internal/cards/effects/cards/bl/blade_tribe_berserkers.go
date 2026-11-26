package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blade-Tribe Berserkers — Creature — Human Berserker {3}{R}, 3/3 (Scars of
// Mirrodin, S82 OLD batch 469, #5577):
//
//	"Metalcraft — When this creature enters, if you control three or more
//	 artifacts, this creature gets +3/+3 and gains haste until end of turn."
//
// The enter trigger on the metalcraft gate (checked again as it resolves,
// CR 603.4); the pump and the haste are two until-end-of-turn effects on the
// Berserkers, which do nothing if it has left or come back as a new object
// (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "655847c6-7c5e-4659-870e-825d0cf5e90c",
		Name:         "Blade-Tribe Berserkers",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OldB469WhenThisEntersIfMetalcraft("Blade-Tribe Berserkers — +3/+3 and haste until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.SourceIsStillThisPermanent(g, item) {
						return nil
					}
					ctx := effects.NewContext(g, item)
					if err := (effects.BoostUntilEOT{Target: item.SourceCardID, Power: 3, Toughness: 3,
						Label: "Blade-Tribe Berserkers — +3/+3 until end of turn"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"haste"},
						Label: "Blade-Tribe Berserkers — haste until end of turn"}.Apply(ctx)
				}),
		},
	})
}
