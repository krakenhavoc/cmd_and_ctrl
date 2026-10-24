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
// The body is a step sequence (#4228, CR 608.2c): a mill step and an offer
// step for each seat, so a prompt a mill asks holds the next step behind it.
// The cards milled this way are not held in a slice; each offer step reads
// them back off the event log after a cursor the first step filed, and the
// pick's continuation is a registered key, so a table waiting on an offer is
// a restore point.
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
				return attackDeclaredByYou(ev, source.Controller) && attackerHasSubtype(g, ev, source.TextSubtype("Sphinx"))
			}, urSphinxTriggerLabel, urSphinxSteps)),
		},
	})
}

const urSphinxTriggerLabel = "The Ur-Sphinx — each player mills that many cards; you may cast a card each milled this way"

// attackerHasSubtype reports whether the creature the attack event
// declares has the creature type, as it is now (changelings count,
// CR 702.73a). The Ur-Sphinx's Sphinxes and The Ur-Dragon's Dragons.
// `subtype` is judged as given: a caller reads its printed word through
// its own text first (TextSubtype, CR 612.2, #3984).
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

// urSphinxRecord files the log's cursor from before the first mill and how
// many cards each player mills (#4228).
const urSphinxRecord = "ur-sphinx/mill"

// urSphinxSteps is the trigger's body (CR 608.2c, #4228): one step to count
// and file the cursor, a mill step for each seat, then an offer step for each
// seat. Each is its own step, so the walk stops behind a prompt one asks (a
// commander's CR 903.9 offer on a milled card, the free-cast pick) and a
// restore taken there still runs the rest. The cards a player milled this way
// are read back off the event log after the cursor, not held in a slice. A seat
// the table does not have makes its steps nothing.
var urSphinxSteps = urSphinxBuildSteps()

func urSphinxBuildSteps() Effect {
	steps := []Applier{Instruction(urSphinxCount)}
	for i := 0; i < game.MaxPlayers; i++ {
		steps = append(steps, Instruction(urSphinxMill(i)))
	}
	for i := 0; i < game.MaxPlayers; i++ {
		steps = append(steps, Instruction(urSphinxOffer(i)))
	}
	return Do(steps...)
}

// urSphinxCount counts the Sphinxes that attacked together, "that many", and
// files it with the log's cursor.
func urSphinxCount(ctx *Context) error {
	item := ctx.Item
	if item.Trigger == nil {
		return nil
	}
	g := ctx.Game
	n := attackersOfSubtypeInTheSameBatch(g, item.Trigger.Event, item.Controller, g.TextReaderForEffect(item).TextSubtype("Sphinx"))
	ctx.RecordStep(urSphinxRecord, nil, int(b25LastEventSeq(g)), n)
	return nil
}

// urSphinxMill is the mill of the i-th player at the table, the controller
// first.
func urSphinxMill(i int) func(ctx *Context) error {
	return func(ctx *Context) error {
		players := tablePlayers(ctx)
		r := ctx.StepRecord(urSphinxRecord)
		if i >= len(players) || len(r.Ns) < 2 || r.Ns[1] <= 0 {
			return nil
		}
		return MillToZone{Player: players[i], N: r.Ns[1]}.Apply(ctx)
	}
}

// urSphinxMilledThisWay is the cards `player` milled since the cursor that
// are still in their graveyard and can be cast (a land cannot), oldest first.
// A card a replacement exiled or sent to the command zone instead of milling
// has no EventMill and is not offered (CR 400.7).
func urSphinxMilledThisWay(g *game.Game, player uuid.UUID, cursor uint64) []uuid.UUID {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return nil
	}
	var out []uuid.UUID
	for i := len(g.Events) - 1; i >= 0; i-- {
		ev := g.Events[i]
		if ev.Seq <= cursor {
			break
		}
		if ev.Kind != game.EventMill || ev.Actor != player || ev.NewZone != game.ZoneGraveyard || !p.Graveyard.Contains(ev.CardID) {
			continue
		}
		if c, ok := g.LookupCardForEffect(ev.CardID); ok && !c.IsLand() {
			out = append(out, ev.CardID)
		}
	}
	for l, r := 0, len(out)-1; l < r; l, r = l+1, r-1 {
		out[l], out[r] = out[r], out[l]
	}
	return out
}

// urSphinxOffer asks the controller for up to one nonland card the i-th
// player milled this way, then grants the free cast. The pick's continuation
// is a registered key, so the table waiting on it is a restore point.
func urSphinxOffer(i int) func(ctx *Context) error {
	return func(ctx *Context) error {
		players := tablePlayers(ctx)
		r := ctx.StepRecord(urSphinxRecord)
		if i >= len(players) || len(r.Ns) < 2 || r.Ns[1] <= 0 {
			return nil
		}
		offered := urSphinxMilledThisWay(ctx.Game, players[i], uint64(r.Ns[0]))
		if len(offered) == 0 {
			return nil
		}
		ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    ctx.Controller(),
			FromPlayer: players[i],
			Source:     ctx.Item.SourceCardID,
			Question:   "The Ur-Sphinx — you may cast a card milled this way without paying its mana cost",
			Cards:      offered,
			Min:        0,
			Max:        1,
			Zone:       game.ZoneGraveyard,
			ThenKey:    urSphinxChosen,
		})
		return nil
	}
}

// urSphinxChosen grants the free cast over the card picked. The next
// player's offer is the next step, which the walk runs once this answer is in.
var urSphinxChosen = game.RegisterChooseCardsThen("choose-cards/the-ur-sphinx-cast", func(g *game.Game, r game.CardsChosen) error {
	if !r.Answered || len(r.Picked) == 0 {
		return nil
	}
	g.GrantCastPermissionOverCardForEffect(r.Picked[0], game.CastPermission{
		Player:   r.Chooser,
		Zone:     game.ZoneGraveyard,
		Cost:     "{0}",
		CastOnly: true,
		Timing:   game.TimingFlash,
		Label:    "The Ur-Sphinx — cast it without paying its mana cost",
	})
	return nil
})
