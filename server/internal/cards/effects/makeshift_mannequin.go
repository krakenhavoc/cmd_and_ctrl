package effects

import (
	"errors"

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
	Register(Spec{
		OracleID:     "d9dfdbbb-75b3-43d3-8ed6-371cb98a0124",
		Name:         "Makeshift Mannequin",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"A creature that asks a question as it enters, such as a Clone, comes back without the sacrifice ability."},
		Targets:      targetCreatureInYourGraveyard(),
		Grants: []AbilityGrant{{
			Key:  makeshiftMannequinGrant,
			Text: "When this creature becomes the target of a spell or ability, sacrifice it.",
			Triggered: []game.TriggeredAbility{
				On(game.EventBecomesTarget, Self, "Makeshift Mannequin — sacrifice it", SacrificeThisIfStillOnBattlefield),
			},
		}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			ts := ctx.LegalTargets()
			if len(ts) == 0 {
				return nil
			}
			id := ts[0].ID
			if _, err := ctx.Game.ReturnFromGraveyardWithCountersForEffect(id, ctx.Controller(), false,
				map[string]int{"mannequin": 1}); err != nil && !errors.Is(err, game.ErrCardNotFound) {
				return err
			}
			// A graveyard return keeps the card's ID (Gift of Immortality).
			if !onBattlefield(ctx.Game, id) {
				return nil
			}
			return grantWhileItHasCounter(ctx, id, "mannequin",
				"Makeshift Mannequin — sacrifice it when it becomes a target, while it has a mannequin counter", makeshiftMannequinGrant)
		},
	})
}
