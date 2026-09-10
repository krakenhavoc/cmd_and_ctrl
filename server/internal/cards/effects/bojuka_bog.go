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
// Enters-tapped follows the Worn Powerstone / Azorius Chancery
// pattern — OnETB taps the permanent as it lands, a beat later than
// a true CR 614 replacement, because a catalog replacement cannot
// fire on its own source's entry (the entering card is not yet on
// the battlefield when the pipeline walks it).
//
// The ETB is a real "target player", so it goes through Triggered
// rather than OnETB: the target is chosen when the trigger goes on
// the stack, which is what makes the graveyard emptied the one that
// existed at resolution rather than at entry. Mandatory, like
// Ravenous Chupacabra — there is always a legal player, including
// its own controller, so it never fizzles for want of a target.
func init() {
	Register(Spec{
		OracleID: "04b7362d-0490-4cb0-b5d7-2a7732f659ce",
		Name:     "Bojuka Bog",
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
		OnETB: func(card *game.Card, ctx *Context) error {
			return TapTarget{Target: card.InstanceID}.Apply(ctx)
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Targets: TargetPlayer("target player"),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Bojuka Bog — exile target player's graveyard",
					func(g *game.Game, item *game.StackItem) error {
						if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
							return nil
						}
						return exileGraveyardForEffect(g, item, item.Targets[0].ID)
					})
			},
		}},
	})
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
