package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "daaf33d8-2ee2-41f8-9872-66b6493c6e13",
		Name:            "Happy Hogan, Bodyguard",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Happy Hogan — the owner of target creature an opponent controls puts it into their library second from the top or on the bottom",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						return effects.PutIntoLibraryAtDepthOrBottom{
							Card:  t.ID,
							Depth: 2,
							Label: "Happy Hogan — put it into your library second from the top or on the bottom",
						}.Apply(ctx)
					}
					return nil
				}), effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
