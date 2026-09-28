package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grand Melee — Enchantment, {3}{R}:
//
//	"All creatures attack each combat if able.
//	 All creatures block each combat if able."
//
// Both halves are enforced. The attack half is #1599's
// AttacksEachCombatWhere; the block half is #1597's
// BlocksEachCombatWhere (CR 509.1c), which puts one "blocks each combat
// if able" requirement on every creature, so a defending player's pass
// is refused while any untapped creature of theirs could still legally
// block something. Neither predicate scopes anything — "ALL creatures",
// no "you control", no "other" — and both are attributed to Grand
// Melee itself, so a refusal names this card.
//
// No simplification.
func init() {
	allCreatures := func(_ *game.Card, _ *game.Game, _ *game.Card) bool { return true }
	Register(Spec{
		OracleID:     "1accf98a-0905-4a9d-9ab3-72e9f853f4ab",
		Name:         "Grand Melee",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			AttacksEachCombatWhere(allCreatures),
			BlocksEachCombatWhere(allCreatures),
		},
	})
}
