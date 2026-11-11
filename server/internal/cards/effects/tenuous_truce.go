package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tenuous Truce — Enchantment — Aura {1}{W} (Bloomburrow Commander, #5368):
//
//	"Enchant opponent
//	 At the beginning of enchanted opponent's end step, you and that
//	 player each draw a card.
//	 When you attack enchanted opponent or a planeswalker they control or
//	 when they attack you or a planeswalker you control, sacrifice this
//	 Aura."
//
// An Aura on a player (Maddening Hex's EnchantPlayer, an opponent only).
// The draw is the Aura controller's and the enchanted player's, on that
// player's end step only. The sacrifice is one ability with two events
// (CR 603.2): an attack declared by the Aura's controller at the enchanted
// player or at a planeswalker that player controls, or one declared by the
// enchanted player at the Aura's controller or a planeswalker they control
// (attacksYouOrYourPlaneswalker: a battle is neither, CR 508.3a). One
// declaration is one trigger however many creatures attack (OncePerBatch).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f0bf2923-aea2-4bac-b061-64bc84437ef1",
		Name:         "Tenuous Truce",
		Completeness: CompletenessFull,
		Targets:      EnchantPlayer(Opponent()),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source.AttachedTo.Kind == game.TargetPlayer && source.AttachedTo.ID != uuid.Nil && ev.Actor == source.AttachedTo.ID
			}, "Tenuous Truce — you and enchanted opponent each draw a card", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (DrawCards{N: 1}).Apply(ctx); err != nil {
					return err
				}
				aura, ok := g.LookupCardForEffect(item.SourceCardID)
				if !ok || aura.AttachedTo.Kind != game.TargetPlayer {
					return nil
				}
				if p := g.PlayerByIDForEffect(aura.AttachedTo.ID); p == nil || p.Eliminated {
					return nil
				}
				return DrawCards{Player: aura.AttachedTo.ID, N: 1}.Apply(ctx)
			}),
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if source.AttachedTo.Kind != game.TargetPlayer || source.AttachedTo.ID == uuid.Nil {
					return false
				}
				enchanted := source.AttachedTo.ID
				switch ev.Actor {
				case source.Controller:
					return attacksYouOrYourPlaneswalker(g, ev.Target, enchanted)
				case enchanted:
					return attacksYouOrYourPlaneswalker(g, ev.Target, source.Controller)
				}
				return false
			}, "Tenuous Truce — sacrifice this Aura (an attack between you and the enchanted opponent)", SacrificeThisIfStillOnBattlefield)),
		},
	})
}
