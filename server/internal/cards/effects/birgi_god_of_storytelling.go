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
// Boast (CR 702.142) is not an ability the catalog can declare: nothing
// in it prints a boast, so the second clause has nothing to double and
// is left off, and the card says so as a caveat.
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
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Boast isn't implemented — the second boast activation is not offered."},
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
			Label: "Discard a card: Exile the top two cards of your library. You may play those cards this turn.",
			Cost:  DiscardACard(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return ExileTopWithPermission{From: item.Controller, GrantTo: item.Controller, N: 2}.Apply(ctx)
			},
		}},
	})
}
