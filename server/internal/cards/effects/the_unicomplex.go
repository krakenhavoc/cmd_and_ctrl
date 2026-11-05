package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Unicomplex — Land (Star Trek, #4973):
//
//	"This land enters tapped unless you control two or more other lands.
//	 When this land enters untapped, choose one —
//	 • Put target creature card from your graveyard on top of your library.
//	 • Surveil 1.
//	 {T}: Add {B}."
//
// The slowland condition (otherLandsAtLeast) and Mystic Sanctuary's
// "enters untapped" trigger, which reads the source's tapped flag after
// the entry replacement has applied. The trigger is modal (CR 700.2): the
// mode is chosen as it goes on the stack, and the first bullet can only be
// chosen with a creature card in your graveyard to target (CR 700.2a). A
// card that leaves the graveyard in response makes that bullet do nothing
// (CR 608.2b). The tuck goes through the shared exit primitive, so a
// commander card gets the CR 903.9 offer.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b38fcf9a-464a-432a-b13c-9a206d52dcd9",
		Name:         "The Unicomplex",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{EntersTappedUnless(otherLandsAtLeast(2))},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: selfEnteredUntapped,
			Key:       "The Unicomplex — choose one",
			Effect:    func(*game.Game, *game.StackItem) error { return nil },
			Modes: ChooseOne(
				ModeDoing("Put target creature card from your graveyard on top of your library.",
					TargetCardInGraveyard("target creature card from your graveyard", Creature(), YouOwn()),
					func(_ *game.StackItem, ctx *Context, occ int) error {
						t, ok := ModeTarget(ctx, occ)
						if !ok || t.Kind != game.TargetCard {
							return nil
						}
						return ctx.Game.TuckToLibraryForEffect(t.ID, false)
					}),
				ModeDoing("Surveil 1.", nil,
					func(item *game.StackItem, ctx *Context, _ int) error {
						return Surveil{Player: item.Controller, N: 1}.Apply(ctx)
					}),
			),
		}},
	})
}
