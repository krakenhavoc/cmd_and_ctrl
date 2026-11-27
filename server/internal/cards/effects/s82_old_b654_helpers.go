package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b654_helpers.go — Invasion, batch 654 (S82, #5616).

// SpellOrAbilityTargetingYourLand is the clause "an opponent controls that
// targets a land you control" (Teferi's Response): the object is controlled by
// an opponent of the chooser and one of its targets is a land the chooser
// controls.
func SpellOrAbilityTargetingYourLand() StackItemPredicate {
	return func(g *game.Game, chooser uuid.UUID, item *game.StackItem) bool {
		if item.Controller == chooser {
			return false
		}
		for _, t := range item.Targets {
			if t.Kind != game.TargetCard {
				continue
			}
			if c, ok := g.LookupCardForEffect(t.ID); ok && c.IsLand() && c.Controller == chooser {
				return true
			}
		}
		return false
	}
}

// TeferisResponseResolve is "Counter target spell or ability an opponent
// controls that targets a land you control. If a permanent's ability is
// countered this way, destroy that permanent. Draw two cards." The ability's
// source is read before the counter removes the item; the permanent is
// destroyed only if the same object is still on the battlefield. A target that
// has left the stack makes the whole spell do nothing (CR 608.2b).
func TeferisResponseResolve(item *game.StackItem, ctx *Context) error {
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	target := ctx.Game.StackItemForEffect(t.ID)
	var source uuid.UUID
	if target != nil && target.Kind != game.StackItemSpell && !ctx.Game.AbilitySourceIsNewObjectForEffect(target) {
		if c, found := ctx.Game.LookupCardForEffect(target.SourceCardID); found && OnBattlefield(ctx.Game, c.InstanceID) {
			source = c.InstanceID
		}
	}
	if err := (CounterTarget{StackID: t.ID}).Apply(ctx); err != nil {
		return err
	}
	if source != uuid.Nil {
		if err := (DestroyTarget{Target: source}).Apply(ctx); err != nil {
			return err
		}
	}
	return DrawCards{N: 2}.Apply(ctx)
}

// SwayOfIllusionResolve is "Any number of target creatures become the color
// of your choice until end of turn. Draw a card." The colour is chosen once
// as the spell resolves; the draw follows the change. With no target declared
// there is nothing to colour and the draw still happens; if every declared
// target has become illegal the spell does nothing (CR 608.2b).
func SwayOfIllusionResolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if len(ctx.Targets()) == 0 {
		return DrawCards{N: 1}.Apply(ctx)
	}
	if len(ctx.LegalTargets()) == 0 {
		return nil
	}
	ChooseColorTail(game.ColorForBenefit, g, item.Controller, item.SourceCardID,
		"Sway of Illusion — choose a color", ctx.ItemTail(swayOfIllusionColorChosen))
	return nil
}

// swayOfIllusionColorChosen recolours every target still legal and draws. An
// on-disk key.
var swayOfIllusionColorChosen = CardTail("sway-of-illusion/color-chosen", func(ctx *Context, o game.TailOutcome) error {
	if color := o.Tail.Word; color != "" {
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			if err := UntilEndOfTurn(ctx, t.ID, nil, "Sway of Illusion — becomes "+game.ColorName(color), game.SetColorsMod(color)); err != nil {
				return err
			}
		}
	}
	return DrawCards{N: 1}.Apply(ctx)
})

// DredgeResolve is "Sacrifice a creature or land. Draw a card.": the
// controller chooses one of their creatures and lands to sacrifice, and draws
// whether or not there was one to sacrifice (CR 701.21a).
func DredgeResolve(item *game.StackItem, ctx *Context) error {
	candidates := B04CreatureIDsControlledBy(ctx.Game, item.Controller)
	for _, id := range LandsControlledByPlayer(ctx.Game, item.Controller) {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && !c.IsCreature() {
			candidates = append(candidates, id)
		}
	}
	return SacrificeChoice{
		Player:     item.Controller,
		Candidates: candidates,
		Question:   "Dredge — sacrifice a creature or land",
		ThenTail:   ctx.ItemTail(dredgeSacrificed),
	}.Apply(ctx)
}

// dredgeSacrificed is "Draw a card", after the sacrifice. An on-disk key.
var dredgeSacrificed = CardTail("then/dredge-sacrificed", func(ctx *Context, _ game.TailOutcome) error {
	return DrawCards{N: 1}.Apply(ctx)
})
