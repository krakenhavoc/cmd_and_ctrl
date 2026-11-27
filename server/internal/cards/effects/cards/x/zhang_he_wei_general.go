package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zhang He, Wei General — Legendary Creature — Human Soldier {3}{B}{B}, 4/2
// (Portal Three Kingdoms, #5622):
//
//	"Horsemanship (This creature can't be blocked except by creatures with horsemanship.)"
//	"Whenever Zhang He attacks, each other creature you control gets +1/+0 until end of turn."
//
// Horsemanship is an engine keyword carried as printed data: the block
// check refuses any blocker without horsemanship. It doesn't stop this
// creature blocking.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fe35b809-7439-4bdc-bce2-42262095064f",
		Name:            "Zhang He, Wei General",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"horsemanship"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Zhang He — each other creature you control gets +1/+0 until end of turn", func(g *game.Game, item *game.StackItem) error {
				return effects.BoostUntilEOT{
					Match:     effects.And(effects.Creature(), effects.YouControl(), effects.OtherThan(item.SourceCardID)),
					Power:     1,
					Toughness: 0,
					Label:     "Zhang He — each other creature you control gets +1/+0",
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
