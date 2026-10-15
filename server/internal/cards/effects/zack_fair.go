package effects

import (
	"sort"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zack Fair — Legendary Creature — Human Soldier {W}, 0/1 (#3656):
//
//	"Zack Fair enters with a +1/+1 counter on it.
//	 {1}, Sacrifice Zack Fair: Target creature you control gains
//	 indestructible until end of turn. Put Zack Fair's counters on that
//	 creature and attach an Equipment that was attached to Zack Fair to
//	 that creature."
//
// The entry counter is the self-replacement every "enters with" creature
// uses. The ability reads Zack's counters from the sacrificed
// permanent's last-known information (Context.SacrificedPermanent), so
// the counters he had when he was sacrificed move over, and a target
// that left in response gets nothing (CR 608.2b).
//
// Simplification, declared: the engine keeps no record of which
// Equipment was attached to a permanent when it was sacrificed (state
// based actions unattach it before the ability resolves), so the
// Equipment is not moved to the target.
func init() {
	Register(Spec{
		OracleID:     "e84d146a-1c67-49df-b062-622794e45c42",
		Name:         "Zack Fair",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The Equipment that was attached to Zack Fair isn't moved to the target; attach it yourself.",
		},
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters("+1/+1", 1, "Zack Fair enters with a +1/+1 counter on it."),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice Zack Fair: Target creature you control gains indestructible until end of turn. Put Zack Fair's counters on that creature and attach an Equipment that was attached to Zack Fair to that creature.",
			Cost:    Plus(ManaCost("{1}"), SacrificeThis()),
			Targets: TargetCreature("target creature you control", YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				target, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if err := (GrantKeywordUntilEOT{Target: target, Keywords: []string{"indestructible"}, Label: "Zack Fair — indestructible"}).Apply(ctx); err != nil {
					return err
				}
				zack, ok := ctx.SacrificedPermanent()
				if !ok {
					return nil
				}
				kinds := make([]string, 0, len(zack.Counters))
				for kind, n := range zack.Counters {
					if n > 0 {
						kinds = append(kinds, kind)
					}
				}
				sort.Strings(kinds)
				for _, kind := range kinds {
					if err := (AddCounter{Target: target, Kind: kind, N: zack.Counters[kind]}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
