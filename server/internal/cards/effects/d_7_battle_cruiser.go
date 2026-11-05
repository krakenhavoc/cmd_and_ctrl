package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// D-7 Battle Cruiser — {1}{R} Artifact — Spacecraft, 3/3 (Star Trek,
// #4973):
//
//	"Station (Tap another creature you control: Put charge counters
//	 equal to its power on this Spacecraft. Station only as a sorcery.
//	 It's an artifact creature at 8+.)
//	 3+ | At the beginning of combat on your turn, create a 2/1 red
//	 Klingon Warrior creature token with haste. Sacrifice it at the
//	 beginning of the next end step.
//	 8+ | Flying"
//
// Galvanizing Sawship's station body with a charge-gated trigger: the 3+
// line is a triggered ability that exists only while the Spacecraft has
// three or more charge counters (CR 721.2a, TriggeredAbility.ActiveWhen).
// The token is created through the keyed creation door and the sacrifice
// is Determined Iteration's: a CR 603.7 delayed trigger over exactly the
// tokens this trigger made, so a replacement that makes more (or none)
// is followed.
//
// No simplification.
func init() {
	combat := AtBeginningOfYourCombat("D-7 Battle Cruiser — create a 2/1 Klingon Warrior with haste; sacrifice it at the next end step",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			return g.CreateTokensKeyedForEffect(game.TokenCreation{
				Controller: item.Controller,
				Source:     item.SourceCardID,
				Groups:     []game.TokenGroup{{Template: ctx.TextToken(TokenCard("2/1 red Klingon Warrior with haste")), Count: 1}},
			}, ctx.ItemTail(d7BattleCruiserSacrificeAtEnd))
		})
	combat.ActiveWhen = AtChargeCounters(3)
	Register(Spec{
		OracleID:     "5e696f9a-425c-470a-870e-10d2ba3bb27d",
		Name:         "D-7 Battle Cruiser",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Station()},
		Static: append(
			SpacecraftAt(8, 3, 3),
			ThresholdKeywords(8, "flying"),
		),
		Triggered: []game.TriggeredAbility{combat},
	})
}

// d7BattleCruiserSacrificeAtEnd schedules the sacrifice of the token(s) the
// trigger made (the tail's IDs). An on-disk key.
var d7BattleCruiserSacrificeAtEnd = CardTail("d-7-battle-cruiser/sacrifice-at-end", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label: "D-7 Battle Cruiser — sacrifice the Klingon Warrior",
		Cards: o.IDs,
		Body:  sacrificeListedCardsBody,
	}.Apply(ctx)
})
