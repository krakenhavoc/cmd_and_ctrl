package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Mighty Thor, Jane Foster — Legendary Creature — Human God Hero
// {1}{W}{U}, 3/3:
//
//	"Flying
//	 Whenever The Mighty Thor attacks, exile up to one target nontoken
//	 artifact or creature, then return that card to the battlefield
//	 tapped under its owner's control.
//	 Whenever an Equipment you control enters, draw a card."
//
// The attack trigger is an immediate flicker (Flicker, Tapped): exile
// and return resolve together, so the permanent comes back a NEW
// object (CR 400.7) — tapped, which is what makes it removal on a
// blocker and a re-used ETB on your own creature. "Under its owner's
// control" leaves Flicker's Controller zero, so a stolen creature goes
// home. "Up to one" is the (0, 1) count; "nontoken" keeps a token out
// of the picker (it would cease to exist in exile, CR 111.8). Thor
// herself is a legal target — the clause has no "other".
//
// The Equipment trigger watches EventETB for a permanent with the
// Equipment subtype entering under Thor's controller's control; a
// token Equipment counts, and so does Thor's own entry if something
// has made her an Equipment.
func init() {
	Register(Spec{
		OracleID:        "57d02dc8-e22e-4874-9f02-490a2528a28f",
		Name:            "The Mighty Thor, Jane Foster",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverThisAttacks("The Mighty Thor — exile and return a nontoken artifact or creature tapped", mightyThorFlickerTapped),
				TargetPermanent("up to one target nontoken artifact or creature",
					Or(Artifact(), Creature()), Not(IsTokenPredicate())).WithCount(0, 1),
			),
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.HasSubtype("Equipment")
			}, "The Mighty Thor — draw a card (an Equipment entered)", Do(DrawCards{N: 1})),
		},
	})
}

// mightyThorFlickerTapped exiles the chosen permanent, if it is still
// a legal target, and returns it tapped under its owner's control.
func mightyThorFlickerTapped(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return Flicker{Target: id, Tapped: true}.Apply(ctx)
}
