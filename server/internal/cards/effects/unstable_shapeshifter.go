package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Unstable Shapeshifter — Creature — Shapeshifter {3}{U}, 0/1:
//
//	"Whenever another creature enters, this creature becomes a copy of
//	 that creature, except it has this ability."
//
// A duration copy with NO stated duration (CR 611.2a, #1593): it lasts
// until the Shapeshifter leaves the battlefield or copies the next
// creature. Each new copy makes the last one invisible for as long as
// the permanent lasts, so the engine keeps only the newest record
// (game.BecomeCopyForEffect) rather than one per creature that ever
// entered.
//
// "Except it has this ability" is a GRANT in the copiable values
// (CR 707.9a), declared as the bundle below and named from the except
// clause — so the trigger survives every copy, and a Clone copying the
// Shapeshifter-as-a-Bear is a Bear with the trigger too. The card's
// own printed trigger and the granted one are the same TriggeredAbility
// value, so they cannot drift.
//
// The copy is of the creature as it is on the battlefield when the
// trigger resolves. If it has left by then, the copy is of the card
// itself — the same thing, except for a creature that was itself a copy
// (a Clone), whose last-known copiable values the engine does not
// keep. That is the caveat.
const unstableShapeshifterGrant = "unstable-shapeshifter/this-ability"

var unstableShapeshifterTrigger = On(game.EventETB,
	func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if ev.CardID == uuid.Nil || ev.CardID == source.InstanceID {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && c.IsCreature()
	},
	"Unstable Shapeshifter — becomes a copy of that creature",
	func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		self := ctx.Source()
		entered := ctx.Trigger().Event.CardID
		copyOf := BecomeCopy{
			Targets:  []uuid.UUID{self},
			Of:       entered,
			Duration: CopyIndefinite,
			Except:   func(v *game.PrintedValues) { v.GrantAbility(unstableShapeshifterGrant) },
			Label:    "Unstable Shapeshifter — becomes a copy",
		}
		if info, ok := ctx.TriggeringPermanent(); !ok || info.Left {
			card, found := g.LookupCardForEffect(entered)
			if !found {
				return nil
			}
			own := game.OwnPrintedValues(card)
			copyOf.Values = &own
		}
		return copyOf.Apply(ctx)
	})

func init() {
	Register(Spec{
		OracleID:     "6a71deb3-6659-47ea-8687-136cf75fa381",
		Name:         "Unstable Shapeshifter",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"If the creature that entered was itself a copy of something and leaves the battlefield before the ability resolves, this becomes a copy of the card rather than of what that creature was copying."},
		Grants: []AbilityGrant{{
			Key:       unstableShapeshifterGrant,
			Text:      "Whenever another creature enters, this creature becomes a copy of that creature, except it has this ability.",
			Triggered: []game.TriggeredAbility{unstableShapeshifterTrigger},
		}},
		Triggered: []game.TriggeredAbility{unstableShapeshifterTrigger},
	})
}
