package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Augusta, Order Returned — Legendary Creature — Spirit Advisor {2}{W}, 1/3:
//
//	"Flying, vigilance
//	 Whenever Augusta attacks, each player exiles a card from their
//	 graveyard. When one or more nonland cards are exiled this way, put
//	 that many +1/+1 counters on target attacking creature."
//
// Every player chooses their own card, so it is one choose_cards run
// (Ill-Gotten Gains' shape): each player with a nonempty graveyard is
// asked about THEIR graveyard in APNAP order (CR 101.4), and nothing moves
// until the last has answered; then every chosen card is exiled. A player
// with an empty graveyard exiles nothing.
//
// "When one or more nonland cards are exiled this way" is a reflexive
// trigger (CR 603.12), created only when at least one card that reached
// exile is a nonland card; "that many" rides it as its amount, and its
// target, an attacking creature, is chosen as it goes on the stack and
// re-checked at resolution (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a80a1f84-ea49-4c8d-912d-c7036b814a6c",
		Name:            "Augusta, Order Returned",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Augusta, Order Returned — each player exiles a card from their graveyard", augustaEachPlayerExiles),
		},
	})
}

// augustaEachPlayerExiles asks each player for a card from their own
// graveyard.
func augustaEachPlayerExiles(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	players := apnapOrder(g, tablePlayers(ctx))
	var legs []game.ChooseCardsPrompt
	for _, id := range players {
		cards := graveyardCardsOfPlayer(g, id)
		if len(cards) == 0 {
			continue
		}
		legs = append(legs, game.ChooseCardsPrompt{
			Chooser:    id,
			FromPlayer: id,
			Source:     ctx.Source(),
			Question:   "Augusta, Order Returned — exile a card from your graveyard",
			Cards:      cards,
			Min:        1,
			Max:        1,
			Zone:       game.ZoneGraveyard,
		})
	}
	_, err := g.ChooseCardsRunThenKeyedForEffect(legs, augustaExiled, ctx.Source(), ctx.RunData(players))
	return err
}

// augustaExiled exiles every chosen card at once when all have answered;
// the count rides the exile's tail. An on-disk key.
var augustaExiled = RunThen("prompt-run/augusta-order-returned-exile", func(ctx *Context, r game.RunDone) error {
	picks := r.Picks()
	var chosen []uuid.UUID
	for _, player := range r.Carry {
		chosen = append(chosen, picks.By(player)...)
	}
	if len(chosen) == 0 {
		return nil
	}
	return ctx.Game.ExileCardsKeyedForEffect(chosen, ctx.ItemTail(augustaCountExiled))
})

// augustaCountExiled makes the reflexive trigger when a nonland card
// reached exile this way: a card a replacement sent elsewhere is not
// counted (CR 400.7). An on-disk key.
var augustaCountExiled = CardTail("augusta-order-returned/after-exile", func(ctx *Context, o game.TailOutcome) error {
	nonland := 0
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && !c.IsLand() {
			nonland++
		}
	}
	if nonland == 0 {
		return nil
	}
	return ReflexiveTrigger{
		Label:  "Augusta, Order Returned — put that many +1/+1 counters on target attacking creature",
		Body:   augustaCountersBody,
		Params: game.EffectParams{Amount: nonland},
	}.Apply(ctx)
})

// augustaCountersBody is "put that many +1/+1 counters on target attacking
// creature".
var augustaCountersBody = game.ReflexiveBody("augusta-order-returned/counters", simpleBody(augustaCounters),
	constTargets(func() *game.TargetSpec {
		return TargetCreature("target attacking creature", AttackingCreature())
	}))

func augustaCounters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: item.Params.Amount}.Apply(ctx)
	}
	return nil
}
