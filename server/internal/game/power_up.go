package game

import (
	"fmt"
	"slices"

	"github.com/google/uuid"
)

// power_up.go — power-up, CR 702.193 (#4527, Marvel Super Heroes).
//
//	702.193a Power-up is a keyword that adds additional rules to the
//	activated ability that follows it. "Power-up — [Cost]: [Effect]"
//	means "[Cost]: [Effect]. If this permanent entered this turn, this
//	ability's cost is reduced by this permanent's mana cost. Activate
//	this ability only once."
//
//	702.193b Generic mana in the permanent's mana cost reduces generic
//	mana in the cost to activate its power-up ability. Colored and
//	colorless mana in the permanent's mana cost reduces mana of the
//	same type, and any excess reduces that much generic mana. (See
//	rule 118.7.)
//
// The bit is ActivatedAbilityShape.PowerUp. Its two rules live in the
// two places every activation already goes through:
//
//   - "Activate this ability only once" is PowerUpUsed below, which
//     reads the activation record exhaust reads (activation_tally.go):
//     per object and per printed label, written at the announce, never
//     reset, and carried by the snapshot. A new object after a zone
//     change (CR 400.7) reads a key nothing has written, so a flickered
//     creature may power up again; a copy is a new instance and has its
//     own.
//   - "this ability's cost is reduced by this permanent's mana cost" is
//     powerUpReductionLocked, taken in the CR 601.2f pass beside the
//     other reductions, so the payment, the legal-move enumerator and
//     the view's charged cost cannot disagree.
//
// "The first power-up ability you activate during each of your turns"
// is PlayerTurnTally.PowerUpActivations, bumped at the announce.
//
// Three more rules sit on the same bit (#4571):
//
//   - "can be activated an additional time" (Wonder Man) is a
//     PowerUpExtra static, summed into PowerUpLimitFor, which PowerUpUsed
//     compares the activation record against;
//   - "power-up abilities can't be activated" for a turn (Kang the
//     Conqueror) is a ModCantActivatePowerUp ScopedEffect queued with the
//     extra turn, read by ActivationGateLocked through powerUpBanLocked;
//   - "spend this mana only to activate power-up abilities" (Quinjet
//     Technician) is the ManaRestrictPowerUp tag, matched against
//     ManaSpendContext.PowerUp.

// PowerUpExtra is one "each power-up ability of permanents you control
// can be activated an additional time" static contributed by a permanent
// on the battlefield (Wonder Man, Hollywood Hero). A struct of hooks, for
// the reason BoastLimit, StaticAbility and CostModifier are: a card file
// writes a literal.
//
// Unlike a BoastLimit, which REPLACES the printed number ("twice rather
// than once"), this ADDS to it: the text says "an additional time", so
// two of them let a power-up ability be activated three times in all.
type PowerUpExtra struct {
	// Label is the clause as printed.
	Label string

	// Times is how many additional activations each power-up ability
	// the static reaches may have. Wonder Man: 1. A value below 1 changes
	// nothing and is refused at Register.
	Times int

	// Applies decides whether this permanent's static reaches the
	// power-up ability's source. `grantor` is the permanent contributing
	// the static, so grantor.Controller is the "you" of the ability.
	// Wonder Man: source.Controller == grantor.Controller. Nil is refused
	// at Register: it would reach every permanent of every player.
	//
	// Evaluated under g.mu. READ-ONLY, as BoastLimit.Applies is.
	Applies func(g *Game, source Card, grantor Card) bool

	// ActiveWhen is the ADR 0071 designation gate, the same slot
	// BoastLimit carries.
	ActiveWhen Designation
}

// CatalogPowerUpExtras is the catalog hook, mirroring CatalogBoastLimits.
// Nil, or a nil return, means the card changes nobody's power-up limit.
var CatalogPowerUpExtras func(key string) []PowerUpExtra

// powerUpExtrasOf is the extra activations a permanent contributes right
// now, ability removal and designation gate applied (CatalogAbilityKey,
// not CatalogKey: this is a static ability, ADR 0046).
func powerUpExtrasOf(c *Card) []PowerUpExtra {
	if CatalogPowerUpExtras == nil {
		return nil
	}
	key := catalogAbilityKeyOf(c)
	if key == "" {
		return nil
	}
	return activeOnly(c, CatalogPowerUpExtras(key), func(x PowerUpExtra) Designation {
		return x.ActiveWhen
	})
}

// PowerUpLimitFor is how many times each power-up ability of `source`
// may be activated: the printed once (CR 702.193a), plus one for every
// battlefield static that grants an additional time (CR 113.6 — a static
// works from the battlefield, so this is a battlefield walk).
//
// Read at the moment of activation, like BoastLimitFor: nothing is stored
// that could outlive the static, which is the duration. If Wonder Man
// leaves after a power-up ability has been activated twice, it is simply
// spent (two activations against a limit of one).
//
// Caller must hold g.mu.
func (g *Game) PowerUpLimitFor(source Card) int {
	limit := 1
	if g == nil || g.Battlefield == nil || CatalogPowerUpExtras == nil {
		return limit
	}
	for i := range g.Battlefield.Cards {
		grantor := &g.Battlefield.Cards[i]
		for _, x := range powerUpExtrasOf(grantor) {
			if x.Times > 0 && x.Applies != nil && x.Applies(g, source, *grantor) {
				limit += x.Times
			}
		}
	}
	return limit
}

// PowerUpUsed reports whether `ab` is a power-up ability that the object
// `source` names has already been activated as many times as it may:
// "Activate this ability only once" (CR 702.193a), or once more for each
// "an additional time" static in play (PowerUpLimitFor). False for every
// ability without the bit.
//
// One function and three readers, as AbilityExhausted is:
// ActivateCatalogAbility refuses with ErrPowerUpUsed, internal/legal
// does not offer the move, and the view stamps `power_up_used`.
//
// Caller must hold g.mu.
func (g *Game) PowerUpUsed(source *Card, ab ActivatedAbilityShape) bool {
	if !ab.PowerUp || source == nil {
		return false
	}
	used := g.ActivatedThisGame(source.InstanceID, ab.Label)
	if used == 0 {
		// The walk for a static happens only once the ability has been
		// used at all.
		return false
	}
	return used >= g.PowerUpLimitFor(*source)
}

// powerUpBanLocked is the clause of a live ModCantActivatePowerUp
// ScopedEffect and the permanent that made it, for the activation gate
// (Kang the Conqueror's "During that turn, power-up abilities can't be
// activated"). ok is false when nothing bans them.
//
// Caller must hold g.mu.
func (g *Game) powerUpBanLocked() (reason string, source uuid.UUID, ok bool) {
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		if !scopedEffectHasMod(e, ModCantActivatePowerUp) {
			continue
		}
		reason = "Power-up abilities can't be activated"
		if e.SourceName != "" {
			reason += " — " + e.SourceName
		}
		return reason, e.Source.ID, true
	}
	return "", uuid.Nil, false
}

// PowerUpActivationsThisTurn is how many power-up abilities `player`
// has activated this turn. Zero means the next one is the turn's first.
//
// Caller must hold g.mu.
func (g *Game) PowerUpActivationsThisTurn(player uuid.UUID) int {
	return g.TurnTallyFor(player).PowerUpActivations
}

// powerUpReductionLocked is the mana cost a power-up activation's cost
// is reduced by: its source's mana cost while the source entered the
// battlefield this turn (CR 702.193a), and the zero cost otherwise.
//
// "This permanent's mana cost" is the source's mana cost as it is now
// (Card.ManaCost: a transformed back face has none, and a face-down
// permanent has none, CR 708.2). {X} in it is 0 (CR 107.3g), so it
// reduces nothing. An unreadable mana cost refuses the activation
// rather than pricing it at the printed cost, #289's lesson.
//
// Caller must hold g.mu.
func (g *Game) powerUpReductionLocked(q CostQuery) (ParsedCost, error) {
	if q.Ability == nil || !q.Ability.PowerUp || q.Card.FaceDown {
		return ParsedCost{}, nil
	}
	if !g.EnteredThisTurn(q.Card.InstanceID) {
		return ParsedCost{}, nil
	}
	by, err := ParseCost(q.Card.ManaCost)
	if err != nil {
		return ParsedCost{}, fmt.Errorf("%w: the power-up reduction can't read %s's mana cost %q",
			ErrCostModifier, q.Card.Name, q.Card.ManaCost)
	}
	return by, nil
}

// reduceByManaCost reduces `cost` by the mana cost `by`, as CR 702.193b
// and 118.7b–g say:
//
//   - generic mana in `by` reduces generic mana (118.7a);
//   - a coloured or colourless symbol reduces one mana of the same type,
//     and when the cost has none of that type left it reduces one
//     generic instead (118.7b–d). A Phyrexian symbol is one mana of its
//     colour (118.7f); snow is generic (118.7g);
//   - for a hybrid symbol the payer chooses a half (118.7e). The engine
//     chooses for them, and chooses the half that takes a coloured
//     symbol off the cost when one matches; otherwise the symbol
//     reduces generic mana (one, or N for a {N/C} symbol).
//
// A symbol "of the same type" is a requirement of the cost that type of
// mana can pay. An exact single-colour requirement is taken before a
// hybrid one, so {R} off {R/G}{R} takes the {R} and leaves the hybrid,
// which either colour can still pay.
//
// What no symbol can take comes off the generic mana, floored at zero
// (reduceGeneric, which also lets it reach the mana announced for {X},
// #2701). It never goes below {0} (CR 601.2f).
func reduceByManaCost(cost ParsedCost, by ParsedCost) ParsedCost {
	if by.Generic == 0 && len(by.Required) == 0 {
		return cost
	}
	out := cost
	req := slices.Clone(cost.Required)
	generic := by.Generic
	var hybrids []ColorRequirement
	for _, r := range by.Required {
		switch {
		case r.Snow:
			generic++
		case len(r.Options) == 1 && !r.HasNumericAlt:
			var ok bool
			if req, ok = takeRequirement(req, r.Options[0]); !ok {
				generic++
			}
		default:
			hybrids = append(hybrids, r)
		}
	}
	// Hybrids last, so a half is chosen against what the single-colour
	// symbols left.
	for _, r := range hybrids {
		took := false
		for _, color := range r.Options {
			var ok bool
			if req, ok = takeRequirement(req, color); ok {
				took = true
				break
			}
		}
		if took {
			continue
		}
		if r.HasNumericAlt {
			generic += r.NumericAlt
		} else {
			generic++
		}
	}
	out.Required = req
	return reduceGeneric(out, generic)
}

// takeRequirement removes one requirement of `req` that mana of `color`
// pays, an exact single-colour one first, and reports whether there was
// one. `req` is the caller's own copy.
func takeRequirement(req []ColorRequirement, color string) ([]ColorRequirement, bool) {
	exact := -1
	loose := -1
	for i, r := range req {
		if !slices.Contains(r.Options, color) {
			continue
		}
		if len(r.Options) == 1 && !r.HasNumericAlt {
			exact = i
			break
		}
		if loose < 0 {
			loose = i
		}
	}
	i := exact
	if i < 0 {
		i = loose
	}
	if i < 0 {
		return req, false
	}
	return slices.Delete(req, i, i+1), true
}

// zeroAlternativeApplies reports whether a CostZeroAlternative modifier
// replaces `cost` for this activation: an activation, a cost with no
// {X} (CR 107.3b — see CostZeroAlternative), and the modifier's own
// predicate.
func zeroAlternativeApplies(bm boundCostModifier, q CostQuery, cost ParsedCost) bool {
	if q.Ability == nil || cost.XSlots > 0 {
		return false
	}
	q.Source = bm.source
	q.Cost = cost
	if !bm.modifier.ReadsTargets {
		q.Targets = nil
	}
	return bm.modifier.AppliesTo == nil || bm.modifier.AppliesTo(q)
}
