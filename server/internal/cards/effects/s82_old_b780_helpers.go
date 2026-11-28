package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b780_helpers.go — S82 OLD batch 780 (Weatherlight, #5629).

// S82OldB780UpkeepSacrificeUnlessExileTopCreatureCard is Circling Vultures'
// "At the beginning of your upkeep, sacrifice this creature unless you exile the
// top creature card of your graveyard": the Elder Dragons' upkeep tax paid in
// cards (UpkeepPayUnless with the #6194 pay action). The card is not the
// payer's to pick, it is the topmost creature card of their graveyard
// (CR 404.2), and a graveyard with none can't pay (CR 118.3), so declining or
// being unable to pay sacrifices the creature.
func S82OldB780UpkeepSacrificeUnlessExileTopCreatureCard(name string) game.TriggeredAbility {
	pay := ExileTopGraveyardCardPayment("creature")
	return AtYourUpkeep(name+" — sacrifice this creature unless you exile the top creature card of your graveyard",
		func(g *game.Game, item *game.StackItem) error {
			return UpkeepPayUnless{
				Chooser:  item.Controller,
				Cost:     pay.Words,
				Action:   pay.Action,
				Question: name + " — exile the top creature card of your graveyard or sacrifice " + name + "?",
				Else:     UpkeepSacrificeSource,
			}.Apply(NewContext(g, item))
		})
}

// S82OldB780ThisHasNoCounters is the condition of "as long as this creature has
// no <kind> counters on it" (Roc Hatchling), read live as the layers are
// recomputed.
func S82OldB780ThisHasNoCounters(kind string) func(target *game.Card, g *game.Game, source *game.Card) bool {
	return func(target *game.Card, g *game.Game, source *game.Card) bool {
		return SelfOnly(target, g, source) && target.Counters[kind] <= 0
	}
}

// S82OldB780GoblinBombUpkeep is Goblin Bomb's upkeep trigger: a "you may flip a
// coin" asked as the trigger resolves (CR 603.5), the call and flip following
// the yes.
func S82OldB780GoblinBombUpkeep(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return MayChoice{
			Question: label + "?",
			YesLabel: "Flip",
			NoLabel:  "Decline",
			Then:     s82OldB780GoblinBombFlip,
		}.Apply(NewContext(g, item))
	}
}

// s82OldB780GoblinBombFlip is the "yes": the coin is called and flipped.
// An on-disk key.
var s82OldB780GoblinBombFlip = MayYes("confirm/s82-old-b780-goblin-bomb-flip", func(ctx *Context, _ game.ConfirmAnswer) error {
	ctx.Game.FlipCoinForEffect(game.CoinFlipSpec{
		Flipper:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Goblin Bomb — call the coin flip",
		ThenKey:  s82OldB780GoblinBombFlipped,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
})

// s82OldB780GoblinBombFlipped is the flip settled: a win puts a fuse counter on
// the Bomb, a loss removes one (none to remove, nothing happens). An on-disk key.
var s82OldB780GoblinBombFlipped = CoinFlipThen("coin-flip/s82-old-b780-goblin-bomb", func(ctx *Context, r game.CoinFlipped) error {
	if !OnBattlefield(ctx.Game, ctx.Source()) {
		return nil
	}
	if WonFirstFlip(r) {
		return AddCounter{Target: ctx.Source(), Kind: "fuse", N: 1}.Apply(ctx)
	}
	return RemoveACounterFromThis("fuse")(ctx.Game, ctx.Item)
})

// S82OldB780ThisEnteredThisTurn is "Activate only if this creature entered this
// turn" (Fungus Elemental): game.EnteredThisTurn, not summoning sickness (a
// creature that entered during an opponent's turn did not enter this turn).
func S82OldB780ThisEnteredThisTurn() ActivationCondition {
	return func(g *game.Game, _, source uuid.UUID) bool {
		c, ok := g.LookupCardForEffect(source)
		return ok && SourceEnteredThisTurn(g, c)
	}
}

// S82OldB780LiegeOfTheHollows is Liege of the Hollows' resolution: starting with
// the trigger's controller and in turn order, each player may pay any amount of
// mana (their own, CR 608.2d), and creates that many 1/1 green Squirrel creature
// tokens, each under their own control.
type S82OldB780LiegeOfTheHollows struct{ then game.PayUnlessElse }

// NewS82OldB780LiegeOfTheHollows builds the process; call it once, from a
// package-level var (the key is an on-disk identity: never renamed, never reused).
func NewS82OldB780LiegeOfTheHollows(key string) *S82OldB780LiegeOfTheHollows {
	l := &S82OldB780LiegeOfTheHollows{}
	l.then = PayUnlessBranches(key,
		func(ctx *Context, r game.PayUnlessAnswer) error {
			if r.Number > 0 {
				if err := (CreateToken{Controller: r.Chooser, Template: TokenCard("1/1 green Squirrel"), N: r.Number}).Apply(ctx); err != nil {
					return err
				}
			}
			return l.ask(ctx, r.Carry)
		},
		func(ctx *Context, r game.PayUnlessAnswer) error { return l.ask(ctx, r.Carry) })
	return l
}

// Start asks every living player in turn, from the controller.
func (l *S82OldB780LiegeOfTheHollows) Start(ctx *Context) error {
	return l.ask(ctx, SeatsFromController(ctx.Game, ctx.Controller()))
}

func (l *S82OldB780LiegeOfTheHollows) ask(ctx *Context, seats []uuid.UUID) error {
	for len(seats) > 0 {
		if p := ctx.Game.PlayerByIDForEffect(seats[0]); p != nil && !p.Eliminated {
			break
		}
		seats = seats[1:]
	}
	if len(seats) == 0 {
		return nil
	}
	return MayPayX{
		Chooser: seats[0],
		Symbol:  "{1}",
		Label:   "Liege of the Hollows",
		Buys:    "create X 1/1 green Squirrel creature tokens",
		Unit:    game.PayAmountOther,
		Then:    l.then,
		Carry:   append([]uuid.UUID(nil), seats[1:]...),
	}.Apply(ctx)
}
