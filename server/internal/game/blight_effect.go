package game

import "github.com/google/uuid"

// blight_effect.go — #4683: CR 701.68a's "blight N" done by a RESOLVING
// effect rather than paid as a cost. "When this creature enters, you may
// blight 2. If you do, …" (Sourbread Auntie), "Whenever this creature
// attacks, blight 1" (Shadow Urchin), "each opponent blights 1" (High
// Perfect Morcant).
//
//	701.68a To "blight N" means to put N -1/-1 counters on a creature you
//	control.
//	701.68b If a player is given the choice to blight but is unable to put
//	N -1/-1 counters on a creature they control (usually because they
//	control no creatures), they can't choose to blight.
//
// The two halves of the action are the same for a cost and an effect:
// which creatures may be chosen is BlightOptionsForEffect (the one walk
// the cast validator, the view and the enumerator read), and the counters
// go through the CR 614 counter window with the blighting player named as
// their placer. What differs is the mark on the event. A cost's counters
// (blightLocked) are CounterFromCost, so a replacement worded "if an
// effect would put" (Doubling Season) does not see them; an effect's are
// not, so it does (CR 614.16). And an effect may pause on a CR 616
// ordering prompt where a cost settles at once, so the rest of the effect
// is a keyed tail.
//
// The choice of creature is the caller's (effects.Blight asks it as the
// ability resolves, CR 608.2c); it is not a target, so hexproof and
// protection do not stop it.

// ChooseBlight: the named creature, one the chooser controls, gets the
// blight's -1/-1 counters (CR 701.68a, #4683). The client shows the
// question; the bot prices the pick as the counters it costs.
const ChooseBlight ChooseDestination = "blight"

// BlightKeyedForEffect puts `n` -1/-1 counters on `creatureID` as the
// EFFECT of a resolving spell or ability, `playerID` blighting: the
// counters' placer is `playerID` (CR 120.3d), and the event is not
// CounterFromCost, so Doubling Season doubles it (CR 614.16). `then`, a
// registered tail, runs once the placement has settled, told the counters
// that landed as TailOutcome.Amount — after the answer when two counter
// replacements ask to be ordered (CR 616.1).
//
// The creature must still be one `playerID` controls on the battlefield;
// one that is not (or that can't have counters put on it, CR 701.68b) gets
// nothing, and `then` is told zero. Whether the player BLIGHTED is the
// caller's question, not this function's: the player blighted once the
// creature was chosen, whatever the replacements made of the counters —
// the reading CR 701.68d gives a "whenever a player blights" trigger
// ("regardless of what events actually occurred").
//
// A cost that blights (an activated ability's "Blight 1", #4684) pays with
// blightLocked instead, which marks the event CounterFromCost and settles
// without pausing.
//
// Caller must hold g.mu (resolution frame).
func (g *Game) BlightKeyedForEffect(playerID, creatureID uuid.UUID, n int, then KeyedTail) error {
	c := findBattlefieldCard(g, creatureID)
	if n <= 0 || c == nil || c.Controller != playerID || !c.IsCreature() {
		return then.Run(g, nil)
	}
	return g.AddCounterByKeyedForEffect(playerID, creatureID, CounterMinusOne, n, then)
}
