package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// God-Eternal Oketra — Legendary Creature — Zombie God {3}{W}{W}, 3/6
// (EDHREC rank 2435):
//
//	"Double strike
//	 Whenever you cast a creature spell, create a 4/4 black Zombie
//	 Warrior creature token with vigilance.
//	 When God-Eternal Oketra dies or is put into exile from the
//	 battlefield, you may put it into its owner's library third from
//	 the top."
//
// The creature deck's token engine that will not stay dead. The cast
// trigger is Lifecrafter's Bestiary's condition (the spell's type
// read off the stack) making the 4/4; cast, not resolve, so a
// countered creature still makes a Zombie. The return is one
// printed ability with two conditions — its own EventLTB bound for a
// graveyard or exile; a bounce or a library tuck is neither — with a
// "you may" prompt, and on "yes" the card is tucked to the top of
// its owner's library and slid under the top two (on the bottom when
// the library holds fewer than two, the printed ruling). The card is
// looked up as the trigger resolves: an Oketra that has since moved
// on — a commander sent to the command zone by the CR 903.9 prompt,
// a graveyard exiled in response — is left where it is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3b03358d-f87e-4939-afc9-5ee3f044146a",
		Name:            "God-Eternal Oketra",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b12CreatureSpellCastByYou(ev, source, g)
			}, "God-Eternal Oketra — create a 4/4 black Zombie Warrior with vigilance", Do(CreateToken{Template: TokenCard("4/4 black Zombie Warrior with vigilance"), N: 1})),
			Optional(On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b22SelfDiedOrWasExiledFromBattlefield(ev, source)
			}, "God-Eternal Oketra — put it into its owner's library third from the top", func(g *game.Game, item *game.StackItem) error {
				z := g.FindCardZoneForEffect(item.SourceCardID)
				if z == nil || (z.Kind != game.ZoneGraveyard && z.Kind != game.ZoneExile) {
					return nil
				}
				return b22TuckThirdFromTop(g, item.SourceCardID)
			}), "God-Eternal Oketra — put it into its owner's library third from the top?"),
		},
	})
}
