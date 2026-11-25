package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rite of Belzenlok — Enchantment — Saga {2}{B}{B} (Dominaria):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I, II — Create two 0/1 black Cleric creature tokens.
//	 III — Create a 6/6 black Demon creature token with flying, trample, and
//	 "At the beginning of your upkeep, sacrifice another creature. If you can't,
//	 this token deals 6 damage to you.""
//
// I and II are one body over two chapters: two Clerics together. III is Thomil's
// Lord of the Pit at 6/6 with 6 damage: a catalog token whose upkeep trigger
// makes its controller sacrifice another creature (their choice, a creature
// that can't be sacrificed excepted), or take 6 damage from the token.
//
// No simplification.
func init() {
	const name = "Rite of Belzenlok"
	clerics := effects.Do(effects.CreateToken{Template: effects.TokenCard("0/1 black Cleric"), N: 2})
	effects.Register(effects.Spec{
		OracleID:     "a3b2e0a2-f4ad-457b-ae70-7732bc3479ff",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 5, CreatureTokens: 5},
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "create two 0/1 black Cleric creature tokens"), clerics),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "create two 0/1 black Cleric creature tokens"), clerics),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "create a 6/6 black Demon creature token with flying, trample, and an upkeep sacrifice"),
				effects.Do(effects.CreateToken{Template: effects.OldB207BelzenlokDemonToken(), N: 1})),
		},
	})
}
