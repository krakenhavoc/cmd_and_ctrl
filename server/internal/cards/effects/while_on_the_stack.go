package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// while_on_the_stack.go — #6222: abilities of a spell that function
// while it is on the stack (CR 113.6b, "As long as <this> is on the
// stack, …").
//
//   - A cost modifier: Kaervek's Torch's "spells that target it cost {2}
//     more to cast" (SpellsThatTargetThisCostMore), gathered from the
//     stack by the CR 601.2f pass (game/stack_cost_modifiers.go).
//   - Any other static: declare Zones: []game.ZoneKind{game.ZoneStack}
//     on the game.StaticAbility (Torrent of Lava's grant to each
//     creature), gathered by the layer pass (game/static_zones.go). The
//     source is the spell, and "you" is its controller.
//
// Append-only: add a builder, never change what one means.

// WhileOnTheStack marks a cost modifier as one that functions while its
// card is a spell on the stack, and only then.
func WhileOnTheStack(m game.CostModifier) game.CostModifier {
	m.OnStack = true
	return m
}

// SpellsThatTargetThisCostMore is "As long as <this> is on the stack,
// spells that target it cost {n} more to cast" (Kaervek's Torch). Every
// spell, whoever casts it; an ability that targets it is not a spell.
func SpellsThatTargetThisCostMore(n int, label string) game.CostModifier {
	m := WhileOnTheStack(CostsMore(n, label, ThatTargetThisPermanent()))
	m.ReadsTargets = true
	return m
}
