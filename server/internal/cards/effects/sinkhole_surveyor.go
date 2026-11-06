package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sinkhole Surveyor — Creature — Bird Scout {1}{B}, 1/3 (Tarkir:
// Dragonstorm, #4463):
//
//	"Flying
//	 Whenever this creature attacks, you lose 1 life and this creature
//	 endures 1. (Put a +1/+1 counter on it or create a 1/1 white Spirit
//	 creature token.)"
//
// One attack trigger, in printed order: the life loss, then endure 1
// (endure_n.go, CR 701.63a). The life is lost whichever way the endure is
// answered, and a Surveyor that has left the battlefield by resolution can
// only make the Spirit.
//
// No simplification.
func init() {
	endure := ThisEndures("Sinkhole Surveyor", 1)
	Register(Spec{
		OracleID:        "83a47724-b8d1-4884-9442-b62aa1510dc3",
		Name:            "Sinkhole Surveyor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Sinkhole Surveyor — you lose 1 life and this creature endures 1",
				func(g *game.Game, item *game.StackItem) error {
					if err := (GainLife{Amount: -1}).Apply(NewContext(g, item)); err != nil {
						return err
					}
					return endure(g, item)
				}),
		},
	})
}
