package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// balambGardenOracle is shared by both faces: the front registers under
// it, the back under balambGardenOracle + "#1".
const balambGardenOracle = "8b84fec5-617c-4088-8250-2ba1f1f9479a"

// Balamb Garden, SeeD Academy // Balamb Garden, Airborne — a
// transforming land (Final Fantasy, #4462):
//
//	Balamb Garden, SeeD Academy, Land — Town
//	  "This land enters tapped.
//	   {T}: Add {G} or {U}.
//	   {5}{G}{U}, {T}: Transform this land. This ability costs {1} less
//	   to activate for each other Town you control."
//
//	Balamb Garden, Airborne, Legendary Artifact — Vehicle, 5/4
//	  "Flying
//	   Whenever Balamb Garden attacks, draw a card.
//	   Crew 1"
//
// Westvale Abbey's shape: the permanent transforms in place
// (TransformThis), so it is the same object, nothing re-enters and it
// stays tapped from the {T} it paid. The discount is the ability's own
// cost clause (ActivatedAbility.CostModifiers, The Lonely Mountain's
// shape) counting the OTHER Towns its controller controls, by object,
// after the layers; it comes off the generic part only, so {G}{U} is
// always paid. The back face is a noncreature Vehicle until it is
// crewed; its attack trigger and Crew 1 are the ordinary ones.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      balambGardenOracle,
		Name:          "Balamb Garden, SeeD Academy",
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility("G", "U")},
		Activated: []ActivatedAbility{{
			Label:   "{5}{G}{U}, {T}: Transform this land. This ability costs {1} less to activate for each other Town you control.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{5}{G}{U}"), TapCost()),
			CostModifiers: []game.CostModifier{
				CostsLessEach(balambGardenOtherTowns,
					"This ability costs {1} less to activate for each other Town you control."),
			},
			Effect: Do(TransformThis{}),
		}},
	})
	Register(Spec{
		OracleID:        balambGardenOracle + "#1",
		Name:            "Balamb Garden, Airborne",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Balamb Garden, Airborne — draw a card", Do(DrawCards{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Balamb Garden, Airborne"),
		}},
	})
}

// balambGardenOtherTowns counts the Towns the activator controls other
// than the land whose ability is being priced.
func balambGardenOtherTowns(q game.CostQuery) int {
	town := MatchLandSubtype("Town")
	return countBattlefield(q, func(c game.Card) bool {
		return c.Controller == q.Controller && c.InstanceID != q.Source.InstanceID && town(c)
	})
}
