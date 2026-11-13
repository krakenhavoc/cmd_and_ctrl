package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glissa Sunseeker — Legendary Creature — Elf Warrior {2}{G}{G}, 3/2
// (Mirrodin, #5606):
//
//	"First strike
//	 {T}: Destroy target artifact if its mana value is equal to the
//	 amount of unspent mana you have."
//
// Any artifact may be targeted; the comparison is made as the ability
// resolves. "Unspent mana you have" is every mana in the activator's
// pool at that moment (CR 106.4), of any colour or none, and the
// artifact's mana value is its mana cost's (CR 202.3). The mana is only
// counted, never spent, and empties at the end of the step as usual. A
// target that is no longer legal is left alone (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "bf4b9330-72df-426f-820c-77e37f7e9130",
		Name:            "Glissa Sunseeker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Destroy target artifact if its mana value is equal to the amount of unspent mana you have.",
			Cost:    TapCost(),
			Targets: TargetPermanent("target artifact", Artifact()),
			Effect:  glissaSunseekerDestroy,
		}},
	})
}

// glissaSunseekerDestroy destroys the target artifact when its mana value
// equals the number of mana in the controller's pool as the ability
// resolves.
func glissaSunseekerDestroy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	artifact, ok := g.LookupCardForEffect(id)
	p := g.PlayerByIDForEffect(item.Controller)
	if !ok || p == nil || artifact.ManaValue() != p.ManaPool.Total() {
		return nil
	}
	return DestroyTarget{Target: id}.Apply(ctx)
}
