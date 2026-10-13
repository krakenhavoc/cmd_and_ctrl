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
// # Kicker
//
// Kicker {G} is the ordinary optional additional cost (CR 702.33):
// OptionalCosts: Kicker("{G}"), read back at resolution with
// ctx.WasKicked(). A kicked Vines gives the same creature +4/+4 until
// end of turn after the protection.
//
// The remaining simplification (1) is the only one left, and it only
// takes targets away.
func init() {
	Register(Spec{
		OracleID:     "8998d211-5ca6-41d2-bc64-f51f70bd41e5",
		Name:         "Vines of Vastwood",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"It can only be cast on a creature you control.",
		},
		Targets:       TargetCreature("target creature you control", YouControl()),
		OptionalCosts: []game.AdditionalCost{Kicker("{G}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			if err := (GrantKeywordUntilEOT{
				Target:   item.Targets[0].ID,
				Keywords: []string{"hexproof"},
				Label:    "Vines of Vastwood — hexproof",
			}).Apply(ctx); err != nil {
				return err
			}
			if !ctx.WasKicked() {
				return nil
			}
			return BoostUntilEOT{Target: item.Targets[0].ID, Power: 4, Toughness: 4, Label: "Vines of Vastwood — kicked: +4/+4"}.Apply(ctx)
		},
	})
}
