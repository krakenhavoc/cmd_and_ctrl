package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Speedball, New Warrior — Legendary Creature — Human Hero {2}{U/R}, 2/2
// (Marvel Super Heroes, #4455):
//
//	"Whenever a player casts a spell that targets Speedball, he gets +2/+2
//	 until end of turn. You may choose new targets for that spell."
//
// Heroic's trigger (a spell on the stack with this creature among its
// targets as it is cast, CR 601.2c) without the "you" gate: any player's
// spell qualifies, an ability does not, and a spell naming Speedball twice
// triggers it once. The boost lands first and the retarget is Deflecting
// Swat's (ChangeTargets, choose new targets, optional, CR 115.7): Speedball's
// controller is the chooser, whoever cast the spell, and may leave every
// target as it was. A spell that has left the stack by then is left alone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "53d3ef8f-8e0f-4119-8cdc-a0bf076b4e91",
		Name:         "Speedball, New Warrior",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				item := g.StackItemForEffect(ev.CardID)
				if item == nil || item.Kind != game.StackItemSpell {
					return false
				}
				for _, t := range item.Targets {
					if t.Kind == game.TargetCard && t.ID == source.InstanceID {
						return true
					}
				}
				return false
			}, "Speedball — +2/+2 until end of turn, and you may choose new targets for that spell", Do(
				Instruction(func(ctx *Context) error {
					return BoostUntilEOT{Target: ctx.Source(), Power: 2, Toughness: 2}.Apply(ctx)
				}),
				Instruction(func(ctx *Context) error {
					return ChangeTargets{
						StackID:  ctx.Item.Trigger.Event.CardID,
						Policy:   game.RetargetChooseNew,
						Optional: true,
						Reason:   "Speedball — choose new targets for that spell",
					}.Apply(ctx)
				}),
			)),
		},
	})
}
