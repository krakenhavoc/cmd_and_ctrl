package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Inside Information — Sorcery {X}{B}{B} (The Hobbit, #4454):
//
//	"Exile the top X cards of target opponent's library. You may play
//	 those cards this turn. If you cast a spell this way, pay life equal
//	 to its mana value rather than pay its mana cost."
//
// An impulse exile of ANOTHER player's library (Game.ExileTopWithPermission-
// ForEffect: the library it reads and the player it grants to are
// separate), whose permission ends with the turn (the zero Duration) and
// carries Bolas's Citadel's price: LifeEqualToManaValue (CR 118.9, 119.4),
// the same alternative cost Valgavoth, Terror Eater's permission claims. It
// is a cost, so a player without the life cannot claim it; it replaces the
// mana cost, so the printed cost is not on offer; and a land, which is
// played rather than cast, has no cost to replace. The cards are the
// opponent's and the caster plays them. X = 0 exiles nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "66961bcb-f588-47a7-b3f3-12f4ac674477",
		Name:         "Inside Information",
		XMatters:     true,
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetPlayer {
					continue
				}
				_, err := ctx.Game.ExileTopWithPermissionForEffect(t.ID, item.Controller, ctx.X(), game.CastPermission{
					AltCostKey:           "inside_information",
					LifeEqualToManaValue: true,
					Label:                "Pay life equal to its mana value (Inside Information)",
				})
				return err
			}
			return nil
		},
	})
}
