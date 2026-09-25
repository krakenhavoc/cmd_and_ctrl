package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bojuka Bog — Land (EDHREC rank 24):
//
//	"This land enters tapped. When this land enters, exile target
//	player's graveyard. {T}: Add {B}."
//
// A land that is also the format's cheapest graveyard hate: it
// costs no card and no mana, which is why it is in essentially
// every black deck.
//
// Enters-tapped is a real CR 614 self-replacement (SelfEntersTapped):
// the land is never untapped on the battlefield and no tap event is
// emitted. It used to be a hook that tapped the land a beat after
// entry, written when a catalog replacement could not see its own
// source's entry; the entering-card block in
// gatherActiveReplacementsLocked has made that possible since the
// Temple cycle (#360, #578).
//
// The ETB is a real "target player", so it goes through Triggered
// rather than the entry hook: the target is chosen when the trigger goes on
// the stack, which is what makes the graveyard emptied the one that
// existed at resolution rather than at entry. Mandatory, like
// Ravenous Chupacabra — there is always a legal player, including
// its own controller, so it never fizzles for want of a target.
func init() {
	Register(Spec{
		OracleID:     "04b7362d-0490-4cb0-b5d7-2a7732f659ce",
		Name:         "Bojuka Bog",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Targets: TargetPlayer("target player"),
			Key:     "Bojuka Bog — exile target player's graveyard",
			Effect:  exileTargetPlayersGraveyard,
		}},
	})
}

// exileTargetPlayersGraveyard is the whole body of "exile target
// player's graveyard" — Bojuka Bog's trigger and Angel of Finality's,
// which are the same sentence on a land and on a 3/4 flier.
//
// The player is read off the item's first target rather than
// recomputed, because the clause targets: the graveyard emptied is the
// one that player has at RESOLUTION, and a player who mills themselves
// in response loses those cards too.
func exileTargetPlayersGraveyard(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
		return nil
	}
	return exileGraveyardForEffect(g, item, item.Targets[0].ID)
}

// exileGraveyardForEffect exiles every card in one player's
// graveyard.
//
// The instance IDs are snapshotted BEFORE the first exile: each
// ExileTarget removes a card from the pile, so ranging over the
// live slice would skip every other card and leave half the
// graveyard behind — a bug that would look like "Bojuka Bog only
// half works" rather than like an iteration error.
func exileGraveyardForEffect(g *game.Game, item *game.StackItem, playerID uuid.UUID) error {
	p := g.PlayerByIDForEffect(playerID)
	if p == nil {
		return nil
	}
	ids := make([]uuid.UUID, 0, len(p.Graveyard.Cards))
	for _, c := range p.Graveyard.Cards {
		ids = append(ids, c.InstanceID)
	}
	ctx := NewContext(g, item)
	for _, id := range ids {
		if err := (ExileTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
