package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Winter's Chill — Instant {X}{U}:
//
//	"Cast this spell only during combat before blockers are declared.
//	 X can't be greater than the number of snow lands you control.
//	 Choose X target attacking creatures. For each of those creatures,
//	 its controller may pay {1} or {2}. If that player doesn't, destroy
//	 that creature at end of combat. If that player pays only {1},
//	 prevent all combat damage that would be dealt to and dealt by that
//	 creature this combat."
//
// The timing is a CastCondition: the beginning of combat or the declare
// attackers step, in any combat phase (the 2013-09-20 ruling), as Blaze
// of Glory's. The X ceiling is XCeilingSnowLandsYouControl (#2581), and
// the target count is the announced X (CountFromX).
//
// The payments are made as the spell resolves (the 2004-10-04 ruling,
// CR 118.12), and they are one choice with three outcomes: pay {2} and
// the creature fights as normal, pay {1} and it neither deals nor is
// dealt combat damage, pay nothing and it is destroyed at end of combat
// (#2854). That is an option_pick whose options carry their mana cost:
// an option its controller cannot pay is not offered (CR 118.3), so a
// player with one untapped land sees "pay nothing" and "pay {1}" only.
// The engine pays the chosen option through the auto-tapper.
//
// One prompt per creature still a legal target (CR 608.2b), in target
// order, asked of its controller. That is always the active player: a
// creature whose controller changes is removed from combat (CR 506.4),
// so it is no longer an attacking creature and no longer a legal
// target. Each answer is settled before the next question, so a player
// who pays {2} for one creature is offered only what they have left for
// the next.
//
// The {1} branch is one prevention record over both directions
// (AndDealtBy, Maze of Ith's shape) lasting this combat (#2027). The
// no-payment branch is a delayed trigger at the beginning of this
// combat's end of combat step that destroys the creature if it is still
// the same object (CR 400.7, 603.7).
//
// The continuation is a registered key (OptionPickThen), so a table
// waiting on any of these questions is still a restore point.
//
// No simplifications.
func init() {
	wintersChillAnswer = OptionPickThen("option-pick/winters-chill-pay", wintersChillAnswered)
	Register(Spec{
		OracleID:     "ae78d498-11c7-427f-9d13-3f1b7096f803",
		Name:         "Winter's Chill",
		Completeness: CompletenessFull,
		CastCondition: func(g *game.Game, _ uuid.UUID, _ game.Card) bool {
			return g.Turn.Step == game.StepBeginCombat || g.Turn.Step == game.StepDeclareAttackers
		},
		CastConditionLabel: "Cast this spell only during combat before blockers are declared.",
		XCeiling:           XCeilingSnowLandsYouControl,
		Targets:            wintersChillTargets(),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			var creatures []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					creatures = append(creatures, t.ID)
				}
			}
			return wintersChillAsk(ctx, creatures)
		},
	})
}

// wintersChillTargets is "X target attacking creatures".
func wintersChillTargets() *game.TargetSpec {
	spec := TargetCreature("X target attacking creatures", AttackingCreature())
	spec.CountFromX = true
	return spec
}

// wintersChillAnswer is the registered continuation of each question.
// Its key is an on-disk identity: never renamed, never reused.
//
// Assigned in init, not in the declaration: the continuation asks
// the next question, which names this key, and a package-level
// initializer may not refer to itself (an initialization cycle).
var wintersChillAnswer game.OptionPickThen

// wintersChillAsk asks about the first creature in `creatures` that is
// still on the battlefield, carrying the rest. Nothing is asked once
// every creature has been answered for.
func wintersChillAsk(ctx *Context, creatures []uuid.UUID) error {
	for i, id := range creatures {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || !onBattlefield(ctx.Game, id) {
			continue
		}
		return PickOption{
			Player:   c.Controller,
			Question: "Winter's Chill — pay {1} or {2} for " + c.Name + "?",
			Options: []game.ChoiceOption{
				{Label: "Pay nothing: destroy " + c.Name + " at end of combat"},
				{Label: "Pay {1}: prevent all combat damage dealt to and by " + c.Name + " this combat", ManaCost: "{1}"},
				{Label: "Pay {2}: " + c.Name + " is unaffected", ManaCost: "{2}"},
			},
			ThenKey: wintersChillAnswer,
			Carry:   append([]uuid.UUID(nil), creatures[i:]...),
		}.Apply(ctx)
	}
	return nil
}

// wintersChillAnswered settles the answer about Carry[0] and asks about
// the next creature. The branch is read off what was PAID, which is the
// card's own test ("if that player pays only {1}"): nothing, {1} or {2}.
// Nobody choosing (the controller left the game, and the creature with
// them) is not paying.
func wintersChillAnswered(ctx *Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 {
		return nil
	}
	id, rest := r.Carry[0], r.Carry[1:]
	paid := ""
	if r.Option != nil {
		paid = r.Option.ManaCost
	}
	if err := wintersChillSettle(ctx, id, paid); err != nil {
		return err
	}
	return wintersChillAsk(ctx, rest)
}

func wintersChillSettle(ctx *Context, id uuid.UUID, paid string) error {
	c, ok := ctx.Game.LookupCardForEffect(id)
	if !ok || !onBattlefield(ctx.Game, id) {
		return nil
	}
	switch paid {
	case "{2}":
		return nil
	case "{1}":
		return PreventDamageFromSource{
			Protect: ShieldObject(id), AndDealtBy: true, CombatOnly: true, Lasts: ShieldThisCombat,
			Label: "Winter's Chill — prevent all combat damage dealt to and by " + c.Name + " this combat",
		}.Apply(ctx)
	}
	return ScheduleDelayedTrigger{
		At:     game.StepEndCombat,
		Label:  "Winter's Chill — destroy " + c.Name,
		Cards:  []uuid.UUID{id},
		Body:   destroyTheObjectBody,
		Params: game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
	}.Apply(ctx)
}
