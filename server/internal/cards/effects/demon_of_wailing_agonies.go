package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Demon of Wailing Agonies — Creature — Demon {3}{B}{B}, 4/4 (Commander
// 2014, #5553):
//
//	"Flying
//	 Lieutenant — As long as you control your commander, this creature gets
//	 +2/+2 and has "Whenever this creature deals combat damage to a player,
//	 that player sacrifices a creature of their choice.""
//
// Tyrant's Familiar's lieutenant: the condition is a commander you own and
// control on the battlefield (b18ControlsYourCommander), the +2/+2 a layer
// 7c static, and the granted trigger is written as the card's own, gated on
// the same condition when the damage is dealt. Once it has triggered,
// losing the commander does not stop it resolving. The damaged player picks
// the creature as it resolves (Cabal Executioner's edict); the edict is not
// targeted, so hexproof does not stop it.
//
// No simplification.
func init() {
	edict := TriggerEdict(WheneverThisDealsCombatDamageToAPlayer("Demon of Wailing Agonies — Lieutenant: that player sacrifices a creature",
		func(g *game.Game, item *game.StackItem) error {
			victim := NewContext(g, item).Trigger().Event.Target
			if g.PlayerByIDForEffect(victim) == nil {
				return nil
			}
			g.PlayerSacrificesForEffect(item.SourceCardID, victim, SacrificeSpec("a creature", Creature()),
				"Demon of Wailing Agonies — sacrifice a creature")
			return nil
		}))
	Register(Spec{
		OracleID:        "2ee46c98-aee7-459a-80fa-4ed9d8a3acda",
		Name:            "Demon of Wailing Agonies",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{C14LieutenantPump()},
		Triggered:       []game.TriggeredAbility{C14WhileLieutenant(edict)},
	})
}

// c14LieutenantPump is the "+2/+2 as long as you control your commander"
// half of Commander 2014's lieutenant creatures: a layer 7c pump of the
// source alone, on while its controller controls a commander they own.
func C14LieutenantPump() game.StaticAbility {
	return game.StaticAbility{
		Label:    "Lieutenant — As long as you control your commander, this creature gets +2/+2.",
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7C_Modify,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return target.InstanceID == source.InstanceID && B18ControlsYourCommander(g, source.Controller)
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Power += 2
			c.Toughness += 2
		},
	}
}

// c14WhileLieutenant gates a trigger the lieutenant grants: the creature
// has it only while its controller controls their commander, so it
// triggers only then.
func C14WhileLieutenant(t game.TriggeredAbility) game.TriggeredAbility {
	when := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return when(ev, source, lki, g) && B18ControlsYourCommander(g, source.Controller)
	}
	return t
}
