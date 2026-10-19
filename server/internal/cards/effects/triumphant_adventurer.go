package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Triumphant Adventurer — Creature — Human Knight {W}{B}, 1/1:
//
//	"Deathtouch
//	 During your turn, this creature has first strike.
//	 Whenever this creature attacks, venture into the dungeon."
//
// "During your turn" is Leonin Surveyor's layer 6 grant, read on whose
// turn it is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6aac0ae4-d9a7-4c40-a39e-229186c84d74",
		Name:            "Triumphant Adventurer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return thisPermanent(target, g, source) && isActivePlayer(g, source.Controller)
			}, "first strike"),
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks(ventureLabel("Triumphant Adventurer"), Venture),
		},
	})
}
