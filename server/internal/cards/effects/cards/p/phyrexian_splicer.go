package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phyrexian Splicer — Artifact {2} (Tempest, S82 OLD b753, #5628):
//
//	"{2}, {T}, Choose flying, first strike, trample, or shadow: Until end of
//	 turn, target creature with the chosen ability loses it and another
//	 target creature gains it."
//
// The choice is made as the ability is activated, before its targets, and
// decides what the first target must have, so it is four activated rows, one
// per ability, each naming the choice in its effect line. A row's first target
// must have the ability (a creature that has lost it by resolution is an
// illegal target, CR 608.2b); the second is a different creature (Distinct,
// CR 601.2c). The loss is a layer-6 removal and the grant a layer-6 addition,
// each with its own timestamp (CR 613.1f): the grant is to the other creature,
// so neither undoes the other.
//
// No Purpose: the bot never pays to move an ability between creatures.
//
// No simplification.
func init() {
	var rows []effects.ActivatedAbility
	for _, kw := range []string{"flying", "first strike", "trample", "shadow"} {
		rows = append(rows, phyrexianSplicerRow(kw))
	}
	effects.Register(effects.Spec{
		OracleID:     "4433a5fd-2daa-440a-b9cc-4ad028358cbc",
		Name:         "Phyrexian Splicer",
		Completeness: effects.CompletenessFull,
		Activated:    rows,
	})
}

// phyrexianSplicerRow is the Splicer's ability with `kw` chosen.
func phyrexianSplicerRow(kw string) effects.ActivatedAbility {
	return effects.ActivatedAbility{
		Label: "{2}, {T}, Choose flying, first strike, trample, or shadow: Until end of turn, target creature with " + kw +
			" loses it and another target creature gains it.",
		Cost: effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
		Targets: effects.Clauses(
			effects.TargetCreature("target creature with "+kw, effects.HasKeyword(kw)),
			effects.Distinct(effects.TargetCreature("another target creature")),
		),
		Effect: func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			from, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			if err := effects.UntilEndOfTurn(ctx, from.ID, nil, "Phyrexian Splicer — loses "+kw,
				game.RemoveKeywordsMod(kw)); err != nil {
				return err
			}
			to, ok := ctx.ClauseTarget(1)
			if !ok {
				return nil
			}
			return effects.UntilEndOfTurn(ctx, to.ID, nil, "Phyrexian Splicer — gains "+kw,
				game.AddKeywordsMod(kw))
		},
	}
}
