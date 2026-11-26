package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b366_helpers.go — bodies shared by the Journey into Nyx strive
// cards of OLD batch 366 (#5558, S82). Each reads its targets as the spell
// resolves, so a target that left or stopped being legal is skipped (CR 608.2b).

// S82OldB366LegalTargetCards is the distinct card targets still legal as the
// resolving spell finishes, in announcement order. A permanent named by two of
// the spell's "target" words is listed once (CR 115.3).
func S82OldB366LegalTargetCards(ctx *Context) []uuid.UUID {
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard || ctx.IsNewSourceObject(t.ID) || slices.Contains(ids, t.ID) {
			continue
		}
		ids = append(ids, t.ID)
	}
	return ids
}

// S82OldB366PumpAndUntapEachTarget is "any number of target creatures each get
// +P/+T until end of turn. Untap those creatures." (Colossal Heroics).
func S82OldB366PumpAndUntapEachTarget(power, toughness int, label string) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		for _, id := range S82OldB366LegalTargetCards(ctx) {
			if err := (BoostUntilEOT{Target: id, Power: power, Toughness: toughness, Label: label}).Apply(ctx); err != nil {
				return err
			}
		}
		for _, id := range S82OldB366LegalTargetCards(ctx) {
			if err := (UntapTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}
}

// S82OldB366WithAttachedAuras is ids followed by every Aura on the battlefield
// attached to one of them (Silence the Believers), each listed once.
func S82OldB366WithAttachedAuras(ctx *Context, ids []uuid.UUID) []uuid.UUID {
	out := slices.Clone(ids)
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if !c.IsAura() || c.AttachedTo.Kind != game.TargetCard || !slices.Contains(ids, c.AttachedTo.ID) {
			continue
		}
		if !slices.Contains(out, c.InstanceID) {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// s82OldB366Spheres is Hour of Need's continuation, run once the exile has
// settled: the tail's Carry is the targets (the first N) and each one's
// controller as it last existed on the battlefield (CR 608.2h); a token is
// made for each target that really went to exile, for that controller.
var s82OldB366Spheres = CardTail("hour-of-need/sphinx-for-each-exiled", func(ctx *Context, o game.TailOutcome) error {
	n := o.Tail.N
	if n <= 0 || len(o.Tail.Carry) < 2*n {
		return nil
	}
	for i := 0; i < n; i++ {
		if !slices.Contains(o.IDs, o.Tail.Carry[i]) {
			continue
		}
		controller := o.Tail.Carry[n+i]
		if controller == uuid.Nil {
			continue
		}
		if err := (CreateToken{Controller: controller, Template: TokenCard("4/4 blue Sphinx with flying"), N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})

// S82OldB366ExileAndMakeSpheres is Hour of Need: exile the legal targets as
// one event, then each exiled creature's controller creates a 4/4 blue
// Sphinx creature token with flying.
func S82OldB366ExileAndMakeSpheres(_ *game.StackItem, ctx *Context) error {
	ids := S82OldB366LegalTargetCards(ctx)
	if len(ids) == 0 {
		return nil
	}
	controllers := make([]uuid.UUID, len(ids))
	for i, id := range ids {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok {
			controllers[i] = c.Controller
		}
	}
	tail := ctx.ItemTail(s82OldB366Spheres)
	tail.N = len(ids)
	tail.Carry = append(slices.Clone(ids), controllers...)
	return ctx.Game.ExileCardsKeyedForEffect(ids, tail)
}

// s82OldB366CopyModel is the keyed continuation of Polymorphous Rush's
// "choose a creature on the battlefield": each target creature still legal
// becomes a copy of the creature chosen, until end of turn.
var s82OldB366CopyModel = OptionPickThen("option-pick/polymorphous-rush-model", func(ctx *Context, r game.OptionPicked) error {
	if r.Index < 0 || r.Index >= len(r.Carry) {
		return nil
	}
	model := r.Carry[r.Index]
	if _, ok := ctx.Game.LookupCardForEffect(model); !ok {
		return nil
	}
	var targets []uuid.UUID
	for _, id := range S82OldB366LegalTargetCards(ctx) {
		if id != model {
			targets = append(targets, id)
		}
	}
	if len(targets) == 0 {
		return nil
	}
	return BecomeCopy{Targets: targets, Of: model, Label: "Polymorphous Rush — becomes a copy until end of turn"}.Apply(ctx)
})

// S82OldB366ChooseCreatureThenCopy is Polymorphous Rush's body: the controller
// chooses one creature among every creature on the battlefield, then the
// targets copy it. With no creature to choose nothing happens.
func S82OldB366ChooseCreatureThenCopy(item *game.StackItem, ctx *Context) error {
	if len(S82OldB366LegalTargetCards(ctx)) == 0 {
		return nil
	}
	var ids []uuid.UUID
	var options []game.ChoiceOption
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if !c.IsCreature() {
			continue
		}
		ids = append(ids, c.InstanceID)
		options = append(options, game.ChoiceOption{Label: c.Name, Cards: []uuid.UUID{c.InstanceID}})
	}
	if len(ids) == 0 {
		return nil
	}
	return PickOption{
		Question: "Polymorphous Rush — choose a creature on the battlefield",
		Options:  options,
		ThenKey:  s82OldB366CopyModel,
		Carry:    ids,
		WithItem: true,
	}.Apply(ctx)
}
