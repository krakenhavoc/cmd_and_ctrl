package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sword of Light and Shadow — Artifact — Equipment {3} (the Squall
// deck request, #3918):
//
//	"Equipped creature gets +2/+2 and has protection from white and
//	 from black.
//	 Whenever equipped creature deals combat damage to a player, you
//	 gain 3 life and you may return up to one target creature card from
//	 your graveyard to your hand.
//	 Equip {2}"
//
// The Mirrodin sword cycle's shared shape (Sword of Body and Mind, Sword
// of Fire and Ice): a layer-6 protection grant, the combat-damage
// trigger matched through the attachment, Equip {2}.
//
// "Up to one target creature card" is chosen as the trigger goes on the
// stack and may be none. The life is not conditional on the return: it
// is gained first, whatever happened to the target. "You may" is asked
// as the trigger resolves (CR 608.2d), once the card is known to still
// be there (CR 608.2b), so a player can decline a card they targeted.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "01c8a554-45a6-46ca-b40a-af0fd4529e40",
		Name:         "Sword of Light and Shadow",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("protection from white", "protection from black"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			},
			Targets: TargetCardInGraveyard("up to one target creature card from your graveyard", YouOwn(), Creature()).WithCount(0, 1),
			Key:     "Sword of Light and Shadow — gain 3 life, and you may return a creature card",
			Effect:  swordOfLightAndShadowHit,
		}},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}

// swordOfLightAndShadowHit gains the life, then offers the return of the
// still-legal target, if there is one.
func swordOfLightAndShadowHit(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (GainLife{Player: item.Controller, Amount: 3}).Apply(ctx); err != nil {
		return err
	}
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return MayChoice{
		Question: "Sword of Light and Shadow — return that creature card from your graveyard to your hand?",
		OnYes: func(ctx *Context) error {
			return ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}.Apply(ctx)
		},
	}.Apply(ctx)
}
