package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Saheeli, the Gifted — Legendary Planeswalker — Saheeli {2}{U}{R},
// loyalty 4:
//
//		"+1: Create a 1/1 colorless Servo artifact creature token.
//		 +1: The next spell you cast this turn has affinity for artifacts.
//		     (It costs {1} less to cast for each artifact you control as you
//		     cast it.)
//		 −7: For each artifact you control, create a token that's a copy of
//		     it. Those tokens gain haste. Exile those tokens at the beginning
//		     of the next end step.
//		 Saheeli, the Gifted can be your commander."
//
//	  - The second +1 is #1852's one-use promise with an affinity rider:
//	    it prices the next spell cast, of any kind, at {1} less per
//	    artifact its caster controls when it is cast (CR 601.2f), and it
//	    is spent by that spell. A second activation the same turn is a
//	    second promise, and both apply to the one spell.
//	  - The −7 snapshots the artifacts before creating their copies. The
//	    creation continuation grants haste and schedules exile after all
//	    entry choices finish; haste is a continuous grant, not a copiable
//	    exception to the tokens' printed characteristics.
func init() {
	Register(Spec{
		OracleID:        "0e084342-31ff-47a5-9fd8-5b385add7a1a",
		Name:            "Saheeli, the Gifted",
		Completeness:    CompletenessFull,
		StartingLoyalty: 4,
		Activated: []ActivatedAbility{
			{
				Label:  "+1: Create a 1/1 colorless Servo artifact creature token.",
				Cost:   LoyaltyCost(1),
				Effect: saheeliGiftedServo,
			},
			{
				Label:  "+1: The next spell you cast this turn has affinity for artifacts.",
				Cost:   LoyaltyCost(1),
				Effect: saheeliGiftedAffinity,
			},
			{
				Label:  "−7: For each artifact you control, create a token that's a copy of it. Those tokens gain haste. Exile those tokens at the beginning of the next end step.",
				Cost:   LoyaltyCost(-7),
				Effect: saheeliGiftedCopyArtifacts,
			},
		},
	})
}

func saheeliGiftedServo(g *game.Game, item *game.StackItem) error {
	return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 colorless Servo artifact"), N: 1}.Apply(NewContext(g, item))
}

func saheeliGiftedAffinity(g *game.Game, item *game.StackItem) error {
	return GrantNextSpellPromise{From: "Saheeli, the Gifted", Promise: game.NextSpellPromise{
		AffinityFor: "Artifact",
		Text:        "The next spell you cast this turn has affinity for artifacts.",
	}}.Apply(NewContext(g, item))
}

func saheeliGiftedCopyArtifacts(g *game.Game, item *game.StackItem) error {
	var groups []game.TokenGroup
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsArtifact() {
			if tmpl, ok := TokenCopyTemplate(g, c.InstanceID); ok {
				groups = append(groups, game.TokenGroup{Template: tmpl, Count: 1})
			}
		}
	}
	return g.CreateTokensThenForEffect(game.TokenCreation{
		Controller: item.Controller, Source: item.SourceCardID, Groups: groups, Together: true,
	}, func(g *game.Game, made []uuid.UUID) error {
		ctx := NewContext(g, item)
		for _, id := range made {
			if err := (ScopedEffectFor{Target: id, Mods: []game.Mod{game.AddKeywordsMod("haste")},
				Duration: game.IndefiniteDuration(), Label: "Saheeli, the Gifted — haste"}).Apply(ctx); err != nil {
				return err
			}
		}
		if len(made) == 0 {
			return nil
		}
		return (ScheduleDelayedTrigger{At: game.StepEnd, Controller: item.Controller,
			Label: "Saheeli, the Gifted — exile the copies", Cards: made, Body: exileListedCardsBody}).Apply(ctx)
	})
}
