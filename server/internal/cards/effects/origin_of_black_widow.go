package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Origin of Black Widow — {2}{B} Enchantment — Saga (Marvel Super Heroes
// Commander, #4455):
//
//	"I — Each opponent sacrifices a creature of their choice.
//	 II — Creatures you control gain deathtouch until end of turn.
//	 III — Each opponent loses 1 life for each creature card in their
//	       graveyard."
//
// Chapter I is the edict of The Eldest Reborn (each opponent chooses their
// own). Chapter III counts per opponent, as it resolves, in that opponent's
// own graveyard (The One Ring Rule Them All's chapter III, oneRingDrain); the
// life loss is not damage and is one change per opponent.
//
// No simplification.
func init() {
	const name = "Origin of Black Widow"
	Register(Spec{
		OracleID:     "3bc0479f-0f45-4b36-9963-f8eb5973cf47",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "each opponent sacrifices a creature"), originOfBlackWidowEdict),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "creatures you control gain deathtouch until end of turn"), originOfBlackWidowDeathtouch),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "each opponent loses 1 life for each creature card in their graveyard"), oneRingDrain),
		},
	})
}

func originOfBlackWidowEdict(g *game.Game, item *game.StackItem) error {
	return EachPlayerSacrifices{
		ExceptController: true,
		Match:            Creature(),
		Label:            "a creature",
	}.Apply(NewContext(g, item))
}

func originOfBlackWidowDeathtouch(g *game.Game, item *game.StackItem) error {
	return GrantKeywordUntilEOT{
		Match:    And(Creature(), YouControl()),
		Keywords: []string{"deathtouch"},
		Label:    "Origin of Black Widow — creatures you control gain deathtouch",
	}.Apply(NewContext(g, item))
}
