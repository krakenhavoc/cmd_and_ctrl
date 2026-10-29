package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Avengers: Under Siege — Enchantment — Saga {2}{B}{R} (Marvel Super Heroes):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Create two 2/1 black Villain creature tokens with menace.
//	 II — This Saga deals 2 damage to each non-Villain creature and each
//	 opponent.
//	 III — Create a Treasure token for each Villain you control."
//
// Chapter I is the table's 2/1 black Villain row, twice. Chapter II is The
// Elder Dragon War's one damage instance (every matching creature, then each
// opponent still in the game, in seat order) with the creature set narrowed
// to the non-Villains; the Saga itself is the damage source, and the Villain
// test reads the creature's effective subtypes as it resolves, so a
// changeling is spared. Chapter III counts the Villains you control as it
// resolves, the tokens chapter I made included, and makes that many Treasures.
//
// No simplification.
func init() {
	const name = "Avengers: Under Siege"
	Register(Spec{
		OracleID:     "d911cf65-7273-4162-a766-7d4107699dc2",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "create two 2/1 black Villains with menace"),
				Do(CreateToken{Template: TokenCard("2/1 black Villain with menace"), N: 2})),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "2 damage to each non-Villain creature and each opponent"),
				avengersUnderSiegeDamage),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "a Treasure for each Villain you control"),
				avengersUnderSiegeTreasures),
		},
	})
}

// avengersUnderSiegeDamage is chapter II.
func avengersUnderSiegeDamage(g *game.Game, item *game.StackItem) error {
	return damageEachMatchingAndEachOpponent(NewContext(g, item), And(Creature(), Not(Subtype("Villain"))), 2)
}

// avengersUnderSiegeTreasures is chapter III.
func avengersUnderSiegeTreasures(g *game.Game, item *game.StackItem) error {
	n := len(permanentsControlledByMatching(g, item.Controller, Subtype("Villain")))
	if n == 0 {
		return nil
	}
	return CreateToken{Template: TreasureToken(), N: n}.Apply(NewContext(g, item))
}
