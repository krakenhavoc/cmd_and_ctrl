package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Herald of Hoofbeats — Creature — Human Knight {3}{U}, 3/3 (Commander
// Legends of March of the Machine):
//
//	"Horsemanship (This creature can't be blocked except by creatures with
//	 horsemanship.)
//	 Other Knights you control have horsemanship."
func init() {
	Register(Spec{
		OracleID:        "c9b40f24-c08e-4e05-a141-0a8a670d37c9",
		Name:            "Herald of Hoofbeats",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"horsemanship"},
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Knight"}, Others: true, YoursOnly: true}, "horsemanship"),
		},
	})
}
