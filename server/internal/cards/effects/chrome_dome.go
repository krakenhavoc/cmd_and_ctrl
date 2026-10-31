package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chrome Dome — {2} Artifact Creature — Robot Ninja 1/3 (Teenage Mutant
// Ninja Turtles, #4457):
//
//	"Other artifact creatures you control get +1/+0.
//	 {5}: Create a token that's a copy of another target artifact you
//	 control. That token gains haste. Sacrifice it at the beginning of
//	 the next end step."
//
// The lord is a layer 7c pump (CR 613.4c) over the other artifact creatures
// its controller controls. The activation is Saheeli, the Gifted's −7 for one
// target: the copy takes the target's copiable values (CR 707.2), and haste
// is NOT one of them: it is a continuous grant to that token with no stated
// duration (CR 611.2a), so a copy of the token made later does not have it.
// The sacrifice is a delayed trigger at the beginning of the next end step
// (CR 603.7). The token is created through the keyed creation door, so a
// copy that asks an entry question (a Clone-like artifact) is a restore
// point while it waits, and the haste and sacrifice run after it settles.
// The target is re-read as the ability resolves (CR 608.2b): an artifact
// that left in response makes nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3dc71e16-b935-4ad7-b3f7-945ac2c179a7",
		Name:         "Chrome Dome",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			Label:    "Other artifact creatures you control get +1/+0.",
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID != source.InstanceID && target.Controller == source.Controller &&
					target.IsArtifact() && target.IsCreature()
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{5}: Create a token that's a copy of another target artifact you control. That token gains haste. Sacrifice it at the beginning of the next end step.",
			Cost:    ManaCost("{5}"),
			Targets: Another(TargetPermanent("another target artifact you control", Artifact(), YouControl())),
			Effect:  chromeDomeCopy,
		}},
	})
}

// chromeDomeCopy makes the token copy of the still-legal target.
func chromeDomeCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	copyOf := FirstLegalBattlefieldTarget(ctx)
	if copyOf == uuid.Nil {
		return nil
	}
	tmpl, ok := TokenCopyTemplate(g, copyOf)
	if !ok {
		return nil
	}
	return g.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: item.Controller,
		Source:     item.SourceCardID,
		Groups:     []game.TokenGroup{{Template: copiedToken(tmpl), Count: 1}},
	}, ctx.ItemTail(chromeDomeCopyRest))
}

// chromeDomeCopyRest is "That token gains haste. Sacrifice it at the
// beginning of the next end step." as a keyed tail, told the tokens the
// creation actually made. An on-disk key: never renamed, never reused.
var chromeDomeCopyRest = CardTail("chrome-dome/haste-and-sacrifice-copy", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if err := (ScopedEffectFor{Target: id, Mods: []game.Mod{game.AddKeywordsMod("haste")},
			Duration: game.IndefiniteDuration(), Label: "Chrome Dome — the token gains haste"}).Apply(ctx); err != nil {
			return err
		}
	}
	if len(o.IDs) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		At:         game.StepEnd,
		Controller: ctx.Item.Controller,
		Label:      "Chrome Dome — sacrifice the token",
		Cards:      o.IDs,
		Body:       sacrificeListedCardsBody,
	}.Apply(ctx)
})
