package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Warring Triad — Legendary Artifact Creature — God {3}, 5/5
// (EDHREC rank 3283):
//
//	"Flying, trample, haste
//	 As long as there are fewer than eight cards in your graveyard,
//	 The Warring Triad isn't a creature.
//	 {T}, Mill a card: Target player adds one mana of any color.
//	 (Activate only as an instant.)"
//
// A three-mana 5/5 flying trampler with haste that is a plain
// artifact until the graveyard fills. The keywords ride
// PrintedKeywords. "Isn't a creature" is a layer 4 static on itself
// gated on the controller's graveyard (b31NotACreatureWhileGraveyardBelow):
// while the graveyard holds fewer than eight cards its Creature type
// is dropped — and its creature type God with it (CR 205.1b) — so it
// cannot attack or block, does not die to a wrath or to damage, and
// is not summoning sick when it wakes because it has been on the
// battlefield all along; Legendary and Artifact stay. Eight cards
// and it is a creature at once.
//
// Engine gap it shares with Elvish Reclaimer and Multani, not the
// card's: the layer cache is invalidated by battlefield motion,
// counters, taps and turn changes, not by a card reaching or
// leaving a graveyard from a hand, a library or the stack, so an
// eighth card milled or discarded shows on the Triad at the next
// recompute — the next tap, counter, zone move onto or off the
// battlefield, or turn — rather than the instant it lands.
//
// DECLARED SIMPLIFICATION, weaker than printed: the mana ability is
// not implemented. "Mill a card" is a cost component the engine
// cannot express — AbilityCost carries tap, sacrifice, mana, life,
// loyalty and crew, and nothing that mills — and deferring the mill
// to resolution would let a response change which card is milled,
// or let an empty-library activation add mana the printed card
// could not, the #259 direction. The Triad is still the god that
// wakes up, which is what it is played for.
func init() {
	Register(Spec{
		OracleID:        "df0b5995-117b-4ba8-964d-ea3c592621ef",
		Name:            "The Warring Triad",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"The tap-and-mill ability isn't implemented — it can't add mana for a player."},
		PrintedKeywords: []string{"flying", "trample", "haste"},
		Static: []game.StaticAbility{
			b31NotACreatureWhileGraveyardBelow(8),
		},
	})
}
