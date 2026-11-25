package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tahngarth, First Mate — Legendary Creature — Minotaur Warrior {2}{R}{G},
// 5/5 (Commander 2019):
//
//	"Tahngarth can't be blocked by more than one creature.
//	 Whenever an opponent attacks with one or more creatures, if Tahngarth is
//	 tapped, you may have that opponent gain control of Tahngarth until end of
//	 combat. If you do, choose a player or planeswalker that opponent is
//	 attacking. Tahngarth is attacking that player or planeswalker."
//
// The block rule is Bristling Boar's. The trigger is one per attacking
// opponent however many creatures they send (Emberwilde Captain's batch), with
// the CR 603.4 intervening "if": Tahngarth is tapped as the attack is
// declared and again as the ability resolves, before its "you may" is asked.
// If you do, that opponent gains control until end of combat (CR 611.2b, a
// duration that exists only inside a combat phase), then you choose, from the
// players and planeswalkers that opponent's creatures are attacking, the one
// Tahngarth attacks: CR 506.3c's "put into combat attacking" door
// (Game.PutOnBattlefieldAttackingForEffect), which declares nothing, so
// nothing triggers on its attack. Tahngarth is already tapped. A control
// change that did not take place (the duration never began) puts it nowhere.
//
// No simplification.
func init() {
	const label = "Tahngarth, First Mate — you may have that opponent gain control of Tahngarth until end of combat; it attacks a player or planeswalker they are attacking"
	trig := game.TriggeredAbility{
		Watches: []game.EventKind{game.EventAttack},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			return ev.Kind == game.EventAttack && ev.Actor != uuid.Nil && ev.Actor != source.Controller &&
				source.Tapped && g.PlayerByIDForEffect(ev.Actor) != nil
		},
		Key: label,
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			return game.NewTriggeredItem(source, label)
		},
		Effect: tahngarthResolve,
	}
	trig = effects.OncePerBatchPerAttacker(trig)
	effects.Register(effects.Spec{
		OracleID:     "1a8a41a1-6fc5-4dc8-bd28-9bf404c777d7",
		Name:         "Tahngarth, First Mate",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.MaxBlockers(effects.OnSelf(), 1),
		},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{Name: "Tahngarth is tapped", Test: func(ctx *effects.Context) bool {
				c, ok := ctx.Game.LookupCardForEffect(ctx.Source())
				return ok && c.Tapped && ctx.Game.Battlefield.Contains(ctx.Source())
			}}, trig),
		},
	})
}

// tahngarthResolve asks "you may", then hands Tahngarth over and puts it
// into the opponent's combat.
func tahngarthResolve(g *game.Game, item *game.StackItem) error {
	return effects.MayChoice{
		Question: "Tahngarth, First Mate — have that opponent gain control of Tahngarth until end of combat?",
		YesLabel: "Give control",
		NoLabel:  "Decline",
		Then:     tahngarthAnswer,
	}.Apply(effects.NewContext(g, item))
}

// tahngarthAnswer is the "yes". Keyed.
var tahngarthAnswer = effects.MayYes("confirm/tahngarth-first-mate-give", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	if ctx.Item == nil || ctx.Item.Trigger == nil {
		return nil
	}
	opp := ctx.Item.Trigger.Event.Actor
	if opp == uuid.Nil || !ctx.Game.Battlefield.Contains(ctx.Source()) {
		return nil
	}
	d, ok := effects.DurationUntilEndOfCombat(ctx)
	if !ok {
		return nil
	}
	if err := (effects.GainControl{
		Target:     ctx.Source(),
		Controller: opp,
		Duration:   d,
		Label:      "Tahngarth, First Mate — that opponent gains control until end of combat",
	}).Apply(ctx); err != nil {
		return err
	}
	ctx.Game.RecomputeLayersIfStaleLocked()
	if c, ok := ctx.Game.LookupCardForEffect(ctx.Source()); !ok || c.Controller != opp {
		return nil
	}
	var options []uuid.UUID
	seen := map[uuid.UUID]bool{}
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller != opp || c.AttackingTarget == uuid.Nil || seen[c.AttackingTarget] {
			continue
		}
		seen[c.AttackingTarget] = true
		if k := ctx.Game.ClassifyAttackTargetForEffect(c.AttackingTarget); k == game.AttackTargetPlayer || k == game.AttackTargetPlaneswalker {
			options = append(options, c.AttackingTarget)
		}
	}
	if len(options) == 0 {
		return nil
	}
	if len(options) == 1 {
		ctx.Game.PutOnBattlefieldAttackingForEffect(ctx.Source(), options[0])
		return nil
	}
	choices := make([]game.ChoiceOption, 0, len(options))
	for _, id := range options {
		var opt game.ChoiceOption
		if p := ctx.Game.PlayerByIDForEffect(id); p != nil {
			opt = game.ChoiceOption{Label: "Attack " + p.Name, Player: id}
		} else if c, ok := ctx.Game.LookupCardForEffect(id); ok {
			opt = game.ChoiceOption{Label: "Attack " + c.Name, Cards: []uuid.UUID{id}}
		} else {
			continue
		}
		choices = append(choices, opt)
	}
	return effects.PickOption{
		Question: "Tahngarth, First Mate — choose a player or planeswalker that opponent is attacking",
		Options:  choices,
		ThenKey:  tahngarthDefenderChosen,
		Carry:    []uuid.UUID{ctx.Source()},
	}.Apply(ctx)
})

// tahngarthDefenderChosen puts Tahngarth into combat attacking the chosen
// player or planeswalker. An on-disk key: never renamed, never reused.
var tahngarthDefenderChosen = effects.OptionPickThen("option-pick/tahngarth-first-mate-defender", func(ctx *effects.Context, r game.OptionPicked) error {
	if r.Option == nil || len(r.Carry) == 0 {
		return nil
	}
	defender := r.Option.Player
	if defender == uuid.Nil && len(r.Option.Cards) > 0 {
		defender = r.Option.Cards[0]
	}
	if defender == uuid.Nil {
		return nil
	}
	ctx.Game.PutOnBattlefieldAttackingForEffect(r.Carry[0], defender)
	return nil
})
