package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eldrazi Confluence — Instant {2}{C}{C} (Modern Horizons 3 Commander,
// #5468):
//
//	"Choose three. You may choose the same mode more than once.
//	 • Target creature gets +3/-3 until end of turn.
//	 • Exile target nonland permanent, then return it to the battlefield
//	   tapped under its owner's control.
//	 • Create a 1/1 colorless Eldrazi Scion creature token with
//	   "Sacrifice this token: Add {C}.""
//
// The Confluence frame (Fiery Confluence's): three choices, repeats
// allowed, each targeted bullet with its own target group, resolved in
// the order chosen. The flicker is Eldrazi Displacer's (Flicker,
// returned tapped under the OWNER's control as a new object, CR 400.7),
// so the same permanent chosen for that bullet twice is flickered once:
// the second occurrence's target is gone (CR 608.2b). The Scion is the
// printed token, sacrifice ability included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "01160cb5-6d83-4317-9522-74083b3a83bb",
		Name:         "Eldrazi Confluence",
		Completeness: CompletenessFull,
		Modes: ChooseNRepeating("Choose three. You may choose the same mode more than once.", 3, 3,
			ModeDoing("Target creature gets +3/-3 until end of turn.",
				TargetCreature("target creature"),
				BoostTheModesTarget(3, -3, "Eldrazi Confluence — +3/-3")),
			ModeDoing("Exile target nonland permanent, then return it to the battlefield tapped under its owner's control.",
				TargetPermanent("target nonland permanent", Nonland()),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return Flicker{Target: t.ID, Tapped: true}.Apply(ctx)
				}),
			ModeWithPurpose(ModeDoing("Create a 1/1 colorless Eldrazi Scion creature token with \"Sacrifice this token: Add {C}.\"", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return CreateToken{Template: b28EldraziScionToken(), N: 1}.Apply(ctx)
				}), game.Purpose{Tokens: 1, CreatureTokens: 1}),
		),
	})
}
