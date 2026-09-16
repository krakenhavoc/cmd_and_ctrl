package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Siege-Gang Commander — Creature — Goblin {3}{R}{R}, 2/2 (EDHREC
// rank 1590):
//
//	"When this creature enters, create three 1/1 red Goblin creature
//	 tokens.
//	 {1}{R}, Sacrifice a Goblin: This creature deals 2 damage to any
//	 target."
//
// Four bodies and a sac outlet. The ETB is a real trigger with a
// response window; the Goblins are Krenko's template. The sacrifice
// is Skirk Prospector's "a Goblin" clause — effective subtypes, so a
// changeling counts and the Commander can eat itself, as printed —
// paid at announce, so a Goblin's dies-triggers resolve above the
// damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ddc7f59a-bbb1-4ba1-82c8-6813fd191940",
		Name:         "Siege-Gang Commander",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Siege-Gang Commander — create three 1/1 red Goblins", Do(CreateToken{Template: RedGoblinToken(), N: 3})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}, Sacrifice a Goblin: This creature deals 2 damage to any target.",
			Cost:    Plus(ManaCost("{1}{R}"), game.AbilityCost{SacrificeOther: b12SacrificeAGoblin()}),
			Targets: TargetAny(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if err := (DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 2}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
