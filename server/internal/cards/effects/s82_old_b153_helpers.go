package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b153_helpers.go — shared bodies for the Commander 2019 cards of
// S82 set OLD, batch 153 (#5517).

// oldB153FaithPaid is Faith of the Devoted's paid half: each opponent loses 2
// life and its controller gains 2. An on-disk key: never renamed, never
// reused.
var oldB153FaithPaid = PayUnlessBranches("pay-unless/faith-of-the-devoted-drain",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		for _, opp := range ctx.Opponents() {
			if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), opp, -2); err != nil {
				return err
			}
		}
		return GainLife{Player: ctx.Controller(), Amount: 2}.Apply(ctx)
	}, nil)

// OldB153FaithOffer asks for Faith of the Devoted's {1} as the trigger
// resolves.
func OldB153FaithOffer(g *game.Game, item *game.StackItem) error {
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{1}",
		Question: "Faith of the Devoted — pay {1}: each opponent loses 2 life and you gain 2 life?",
		Then:     oldB153FaithPaid,
	}.Apply(NewContext(g, item))
}

// OldB153FaceDownCreaturesYouControl is "face-down creatures you control", the
// AppliesTo of Secret Plans's +0/+1. The state is read, never the identity
// under it: a face-down permanent is a 2/2 creature whatever it hides (CR 708.2a).
func OldB153FaceDownCreaturesYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.FaceDownIsPermanent() && target.IsCreature() && target.Controller == source.Controller
}

// OldB153PermanentYouControlTurnedFaceUp is Secret Plans's "whenever a
// permanent you control is turned face up" (the permanent itself may be the
// source, which is never the case for a card that has no text face down).
func OldB153PermanentYouControlTurnedFaceUp(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventTurnedFaceUp {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Controller == source.Controller
}

// OldB153FirstFaceDownCreatureSpell is Kadena's "the first face-down creature
// spell you cast each turn": the spell is being cast face down, and none was
// before it this turn (the tally is bumped after the cost is determined,
// CR 601.2f).
func OldB153FirstFaceDownCreatureSpell() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Game != nil && q.Card.FaceDown && q.Game.CastTallyFor(q.Controller).FaceDownCreatureSpells == 0
	}
}

// OldB153AssassinToken is Vraska the Unseen's Assassin for CreateToken.
func OldB153AssassinToken() game.Card { return TokenFromCatalog(PrintedOldB153Assassin) }

// PrintedOldB153Assassin is the Assassin as PRINTED (Commander 2019 token): a
// 1/1 black Assassin creature with "Whenever this token deals combat damage to
// a player, that player loses the game." Its slug is an on-disk identity:
// never renamed, never reused.
func PrintedOldB153Assassin() TokenTemplate {
	return TokenTemplate{
		Slug: "assassin-loses-the-game",
		Card: game.Card{
			Name:      "Assassin",
			TypeLine:  "Token Creature — Assassin",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"B"},
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Assassin — that player loses the game",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return LoseTheGame{Player: ctx.Trigger().Event.Target}.Apply(ctx)
				}),
		},
		Text: "Whenever this token deals combat damage to a player, that player loses the game.",
	}
}

// oldB153VraskaCombatDamageCondition is "a creature deals combat damage to
// Vraska": a positive combat damage event from a creature whose target is
// the planeswalker that scheduled the trigger.
var oldB153VraskaCombatDamageCondition = game.DelayedCondition("damage/a-creature-deals-combat-damage-to-this-planeswalker",
	func(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) bool {
		if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || ev.Target != dt.SourceCardID {
			return false
		}
		src, ok := g.LookupCardForEffect(ev.Source)
		return ok && src.IsCreature()
	})

// oldB153VraskaDestroyBody destroys the creature that dealt the combat damage.
var oldB153VraskaDestroyBody = game.SimpleDelayedBody("vraska-the-unseen/destroy-the-damaging-creature", func(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id := ctx.Trigger().Event.Source
	if _, ok := g.PermanentRefForEffect(id); !ok {
		return nil
	}
	return DestroyTarget{Target: id}.Apply(ctx)
})

// OldB153VraskaPlusOne is Vraska the Unseen's +1: "Until your next turn,
// whenever a creature deals combat damage to Vraska, destroy that creature."
// A CR 603.7b delayed trigger with a stated duration, repeating (Jace,
// Architect of Thought's +1 is the template).
func OldB153VraskaPlusOne(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	until := DurationUntilYourNextTurn(ctx, ctx.Controller())
	g.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
		Controller:   ctx.Controller(),
		SourceCardID: ctx.Source(),
		Label:        "Vraska the Unseen — destroy the creature that dealt combat damage to her",
		On:           []game.EventKind{game.EventDealDamage},
		Condition:    oldB153VraskaCombatDamageCondition,
		Body:         oldB153VraskaDestroyBody,
		Repeats:      true,
		Duration:     &until,
	})
	return nil
}

// oldB153GhiredPopulate is the delayed body: "populate" for the trigger's
// controller.
var oldB153GhiredPopulate = game.SimpleDelayedBody("ghireds-belligerence/populate", func(g *game.Game, item *game.StackItem) error {
	return Populate{}.Apply(NewContext(g, item))
})

// oldB153GhiredDealt is the continuation of one creature's share of the
// damage: if it was really dealt damage (not prevented to nothing), "whenever
// it dies this turn, populate" is created about the object it is now
// (CR 603.7a). An on-disk key: never renamed, never reused.
var oldB153GhiredDealt = CardTail("then/ghireds-belligerence-dealt-damage", func(ctx *Context, o game.TailOutcome) error {
	if o.Amount <= 0 || len(o.IDs) == 0 {
		return nil
	}
	return WhenThatCreatureDiesThisTurn(ctx, o.IDs[0], "Ghired's Belligerence — populate", oldB153GhiredPopulate)
})

// OldB153GhiredsBelligerence resolves the spell: X damage divided as announced
// among the target creatures, each one dealt damage getting a delayed
// "when it dies this turn, populate".
func OldB153GhiredsBelligerence(ctx *Context) error {
	legal := ctx.LegalTargets()
	shares := ctx.Item.Distribution
	return ctx.Game.DamageInstanceForEffect(func() error {
		for _, t := range legal {
			if t.Kind != game.TargetCard || shares[t.ID] <= 0 {
				continue
			}
			if err := ctx.Game.DealDamageToCreatureKeyedForEffect(ctx.Source(), t.ID, shares[t.ID], ctx.Tail(oldB153GhiredDealt)); err != nil {
				return err
			}
		}
		return nil
	})
}
