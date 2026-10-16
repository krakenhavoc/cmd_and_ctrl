package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zurgo Helmsmasher — Legendary Creature — Orc Warrior, {2}{R}{W}{B}, 7/2:
//
//	"Haste"
//	"Zurgo Helmsmasher attacks each combat if able."
//	"During your turn, Zurgo Helmsmasher has indestructible."
//	"Whenever a creature dealt damage by Zurgo Helmsmasher this turn
//	 dies, put a +1/+1 counter on Zurgo Helmsmasher."
//
// The voltron commander that dares you to block it. Seven power on a
// four-mana commander is three hits to the 21-damage clock
// (CR 903.10a), and the two-toughness body that should make that
// suicidal is covered by conditional indestructible — but only on
// YOUR turn, which is the entire design.
//
// # The conditional static is the interesting part
//
// "During your turn" is a CR 613 continuous effect whose predicate
// reads the TURN, not the battlefield. `AppliesTo` runs on every
// recompute pass and re-answers the question each time, so the
// keyword genuinely blinks off the moment the turn passes: Zurgo
// attacks into a 5/5, survives, and then dies to a Shock on your
// opponent's upkeep. The layer engine bumps its version on turn
// change (layer_listener.go), so the recompute really does happen at
// the boundary rather than lazily.
//
// This is also the first card in the catalog whose static grants
// indestructible CONDITIONALLY, which is worth having: a predicate
// that can go false is the case a "grant once and forget"
// implementation of the keyword would have got wrong.
//
// # The counter trigger
//
// "Whenever a creature dealt damage by Zurgo Helmsmasher this turn dies"
// reads the turn's record of damaged objects through dealtDamageByThisTurn
// (#1946): the source and the victim are each a new object once they
// re-enter the battlefield (CR 400.7), and the damage counts as dealt
// after prevention. Zurgo must still be on the battlefield to get the
// counter, and the creature that died may be any player's.
//
// "Attacks each combat if able" is no longer on that list. #1595 gave
// the engine CR 508.1d attack-requirement enforcement
// (game/attack_requirements.go), and Zurgo's own line is the plainest
// shape it takes: AttacksEachCombat() is "~ attacks each combat if
// able" on the creature's own ability list, so a Zurgo that loses all
// abilities (CR 613.1f) stops being required to swing.
//
// All four lines are live.
func init() {
	Register(Spec{
		OracleID:        "6c48d888-9f5d-43f4-adbd-61dbdba09260",
		Name:            "Zurgo Helmsmasher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, died := diedCreature(ev, g)
				return died && ev.CardID != source.InstanceID && dealtDamageByThisTurn(g, source.InstanceID, ev.CardID)
			}, "Zurgo Helmsmasher — put a +1/+1 counter on it", func(g *game.Game, item *game.StackItem) error {
				if !onBattlefield(g, item.SourceCardID) {
					return nil
				}
				return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
			}),
		},
		Static: []game.StaticAbility{
			AttacksEachCombat(),
			{
				Layer: game.Layer6Ability,
				AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
					return target.InstanceID == source.InstanceID &&
						isActivePlayer(g, source.Controller)
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Abilities = append(c.Abilities, "indestructible")
				},
			},
		},
	})
}
