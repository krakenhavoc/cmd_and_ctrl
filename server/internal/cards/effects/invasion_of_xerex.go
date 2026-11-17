package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Invasion of Xerex // Vertex Paladin — March of the Machine (#5477).
//
// Front, Battle — Siege 4, {2}{W}{U}: "When this Siege enters, return up to
// one target creature to its owner's hand." Back, Vertex Paladin, Creature —
// Angel Knight */*: "Flying. Vertex Paladin's power and toughness are each
// equal to the number of creatures you control."
//
// No simplification.
const invasionOfXerexOracleID = "8b5add91-2e82-4430-bfc5-f4a6366dbbdb"

func init() {
	Register(Spec{
		OracleID:     invasionOfXerexOracleID,
		Name:         "Invasion of Xerex",
		Completeness: CompletenessFull,
		Battle:       &BattleSpec{Defense: 4, Subtype: BattleSubtypeSiege},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Invasion of Xerex — return up to one target creature to its owner's hand",
				bounceUpToOneTarget), TargetCreature("up to one target creature").WithCount(0, 1)),
			DefeatedTrigger("Invasion of Xerex — defeated: exile it, then cast Vertex Paladin", SiegeDefeated()),
		},
	})
	Register(Spec{
		OracleID:        invasionOfXerexOracleID + "#1",
		Name:            "Vertex Paladin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			s82MomB17CountPT("Vertex Paladin's power and toughness are each equal to the number of creatures you control.",
				func(g *game.Game, source *game.Card) int { return b04CreaturesControlled(g, source.Controller) }),
		},
	})
}
