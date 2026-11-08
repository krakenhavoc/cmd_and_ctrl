package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lupinflower Village — Land (Bloomburrow, #4467):
//
//	"{T}: Add {C}.
//	 {T}: Add {W}. Spend this mana only to cast a creature spell.
//	 {1}{W}, {T}, Sacrifice this land: Look at the top six cards of your
//	 library. You may reveal a Bat, Bird, Mouse, or Rabbit card from
//	 among them and put it into your hand. Put the rest on the bottom of
//	 your library in a random order."
//
// Bloomburrow's white creature-type land, built like Lilypad Village:
// the {W} carries the cast and creature restriction tags, and is a
// deliberate click since the auto-tapper leaves restricted abilities
// alone. The sacrifice ability is Dáin's Company's look: a choose-cards
// prompt over the six cards with a "may" floor of zero, the taken card
// revealed, and the rest to the bottom in a random order
// (takeRevealRestBottomRandom). The four types are words of the
// Village's text (AnySubtype reads them through it, CR 612.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b6c7c708-5212-4100-b954-b77855b27915",
		Name:         "Lupinflower Village",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{W}",
				Label:    "Add {W} (creature spells only)",
				Restrictions: []string{
					ManaRestrictCast,
					ManaRestrictType("Creature"),
				},
			},
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}, {T}, Sacrifice this land: Look at the top six cards of your library. You may reveal a Bat, Bird, Mouse, or Rabbit card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}{W}"), TapCost(), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				player := item.Controller
				return TakeFromLibraryToHand{
					Player:   player,
					Cards:    g.LookAtTopOfLibraryForEffect(player, 6),
					Match:    AnySubtype("Bat", "Bird", "Mouse", "Rabbit"),
					Max:      1,
					Optional: true,
					Label:    "Lupinflower Village — you may reveal a Bat, Bird, Mouse, or Rabbit card and put it into your hand",
					Key:      takeRevealRestBottomRandom,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
