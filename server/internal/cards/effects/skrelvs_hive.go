package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skrelv's Hive — Enchantment {1}{W}:
//
//	"At the beginning of your upkeep, you lose 1 life and create a 1/1
//	 colorless Phyrexian Mite artifact creature token with toxic 1 and
//	 'This token can't block.'
//	 Corrupted — As long as an opponent has three or more poison
//	 counters, creatures you control with toxic have lifelink."
//
// The corrupted clause is a layer 6 grant to your creatures that have
// toxic, re-answered on every recompute, so it switches on and off as
// poison counters come and go. Toxic is read through game.ToxicTotal,
// so a creature that was granted toxic counts as having it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "06d219ff-0083-4c2a-b5b3-2b84bb58f57e",
		Name:         "Skrelv's Hive",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Skrelv's Hive — lose 1 life and create a Phyrexian Mite",
				func(g *game.Game, item *game.StackItem) error {
					if err := g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -1); err != nil {
						return err
					}
					return CreateToken{Controller: item.Controller, Template: PhyrexianMiteToken(), N: 1}.Apply(NewContext(g, item))
				}),
		},
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller &&
					game.ToxicTotal(target) > 0 && rfReprintBCorrupted(g, source.Controller)
			}, "lifelink"),
		},
	})
}
