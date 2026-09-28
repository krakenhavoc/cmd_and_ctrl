package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grand Melee — Enchantment, {3}{R}:
//
//	"All creatures attack each combat if able.
//	 All creatures block each combat if able."
//
// #1599: the attack half only. The block half needs the
// block-requirement seam (#1597, CR 509.1c "must block if able"),
// which the engine does not have yet — #1595 built attack
// requirements, not block ones, and #1597 is untouched. Building the
// attack half alone ships a card that is weaker than printed and
// declares exactly the missing line, never one that is stronger.
//
// The attack half is AttacksEachCombatWhere with no scoping predicate
// at all — "ALL creatures", no "you control", no "other" — attributed
// to Grand Melee itself, so a table full of forced attackers still
// names this card in the refusal sentence.
func init() {
	Register(Spec{
		OracleID:     "1accf98a-0905-4a9d-9ab3-72e9f853f4ab",
		Name:         "Grand Melee",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Creatures aren't forced to block — only the attack requirement is enforced."},
		Static: []game.StaticAbility{
			AttacksEachCombatWhere(func(_ *game.Card, _ *game.Game, _ *game.Card) bool { return true }),
		},
	})
}
