package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wildwood Escort — Creature — Elf Warrior {4}{G}, 3/3 (March of the
// Machine):
//
//	"When this creature enters, return target creature or battle card
//	 from your graveyard to your hand.
//	 If this creature would die, exile it instead."
//
// The enters trigger is Along the Crooked Way's return. The static is a
// CR 614 replacement on this creature moving from the battlefield to the
// graveyard: it is exiled instead, so it never dies and its dies-triggers
// do not fire (a token is exiled too).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ce807ee3-a27c-4f3b-92a4-37faaca42aa1",
		Name:         "Wildwood Escort",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return src != nil && ev.Kind == game.RepEventMove && ev.CardID == src.InstanceID &&
					ev.OldZone == game.ZoneBattlefield && ev.NewZone == game.ZoneGraveyard
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.NewZone = game.ZoneExile
				ev.NewZoneOwner = uuid.Nil
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label:    "Wildwood Escort: exile instead of dying",
			RowLabel: "If this creature would die, exile it instead.",
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Wildwood Escort — return target creature or battle card from your graveyard to your hand",
					func(g *game.Game, item *game.StackItem) error {
						return returnLegalGraveyardTargetsToHand(NewContext(g, item))
					}),
				TargetCardInGraveyard("target creature or battle card from your graveyard", Or(Creature(), Battle()), YouOwn()),
			),
		},
	})
}
