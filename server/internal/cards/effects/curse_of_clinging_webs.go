package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Curse of Clinging Webs — Enchantment — Aura Curse {2}{G}:
//
//	"Enchant player
//	 Whenever a nontoken creature enchanted player controls dies, exile
//	 it and you create a 1/2 green Spider creature token with reach."
//
// Curse of Opulence's enchant-player shape watching a death: the dead
// creature is read as it last existed (CR 603.10a), so it counts when it
// was a nontoken creature under the enchanted player's control as it
// left. The body exiles the card only if it is still in a graveyard
// (a commander that took the command zone, or a card already
// reanimated, is left alone) and then makes the Spider for the Curse's
// controller whether or not the exile happened ("exile it and you
// create" are two instructions of the one effect).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d063f1d4-eb6b-40fa-97b7-f15ea53397fb",
		Name:         "Curse of Clinging Webs",
		Completeness: CompletenessFull,
		Targets:      EnchantPlayer(),
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if source.AttachedTo.Kind != game.TargetPlayer {
					return false
				}
				dead, ok := diedCreature(ev, g)
				return ok && !IsToken(dead) && leftUnderControlOf(ev, dead) == source.AttachedTo.ID
			}, "Curse of Clinging Webs — exile it and create a 1/2 green Spider with reach",
				Do(Instruction(func(ctx *Context) error {
					dead := ctx.Item.Trigger.Event.CardID
					if z := ctx.Game.FindCardZoneForEffect(dead); z != nil && z.Kind == game.ZoneGraveyard {
						if err := (ExileTarget{Target: dead}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
					CreateToken{Template: TokenCard("1/2 green Spider with reach"), N: 1})),
		},
	})
}
