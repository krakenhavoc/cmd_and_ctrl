package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Varchild, Betrayer of Kjeldor — Legendary Creature — Human Knight {2}{R},
// 3/3 (Commander 2018, #5526):
//
//	"Whenever Varchild deals combat damage to a player, that player creates
//	 that many 1/1 red Survivor creature tokens.
//	 Survivors your opponents control can't block, and they can't attack you
//	 or planeswalkers you control.
//	 When Varchild leaves the battlefield, gain control of all Survivors."
//
// The damage trigger reads the damaged player and the amount off the
// triggering damage event, and that player creates the tokens under their
// own control.
//
// The block half is a layer 6 restriction (Bloodfeather Phoenix's CantBlock)
// on every Survivor controlled by someone other than Varchild's controller,
// judged on the effective creature type through Varchild's text. The attack
// half is Combat Calligrapher's restriction for Survivors: none may attack
// Varchild's controller or a planeswalker they control. A Survivor you
// control could never attack you, so binding every Survivor is the printed
// rule. Both follow a change of control of Varchild.
//
// The leaves trigger takes every Survivor on the battlefield as it
// resolves, under the trigger's controller (Varchild's controller as it
// left), with no end (Thieving Skydiver's indefinite duration).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "26ae37ab-91f8-4113-beb0-1154f9b62d28",
		Name:         "Varchild, Betrayer of Kjeldor",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label: "Survivors your opponents control can't block.",
				Layer: game.Layer6Ability,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.IsCreature() && target.Controller != source.Controller &&
						target.HasSubtype(source.TextSubtype("Survivor"))
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Restrictions |= game.CantBlock
				},
			},
			socB3SubtypeCantAttackYouOrYourPlaneswalkers("Survivors your opponents control can't attack you or planeswalkers you control.", "Survivor"),
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Varchild, Betrayer of Kjeldor — that player creates that many 1/1 red Survivor creature tokens",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil || item.Trigger.Event.Amount <= 0 {
						return nil
					}
					ev := item.Trigger.Event
					victim := g.PlayerByIDForEffect(ev.Target)
					if victim == nil || victim.Eliminated {
						return nil
					}
					return CreateToken{Controller: victim.ID, Template: TokenCard("1/1 red Survivor"), N: ev.Amount}.Apply(NewContext(g, item))
				}),
			WhenThisLeaves("Varchild, Betrayer of Kjeldor — gain control of all Survivors",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					survivor := ctx.TextSubtype("Survivor")
					for _, c := range g.BattlefieldCardsForEffect() {
						if !c.HasSubtype(survivor) {
							continue
						}
						if err := (GainControl{
							Target:   c.InstanceID,
							Duration: game.IndefiniteDuration(),
							Label:    "Varchild, Betrayer of Kjeldor — gain control of all Survivors",
						}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
