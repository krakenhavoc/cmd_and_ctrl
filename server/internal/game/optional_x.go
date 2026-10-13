package game

// optional_x.go — a kicker (or any optional cost) whose mana half is
// {X} (#3232, ADR 0073's amendment of 2026-10-13).
//
//	Thieving Skydiver  Kicker {X}. X can't be 0.
//	Verdeloth          Kicker {X}.
//	Kangee             Kicker {X}{2}.
//
// CR 107.3a and 601.2b make X one number the caster announces; CR
// 601.2f adds the kicker's {X} to the total cost, so the price was
// already right (AdditionalCostMana appends the XSlots and the pool
// multiplies the announced X into them). Three things were not:
//
//   - "X can't be 0" had a home on an activated ability
//     (AbilityCost.MinX) and none on a spell's optional cost. It is
//     AdditionalCost.MinX, and the announce gate refuses a kicked cast
//     whose X is under it.
//   - The X belongs to the KICKER. A caster who does not kick has no
//     {X} in the cost they pay, so X is 0 and nothing is charged for
//     it; an unkicked announcement of X=3 is refused rather than
//     recorded on the stack item, where Card.CastX would carry a number
//     nothing paid for.
//   - The two readers that must agree with the gate (the bot
//     enumerator's X floor and the view's `x_min`) ask the same
//     functions.

// OptionalCostHasX reports whether paying `c` charges {X}: its mana
// half has an {X} slot. A cost whose mana string does not parse has
// none (Register refuses that string at boot).
func OptionalCostHasX(c AdditionalCost) bool {
	if c.ManaCost == "" {
		return false
	}
	pc, err := ParseCost(c.ManaCost)
	return err == nil && pc.XSlots > 0
}

// OptionalCostsOfferX reports whether any of a card's optional costs
// charges {X}.
func OptionalCostsOfferX(optional []AdditionalCost) bool {
	for _, c := range optional {
		if OptionalCostHasX(c) {
			return true
		}
	}
	return false
}

// OptionalCostsPayX reports whether the announced optional costs
// include one that charges {X}.
func OptionalCostsPayX(optional []AdditionalCost, chosen []int) bool {
	for _, i := range chosen {
		if i >= 0 && i < len(optional) && OptionalCostHasX(optional[i]) {
			return true
		}
	}
	return false
}

// OptionalCostsXFloor is the smallest X the announced optional costs
// allow: the largest MinX among the costs chosen, 0 when none names a
// floor. Pure, so the announce gate, the bot enumerator and the view
// quote one number.
func OptionalCostsXFloor(optional []AdditionalCost, chosen []int) int {
	floor := 0
	for _, i := range chosen {
		if i >= 0 && i < len(optional) && optional[i].MinX > floor {
			floor = optional[i].MinX
		}
	}
	return floor
}

// validateOptionalCostX is the announce-time rule for an optional
// cost's X (#3232), judged after validateOptionalCostChoice has
// accepted the indices:
//
//   - a chosen cost with a floor refuses an X under it ("X can't be 0");
//   - when the card offers an {X} optional cost, none is chosen and the
//     cost actually being paid (`paidHasX`: the printed cost or the
//     claimed alternative cost) has no {X} of its own, the only legal
//     announcement is 0 — there is nothing for X to be.
//
// `otherXSource` is true when some other component prices X (a "pay X
// life" or "blight X" additional cost, a waterbend {X}); the second
// rule stands down then, because the X is meaningful without the kicker.
func validateOptionalCostX(optional []AdditionalCost, chosen []int, x int, paidHasX, otherXSource bool) error {
	if floor := OptionalCostsXFloor(optional, chosen); x < floor {
		return ErrInvalidParam
	}
	if x != 0 && !paidHasX && !otherXSource && OptionalCostsOfferX(optional) && !OptionalCostsPayX(optional, chosen) {
		return ErrInvalidParam
	}
	return nil
}
