package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tamiyo, Upriser Crowned — Legendary Creature — Moonfolk Warrior
// {4}{R}{W}, 3/5:
//
//	"Flying, double strike, haste
//	 When Tamiyo enters, you become the monarch.
//	 Whenever one or more creatures deal combat damage to you while
//	 you're the monarch, tap those creatures and put a stun counter on
//	 each of them."
//
// The engine emits one damage event per creature, so the trigger fires
// once per creature rather than once per batch; each one taps and
// stuns exactly its own creature, which leaves the board the same as
// one batched trigger would (the creatures are distinct, so there is no
// double stun). "While you're the monarch" is judged as the damage is
// dealt, the trigger-time reading.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8e0f2e26-1e14-422a-aafe-e1da38ab10a3",
		Name:            "Tamiyo, Upriser Crowned",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "double strike", "haste"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouBecomeTheMonarch("Tamiyo, Upriser Crowned"),
			On(game.EventDealDamage, rfCreatureECombatDamageToYouWhileMonarch,
				"Tamiyo, Upriser Crowned — tap the creature and put a stun counter on it", rfCreatureETapAndStunTheDamager),
		},
	})
}

// rfCreatureECombatDamageToYouWhileMonarch is Tamiyo's condition: a
// creature dealt combat damage to the source's controller, who is the
// monarch.
func rfCreatureECombatDamageToYouWhileMonarch(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || ev.Target != source.Controller {
		return false
	}
	if p := g.PlayerByIDForEffect(ev.Target); p == nil {
		return false
	}
	src, ok := g.LookupCardForEffect(ev.Source)
	return ok && src.IsCreature() && YoureTheMonarch(g, source.Controller)
}

// rfCreatureETapAndStunTheDamager taps the creature that dealt the
// damage and puts a stun counter on it, if it is still on the
// battlefield.
func rfCreatureETapAndStunTheDamager(g *game.Game, item *game.StackItem) error {
	id := item.Trigger.Event.Source
	if id == uuid.Nil {
		return nil
	}
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	ctx := NewContext(g, item)
	if err := (TapTarget{Target: id}).Apply(ctx); err != nil {
		return err
	}
	return AddCounter{Target: id, Kind: game.CounterStun, N: 1}.Apply(ctx)
}
