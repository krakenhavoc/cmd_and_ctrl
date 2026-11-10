package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rhonas the Indomitable — Legendary Creature — God {2}{G}, 5/5:
//
//	"Deathtouch, indestructible
//	 Rhonas can't attack or block unless you control another creature with
//	 power 4 or greater.
//	 {2}{G}: Another target creature gets +2/+0 and gains trample until
//	 end of turn."
//
// The restriction is Tiger-Dillo's fold (game/power_floor.go): the static
// records the threshold and the engine applies it once every layer has run
// (CR 613.11), so it is checked at declaration only and a creature that has
// attacked or blocked stays in combat. The ability is Oliphaunt's pump,
// aimed at ANOTHER creature by object identity.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5d2acd3e-4782-42e3-bcd5-80f8b439cc28",
		Name:            "Rhonas the Indomitable",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch", "indestructible"},
		Static: []game.StaticAbility{{
			Label:     "Rhonas can't attack or block unless you control another creature with power 4 or greater.",
			Layer:     game.Layer6Ability,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				if c.AttackBlockNeedsOtherPower < 4 {
					c.AttackBlockNeedsOtherPower = 4
				}
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}{G}: Another target creature gets +2/+0 and gains trample until end of turn.",
			Cost:    ManaCost("{2}{G}"),
			Targets: Another(TargetCreature("another target creature")),
			Purpose: ForTargets(BenefitUntilEOT(0, 2, 0, "trample")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (BoostUntilEOT{Target: t.ID, Power: 2, Label: "Rhonas the Indomitable — +2/+0"}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{
						Target:   t.ID,
						Keywords: []string{"trample"},
						Label:    "Rhonas the Indomitable — trample",
					}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
