package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prosperity Tycoon — Creature — Human Noble {3}{W}, 4/2 (Outlaws of
// Thunder Junction, #4468):
//
//	"When this creature enters, create a 1/1 red Mercenary creature token
//	 with '{T}: Target creature you control gets +1/+0 until end of
//	 turn. Activate only as a sorcery.'
//	 {2}, Sacrifice a token: This creature gains indestructible until end
//	 of turn. Tap it. (Damage and effects that say "destroy" don't
//	 destroy it.)"
//
// The enters trigger makes the set's Mercenary template (MercenaryToken).
// The activation is Underrealm Lich's body for a {2}-and-a-token cost:
// indestructible until end of turn on this object (CR 702.12b), then it
// taps. "Tap it" is effect, not cost, so a tapped Tycoon can still
// activate, and any token you control pays, the Tycoon itself included
// when it is a token copy.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "07acb7e5-7aae-4d89-aabf-cbd329b3480b",
		Name:         "Prosperity Tycoon",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Prosperity Tycoon — create a 1/1 red Mercenary", effects.CreateAMercenary()),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Sacrifice a token: This creature gains indestructible until end of turn. Tap it.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive | game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.B08SacrificeAToken()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := (effects.GrantKeywordUntilEOT{
					Target:   item.SourceCardID,
					Keywords: []string{"indestructible"},
					Label:    "Prosperity Tycoon — indestructible",
				}).Apply(ctx); err != nil {
					return err
				}
				return effects.TapTarget{Target: item.SourceCardID}.Apply(ctx)
			},
		}},
	})
}
