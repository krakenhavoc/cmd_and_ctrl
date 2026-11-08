package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kitnap — Enchantment — Aura {2}{U}{U} (Bloomburrow, #4467):
//
//	"Gift a card (You may promise an opponent a gift as you cast this
//	 spell. If you do, when it enters, they draw a card.)
//	 Enchant creature
//	 When this Aura enters, tap enchanted creature. If the gift wasn't
//	 promised, put three stun counters on it.
//	 You control enchanted creature."
//
// Control Magic with a gift. The gift is the shared keyword (gift.go),
// which grows its own entry trigger. The control change is Control
// Magic's layer 2 static (ControlAttachedBySource).
//
// The entry trigger is Claustrophobia's tap (tapEnchantedCreatureOnEntry
// reads the creature the Aura enchants, or last enchanted if it has left
// in response, CR 608.2h) followed by the stun counters. Whether the gift
// was promised is the Aura's cast record (CR 400.7d), read as the trigger
// goes on the stack and carried on the item as the counter count, since
// the Aura may be gone by resolution. A Kitnap put onto the battlefield
// without being cast was never promised a gift, so it stuns.
//
// No simplification.
func init() {
	const label = "Kitnap — tap enchanted creature; if the gift wasn't promised, put three stun counters on it"
	enters := WhenThisEnters(label, kitnapTapAndStun)
	enters.Build = func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		if !source.GiftPromised() {
			item.Params.Amount = 3
		}
		return item
	}
	Register(Spec{
		OracleID:     "a1bd1e83-40ac-449f-838c-ef308f275a5d",
		Name:         "Kitnap",
		Completeness: CompletenessFull,
		Gift:         GiftACard(),
		Targets:      EnchantCreature(),
		Static:       []game.StaticAbility{ControlAttachedBySource()},
		Triggered:    []game.TriggeredAbility{enters},
	})
}

// kitnapTapAndStun taps the enchanted creature, then puts the stun
// counters the trigger carries (three, or none if the gift was promised)
// on that same creature.
func kitnapTapAndStun(g *game.Game, item *game.StackItem) error {
	if err := tapEnchantedCreatureOnEntry(g, item); err != nil {
		return err
	}
	if item.Params.Amount <= 0 {
		return nil
	}
	ctx := NewContext(g, item)
	aura, ok := ctx.TriggeringPermanent()
	if !ok || aura.AttachedTo.Kind != game.TargetCard {
		return nil
	}
	return AddCounter{Target: aura.AttachedTo.ID, Kind: game.CounterStun, N: item.Params.Amount}.Apply(ctx)
}
