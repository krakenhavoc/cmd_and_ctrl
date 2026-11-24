package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghost Ark — Artifact — Vehicle {4}, 3/3 (Doctor Who):
//
//	"Flying
//	 Repair Barge — Whenever this Vehicle becomes crewed, each artifact
//	 creature card in your graveyard gains unearth {3} until end of turn.
//	 Crew 2"
//
// The grant is Cursecloth Wrappings' (#4464): an ordinary graveyard
// AbilityGrant whose one row is the shared unearth body (Unearth), given to
// each card by GrantAbilityToGraveyardCardUntilEndOfTurnForEffect. The set
// is locked as the trigger resolves (CR 611.2c): a card that reaches the
// graveyard later gains nothing. The cleanup sweep ends the grant (CR
// 514.2), and a card that leaves the graveyard loses it (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bb1b4786-3804-486b-9fd9-85848d24d362",
		Name:            "Ghost Ark",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Grants: []effects.AbilityGrant{{
			Key:       ghostArkUnearth,
			Zone:      game.ZoneGraveyard,
			Activated: []effects.ActivatedAbility{effects.Unearth("{3}")},
			Text:      "Unearth {3}",
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenBecomesCrewed("Ghost Ark — Repair Barge: each artifact creature card in your graveyard gains unearth {3} until end of turn",
				ghostArkRepairBarge),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    effects.CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Ghost Ark"),
		}},
	})
}

// ghostArkUnearth is the bundle key of the unearth the Ark gives. On disk.
const ghostArkUnearth = "ghost-ark/unearth"

func ghostArkRepairBarge(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	p := ctx.PlayerByID(ctx.Controller())
	if p == nil || p.Graveyard == nil {
		return nil
	}
	var ids []game.Card
	for _, c := range p.Graveyard.Cards {
		if c.IsArtifact() && c.IsCreature() {
			ids = append(ids, c)
		}
	}
	for _, c := range ids {
		g.GrantAbilityToGraveyardCardUntilEndOfTurnForEffect(c.InstanceID, ctx.Source(), ctx.Controller(), ghostArkUnearth)
	}
	return nil
}
