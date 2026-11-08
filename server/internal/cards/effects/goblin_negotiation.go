package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Negotiation — Sorcery {X}{R}{R} (Foundations, #4465):
//
//	"Goblin Negotiation deals X damage to target creature. Create a
//	 number of 1/1 red Goblin creature tokens equal to the amount of
//	 excess damage dealt to that creature this way."
//
// Hell to Pay with Goblins for Treasures. X rides the cast; the damage
// goes through b27DealDamageWithExcess, which reads the creature's
// lethal requirement (toughness less damage already marked, CR 120.4a)
// before the damage lands and the amount actually dealt afterwards, so a
// prevention shield lowers the excess with the damage. A target gone by
// resolution takes nothing and makes no Goblins (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "79824266-bffa-476c-bdd3-976e9d789a42",
		Name:         "Goblin Negotiation",
		XMatters:     true,
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				excess, err := b27DealDamageWithExcess(ctx, t.ID, ctx.X())
				if err != nil || excess <= 0 {
					return err
				}
				return CreateToken{Controller: item.Controller, Template: RedGoblinToken(), N: excess}.Apply(ctx)
			}
			return nil
		},
	})
}
