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
// Both halves are live. The doubled branch reads `CastFromZone`
// rather than the alternative cost that was paid, because that is
// what the card says: "if this spell was cast from a GRAVEYARD", not
// "if you paid its flashback cost". The two coincide for every
// printing that exists, but a future effect that lets you cast this
// from a graveyard for its normal cost (Yawgmoth's Will, Underworld
// Breach) doubles it correctly for free, and a hypothetical flashback
// paid from somewhere else would not.
//
// Flashback rides the S29 machinery, which binds the two fields that
// must agree: the cost is claimable only out of the graveyard, and
// `CastableZones` has to open that zone or `Register` panics. The
// exile-on-leaving-stack half is what keeps the card from flashing
// back every turn forever, and it is a replacement rather than an
// appended exile — so a fizzled or countered Vengeance is exiled too.
func init() {
	Register(Spec{
		OracleID:     "a5ca7bd9-0964-405f-adb9-7c27153595e6",
		Name:         "Increasing Vengeance",
		Completeness: CompletenessFull,
		Targets: instantOrSorcerySpell(
			"target instant or sorcery spell you control", YouControl()),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{3}{R}{R}")},
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
