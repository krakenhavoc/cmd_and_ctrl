package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pia, Determined Rebuilder — Legendary Creature — Human Artificer
// {2}{R}, 2/2:
//
//	"When Pia enters, create a 1/1 colorless Thopter artifact creature
//	 token with flying.
//	 {5}{R}: Target creature gets +X/+0 until end of turn, where X is
//	 the number of artifacts you control."
//
// X is counted as the ability resolves, so an artifact destroyed in
// response shrinks the pump.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a49c0c64-ea87-44f4-9acf-e7181962aeb2",
		Name:         "Pia, Determined Rebuilder",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Pia, Determined Rebuilder — create a 1/1 Thopter with flying",
				Do(CreateToken{Template: TokenCard("1/1 colorless Thopter artifact with flying"), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{5}{R}: Target creature gets +X/+0 until end of turn, where X is the number of artifacts you control",
			Cost:    ManaCost("{5}{R}"),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				x := countControlled(g, item.Controller, game.Card.IsArtifact)
				for _, t := range ctx.LegalTargets() {
					return BoostUntilEOT{Target: t.ID, Power: x, Label: "Pia, Determined Rebuilder — +X/+0"}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
