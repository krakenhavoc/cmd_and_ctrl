package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brenard, Ginger Sculptor — Legendary Creature — Human Artificer
// {1}{G}{W}{U}, 3/3:
//
//	"Each creature you control that's a Food or a Golem gets +2/+2 and
//	 has trample.
//	 Whenever another nontoken creature you control dies, you may exile
//	 it. If you do, create a token that's a copy of that creature,
//	 except it's a 1/1 Food Golem artifact creature in addition to its
//	 other types and it has "{2}, {T}, Sacrifice this token: You gain 3
//	 life.""
//
// The statics are a tribal anthem and keyword grant over the creatures
// Brenard's controller controls that are Food or Golem (TribeFilter reads
// the layered view, so the copies' own types count). The trigger is
// Nightmare Shepherd's exile-then-copy; the copy's exceptions (1/1, the
// Artifact type and the Food and Golem subtypes, and the quoted Food
// ability) are part of the copiable values (CR 707.9a), the ability as a
// granted bundle that works the way a Food's does (cost {2}, {T},
// sacrifice itself).
//
// No simplification.
const brenardFoodGrant = "brenard-ginger-sculptor/food-ability"

func init() {
	foodOrGolem := TribeFilter{Tribes: []string{"Food", "Golem"}, YoursOnly: true}
	Register(Spec{
		OracleID:     "1a2ab12a-5778-40df-a2ec-2733a7121cb4",
		Name:         "Brenard, Ginger Sculptor",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(foodOrGolem, 2, 2),
			TribalKeywordGrant(foodOrGolem, "trample"),
		},
		Grants: []AbilityGrant{{
			Key:  brenardFoodGrant,
			Text: "{2}, {T}, Sacrifice this token: You gain 3 life.",
			Activated: []ActivatedAbility{{
				Label:   "{2}, {T}, Sacrifice this token: You gain 3 life.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{2}"), TapCost(), SacrificeThis()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return GainLife{Player: item.Controller, Amount: 3}.Apply(NewContext(g, item))
				},
			}},
		}},
		Triggered: []game.TriggeredAbility{
			cmdr2S47ExileThenCopyTrigger(
				"Brenard, Ginger Sculptor — exile it; if you do, create a 1/1 Food Golem token copy",
				"Brenard, Ginger Sculptor — exile that creature and create a 1/1 Food Golem token copy of it?",
				"brenard-ginger-sculptor/copy",
				func(t *game.Card) {
					v := game.PrintedValues{TypeLine: t.TypeLine, GrantedAbilities: t.GrantedAbilities}
					v.AddCardType("Artifact")
					v.AddSubtype("Food")
					v.AddSubtype("Golem")
					v.GrantAbility(brenardFoodGrant)
					t.TypeLine = v.TypeLine
					t.GrantedAbilities = v.GrantedAbilities
					t.Power, t.Toughness, t.VariableToughness = 1, 1, false
				}),
		},
	})
}
