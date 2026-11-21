package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Katsumasa, the Animator — Legendary Creature — Moonfolk Artificer
// {2}{U}{U}, 3/3 (Neon Dynasty):
//
//	"Flying
//	 {2}{U}: Until end of turn, target noncreature artifact you control
//	 becomes an artifact creature and gains flying. If it's not a Vehicle,
//	 it has base power and toughness 1/1 until end of turn.
//	 At the beginning of your upkeep, put a +1/+1 counter on each of up to
//	 three target noncreature artifacts."
//
// The animation is ONE effect at one timestamp (CR 613.7): a layer-4 type set
// (Suit Up's SetTypesMod, so an artifact land or enchantment stops being one),
// flying, and, only when the target is not a Vehicle as the ability resolves,
// the layer-7b 1/1 (a Vehicle keeps its printed power and toughness). The
// upkeep trigger is Leo's Guidance's up-to-three targets over noncreature
// artifacts of any controller, counters on each still-legal target
// (CR 608.2b).
//
// No simplification.
func init() {
	const upkeep = "Katsumasa, the Animator — put a +1/+1 counter on each of up to three target noncreature artifacts"
	trigger := effects.AtYourUpkeep(upkeep, func(g *game.Game, item *game.StackItem) error {
		return effects.PutPlusOneCounterOnEachLegalTarget(g, item)
	})
	trigger.Targets = effects.TargetPermanent("up to three target noncreature artifacts",
		effects.Artifact(), effects.Noncreature()).WithCount(0, 3)
	effects.Register(effects.Spec{
		OracleID:        "80ba76a6-cc48-4820-bc0b-c9ec18088a3b",
		Name:            "Katsumasa, the Animator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}: Until end of turn, target noncreature artifact you control becomes an artifact creature and gains flying. If it's not a Vehicle, it has base power and toughness 1/1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{2}{U}"),
			Targets: effects.TargetPermanent("target noncreature artifact you control", effects.Artifact(), effects.Noncreature(), effects.YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					c, ok := g.LookupCardForEffect(t.ID)
					if !ok {
						return nil
					}
					mods := []game.Mod{
						game.SetTypesMod([]string{"Artifact", "Creature"}, nil),
						game.AddKeywordsMod("flying"),
					}
					if !c.HasSubtype("Vehicle") {
						mods = append(mods, game.SetBasePTMods(1, 1)...)
					}
					return effects.ScopedEffectFor{
						Target:   t.ID,
						Mods:     mods,
						Duration: effects.DurationUntilEndOfTurn(ctx),
						Label:    "Katsumasa, the Animator — becomes an artifact creature with flying until end of turn",
					}.Apply(ctx)
				}
				return nil
			},
		}},
		Triggered: []game.TriggeredAbility{trigger},
	})
}
