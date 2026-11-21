package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ajani, Sleeper Agent — Legendary Planeswalker — Ajani {1}{G}{G/W/P}{W},
// starting loyalty 4 (Dominaria United, #5487):
//
//	"Compleated ({G/W/P} can be paid with {G}, {W}, or 2 life. If life was
//	 paid, this planeswalker enters with two fewer loyalty counters.)
//	 +1: Reveal the top card of your library. If it's a creature or
//	 planeswalker card, put it into your hand. Otherwise, you may put it on
//	 the bottom of your library.
//	 −3: Distribute three +1/+1 counters among up to three target creatures.
//	 They gain vigilance until end of turn.
//	 −6: You get an emblem with "Whenever you cast a creature or planeswalker
//	 spell, target opponent gets two poison counters.""
//
// Compleated is ADR 0183's replacement (effects.Compleated). The +1 reveals
// the top card, takes it if it is a creature or planeswalker card and
// otherwise offers to bottom it (declining leaves it on top). The −3 divides
// the three counters as the ability is put on the stack (CR 601.2d); each
// target is rechecked on resolution (CR 608.2b), so one that has gone takes
// no counters and its share is not moved, and the others still do. The emblem
// is the controller's alone (CR 114.4): its trigger takes its target when put
// on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "37714eb9-a39d-4534-a2d3-27908c418f8a",
		Name:            "Ajani, Sleeper Agent",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		PrintedKeywords: []string{"compleated"},
		Replacements:    []game.ReplacementEffect{effects.Compleated()},
		Emblem: &effects.EmblemSpec{
			Label: "Ajani, Sleeper Agent emblem",
			Text:  "Whenever you cast a creature or planeswalker spell, target opponent gets two poison counters.",
			Triggered: []game.TriggeredAbility{
				effects.Targeting(
					effects.WheneverYouCast(effects.Or(effects.Creature(), effects.Planeswalker()),
						"Ajani, Sleeper Agent emblem — target opponent gets two poison counters",
						ajaniSleeperAgentPoison),
					effects.TargetPlayer("target opponent", effects.Opponent())),
			},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:  "+1: Reveal the top card of your library. If it's a creature or planeswalker card, put it into your hand. Otherwise, you may put it on the bottom of your library.",
				Cost:   effects.LoyaltyCost(1),
				Effect: ajaniSleeperAgentReveal,
			},
			{
				Label: "−3: Distribute three +1/+1 counters among up to three target creatures. They gain vigilance until end of turn.",
				Cost:  effects.LoyaltyCost(-3),
				Targets: effects.TargetCreature("up to three target creatures").
					WithCount(0, 3).Dividing(effects.Divide(3)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := effects.PutDividedCounters(ctx, game.CounterPlusOne); err != nil {
						return err
					}
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"vigilance"},
							Label: "Ajani, Sleeper Agent — vigilance until end of turn"}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Label:  "−6: You get an emblem with \"Whenever you cast a creature or planeswalker spell, target opponent gets two poison counters.\"",
				Cost:   effects.LoyaltyCost(-6),
				Effect: effects.Do(effects.CreateEmblem{}),
			},
		},
	})
}

func ajaniSleeperAgentPoison(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	t, ok := effects.FirstLegalTarget(ctx)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	return g.AddPlayerCounterByForEffect(item.Controller, t.ID, game.CounterPoison, 2)
}

func ajaniSleeperAgentReveal(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	revealed := g.RevealTopOfLibraryForEffect(item.Controller, item.SourceCardID, 1,
		"Ajani, Sleeper Agent — reveal the top card of your library")
	if len(revealed) == 0 {
		return nil
	}
	c, ok := g.LookupCardForEffect(revealed[0])
	if !ok {
		return nil
	}
	if c.IsCreature() || c.IsPlaneswalker() {
		return g.BounceToHandForEffect(revealed[0])
	}
	return effects.MayChoice{
		Player:   ctx.Controller(),
		Question: "Ajani, Sleeper Agent — put " + c.Name + " on the bottom of your library?",
		YesLabel: "Put it on the bottom",
		NoLabel:  "Leave it on top",
		Then:     ajaniSleeperAgentBottom,
		Carry:    []uuid.UUID{revealed[0]},
	}.Apply(ctx)
}

var ajaniSleeperAgentBottom = effects.MayYes("s82-dmu-b16/ajani-bottom", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	id := effects.CarryAt(r, 0)
	if !effects.BisonWhistleStillInLibrary(ctx.Game, id) {
		return nil
	}
	return ctx.Game.PutOnBottomInRandomOrderForEffect(r.Chooser, game.ZoneLibrary, []uuid.UUID{id})
})
