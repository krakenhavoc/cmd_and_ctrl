package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Winter's Night — World Enchantment {R}{G}{W}:
//
//	"Whenever a player taps a snow land for mana, that player adds one
//	 mana of any type that land produced. That land doesn't untap during
//	 its controller's next untap step."
//
// #1997, one of the world enchantments ADR 0109 PR 7 left out. A
// triggered mana ability (CR 605.1b): Mana Flare's, narrowed to a snow
// land, so it fires for every player's snow land, adds its mana to the
// player who tapped it with no priority window (CR 605.4a), and offers
// only the types the land just produced. Its second sentence is the
// ability's rider (game.ManaTrigger.TappedSkipsNextUntap): the land gets
// the ordinary next-untap marker for its controller (CR 502.3), so a
// snow land tapped on an opponent's turn stays tapped through its
// controller's next untap step and untaps the one after.
//
// The snow supertype is read off the land as it was when it produced,
// so an effect that makes a land snow counts. Like Mana Flare, the
// auto-tap planner does not count the extra mana (ADR 0074 §7).
//
// A world permanent: the world rule (CR 704.5k, ADR 0109 §8) puts it
// into its owner's graveyard when a newer one enters.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "e4f73802-b639-47ab-a6c6-f761afa4e00e",
		Name:         "Winter's Night",
		Completeness: CompletenessFull,
		ManaTriggers: []game.ManaTrigger{{
			Label: "Winter's Night — add one mana of any type that snow land produced; it doesn't untap during its controller's next untap step",
			AppliesTo: func(prod game.ManaProduced, _ *game.Card, _ *game.Game) bool {
				return prod.Source.IsLand() && prod.Source.HasSupertype("Snow")
			},
			Produced:             AddsOneManaOfAnyTypeProduced(),
			TappedSkipsNextUntap: true,
		}},
	})
}
