package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chocobo Camp — Land (Final Fantasy Commander, #4462):
//
//	"This land enters tapped unless you control a legendary creature.
//	 {T}: Add {G}. When you next cast a Bird creature spell this turn, it
//	 enters with an additional +1/+1 counter on it.
//	 {2}{G}{G}, {T}: Create a 2/2 green Bird creature token with "Whenever
//	 a land you control enters, this token gets +1/+0 until end of turn.""
//
// The tapped entry reads the battlefield as the land enters (the
// legendary creature must already be there, CR 614.12), Minas Tirith's
// condition. The mana ability's rider is Yuna, Grand Summoner's one-use
// promise narrowed to Bird creature spells: the first Bird creature spell
// cast this turn enters with the counter, a non-Bird creature spell cast
// first does not use it up (CR 603.7b), and it ends at cleanup. The
// token is the set's Chocobo Bird.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ed77fdf2-59c0-4310-9b12-80d28beeaeef",
		Name:         "Chocobo Camp",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTappedUnless(b08ControlsLegendaryCreature)},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}. When you next cast a Bird creature spell this turn, it enters with an additional +1/+1 counter on it",
			Rider: func(g *game.Game, controller, source uuid.UUID) error {
				bird := "Bird"
				if c, ok := g.LookupCardForEffect(source); ok {
					bird = c.TextSubtype("Bird")
				}
				g.GrantNextSpellPromiseForEffect(controller, game.NextSpellPromise{
					Filter:      game.PermissionFilter{CreatureOnly: true, CreatureType: bird},
					CounterKind: game.CounterPlusOne,
					Counters:    1,
					Text:        "The next Bird creature spell you cast this turn enters with an additional +1/+1 counter.",
				}, "Chocobo Camp", source, g.UntilEndOfTurnDuration())
				return nil
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}{G}{G}, {T}: Create a 2/2 green Bird creature token with \"Whenever a land you control enters, this token gets +1/+0 until end of turn.\"",
			Purpose: game.Purpose{Tokens: 1, Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}{G}{G}"), TapCost()),
			Effect:  Do(CreateToken{Template: chocoboBirdToken(), N: 1}),
		}},
	})
}
