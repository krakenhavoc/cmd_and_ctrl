package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Invasion of Xerex // Vertex Paladin — March of the Machine (#5477).
//
// Front, Battle — Siege 4, {2}{W}{U}: "When this Siege enters, return up to
// one target creature to its owner's hand." Back, Vertex Paladin, Creature —
// Angel Knight */*: "Flying. Vertex Paladin's power and toughness are each
// equal to the number of creatures you control."
//
// No simplification.
const InvasionOfXerexOracleID = "8b5add91-2e82-4430-bfc5-f4a6366dbbdb"

func init() {
	effects.Register(effects.Spec{
		OracleID:     InvasionOfXerexOracleID,
		Name:         "Invasion of Xerex",
		Completeness: effects.CompletenessFull,
		Battle:       &effects.BattleSpec{Defense: 4, Subtype: effects.BattleSubtypeSiege},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Invasion of Xerex — return up to one target creature to its owner's hand",
				effects.BounceUpToOneTarget), effects.TargetCreature("up to one target creature").WithCount(0, 1)),
			effects.DefeatedTrigger("Invasion of Xerex — defeated: exile it, then cast Vertex Paladin", effects.SiegeDefeated()),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        InvasionOfXerexOracleID + "#1",
		Name:            "Vertex Paladin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			effects.S82MomB17CountPT("Vertex Paladin's power and toughness are each equal to the number of creatures you control.",
				func(g *game.Game, source *game.Card) int { return effects.B04CreaturesControlled(g, source.Controller) }),
		},
	})
}
