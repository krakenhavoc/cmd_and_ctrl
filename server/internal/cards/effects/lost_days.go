package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lost Days — Instant — Lesson {4}{U}:
//
//	"The owner of target creature or enchantment puts it into their
//	 library second from the top or on the bottom. You create a Clue
//	 token."
//
// Temporal Cleansing's tuck (PutIntoLibraryAtDepthOrBottom: the OWNER
// picks second from the top or the bottom, the ruling), then the
// caster's Clue, which waits for the owner's answer (ResolveSteps,
// CR 608.2c). Not investigating (the ruling). With its only target
// illegal the spell doesn't resolve, so no Clue (CR 608.2b).
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:     "3bc47eae-cf6f-4b5e-96e1-94a4f307f474",
		Name:         "Lost Days",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target creature or enchantment", Or(Creature(), Enchantment())),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return PutIntoLibraryAtDepthOrBottom{
						Card:  t.ID,
						Depth: 2,
						Label: "Lost Days — put it into your library second from the top or on the bottom",
					}.Apply(ctx)
				}
				return nil
			}),
			CreateToken{Template: ClueToken(), N: 1},
		),
	})
}
