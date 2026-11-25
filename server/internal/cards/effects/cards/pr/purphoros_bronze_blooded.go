package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Purphoros, Bronze-Blooded — Legendary Enchantment Creature — God {4}{R},
// 7/6 (Theros Beyond Death, #5514):
//
//	"Indestructible
//	 As long as your devotion to red is less than five, Purphoros isn't a
//	 creature.
//	 Other creatures you control have haste.
//	 {2}{R}: You may put a red creature card or an artifact creature card
//	 from your hand onto the battlefield. Sacrifice it at the beginning of
//	 the next end step."
//
// The God clause is the layer-4 devotion static every God shares; the haste
// is a layer-6 grant to its controller's other creatures. The ability is
// Incandescent Soulstoke's put from hand ("may" is a real decline), whose
// tail schedules the sacrifice of that object (CR 603.7c) and makes none when
// nothing was put. A multicoloured card with red, or a colourless artifact
// creature, is a legal pick; a nonartifact colourless card is not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6728b9d0-d3ab-4ba3-af13-63cb83e0ed99",
		Name:            "Purphoros, Bronze-Blooded",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Static: []game.StaticAbility{
			effects.GodUnlessDevotion("R"),
			effects.KeywordGrant(effects.OtherCreaturesYouControl, "haste"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}: You may put a red creature card or an artifact creature card from your hand onto the battlefield. Sacrifice it at the beginning of the next end step.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.ManaCost("{2}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.PutFromHandOntoBattlefield{
					Match:    effects.And(effects.Creature(), effects.Or(effects.OfColor("R"), effects.Artifact())),
					Optional: true,
					Label:    "Purphoros, Bronze-Blooded — you may put a red creature card or an artifact creature card from your hand onto the battlefield",
					ThenTail: effects.OldB137PurphorosEntered,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
