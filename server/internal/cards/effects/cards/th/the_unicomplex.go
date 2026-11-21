package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "b38fcf9a-464a-432a-b13c-9a206d52dcd9",
		Name:         "The Unicomplex",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.EntersTappedUnless(effects.OtherLandsAtLeast(2))},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: effects.SelfEnteredUntapped,
			Key:       "The Unicomplex — choose one",
			Effect:    func(*game.Game, *game.StackItem) error { return nil },
			Modes: effects.ChooseOne(
				effects.ModeDoing("Put target creature card from your graveyard on top of your library.",
					effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn()),
					effects.TuckTheModesTargetOnTopOfItsOwnersLibrary),
				effects.ModeDoing("Surveil 1.", nil,
					func(item *game.StackItem, ctx *effects.Context, _ int) error {
						return effects.Surveil{Player: item.Controller, N: 1}.Apply(ctx)
					}),
			),
		}},
	})
}
