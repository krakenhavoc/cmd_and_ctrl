package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arnim Zola, Bio-Fanatic — Legendary Artifact Creature — Scientist Villain
// {2}{B}, 2/3 (Marvel Super Heroes):
//
//	"{3}, {T}: Create a tapped 2/1 black Villain creature token with
//	 menace. Activate only if there are two or more creature cards in your
//	 graveyard."
//
// The activation condition is the graveyard-count gate (GraveyardAtLeast over
// creature cards), checked when the ability is activated (CR 602.5b) and by
// the legal-action view. The token enters tapped, the way Ainok Strike
// Leader's Goblins do.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "95161e23-e4c4-411f-bce5-58a02abc0958",
		Name:         "Arnim Zola, Bio-Fanatic",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{3}, {T}: Create a tapped 2/1 black Villain creature token with menace. Activate only if there are two or more creature cards in your graveyard.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      Plus(ManaCost("{3}"), TapCost()),
			Condition: GraveyardAtLeast(2, func(c game.Card) bool { return c.IsCreature() }),
			Effect:    Do(CreateToken{Template: mshB1TappedVillainToken(), N: 1}),
		}},
	})
}

// mshB1TappedVillainToken is the 2/1 black Villain with menace, entering
// tapped.
func mshB1TappedVillainToken() game.Card {
	tmpl := TokenCard("2/1 black Villain with menace")
	tmpl.Tapped = true
	return tmpl
}
