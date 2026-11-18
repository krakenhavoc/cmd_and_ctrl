package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chrome Courier — Artifact Creature — Thopter {1}{W}{U}, 1/1:
//
//	"Flying
//	 When this creature enters, reveal the top two cards of your library.
//	 Put one of them into your hand and the other into your graveyard. If
//	 you put an artifact card into your hand this way, you gain 3 life."
//
// Faerie Snoop's take (one to hand, the rest to the graveyard) over a
// REVEALED pair, with the life gain read off the card that reached the
// hand. A library of one card gives that card to the hand with no
// choice; an empty library does nothing. The take is registered, so a
// table waiting on the pick is a restore point.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a07ceaca-9964-4c24-937e-090b6643b327",
		Name:            "Chrome Courier",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Chrome Courier — reveal the top two cards of your library, one to hand and the other to your graveyard; gain 3 life if it's an artifact",
				chromeCourierReveal),
		},
	})
}

// chromeCourierTake is the take's registration: the rest to the
// graveyard, then the life if the card taken is an artifact. On disk:
// never renamed, never reused.
var chromeCourierTake = RegisterTake("choose-cards/chrome-courier-take",
	TakeDef{Then: func(g *game.Game, res TakeFromLibraryResult) error {
		if err := TakeRestIntoGraveyard(g, res); err != nil {
			return err
		}
		if len(res.Taken) == 0 {
			return nil
		}
		c, ok := g.LookupCardForEffect(res.Taken[0])
		if !ok || !c.IsArtifact() {
			return nil
		}
		return g.ChangePlayerLifeForEffect(res.Source, res.Player, 3)
	}})

func chromeCourierReveal(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := ctx.Controller()
	revealed := g.RevealTopOfLibraryForEffect(player, ctx.Source(), 2,
		"Chrome Courier — reveal the top two cards of your library")
	return TakeFromLibraryToHand{
		Player: player,
		Cards:  revealed,
		Max:    1,
		Label:  "Chrome Courier — put one of them into your hand and the other into your graveyard",
		Key:    chromeCourierTake,
	}.Apply(ctx)
}
