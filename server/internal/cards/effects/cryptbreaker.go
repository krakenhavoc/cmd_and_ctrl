package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cryptbreaker — Creature — Zombie {B}, 1/1:
//
//	"{1}{B}, {T}, Discard a card: Create a 2/2 black Zombie creature token.
//	 Tap three untapped Zombies you control: You draw a card and you lose 1
//	 life."
//
// The first ability is a mana cost, {T} and a discard. The second taps three
// untapped Zombies you control and names no {T}, so Cryptbreaker may be one of
// the three, even with summoning sickness (CR 302.6 covers only the {T}
// symbol), and a Zombie's summoning sickness does not matter either; it is
// tapped as a cost and taps nothing twice.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "16f75a65-d067-4f79-81b9-833c373cf4d2",
		Name:         "Cryptbreaker",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label: "{1}{B}, {T}, Discard a card: Create a 2/2 black Zombie creature token.",
				Cost:  Plus(ManaCost("{1}{B}"), TapCost(), DiscardACard()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateToken{Template: TokenCard("2/2 black Zombie"), N: 1}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "Tap three untapped Zombies you control: You draw a card and you lose 1 life.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
				Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
					Count:  3,
					Filter: TargetPermanent("three untapped Zombies you control", HasSubtype("Zombie")),
					Label:  "three untapped Zombies you control",
				}},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (DrawCards{N: 1}).Apply(ctx); err != nil {
						return err
					}
					return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -1)
				},
			},
		},
	})
}
