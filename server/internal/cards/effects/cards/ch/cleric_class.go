package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cleric Class — Enchantment — Class {W} (Adventures in the Forgotten
// Realms):
//
//	"(Gain the next level as a sorcery to add its ability.)
//	 If you would gain life, you gain that much life plus 1 instead.
//	 {3}{W}: Level 2
//	 Whenever you gain life, put a +1/+1 counter on target creature you
//	 control.
//	 {4}{W}: Level 3
//	 When this Class becomes level 3, return target creature card from your
//	 graveyard to the battlefield. You gain life equal to that creature's
//	 toughness."
//
// One line per level (ADR 0071):
//
//   - Level 1 is Leyline of Hope's replacement on the life gained,
//     ungated (CR 716.3).
//   - Level 2 is Wizard Class's targeted trigger, gated at level 2, on the
//     life-gain event. It exists only from level 2, so a gain made at
//     level 1 puts nothing on anything.
//   - Level 3 is BecomesLevel's trigger with a graveyard target. The
//     creature returns first; "that creature's toughness" is read from the
//     permanent as it landed (a keyed tail, so a replacement-ordering
//     prompt cannot lose it), or from its last-known information if it
//     left already (CR 608.2h). A creature that did not enter gains
//     nothing. The gain runs the level 1 replacement, so it is +1 more.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f0ead7f5-5c82-40fd-8e89-3e3c5d31fbad",
		Name:         "Cleric Class",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.LifeGainBecomes{
				Amount: func(n int) int { return n + 1 },
				Scope:  effects.LifeGainByController,
				Label:  "Cleric Class: gain that much life plus 1",
			}.Build(),
		},
		Activated: []effects.ActivatedAbility{
			effects.LevelUp(2, effects.ManaCost("{3}{W}")),
			effects.LevelUp(3, effects.ManaCost("{4}{W}")),
		},
		Triggered: []game.TriggeredAbility{
			effects.AtLevel(2, effects.Targeting(
				effects.WheneverYouGainLife("Cleric Class — put a +1/+1 counter on target creature you control",
					effects.PutPlusOneCounterOnEachLegalTarget),
				effects.TargetCreature("target creature you control", effects.YouControl()))),
			effects.Targeting(
				effects.BecomesLevel(3, "Cleric Class — level 3: return target creature card from your graveyard to the battlefield, then gain life equal to its toughness",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						ids := effects.LegalTargetCardIDs(ctx)
						if len(ids) == 0 {
							return nil
						}
						return g.ReturnFromGraveyardToBattlefieldKeyedForEffect(ids[0], ctx.Controller(), false,
							ctx.ItemTail(clericClassGainToughness))
					}),
				effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn())),
		},
	})
}

// clericClassGainToughness is "you gain life equal to that creature's
// toughness" as a keyed tail, told the permanent that entered. An on-disk key.
var clericClassGainToughness = effects.EnteredTail("cleric-class/gain-life-equal-to-toughness", func(ctx *effects.Context, entered uuid.UUID, _ game.TailOutcome) error {
	if entered == uuid.Nil {
		return nil
	}
	_, toughness, ok := effects.CmdrB13PowerAndToughness(ctx.Game, entered)
	if !ok || toughness <= 0 {
		return nil
	}
	return effects.GainLife{Amount: toughness}.Apply(ctx)
})
