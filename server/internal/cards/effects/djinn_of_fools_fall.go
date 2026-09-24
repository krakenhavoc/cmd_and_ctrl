package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Djinn of Fool's Fall — {4}{U} Creature — Djinn 4/3:
//
//	"Flying
//	 Plot {3}{U} (You may pay {3}{U} and exile this card from your
//	 hand. Cast it as a sorcery on a later turn without paying its mana
//	 cost. Plot only as a sorcery.)"
//
// The plot keyword's plainest proof card (#1342): a vanilla flier
// whose only other text is the keyword. Flying is enforced for every
// card from its printed keywords; plot is the CR 116.2 special action
// (effects.Plot) and the free later cast is #1318's plotted-card
// permission.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7dbb540f-ee96-4115-b6ef-28ffc73da8b7",
		Name:            "Djinn of Fool's Fall",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SpecialActions:  []game.SpecialAction{Plot("{3}{U}")},
	})
}
