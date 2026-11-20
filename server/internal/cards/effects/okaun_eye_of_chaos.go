package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Okaun, Eye of Chaos — Legendary Creature — Cyclops Berserker {4}{R},
// 3/3 (Battlebond, #5529):
//
//	"Partner with Zndrsplt, Eye of Wisdom (When this creature enters,
//	 target player may put Zndrsplt into their hand from their library,
//	 then shuffle.)
//	 At the beginning of combat on your turn, flip a coin until you lose a
//	 flip.
//	 Whenever a player wins a coin flip, double Okaun's power and toughness
//	 until end of turn."
//
// Partner with is the shared CR 702.124j pair (PartnerWith). The combat
// trigger is the shared Battlebond chain (bbdFlipUntilYouLose): one called
// flip at a time (CR 705.2), each won flip asking the next, until a flip is
// lost. Every won flip, by any player and from any source, triggers the
// doubling (bbdAPlayerWonACoinFlip), so Okaun's own chain doubles it once
// per win and a partnered Zndrsplt's chain doubles it too. Each doubling
// gives +X/+X where X is Okaun's power and toughness as it resolves
// (DoublePowerAndToughnessOfEachUntilEOT); an Okaun that left and came back
// is a new object and is not doubled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7dec219b-96d1-4d4c-b3ac-15eda9d2ccc6",
		Name:         "Okaun, Eye of Chaos",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			PartnerWith("Okaun, Eye of Chaos", "Zndrsplt, Eye of Wisdom"),
			AtBeginningOfYourCombat("Okaun, Eye of Chaos — flip a coin until you lose a flip",
				bbdFlipUntilYouLose("Okaun, Eye of Chaos — call the coin flip (flip until you lose)")),
			On(game.EventFlipCoin, bbdAPlayerWonACoinFlip,
				"Okaun, Eye of Chaos — double its power and toughness until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if sourceIsNewObject(g, item) {
						return nil
					}
					return DoublePowerAndToughnessOfEachUntilEOT(NewContext(g, item), []uuid.UUID{item.SourceCardID},
						"Okaun, Eye of Chaos — double power and toughness until end of turn")
				}),
		},
	})
}
