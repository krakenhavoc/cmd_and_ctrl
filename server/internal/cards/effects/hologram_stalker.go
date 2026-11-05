package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hologram Stalker — {4}{G}{G} Creature — Hirogen Warrior, 2/2 (Star
// Trek, #4973):
//
//	"When this creature enters, return target permanent card from your
//	 graveyard to your hand.
//	 Whenever a card leaves your graveyard, create a 3/3 colorless
//	 Illusion artifact creature token. This ability triggers only once
//	 each turn."
//
// The enters trigger is Evolution Witness's return; the target is chosen
// as it goes on the stack (CR 603.3d), so with no permanent card in the
// graveyard it is removed. The second ability watches every route a card
// leaves its controller's graveyard by (leftGraveyardEvents: an exile, a
// return, a cast from it), and "only once each turn" is the per-turn
// trigger tally of this object (b11TriggeredThisTurn); OncePerBatch keeps
// cards leaving together from getting past the tally twice. The enters
// trigger's own return is a card leaving, so it makes the first Illusion.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8f91e7aa-a2d4-42ed-a959-3acd621679a5",
		Name:         "Hologram Stalker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Hologram Stalker — return target permanent card from your graveyard to your hand",
				returnTargetGraveyardCardToHand),
				TargetCardInGraveyard("target permanent card from your graveyard", Permanent(), YouOwn())),
			OncePerBatch(OnAny(leftGraveyardEvents, hologramStalkerCardLeft, hologramStalkerIllusionLabel,
				Do(CreateToken{Template: TokenCard("3/3 colorless Illusion artifact"), N: 1}))),
		},
	})
}

const hologramStalkerIllusionLabel = "Hologram Stalker — create a 3/3 colorless Illusion artifact creature token"

// hologramStalkerCardLeft is "a card leaves your graveyard", while this
// ability has not been put on the stack yet this turn.
func hologramStalkerCardLeft(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return cardLeftYourGraveyard(ev, source, g) && !b11TriggeredThisTurn(g, source.InstanceID, hologramStalkerIllusionLabel)
}
