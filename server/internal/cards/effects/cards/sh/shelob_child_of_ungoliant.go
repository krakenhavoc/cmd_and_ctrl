package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shelob, Child of Ungoliant — Legendary Creature — Spider Demon {4}{B}{G},
// 8/8 (The Lord of the Rings: Tales of Middle-earth, S82 GAP2 b11, #6098):
//
//	"Deathtouch, ward {2}
//	 Other Spiders you control have deathtouch and ward {2}.
//	 Whenever another creature dealt damage this turn by a Spider you
//	 controlled dies, create a token that's a copy of that creature,
//	 except it's a Food artifact with "{2}, {T}, Sacrifice this token: You
//	 gain 3 life," and it loses all other card types."
//
// The grants are Rith's (ward, a triggered ability granted to each other
// Spider you control) and a tribal keyword grant (deathtouch).
//
// "Dealt damage this turn by a Spider you controlled" is the turn's damage
// record (TurnTally.DamagedObjects): each event names who controlled the
// source and the source's subtypes AS IT DEALT the damage (Aegar's reader,
// PermanentDealtDamageThisTurnByKindForEffect), so a Spider that has since
// died, changed hands or stopped being a Spider still counts, and one that
// became a Spider later does not. The victim is the object that died,
// read as it last existed; a creature that left and came back is a new
// object (CR 400.7).
//
// The token copies the creature as it last existed on the battlefield (CR
// 707.2, the trigger's object), then the exception (CR 707.9b): its card
// types become Artifact alone, which takes its creature types with the
// creature type (CR 205.1a), it is a Food, and it has the quoted ability
// as a granted bundle. Its other abilities are copied and kept.
//
// No simplification.
const shelobFoodGrant = "shelob-child-of-ungoliant/food-ability"

func init() {
	otherSpiders := effects.TribeFilter{Tribes: []string{"Spider"}, Others: true, YoursOnly: true}
	effects.Register(effects.Spec{
		OracleID:        "8372115d-c963-4a41-8759-72f7f6a8935a",
		Name:            "Shelob, Child of Ungoliant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch", "ward"},
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(otherSpiders, "deathtouch"),
		},
		Grants: []effects.AbilityGrant{{
			Key:  shelobFoodGrant,
			Text: "{2}, {T}, Sacrifice this token: You gain 3 life.",
			Activated: []effects.ActivatedAbility{{
				Label:   "{2}, {T}, Sacrifice this token: You gain 3 life.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeThis()),
				Effect:  effects.B36GainLife(3),
			}},
		}},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{2}"), "Shelob, Child of Ungoliant — ward {2}"),
			effects.WardGranted(effects.WardMana("{2}"), "Shelob, Child of Ungoliant — other Spiders you control have ward {2}",
				func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.InstanceID != source.InstanceID && target.Controller == source.Controller &&
						target.IsCreature() && target.HasSubtype(source.TextSubtype("Spider"))
				}),
			effects.On(game.EventLTB, shelobSpiderDamagedCreatureDied,
				"Shelob, Child of Ungoliant — create a token copy of that creature, except it's a Food artifact",
				shelobCopyAsFood),
		},
	})
}

// shelobSpiderDamagedCreatureDied is "another creature dealt damage this
// turn by a Spider you controlled dies".
func shelobSpiderDamagedCreatureDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard || ev.CardID == source.InstanceID {
		return false
	}
	if was, known := ev.WasType("creature"); known && !was {
		return false
	}
	return g.PermanentDealtDamageThisTurnByKindForEffect(ev.CardID, source.Controller, false, source.TextSubtype("Spider"))
}

func shelobCopyAsFood(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	obj := ctx.Trigger().Object
	if obj == nil {
		return nil
	}
	tmpl, ok := g.TokenCopyTemplateForObjectForEffect(obj.Ref())
	if !ok {
		return nil
	}
	v := game.PrintedValues{TypeLine: tmpl.TypeLine, GrantedAbilities: tmpl.GrantedAbilities}
	v.SetCardTypes("Artifact")
	v.AddSubtype("Food")
	v.GrantAbility(shelobFoodGrant)
	tmpl.TypeLine = v.TypeLine
	tmpl.GrantedAbilities = v.GrantedAbilities
	return effects.CreateTokenCopies(ctx, ctx.Controller(), tmpl, 1)
}
