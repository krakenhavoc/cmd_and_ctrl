package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Growing Rites of Itlimoc — Legendary Enchantment {2}{G} // Itlimoc,
// Cradle of the Sun:
//
//	"When Growing Rites of Itlimoc enters, look at the top four cards
//	 of your library. You may reveal a creature card from among them
//	 and put it into your hand. Put the rest on the bottom of your
//	 library in any order.
//	 At the beginning of your end step, if you control four or more
//	 creatures, transform Growing Rites of Itlimoc."
//
// The front face of a transforming double-faced card (CR 712). Its
// back, Itlimoc, Cradle of the Sun, is itlimoc_cradle_of_the_sun.go,
// registered under "<oracle_id>#1".
//
// The enters trigger is Oath of Nissa's look: only the controller sees
// the four, the creature taken is revealed, and the rest go to the
// bottom in an order the player chooses.
//
// The end-step trigger has an intervening if (CR 603.4): it triggers
// only if you control four or more creatures as your end step begins,
// and does nothing if you don't as it resolves (the 2017-09-29
// rulings), Storm the Vault's shape. A creature that will leave "at
// the beginning of the next end step", or a permanent that is a
// creature until end of turn, still counts while it is there. The
// transform is the ADR 0079 in-place verb, which ignores a second
// instruction once the permanent has transformed (CR 701.27f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     growingRitesOfItlimocOracleID,
		Name:         "Growing Rites of Itlimoc",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Growing Rites of Itlimoc — look at the top four cards of your library",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					player := ctx.Controller()
					return TakeFromLibraryToHand{
						Player:   player,
						Cards:    g.LookAtTopOfLibraryForEffect(player, 4),
						Match:    Creature(),
						Max:      1,
						Optional: true,
						Reveal:   true,
						Label:    "Growing Rites of Itlimoc — reveal a creature card and put it into your hand",
						Then:     TakeRestOnBottomInAnyOrder,
					}.Apply(ctx)
				}),
			On(game.EventBeginEndStep, growingRitesHasFourCreatures,
				"Growing Rites of Itlimoc — transform it",
				func(g *game.Game, item *game.StackItem) error {
					// CR 603.4's second check, as it resolves.
					if countControlled(g, item.Controller, MatchCreature) < 4 {
						return nil
					}
					return TransformThis{}.Apply(NewContext(g, item))
				}),
		},
	})
}

// growingRitesHasFourCreatures is the intervening if's first check
// (CR 603.4), as your end step begins.
func growingRitesHasFourCreatures(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.Actor == source.Controller &&
		countControlled(g, source.Controller, MatchCreature) >= 4
}

// growingRitesOfItlimocOracleID is shared with the back face, which
// registers under it plus "#1" (game.CatalogKey).
const growingRitesOfItlimocOracleID = "ea9c459a-6047-43aa-968f-a582be4000e8"
