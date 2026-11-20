package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Teferi, Druid of Argoth — Legendary Creature — Human Druid
// {2}{G}{G}{G}, 3/4 (Mystery Booster 2):
//
//	"Flash
//	 Creature cards you own that aren't on the battlefield have flash.
//	 Each opponent can cast spells only any time they could cast a
//	 sorcery."
//
// Teferi, Mage of Zhalfir's two clauses, shared (teferi_mage_of_zhalfir.go):
// the owned-creature flash grant and the opponents' sorcery-speed bar
// (CR 101.2). Teferi's own flash is the printed keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f2c9957e-d673-4b32-a201-811849c5de31",
		Name:            "Teferi, Druid of Argoth",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		CastTimings: []game.CastTimingRule{
			finB12OwnedCreatureCardsHaveFlash("Creature cards you own that aren't on the battlefield have flash."),
			OpponentsCastAtSorcerySpeed(),
		},
	})
}
