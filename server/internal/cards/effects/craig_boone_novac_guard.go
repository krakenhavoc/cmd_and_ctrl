package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Craig Boone, Novac Guard — Legendary Creature — Human Soldier {1}{R}{W},
// 3/3 (Fallout, #5948):
//
//	"Reach, lifelink
//	 One for My Baby — Whenever you attack with two or more creatures, put two
//	 quest counters on Craig Boone. When you do, Craig Boone deals damage equal
//	 to the number of quest counters on it to up to one target creature unless
//	 that creature's controller has Craig Boone deal that much damage to
//	 them."
//
// "Whenever you attack" is once per declaration, and only for two or more
// attackers (OncePerBatch with the settled batch, ThatManyEventsAtLeast(2)).
// The counters go on as it resolves, and "when you do" is a CR 603.12
// reflexive trigger whose "up to one target creature" is chosen as it goes on
// the stack. As it resolves the quest counters on Boone are counted (his
// last-known count if he has left) and the target creature's controller is
// asked whether to have Boone deal that much damage to them instead: yes
// deals it to that player and spares the creature, no (or nobody left to ask)
// deals it to the creature. Boone deals it as he last existed, lifelink
// included; a creature that has left or changed legality by then takes none.
//
// No simplification.
func init() {
	attack := withAtBatchEnd(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return attackDeclaredByYou(ev, source.Controller)
	}, "Craig Boone — put two quest counters on it; when you do, it deals damage to up to one target creature unless its controller has it deal that damage to them",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			if err := (AddCounter{Target: item.SourceCardID, Kind: "quest", N: 2}).Apply(ctx); err != nil {
				return err
			}
			return ReflexiveTrigger{
				Label: "Craig Boone — damage equal to its quest counters to up to one target creature",
				Body:  craigBooneDamageBody,
			}.Apply(ctx)
		}), ThatManyEventsAtLeast(2))
	Register(Spec{
		OracleID:        "74f4743b-06a9-48e3-9fdf-27a56660ac8e",
		Name:            "Craig Boone, Novac Guard",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "lifelink"},
		Triggered:       []game.TriggeredAbility{OncePerBatch(attack)},
	})
}

var craigBooneDamageBody = game.ReflexiveBody("craig-boone/damage", func(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	target, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	n := 0
	if c, found := g.LookupCardForEffect(item.SourceCardID); found && onBattlefield(g, c.InstanceID) {
		n = c.Counters["quest"]
	} else if info, found := g.LastKnownPermanentForEffect(item.SourceCardID); found {
		n = info.Counters["quest"]
	}
	if n <= 0 {
		return nil
	}
	return PickOption{
		Player:   target.Controller,
		Question: "Craig Boone — have it deal " + strconv.Itoa(n) + " damage to you instead of to " + target.Name + "?",
		Options: []game.ChoiceOption{
			{Label: "No — it deals the damage to " + target.Name},
			{Label: "Yes — it deals " + strconv.Itoa(n) + " damage to you"},
		},
		ThenKey:  craigBooneChoice,
		WithItem: true,
		Carry:    []uuid.UUID{id},
	}.Apply(ctx)
}, constTargets(func() *game.TargetSpec {
	return TargetCreature("up to one target creature").WithCount(0, 1)
}))

// craigBooneChoice deals the damage as the chosen branch says. An on-disk
// key: never renamed, never reused.
var craigBooneChoice = OptionPickThen("option-pick/craig-boone-damage", func(ctx *Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 {
		return nil
	}
	g := ctx.Game
	target := r.Carry[0]
	n := 0
	if c, ok := g.LookupCardForEffect(ctx.Item.SourceCardID); ok && onBattlefield(g, c.InstanceID) {
		n = c.Counters["quest"]
	} else if info, ok := g.LastKnownPermanentForEffect(ctx.Item.SourceCardID); ok {
		n = info.Counters["quest"]
	}
	if n <= 0 {
		return nil
	}
	if r.Index == 1 {
		return DealDamage{Source: ctx.Item.SourceCardID, Target: r.Chooser, Amount: n}.Apply(ctx)
	}
	if !onBattlefield(g, target) {
		return nil
	}
	return DealDamage{Source: ctx.Item.SourceCardID, Target: target, Amount: n}.Apply(ctx)
})
