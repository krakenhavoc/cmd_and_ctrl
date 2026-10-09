package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lion Sash — Artifact Creature — Equipment Cat {1}{W}, 1/1:
//
//	"{W}: Exile target card from a graveyard. If it was a permanent card,
//	 put a +1/+1 counter on this permanent.
//	 Equipped creature gets +1/+1 for each +1/+1 counter on this
//	 Equipment.
//	 Reconfigure {2}"
//
// The exile clause is Cling to Dust's ExileThenIfItWas: the card's type
// is read before it moves, and the counter waits for the exile to
// happen. The counter goes on the Sash only while it is still the same
// object on the battlefield. While the Sash fights on its own its
// counters grow it directly; while attached they pump the host
// (PumpAttachedPer). Reconfigure is reconfigure.go (#2639).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "50330e48-db74-4c5f-a0bc-f9a8607e8f31",
		Name:         "Lion Sash",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttachedPer(1, 1, func(_ *game.Game, source *game.Card) int {
				return source.Counters["+1/+1"]
			}),
		},
		Activated: append([]ActivatedAbility{{
			Label:   "{W}: Exile target card from a graveyard. If it was a permanent card, put a +1/+1 counter on this permanent.",
			Cost:    ManaCost("{W}"),
			Targets: TargetCardInGraveyard("target card from a graveyard"),
			Effect:  lionSashExile,
		}}, Reconfigure("{2}")...),
	})
}

func lionSashExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return ExileThenIfItWas{
			Target: t.ID,
			Was:    func(c game.Card) bool { return c.IsPermanent() },
			Then: func(ctx *Context) error {
				if !onBattlefield(ctx.Game, ctx.Source()) {
					return nil
				}
				return AddCounter{Target: ctx.Source(), Kind: "+1/+1", N: 1}.Apply(ctx)
			},
		}.Apply(ctx)
	}
	return nil
}
