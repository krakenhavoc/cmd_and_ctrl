package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b166_helpers.go — shared vocabulary of Modern Horizons batch 166
// (S82 older sets, #5519).

// OldB166WardForYouAndYourPermanents is "Whenever you or a permanent you
// control becomes the target of a spell or ability an opponent controls,
// counter that spell or ability unless its controller pays <cost>"
// (Unsettled Mariner): ward's pay-or-counter body (CR 702.21a) over the source's
// controller and every permanent they control, the source included.
func OldB166WardForYouAndYourPermanents(cost WardCost, label string) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventBecomesTarget},
		Key:     label,
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			if ev.StackItemID == uuid.Nil || ev.Actor == uuid.Nil || ev.Actor == source.Controller {
				return false
			}
			if ev.CardID == uuid.Nil {
				return ev.Target == source.Controller
			}
			c, ok := g.LookupCardForEffect(ev.CardID)
			return ok && g.Battlefield.Contains(ev.CardID) && c.Controller == source.Controller
		},
		Effect: WardEffect(cost),
	}
}

// OldB166TapAndFreezeDamagedCreature is "tap that creature and it doesn't untap
// during its controller's next untap step", about the creature the source dealt
// combat damage to: the object stamped on the item as the trigger was built
// (CR 400.7), and only while it is still that object on the battlefield.
func OldB166TapAndFreezeDamagedCreature(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ref := item.Params.Object
		c, ok := g.LookupCardForEffect(ref.ID)
		if !ok || c.ObjectEpoch != ref.Epoch {
			return nil
		}
		if z := g.FindCardZoneForEffect(ref.ID); z == nil || z.Kind != game.ZoneBattlefield {
			return nil
		}
		return TapAndFreeze{Targets: []uuid.UUID{ref.ID}, Label: label}.Apply(NewContext(g, item))
	}
}

// OldB166YouControlAnotherSnowPermanent is Blizzard Strix's intervening "if"
// (CR 603.4): its controller controls a snow permanent other than the source.
func OldB166YouControlAnotherSnowPermanent(g *game.Game, you, source uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == you && c.InstanceID != source && c.HasSupertype("snow") {
			return true
		}
	}
	return false
}

// OldB166IfYouControlAnotherSnowPermanent narrows a trigger of the source to
// "if you control another snow permanent": read as it triggers and again as it
// resolves.
func OldB166IfYouControlAnotherSnowPermanent(t game.TriggeredAbility) game.TriggeredAbility {
	when := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return source != nil && OldB166YouControlAnotherSnowPermanent(g, source.Controller, source.InstanceID) &&
			(when == nil || when(ev, source, lki, g))
	}
	return InterveningIf(CondBoard{Name: "you control another snow permanent", Test: func(ctx *Context) bool {
		return OldB166YouControlAnotherSnowPermanent(ctx.Game, ctx.Controller(), ctx.Item.SourceCardID)
	}}, t)
}

// OldB166ArtifactCardsInGraveyard counts the artifact cards in a player's
// graveyard (Mirrodin Besieged's fifteen).
func OldB166ArtifactCardsInGraveyard(g *game.Game, player uuid.UUID) int {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return 0
	}
	n := 0
	for _, c := range p.Graveyard.Cards {
		if c.IsArtifact() {
			n++
		}
	}
	return n
}
