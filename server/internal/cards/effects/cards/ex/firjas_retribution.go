package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Firja's Retribution — Enchantment — Saga {1}{W}{W}{B} (Commander Legends:
// Battle for Baldur's Gate, #5948):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Create a 4/4 white Angel Warrior creature token with flying and
//	     vigilance.
//	 II — Until end of turn, Angels you control gain "{T}: Destroy target
//	      creature with power less than this creature's power."
//	 III — Angels you control gain double strike until end of turn."
//
// Chapter I is the printed token. Chapter II hands a bundle to the Angels you
// control as it resolves (The Grand Evolution's GrantAbilitiesFor; CR 611.2c
// locks the set in): the ability's {T} is the creature's own, so one that
// entered this turn can't use it (CR 302.6), and "this creature's power" is
// the granted creature's, judged at announce and again at resolution
// (RelativeToSource). Chapter III grants double strike the same way, to the
// Angels you control when it resolves.
//
// No simplification.
const firjasRetributionDestroyGrant = "firjas-retribution/destroy-lesser-power"

func init() {
	angels := effects.And(effects.Creature(), effects.YouControl(), effects.Subtype("Angel"))
	effects.Register(effects.Spec{
		OracleID:     "fea32d4e-ceef-4954-a84d-e7e2509e4a13",
		Name:         "Firja's Retribution",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: firjasRetributionDestroyGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: Destroy target creature with power less than this creature's power.",
				Cost:    game.AbilityCost{Tap: true},
				Purpose: game.Purpose{Answers: game.AnswerRemove},
				Targets: effects.RelativeToSource(
					effects.TargetCreature("target creature with power less than this creature's power"), effects.LesserPower()),
				Effect: effects.DestroyFirstLegalCardTarget,
			}},
			Text: "{T}: Destroy target creature with power less than this creature's power.",
		}},
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel("Firja's Retribution", 1, "create a 4/4 white Angel Warrior creature token with flying and vigilance"),
				effects.Do(effects.CreateToken{Template: effects.TokenCard("4/4 white Angel Warrior with flying and vigilance"), N: 1})),
			effects.ChapterTrigger(2, effects.SagaChapterLabel("Firja's Retribution", 2, "until end of turn, Angels you control gain \"{T}: Destroy target creature with power less than this creature's power.\""),
				func(g *game.Game, item *game.StackItem) error {
					return effects.GrantAbilitiesFor{
						Match: angels,
						Keys:  []string{firjasRetributionDestroyGrant},
						Label: "Firja's Retribution — {T}: destroy target creature with power less than this creature's power",
					}.Apply(effects.NewContext(g, item))
				}),
			effects.ChapterTrigger(3, effects.SagaChapterLabel("Firja's Retribution", 3, "Angels you control gain double strike until end of turn"),
				func(g *game.Game, item *game.StackItem) error {
					return effects.GrantKeywordUntilEOT{Match: angels, Keywords: []string{"double strike"},
						Label: "Firja's Retribution — double strike"}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
