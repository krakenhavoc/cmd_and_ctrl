package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// westvaleAbbeyOracle is shared by both faces: the front registers
// under it, the back under westvaleAbbeyOracle + "#1".
const WestvaleAbbeyOracle = "04eeb9ad-5c59-411b-8809-db8349838588"

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
	effects.Register(effects.Spec{
		OracleID:     WestvaleAbbeyOracle,
		Name:         "Westvale Abbey",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "{T}: Add {C}.",
		}},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{5}, {T}, Pay 1 life: Create a 1/1 white and black Human Cleric creature token.",
				Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1},
				Cost:    effects.Plus(effects.ManaCost("{5}"), effects.TapCost(), effects.PayLife(1)),
				Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white and black Human Cleric"), N: 1}),
			},
			{
				Label:   "{5}, {T}, Sacrifice five creatures: Transform this land, then untap it.",
				Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
				Cost:    effects.Plus(effects.ManaCost("{5}"), effects.TapCost(), effects.SacrificeN(5, "five creatures", effects.Creature())),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.TransformThis{}).Apply(ctx); err != nil {
						return err
					}
					return effects.UntapTarget{Target: item.SourceCardID}.Apply(ctx)
				},
			},
		},
	})
	effects.Register(effects.Spec{
		OracleID:        WestvaleAbbeyOracle + "#1",
		Name:            "Ormendahl, Profane Prince",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink", "indestructible", "haste"},
	})
}
