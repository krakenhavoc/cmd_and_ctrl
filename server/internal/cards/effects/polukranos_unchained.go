package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Polukranos, Unchained — Legendary Creature — Zombie Hydra {2}{B}{G}, 0/0:
//
//	"Polukranos enters with six +1/+1 counters on it. It escapes with twelve +1/+1 counters on it instead.
//	 If damage would be dealt to Polukranos while it has a +1/+1 counter on it, prevent that damage and remove that many +1/+1 counters from it.
//	 {1}{B}{G}: Polukranos fights another target creature.
//	 Escape—{4}{B}{G}, Exile six other cards from your graveyard."
//
// ADR 0108 §8 (#1906): a prevention static with an additional effect. It
// removes "that many" counters — the damage, prevented or not, so damage
// that can't be prevented is dealt and still removes them before lethal
// damage is checked (CR 615.12) — or every counter it has when that is
// fewer, and all of the damage is still prevented (the rulings).
//
// "Escapes with twelve instead": every entry gets the printed six, and the
// escape cost adds six more (EscapeWithCounters). The two are summed into
// one entry before any counter is placed, so the total is twelve, and a
// Doubling Season or a Hardened Scales sees one placement of twelve,
// exactly as it would the printed "instead".
//
// The fight is b30SourceFightsFirstLegalTarget: a Polukranos no longer on
// the battlefield fights nothing (the ruling).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:         "649e7237-b38b-43e9-83f4-763751fb1bea",
		Name:             "Polukranos, Unchained",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{EscapeWithCounters("{4}{B}{G}", 6, 6)},
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterPlusOne, 6, "Polukranos, Unchained: enters with six +1/+1 counters"),
			PreventDamageDealtTo(PreventionStatic{
				To:    ToThisCreature,
				While: WhileItHasAPlusOneCounter,
				Then:  removeThatManyCountersFromThisBody,
				Label: "Polukranos, Unchained — prevent damage to it and remove that many +1/+1 counters",
			}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}{G}: Polukranos fights another target creature.",
			Cost:    ManaCost("{1}{B}{G}"),
			Targets: Another(TargetCreature("another target creature")),
			Effect:  b30SourceFightsFirstLegalTarget,
		}},
	})
}
