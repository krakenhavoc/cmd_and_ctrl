package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const tocasiaGrant = "tocasia-dig-site-mentor/surveil"

// Tocasia, Dig Site Mentor — Legendary Creature — Human Artificer
// {1}{G}{W}{U}, 4/3 (The Brothers' War, #5482):
//
//	"Creatures you control have vigilance and "{T}: Surveil 1."
//	 {2}{G}{G}{W}{W}{U}{U}, Exile this card from your graveyard: Return any
//	 number of target artifact cards with total mana value 10 or less from
//	 your graveyard to the battlefield. Activate only as a sorcery."
//
// The grant is Cryptolith Rite's shape: a layer-6 bundle of the activated
// ability, plus a vigilance grant. The graveyard ability's targets carry a
// total-mana-value ceiling, judged at announce and again at resolution, so
// a set that no longer fits returns nothing (CR 608.2b). The artifacts
// enter together.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e790f7e6-45f5-45df-b8e5-bcbda4f69f8a",
		Name:         "Tocasia, Dig Site Mentor",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:  tocasiaGrant,
			Text: "{T}: Surveil 1.",
			Activated: []ActivatedAbility{{
				Label:   "{T}: Surveil 1.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    TapCost(),
				Effect:  Do(Surveil{N: 1}),
			}},
		}},
		Static: []game.StaticAbility{
			Labelled(b16GrantKeywords(b16CreaturesYouControl, "vigilance"), "Creatures you control have vigilance."),
			Labelled(GrantAbilities(b16CreaturesYouControl, tocasiaGrant), "Creatures you control have \"{T}: Surveil 1.\""),
		},
		Activated: []ActivatedAbility{{
			Label: "{2}{G}{G}{W}{W}{U}{U}, Exile this card from your graveyard: Return any number of target artifact cards with total mana value 10 or less from your graveyard to the battlefield. Activate only as a sorcery.",
			Cost:  Plus(ManaCost("{2}{G}{G}{W}{W}{U}{U}"), ExileThis()),
			Zones: []game.ZoneKind{game.ZoneGraveyard},
			Targets: TargetCardInGraveyard("any number of target artifact cards with total mana value 10 or less from your graveyard", YouOwn(), Artifact()).
				WithCount(0, 0).WithTotal(TotalManaValueAtMost(10)),
			SorcerySpeed: true,
			Purpose:      ForTargets(Reanimates(0)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				ids := legalTargetCardIDs(ctx)
				if len(ids) == 0 {
					return nil
				}
				return ReturnFromGraveyardTogether{Targets: ids, Controller: ctx.Controller()}.Apply(ctx)
			},
		}},
	})
}
