package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Makeshift Mannequin — Instant for {3}{B}:
//
//	"Return target creature card from your graveyard to the battlefield with a mannequin counter on it. For as long as that creature has a mannequin counter on it, it has "When this creature becomes the target of a spell or ability, sacrifice it.""
//
// ADR 0109 §2 (#1604): the creature returns with its mannequin counter
// (a CR 614.1c "enters with", so it is on the creature as it arrives),
// then has the sacrifice trigger, a granted ability bundle (ADR 0093
// PR 4), for as long as it has a mannequin counter on it.
const makeshiftMannequinGrant = "makeshift-mannequin/sacrifice"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "d9dfdbbb-75b3-43d3-8ed6-371cb98a0124",
		Name:         "Makeshift Mannequin",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreatureInYourGraveyard(),
		Grants: []effects.AbilityGrant{{
			Key:  makeshiftMannequinGrant,
			Text: "When this creature becomes the target of a spell or ability, sacrifice it.",
			Triggered: []game.TriggeredAbility{
				effects.On(game.EventBecomesTarget, effects.Self, "Makeshift Mannequin — sacrifice it", effects.SacrificeThisIfStillOnBattlefield),
			},
		}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			ts := ctx.LegalTargets()
			if len(ts) == 0 {
				return nil
			}
			// The grant is the entry's tail (ADR 0154): a creature that
			// asks a question as it enters (a Clone's copy pick) still gets
			// it once the pick is answered. Plain values only; the tail
			// runs from a resume as well as inline.
			return ctx.Game.ReturnFromGraveyardWithCountersKeyedForEffect(ts[0].ID, item.Controller, false,
				map[string]int{"mannequin": 1}, ctx.Tail(makeshiftMannequinGrantTail))
		},
	})
}

// makeshiftMannequinGrantTail is the entry's tail as a keyed tail (#4253
// slice 6): the sacrifice grant while it has a mannequin counter. An
// on-disk key.
var makeshiftMannequinGrantTail = effects.EnteredTail("makeshift-mannequin/grant-while-counter", func(ctx *effects.Context, entered uuid.UUID, _ game.TailOutcome) error {
	// A graveyard return keeps the card's ID (Gift of Immortality).
	if entered == uuid.Nil || !effects.OnBattlefield(ctx.Game, entered) {
		return nil
	}
	return effects.GrantWhileItAlreadyHasCounter(ctx, entered, "mannequin",
		"Makeshift Mannequin — sacrifice it when it becomes a target, while it has a mannequin counter", makeshiftMannequinGrant)
})
