package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const eowynLabel = "Éowyn, Fearless Knight — exile target creature an opponent controls with greater power; legendary creatures you control gain protection from each of its colors until end of turn"

// Éowyn, Fearless Knight — Legendary Creature — Human Knight {2}{R}{W}, 3/4
// (The Lord of the Rings: Tales of Middle-earth Commander):
//
//	"Haste
//	 When Éowyn enters, exile target creature an opponent controls with
//	 greater power. Legendary creatures you control gain protection from
//	 each of that creature's colors until end of turn."
//
// Haste is printed. The target clause is built from Éowyn's power as the
// trigger goes on the stack (TargetsFrom: power at least one more than
// hers), and checked again as it resolves, so a target that is no longer
// larger than Éowyn is illegal (CR 608.2b). The creature's colors are read
// as it last stood on the battlefield, before it is exiled (CR 608.2h);
// the legendary creatures that gain protection are those its controller
// controls as the ability resolves (CR 611.2c). A colorless creature gives
// no protection.
//
// No simplification.
func init() {
	enters := effects.WhenThisEnters(eowynLabel, eowynExileAndProtect)
	enters.TargetsFrom = func(_ game.TriggerContext, source *game.Card, _ *game.Game) *game.TargetSpec {
		return effects.TargetCreature("target creature an opponent controls with greater power",
			effects.OpponentControls(), effects.PowerGE(source.PowerForComparison()+1))
	}
	effects.Register(effects.Spec{
		OracleID:        "dcc349e9-399d-4f33-8131-d1801272466a",
		Name:            "Éowyn, Fearless Knight",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered:       []game.TriggeredAbility{enters},
	})
}

func eowynExileAndProtect(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	target, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	victim, ok := g.LookupCardForEffect(target)
	if !ok {
		return nil
	}
	if src, ok := g.LookupCardForEffect(item.SourceCardID); ok && victim.PowerForComparison() <= src.PowerForComparison() {
		return nil
	}
	tail := ctx.ItemTail(eowynProtectTail)
	tail.Words = effects.GapB127ColorProtections(victim)
	return effects.ExileTarget{Target: target, ThenTail: tail}.Apply(ctx)
}

// eowynProtectTail is the second sentence, run once the exile has settled
// (the exile can pause on a replacement): the protection tokens ride the
// tail as its words, read from the creature before it left. It applies
// whether or not the exile happened, as printed. An on-disk key.
var eowynProtectTail = effects.CardTail("eowyn-fearless-knight/protect", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.Tail.Words) == 0 {
		return nil
	}
	return effects.GrantKeywordUntilEOT{
		Match:    effects.And(effects.Creature(), effects.YouControl(), effects.Legendary()),
		Keywords: o.Tail.Words,
		Label:    "Éowyn, Fearless Knight — protection from each of that creature's colors",
	}.Apply(ctx)
})
