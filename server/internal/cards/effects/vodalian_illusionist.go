package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vodalian Illusionist — Creature — Merfolk Wizard {2}{U}, 1/2:
//
//	"{U}{U}, {T}: Target creature phases out. (While it's phased out,
//	 it's treated as though it doesn't exist. It phases in before its
//	 controller untaps during their next untap step.)"
//
// The Mirage-block phasing engine, and the smallest complete
// statement of CR 702.26 in the catalog: one target, one verb, and
// every consequence of the verb is the engine's.
//
// It is a Fog for one creature and a removal spell for none. Pointed
// at an attacker it takes it out of combat (CR 506.4, named in
// CR 702.26b's own text); pointed at a blocker in the declare-blockers
// step the attacker stays blocked, which is what the rule says and is
// the trap the card is known for. Pointed at your own creature in
// response to a Doom Blade it saves it — the spell loses its only
// target and is countered on resolution (CR 608.2b) — and the
// creature comes back at your untap step with its counters, its
// damage, its Auras and its Equipment, because none of that is a zone
// change (CR 702.26d).
//
// That last clause is the whole reason this card is the proof card:
// an Equipment on the target phases out with it and phases back in
// still attached (CR 702.26g), and nothing in this file says so.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ac935639-ba30-4b04-86f6-363527e85a8b",
		Name:         "Vodalian Illusionist",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{U}{U}, {T}: Target creature phases out.",
			Cost:    Plus(ManaCost("{U}{U}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return PhaseOut{Targets: legalTargetCards(item, g)}.Apply(ctx)
			},
		}},
	})
}
