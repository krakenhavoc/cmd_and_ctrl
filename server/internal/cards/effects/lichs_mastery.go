package effects

import (
	"fmt"
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lich's Mastery — Legendary Enchantment {3}{B}{B}{B}:
//
//	"Hexproof
//	 You can't lose the game.
//	 Whenever you gain life, draw that many cards.
//	 Whenever you lose life, for each 1 life you lost, exile a permanent
//	 you control or a card from your hand or graveyard.
//	 When Lich's Mastery leaves the battlefield, you lose the game."
//
// #2065, unblocked by #2105. Five lines, five existing shapes:
//
//   - Hexproof is a printed keyword, Nine Lives' shape.
//   - "You can't lose the game" is a GateYou CantLose gate with no
//     cause narrowing (ADR 0057): every cause but a concession. The
//     rulings' "your opponents can still win the game if an effect says
//     so" is the gate saying nothing about winning.
//   - The life-gain draw reads the amount off the EventChangeLife the
//     trigger saw, Enduring Tenacity's shape. The rulings: the life still
//     changes; Lich's Mastery replaces neither the gain nor the loss.
//   - The life-loss trigger is Vilis's: s22PlayerLostLife over both
//     shapes a life loss takes, with the amount captured as it triggers.
//     Since #2105 it reads the life the damage COST, so infect damage
//     (and Phyrexian Unlife's as-though-infect damage) exiles nothing:
//     it gives poison and loses no life (CR 702.90b). As it resolves,
//     its controller picks that many objects from one pool of the
//     permanents they control and the cards in their hand and graveyard,
//     in any mix (the ruling), and they are exiled at once. With fewer
//     than that many, everything is exiled without a prompt, Lich's
//     Mastery itself included (the rulings).
//   - The leaves-the-battlefield loss is Nine Lives' trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "785c306c-471c-4699-8ab2-43c253d569cf",
		Name:            "Lich's Mastery",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
		GameEndGates:    []game.GameEndGate{{Scope: game.GateYou, CantLose: true}},
		Triggered: []game.TriggeredAbility{
			WheneverYouGainLife("Lich's Mastery — draw that many cards", lichsMasteryDraw),
			{
				Watches: []game.EventKind{game.EventChangeLife, game.EventDealDamage},
				Key:     lichsMasteryExileLabel,
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					_, ok := s22PlayerLostLife(ev, source.Controller, g)
					return ok
				},
				// A fill-in Build (ADR 0041 P9): the life lost is a fact
				// of the moment the ability triggered.
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, lichsMasteryExileLabel)
					item.Params.Amount, _ = s22PlayerLostLife(ev, source.Controller, g)
					return item
				},
				Effect: lichsMasteryExile,
			},
			WhenThisLeaves("Lich's Mastery — you lose the game", Do(LoseTheGame{})),
		},
	})
}

const lichsMasteryExileLabel = "Lich's Mastery — for each 1 life you lost, exile a permanent you control or a card from your hand or graveyard"

// lichsMasteryDraw is "draw that many cards": the life gained by the
// event the trigger saw.
func lichsMasteryDraw(g *game.Game, item *game.StackItem) error {
	return DrawCards{Player: item.Controller, N: item.Trigger.Event.Amount}.Apply(NewContext(g, item))
}

// lichsMasteryExile exiles item.Params.Amount objects from the pool of
// the controller's permanents, hand and graveyard, chosen by the
// controller, or the whole pool when it is no larger than that.
func lichsMasteryExile(g *game.Game, item *game.StackItem) error {
	n := item.Params.Amount
	you := item.Controller
	pool := lichsMasteryPool(g, you)
	if n <= 0 || len(pool) == 0 {
		return nil
	}
	if n >= len(pool) {
		g.ExileCardsForEffect(pool)
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  you,
		Source:   item.SourceCardID,
		Question: fmt.Sprintf("Lich's Mastery — exile %d: permanents you control, or cards from your hand or graveyard", n),
		Cards:    pool,
		Min:      n,
		Max:      n,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			return lichsMasteryExilePicked(g, you, pool, picked, n)
		},
	})
	return nil
}

// lichsMasteryPool is every permanent `you` control, then every card in
// their hand, then every card in their graveyard.
func lichsMasteryPool(g *game.Game, you uuid.UUID) []uuid.UUID {
	pool := PermanentsControlledBy(g, you)
	p := g.PlayerByIDForEffect(you)
	if p == nil {
		return pool
	}
	if p.Hand != nil {
		for _, c := range p.Hand.Cards {
			pool = append(pool, c.InstanceID)
		}
	}
	if p.Graveyard != nil {
		for _, c := range p.Graveyard.Cards {
			pool = append(pool, c.InstanceID)
		}
	}
	return pool
}

// lichsMasteryStillEligible reports whether `id` is still a permanent
// `you` control or a card in their hand or graveyard.
func lichsMasteryStillEligible(g *game.Game, you, id uuid.UUID) bool {
	z := g.FindCardZoneForEffect(id)
	if z == nil {
		return false
	}
	switch z.Kind {
	case game.ZoneBattlefield:
		c, ok := g.LookupCardForEffect(id)
		return ok && c.Controller == you
	case game.ZoneHand, game.ZoneGraveyard:
		return z.Owner == you
	}
	return false
}

// lichsMasteryExilePicked exiles the picks that are still eligible, and
// makes any shortfall up from the other candidates still eligible, in
// pool order, so an object that moved before the answer never lowers
// the count (Immortal Coil's rule).
func lichsMasteryExilePicked(g *game.Game, you uuid.UUID, candidates, picked []uuid.UUID, n int) error {
	var exile []uuid.UUID
	for _, list := range [][]uuid.UUID{picked, candidates} {
		for _, id := range list {
			if len(exile) < n && !slices.Contains(exile, id) && lichsMasteryStillEligible(g, you, id) {
				exile = append(exile, id)
			}
		}
	}
	g.ExileCardsForEffect(exile)
	return nil
}
