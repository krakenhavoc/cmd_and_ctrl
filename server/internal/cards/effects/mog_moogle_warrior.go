package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mog, Moogle Warrior — Legendary Creature — Moogle Warrior {1}{R}{W},
// 1/2 (deck request #3655):
//
//	"Lifelink
//	 Dance — At the beginning of your end step, each player may discard a
//	 card. Each player who discarded a card this way draws a card. If a
//	 creature card was discarded this way, you create a 1/2 white Moogle
//	 creature token with lifelink. Then if a noncreature card was
//	 discarded this way, put a +1/+1 counter on each Moogle you control."
//
// Each player, starting with the controller, chooses whether to discard
// and which card, then draws for it; the token and the counters come
// once the last player has answered, so "a creature card was discarded"
// and "a noncreature card was discarded" read the whole table. The new
// Moogle is one of the Moogles that get the counter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f18cce8c-8fb6-49a7-8f79-2529c0ee3ecc",
		Name:            "Mog, Moogle Warrior",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Mog, Moogle Warrior — each player may discard a card and draw a card",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return mogAsk(ctx, tablePlayers(ctx))
				}),
		},
	})
}

// What the table has discarded this resolution (one number each: 1 when a
// creature card, 1 when a noncreature card went), and who is being asked and
// who is left, are step records (#4163): the answers arrive one at a time, and
// a Go variable carried through them would be lost to a restore.
const (
	mogDanceRecord  = "mog-moogle-warrior/dance"
	mogSeatRecord   = "mog-moogle-warrior/seat"
	mogRestRecord   = "mog-moogle-warrior/rest"
	mogFlagCreature = 0
	mogFlagOther    = 1
)

// mogDanceFlags is what the table discarded so far: a creature card, a
// noncreature card.
func mogDanceFlags(ctx *Context) (creature, noncreature bool) {
	ns := ctx.StepRecord(mogDanceRecord).Ns
	if len(ns) < 2 {
		return false, false
	}
	return ns[mogFlagCreature] == 1, ns[mogFlagOther] == 1
}

func mogAsk(ctx *Context, players []uuid.UUID) error {
	if len(players) == 0 {
		return mogFinish.Run(ctx)
	}
	player, rest := players[0], players[1:]
	return MayChoice{
		Player:   player,
		Question: "Mog, Moogle Warrior — discard a card (and draw a card)?",
		Then:     mogAnswer,
		Carry:    rest,
	}.Apply(ctx)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// mogDrawThenAsk is the player's draw for the card they discarded and then the
// question to the next player: the question waits for a prompt the draw asks
// (a dredge offer; CR 608.2c, #4155). It asks again, so it is registered in
// init.
var mogDrawThenAsk ThenSteps

// mogFinish makes the Moogle if a creature card went, then counters the
// Moogles if a noncreature card went: the counters wait for a prompt the
// token's creation asks (CR 608.2c, #4155).
var mogFinish = RegisterThenSteps("then-steps/mog-moogle-warrior-finish",
	Instruction(func(ctx *Context) error {
		if creature, _ := mogDanceFlags(ctx); creature {
			return CreateToken{Controller: ctx.Controller(), Template: TokenCard("1/2 white Moogle with lifelink"), N: 1}.Apply(ctx)
		}
		return nil
	}),
	Instruction(func(ctx *Context) error {
		if _, noncreature := mogDanceFlags(ctx); !noncreature {
			return nil
		}
		var moogles []uuid.UUID
		// "Moogle" is a word of the source's text (CR 612.2, #3984).
		moogle := ctx.TextSubtype("Moogle")
		for _, c := range ctx.Game.Battlefield.Cards {
			if c.Controller == ctx.Controller() && c.HasSubtype(moogle) {
				moogles = append(moogles, c.InstanceID)
			}
		}
		for _, id := range moogles {
			if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.asGroupMember()); err != nil {
				return err
			}
		}
		return nil
	}),
)

func init() {
	mogDrawThenAsk = RegisterThenSteps("then-steps/mog-moogle-warrior-draw",
		Instruction(func(ctx *Context) error {
			seat := ctx.StepRecord(mogSeatRecord)
			if len(seat.IDs) == 0 || seat.N() == 0 {
				return nil
			}
			return ctx.Game.DrawNForEffect(seat.IDs[0], 1)
		}),
		Instruction(func(ctx *Context) error {
			return mogAsk(ctx, ctx.StepRecord(mogRestRecord).IDs)
		}),
	)
}

// mogAnswer is one player's answer to "discard a card (and draw a card)?": a
// yes opens their discard and the steps behind it, a no asks the next player
// (the carried IDs). Registered in init because the no asks through mogAsk,
// which names it. Keyed (#4291).
var mogAnswer game.ConfirmThen

func init() {
	mogAnswer = MayYesNo("confirm/mog-moogle-warrior-discard", mogDiscard,
		func(c *Context, r game.ConfirmAnswer) error { return mogAsk(c, r.Carry) })
}

// mogDiscard is a player's "yes": they discard a card, and the steps behind
// it draw and ask the rest (the carried IDs).
func mogDiscard(c *Context, r game.ConfirmAnswer) error {
	player, rest := r.Chooser, r.Carry
	// The resolving item rides the prompt as data (the step records are
	// filed under its id), with the controller and then the players still
	// to ask as the carried IDs.
	carry := append([]uuid.UUID{c.Controller()}, rest...)
	return c.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
		Player:   player,
		Source:   c.Source(),
		N:        1,
		Question: "Mog, Moogle Warrior — discard a card",
		Asked:    game.NewAskedItem(c.Item),
	}, mogDiscarded, carry)
}

// mogDiscarded is a player's discard landed, keyed (#4253): the table's
// running tally is filed, the draw follows and the next player is asked.
var mogDiscarded = game.RegisterDiscardThen("choose-cards/mog-moogle-warrior-discarded", func(g *game.Game, r game.DiscardsDone) error {
	next, rest := askedContext(g, r.Item, r.Player, r.Source, r.Carry)
	discarded := r.Discarded.Cards()
	creature, other := mogDanceFlags(next)
	for _, id := range discarded {
		if card, ok := g.LookupCardForEffect(id); ok {
			if card.IsCreature() {
				creature = true
			} else {
				other = true
			}
		}
	}
	next.RecordStep(mogDanceRecord, nil, boolToInt(creature), boolToInt(other))
	next.RecordStep(mogSeatRecord, []uuid.UUID{r.Player}, len(discarded))
	next.RecordStep(mogRestRecord, rest)
	return mogDrawThenAsk.Run(next)
})
