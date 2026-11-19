package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Archfiend's Vessel — Creature — Human Cleric {B}, 1/1 (Foundations
// Jumpstart, #5929):
//
//	"Lifelink
//	 When this creature enters, if it entered from your graveyard or you
//	 cast it from your graveyard, exile it. If you do, create a 5/5 black
//	 Demon creature token with flying."
//
// An intervening "if" (CR 603.4) read off the entry (EnteredFromYourGraveyard)
// or off the cast (a cast by its controller from the graveyard), so it
// doesn't trigger from a hand cast, and is looked at again as it resolves.
// The exile is Kinzu's: it is the creature itself, still that object on the
// battlefield (a Vessel that has left is a new object, CR 400.7), and the
// Demon comes only when it reached exile.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "62c240ac-f2dd-4e39-9e82-50307efa833e",
		Name:            "Archfiend's Vessel",
		Purpose:         game.Purpose{Tokens: 1, CreatureTokens: 1},
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "it entered from your graveyard or you cast it from your graveyard",
				Test: func(ctx *Context) bool {
					c, ok := ctx.Game.LookupCardForEffect(ctx.Item.SourceCardID)
					if !ok {
						return false
					}
					return archfiendsVesselReturned(ctx.Trigger().Event, &c, game.Characteristic{}, ctx.Game)
				},
			}, On(game.EventETB, archfiendsVesselReturned,
				"Archfiend's Vessel — exile it; if you do, create a 5/5 black Demon creature token with flying",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return ExileTarget{Target: item.SourceCardID, ThenTail: ctx.ItemTail(archfiendsVesselDemon)}.Apply(ctx)
				})),
		},
	})
}

// archfiendsVesselDemon is "create a 5/5 black Demon creature token with
// flying", run once the Vessel reached exile. An on-disk key.
var archfiendsVesselDemon = CardTail("archfiends-vessel/demon", func(ctx *Context, o game.TailOutcome) error {
	if !oneWentThisWay(o) {
		return nil
	}
	return CreateToken{Template: TokenCard("5/5 black Demon with flying"), N: 1}.Apply(ctx)
})
