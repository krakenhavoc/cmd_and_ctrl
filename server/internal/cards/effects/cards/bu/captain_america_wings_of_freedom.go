package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Captain America, Wings of Freedom — Legendary Creature — Human Soldier
// Hero {2}{W}, 3/1 (Marvel Super Heroes, #4455):
//
//	"Flying, first strike, ward {1}
//	 Whenever Captain America attacks, each other Hero you control gets
//	 +X/+X until end of turn, where X is Captain America's toughness."
//
// X is read as the ability resolves: Captain America's toughness now, or
// his last known toughness if he has left (CR 608.2h). The set of Heroes is
// fixed at resolution (CR 611.2c), so a Hero that enters afterwards gets
// nothing, and he does not pump himself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "942ac2ac-7528-4768-8398-8520a8ee15e5",
		Name:            "Captain America, Wings of Freedom",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike"},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{1}"), "Captain America, Wings of Freedom — ward {1}"),
			effects.WheneverThisAttacks("Captain America, Wings of Freedom — each other Hero you control gets +X/+X",
				func(g *game.Game, item *game.StackItem) error {
					x := effects.ToughnessNowOrLastKnown(g, item.SourceCardID)
					return effects.BoostUntilEOT{
						Match:     effects.And(effects.Creature(), effects.YouControl(), effects.Subtype("Hero"), effects.NotSelf(item.SourceCardID)),
						Power:     x,
						Toughness: x,
						Label:     "Captain America, Wings of Freedom",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
