package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Ur-Sphinx — Legendary Creature — Sphinx Avatar {6}{W}{U}{B}, 10/10
// (Reality Fracture Commander, tracker #2795, #2797):
//
//	"Eminence — As long as The Ur-Sphinx is in the command zone or on
//	 the battlefield, other Sphinx spells you cast cost {1} less to cast.
//	 Flying
//	 Whenever one or more Sphinxes you control attack, each player mills
//	 that many cards. For each player, you may cast a card that player
//	 milled this way without paying its mana cost."
//
// # Eminence
//
// The first command-zone static (ADR 0140): the discount is a cost
// modifier declared with Eminence(...), so the cast pricer also gathers it
// from its owner's command zone, bound with that player as "you". It
// applies from nowhere else — not a hand, library, graveyard or exile —
// and not to The Ur-Sphinx itself ("other"). A reduction spends against
// generic mana only (CR 601.2f), so {W}{U}{B} Sphinx costs are untouched.
//
// # The attack trigger
//
// "Whenever one or more Sphinxes you control attack" is one trigger per
// declaration (OncePerBatch; the engine emits EventAttack per creature).
// "That many" is the number of Sphinxes declared in that batch, counted
// from the events so a Sphinx that has left combat by resolution still
// counts. Every player mills it, starting with the controller (CR 101.4
// would order a simultaneous choice, but a mill makes none). Then, for
// each player, the controller picks up to one nonland card from the ones
// that player milled THIS WAY and is granted the free cast.
//
// The free cast is a grant, the engine's shape for every "cast it without
// paying its mana cost" (Malcolm, cascade, madness; game/cascade.go): the
// chosen card's own object gets a per-instance cast permission priced at
// {0}, flash-timed, that expires at end of turn, rather than a cast inside
// the trigger's resolution. The consequences are Malcolm's: there is a
// response window paper does not have, the cast may be held until later in
// the same turn, and it is the controller who casts. Lands are not offered
// (a land can't be cast). A card a replacement exiled instead of milling is
// not "milled this way" and is not offered (CR 400.7).
//
// No simplification beyond the shared grant shape.
func init() {
	Register(Spec{
		OracleID:        "4a3fdb8e-4699-4bd9-84e6-3cc7fea0e1ef",
		Name:            "The Ur-Sphinx",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CostModifiers: []game.CostModifier{
			Eminence(CostsLess(1, "Eminence — other Sphinx spells you cast cost {1} less to cast.",
				YourSpell(), OtherSpellOfCreatureType("Sphinx"))),
		},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller) && attackerHasSubtype(g, ev, "Sphinx")
			}, urSphinxTriggerLabel, urSphinxMillAndCast)),
		},
	})
}

const urSphinxTriggerLabel = "The Ur-Sphinx — each player mills that many cards; you may cast a card each milled this way"

// attackerHasSubtype reports whether the creature the attack event
// declares has the creature type, as it is now (changelings count,
// CR 702.73a). The Ur-Sphinx's Sphinxes and The Ur-Dragon's Dragons.
func attackerHasSubtype(g *game.Game, ev game.Event, subtype string) bool {
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && c.HasSubtype(subtype)
}

// attackersOfSubtypeInTheSameBatch is "that many" for an "whenever one
// or more [type] you control attack" trigger: the number of creatures
// of the type `you` declared as attackers in the event batch ev belongs
// to. Read off the log rather than the board, so one that died or left
// combat before the trigger resolved still counts.
//
// Caller must hold g.mu.
func attackersOfSubtypeInTheSameBatch(g *game.Game, ev game.Event, you uuid.UUID, subtype string) int {
	n := 0
	for i := len(g.Events) - 1; i >= 0; i-- {
		e := g.Events[i]
		if e.Batch < ev.Batch {
			break
		}
		if e.Batch == ev.Batch && attackDeclaredByYou(e, you) && attackerHasSubtype(g, e, subtype) {
			n++
		}
	}
	return n
}

// urSphinxMillAndCast is the trigger's body: every player mills N, then
// the controller picks a free cast from each milled pile in turn.
func urSphinxMillAndCast(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	n := attackersOfSubtypeInTheSameBatch(g, item.Trigger.Event, item.Controller, "Sphinx")
	if n <= 0 {
		return nil
	}
	ctx := NewContext(g, item)
	players := tablePlayers(ctx)
	milledBy := make([][]uuid.UUID, len(players))
	var millFrom func(i int) error
	millFrom = func(i int) error {
		if i >= len(players) {
			return urSphinxOfferCasts(NewContext(ctx.Game, item), players, milledBy, 0)
		}
		return MillToZone{Player: players[i], N: n, Then: func(c *Context, milled []uuid.UUID) error {
			milledBy[i] = milled
			return millFrom(i + 1)
		}}.Apply(NewContext(ctx.Game, item))
	}
	return millFrom(0)
}

// urSphinxOfferCasts walks the players in order and, for each, asks the
// controller for up to one nonland card that player milled this way,
// then grants the free cast. The next player is asked once the previous
// answer is in.
func urSphinxOfferCasts(ctx *Context, players []uuid.UUID, milledBy [][]uuid.UUID, i int) error {
	g, item := ctx.Game, ctx.Item
	for ; i < len(players); i++ {
		var offered []uuid.UUID
		for _, id := range milledBy[i] {
			// Still in that player's graveyard, and a card that can be cast.
			if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
				continue
			}
			if c, ok := g.LookupCardForEffect(id); ok && !c.IsLand() {
				offered = append(offered, id)
			}
		}
		if len(offered) == 0 {
			continue
		}
		owner, next := players[i], i+1
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    item.Controller,
			FromPlayer: owner,
			Source:     item.SourceCardID,
			Question:   "The Ur-Sphinx — you may cast a card milled this way without paying its mana cost",
			Cards:      offered,
			Min:        0,
			Max:        1,
			Zone:       game.ZoneGraveyard,
			Then: func(g *game.Game, picked []uuid.UUID) error {
				if len(picked) > 0 {
					g.GrantCastPermissionOverCardForEffect(picked[0], game.CastPermission{
						Player:   item.Controller,
						Zone:     game.ZoneGraveyard,
						Cost:     "{0}",
						CastOnly: true,
						Timing:   game.TimingFlash,
						Label:    "The Ur-Sphinx — cast it without paying its mana cost",
					})
				}
				return urSphinxOfferCasts(NewContext(g, item), players, milledBy, next)
			},
		})
		return nil
	}
	return nil
}
