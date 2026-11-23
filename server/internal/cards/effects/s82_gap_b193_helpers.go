package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b193_helpers.go — shared pieces of S82 GAP batch 193 (#5948).

// GapB193NextXThingCondition is the event condition of "when you next cast a
// spell with {X} in its mana cost or activate an ability with {X} in its
// activation cost this turn" (Magus Lucea Kane): an EventCast by the trigger's
// controller whose spell has an {X} in its own mana cost, or an
// EventActivateAbility by that player of a stack ability whose cost has an {X}
// mana symbol (Event.XCost, set only for a stack ability, so a mana ability
// never answers it). A cast or an activation that is not that event does not
// use the trigger up (CR 603.7b).
var GapB193NextXThingCondition = game.DelayedCondition("cast/you-next-cast-or-activate-x",
	func(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) bool {
		if ev.Actor != dt.Controller {
			return false
		}
		switch ev.Kind {
		case game.EventCast:
			c, ok := g.LookupCardForEffect(ev.CardID)
			if !ok {
				return false
			}
			cost, err := game.ParseCost(c.ManaCost)
			return err == nil && cost.XSlots > 0
		case game.EventActivateAbility:
			return ev.XCost && ev.StackItemID != uuid.Nil
		}
		return false
	})

// GapB193CopyTheXThingBody copies what the delayed trigger fired on: the spell
// (from last-known information, CR 608.2h) or the stack ability, with the
// new-targets offer for either.
var GapB193CopyTheXThingBody = game.SimpleDelayedBody("copy/the-x-spell-or-ability-you-just-used",
	func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		if ctx.Trigger().Event.Kind == game.EventActivateAbility {
			return CopyTheTriggeringAbility(g, item)
		}
		return CopyTheSpellYouJustCast(g, item)
	})

// GapB193ScheduleCopyNextXThing is the rider of "When you next cast a spell
// with {X} in its mana cost or activate an ability with {X} in its activation
// cost this turn, copy that spell or ability."
func GapB193ScheduleCopyNextXThing(source string) func(g *game.Game, controller, src uuid.UUID) error {
	label := source + " — copy that spell or ability"
	return func(g *game.Game, controller, src uuid.UUID) error {
		g.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
			Controller:   controller,
			SourceCardID: src,
			Label:        label,
			On:           []game.EventKind{game.EventCast, game.EventActivateAbility},
			Condition:    GapB193NextXThingCondition,
			Body:         GapB193CopyTheXThingBody,
		})
		return nil
	}
}

// GapB193YouControlACreatureEntered is "a creature entered the battlefield
// under your control this turn" for a trigger or item controlled by
// `controller`: the turn's creature-entry tally (Wary Farmer's reading),
// itself included.
func GapB193YouControlACreatureEntered(g *game.Game, controller uuid.UUID) bool {
	return g.EnteredWithCardTypeThisTurn(controller, "creature") >= 1
}

// GapB193ThisOrAnotherTyranidDealtCombatDamageToAPlayer is "whenever this
// creature or another Tyranid you control deals combat damage to a player".
func GapB193ThisOrAnotherTyranidDealtCombatDamageToAPlayer(ev game.Event, source *game.Card, c game.Characteristic, g *game.Game) bool {
	return ThisDealtCombatDamageToAPlayer(ev, source, c, g) ||
		B43CreatureOfSubtypeYouControlDealtCombatDamageToPlayer(ev, source, g, "Tyranid")
}

// GapB193ACreatureWithACounterYouControlDied is "a creature you control with a
// counter on it dies": the creature's counters of any kind as it last existed
// (CR 603.10a).
func GapB193ACreatureWithACounterYouControlDied(ev game.Event, source *game.Card, c game.Characteristic, g *game.Game) bool {
	if !ACreatureYouControlDied(ev, source, c, g) {
		return false
	}
	for _, n := range g.LastKnownCountersForEffect(ev.CardID) {
		if n > 0 {
			return true
		}
	}
	return false
}

// GapB193DestroyedArtifactControllerDraws is the tail of "destroy up to one
// target artifact. If an artifact is destroyed this way, its controller draws
// a card": the controller is the first carried id, and the draw happens only
// for a permanent that really was destroyed.
var GapB193DestroyedArtifactControllerDraws = CardTail("then/acolyte-hybrid-controller-draws", func(ctx *Context, o game.TailOutcome) error {
	owner := TailCarry0(o)
	if owner == uuid.Nil || len(o.IDs) == 0 {
		return nil
	}
	return DrawCards{Player: owner, N: 1}.Apply(ctx)
})

// GapB193DestroyTargetArtifactControllerDraws is the body of Acolyte Hybrid's
// attack trigger.
func GapB193DestroyTargetArtifactControllerDraws(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	victim := FirstLegalBattlefieldTarget(ctx)
	if victim == uuid.Nil {
		return nil
	}
	c, ok := g.LookupCardForEffect(victim)
	if !ok {
		return nil
	}
	return g.DestroyPermanentsKeyedForEffect([]uuid.UUID{victim},
		WithCarry(ctx.ItemTail(GapB193DestroyedArtifactControllerDraws), c.Controller))
}

// GapB193MalanthropeCounters is the tail of "exile target player's graveyard.
// Put a +1/+1 counter on this creature for each creature card exiled this
// way": the creature cards among the cards that actually reached exile, read
// off the exiled cards (a graveyard card keeps what it is on its way out).
// The counters go on the Malanthrope only while it is still that permanent.
var GapB193MalanthropeCounters = GraveyardExileTail("graveyard-exile/malanthrope-counters",
	func(ctx *Context, _ []uuid.UUID, wasCreature map[uuid.UUID]bool) error {
		n := 0
		for _, is := range wasCreature {
			if is {
				n++
			}
		}
		if n == 0 || !SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
			return nil
		}
		return AddCounter{Target: ctx.Item.SourceCardID, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
	})

// GapB193ExileTargetGraveyardCountCreatures is Malanthrope's enters body.
func GapB193ExileTargetGraveyardCountCreatures(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
		return nil
	}
	ctx := NewContext(g, item)
	p := g.PlayerByIDForEffect(item.Targets[0].ID)
	if p == nil || p.Graveyard == nil {
		return nil
	}
	ids := make([]uuid.UUID, 0, len(p.Graveyard.Cards))
	for _, c := range p.Graveyard.Cards {
		ids = append(ids, c.InstanceID)
	}
	return ExileCardsThen(ctx, ids, GapB193MalanthropeCounters)
}

// GapB193TrygonPrimeAttack is the body of "Whenever this creature attacks,
// put a +1/+1 counter on it and a +1/+1 counter on up to one other target
// attacking creature. That creature can't be blocked this turn."
func GapB193TrygonPrimeAttack(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if SourceIsStillThisPermanent(g, item) {
		if err := (AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	target := FirstLegalBattlefieldTarget(ctx)
	if target == uuid.Nil {
		return nil
	}
	if err := (AddCounter{Target: target, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	return RestrictUntilEOT{
		Target:       target,
		Restrictions: game.CantBeBlocked,
		Label:        "Trygon Prime — can't be blocked",
	}.Apply(ctx)
}

// gapB193AspiringChampionShuffleThenDemonDamage is the put's tail: shuffle the
// rest into the library, then, if the creature that arrived is a Demon, it
// deals damage equal to its power to each opponent.
var gapB193AspiringChampionShuffleThenDemonDamage = PutFromLibraryTail("put-from-library/aspiring-champion-shuffle-then-demon-damage",
	func(ctx *Context, res PutFromLibraryResult) error {
		if err := ctx.Game.ShuffleLibraryForEffect(res.Player); err != nil {
			return err
		}
		if len(res.Entered) == 0 {
			return nil
		}
		id := res.Entered[0]
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || !c.HasSubtype(ctx.TextSubtype("Demon")) || !OnBattlefield(ctx.Game, id) {
			return nil
		}
		power := c.CurrentPower()
		if power <= 0 {
			return nil
		}
		return ctx.Game.DamageInstanceForEffect(func() error {
			for _, opp := range ctx.Opponents() {
				if err := ctx.Game.DealDamageToPlayerForEffect(id, opp, power); err != nil {
					return err
				}
			}
			return nil
		})
	})

// GapB193AspiringChampionAscend is the "If you do" half of Aspiring Champion's
// trigger, after it is sacrificed: reveal cards from the top of your library
// until you reveal a creature card, put that card onto the battlefield, then
// shuffle the rest into your library; if it is a Demon, it deals damage equal
// to its power to each opponent. A library with no creature card is revealed
// whole and shuffled.
func GapB193AspiringChampionAscend(ctx *Context) error {
	player := ctx.Controller()
	run, hit := RevealUntil(ctx, player, game.Card.IsCreature, "Aspiring Champion — reveal until a creature card")
	return PutFromLibraryOntoBattlefield{
		Player:   player,
		Cards:    run,
		Match:    func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return hit != uuid.Nil && c.InstanceID == hit },
		All:      true,
		ThenTail: gapB193AspiringChampionShuffleThenDemonDamage,
	}.Apply(ctx)
}

// GapB193TargetCreatureAttacksTargetOpponent is Ravener's "target creature
// attacks target opponent this turn if able" (Dulcet Sirens' body).
func GapB193TargetCreatureAttacksTargetOpponent(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var creature, opp uuid.UUID
	for _, t := range ctx.LegalTargets() {
		switch t.Kind {
		case game.TargetCard:
			creature = t.ID
		case game.TargetPlayer:
			opp = t.ID
		}
	}
	if creature == uuid.Nil || opp == uuid.Nil {
		return nil
	}
	return UntilEndOfTurn(ctx, creature, nil, "Ravener — attacks that opponent this turn if able",
		game.AttackPlayerIfAbleMod(opp))
}
