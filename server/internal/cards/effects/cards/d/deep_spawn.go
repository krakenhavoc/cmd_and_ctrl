package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Deep Spawn — Creature — Homarid {5}{U}{U}{U}, 6/6 (Fallen Empires, S82 OLD batch
// 852, #5637):
//
//	"Trample
//	 At the beginning of your upkeep, sacrifice this creature unless you mill two
//	 cards.
//	 {U}: This creature gains shroud until end of turn and doesn't untap during
//	 your next untap step. Tap this creature. (A creature with shroud can't be the
//	 target of spells or abilities.)"
//
// Homarid Warrior's shroud, freeze and tap (S82OldB852ShroudFreezeAndTap). The
// upkeep tax is a choice between milling two cards and the sacrifice
// (S82OldB852SacrificeUnlessYouMill); a library with fewer than two cards can't
// pay and the Spawn is sacrificed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d1aba343-d85a-4829-bdfe-db31fcb3c320",
		Name:            "Deep Spawn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Deep Spawn — sacrifice this creature unless you mill two cards", effects.S82OldB852SacrificeUnlessYouMill("Deep Spawn", 2)),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}: This creature gains shroud until end of turn and doesn't untap during your next untap step. Tap this creature.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    effects.ManaCost("{U}"),
			Effect:  effects.S82OldB852ShroudFreezeAndTap("Deep Spawn"),
		}},
	})
}
