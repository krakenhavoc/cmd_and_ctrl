package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Broodcaller Scourge — Creature — Dragon {5}{G}{G}, 5/7 (Tarkir:
// Dragonstorm Commander):
//
//	"Flying
//	 Whenever one or more Dragons you control deal combat damage to a player,
//	 you may put a permanent card with mana value less than or equal to that
//	 damage from your hand onto the battlefield."
//
// Keeper of Fables' batch trigger (one per damaged player per damage step,
// CR 603.2c) narrowed to Dragons, the Scourge itself included. "That damage" is
// the TOTAL the Dragons dealt that player in the batch: the amount is summed
// from the batch's damage events as the ability resolves, so two Dragons that
// hit for 3 and 4 make 7. The put is Elvish Piper's optional hand put (a
// permanent card only, the ruling: an {X} card's mana value in hand counts X as
// 0), narrowed to mana value at most the damage, and it enters through the
// ordinary pipeline.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "16aa790a-8f7c-401c-9d43-dc9440a6f585",
		Name:            "Broodcaller Scourge",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			OncePerBatchPerPlayer(On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return cmdrB13DragonCombatDamageToPlayer(ev, source, g, source.Controller)
			}, "Broodcaller Scourge — you may put a permanent card with mana value less than or equal to that damage from your hand onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					ctx := NewContext(g, item)
					total := 0
					reader, _ := g.LookupCardForEffect(item.SourceCardID)
					for _, ev := range g.Events {
						if ev.Batch != item.Trigger.Event.Batch || ev.Target != item.Trigger.Event.Target {
							continue
						}
						if cmdrB13DragonCombatDamageToPlayer(ev, &reader, g, item.Controller) {
							total += ev.Amount
						}
					}
					if total <= 0 {
						return nil
					}
					return PutFromHandOntoBattlefield{
						Optional: true,
						Label:    "Broodcaller Scourge — put a permanent card with mana value " + strconv.Itoa(total) + " or less from your hand onto the battlefield",
						Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
							return c.ManaValue() <= total
						},
					}.Apply(ctx)
				})),
		},
	})
}
