package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kain, Traitorous Dragoon — Legendary Creature — Human Knight {2}{B}, 2/4
// (Final Fantasy, #4462):
//
//	"Jump — During your turn, Kain has flying.
//	 Whenever Kain deals combat damage to a player, that player gains control
//	 of Kain. If they do, you draw that many cards, create that many tapped
//	 Treasure tokens, then lose that much life."
//
// Jump is Cid, Freeflier Pilot's: a layer 6 flying grant on Kain while his
// controller is the active player, re-read every pass (so it follows the
// controller, and the new one has it on their turn). The trigger's "that
// many" is the combat damage dealt, carried on the triggering event, and the
// damaged player takes control with no stated duration (CR 611.2a). "If they
// do" is read after the layers settle: the damaged player really controls
// Kain now. The three payoffs are the original controller's ("you", CR
// 603.3a), and happen in the printed order: cards, then tapped Treasures,
// then the life loss.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "70fbb378-d905-47cf-8c54-cc166a7867b3",
		Name:         "Kain, Traitorous Dragoon",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && target.IsCreature() && isActivePlayer(g, source.Controller)
			}, "flying"),
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer(
				"Kain, Traitorous Dragoon — that player gains control of Kain; if they do, you draw that many cards, create that many tapped Treasures, then lose that much life",
				kainTraitor),
		},
	})
}

func kainTraitor(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ev := ctx.Trigger().Event
	n, victim := ev.Amount, ev.Target
	if victim == uuid.Nil || n <= 0 || !onBattlefield(g, item.SourceCardID) {
		return nil
	}
	if err := (GainControl{
		Target: item.SourceCardID, Controller: victim, Duration: game.IndefiniteDuration(),
		Label: "Kain, Traitorous Dragoon — that player gains control of Kain",
	}).Apply(ctx); err != nil {
		return err
	}
	g.RecomputeLayersIfStaleLocked()
	if c, ok := g.LookupCardForEffect(item.SourceCardID); !ok || c.Controller != victim {
		return nil
	}
	if err := (DrawCards{Player: item.Controller, N: n}).Apply(ctx); err != nil {
		return err
	}
	if err := (CreateTokenAdvanced{Spec: Token(TreasureToken()).EntersTapped(), N: n}).Apply(ctx); err != nil {
		return err
	}
	return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -n)
}
