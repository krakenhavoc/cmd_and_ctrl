package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lazav, Dimir Mastermind — Legendary Creature — Shapeshifter
// {U}{U}{B}{B}, 3/3:
//
//	"Hexproof
//	 Whenever a creature card is put into an opponent's graveyard from
//	 anywhere, you may have Lazav become a copy of that card, except its
//	 name is Lazav, Dimir Mastermind, it's legendary in addition to its
//	 other types, and it has hexproof and this ability."
//
// A duration copy with no stated duration (#1593, become_copy.go): it
// lasts until Lazav leaves the battlefield or copies the next card, and
// the engine keeps only the newest record.
//
// The trigger watches the four events that put a card into a graveyard
// — the same set, and the same "is it now in an opponent's graveyard"
// question, as Bloodchief Ascension (b06CardInOpponentsGraveyard) — so a
// creature that dies, a creature card milled or discarded, and a
// countered creature spell all count, each once.
//
// The copy is of the CARD, read as it resolves. A card that has left
// the graveyard by then is still copied (CR 608.2h): its values are its
// own printed ones, which game.OwnPrintedValues reads whatever it has
// become since.
//
// The except clause is four edits to the copiable values (CR 707.9):
// the name, the Legendary supertype, the hexproof keyword, and a grant
// of this trigger as a named bundle, so a Lazav copying a Grizzly Bears
// is a legendary Grizzly Bears named Lazav that keeps watching.
const lazavDimirMastermindGrant = "lazav-dimir-mastermind/this-ability"

var lazavDimirMastermindTrigger = game.TriggeredAbility{
	Watches: []game.EventKind{game.EventZoneMove, game.EventDiscardCard, game.EventMill, game.EventCounterSpell},
	AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if _, ok := b06CardInOpponentsGraveyard(ev, source.Controller, g); !ok {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && c.IsCreature()
	},
	OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Lazav, Dimir Mastermind — become a copy of that card?"},
	Key:            "Lazav, Dimir Mastermind — becomes a copy of that card",
	Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, "Lazav, Dimir Mastermind — becomes a copy of that card")
		item.Params.Object = game.ObjectRef{ID: ev.CardID}
		return item
	},
	Effect: func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		self := ctx.Source()
		card, found := g.LookupCardForEffect(item.Params.Object.ID)
		if !found {
			return nil
		}
		own := game.OwnPrintedValues(card)
		return BecomeCopy{
			Targets:  []uuid.UUID{self},
			Of:       card.InstanceID,
			Values:   &own,
			Duration: CopyIndefinite,
			Except: func(v *game.PrintedValues) {
				v.SetName("Lazav, Dimir Mastermind")
				v.AddSupertype("Legendary")
				v.AddKeyword("hexproof")
				v.GrantAbility(lazavDimirMastermindGrant)
			},
			Label: "Lazav, Dimir Mastermind — becomes a copy",
		}.Apply(ctx)
	},
}

func init() {
	Register(Spec{
		OracleID:        "8027a610-613e-4640-9840-c8778694f312",
		Name:            "Lazav, Dimir Mastermind",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
		Grants: []AbilityGrant{{
			Key:       lazavDimirMastermindGrant,
			Text:      "Whenever a creature card is put into an opponent's graveyard from anywhere, you may have Lazav become a copy of that card, except its name is Lazav, Dimir Mastermind, it's legendary in addition to its other types, and it has hexproof and this ability.",
			Triggered: []game.TriggeredAbility{lazavDimirMastermindTrigger},
		}},
		Triggered: []game.TriggeredAbility{lazavDimirMastermindTrigger},
	})
}
