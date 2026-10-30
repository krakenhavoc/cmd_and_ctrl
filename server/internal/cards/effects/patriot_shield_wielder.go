package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Patriot, Shield Wielder — Legendary Creature — Human Hero {1}{W}, 2/2
// (Marvel Super Heroes, #4455):
//
//	"{2}, {T}: Another target creature you control gets +2/+0 and gains
//	 hexproof until end of turn. (It can't be the target of spells or
//	 abilities your opponents control.)"
//
// Ancestors' Aid's pump and keyword grant on a targeted activated ability
// whose clause excludes Patriot himself (Another, CR 109.1). Both halves
// are keyed to the target chosen on announcement.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1df7a09f-fbc8-4c2a-a903-bab9e680eda7",
		Name:         "Patriot, Shield Wielder",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Another target creature you control gets +2/+0 and gains hexproof until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 2, 0, "hexproof")),
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Targets: Another(TargetCreature("another target creature you control", YouControl())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (BoostUntilEOT{Target: t.ID, Power: 2, Label: "Patriot — +2/+0 until end of turn"}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"hexproof"}, Label: "Patriot — hexproof until end of turn"}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
