package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Impact Tremors — Enchantment {1}{R} (EDHREC rank 197):
//
//	"Whenever a creature you control enters, this enchantment deals
//	 1 damage to each opponent."
//
// The go-wide payoff: every token, every creature spell, every
// reanimation pings the table. Fires once per creature (the printed
// text is "a creature", not "one or more"), so a Krenko activation
// that makes five Goblins triggers five times — which is the paper
// behaviour and the reason the card is played.
//
// The damage's source is the enchantment itself, so a damage
// doubler (Solphim) or a prevention effect sees it as noncombat
// damage from a red permanent, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9242cd3e-1a71-4700-8182-9c1005616033",
		Name:         "Impact Tremors",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, CreatureEnteredUnderYourControl, "Impact Tremors — 1 damage to each opponent", func(g *game.Game, item *game.StackItem) error {
				return damageToEachOpponent(g, item, 1)
			}),
		},
	})
}
