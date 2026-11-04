package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b32_helpers.go — shared pieces for the thirty-second Secrets of
// Strixhaven batch (#4456).

// socB32EachPlayerLosesHalfTheirLife is "each player loses half their life,
// rounded down" (Pox Plague): the caster first, then each opponent in seat
// order, each a life change that names the item's source and no damage. A
// player at zero or less loses nothing.
func socB32EachPlayerLosesHalfTheirLife(ctx *Context) error {
	for _, id := range tablePlayers(ctx) {
		p := ctx.PlayerByID(id)
		if p == nil || p.Eliminated || p.Life <= 0 {
			continue
		}
		if half := p.Life / 2; half > 0 {
			if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), id, -half); err != nil {
				return err
			}
		}
	}
	return nil
}

// socB32PutCounterOnEachCreaturePlayerControls is "put a +1/+1 counter on each
// creature target player controls" (Practiced Offense). The set is read before
// the first counter goes on, so a creature a counter-placement trigger makes
// mid-loop gets none.
func socB32PutCounterOnEachCreaturePlayerControls(ctx *Context, player uuid.UUID) error {
	var ids []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
			continue
		}
		if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}

// socB32YouCastASpellThat is the AppliesTo of a permanent that gives a keyword
// to spells its controller casts (GrantsCascade, socB32GrantsStorm): the cast
// event is the controller's own, and `when` (nil admits every spell) judges the
// spell, looked up off the stack, against the granting permanent.
func socB32YouCastASpellThat(when func(spell game.Card, source *game.Card, g *game.Game) bool) func(game.Event, *game.Card, game.Characteristic, *game.Game) bool {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if ev.Actor != source.Controller {
			return false
		}
		spell, ok := g.LookupCardForEffect(ev.CardID)
		if !ok {
			return false
		}
		return when == nil || when(spell, source, g)
	}
}

// socB32GrantsStorm builds the trigger for a permanent that gives storm to
// spells its controller casts (Prismari, the Inspiration: "Instant and
// sorcery spells you cast have storm"), the storm sibling of GrantsCascade.
//
// A spell that has storm gets the keyword's trigger when it is cast
// (CR 702.40a), so the grant is a battlefield trigger watching the cast; each
// grantor is its own instance (CR 702.40b), and a spell that also prints storm
// triggers once more. The item names the grantor as its source (its catalog
// row, ADR 0041 P9) and carries the spell as Params.Object, because the copies
// are copies of the spell, not of the permanent. The copy count is read as the
// item resolves (stormCopies), and the spell is copied from last-known
// information if it has left the stack (CR 608.2h).
func socB32GrantsStorm(label string, when func(spell game.Card, source *game.Card, g *game.Game) bool) game.TriggeredAbility {
	return game.TriggeredAbility{
		Keyword:   KeywordStorm,
		Key:       label,
		Watches:   []game.EventKind{game.EventCast},
		AppliesTo: socB32YouCastASpellThat(when),
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label+" — storm")
			item.Controller, item.Owner = ev.Actor, ev.Actor
			item.Params.Object = game.ObjectRef{ID: ev.CardID}
			return item
		},
		Effect: func(g *game.Game, item *game.StackItem) error {
			return stormCopies(g, item, item.Params.Object.ID)
		},
	}
}
