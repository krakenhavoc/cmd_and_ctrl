package effects

// Joo Dee, One of Many — Creature — Human Advisor {1}{B}, 2/2 (Avatar:
// The Last Airbender, #4459):
//
//	"{B}, {T}: Surveil 1. Create a token that's a copy of this creature,
//	 then sacrifice an artifact or creature. Activate only as a
//	 sorcery."
//
// Three steps of one Do, each waiting for the one before (CR 608.2c):
// the surveil prompt is answered, then the token copy is made from Joo
// Dee's copiable values (CR 707.2; its last-known ones if it has left,
// CR 608.2h), then its controller sacrifices an artifact or creature,
// chosen as that step runs, so the new token and Joo Dee herself are
// among the choices (the first ruling). Nobody can act between the copy
// and the sacrifice. A copy of a token copy of Joo Dee is still Joo Dee.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4601ed7b-716e-4c89-b733-1eeabe8bf5a2",
		Name:         "Joo Dee, One of Many",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{B}, {T}: Surveil 1. Create a token that's a copy of this creature, then sacrifice an artifact or creature. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{B}"), TapCost()),
			SorcerySpeed: true,
			Effect: Do(
				Surveil{N: 1},
				Instruction(func(ctx *Context) error {
					return CreateTokenCopy{Controller: ctx.Controller(), Copy: ctx.Source(), N: 1}.Apply(ctx)
				}),
				Instruction(func(ctx *Context) error {
					return SacrificeChoice{
						Player:     ctx.Controller(),
						Candidates: permanentsControlledByMatching(ctx.Game, ctx.Controller(), Or(Artifact(), Creature())),
						Question:   "Joo Dee, One of Many — sacrifice an artifact or creature",
					}.Apply(ctx)
				}),
			),
		}},
	})
}
