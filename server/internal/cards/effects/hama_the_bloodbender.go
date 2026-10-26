package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hama, the Bloodbender — Legendary Creature — Human Warlock
// {2}{U/B}{U/B}{U/B}, 3/3:
//
//	"When Hama enters, target opponent mills three cards. Exile up to one
//	 noncreature, nonland card from that player's graveyard. For as long
//	 as you control Hama, you may cast the exiled card during your turn
//	 by waterbending {X} rather than paying its mana cost, where X is its
//	 mana value. (While paying a waterbend cost, you can tap your
//	 artifacts and creatures to help. Each one pays for {1}.)"
//
// The mill comes first, then Hama's controller chooses up to one
// noncreature, nonland card from anywhere in that player's graveyard
// (not only the three milled; the pick does not target), and it is
// exiled. If it landed in exile, Hama's controller gets a cast permission
// over it: cast only, on their own turn (TimingYourTurnOnly, the card's
// own timing still applies), for as long as they control Hama
// (forAsLongAsYouControlThisWindow), and priced by waterbending its mana
// value (CastPermission.WaterbendManaValue, #3989): the waterbend {N} is
// the alternative cost (CR 118.9), N the mana value of the face being
// cast, paid with mana or with tapped artifacts and creatures, {1} each
// (CR 701.67a). A cost reduction or a tax applies to it (CR 118.9d), and
// the taps pay only the waterbend's {N}, so a kicker or a tax the spell
// picks up is paid with mana (CR 701.67b). A spell with {X} in its cost
// is cast with X 0 (CR 107.3b).
//
// The spell keeps any convoke or improvise it has, its own or granted
// (#4177, ADR 0073's 2026-10-19 amendment): the waterbend and that tap
// cost are chained on the one cast (game.TapPermanentsCost.Also), and
// the engine shares the tapped permanents between them. The waterbend's
// taps pay only its {N} (CR 701.67b); a convoke or improvise tap pays
// any other part of the total, a tax for instance (CR 702.51a, 702.126a).
// A waterbend of the spell's own (Waterbender's Restoration) is still
// owed beside it. No simplification remains.
func init() {
	Register(Spec{
		OracleID:     "ef80e174-78e8-412d-a6fe-89c7af746669",
		Name:         "Hama, the Bloodbender",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Hama — target opponent mills three cards; exile up to one noncreature, nonland card from their graveyard", hamaEnters),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}

func hamaEnters(g *game.Game, item *game.StackItem) error {
	opp := TargetedPlayer(NewContext(g, item))
	if g.PlayerByIDForEffect(opp) == nil {
		return nil
	}
	tail := NewContext(g, item).ItemTail(hamaMilled)
	tail.Carry = []uuid.UUID{opp}
	return g.MillToZoneKeyedForEffect(opp, 3, game.ZoneGraveyard, game.MillUntil{}, tail)
}

// hamaMilled, hamaPicked and hamaExiled are Hama's rest of the trigger as
// registered continuations (#4253 slice 9): the opponent rides the carry
// and the trigger rides the tail, so a table waiting on a milled card's
// prompt, the pick or the exile is a restore point.
var (
	hamaMilled = CardTail("hama/choose-from-graveyard", func(ctx *Context, o game.TailOutcome) error {
		if len(o.Tail.Carry) == 0 {
			return nil
		}
		return hamaChooseFromGraveyard(ctx.Game, ctx.Item, o.Tail.Carry[0])
	})
	hamaPicked = game.RegisterChooseCardsThen("hama/exile-pick", func(g *game.Game, r game.CardsChosen) error {
		if len(r.Picked) == 0 || r.Asked == nil {
			return nil
		}
		return hamaExile(g, r.Asked, r.Picked[0])
	})
	hamaExiled = CardTail("hama/grant-waterbend-cast", func(ctx *Context, o game.TailOutcome) error {
		hamaGrantCasts(ctx.Game, ctx.Item, o.IDs)
		return nil
	})
)

// hamaChooseFromGraveyard asks Hama's controller for up to one
// noncreature, nonland card in the opponent's graveyard.
func hamaChooseFromGraveyard(g *game.Game, item *game.StackItem, opp uuid.UUID) error {
	p := g.PlayerByIDForEffect(opp)
	if p == nil || p.Graveyard == nil {
		return nil
	}
	var candidates []uuid.UUID
	for _, c := range p.Graveyard.Cards {
		if !c.IsCreature() && !c.IsLand() {
			candidates = append(candidates, c.InstanceID)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    item.Controller,
		FromPlayer: opp,
		Source:     item.SourceCardID,
		Question:   "Hama — exile up to one noncreature, nonland card from that player's graveyard",
		Cards:      candidates,
		Min:        0,
		Max:        1,
		Zone:       game.ZoneGraveyard,
		ThenKey:    hamaPicked,
		Asked:      game.NewAskedItem(item),
		LiveZone:   game.ZoneGraveyard,
	})
	return nil
}

// hamaExile exiles the pick and, if it landed in exile, grants the cast.
func hamaExile(g *game.Game, item *game.StackItem, id uuid.UUID) error {
	return g.ExileCardsKeyedForEffect([]uuid.UUID{id}, NewContext(g, item).ItemTail(hamaExiled))
}

// hamaGrantCasts grants the waterbend cast of each card exiled this way.
func hamaGrantCasts(g *game.Game, item *game.StackItem, exiled []uuid.UUID) {
	for _, eid := range exiled {
		c, ok := g.LookupCardForEffect(eid)
		if !ok {
			continue
		}
		g.GrantCastPermissionToCardsForEffect(game.CastPermission{
			Player:             item.Controller,
			Zone:               game.ZoneExile,
			CastOnly:           true,
			Timing:             game.TimingYourTurnOnly,
			WaterbendManaValue: true,
			Duration:           forAsLongAsYouControlThisWindow(g, item),
			Source:             item.SourceCardID,
			SourceName:         "Hama, the Bloodbender",
			Label:              "You may cast it during your turn by waterbending its mana value, for as long as you control Hama",
		}, []game.Card{c})
	}
}
