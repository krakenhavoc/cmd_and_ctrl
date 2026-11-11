package effects

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s38_helpers.go — shared pieces of the Karlov Manor Commander
// face-up payoffs (#5368, batch 38).

// cmdr2S38ChooseNewTargetsForOthers is Boltbender's "you may choose new
// targets for any number of other spells and/or abilities" (CR 115.7c): one
// optional choose-new-targets offer per other stack item that has a target,
// the trigger's own item excluded. Each offer is its own prompt the
// controller may decline, so "any number" is the number they answer.
func cmdr2S38ChooseNewTargetsForOthers(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for id := range g.StackMeta {
		if id != item.ID && g.RetargetableForEffect(id) {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i].String() < ids[j].String() })
	for _, id := range ids {
		if err := (ChangeTargets{
			StackID:  id,
			Chooser:  ctx.Controller(),
			Policy:   game.RetargetChooseNew,
			Optional: true,
			Reason:   "Boltbender — you may choose new targets",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// cmdr2S38PutCountersEqualToPower is Experiment Twelve's "put +1/+1
// counters on that creature equal to its power": the creature that was
// turned face up, its power read as the ability resolves.
func cmdr2S38PutCountersEqualToPower(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	id := item.Trigger.Event.CardID
	if !onBattlefield(g, id) {
		return nil
	}
	n := b31CurrentPowerOf(g, id)
	if n <= 0 {
		return nil
	}
	return AddCounter{Target: id, Kind: game.CounterPlusOne, N: n}.Apply(NewContext(g, item))
}

// cmdr2S38CreaturesYouControl counts the creatures `player` controls.
func cmdr2S38CreaturesYouControl(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() {
			n++
		}
	}
	return n
}

// cmdr2S38OozeFromFaceUp is Printlifter Ooze's trigger body: a 0/0 green
// Ooze with trample that enters with a +1/+1 counter for each other
// creature the controller controls. The token is not on the battlefield
// yet, so every creature counted is "other" than it.
func cmdr2S38OozeFromFaceUp(g *game.Game, item *game.StackItem) error {
	x := cmdr2S38CreaturesYouControl(g, item.Controller)
	ctx := NewContext(g, item)
	_, err := g.CreateTokensForEffect(item.Controller,
		ctx.TextToken(TokenCard("0/0 green Ooze with trample")), 1,
		game.TokenEntryOptions{Counters: map[string]int{game.CounterPlusOne: x}})
	return err
}

// cmdr2S38TurnedFaceUpThisTurn is "this creature was turned face up this
// turn": an EventTurnedFaceUp this turn names it (Kaust).
func cmdr2S38TurnedFaceUpThisTurn(g *game.Game, id uuid.UUID) bool {
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind == game.EventTurnedFaceUp && ev.CardID == id {
			return true
		}
	}
	return false
}

// cmdr2S38FaceUpCreatureDealtCombatDamageToAPlayer is Kaust's "whenever a
// creature you control that was turned face up this turn deals combat
// damage to a player".
func cmdr2S38FaceUpCreatureDealtCombatDamageToAPlayer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return combatDamageToPlayerBy(ev, source.Controller, g) && cmdr2S38TurnedFaceUpThisTurn(g, ev.Source)
}

// cmdr2S38SeatClauses are Unexplained Absence's "for each player, up to one
// target nonland permanent that player controls": one clause per seat,
// counted from the caster in turn order, so a table of any size up to four
// is covered and a seat that is not there has nothing to pick.
func cmdr2S38SeatClauses() *game.TargetSpec {
	var clauses []*game.TargetSpec
	for k := 0; k < game.MaxPlayers; k++ {
		k := k
		clauses = append(clauses, TargetPermanent("up to one target nonland permanent (player "+string(rune('1'+k))+" from you)",
			Nonland(), func(g *game.Game, caster uuid.UUID, c game.Card) bool {
				return cmdr2S38SeatFrom(g, caster, k) == c.Controller
			}).WithCount(0, 1))
	}
	return Clauses(clauses[0], clauses[1:]...)
}

// cmdr2S38SeatFrom is the player k seats after `from` in turn order (0 is
// `from` itself), uuid.Nil when the table has no such seat.
func cmdr2S38SeatFrom(g *game.Game, from uuid.UUID, k int) uuid.UUID {
	n := len(g.Seats)
	for i, p := range g.Seats {
		if p != nil && p.ID == from {
			if k >= n {
				return uuid.Nil
			}
			return g.Seats[(i+k)%n].ID
		}
	}
	return uuid.Nil
}
