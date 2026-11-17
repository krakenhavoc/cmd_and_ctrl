package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Whispersteel Dagger — Artifact — Equipment {2}{B} (Zendikar Rising
// Commander, #5509):
//
//	"Equipped creature gets +2/+0.
//	 Whenever equipped creature deals combat damage to a player, you may
//	 cast a creature spell from that player's graveyard this turn, and you
//	 may spend mana as though it were mana of any color to cast that spell.
//	 Equip {3}"
//
// The trigger is Dream-Thief's Bandana's "whenever equipped creature deals
// combat damage to a player", the damaged player the event's Target. Its
// resolution grants the Dagger's controller one stored permission, until
// end of turn, over that player's graveyard (ZoneOwner, Shaman's Trance's
// cross-seat grant): creature spells only, casting only, good for ONE spell
// ("a creature spell": CastsLeft 1, spent by the cast it opens), with the
// any-colour spend for that spell (CR 609.4b). Nothing is chosen as the
// trigger resolves (the ruling): the permission covers whatever creature
// card is in that graveyard when the player casts. The spell keeps its own
// timing and every cost, additional and alternative ones included (the
// rulings); it is the caster's to control (CR 601.2a) and goes to its
// owner's graveyard afterwards (CR 400.3).
//
// Each damaged player gets a permission of their own, labelled with that
// player, so a Dagger that connects with two players in one turn opens one
// creature spell from each graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ca59fbf4-c774-49d3-8630-108c053e01fb",
		Name:         "Whispersteel Dagger",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(2, 0)},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, cmdr2S45EquippedCreatureDealtCombatDamageToAPlayer,
				"Whispersteel Dagger — you may cast a creature spell from that player's graveyard this turn, spending mana as though it were mana of any color",
				whispersteelDaggerGrant),
		},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}

// whispersteelDaggerGrant opens one creature spell from the damaged
// player's graveyard to the Dagger's controller until end of turn.
func whispersteelDaggerGrant(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	victim := g.PlayerByIDForEffect(item.Trigger.Event.Target)
	if victim == nil {
		return nil
	}
	g.GrantCastPermissionForEffect(game.CastPermission{
		Player:     item.Controller,
		Zone:       game.ZoneGraveyard,
		Scope:      game.ScopeStanding,
		ZoneOwner:  victim.ID,
		Filter:     game.PermissionFilter{CreatureOnly: true},
		CastOnly:   true,
		CastsLeft:  1,
		AnyColor:   true,
		Source:     item.SourceCardID,
		SourceName: "Whispersteel Dagger",
		Label:      "Whispersteel Dagger — cast a creature spell from " + victim.Name + "'s graveyard",
	})
	return nil
}
