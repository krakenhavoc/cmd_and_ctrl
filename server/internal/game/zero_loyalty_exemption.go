package game

import "github.com/google/uuid"

// zero_loyalty_exemption.go — "planeswalkers you control aren't put into
// their owners' graveyards for having 0 loyalty" (CR 704.5i, #2797):
// Sanctum Lurker.
//
// CR 704.5i is a state-based action; this is a static ability that
// stops it applying to a class of permanents, so a planeswalker the
// exemption covers stays on the battlefield at 0 loyalty (or fewer) and
// the state-based pass skips it. It is not a replacement effect and not
// indestructible: nothing is "prevented", the action simply does not
// apply. Every other state-based action still does, so a planeswalker
// that is also a creature with lethal damage dies to CR 704.5g, and one
// that stops being a planeswalker is a different permanent altogether.
//
// Read live off the battlefield each time the state-based pass meets a
// planeswalker with no loyalty, keyed by CatalogAbilityKey, and never
// stored (the LegendRuleExemption pattern, which this copies): two
// exempting permanents compose, one leaving cannot revoke the other's
// exemption, and one that lost its abilities exempts nothing. When the
// last one leaves, the very next pass puts every uncovered 0-loyalty
// planeswalker into its owner's graveyard, all as one event (CR 704.3).
//
// WHAT THE EXEMPTION DOES NOT DO. It does not let the walker activate a
// loyalty ability it could not pay for: CR 606.6 still asks for N
// counters, so a 0-loyalty walker can use a plus ability and nothing
// else. It does not stop the walker being attacked or dealt damage, and
// damage dealt to it still removes loyalty counters (CR 120.3c), which
// simply stay at zero.

// ZeroLoyaltyExemptionWhose is whose planeswalkers a printed exemption
// covers.
type ZeroLoyaltyExemptionWhose uint8

const (
	// ZeroLoyaltyExemptYours is "Planeswalkers you control aren't put
	// into their owners' graveyards for having 0 loyalty": the
	// exempting permanent's controller (Sanctum Lurker).
	ZeroLoyaltyExemptYours ZeroLoyaltyExemptionWhose = iota
	// ZeroLoyaltyExemptEveryone is the same sentence about every
	// player's planeswalkers. No printed card says it today; the shape
	// is here so a card that does is a declaration, not an engine change.
	ZeroLoyaltyExemptEveryone
)

// ZeroLoyaltyExemption is one printed CR 704.5i exemption. Catalog data.
type ZeroLoyaltyExemption struct {
	// Label is the clause as printed.
	Label string
	Whose ZeroLoyaltyExemptionWhose
}

// CatalogZeroLoyaltyExemptions returns the exemptions a permanent with
// the given catalog key has. carddef.go sets it from CardDef.
var CatalogZeroLoyaltyExemptions func(key string) []ZeroLoyaltyExemption

// zeroLoyaltyExempt is the set of controllers the exemptions on the
// battlefield cover right now.
type zeroLoyaltyExempt struct {
	all         bool
	controllers map[uuid.UUID]bool
}

func (e zeroLoyaltyExempt) covers(controller uuid.UUID) bool {
	return e.all || e.controllers[controller]
}

// zeroLoyaltyExemptControllersLocked reads every exemption on the
// battlefield. Caller must hold g.mu (read or write). Reads only.
func (g *Game) zeroLoyaltyExemptControllersLocked() zeroLoyaltyExempt {
	var out zeroLoyaltyExempt
	if CatalogZeroLoyaltyExemptions == nil || g.Battlefield == nil {
		return out
	}
	for i := range g.Battlefield.Cards {
		src := &g.Battlefield.Cards[i]
		key := catalogAbilityKeyOf(src)
		if key == "" {
			continue
		}
		for _, s := range CatalogZeroLoyaltyExemptions(key) {
			switch s.Whose {
			case ZeroLoyaltyExemptEveryone:
				out.all = true
			case ZeroLoyaltyExemptYours:
				if out.controllers == nil {
					out.controllers = map[uuid.UUID]bool{}
				}
				out.controllers[src.Controller] = true
			}
		}
	}
	return out
}

// ZeroLoyaltyExemptForEffect reports whether CR 704.5i does not apply
// to `walker` right now, for a caller outside the state-based pass (a
// test, a bot's "is this walker about to die" read). Caller must hold
// g.mu (read or write).
func (g *Game) ZeroLoyaltyExemptForEffect(walker Card) bool {
	return g.zeroLoyaltyExemptControllersLocked().covers(walker.Controller)
}
