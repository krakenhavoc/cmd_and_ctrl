package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b140_helpers.go — shared bodies for the Throne of Eldraine adamant
// cards of S82 set OLD batch 140 (#5516). Adamant is an ability word
// (CR 207.2c): the card counts what was actually spent to cast it.

// OldB140CountersIfAdamantSpent is "Adamant — If at least three <color> mana was
// spent to cast this spell, this creature enters with a <kind> counter on it"
// (CR 614.1c): one counter when the spell's payment holds three or more of
// `color`, none when it does not or the permanent did not come from a spell.
// A waived payment counts only what it declared (the weaker direction).
func OldB140CountersIfAdamantSpent(kind, color string) game.EntryCountersFromCast {
	return game.EntryCountersFromCast{
		Kind: kind,
		Count: func(cast game.CastCounts) int {
			if cast.Spent.Count(color) >= 3 {
				return 1
			}
			return 0
		},
	}
}

// OldB140ThreeOfOneColor reports "at least three mana of the same color was
// spent": three or more of any one of white, blue, black, red or green.
// Colourless is not a colour (CR 105.1).
func OldB140ThreeOfOneColor(m game.ManaSpent) bool {
	for _, color := range []string{"W", "U", "B", "R", "G"} {
		if m.Count(color) >= 3 {
			return true
		}
	}
	return false
}

// OldB140CountersIfThreeOfOneColorSpent is OldB140CountersIfAdamantSpent for any one
// colour (Henge Walker).
func OldB140CountersIfThreeOfOneColorSpent(kind string) game.EntryCountersFromCast {
	return game.EntryCountersFromCast{
		Kind: kind,
		Count: func(cast game.CastCounts) int {
			if OldB140ThreeOfOneColor(cast.Spent) {
				return 1
			}
			return 0
		},
	}
}

// OldB140ClockworkServantSpentTrigger is "Adamant — When this creature enters, if at
// least three mana of the same color was spent to cast it, draw a card": an
// intervening "if" (CR 603.4) over the payment of the spell that became this
// permanent (CR 400.7d), checked as it enters and again as it resolves.
func OldB140ClockworkServantSpentTrigger() game.TriggeredAbility {
	const label = "Clockwork Servant — draw a card"
	paid := func(g *game.Game, id uuid.UUID) bool {
		return OldB140ThreeOfOneColor(g.CastProvenanceForEffect(id).Spent())
	}
	t := On(game.EventETB, Self, label, Do(DrawCards{N: 1}))
	when := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return source != nil && paid(g, source.InstanceID) && when(ev, source, lki, g)
	}
	return InterveningIf(CondBoard{
		Name: "at least three mana of the same color was spent to cast it",
		Test: func(ctx *Context) bool { return paid(ctx.Game, ctx.Item.SourceCardID) },
	}, t)
}

// OldB140GainLifePerCreature is "you gain 1 life for each creature you
// control", counted as it resolves.
type OldB140GainLifePerCreature struct{}

func (OldB140GainLifePerCreature) Apply(ctx *Context) error {
	n := len(MatchingBattlefield(ctx, And(Creature(), YouControl())))
	return GainLife{Amount: n}.Apply(ctx)
}

// OldB140IndestructibleToYourTarget is "the creature you control gains
// indestructible until end of turn": the first target clause, if it is still
// legal.
type OldB140IndestructibleToYourTarget struct{}

func (OldB140IndestructibleToYourTarget) Apply(ctx *Context) error {
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
		return nil
	}
	return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"indestructible"}, Label: "Outmuscle — indestructible"}.Apply(ctx)
}

// OldB140SearingBarrageAdamantSpent is "deals 5 damage to target creature. Adamant — if at
// least three red mana was spent, 3 damage to that creature's controller".
func OldB140SearingBarrageAdamantSpent(item *game.StackItem, ctx *Context) error {
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	c, found := ctx.Game.LookupCardForEffect(id)
	if !found {
		return nil
	}
	controller := c.Controller
	if err := (DealDamage{Source: item.SourceCardID, Target: id, Amount: 5}).Apply(ctx); err != nil {
		return err
	}
	if !AdamantSpent(ctx, "R", 3) {
		return nil
	}
	return DealDamage{Source: item.SourceCardID, Target: controller, Amount: 3}.Apply(ctx)
}

// OldB140CauldronsGiftChoose is "You may choose a creature card in your
// graveyard. If you do, return it to the battlefield with an additional
// +1/+1 counter on it": one optional pick over the creature cards in the
// graveyard as the spell resolves (not a target).
func OldB140CauldronsGiftChoose(ctx *Context) error {
	player := ctx.Controller()
	candidates := GraveyardCardIDs(ctx, player, func(c game.Card) bool { return c.IsCreature() && !c.IsToken() })
	if len(candidates) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:     player,
		FromPlayer:  player,
		Source:      ctx.Source(),
		Question:    "Cauldron's Gift — you may choose a creature card in your graveyard to return to the battlefield with an additional +1/+1 counter",
		Cards:       candidates,
		Min:         0,
		Max:         1,
		Destination: game.ChooseOntoBattlefield,
		ThenKey:     oldB140CauldronsGiftThen,
		LiveZone:    game.ZoneGraveyard,
		Asked:       game.NewAskedItem(ctx.Item),
	})
	return nil
}

// oldB140CauldronsGiftThen is the answer: the chosen creature card returns
// with the counter. An on-disk key: never renamed, never reused.
var oldB140CauldronsGiftThen = game.RegisterChooseCardsThen("choose-cards/cauldrons-gift-return",
	func(g *game.Game, r game.CardsChosen) error {
		if !r.Answered || len(r.Picked) == 0 {
			return nil
		}
		ctx, _ := AskedContext(g, r.Asked, r.Chooser, r.Source, nil)
		return ctx.Game.ReturnFromGraveyardWithCountersKeyedForEffect(r.Picked[0], r.Chooser, false,
			map[string]int{game.CounterPlusOne: 1}, game.KeyedTail{})
	})
