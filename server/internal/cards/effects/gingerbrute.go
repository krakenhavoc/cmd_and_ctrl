package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gingerbrute — Artifact Creature — Food Golem {1}, 1/1:
//
//	"Haste
//	 {1}: This creature can't be blocked this turn except by creatures
//	 with haste.
//	 {2}, {T}, Sacrifice this creature: You gain 3 life."
//
// The reference card for an UNTIL-END-OF-TURN block rule (#750, ADR
// 0045 addendum Decision 11): the {1} registers a cantBeBlockedExceptBy
// ScopedEffect record (ADR 0041 phase 3 tier 3b, #1497), pinned to this
// object at resolution (CR 611.2c) and swept at the turn's end. A
// Gingerbrute that leaves and returns is a new object (CR 400.7) and
// is blockable again. Restrictions are checked at declaration
// (CR 509.1b), so activating it after blocks are declared does
// nothing — the rule, not a gap. Activating it twice is harmless:
// two identical records refuse the same blockers.
//
// "Creatures with haste" reads the blocker's effective abilities, so a
// granted haste counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "10b8d4c7-7553-4d76-b643-d98b80701e13",
		Name:            "Gingerbrute",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []ActivatedAbility{
			{
				Label: "{1}: This creature can't be blocked this turn except by creatures with haste.",
				Cost:  ManaCost("{1}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CantBeBlockedThisTurnExceptBy{
						Target:   item.SourceCardID,
						Keywords: []string{"haste"},
						Text:     "creatures with haste",
						Label:    "Gingerbrute — can't be blocked except by creatures with haste",
					}.Apply(NewContext(g, item))
				},
			},
			{
				Label: "{2}, {T}, Sacrifice this creature: You gain 3 life.",
				Cost:  Plus(ManaCost("{2}"), TapCost(), SacrificeThis()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return GainLife{Player: item.Controller, Amount: 3}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
