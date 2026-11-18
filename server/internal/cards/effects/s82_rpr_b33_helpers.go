package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_rpr_b33_helpers.go — shared bodies for the S82 reprint pass, batch 33.

// demandingDragonName is the label stem of the Dragon's trigger.
const demandingDragonName = "Demanding Dragon"

const (
	demandingDragonTakeLabel = "Take 5 damage"
	demandingDragonSacLabel  = "Sacrifice a creature"
)

// demandingDragonAsk is "it deals 5 damage to target opponent unless that
// player sacrifices a creature of their choice": the opponent decides, and
// with no creature to sacrifice the damage is simply dealt.
func demandingDragonAsk(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ts := ctx.LegalTargets()
	if len(ts) == 0 || ts[0].Kind != game.TargetPlayer {
		return nil
	}
	victim := ts[0].ID
	p := g.PlayerByIDForEffect(victim)
	if p == nil || p.Eliminated {
		return nil
	}
	if len(creaturesControlledByPlayer(g, victim)) == 0 {
		return demandingDragonDamage(ctx, victim)
	}
	return PickOption{
		Player:   victim,
		Question: demandingDragonName + " — take 5 damage, or sacrifice a creature?",
		Options: []game.ChoiceOption{
			{Label: demandingDragonTakeLabel},
			{Label: demandingDragonSacLabel},
		},
		ThenKey: demandingDragonAnswer,
	}.Apply(ctx)
}

func demandingDragonDamage(ctx *Context, victim uuid.UUID) error {
	return DealDamage{Source: ctx.Source(), Target: victim, Amount: 5}.Apply(ctx)
}

// demandingDragonAnswer is the opponent's answer. An on-disk key: never
// renamed, never reused. A refusal to answer takes the damage.
var demandingDragonAnswer = OptionPickThen("option-pick/demanding-dragon-answer", func(ctx *Context, r game.OptionPicked) error {
	victim := r.Chooser
	if r.Option != nil && r.Option.Label == demandingDragonSacLabel {
		return SacrificeChoice{
			Player:     victim,
			Candidates: creaturesControlledByPlayer(ctx.Game, victim),
			Question:   demandingDragonName + " — sacrifice a creature",
		}.Apply(ctx)
	}
	return demandingDragonDamage(ctx, victim)
})

// eachAttackingOpponentDoesTheSame is the second half of a Curse that
// reads "Whenever enchanted player is attacked, you <do X>. Each opponent
// attacking that player does the same": a trigger for the ATTACKING
// opponent (never the Curse's controller), who becomes the item's
// controller so the effect is theirs. One declaration is one "is attacked"
// (CR 506.3), so it is batched per attacker: two creatures from the same
// player still earn one. Curse of Opulence and Curse of Verbosity.
func eachAttackingOpponentDoesTheSame(label string, effect Effect) game.TriggeredAbility {
	t := game.TriggeredAbility{
		Watches: []game.EventKind{game.EventAttack},
		Key:     label,
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			if source.AttachedTo.Kind != game.TargetPlayer || source.AttachedTo.ID != ev.Target {
				return false
			}
			return ev.Actor != uuid.Nil && ev.Actor != source.Controller
		},
		// A fill-in Build (ADR 0041 P9): the attacking opponent becomes
		// the item's controller, a fact of the moment the trigger fired.
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			item.Controller, item.Owner = ev.Actor, ev.Actor
			return item
		},
		Effect: effect,
	}
	t.OncePerBatch = true
	t.BatchKey = func(ev game.Event, _ *game.Card, _ *game.Game) string {
		return ev.Actor.String()
	}
	return t
}

// curseOfVerbosityTriggers is the Curse's two abilities, batched per
// declaration exactly as Curse of Opulence's are.
func curseOfVerbosityTriggers() []game.TriggeredAbility {
	mine := OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return source.AttachedTo.Kind == game.TargetPlayer && source.AttachedTo.ID == ev.Target
	}, "Curse of Verbosity — you draw a card", Do(DrawCards{N: 1})))
	return []game.TriggeredAbility{
		mine,
		eachAttackingOpponentDoesTheSame("Curse of Verbosity — the attacking opponent draws a card", Do(DrawCards{N: 1})),
	}
}
