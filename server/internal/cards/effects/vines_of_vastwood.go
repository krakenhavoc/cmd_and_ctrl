package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vines of Vastwood — Instant {G} (EDHREC rank 4048):
//
//	"Kicker {G} (You may pay an additional {G} as you cast this
//	 spell.)
//	 Target creature can't be the target of spells or abilities your
//	 opponents control this turn. If this spell was kicked, that
//	 creature gets +4/+4 until end of turn."
//
// One mana to blank a removal spell, two to blank it and win the
// combat. It is the cheapest protection spell green has and the
// reason a Voltron deck can hold up a single Forest.
//
// # What the first clause is
//
// "Can't be the target of spells or abilities your opponents control"
// is hexproof, word for word (CR 702.11b) — the engine's own
// definition of the keyword, granted for the turn, rather than a new
// restriction. Everything follows from that: the creature's
// controller can still target it (with a pump spell, an Aura, a
// sacrifice outlet), an opponent's board wipe still kills it because
// a wipe targets nothing, and an opponent's spell that had ALREADY
// targeted it before the Vines resolved is not affected — it simply
// fizzles later when the re-check finds an illegal target (CR
// 608.2b).
//
// # Declared simplification 1 (weaker): it only protects your own
//
// The printed clause protects against "YOUR opponents" — the Vines
// caster's — while the hexproof keyword protects against the
// CREATURE'S CONTROLLER's opponents. Those agree in every game the
// card is actually played in and come apart only when Vines is cast
// on a creature somebody else controls, where the printed card
// locks the creature's own controller out of targeting it and
// hexproof does not.
//
// Rather than ship a card that reads differently from its text in
// that corner, the target clause is narrowed to "target creature you
// control", where hexproof IS the printed effect exactly. That is
// strictly fewer legal targets than printed and never more.
//
// # Declared simplification 2 (weaker): no kicker
//
// Kicker is an additional cost paid as the spell is cast (CR 702.33),
// and the engine's additional-cost slot carries only the discard,
// sacrifice and pay-life shapes — there is no way to offer an extra
// mana payment at announce or to read back whether it was made
// (#664). So the spell is never kicked and the +4/+4 never happens.
//
// Vines therefore ships as the protection half only, for {G}. Both
// simplifications take something away and neither adds anything, which
// is the direction #259 allows.
func init() {
	Register(Spec{
		OracleID:     "8998d211-5ca6-41d2-bc64-f51f70bd41e5",
		Name:         "Vines of Vastwood",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Kicker isn't available, so the +4/+4 half never happens — Vines only grants the protection.",
			"It can only be cast on a creature you control.",
		},
		Targets: TargetCreature("target creature you control", YouControl()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return GrantKeywordUntilEOT{
				Target:   item.Targets[0].ID,
				Keywords: []string{"hexproof"},
				Label:    "Vines of Vastwood — hexproof",
			}.Apply(ctx)
		},
	})
}
