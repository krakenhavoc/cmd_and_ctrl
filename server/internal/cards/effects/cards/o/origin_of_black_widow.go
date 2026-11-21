package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "3bc0479f-0f45-4b36-9963-f8eb5973cf47",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "each opponent sacrifices a creature"), effects.EachOpponentSacrificesACreature),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "creatures you control gain deathtouch until end of turn"), originOfBlackWidowDeathtouch),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "each opponent loses 1 life for each creature card in their graveyard"), effects.OneRingDrain),
		},
	})
}

func originOfBlackWidowDeathtouch(g *game.Game, item *game.StackItem) error {
	return effects.GrantKeywordUntilEOT{
		Match:    effects.And(effects.Creature(), effects.YouControl()),
		Keywords: []string{"deathtouch"},
		Label:    "Origin of Black Widow — creatures you control gain deathtouch",
	}.Apply(effects.NewContext(g, item))
}
