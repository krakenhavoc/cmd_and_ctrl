package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Clive, Ifrit's Dominant // Ifrit, Warden of Inferno — a transforming card
// (Final Fantasy, #4462).
//
// Front, Clive, Ifrit's Dominant — Legendary Creature — Human Noble Warrior
// {4}{R}{R}, 5/5:
//
//	"When Clive enters, you may discard your hand, then draw cards equal to
//	 your devotion to red.
//	 {4}{R}{R}, {T}: Exile Clive, then return it to the battlefield
//	 transformed under its owner's control. Activate only as a sorcery."
//
// Back, Ifrit, Warden of Inferno — Legendary Enchantment Creature — Saga
// Demon, 9/9:
//
//	"(As this Saga enters and after your draw step, add a lore counter.)
//	 I — Lunge — Ifrit fights up to one other target creature.
//	 II, III — Brimstone — Add {R}{R}{R}{R}. If Ifrit has three or more lore
//	 counters on it, exile it, then return it to the battlefield (front
//	 face up)."
//
// The enter trigger's "may" is asked as it resolves; devotion is counted
// after the hand is gone, off the permanents then on the battlefield (the
// printed ruling: Clive is in the count). The transform is the exile-and-
// return verb (CR 712.14a), which makes Ifrit a new object that enters as a
// Saga and so gets its first lore counter. Chapters II and III are each an
// ordinary Saga chapter whose mana is added as the ability resolves (it
// empties with the phase, CR 106.4); the flicker back to the front face is
// gated by the live lore count, and the Saga is not sacrificed afterwards
// because by then it is a creature again (CR 714.4 checks Sagas only).
func init() {
	const cliveOracle = "ceba3d03-be5a-4dd7-b829-7db216a0f44e"
	effects.Register(effects.Spec{
		OracleID:     cliveOracle,
		Name:         "Clive, Ifrit's Dominant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Clive, Ifrit's Dominant — you may discard your hand, then draw cards equal to your devotion to red",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayChoice{
						Question: "Clive, Ifrit's Dominant — discard your hand, then draw cards equal to your devotion to red?",
						Then:     cliveDiscardAndDraw,
					}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{4}{R}{R}, {T}: Exile Clive, then return it to the battlefield transformed under its owner's control. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{4}{R}{R}"), effects.TapCost()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ExileAndReturnTransformed{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
	brimstone := func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		if err := (effects.AddMana{Produced: "{R}{R}{R}{R}"}).Apply(ctx); err != nil {
			return err
		}
		src, ok := g.LookupCardForEffect(item.SourceCardID)
		if !ok || ctx.IsNewSourceObject(item.SourceCardID) || src.Counters[game.CounterLore] < 3 || !effects.OnBattlefield(g, item.SourceCardID) {
			return nil
		}
		return effects.Flicker{Target: item.SourceCardID}.Apply(ctx)
	}
	effects.Register(effects.Spec{
		OracleID:     cliveOracle + "#1",
		Name:         "Ifrit, Warden of Inferno",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTriggerTargeting(1, effects.SagaChapterLabel("Ifrit, Warden of Inferno", 1, "Lunge — fights up to one other target creature"),
				effects.Another(effects.TargetCreature("up to one other target creature").WithCount(0, 1)),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							return effects.B10Fight(ctx, item.SourceCardID, t.ID)
						}
					}
					return nil
				}),
			effects.ChapterTrigger(2, effects.SagaChapterLabel("Ifrit, Warden of Inferno", 2, "Brimstone — add {R}{R}{R}{R}; with three or more lore counters, exile it, then return it front face up"), brimstone),
			effects.ChapterTrigger(3, effects.SagaChapterLabel("Ifrit, Warden of Inferno", 3, "Brimstone — add {R}{R}{R}{R}; with three or more lore counters, exile it, then return it front face up"), brimstone),
		},
	})
}

// cliveDiscardAndDraw is the "yes": discard the hand, then draw cards equal
// to the devotion to red as it stands now. An on-disk key.
var cliveDiscardAndDraw = effects.MayYes("may/clive-ifrits-dominant-discard-and-draw", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	if _, err := effects.DiscardWholeHand(ctx.Game, ctx.Controller()); err != nil {
		return err
	}
	n := effects.DevotionTo(ctx.Game, ctx.Controller(), ctx.TextColor("R"))
	if n <= 0 {
		return nil
	}
	return effects.DrawCards{N: n}.Apply(ctx)
})
