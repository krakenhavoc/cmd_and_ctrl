package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dazzling Sphinx — Creature — Sphinx {3}{U}{U}, 4/5 (Outlaws of Thunder
// Junction Commander):
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, that player
//	 exiles cards from the top of their library until they exile an instant
//	 or sorcery card. You may cast that card without paying its mana cost.
//	 Then that player puts the exiled cards that weren't cast this way on the
//	 bottom of their library in a random order."
//
// Gríma, Saruman's Footman's trigger word for word (grimaSarumansFootmanExile).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "00e8b4c9-3a79-4864-a730-c0081d397c5b",
		Name:            "Dazzling Sphinx",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Dazzling Sphinx — that player exiles cards from the top of their library until an instant or sorcery card; you may cast it without paying its mana cost",
				grimaSarumansFootmanExile),
		},
	})
}
