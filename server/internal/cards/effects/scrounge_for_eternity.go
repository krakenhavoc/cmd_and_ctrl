package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scrounge for Eternity — Sorcery {2}{B} (Edge of Eternities, #4461):
//
//	"As an additional cost to cast this spell, sacrifice an artifact or
//	 creature.
//	 Return target creature or Spacecraft card with mana value 5 or less
//	 from your graveyard to the battlefield. Then create a Lander token."
//
// The sacrifice is Anchor to Reality's additional cost. Targets are
// chosen before costs are paid (CR 601.2c, 601.2h), so the permanent
// sacrificed for this spell cannot be its target. The return is Doomed
// Necromancer's, over Stray from History's "creature or Spacecraft
// card", with the mana value read where the card is. The Lander is a
// second step that waits for any question the returning card asks as it
// enters (CR 608.2c). A target that has left the graveyard makes the
// whole spell do nothing, Lander included (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "71982b17-1d9a-4f98-b8b5-d232e3a8cf1d",
		Name:           "Scrounge for Eternity",
		Completeness:   CompletenessFull,
		Purpose:        game.Purpose{Tokens: 1},
		AdditionalCost: SacrificeCost("an artifact or creature", Or(Artifact(), Creature())),
		Targets: TargetCardInGraveyard("target creature or Spacecraft card with mana value 5 or less from your graveyard",
			YouOwn(), Or(Creature(), Subtype("Spacecraft")), ManaValueLE(5)),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				_, _ = reanimateSingleTarget(ctx, ctx.Controller())
				return nil
			}),
			CreateToken{Template: b16LanderToken(), N: 1},
		),
	})
}
