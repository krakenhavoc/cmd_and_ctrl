package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Birgi, God of Storytelling // Harnfel, Horn of Bounty — modal double-
// faced card, {2}{R} // {4}{R}.
//
// Front, Legendary Creature — God, 3/3:
//
//	"Whenever you cast a spell, add {R}. Until end of turn, you don't
//	 lose this mana as steps and phases end.
//	 Creatures you control can boast twice during each of your turns
//	 rather than once."
//
// The mana carries the keep mark (KeepManaUntilEndOfTurn, #2166), the
// same one Savage Ventmaw uses. The trigger goes on the stack above the
// spell, so the {R} arrives before the spell resolves, as printed.
//
// "Creatures you control can boast twice during each of your turns
// rather than once" is a BoastLimit (#2697, CR 702.142): a per-controller
// replacement of the printed limit of one, read from the battlefield at
// the moment a boast is activated, so it covers every creature the
// controller has, ends when Birgi leaves, and reaches only that
// player's own turn. It is a limit and not an extra activation: a
// creature that has not attacked still cannot boast at all, and two
// Birgis do not make three (the engine takes the largest limit). Each
// boast ability is counted separately, as the activation tally keys it.
//
// Back, Legendary Artifact (registered under "<oracle_id>#1", ADR
// 0034), below:
//
//	"Discard a card: Exile the top two cards of your library. You may
//	 play those cards this turn."
//
// "Play", not "cast", so a land exiled this way can be played.
func init() {
	Register(Spec{
		OracleID:     "fb81e4d3-1d8c-4779-be62-87cf49277e51",
		Name:         "Birgi, God of Storytelling",
		Completeness: CompletenessFull,
		BoastLimits: []game.BoastLimit{
			YourCreaturesBoastTimes("Creatures you control can boast twice during each of your turns rather than once.", 2),
		},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(nil, "Birgi — add {R}, kept until end of turn",
				Do(AddMana{
					Produced: "{R}",
					Riders:   []game.ManaSpendRider{KeepManaUntilEndOfTurn()},
				})),
		},
	})
	Register(Spec{
		OracleID:     "fb81e4d3-1d8c-4779-be62-87cf49277e51#1",
		Name:         "Harnfel, Horn of Bounty",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Discard a card: Exile the top two cards of your library. You may play those cards this turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    DiscardACard(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return ExileTopWithPermission{From: item.Controller, GrantTo: item.Controller, N: 2}.Apply(ctx)
			},
		}},
	})
}
