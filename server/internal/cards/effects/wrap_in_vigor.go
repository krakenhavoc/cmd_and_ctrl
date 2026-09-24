package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wrap in Vigor — Instant {1}{G}:
//
//	"Regenerate each creature you control."
//
// The green two-mana answer to a board wipe: every creature you
// control gets a regeneration shield (CR 701.19a), so the Wrath that
// follows taps your team and removes all damage from it instead of
// killing it.
//
// It does NOT target. "Each creature you control" is a set read at
// resolution, so hexproof and shroud on your own creatures are
// irrelevant and nothing can fizzle it by removing one creature in
// response — the rest still get their shields.
//
// What beats it, as printed: Damnation, Wrath of God and Damn say the
// creatures can't be regenerated (CR 701.19c), which ignores every
// shield this made and does not even spend them (CR 701.19c); a
// sacrifice is not a destruction (CR 701.21a); and an effect that
// exiles rather than destroys never opens the window at all. A
// creature that regenerates is TAPPED afterwards, which is the real
// cost of casting this on your own turn.
func init() {
	Register(Spec{
		OracleID:     "39da2aa8-f4d9-44f6-a446-488beaec821f",
		Name:         "Wrap in Vigor",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			me := ctx.Controller()
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if !c.IsCreature() || c.Controller != me {
					continue
				}
				if err := (Regenerate{Target: c.InstanceID}).Apply(ctx.asGroupMember()); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
