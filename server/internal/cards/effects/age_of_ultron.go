package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Age of Ultron — Enchantment — Saga {4}{B} (Marvel Super Heroes Commander):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — For each opponent, destroy up to one target nonartifact creature that
//	 player controls.
//	 II — For each opponent, you create a 2/2 colorless Robot Villain artifact
//	 creature token.
//	 III — Artifact creatures you control gain deathtouch until end of turn.
//	 Put a +1/+1 counter on each of them."
//
// Chapter I is Dismantling Wave's per-opponent "up to one" clause list
// (perOpponentClauses): each pick is bound to its opponent, so a creature that
// changes hands in response is an illegal target for the clause it was chosen
// under (CR 608.2b). They die together (destroyEachLegalTarget). Chapter II
// makes one Robot Villain per opponent still in the game, the table's 2/2
// artifact row. Chapter III reads the artifact creatures as it resolves, the
// tokens of chapter II included, and counters exactly the creatures it gave
// deathtouch to.
//
// No simplification.
func init() {
	const name = "Age of Ultron"
	destroy := ChapterTrigger(1, SagaChapterLabel(name, 1, "for each opponent, destroy up to one target nonartifact creature that player controls"),
		func(g *game.Game, item *game.StackItem) error {
			return destroyEachLegalTarget(item, NewContext(g, item))
		})
	destroy.Targets = perOpponentClauses(game.MaxPlayers-1, "nonartifact creature", And(Creature(), Not(Artifact())))
	Register(Spec{
		OracleID:     "51042581-4add-4487-a1b6-85446539b975",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			destroy,
			ChapterTrigger(2, SagaChapterLabel(name, 2, "a 2/2 Robot Villain artifact for each opponent"), ageOfUltronRobots),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "artifact creatures you control gain deathtouch and get a +1/+1 counter"), ageOfUltronSurge),
		},
	})
}

// ageOfUltronRobots is chapter II.
func ageOfUltronRobots(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	n := len(ctx.Opponents())
	if n == 0 {
		return nil
	}
	return CreateToken{Template: TokenCard("2/2 colorless Robot Villain artifact"), N: n}.Apply(ctx)
}

// ageOfUltronSurge is chapter III.
func ageOfUltronSurge(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	artifactCreatures := And(Artifact(), Creature(), YouControl())
	if err := (GrantKeywordUntilEOT{Match: artifactCreatures, Keywords: []string{"deathtouch"},
		Label: "Age of Ultron — deathtouch until end of turn"}).Apply(ctx); err != nil {
		return err
	}
	var ids []uuid.UUID
	for _, c := range MatchingBattlefield(ctx, artifactCreatures) {
		ids = append(ids, c.InstanceID)
	}
	return b13PutCounterOnEach(ctx, ids)
}
