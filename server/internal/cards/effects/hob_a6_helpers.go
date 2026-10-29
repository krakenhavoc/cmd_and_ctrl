package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// hob_a6_helpers.go — the shared body of The Hobbit's tribal "counter
// lands" (#4454): Mirkwood, Elvenking's Halls and Lake-town.
//
//	"This land enters tapped.
//	 {T}: Add {X} or {Y}.
//	 {2}{X}{Y}, {T}, Sacrifice this land: Put two +1/+1 counters on
//	 target <tribe> you control. Activate only as a sorcery."
//
// They differ only in the two colours and the tribe, so each card file
// is one call.

// hobA6CounterLand builds the Spec of one counter land. `colors` is the
// two basic-colour letters ("BG"), `tribes` the creature types the
// target may have ("Bear", "Spider", "Wolf") and `noun` the printed
// clause around them ("Bear, Spider, or Wolf").
func hobA6CounterLand(oracleID, name, colors, noun string, tribes ...string) Spec {
	first, second := "{"+colors[:1]+"}", "{"+colors[1:]+"}"
	cost := "{2}" + first + second
	preds := make([]CardPredicate, 0, len(tribes))
	for _, t := range tribes {
		preds = append(preds, HasSubtype(t))
	}
	label := "target " + noun + " you control"
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{" + colors[:1] + "|" + colors[1:] + "}",
			Label:    "Add " + first + " or " + second,
		}},
		Activated: []ActivatedAbility{{
			Label: cost + ", {T}, Sacrifice this land: Put two +1/+1 counters on " + label +
				". Activate only as a sorcery.",
			Cost:         Plus(ManaCost(cost), TapCost(), SacrificeThis()),
			Targets:      PermanentYouControl(label, Or(preds...)),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
				}
				return nil
			},
		}},
	}
}
