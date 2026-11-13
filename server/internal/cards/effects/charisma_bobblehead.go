package effects

// Charisma Bobblehead — Artifact — Bobblehead {3} (Fallout, #5469):
//
//	"{T}: Add one mana of any color.
//	 {4}, {T}: Create X 1/1 white Soldier creature tokens, where X is
//	 the number of Bobbleheads you control. Activate only as a sorcery."
//
// Intelligence Bobblehead's count (pipBobbleheadsYouControl), read as
// the ability resolves (CR 608.2h), as the number of Soldiers made in
// one creation.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "77897962-7d67-4fa8-858f-2f35a322d7c6",
		Name:         "Charisma Bobblehead",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []ActivatedAbility{{
			Label:        "{4}, {T}: Create X 1/1 white Soldier creature tokens, where X is the number of Bobbleheads you control. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{4}"), TapCost()),
			SorcerySpeed: true,
			Effect: Do(Instruction(func(ctx *Context) error {
				n := pipBobbleheadsYouControl(ctx.Game, ctx.Controller())
				if n <= 0 {
					return nil
				}
				return CreateToken{Template: TokenCard("1/1 white Soldier"), N: n}.Apply(ctx)
			})),
		}},
	})
}
