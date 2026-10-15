package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dragon Tempest — Enchantment {1}{R} (EDHREC top 1,000, #3654):
//
//	"Whenever a creature you control with flying enters, it gains haste
//	 until end of turn.
//	 Whenever a Dragon you control enters, it deals X damage to any
//	 target, where X is the number of Dragons you control."
//
// Two independent triggers, both off a creature of yours entering (the
// source is an enchantment, so "another" never matters). A creature
// with flying is judged as it entered, so a Dragon with flying fires
// both. The damage is dealt by the Dragon that entered, not by the
// enchantment, and X is counted as the trigger resolves; if that Dragon
// has left the battlefield by then the damage is dealt by its last
// known self and X counts whatever Dragons remain.
//
// No simplification.
func init() {
	const hasteLabel = "Dragon Tempest — it gains haste until end of turn"
	const damageLabel = "Dragon Tempest — it deals X damage to any target, where X is the number of Dragons you control"
	Register(Spec{
		OracleID:     "b9bacb46-fd1d-459b-81c2-d7d08fe73848",
		Name:         "Dragon Tempest",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature() && game.HasKeyword(&c, "flying")
			}, hasteLabel, func(g *game.Game, item *game.StackItem) error {
				return GrantKeywordUntilEOT{
					Target:   item.Trigger.Event.CardID,
					Keywords: []string{"haste"},
					Label:    "Dragon Tempest",
				}.Apply(NewContext(g, item))
			}),
			{
				Watches: []game.EventKind{game.EventETB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := enteredUnderYourControl(ev, source, g, false)
					return ok && c.IsCreature() && hasSubtype(c, "Dragon")
				},
				Targets: TargetAny(),
				Key:     damageLabel,
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					ts := ctx.LegalTargets()
					if len(ts) == 0 {
						return nil
					}
					x := 0
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.IsCreature() && hasSubtype(c, "Dragon") {
							x++
						}
					}
					return DealDamage{
						Source: item.Trigger.Event.CardID,
						Target: ts[0].ID,
						Amount: x,
					}.Apply(ctx)
				},
			},
		},
	})
}
