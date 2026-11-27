package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// scoped_combat_tax.go — #6191: attack and block taxes scoped to the
// taxed creature, priced per blocked attacker, or paid in life
// (CR 508.1h, CR 509.1d).
//
//   - Myr Prototype's "can't attack or block unless you pay {1} for each
//     +1/+1 counter on it", Phyrexian Marauder's attack half: a self
//     attack tax (AttackTax.OnAttacker, #6190) and a self block tax
//     (game.BlockedBySelf), each priced from the creature's counters.
//   - Hipparion's "can't block creatures with power 3 or greater unless
//     you pay {1}" and Awesome Presence's "enchanted creature can't be
//     blocked unless defending player pays {3} for each creature they
//     control that's blocking it": game.BlockTax.PerAttacker, a price per
//     (blocker, attacker) pair.
//   - Heat Wave's "Nonblue creatures can't block creatures you control
//     unless their controller pays 1 life for each blocking creature
//     they control": game.BlockTax.LifeCost, judged on what the creature
//     blocks (BlockTaxQuery.Blocking) and charged once per creature.

// GenericMana is "{n}", or "" (no tax) for n <= 0.
func GenericMana(n int) string {
	if n <= 0 {
		return ""
	}
	return "{" + strconv.Itoa(n) + "}"
}

// CantAttackUnlessYouPayPerCounter is "This creature can't attack unless
// you pay {1} for each <kind> counter on it": owed by its controller
// whatever it attacks, {N} for the N counters on it as attackers are
// declared (CR 508.1h). No counters, no tax.
func CantAttackUnlessYouPayPerCounter(kind, label string) game.AttackTax {
	return game.AttackTax{
		Label:      label,
		OnAttacker: true,
		AppliesTo:  SelfAttacker,
		ManaCost: func(q game.AttackTaxQuery) string {
			return GenericMana(q.Attacker.Counters[kind])
		},
	}
}

// CantBlockUnlessYouPayPerCounter is the block half: "This creature
// can't block unless you pay {1} for each <kind> counter on it"
// (CR 509.1d).
func CantBlockUnlessYouPayPerCounter(kind, label string) game.BlockTax {
	return game.BlockTax{
		Label:     label,
		AppliesTo: game.BlockedBySelf,
		ManaCost: func(q game.BlockTaxQuery) string {
			return GenericMana(q.Blocker.Counters[kind])
		},
	}
}

// CantBlockAttackersUnlessYouPay is "This creature can't block <attackers
// matching attacks> unless you pay <cost>" (Hipparion: creatures with
// power 3 or greater, {1}). Priced per attacker blocked
// (BlockTax.PerAttacker): blocking a 2/2 is free, a 3/3 costs the price.
func CantBlockAttackersUnlessYouPay(cost string, attacks func(attacker game.Card) bool, label string) game.BlockTax {
	return game.BlockTax{
		Label:       label,
		PerAttacker: true,
		AppliesTo: func(q game.BlockTaxQuery) bool {
			return game.BlockedBySelf(q) && attacks(q.Attacker)
		},
		ManaCost: func(game.BlockTaxQuery) string { return cost },
	}
}

// PowerAtLeast is "creatures with power N or greater", read off the
// creature's effective power.
func PowerAtLeast(n int) func(c game.Card) bool {
	return func(c game.Card) bool { return c.Effective().Power >= n }
}

// EnchantedCantBeBlockedUnlessPay is "Enchanted creature can't be blocked
// unless defending player pays <cost> for each creature they control
// that's blocking it" (Awesome Presence): every creature blocking the
// enchanted creature costs its controller the price.
func EnchantedCantBeBlockedUnlessPay(cost, label string) game.BlockTax {
	return game.BlockTax{
		Label:       label,
		PerAttacker: true,
		AppliesTo: func(q game.BlockTaxQuery) bool {
			return q.Source.IsAttachedTo(q.Attacker.InstanceID)
		},
		ManaCost: func(game.BlockTaxQuery) string { return cost },
	}
}

// CantBlockYourCreaturesUnlessPayLife is "<blockers matching blockers>
// can't block creatures you control unless their controller pays <life>
// life for each blocking creature they control" (Heat Wave's nonblue
// half): a creature blocking at least one creature the source's
// controller controls costs its controller the life, once however many
// it blocks (CR 509.1d, CR 119.4).
func CantBlockYourCreaturesUnlessPayLife(life int, blockers func(blocker game.Card) bool, label string) game.BlockTax {
	return game.BlockTax{
		Label: label,
		AppliesTo: func(q game.BlockTaxQuery) bool {
			if !blockers(q.Blocker) {
				return false
			}
			for _, a := range q.Blocking {
				if a.Controller == q.Source.Controller {
					return true
				}
			}
			return false
		},
		LifeCost: func(game.BlockTaxQuery) int { return life },
	}
}
