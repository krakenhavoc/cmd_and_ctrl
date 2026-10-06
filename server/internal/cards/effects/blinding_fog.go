package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blinding Fog — Instant {2}{G}:
//
//	"Prevent all damage that would be dealt to creatures this turn.
//	 Creatures you control gain hexproof until end of turn."
//
// The two sentences fix their sets differently, and that is CR 611.2c:
//
//   - The shield is Forfend's (#2045): every creature, whoever controls
//     it, read as the damage would be dealt. A prevention effect doesn't
//     modify characteristics, so a creature that enters later is
//     protected too.
//   - "Gain hexproof" adds an ability, which modifies characteristics,
//     so the creatures you control as Blinding Fog resolves are the ones
//     that gain it, and a creature that enters later doesn't.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "50555471-a03c-4735-a6ce-6ae96e27d7c0",
		Name:         "Blinding Fog",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (PreventDamageFromSource{Protect: ShieldCreatures}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{
				Match:    And(Creature(), YouControl()),
				Keywords: []string{"hexproof"},
				Label:    "Blinding Fog — creatures you control gain hexproof",
			}.Apply(ctx)
		},
	})
}
