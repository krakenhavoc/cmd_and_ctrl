package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wheel of Sun and Moon — Enchantment — Aura {G/W}{G/W}:
//
//	"Enchant player
//	 If a card would be put into enchanted player's graveyard from anywhere,
//	 instead that card is revealed and put on the bottom of that player's
//	 library."
//
// A graveyard replacement (CR 614.1a), Rest in Peace's shape with the
// destination changed: it watches every way into the enchanted player's
// graveyard (a death, a discard, a mill, a spell finishing, a countered spell),
// and only a CARD (a token is never put into a graveyard, CR 704.5d). The
// graveyard is the card's owner's, so a card the enchanted player owns is
// affected wherever it came from and another player's is not. The card is
// revealed from the zone it was leaving (CR 701.20b, nothing moves for a
// reveal) and lands on the BOTTOM of its owner's library without a shuffle
// (ReplacementEvent.PutOnLibraryBottom, which is the new seam this card
// brought); it never reaches the graveyard, so nothing that watches a graveyard
// sees it. Against another replacement of the same move (Rest in Peace's exile)
// the affected card's owner chooses which applies (CR 616.1).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6b6e4ee2-52e6-452b-af15-c90eab5fc746",
		Name:         "Wheel of Sun and Moon",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPlayer(),
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove, game.EventDiscardCard},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if src == nil || (ev.Kind != game.RepEventMove && ev.Kind != game.RepEventDiscard) || ev.NewZone != game.ZoneGraveyard {
					return false
				}
				player, ok := wheelOfSunAndMoonPlayer(src)
				if !ok {
					return false
				}
				moving, ok := g.LookupCardForEffect(ev.CardID)
				if !ok || moving.IsToken() {
					return false
				}
				graveyard := ev.NewZoneOwner
				if graveyard == uuid.Nil {
					graveyard = moving.Owner
				}
				return graveyard == player
			},
			Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
				player, _ := wheelOfSunAndMoonPlayer(src)
				g.RevealForEffect(game.RevealSpec{
					Player: player,
					Source: src.InstanceID,
					Reason: "Wheel of Sun and Moon — revealed and put on the bottom of its owner's library",
					Cards:  []uuid.UUID{ev.CardID},
				})
				ev.NewZone = game.ZoneLibrary
				ev.NewZoneOwner = player
				ev.PutOnLibraryBottom = true
				return nil
			},
			Controller: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) uuid.UUID {
				if player, ok := wheelOfSunAndMoonPlayer(src); ok {
					return player
				}
				return src.Controller
			},
			Label: "Wheel of Sun and Moon: revealed and put on the bottom of its owner's library instead",
		}},
	})
}

// wheelOfSunAndMoonPlayer is the enchanted player: the Aura must be attached
// to a player.
func wheelOfSunAndMoonPlayer(aura *game.Card) (uuid.UUID, bool) {
	if aura.AttachedTo.Kind != game.TargetPlayer || aura.AttachedTo.ID == uuid.Nil {
		return uuid.Nil, false
	}
	return aura.AttachedTo.ID, true
}
