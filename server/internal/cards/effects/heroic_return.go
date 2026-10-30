package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Heroic Return — Instant {5}{W} (Marvel Super Heroes Commander, #4455):
//
//	"This spell costs {2} less to cast if a creature is attacking you.
//	 Return target creature card from your graveyard to the battlefield. If
//	 a Hero enters this way, it enters with two additional +1/+1 counters
//	 on it."
//
// The discount (CR 601.2f) looks for a creature attacking the caster
// themselves: one attacking a planeswalker or battle the caster controls is
// not attacking them. The counters are the entry's own (EntersWithCounters,
// CR 614.1c, so Hardened Scales and Doubling Season see them), given when the
// card is a Hero as it would exist on the battlefield (CR 614.12), not as it
// sits in the graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4b7a8c05-15cd-452e-b9fc-83800e81793d",
		Name:         "Heroic Return",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			CostsLess(2, "This spell costs {2} less to cast if a creature is attacking you.", aCreatureIsAttackingYou()),
		},
		Targets:   targetCreatureInYourGraveyard(),
		OnResolve: heroicReturnResolve,
	})
}

// aCreatureIsAttackingYou passes when a creature is attacking the caster.
func aCreatureIsAttackingYou() CostPredicate {
	return func(q game.CostQuery) bool {
		if q.Game == nil || q.Game.Battlefield == nil {
			return false
		}
		for _, c := range q.Game.Battlefield.Cards {
			if c.IsCreature() && c.AttackingTarget == q.Controller {
				return true
			}
		}
		return false
	}
}

func heroicReturnResolve(_ *game.StackItem, ctx *Context) error {
	var id uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			id = t.ID
			break
		}
	}
	if id == uuid.Nil {
		return nil
	}
	if _, ok := ctx.Game.LookupCardForEffect(id); !ok {
		return nil
	}
	// "If a Hero enters this way": judged by what the card would be on the
	// battlefield (CR 614.12), so a static that makes it a Hero there counts
	// and one that stops it being a Hero does not, whatever the card is in
	// the graveyard.
	opts := game.ZoneEntryOptions{
		Controller: ctx.Controller(),
		EntersWithCountersIfType: []game.EntryCountersIfType{
			{Subtype: ctx.TextSubtype("Hero"), Counter: game.CounterPlusOne, N: 2},
		},
	}
	return putCardsFromZoneTogetherWith(ctx, game.ZoneGraveyard, []uuid.UUID{id}, opts, game.TailRef{}, 0)
}
