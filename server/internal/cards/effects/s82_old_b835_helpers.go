package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ice Age pump and combat tricks, batch 835 (#5636): helpers the cards of the
// batch read.

// oldB835SacrificeTheObjectBody is the delayed "sacrifice it" of a creature an
// ability named: the object, if it is still that object on the battlefield
// (CR 603.7c). Its current controller sacrifices it ("its controller
// sacrifices it"). An on-disk key.
var oldB835SacrificeTheObjectBody = game.DelayedBody("sacrifice/old-b835-the-object", func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	c, ok := g.LookupCardForEffect(p.Object.ID)
	if !ok || c.ObjectEpoch != p.Object.Epoch || !OnBattlefield(g, p.Object.ID) {
		return nil
	}
	return SacrificePermanent{Target: p.Object.ID}.Apply(NewContext(g, item))
})

// oldB835SacrificeAtNextEndStep schedules the sacrifice of `id` at the
// beginning of the next end step.
func oldB835SacrificeAtNextEndStep(ctx *Context, id uuid.UUID, label string) error {
	c, ok := ctx.Game.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	return ScheduleDelayedTrigger{
		At:     game.StepEnd,
		Label:  label + " — sacrifice " + c.Name,
		Cards:  []uuid.UUID{id},
		Body:   oldB835SacrificeTheObjectBody,
		Params: game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
	}.Apply(ctx)
}

// OldB835TargetGetsThenSacrificeAtEnd is Celestial Sword: the target gets
// +3/+3 until end of turn and is sacrificed at the beginning of the next end
// step.
func OldB835TargetGetsThenSacrificeAtEnd(power, toughness int, label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id := FirstLegalBattlefieldTarget(ctx)
		if id == uuid.Nil {
			return nil
		}
		if err := (BoostUntilEOT{Target: id, Power: power, Toughness: toughness, Label: label}).Apply(ctx); err != nil {
			return err
		}
		return oldB835SacrificeAtNextEndStep(ctx, id, label)
	}
}

// OldB835ThisGetsAndFliesThenSacrificeAtEnd is Goblin Ski Patrol: it gets
// +2/+0 and gains flying, and is sacrificed at the beginning of the next end
// step.
func OldB835ThisGetsAndFliesThenSacrificeAtEnd(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if !SourceIsStillThisPermanent(g, item) {
		return nil
	}
	label := "Goblin Ski Patrol"
	if err := (BoostUntilEOT{Target: item.SourceCardID, Power: 2, Label: label + " — +2/+0"}).Apply(ctx); err != nil {
		return err
	}
	if err := (GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"flying"}, Label: label + " — flying"}).Apply(ctx); err != nil {
		return err
	}
	return oldB835SacrificeAtNextEndStep(ctx, item.SourceCardID, label)
}

// OldB835TargetGetsThenSacrificeThisWhenItLeaves is Kjeldoran Elite Guard:
// "Target creature gets +P/+T until end of turn. When that creature leaves the
// battlefield this turn, sacrifice this creature." Runesword's delayed trigger
// on the target, sacrificing the source.
func OldB835TargetGetsThenSacrificeThisWhenItLeaves(power, toughness int, label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id := FirstLegalBattlefieldTarget(ctx)
		if id == uuid.Nil {
			return nil
		}
		if err := (BoostUntilEOT{Target: id, Power: power, Toughness: toughness, Label: label}).Apply(ctx); err != nil {
			return err
		}
		d := g.PinnedTo(g.UntilEndOfTurnDuration(), id)
		g.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
			Controller:   ctx.Controller(),
			SourceCardID: ctx.Source(),
			Label:        label + " — sacrifice this creature",
			On:           []game.EventKind{game.EventLTB},
			Condition:    theListedObjectLeftCondition,
			Cards:        []uuid.UUID{id},
			Duration:     &d,
			Body:         runeswordSacrificeBody,
		})
		return nil
	}
}

// OldB835DuringCombat is "Activate only during combat": any step from the
// beginning of combat to the end of combat.
func OldB835DuringCombat() ActivationCondition {
	return func(g *game.Game, _, _ uuid.UUID) bool {
		switch g.Turn.Step {
		case game.StepBeginCombat, game.StepDeclareAttackers, game.StepDeclareBlockers,
			game.StepFirstStrikeDamage, game.StepCombatDamage, game.StepEndCombat:
			return true
		}
		return false
	}
}

// OldB835BlockedByAtLeastOneThisDeclareBlockers is Grizzled Wolverine's
// "only during the declare blockers step, only if at least one creature is
// blocking this creature".
func OldB835BlockedByAtLeastOneThisDeclareBlockers() ActivationCondition {
	return func(g *game.Game, _, source uuid.UUID) bool {
		return g.Turn.Step == game.StepDeclareBlockers && len(BlockersOf(g, source)) > 0
	}
}

// OldB835JohtullWurm is "it gets -2/-1 until end of turn for each creature
// blocking it beyond the first", counted as the trigger resolves.
func OldB835JohtullWurm(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		n := len(BlockersOf(g, item.SourceCardID)) - 1
		if n <= 0 {
			return nil
		}
		return BoostUntilEOT{Target: item.SourceCardID, Power: -2 * n, Toughness: -n, Label: label}.Apply(NewContext(g, item))
	}
}

// OldB835SnowMountain is Goblin Ski Patrol's "only if you control a snow
// Mountain" (the "only once" is the ability's OnlyOnce bit): a snow land with
// the Mountain type under your control, the type read as a word of the card's
// text (CR 612.2).
func OldB835SnowMountain() ActivationCondition {
	return func(g *game.Game, controller, source uuid.UUID) bool {
		src, ok := g.LookupCardForEffect(source)
		if !ok {
			return false
		}
		mountain := src.TextSubtype("Mountain")
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == controller && c.IsLand() && c.HasSupertype("snow") && c.HasSubtype(mountain) {
				return true
			}
		}
		return false
	}
}

// OldB835BattleCry is Battle Cry: untap all white creatures you control, and
// whenever a creature blocks this turn it gets +0/+1 until end of turn.
func OldB835BattleCry(_ *game.StackItem, ctx *Context) error {
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.IsCreature() && c.Controller == ctx.Controller() && c.HasColor(ctx.TextColor("W")) {
			if err := (UntapTarget{Target: c.InstanceID}).Apply(ctx.AsGroupMember()); err != nil {
				return err
			}
		}
	}
	return DelayedOnEvent{
		Label:     "Battle Cry — the blocking creature gets +0/+1 until end of turn",
		On:        []game.EventKind{game.EventBlock},
		Condition: anyCreatureBlocksOldB835Condition,
		Body:      oldB835BlockerGetsBody,
		Repeats:   true,
	}.Apply(ctx)
}

// "Whenever a creature blocks": an EventBlock, once per blocker however many
// attackers it blocks (CR 509.3a). An on-disk key.
var anyCreatureBlocksOldB835Condition = game.DelayedCondition("block/old-b835-a-creature-blocks", func(ev game.Event, _ *game.DelayedTrigger, _ *game.Game, _ game.EffectParams) bool {
	return ev.Kind == game.EventBlock && ev.Amount <= 1
})

// oldB835BlockerGetsBody gives the creature that blocked +0/+1 until end of
// turn. An on-disk key.
var oldB835BlockerGetsBody = game.SimpleDelayedBody("block/old-b835-the-blocker-gets-0-1", func(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, id := range ctx.PayloadCards() {
		if !OnBattlefield(g, id) {
			continue
		}
		if err := (BoostUntilEOT{Target: id, Toughness: 1, Label: "Battle Cry — +0/+1"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
