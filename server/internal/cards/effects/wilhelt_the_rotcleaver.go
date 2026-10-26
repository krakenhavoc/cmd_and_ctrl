package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wilhelt, the Rotcleaver — Legendary Creature — Zombie Warrior
// {2}{U}{B}, 3/3:
//
//	"Whenever another Zombie you control dies, if it didn't have
//	 decayed, create a 2/2 black Zombie creature token with decayed.
//	 (It can't block. When it attacks, sacrifice it at end of combat.)
//	 At the beginning of your end step, you may sacrifice a Zombie. If
//	 you do, draw a card."
//
// The death trigger is Diregraf Captain's condition (another Zombie you
// controlled, judged on its last-known subtypes and controller) with an
// intervening "if" on its last-known ABILITIES: the departed object's
// post-layer ability list from the battlefield record
// (diedWithKeyword), so a creature that had decayed only through an
// effect (Ghouls' Night Out) counts as having it, and a decayed token
// sacrificed at end of combat makes nothing. The condition reads only
// last-known information, which cannot change, so it is judged once as
// the trigger fires (CR 603.4).
//
// The end-step sacrifice is chosen as the trigger resolves
// (maySacrificeOne): any Zombie, Wilhelt included, and the card is
// drawn only if one was really sacrificed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eada96a1-4964-4ad0-b0e3-2df08d6ae188",
		Name:         "Wilhelt, the Rotcleaver",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(On(game.EventLTB, anotherUndecayedZombieYouControlDied,
				"Wilhelt, the Rotcleaver — create a 2/2 Zombie with decayed",
				Do(CreateToken{Template: TokenCard("2/2 black Zombie with decayed"), N: 1})),
				game.Purpose{DeathPayoff: true}),
			AtYourEndStep("Wilhelt, the Rotcleaver — you may sacrifice a Zombie to draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return maySacrificeOne(g, item, "Wilhelt, the Rotcleaver — sacrifice a Zombie to draw a card?",
						HasSubtype("Zombie"), false, wilheltDraw)
				}),
		},
	})
}

// wilheltDraw is the draw once a Zombie has really gone, keyed (#4290).
var wilheltDraw = RegisterSacrificedOne("prompt-run/wilhelt-the-rotcleaver-draw", func(ctx *Context, _ uuid.UUID) error {
	return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
})

// anotherUndecayedZombieYouControlDied is Wilhelt's trigger condition:
// another Zombie the source's controller controlled died, and it did
// not have decayed as it last existed on the battlefield.
func anotherUndecayedZombieYouControlDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return anotherZombieYouControlDied(ev, source, g) && !diedWithKeyword(ev, g, game.KeywordDecayed)
}

// diedWithKeyword reports whether the permanent an EventLTB names had
// keyword `kw` as it last existed on the battlefield (CR 603.10a): the
// post-layer ability list of its battlefield record, so a granted
// keyword counts. A card with no record this turn is judged as it now
// sits, which is its printed keywords.
func diedWithKeyword(ev game.Event, g *game.Game, kw string) bool {
	if info, ok := g.LastKnownPermanentForEffect(ev.CardID); ok && info.Left {
		return slices.Contains(info.Characteristic.Abilities, kw)
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && game.HasKeyword(&c, kw)
}
