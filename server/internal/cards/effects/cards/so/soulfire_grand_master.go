package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soulfire Grand Master — {1}{W} Creature — Human Monk 2/2 (Fate Reforged):
//
//	"Lifelink
//	 Instant and sorcery spells you control have lifelink.
//	 {2}{U/R}{U/R}: The next time you cast an instant or sorcery spell from your
//	 hand this turn, put that card into your hand instead of into your graveyard
//	 as it resolves."
//
// Lifelink is printed; the grant is Heartflame Duelist's spell-keyword static.
// The activated ability promises (NextSpellPromise, #1852) that the next instant
// or sorcery cast from the caster's hand this turn is marked when it is cast
// (CR 601.2i), and as it resolves the card goes to its owner's hand instead of
// its graveyard (StackItem.HandInsteadOfGraveyard, one of the resolution exits).
// A spell cast from any other zone does not spend the promise, and one that is
// countered or fizzles goes to the graveyard as usual, since the clause says "as
// it resolves". A second activation promises a second spell. The promise ends
// with the turn.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e65769f4-50ae-482b-933b-ac5cf3843f2a",
		Name:            "Soulfire Grand Master",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Static: []game.StaticAbility{
			effects.Labelled(effects.SpellsYouControlHave(effects.Or(effects.Instant(), effects.Sorcery()), "lifelink"),
				"Instant and sorcery spells you control have lifelink."),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U/R}{U/R}: The next time you cast an instant or sorcery spell from your hand this turn, put that card into your hand instead of into your graveyard as it resolves.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{2}{U/R}{U/R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				g.GrantNextSpellPromiseForEffect(item.Controller, game.NextSpellPromise{
					Filter:                 game.PermissionFilter{InstantOrSorceryOnly: true},
					FromHandOnly:           true,
					HandInsteadOfGraveyard: true,
					Text:                   "Soulfire Grand Master — put that card into your hand instead of into your graveyard as it resolves",
				}, "Soulfire Grand Master", item.SourceCardID, game.Duration{})
				return nil
			},
		}},
	})
}
