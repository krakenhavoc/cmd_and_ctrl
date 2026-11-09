package game

// power_floor.go is #4678: "This creature can't attack or block unless
// its power is 6 or greater" (Technodrome).
//
// Bast, Panther Goddess and Hazoret, Godseeker print the same "can't
// attack or block unless …" and write CantAttack | CantBlock from a
// layer-6 static whose condition reads the board. That cannot work for a
// condition on the creature's own POWER: layer 6 runs before layer 7
// (CR 613.1f, 613.4), so a layer-6 static would read the power before
// any anthem, counter or -N/-N had applied, and a creature pumped to 6
// would still be refused.
//
// So the card's static only records the floor
// (Characteristic.AttackBlockPowerFloor), and this fold turns it into the
// two restriction bits once every layer has run. CR 613.11 places an
// effect on the game rules after every other continuous effect, and a
// restriction has no layer at all (restrictions.go), so reading the
// finished power here is the rules' own order. From the bits on,
// nothing is new: AttackerEligible, BlockPairRefusalLocked, the legal
// enumerator, the bot and the view's restriction chips all read
// Restrictions, so they agree on the declaration (CR 508.1c, 509.1b)
// without a second copy of the rule (#544).
//
// The bits are recomputed on every pass, and a counter placed or
// removed, or a pump that begins or ends, bumps the layer version, so
// the next read sees the new power. As with every restriction, it is
// checked at declaration only: a Technodrome that shrinks below 6 after
// it was declared attacking keeps attacking (CR 506.4).

// foldPowerFloorLocked adds CantAttack | CantBlock to every permanent
// whose finished power (layers and P/T counters, unclamped, as
// PowerForComparison reads it) is below its AttackBlockPowerFloor, and
// to every permanent with an AttackBlockNeedsOtherPower whose controller
// controls no other creature with that much power (Tiger-Dillo).
//
// Caller must hold g.mu in write mode (the layer pass).
func (g *Game) foldPowerFloorLocked() {
	if g.Battlefield == nil {
		return
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.effective == nil {
			continue
		}
		if n := c.effective.AttackBlockNeedsOtherPower; n > 0 && !g.controlsAnotherCreatureWithPowerLocked(c, n) {
			c.effective.Restrictions |= CantAttackOrBlock
		}
		if c.effective.AttackBlockPowerFloor <= 0 && c.effective.AttackBlockPowerCeiling <= 0 {
			continue
		}
		// Counters are not in effective.Power: they are added at read
		// time, as every power read does (ptWithCounters).
		power, _ := ptWithCounters(*c.effective, c.Counters)
		if f := c.effective.AttackBlockPowerFloor; f > 0 && power < f {
			c.effective.Restrictions |= CantAttackOrBlock
		}
		// Hedron Fields of Agadeem: "creatures with power N or greater
		// can't attack or block".
		if n := c.effective.AttackBlockPowerCeiling; n > 0 && power >= n {
			c.effective.Restrictions |= CantAttackOrBlock
		}
	}
}

// controlsAnotherCreatureWithPowerLocked is Tiger-Dillo's condition:
// does c's controller control a creature other than c whose finished
// power (layers and P/T counters, as foldPowerFloorLocked reads it) is
// n or greater? The fold only writes restriction bits, never power or
// types, so every other permanent's finished values are already final
// when it runs, whatever the order.
//
// Caller must hold g.mu in write mode (the layer pass).
func (g *Game) controlsAnotherCreatureWithPowerLocked(c *Card, n int) bool {
	for i := range g.Battlefield.Cards {
		o := &g.Battlefield.Cards[i]
		if o.InstanceID == c.InstanceID || o.Controller != c.Controller || o.effective == nil ||
			!typeListHas(o.effective.Types, "Creature") {
			continue
		}
		if power, _ := ptWithCounters(*o.effective, o.Counters); power >= n {
			return true
		}
	}
	return false
}
