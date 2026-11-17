package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sword of Once and Future — Artifact — Equipment {3} (March of the
// Machine):
//
//	"Equipped creature gets +2/+2 and has protection from blue and from
//	 black.
//	 Whenever equipped creature deals combat damage to a player, surveil
//	 2. Then you may cast an instant or sorcery spell with mana value 2
//	 or less from your graveyard without paying its mana cost. If that
//	 spell would be put into your graveyard, exile it instead.
//	 Equip {2}"
//
// The surveil runs first and the cast waits for the two cards to be
// filed (Surveil.ThenTail), so a card the surveil binned can be the
// one cast. The cast is the free, exile-instead grant offered during the
// resolution (CR 608.2g); the mana value clause is judged on the face
// being cast (CR 601.3e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1e0cf860-6be6-4b4b-9a67-0c19b6018ae1",
		Name:         "Sword of Once and Future",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("protection from blue", "protection from black"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			},
			Key: "Sword of Once and Future — surveil 2, then cast an instant or sorcery with mana value 2 or less from your graveyard free",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return Surveil{Player: ctx.Controller(), N: 2, ThenTail: ctx.ItemTail(swordOnceAndFutureAfterSurveil)}.Apply(ctx)
			},
		}},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}

// swordOnceAndFutureAfterSurveil offers the free cast over the controller's
// graveyard once the surveil has filed its cards. An on-disk key: never
// renamed, never reused.
var swordOnceAndFutureAfterSurveil = CardTail("sword-of-once-and-future/after-surveil", func(ctx *Context, _ game.TailOutcome) error {
	spell := freeInstantOrSorcery().withManaValueAtMost(2)
	clause := spell.on(game.CastPermission{})
	cmdr2S42FreeGraveyardCast(ctx, "Sword of Once and Future", spell, true,
		graveyardCardIDs(ctx, ctx.Controller(), clause.OffersASpell))
	return nil
})
