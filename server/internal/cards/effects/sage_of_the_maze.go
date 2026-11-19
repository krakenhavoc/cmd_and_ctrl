package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sage of the Maze — Creature — Elf Wizard {2}{G}, 1/3:
//
//	"{T}: Add two mana in any combination of colors.
//	 {T}: Until end of turn, target land you control becomes an X/X Citizen
//	 creature with haste in addition to its other types, where X is twice
//	 the number of Gates you control. Activate only as a sorcery.
//	 Tap an untapped Gate you control: Untap this creature."
//
// The mana ability is Hermitic Herbalist's unrestricted two-in-any-
// combination. The animation is Cactus Preserve's (cmdr2B13Animate): one
// continuous effect pinned to the land for the turn, adding the Creature
// type and the Citizen type and haste, with the base power and toughness
// set to the number counted as the ability RESOLVES, twice the Gates its
// controller controls then (CR 608.2h), the land staying a land. The untap
// is paid by tapping an untapped Gate, which is not the {T} symbol
// (CR 302.6) and so may be a Gate that arrived this turn.
func init() {
	const name = "Sage of the Maze"
	Register(Spec{
		OracleID:     "d82618e7-494a-46d0-a19e-9cc647f96cfe",
		Name:         name,
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: AnyCombinationOfColors(2),
			Label:    "Add two mana in any combination of colors",
		}},
		Activated: []ActivatedAbility{
			{
				Label:        "{T}: Until end of turn, target land you control becomes an X/X Citizen creature with haste in addition to its other types, where X is twice the number of Gates you control. Activate only as a sorcery.",
				Cost:         TapCost(),
				SorcerySpeed: true,
				Targets:      TargetPermanent("target land you control", Land(), YouControl()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					x := 2 * len(MatchingBattlefield(ctx, And(YouControl(), OfSubtype("Gate"))))
					return cmdr2B13Animate(ctx, id, name+" — an X/X Citizen creature with haste until end of turn", x, x, "Citizen", nil, "haste")
				},
			},
			{
				UntapSelf: true,
				Label:     "Tap an untapped Gate you control: Untap this creature.",
				Purpose:   game.Purpose{Answers: game.AnswerValue},
				Cost:      TapAnUntapped("an untapped Gate you control", OfSubtype("Gate")),
				Effect:    gapB44UntapThis,
			},
		},
	})
}
