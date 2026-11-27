package cardsv

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Volition Reins — Enchantment — Aura {3}{U}{U}{U} (Scars of Mirrodin, S82 OLD
// batch 472, #5577):
//
//	"Enchant permanent
//	 When this Aura enters, if enchanted permanent is tapped, untap it.
//	 You control enchanted permanent."
//
// Mind Control's layer 2 control change over any permanent, with a
// self-triggered untap whose "if" is both part of the trigger's condition and
// checked again as it resolves (CR 603.4). The permanent is read through the
// Aura it is attached to, so it is the one the Aura took.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bda496f9-5362-46f2-ad4f-cd96410c3c61",
		Name:         "Volition Reins",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPermanent(),
		Static:       []game.StaticAbility{effects.ControlAttachedBySource()},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, volitionReinsTapped),
				"Volition Reins — if enchanted permanent is tapped, untap it", volitionReinsUntap),
		},
	})
}

// volitionReinsEnchantedTapped is whether the permanent the Aura is attached to
// is tapped now.
func volitionReinsEnchantedTapped(g *game.Game, aura uuid.UUID) (uuid.UUID, bool) {
	a, ok := g.LookupCardForEffect(aura)
	if !ok || a.AttachedTo.Kind != game.TargetCard {
		return uuid.Nil, false
	}
	host, ok := g.LookupCardForEffect(a.AttachedTo.ID)
	return host.InstanceID, ok && host.Tapped
}

func volitionReinsTapped(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	_, ok := volitionReinsEnchantedTapped(g, source.InstanceID)
	return ok
}

func volitionReinsUntap(g *game.Game, item *game.StackItem) error {
	host, ok := volitionReinsEnchantedTapped(g, item.SourceCardID)
	if !ok {
		return nil
	}
	return effects.UntapTarget{Target: host}.Apply(effects.NewContext(g, item))
}
