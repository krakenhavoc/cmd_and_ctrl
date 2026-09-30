package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sonic Screwdriver — Artifact {3} (Doctor Who):
//
//	"{T}: Add one mana of any color.
//	 {1}, {T}: Untap another target artifact.
//	 {2}, {T}: Scry 1.
//	 {3}, {T}: Target creature can't be blocked this turn."
//
// Four abilities sharing one {T}, so only one can be activated a
// turn. The mana ability is a pipe-syntax any-color slot with a
// picker (Birds of Paradise's shape); the three CR 602 abilities are
// Voltaic Key's untap, Sensei's Divining Top-style scry, and Rogue's
// Passage's can't-be-blocked, each with its own mana-plus-tap cost.
//
// "Another" on the untap ability is object identity (effects.Another,
// CR 109.1): Sonic Screwdriver can untap a second Sonic Screwdriver.
func init() {
	Register(Spec{
		OracleID:     "cd68cc31-12fd-48ff-b37b-bccfe4172974",
		Name:         "Sonic Screwdriver",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []ActivatedAbility{
			{
				Label:   "{1}, {T}: Untap another target artifact.",
				Cost:    Plus(ManaCost("{1}"), TapCost()),
				Targets: Another(TargetPermanent("another target artifact", Artifact())),
				Effect:  untapFirstLegalTarget,
			},
			{
				Label: "{2}, {T}: Scry 1.",
				Cost:  Plus(ManaCost("{2}"), TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return Scry{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "{3}, {T}: Target creature can't be blocked this turn.",
				Cost:    Plus(ManaCost("{3}"), TapCost()),
				Targets: TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return RestrictUntilEOT{
						Target:       id,
						Restrictions: game.CantBeBlocked,
						Label:        "Sonic Screwdriver — can't be blocked",
					}.Apply(ctx)
				},
			},
		},
	})
}
