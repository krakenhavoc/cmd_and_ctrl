package effects

import (
	"errors"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Not Dead After All — Instant {B}:
//
//	"Until end of turn, target creature you control gains "When this
//	 creature dies, return it to the battlefield tapped under its
//	 owner's control, then create a Wicked Role token attached to it."
//	 (Enchanted creature gets +1/+1. When this token is put into a
//	 graveyard, each opponent loses 1 life.)"
//
// Fake Your Own Death's duration grant (ADR 0093 PR 4). The granted
// trigger is the creature's own, so "you" is its controller as it died.
// The Role goes on the creature that RETURNED — a new object with a new
// ID — and only if it did return: a token is gone (CR 111.7) and a card
// that left the graveyard first is not the object that died (CR 400.7),
// so neither gets a Role.
//
// No simplifications.
const notDeadAfterAllReturn = "not-dead-after-all/return"

func init() {
	Register(Spec{
		OracleID:     "380c367f-73ea-485f-a37b-5af0ba160893",
		Name:         "Not Dead After All",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		Grants: []AbilityGrant{{
			Key: notDeadAfterAllReturn,
			Triggered: []game.TriggeredAbility{
				WhenThisDies("Not Dead After All — return it tapped, then create a Wicked Role token attached to it",
					notDeadAfterAllReturnBody),
			},
			Text: "When this creature dies, return it to the battlefield tapped under its owner's control, then create a Wicked Role token attached to it.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			target, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return GrantAbilitiesFor{
				Target: target,
				Keys:   []string{notDeadAfterAllReturn},
				Label:  "Not Dead After All — it returns when it dies",
			}.Apply(ctx)
		},
	})
}

// notDeadAfterAllReturnBody returns the dead creature tapped, then
// attaches a Wicked Role to the returned object.
func notDeadAfterAllReturnBody(g *game.Game, item *game.StackItem) error {
	z := g.FindCardZoneForEffect(item.SourceCardID)
	if z == nil || z.Kind != game.ZoneGraveyard || !sameGraveyardObject(z, item) {
		return nil
	}
	back, err := g.ReturnFromGraveyardWithCountersForEffect(item.SourceCardID, uuid.Nil, true, nil)
	if err != nil {
		if errors.Is(err, game.ErrCardNotFound) {
			return nil
		}
		return err
	}
	return CreateRoleToken{Role: RoleWicked, Host: back}.Apply(NewContext(g, item))
}
