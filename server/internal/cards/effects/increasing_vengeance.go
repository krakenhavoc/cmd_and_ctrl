package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Increasing Vengeance — Instant {R}{R}:
//
//	"Copy target instant or sorcery spell you control. If this spell
//	 was cast from a graveyard, copy that spell twice instead. You
//	 may choose new targets for the copies."
//	"Flashback {3}{R}{R}"
//
// Two copies is not one copy that resolves twice: each is created
// separately, each gets its own CR 706.10c target choice, and an
// opponent can respond between them being put on the stack and
// either resolving. CopySpell.Count does exactly that — a loop of
// independent CopySpellForEffect calls.
//
// "You control" is the real narrowing versus Reverberate. This card
// protects your own combo piece rather than stealing the table's
// best spell, which is why it costs the same and sees far less
// Commander play.
//
// DECLARED SIMPLIFICATION — the flashback half is unreachable.
// Flashback (CR 702.34) is not implemented: there is no alternative
// cast-from-graveyard path, so `CastFromZone` can never be
// ZoneGraveyard for this card and the doubled branch never runs. The
// branch is written out anyway, reading the field the cast path
// already stamps, so the card starts doubling the day flashback
// lands with no change here — the same posture Darksteel Citadel's
// indestructible took for five sprints. Shipping without the branch
// would err weaker in a way that hides what the card does; shipping
// with it and pretending flashback works would err stronger, which
// is why the deferral is named rather than left implicit.
func init() {
	Register(Spec{
		OracleID: "a5ca7bd9-0964-405f-adb9-7c27153595e6",
		Name:     "Increasing Vengeance",
		Targets: instantOrSorcerySpell(
			"target instant or sorcery spell you control", YouControl()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			count := 1
			if item.CastFromZone == game.ZoneGraveyard {
				count = 2
			}
			return CopySpell{
				StackID:          item.Targets[0].ID,
				Controller:       ctx.Controller(),
				Count:            count,
				ChooseNewTargets: true,
			}.Apply(ctx)
		},
	})
}
