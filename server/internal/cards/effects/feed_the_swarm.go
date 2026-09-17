package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Feed the Swarm — Sorcery {1}{B}:
//
//	"Destroy target creature or enchantment an opponent controls. You
//	lose life equal to that permanent's mana value."
//
// Black's only mono-coloured enchantment removal at a playable rate,
// which is the whole reason it is a staple rather than a draft
// common.
//
// The life is paid whether or not the destroy sticks — but the mana
// value has to be read from the permanent BEFORE it is destroyed, or
// the lookup finds a graveyard card whose printed cost happens to
// still be right and whose absence would otherwise have to be
// special-cased. Reading first also makes the fizzle case obvious: if
// the target left in response the spell has no legal target at all
// and OnResolve is never reached (CR 608.2b, single-target fizzle),
// so no life is lost.
func init() {
	Register(Spec{
		OracleID:     "5825997b-10d7-4a36-972c-a80ddd90b8ed",
		Name:         "Feed the Swarm",
		Completeness: CompletenessFull,
		Targets: TargetPermanent("target creature or enchantment an opponent controls",
			Or(Creature(), Enchantment()), OpponentControls()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			card, ok := ctx.Game.LookupCardForEffect(target)
			if !ok {
				return nil
			}
			cost := card.ManaValue()
			if err := (DestroyTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -cost)
		},
	})
}
