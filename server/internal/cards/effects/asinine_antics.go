package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Asinine Antics — Sorcery {2}{U}{U}:
//
//	"You may cast this spell as though it had flash if you pay {2} more
//	 to cast it.
//	 For each creature your opponents control, create a Cursed Role
//	 token attached to that creature. (If you control another Role on
//	 it, put that one into the graveyard. Enchanted creature is 1/1.)"
//
// The creature set is read once, at resolution: every creature an
// opponent controls then. A Cursed Role is a base-P/T set to 1/1, so
// each one the caster controls replaces the caster's earlier Role on
// that creature (CR 704.5z).
//
// Simplification: the "as though it had flash if you pay {2} more"
// permission has no shape in the cast-timing engine (an optional
// surcharge that unlocks instant timing), so the spell is
// sorcery-speed only. Weaker than printed, never stronger.
func init() {
	Register(Spec{
		OracleID:     "0a0a051c-59ed-4286-b842-a34da7ef15d5",
		Name:         "Asinine Antics",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"It can't be cast as though it had flash for {2} more — it can only be cast when you could cast a sorcery."},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			var hosts []game.Card
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.IsCreature() && c.Controller != item.Controller {
					hosts = append(hosts, c)
				}
			}
			for _, c := range hosts {
				if err := (CreateRoleToken{Role: RoleCursed, Host: c.InstanceID}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
