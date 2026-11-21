package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rogue Class — {U}{B} Enchantment — Class (Adventures in the Forgotten Realms):
//
//	"(Gain the next level as a sorcery to add its ability.)
//	 Whenever a creature you control deals combat damage to a player, exile the
//	 top card of that player's library face down. You may look at it for as long
//	 as it remains exiled.
//	 {1}{U}{B}: Level 2
//	 Creatures you control have menace.
//	 {2}{U}{B}: Level 3
//	 You may play cards exiled with this Class, and you may spend mana as
//	 though it were mana of any color to cast those spells."
//
// Level 1 is the Draw Arcanum exile (Urianger Augurelt, ADR 0091's
// 2026-10-15 amendment): the card lands face down, linked to THIS Class object
// (CR 607.2a, 400.7) and looked at by that Class's controller alone, and stays
// readable to them while it is exiled (CR 406.3). The exile is the Class's
// trigger, so the damaged player's library is the one that loses its top card,
// and the creature's controller is the Class's controller. A library with no
// card does nothing.
//
// Level 2 is Ninja Teen's menace grant, gated at level 2. Level 3 is a standing
// permission over exile gated at level 3 (CR 716.2a): it reaches the cards this
// Class exiled (PermissionFilter.ExiledWithSource), "play", so a land spends the
// land drop, and spends mana as though it were mana of any color for a spell
// (AnyColor, never any TYPE). At levels 1 and 2 the permission does not exist,
// so the exiled cards are only looked at.
//
// No simplification.
func init() {
	menace := effects.TribalKeywordGrant(effects.TribeFilter{YoursOnly: true}, "menace")
	menace.ActiveWhen = effects.Level(2)
	effects.Register(effects.Spec{
		OracleID:     "76f0f39a-ce1b-4935-a6b8-5c4e1a27ab24",
		Name:         "Rogue Class",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.LevelUp(2, effects.ManaCost("{1}{U}{B}")),
			effects.LevelUp(3, effects.ManaCost("{2}{U}{B}")),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.CombatDamageToPlayerBy(ev, source.Controller, g)
			}, "Rogue Class — exile the top card of that player's library face down; you may look at it", rogueClassExile),
		},
		Static: []game.StaticAbility{menace},
		GatedCastPermissions: []game.CastPermissionGate{{
			Permission: game.CastPermission{
				Zone:     game.ZoneExile,
				Filter:   game.PermissionFilter{ExiledWithSource: true},
				AnyColor: true,
				Label:    "Rogue Class — play cards exiled with this Class, spending mana as though it were mana of any color",
			},
			ActiveWhen: effects.Level(3),
		}},
	})
}

// rogueClassExile exiles the damaged player's top card face down, linked to
// this Class and looked at by its controller.
func rogueClassExile(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	victim := g.PlayerByIDForEffect(item.Trigger.Event.Target)
	if victim == nil || victim.Library == nil || len(victim.Library.Cards) == 0 {
		return nil
	}
	top := victim.Library.Cards[len(victim.Library.Cards)-1].InstanceID
	_, err := g.ExileHiddenLinkedForEffect(effects.HiddenRefOfActivation(item), item.Controller, top)
	return err
}
