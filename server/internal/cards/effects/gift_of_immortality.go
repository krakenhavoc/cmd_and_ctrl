package effects

import (
	"errors"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gift of Immortality — Enchantment — Aura {2}{W}:
//
//	"Enchant creature
//	 When enchanted creature dies, return that card to the battlefield
//	 under its owner's control. Return this card to the battlefield
//	 attached to that creature at the beginning of the next end step."
//
// Two instructions and a delay between them. The first is the Aura's
// dies trigger: the creature card comes back under its OWNER's control
// (not the Aura controller's), as a new object. The second is a
// delayed trigger (CR 603.7) scheduled as that returns, so the table
// can respond to it, and it names the creature by OBJECT: if the
// creature has left the battlefield again by the next end step — or
// come back as a different object — the Aura stays in the graveyard,
// because there is nothing for it to be attached to.
//
// The Aura returns under its owner's control and is attached by the
// same resolution that returns it (CR 303.4f: an Aura put onto the
// battlefield that way is attached to the object the effect names), so
// the CR 704.5m check never sees it unattached.
//
// A creature that cannot return (a token, a card already gone, a
// commander that went to the command zone) schedules nothing: the
// Aura has no creature to come back to.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2c11c72f-5b04-47d8-bdf8-8d9d197ebc1b",
		Name:         "Gift of Immortality",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Triggered: []game.TriggeredAbility{
			WhenEnchantedCreatureDies("Gift of Immortality — return that card to the battlefield", giftOfImmortalityReturnCreature),
		},
	})
}

// giftOfImmortalityReturnCreature is the dies trigger's effect: return
// the creature card under its owner's control and schedule the Aura's
// own return.
func giftOfImmortalityReturnCreature(g *game.Game, item *game.StackItem) error {
	dead := item.Trigger.Event.CardID
	back, err := g.ReturnToBattlefieldForEffect(dead, uuid.Nil, false)
	if errors.Is(err, game.ErrCardNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if back == uuid.Nil {
		// The entry paused or was redirected; a graveyard return keeps
		// the card's ID, so the creature is named by it if it landed.
		back = dead
	}
	c, ok := g.LookupCardForEffect(back)
	if !ok || !onBattlefield(g, back) {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label:  "Gift of Immortality — return this card to the battlefield attached to that creature",
		Cards:  []uuid.UUID{item.SourceCardID},
		Body:   giftOfImmortalityReturnAuraBody,
		Params: game.EffectParams{Object: game.ObjectRef{ID: back, Epoch: c.ObjectEpoch}},
	}.Apply(NewContext(g, item))
}

// giftOfImmortalityReturnAura is the delayed trigger's body
// (giftOfImmortalityReturnAuraBody, delayed_bodies.go): return the
// Aura, which item.Targets names, and attach it to the creature p.Object
// names — while that is still the same object on the battlefield.
func giftOfImmortalityReturnAura(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	if len(item.Targets) == 0 {
		return nil
	}
	host, ok := g.LookupCardForEffect(p.Object.ID)
	if !ok || host.ObjectEpoch != p.Object.Epoch || !onBattlefield(g, p.Object.ID) {
		return nil
	}
	aura := item.Targets[0].ID
	back, err := g.ReturnToBattlefieldForEffect(aura, uuid.Nil, false)
	if errors.Is(err, game.ErrCardNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if back == uuid.Nil {
		back = aura
	}
	return g.AttachForEffect(back, game.TargetRef{Kind: game.TargetCard, ID: p.Object.ID})
}
