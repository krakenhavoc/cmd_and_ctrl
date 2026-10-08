package game

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// energy_cost.go — ADR 0129 §2: paying energy.
//
// CR 107.14: "The energy symbol is {E}. It represents one energy
// counter. To pay {E}, a player removes one energy counter from
// themselves." Energy is an ordinary counter on the player (CR 122.1),
// so nothing new is stored: Player.Counters["energy"] is the total and
// Player.Energy its legacy mirror (ADR 0008 §2).
//
// Three readers ask the same questions here, so they cannot disagree
// (ADR 0033 §1, ADR 0105): ActivateCatalogAbility validates and pays,
// internal/legal offers only what the seat can pay, and the view greys
// a row the seat is short for with the refusal's own text.
//
//	AbilityEnergyCost   the energy an activation charges at X
//	PlayerEnergy        what a player has
//	EnergyShortfall     the refusal, or nil when the energy is there
//	payEnergyLocked     the ONE path that pays energy
//
// The payment opens no replacement window. CR 107.14 calls paying a
// removal, no printed card replaces removing counters from a player,
// and a "remove N counters" cost on a permanent pays the same way
// (payCounterRemovalLocked). It does emit EventPlayerCounterPlaced with
// the negative delta, the payer as Actor and the source, so the layer
// version bumps for a static that reads the total (Razorfield Ripper)
// and the "paid or lost this turn" tally (ADR 0129 §6,
// PlayerTurnTally.EnergyPaidOrLost) counts it.

// ErrInsufficientEnergy is the sentinel for an energy component the
// activating player can't pay (CR 118.3). The error the engine returns
// wraps it with the numbers: "Not enough energy (have 2, need 3)", the
// same text the view stamps on the row (EnergyShortReason).
var ErrInsufficientEnergy = errors.New("game: not enough energy to pay that cost")

// insufficientEnergyError is ErrInsufficientEnergy with the amounts.
type insufficientEnergyError struct{ have, need int }

func (e insufficientEnergyError) Error() string { return EnergyShortReason(e.have, e.need) }

func (e insufficientEnergyError) Is(target error) bool { return target == ErrInsufficientEnergy }

// EnergyShortReason is the sentence for a player with `have` energy
// facing a cost of `need`. One function, so the refusal and the greyed
// row say the same thing.
func EnergyShortReason(have, need int) string {
	return fmt.Sprintf("Not enough energy (have %d, need %d)", have, need)
}

// AbilityEnergyCost is the energy an activation of `cost` charges when
// X is announced as `x`: the printed amount, plus X for "Pay X {E}"
// (CR 107.3a). Zero for every ability with no energy component.
func AbilityEnergyCost(cost AbilityCost, x int) int {
	n := cost.Energy
	if cost.EnergyX && x > 0 {
		n += x
	}
	return n
}

// PlayerEnergy is how many energy counters p has. Nil has none.
func PlayerEnergy(p *Player) int {
	if p == nil {
		return 0
	}
	return p.Counters[CounterEnergy]
}

// EnergyShortfall is nil when p can pay `need` energy, and the refusal
// otherwise (CR 118.3: no partial payment). A need of zero or less
// always passes.
func EnergyShortfall(p *Player, need int) error {
	if need <= 0 {
		return nil
	}
	if have := PlayerEnergy(p); have < need {
		return insufficientEnergyError{have: have, need: need}
	}
	return nil
}

// payEnergyLocked removes n energy counters from `payer` to pay a cost
// whose source is `source` (CR 107.14, CR 602.1a). It is the only path
// that pays energy. It refuses, changing nothing, when the payer has
// fewer than n (CR 118.3). Zero or less pays nothing.
//
// Caller must hold g.mu.
func (g *Game) payEnergyLocked(payer uuid.UUID, n int, source uuid.UUID) error {
	if n <= 0 {
		return nil
	}
	p := g.playerByIDLocked(payer)
	if p == nil {
		return ErrPlayerNotFound
	}
	if err := EnergyShortfall(p, n); err != nil {
		return err
	}
	before := PlayerEnergy(p)
	after := before - n
	setPlayerCounterLocked(p, CounterEnergy, after)
	mirrorLegacyPlayerCounterLocked(p, CounterEnergy, after)
	g.emitPlayerCounterDeltaLocked(payer, CounterEnergy, before, after, payer, source)
	return nil
}
