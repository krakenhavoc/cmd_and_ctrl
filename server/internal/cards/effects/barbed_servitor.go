package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Barbed Servitor — Artifact Creature — Construct {3}{B}, 2/2:
//
//	"Indestructible
//	 When this creature enters, suspect it. (It has menace and can't
//	 block.)
//	 Whenever this creature deals combat damage to a player, you draw a
//	 card and you lose 1 life.
//	 Whenever this creature is dealt damage, target opponent loses that
//	 much life."
//
// The proof card for suspect (CR 701.60, #2698). The first ability is
// the Suspect primitive aimed at the source; menace and the can't-block
// rule are the engine's, not the card's, so there is no "menace" in
// PrintedKeywords and nothing here that would survive being un-suspected.
//
// The damage trigger is Brash Taunter's shape (b35SelfWasDealtDamage)
// with a life loss instead of damage: the controller picks the opponent
// as the trigger goes on the stack, and the life they lose is the
// event's amount. It is life LOSS, not damage, so a damage prevention
// effect on the opponent does not stop it.
//
// The engine emits one damage event per SOURCE, so the trigger waits for
// the damage to finish (AtBatchEnd, ThatMuchDamage) and loses the opponent the
// TOTAL the Servitor was dealt at the same time, as printed (CR 120.3, CR 510.2).
func init() {
	Register(Spec{
		OracleID:        "e3d82066-d8b9-47bf-8821-1370c506970b",
		Name:            "Barbed Servitor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Barbed Servitor — suspect it", func(g *game.Game, item *game.StackItem) error {
				return Suspect{Target: item.SourceCardID}.Apply(NewContext(g, item))
			}),
			WheneverThisDealsCombatDamageToAPlayer("Barbed Servitor — you draw a card and you lose 1 life", b36DrawAndLoseOne),
			{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return b35SelfWasDealtDamage(ev, source)
				},
				Targets:    TargetPlayer("target opponent", Opponent()),
				Key:        "Barbed Servitor — target opponent loses that much life",
				AtBatchEnd: ThatMuchDamage(),
				Effect:     damagedCreatureDrainsChosenOpponent,
			},
		},
	})
}

// damagedCreatureDrainsChosenOpponent is "whenever this creature is
// dealt damage, target opponent loses that much life": the chosen
// player loses the triggering event's amount. Life loss, so it does not
// go through the damage-prevention window; read off the item's event
// rather than captured, like damagedCreatureReflectsChosenAmount.
func damagedCreatureDrainsChosenOpponent(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Trigger == nil {
		return nil
	}
	victim := item.Targets[0]
	if victim.Kind != game.TargetPlayer || !NewContext(g, item).IsTargetLegal(victim) {
		return nil
	}
	return g.ChangePlayerLifeForEffect(item.SourceCardID, victim.ID, -item.Trigger.Event.Amount)
}
