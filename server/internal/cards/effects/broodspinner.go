package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Broodspinner — Creature — Spider {B}{G}, 2/3 (Duskmourn: House of
// Horror, #4466):
//
//	"Reach
//	 When this creature enters, surveil 2.
//	 {4}{B}{G}, {T}, Sacrifice this creature: Create a number of 1/1 black
//	 and green Insect creature tokens with flying equal to the number of
//	 card types among cards in your graveyard."
//
// The enters trigger is Wary Watchdog's surveil, written with the Do
// primitive that declares a footprint. The activated ability's whole cost
// (mana, {T} and the sacrifice) is paid at announce (CR 602.2b), so the
// Broodspinner is already in the graveyard as the ability resolves and its
// own card types (a creature card) are counted with the rest, exactly as
// the count is read then (CR 608.2h: the number is not last-known). The
// Insect is the same table row The Swarmweaver creates. A count of zero
// creates nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ad50214c-4d0e-4f95-9098-fd7d64dbf7f4",
		Name:            "Broodspinner",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Broodspinner — surveil 2", Do(Surveil{N: 2})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{4}{B}{G}, {T}, Sacrifice this creature: Create a number of 1/1 black and green Insect creature tokens with flying equal to the number of card types among cards in your graveyard.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet | game.AnswerMakesBlocker}, // sacrifices itself at instant speed for blockers
			Cost:    game.AbilityCost{Mana: "{4}{B}{G}", Tap: true, SacrificeSelf: true},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				n := b16CardTypesInGraveyard(g, ctx.Controller())
				if n <= 0 {
					return nil
				}
				return CreateToken{Template: TokenCard("1/1 black and green Insect with flying"), N: n}.Apply(ctx)
			},
		}},
	})
}
