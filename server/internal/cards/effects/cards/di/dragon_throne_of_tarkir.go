package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dragon Throne of Tarkir — Legendary Artifact — Equipment {4} (Khans of
// Tarkir):
//
//	"Equipped creature has defender and "{2}, {T}: Other creatures you control
//	 gain trample and get +X/+X until end of turn, where X is this creature's
//	 power."
//	 Equip {3}"
//
// Viridian Longbow's grant of an activated ability (the equipped creature is
// the source, so its {T} keeps CR 302.6's summoning-sickness rule, and moving
// the Throne moves the ability) beside Bladed Pinions's keyword grant. X is the
// source creature's power as the ability resolves, from last-known information
// if it has gone (CR 608.2h); the other creatures you control are fixed as it
// resolves (CR 611.2c) and the creature itself gets neither.
//
// No simplification.
func init() {
	const grant = "dragon-throne-of-tarkir/rally"
	const text = "{2}, {T}: Other creatures you control gain trample and get +X/+X until end of turn, where X is this creature's power."
	effects.Register(effects.Spec{
		OracleID:     "93e40d56-18fd-42e1-a598-636aab79b836",
		Name:         "Dragon Throne of Tarkir",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:   text,
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					x := effects.SourcePowerNow(ctx)
					source := ctx.Source()
					others := effects.And(effects.Creature(), effects.YouControl(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.InstanceID != source })
					if err := (effects.GrantKeywordUntilEOT{Match: others, Keywords: []string{"trample"}, Label: "Dragon Throne of Tarkir — trample"}).Apply(ctx); err != nil {
						return err
					}
					return effects.BoostUntilEOT{Match: others, Power: x, Toughness: x, Label: "Dragon Throne of Tarkir — +X/+X"}.Apply(ctx)
				},
			}},
			Text: text,
		}},
		Static: []game.StaticAbility{
			effects.GrantToAttached("defender"),
			effects.GrantAbilitiesToAttached(grant),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
