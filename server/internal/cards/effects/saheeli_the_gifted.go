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
//	  - The −7 copies every artifact the controller has as the ability
//	    resolves, so the tokens are not copied in turn. The copies are
//	    made one at a time, each with haste (it rides the token's printed
//	    Keywords), and a single delayed trigger exiles all of them at the
//	    beginning of the next end step; one that has already left is
//	    skipped.
//
// Sandbox simplification, inherited from CreateTokenCopy: a token copy
// of a card whose entry lives in an on-enter hook rather than a trigger
// skips that effect.
func init() {
	Register(Spec{
		OracleID:     "0e084342-31ff-47a5-9fd8-5b385add7a1a",
		Name:         "Saheeli, the Gifted",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"A token copy of an artifact whose entry effect is an on-enter hook rather than a trigger skips that effect.",
		},
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
	ctx := NewContext(g, item)
	var artifacts []game.Card
	for i := range g.Battlefield.Cards {
		c := g.Battlefield.Cards[i]
		if c.Controller == item.Controller && c.IsArtifact() {
			artifacts = append(artifacts, c)
		}
	}
	var made []uuid.UUID
	for _, c := range artifacts {
		cursor := b25LastEventSeq(g)
		if err := (CreateTokenCopy{
			Controller: item.Controller,
			Copy:       c.InstanceID,
			N:          1,
			Except:     TokenCopyGainsHaste,
		}).Apply(ctx); err != nil {
			return err
		}
		made = append(made, b27TokensCreatedByAfter(g, item.Controller, cursor)...)
	}
	if len(made) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		At:         game.StepEnd,
		Controller: item.Controller,
		Label:      "Saheeli, the Gifted — exile the copies",
		Cards:      made,
		Body:       exileListedCardsBody,
	}.Apply(ctx)
}
