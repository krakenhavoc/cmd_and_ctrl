package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Happy Hogan, Bodyguard — {5}{U} Legendary Creature — Human Advisor 3/3
// (Marvel Super Heroes Commander, #4455):
//
//	"Vigilance
//	 When Happy Hogan enters, the owner of target creature an opponent
//	 controls puts it into their library second from the top or on the
//	 bottom."
//
// Temporal Cleansing's clause (PutIntoLibraryAtDepthOrBottom): the chooser
// is the creature's OWNER, not Happy Hogan's controller, so a stolen
// creature goes back to its true owner's choice. The move is the tuck route
// either way, so a commander is offered the command zone (CR 903.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "daaf33d8-2ee2-41f8-9872-66b6493c6e13",
		Name:            "Happy Hogan, Bodyguard",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Happy Hogan — the owner of target creature an opponent controls puts it into their library second from the top or on the bottom",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						return PutIntoLibraryAtDepthOrBottom{
							Card:  t.ID,
							Depth: 2,
							Label: "Happy Hogan — put it into your library second from the top or on the bottom",
						}.Apply(ctx)
					}
					return nil
				}), TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
