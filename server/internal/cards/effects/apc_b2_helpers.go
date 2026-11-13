package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// apc_b2_helpers.go — shared bodies for Apocalypse, batch 2 (#5614).
// Append-only: add a function, never change what one means.

// thisDealtDamageToACreature is "whenever this creature deals damage to
// a creature": thisDealtCombatDamageToACreature without the combat
// test, so a fight or a ping counts as well (Spiritmonger). The damage
// event is emitted before the state-based actions that might kill the
// creature, so it is still on the battlefield to be asked about; a
// planeswalker or battle that took the damage is not a creature.
func thisDealtDamageToACreature(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || ev.Amount <= 0 || ev.Source != source.InstanceID {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Target)
	if !ok || !onBattlefield(g, ev.Target) {
		return false
	}
	return c.IsCreature()
}

// thisDealtDamage is "whenever this creature deals damage": each damage
// event this creature is the source of, combat or not, to any player or
// permanent (El-Hajjâj's reading).
func thisDealtDamage(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventDealDamage && ev.Amount > 0 && ev.Source == source.InstanceID
}

// aCreatureBlocked is "whenever a creature blocks": one trigger per
// blocking creature, whoever controls it. A creature that blocks several
// attackers blocks once (CR 509.3a), so only its first pair counts
// (EventBlock's Amount, as in selfBlocksOnce).
func aCreatureBlocked(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventBlock && ev.CardID != uuid.Nil && ev.Amount <= 1
}

// preventNextDamageToTheTarget is "Prevent the next N damage that would
// be dealt to any target this turn" as an activated ability's body: the
// charged shield (PreventNextDamage, CR 615.7) on the one target, at any
// amount. A target gone by resolution gets no shield (CR 608.2b).
func preventNextDamageToTheTarget(name string, n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		ts := ctx.LegalTargets()
		if len(ts) == 0 {
			return nil
		}
		label := name + ": prevent the next " + strconv.Itoa(n) + " damage"
		return PreventNextDamage{Target: ts[0].ID, Amount: n, Label: label}.Apply(ctx)
	}
}

// eachPlayerLifeChanges is "each player loses N life" (a negative
// delta) or "each player gains N life" (a positive one): every live
// player, the controller first, then the opponents. Life loss and gain,
// not damage.
func eachPlayerLifeChanges(ctx *Context, delta int) error {
	for _, id := range tablePlayers(ctx) {
		if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), id, delta); err != nil {
			return err
		}
	}
	return nil
}

// firstLegalTargetCardManaValue is the first chosen target card still
// legal as the spell resolves (CR 608.2b), and its mana value read now,
// before anything moves it (CR 608.2h). ok is false when there is none.
func firstLegalTargetCardManaValue(ctx *Context) (id uuid.UUID, mv int, ok bool) {
	ts := ctx.LegalTargets()
	if len(ts) == 0 || ts[0].Kind != game.TargetCard {
		return uuid.Nil, 0, false
	}
	if c, found := ctx.Game.LookupCardForEffect(ts[0].ID); found {
		mv = c.ManaValue()
	}
	return ts[0].ID, mv, true
}

// saprolingsPerManaValue is "create X 1/1 green Saproling creature
// tokens, where X is that creature's mana value", told X on the tail's N
// (read off the creature before it moved). An on-disk key.
var saprolingsPerManaValue = CardTail("apocalypse/saprolings-per-mana-value", func(ctx *Context, o game.TailOutcome) error {
	if o.Tail.N <= 0 {
		return nil
	}
	return CreateToken{Template: TokenCard("1/1 green Saproling"), N: o.Tail.N}.Apply(ctx)
})
