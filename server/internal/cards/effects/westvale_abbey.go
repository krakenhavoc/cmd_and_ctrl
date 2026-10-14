package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// westvaleAbbeyOracle is shared by both faces: the front registers
// under it, the back under westvaleAbbeyOracle + "#1".
const westvaleAbbeyOracle = "04eeb9ad-5c59-411b-8809-db8349838588"

// Westvale Abbey // Ormendahl, Profane Prince — a transforming land:
//
//	Westvale Abbey, Land
//	  "{T}: Add {C}.
//	   {5}, {T}, Pay 1 life: Create a 1/1 white and black Human Cleric
//	   creature token.
//	   {5}, {T}, Sacrifice five creatures: Transform this land, then
//	   untap it."
//
//	Ormendahl, Profane Prince, Legendary Creature — Demon, 9/7
//	  "Flying, lifelink, indestructible, haste"
//
// The land makes Clerics until it can eat five creatures and turn over.
// The five creatures leave as one simultaneous exit (SacrificeN), the
// permanent transforms in place (TransformThis) and is then untapped, so
// Ormendahl can attack the turn it arrives: haste covers the summoning
// sickness rule (CR 302.6), and the untap covers the {T} paid to flip it.
// It is the same object, so nothing re-enters and no entry trigger fires.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     westvaleAbbeyOracle,
		Name:         "Westvale Abbey",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "{T}: Add {C}.",
		}},
		Activated: []ActivatedAbility{
			{
				Label:   "{5}, {T}, Pay 1 life: Create a 1/1 white and black Human Cleric creature token.",
				Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1},
				Cost:    Plus(ManaCost("{5}"), TapCost(), PayLife(1)),
				Effect:  Do(CreateToken{Template: TokenCard("1/1 white and black Human Cleric"), N: 1}),
			},
			{
				Label:   "{5}, {T}, Sacrifice five creatures: Transform this land, then untap it.",
				Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
				Cost:    Plus(ManaCost("{5}"), TapCost(), SacrificeN(5, "five creatures", Creature())),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (TransformThis{}).Apply(ctx); err != nil {
						return err
					}
					return UntapTarget{Target: item.SourceCardID}.Apply(ctx)
				},
			},
		},
	})
	Register(Spec{
		OracleID:        westvaleAbbeyOracle + "#1",
		Name:            "Ormendahl, Profane Prince",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink", "indestructible", "haste"},
	})
}
