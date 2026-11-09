package effects

import (
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
	Register(Spec{
		OracleID:     cliveOracle,
		Name:         "Clive, Ifrit's Dominant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Clive, Ifrit's Dominant — you may discard your hand, then draw cards equal to your devotion to red",
				func(g *game.Game, item *game.StackItem) error {
					return MayChoice{
						Question: "Clive, Ifrit's Dominant — discard your hand, then draw cards equal to your devotion to red?",
						Then:     cliveDiscardAndDraw,
					}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label:        "{4}{R}{R}, {T}: Exile Clive, then return it to the battlefield transformed under its owner's control. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{4}{R}{R}"), TapCost()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return ExileAndReturnTransformed{Target: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
	})
	brimstone := func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		if err := (AddMana{Produced: "{R}{R}{R}{R}"}).Apply(ctx); err != nil {
			return err
		}
		src, ok := g.LookupCardForEffect(item.SourceCardID)
		if !ok || ctx.isNewSourceObject(item.SourceCardID) || src.Counters[game.CounterLore] < 3 || !onBattlefield(g, item.SourceCardID) {
			return nil
		}
		return Flicker{Target: item.SourceCardID}.Apply(ctx)
	}
	Register(Spec{
		OracleID:     cliveOracle + "#1",
		Name:         "Ifrit, Warden of Inferno",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, SagaChapterLabel("Ifrit, Warden of Inferno", 1, "Lunge — fights up to one other target creature"),
				Another(TargetCreature("up to one other target creature").WithCount(0, 1)),
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							return b10Fight(ctx, item.SourceCardID, t.ID)
						}
					}
					return nil
				}),
			ChapterTrigger(2, SagaChapterLabel("Ifrit, Warden of Inferno", 2, "Brimstone — add {R}{R}{R}{R}; with three or more lore counters, exile it, then return it front face up"), brimstone),
			ChapterTrigger(3, SagaChapterLabel("Ifrit, Warden of Inferno", 3, "Brimstone — add {R}{R}{R}{R}; with three or more lore counters, exile it, then return it front face up"), brimstone),
		},
	})
}

// cliveDiscardAndDraw is the "yes": discard the hand, then draw cards equal
// to the devotion to red as it stands now. An on-disk key.
var cliveDiscardAndDraw = MayYes("may/clive-ifrits-dominant-discard-and-draw", func(ctx *Context, _ game.ConfirmAnswer) error {
	if _, err := discardWholeHand(ctx.Game, ctx.Controller()); err != nil {
		return err
	}
	n := devotionTo(ctx.Game, ctx.Controller(), ctx.TextColor("R"))
	if n <= 0 {
		return nil
	}
	return DrawCards{N: n}.Apply(ctx)
})
