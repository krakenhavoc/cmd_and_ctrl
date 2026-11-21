package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82BnsB12GrantForLife is "Pay 2 life: This creature gains <keyword>
// until end of turn" (Glorifier of Dusk). The life is a cost (CR 119.4).
func S82BnsB12GrantForLife(keyword string) ActivatedAbility {
	return ActivatedAbility{
		Label:   "Pay 2 life: This creature gains " + keyword + " until end of turn.",
		Purpose: game.Purpose{Answers: game.AnswerPump},
		Cost:    PayLife(2),
		Effect: func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			return GrantKeywordUntilEOT{Target: ctx.Source(), Keywords: []string{keyword}, Label: "Glorifier of Dusk — " + keyword}.Apply(ctx)
		},
	}
}

// s82BnsB12OpponentsDealtDamage counts the opponents of controller who were
// dealt damage this turn (damage that was prevented never was dealt).
func S82BnsB12OpponentsDealtDamage(g *game.Game, controller uuid.UUID) int {
	hit := map[uuid.UUID]bool{}
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventDealDamage || ev.Amount <= 0 || ev.Target == controller {
			continue
		}
		if g.PlayerByIDForEffect(ev.Target) == nil {
			continue
		}
		hit[ev.Target] = true
	}
	return len(hit)
}

// s82BnsB12HeraldCast is Herald of Ilharg's body: two counters, then, if the
// cast spell has mana value 5 or greater, damage equal to the counters on it
// to each opponent.
func S82BnsB12HeraldCast(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ctx := NewContext(g, item)
	if err := (CounterOnThis{Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
		return err
	}
	if SpellManaValueForEffect(g, item.Trigger.Event.CardID) < 5 {
		return nil
	}
	self, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || self.Counters[game.CounterPlusOne] <= 0 {
		return nil
	}
	return DamageEachOpponent{N: self.Counters[game.CounterPlusOne]}.Apply(ctx)
}
