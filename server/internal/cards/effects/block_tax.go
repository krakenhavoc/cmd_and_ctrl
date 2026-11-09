package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// block_tax.go — constructors for Spec.BlockTaxes, the CR 509.1d block
// tax: "creatures can't block unless their controller pays {1} for each
// of those creatures". The engine half is game/block_tax.go (ADR 0221,
// #5375); attack_tax.go is the CR 508.1h sibling this mirrors.
//
// The shape to copy:
//
//	BlockTaxes: []game.BlockTax{
//	    BlockTaxWhileAttacking("{1}", "As long as this creature is attacking, creatures can't block unless their controller pays {1} for each of those creatures."),
//	},
//
// Unlike an attack tax there is no "you": the printed clause taxes
// every creature that blocks, whoever controls it.

// BlockTax is the unconditional clause: every blocking creature costs
// `mana`, a cost string for ONE blocker, concatenated per blocker.
func BlockTax(mana, label string) game.BlockTax {
	return game.BlockTax{
		Label:    label,
		ManaCost: func(game.BlockTaxQuery) string { return mana },
	}
}

// BlockTaxWhileAttacking is BlockTax gated on "as long as this creature
// is attacking" — Archangel of Tithes. Judged when the price is
// determined, as the defender's declaration completes (CR 509.1d).
func BlockTaxWhileAttacking(mana, label string) game.BlockTax {
	t := BlockTax(mana, label)
	t.AppliesTo = game.WhileSourceAttacking
	return t
}
