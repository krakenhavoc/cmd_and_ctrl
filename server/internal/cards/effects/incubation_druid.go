package effects

import (
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Incubation Druid — Creature — Elf Druid {1}{G}, 0/2:
//
//	"{T}: Add one mana of any type that a land you control could
//	 produce. If this creature has a +1/+1 counter on it, add three
//	 mana of that type instead.
//	 {3}{G}{G}: Adapt 3. (If this creature has no +1/+1 counters on
//	 it, put three +1/+1 counters on it.)"
//
// The mana ability reads the same union Reflecting Pool's
// DerivedFromOwnLands does — "a land you control could produce",
// which is CR 106.7 and includes colourless — but Reflecting Pool's
// DerivedMatch always produces exactly one of the picked type, and
// this card's amount is CONDITIONAL (one, or three with a +1/+1
// counter). DerivedMatch has no multiplier slot for that, so this
// asks game.ProducibleManaLocked directly, per controlled land, and
// builds its own pipe string with the counter-gated amount —
// druidPipeStringWithAmount below, OneColorOfAmount's shape with an
// arbitrary symbol set instead of the fixed five colours.
//
// Adapt is not a keyword the engine reads generically (nothing grants
// it to another permanent, so it stays out of the closed
// PrintedKeywords table): it is one ordinary activated ability whose
// effect checks the +1/+1 count itself, exactly as Adapt's own
// reminder text says.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "34428f42-03ac-4795-8286-6cbea796df2b",
		Name:         "Incubation Druid",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{Tap: true},
			ProducedFunc: func(g *game.Game, controller, source uuid.UUID) string {
				seen := map[string]bool{}
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller != controller || !c.IsLand() {
						continue
					}
					for _, m := range g.ProducibleManaLocked(c) {
						seen[m] = true
					}
				}
				n := 1
				if src, ok := g.LookupCardForEffect(source); ok && src.Counters["+1/+1"] > 0 {
					n = 3
				}
				return druidPipeStringWithAmount(seen, n)
			},
			Label: "Add one mana of any type a land you control could produce (three with a +1/+1 counter)",
		}},
		Activated: []ActivatedAbility{{
			Label: "{3}{G}{G}: Adapt 3",
			Cost:  ManaCost("{3}{G}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				src, ok := g.LookupCardForEffect(item.SourceCardID)
				if !ok || src.Counters["+1/+1"] > 0 {
					return nil
				}
				return AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: 3}.Apply(NewContext(g, item))
			},
		}},
	})
}

// druidPipeStringWithAmount renders a symbol set (WUBRGC, unordered
// map input) as one pick-one-of-these-N-tokens slot, in canonical
// order — OneColorOfAmount's shape, generalised to an arbitrary
// symbol set instead of the fixed five colours (Incubation Druid's
// "any type a land you control could produce" can include {C}, which
// OneColorOfAmount deliberately excludes for "any COLOR" cards).
func druidPipeStringWithAmount(seen map[string]bool, n int) string {
	if n <= 0 {
		return ""
	}
	count := ""
	if n != 1 {
		count = strconv.Itoa(n)
	}
	var parts []string
	for _, m := range []string{"W", "U", "B", "R", "G", "C"} {
		if seen[m] {
			parts = append(parts, m+count)
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return "{" + strings.Join(parts, "|") + "}"
}
