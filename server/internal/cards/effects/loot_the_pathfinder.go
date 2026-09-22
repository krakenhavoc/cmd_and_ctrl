package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Loot, the Pathfinder — Legendary Creature — Beast Noble
// {2}{G}{U}{R}, 2/4:
//
//	"Double strike, vigilance, haste
//	 Exhaust — {G}, {T}: Add three mana of any one color.
//	 Exhaust — {U}, {T}: Draw three cards.
//	 Exhaust — {R}, {T}: Loot deals 3 damage to any target.
//	 (Activate each exhaust ability only once.)"
//
// The card #1181 shaped its record for and #1183 finally reaches. It
// is the printed proof of two things the record was built to be, and
// of one thing it could not do until now:
//
//   - THREE exhaust abilities on one permanent, each activatable
//     once. That is why the record is keyed by the ability's LABEL
//     and not by the permanent — spending the draw must leave the
//     damage and the mana alone. Nothing in this file says so; it
//     falls out of the key.
//   - Each of them also costs {T}, so in practice Loot spends one per
//     untap even before the keyword is counted. The exhaust record is
//     the half that survives the untap step: a Loot that has drawn
//     three cards never draws three again, however many turns pass,
//     unless it becomes a new object (CR 400.7 — a flicker, not an
//     untap).
//   - The FIRST ability is a MANA ability (CR 605.1a: no target,
//     could add mana, not a loyalty ability), which takes the other
//     entry point — ActivateManaAbility, CR 605.3a, no stack and no
//     priority — and so had no activation record at all until #1183.
//     `ManaAbility.Exhaust` is the marker; ActivateManaAbility and
//     the auto-tapper's executor both write the record now, and the
//     activation path, the enumerator, the wire, the planner and
//     CR 106.7's "could produce" all read one Game.ManaAbilityExhausted.
//
// "Add three mana of any one color" is ONE colour pick that adds three
// tokens of it (#742, Gilded Lotus' shape) — OneColorOfAmount(3) —
// not three independent picks, which would be strictly stronger than
// printed.
//
// The auto-tapper never plans this ability, and that is not a
// simplification of exhaust: a mana ability with a MANA component in
// its cost ({G} here) is excluded from planning outright, because the
// planner would have to solve a second cost to fund the first
// (autoTapAbilityFor). The player floats the {G} and clicks, which is
// how the card is played on paper. Were that to change, a spent Loot
// is already refused by the planner as well.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "68c7e459-0932-4644-a3c0-9a1eae1db7a3",
		Name:            "Loot, the Pathfinder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike", "vigilance", "haste"},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true, Mana: "{G}"},
			Produced: OneColorOfAmount(3),
			Label:    "Exhaust — {G}, {T}: Add three mana of any one color.",
			Exhaust:  true,
		}},
		Activated: []ActivatedAbility{
			{
				Label:   "Exhaust — {U}, {T}: Draw three cards.",
				Exhaust: true,
				Cost:    Plus(ManaCost("{U}"), TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return DrawCards{Player: item.Controller, N: 3}.Apply(ctx)
				},
			},
			{
				Label:   "Exhaust — {R}, {T}: Loot deals 3 damage to any target.",
				Exhaust: true,
				Cost:    Plus(ManaCost("{R}"), TapCost()),
				Targets: TargetAny(),
				Effect:  b33DamageChosenTargetFromSource(3),
			},
		},
	})
}
