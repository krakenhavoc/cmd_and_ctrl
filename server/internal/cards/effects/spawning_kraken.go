package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spawning Kraken — Creature — Kraken {5}{U}, 6/6 (EDHREC rank
// 4082):
//
//	"Whenever a Kraken, Leviathan, Octopus, or Serpent you control
//	 deals combat damage to a player, create a 9/9 blue Kraken
//	 creature token."
//
// The engine of the sea-monster tribe. Six mana for a 6/6 is a bad
// rate; six mana for a 6/6 that turns every connection at the table
// into a 9/9 is why Arixmethes and Runo Stromkirk decks play it. The
// Kraken counts itself among the four types, so its own first hit
// makes a token that is itself a Kraken and triggers next combat —
// the growth is exponential if nobody blocks.
//
// THE FOUR TYPES ARE READ AS EFFECTIVE SUBTYPES, not printed ones, so
// a changeling counts (CR 702.73a) and so does anything a
// type-granting lord or a Maskwood Nexus has touched. "You control"
// is measured at the moment the damage is dealt, against the Kraken's
// controller.
//
// ONE TRIGGER PER CREATURE THAT CONNECTS, not one per combat. The
// printed text is "whenever a … deals combat damage", singular, with
// no "one or more" — so three Krakens connecting is three tokens and
// three separate triggers on the stack. That is why this ability is
// NOT wrapped in OncePerBatch: the batching guard is for the "one or
// more" wording, and applying it here would make the card weaker
// than printed.
//
// The damage must be COMBAT damage to a PLAYER. A Kraken that pings a
// planeswalker, or one that deals damage with an activated ability,
// makes nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f0357833-80b5-40ba-874a-bd22ea6e4e46",
		Name:         "Spawning Kraken",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b39SeaMonsterYouControlDealtCombatDamageToPlayer(ev, source, g)
			}, "Spawning Kraken — create a 9/9 blue Kraken",
				Do(CreateToken{Template: TokenCard("9/9 blue Kraken"), N: 1})),
		},
	})
}
