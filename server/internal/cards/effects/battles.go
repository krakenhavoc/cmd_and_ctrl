package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// battles.go — the card-facing half of S27's battle support
// (CR 310). The lifecycle is engine-side and keys off the card type:
// defense counters on entry, the protector prompt, and the CR 704.5p
// sweep at zero defense are all in server/internal/game/battle.go and
// apply to a battle the catalog has never heard of.
//
// What a card file declares here is:
//
//	BattleSpec     the printed defense, as a FALLBACK for cards that
//	               never go through deck import
//	DefeatedTrigger the "when this Siege is defeated" ability
//
// A battle's own ETB trigger is an ordinary EventETB TriggeredAbility
// and needs nothing special.

// BattleSpec is a battle's printed battle data — the catalog's
// fallback copy of what the deck importer stamps onto
// game.Card.StartingDefense from Scryfall.
//
// LEAVE IT ALONE FOR A REAL CARD. Defense is printed card data like
// power, toughness and starting loyalty, not card-effect data: the
// importer parses Scryfall's per-face `defense` and the engine stamps
// the counters on every battlefield entry, catalog entry or not. This
// slot exists for the cards that never go through import — tokens,
// fixtures, the demo seed — exactly as Spec.StartingLoyalty does
// after #274. Setting it on a card a player can actually own is
// redundant at best and a second source of truth at worst.
//
// It is declared, and the Sieges in this batch do set it, because a
// battle is a new enough card type that a reader deserves to see the
// number next to the rules text — and because that makes the fallback
// path itself testable without a Scryfall fixture.
type BattleSpec struct {
	// Defense is the printed defense (CR 310.4) — the number of
	// defense counters the battle enters with.
	Defense int

	// Subtype is the battle's printed subtype: "Siege" today, and
	// nothing else has been printed. Carried so a future
	// "battles you control" or per-subtype rule has something to
	// read that is not a substring of the type line.
	Subtype string
}

// BattleSubtypeSiege is the only printed battle subtype (CR 310.2).
// A Siege is the one that chooses a protector and that transforms
// when defeated.
const BattleSubtypeSiege = "Siege"

// DefeatedTrigger declares a battle's "when this is defeated"
// ability (CR 310.9) — the last defense counter has come off and the
// battle is about to leave the battlefield.
//
// It is an ordinary triggered ability on an ordinary event, so it
// uses the stack and can be responded to. What is NOT ordinary is
// the timing: the engine announces the defeat from the state-based
// action pass, immediately BEFORE the CR 704.5p move puts the battle
// in the graveyard. So Build runs with the battle still on the
// battlefield and the Effect runs after it has left — which is why
// the effect must read the battle by id off the item rather than
// assuming where it is.
func DefeatedTrigger(label string, effect func(g *game.Game, item *game.StackItem) error) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventBattleDefeated},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return ev.CardID == source.InstanceID
		},
		Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			return game.NewTriggeredItem(source, label, effect)
		},
	}
}

// SiegeDefeated is the reminder-text half every printed Siege shares:
// "exile it, then cast it transformed" (CR 310.9c).
//
// SANDBOX SIMPLIFICATION, WEAKER THAN PRINTED, and the largest one in
// S27: only the EXILE happens. The battle leaves the battlefield and
// goes to exile instead of the graveyard; the free transformed cast
// of its back face does not follow.
//
// The seam, named precisely so the follow-up is a small change rather
// than a rediscovery. Casting the back face needs two things the
// multi-face model (ADR 0034) does not offer yet:
//
//  1. A per-INSTANCE face on the grant. CastableFaces returns front
//     only for a `transform` layout (CR 712.4 — the back is reached
//     by transforming, not by casting), and CastSpell calls
//     SetFace(params.Face) with a default of 0, so an exiled card
//     pre-flipped to face 1 is flipped straight back at announce.
//     ExilePlayPermission would need a Face, and the exile branch of
//     CastSpell would need to read it before the faceCastable gate.
//  2. faceOnResolve returning that face. It returns 0 for every
//     non-MDFC layout today, so even a back-face cast would resolve
//     into a front-face permanent — a battle re-entering the
//     battlefield, which is worse than nothing.
//
// Until both land, exiling is the half that is unambiguously right,
// and losing the back face is the weaker direction (#259). Doing it
// the other way — leaving the battle in the graveyard so a
// reanimation effect could get at it — would be stronger than
// printed, which is the direction that is never acceptable.
func SiegeDefeated() func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		battleID := item.SourceCardID
		if battleID == uuid.Nil {
			return nil
		}
		// The battle is already in its owner's graveyard by the time
		// this resolves: the defeat is announced from the SBA pass,
		// and the CR 704.5p move runs in the same pass, before the
		// trigger drains onto the stack. ExileCardForEffect finds it
		// wherever it is.
		return g.ExileCardForEffect(battleID)
	}
}
