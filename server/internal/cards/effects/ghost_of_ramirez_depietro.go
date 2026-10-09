package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ghost of Ramirez DePietro — Legendary Creature — Spirit Pirate
// {2}{U}, 2/3 (#1112):
//
//	"Ghost of Ramirez DePietro can't be blocked by creatures with
//	 toughness 3 or greater.
//	 Whenever Ghost of Ramirez DePietro deals combat damage to a
//	 player, choose up to one target card in a graveyard that was
//	 discarded or put there from a library this turn. Put that card
//	 into its owner's hand.
//	 Partner"
//
// The evasion is a CR 509.1b block rule with a parameter
// (CantBeBlockedBy, ADR 0045): the blocker's toughness is read as the
// block is declared.
//
// The target clause is "any graveyard", narrowed by where the card
// came from this turn (arrivedFromHandOrLibraryThisTurn,
// graveyard_provenance.go): a card whose most recent arrival in a
// graveyard this turn was a discard, or a move out of a library — a
// mill, a surveil, an Entomb. A creature that died, a spell that
// resolved, or a card that was already there when the turn began is
// not a legal target. The predicate is asked when the trigger goes on
// the stack and again as it resolves (CR 608.2b).
//
// "Put that card into its owner's hand" moves it for its owner, not
// for the Ghost's controller.
//
// Partner is a deck-construction rule (CR 702.124h) that internal/deck
// reads off the oracle text (#2874), so the Ghost can be one of two
// partner commanders.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "07fe0fb5-cf34-4ba4-a3f3-ac0cc919bf91",
		Name:         "Ghost of Ramirez DePietro",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnSelf(), ToughnessGE(3), "creatures with toughness 3 or greater"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverThisDealsCombatDamageToAPlayer(
					"Ghost of Ramirez DePietro — put a card discarded or milled this turn into its owner's hand",
					returnFirstLegalGraveyardTargetToHand),
				TargetCardInGraveyard("up to one target card in a graveyard that was discarded or put there from a library this turn",
					arrivedFromHandOrLibraryThisTurn).WithCount(0, 1),
			),
		},
	})
}

// arrivedFromHandOrLibraryThisTurn is the Ghost's target predicate as a
// CardPredicate over a card in a graveyard.
func arrivedFromHandOrLibraryThisTurn(g *game.Game, _ uuid.UUID, c game.Card) bool {
	return discardedOrMilledThisTurn(g, c.InstanceID)
}
