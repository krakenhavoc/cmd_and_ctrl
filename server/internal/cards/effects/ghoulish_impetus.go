package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ghoulish Impetus — Enchantment — Aura {2}{B} (Secrets of Strixhaven
// Commander, #4834):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1, has deathtouch, and is goaded. (It
//	 attacks each combat if able and attacks a player other than you if
//	 able.)
//	 When enchanted creature dies, return this card to the battlefield at
//	 the beginning of the next end step."
//
// Coercive Impetus's statics (a pump and the goad pair that ends when the
// Aura falls off) with a deathtouch grant beside them. The dies trigger is
// harvested while the Aura is still attached, and resolves with the Aura
// in its owner's graveyard, where the state-based action put it
// (CR 704.5m). It schedules Avacyn, Angel of Horror's delayed return
// (CR 603.7) naming the Aura as the graveyard object it is now, so an
// Aura that leaves the graveyard before the end step, even to come back,
// is a new object (CR 400.7) and stays where it went. An Aura already
// gone from the graveyard as the trigger resolves schedules nothing.
//
// The Aura returns under the control of the trigger's controller
// (CR 110.2a), who chooses what it enchants as it enters (CR 303.4f);
// with nothing it could enchant it stays in the graveyard (CR 303.4g).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "276eccd6-4a76-46e8-b604-94472e618115",
		Name:         "Ghoulish Impetus",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("deathtouch"),
			GoadAttached(),
		},
		Triggered: []game.TriggeredAbility{
			WhenEnchantedCreatureDies("Ghoulish Impetus — return this card to the battlefield at the beginning of the next end step",
				ghoulishImpetusScheduleReturn),
		},
	})
}

// ghoulishImpetusScheduleReturn schedules the end-step return of the Aura,
// if it is in the graveyard as the card that left the battlefield.
func ghoulishImpetusScheduleReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if !sourceCardInGraveyardAfterLeaving(ctx) {
		return nil
	}
	id := item.SourceCardID
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label:  "Ghoulish Impetus — return this card to the battlefield",
		Cards:  []uuid.UUID{id},
		Body:   returnTheObjectFromGraveyardUnderYourControlBody,
		Params: game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
	}.Apply(ctx)
}
