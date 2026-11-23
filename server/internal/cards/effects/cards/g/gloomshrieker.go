package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gloomshrieker — Enchantment Creature — Cat Beast {1}{B}{G}, 2/1
// (Kamigawa: Neon Dynasty, #5493):
//
//	"Menace
//	 When this creature enters, return target permanent card from your
//	 graveyard to your hand.
//	 If this creature would die, exile it instead."
//
// The enters trigger is Coati Scavenger's graveyard-to-hand over a
// permanent card (any card type that can be a permanent, a token never
// sits there); the target is chosen as the trigger goes on the stack, so
// the Gloomshrieker's own earlier death could not be it. The replacement
// is Wildwood Escort's: a battlefield-to-graveyard move of this creature
// becomes a move to exile (ADR 0178).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b58737a4-180e-4e2f-95ae-afd6d446d608",
		Name:            "Gloomshrieker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
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
			Label:    "Gloomshrieker: exile instead of dying",
			RowLabel: "If this creature would die, exile it instead.",
		}},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Gloomshrieker — return target permanent card from your graveyard to your hand", effects.ReturnTargetGraveyardCardToHand),
				effects.TargetCardInGraveyard("target permanent card from your graveyard", effects.Permanent(), effects.YouOwn())),
		},
	})
}
