package game

// counter_types.go is the canonical registry of MTG counter names
// used by the engine. Counter strings are free-form (Card.Counters
// and Player.Counters are both string-keyed) so unknown / homebrew
// names round-trip without validation, but the named constants here
// give the SBA loop and the wire view stable identifiers and let
// the client mirror this list for iconography + colour.
//
// The full canonical taxonomy from the MTG comprehensive rules has
// ~80 entries. The constants below cover the ones the S13.x engine
// references directly (loyalty, defense, +1/+1, -1/-1, poison,
// energy, experience, rad). The client's iconography list
// (client/src/lib/counterTypes.ts) carries a slightly larger
// pinned set; everything else renders as text-only pips.

// Card-level counter type identifiers.
const (
	// CounterPlusOne is the +1/+1 counter. Stacks with other
	// +1/+1 counters (cf. CR 122.1a) and cancels 1-for-1 against
	// CounterMinusOne via the SBA in CR 704.5q.
	CounterPlusOne = "+1/+1"
	// CounterMinusOne is the -1/-1 counter; same shape as
	// CounterPlusOne but with negative effect on P/T.
	CounterMinusOne = "-1/-1"
	// CounterLoyalty is the planeswalker loyalty counter (CR 606).
	// SBA: a planeswalker with 0 loyalty counters → graveyard
	// (CR 704.5i).
	CounterLoyalty = "loyalty"
	// CounterDefense is the battle defense counter (post-MoM).
	// SBA: a battle with 0 defense counters → graveyard
	// (CR 704.5v).
	CounterDefense = "defense"
	// CounterCharge is a generic resource counter (Aether Vial,
	// Coalition Relic, etc.). No SBA.
	CounterCharge = "charge"
	// CounterStorage is the storage counter (Mage-Ring Network,
	// Crucible of the Spirit Dragon, the Mirrodin storage lands). A
	// generic resource counter like CounterCharge, distinct because
	// the cards that bank them spend them by the handful and print
	// "storage counter" in the cost. No SBA. Added with #789's
	// variable counter cost.
	CounterStorage = "storage"
	// CounterStun is the stun counter (post-NEO). A would-be untap
	// removes one instead; untapPermanentLocked enforces it for every
	// untap, not as a state-based action.
	CounterStun = "stun"
	// CounterShield is the shield counter (post-MOM).
	CounterShield = "shield"
	// CounterLore is the saga lore counter (CR 714). SBA: a saga
	// whose final-chapter lore counter is set is sacrificed by
	// its controller (CR 704.5s). The advance-chapter trigger
	// lands in S14+ with the effect catalog.
	CounterLore = "lore"
)

// Player-level counter type identifiers.
const (
	// CounterPoison is the poison counter (CR 122.1d). 10 poison
	// counters lose the game (CR 704.5c). Player.Poison stays as
	// a duplicate int field for backwards-compat with the S10
	// SetPoison action — both fields are kept in sync.
	CounterPoison = "poison"
	// CounterEnergy is the energy counter (Kaladesh resource
	// counter; CR 107.14). No game-loss condition. Player.Energy
	// stays as a duplicate int field for backwards-compat.
	CounterEnergy = "energy"
	// CounterExperience is the experience counter (Commander
	// 2015; CR 122.1d).
	CounterExperience = "experience"
	// CounterRad is the rad counter (Fallout / Warhammer 40K
	// crossover; CR 122.1d). Triggers a draw + life-loss as the
	// player draws — the trigger is S14+ territory.
	CounterRad = "rad"
)

// PoisonLethal is the poison-counter total at which a player loses
// the game (CR 704.5c). Codified as a constant so the SBA threshold
// is easy to find.
const PoisonLethal = 10

// KnownCardCounters is the slice form of the card-level counter
// constants for iteration. The wire surface ships this on demand
// (counter_types endpoint, when added) so clients can stay in sync
// without round-tripping through the action protocol.
var KnownCardCounters = []string{
	CounterPlusOne,
	CounterMinusOne,
	CounterLoyalty,
	CounterDefense,
	CounterCharge,
	CounterStorage,
	CounterStun,
	CounterShield,
	CounterLore,
}

// KnownPlayerCounters is the slice form of the player-level counter
// constants. Same role as KnownCardCounters for the player axis.
var KnownPlayerCounters = []string{
	CounterPoison,
	CounterEnergy,
	CounterExperience,
	CounterRad,
}
