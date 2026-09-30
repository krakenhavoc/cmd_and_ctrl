package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Derelict Attic // Widow's Walk — Enchantment — Room (ADR 0103):
//
//	Derelict Attic {2}{B}: "When you unlock this door, you draw two
//	cards and you lose 2 life."
//	Widow's Walk {3}{B}: "Whenever a creature you control attacks
//	alone, it gets +1/+0 and gains deathtouch until end of turn."
//
// Widow's Walk reads "alone" the way Exalted does (CR 508.3: the only
// creature in the whole declaration) and pumps the ATTACKING creature,
// read off the triggering event.
func init() {
	Register(Room(RoomSpec{
		OracleID:     "993b7b94-ed06-422d-9c7e-52a74ce9d045",
		Name:         "Derelict Attic // Widow's Walk",
		Completeness: CompletenessFull,
		Left: Door{Triggered: []game.TriggeredAbility{
			WhenYouUnlockThisDoor(game.DoorLeft, "Derelict Attic — you draw two cards and you lose 2 life",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (DrawCards{N: 2}).Apply(ctx); err != nil {
						return err
					}
					return g.ChangePlayerLifeForEffect(ctx.Source(), item.Controller, -2)
				}),
		}},
		Right: Door{Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller) && attackedAlone(g)
			}, "Widow's Walk — it gets +1/+0 and gains deathtouch until end of turn", widowsWalkLoneAttacker),
		}},
	}))
}

// widowsWalkLoneAttacker gives the lone attacker (item.Trigger.Event's
// creature) +1/+0 and deathtouch until end of turn (CR 611.2c).
func widowsWalkLoneAttacker(g *game.Game, item *game.StackItem) error {
	return untilEndOfTurn(NewContext(g, item), item.Trigger.Event.CardID, nil,
		"Widow's Walk — +1/+0 and deathtouch", game.ModifyPTMod(1, 0), game.AddKeywordsMod("deathtouch"))
}
