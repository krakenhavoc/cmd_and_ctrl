package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Peregrine Drake — "Flying. When this creature enters, untap up
// to five lands."
//
// The blink engine's mana battery: flickering the Drake re-triggers
// the untap.
//
// Sandbox simplification: "up to five lands" is any lands in paper,
// and the choice is the controller's. Until a multi-target
// "up to N permanents" picker exists (the S20 multi-target work
// covers spells, not trigger-side free picks), this untaps up to
// five TAPPED LANDS THE CONTROLLER CONTROLS, in battlefield order.
// That is the overwhelmingly common use and never untaps an
// opponent's land, so the simplification is strictly conservative.
func init() {
	Register(Spec{
		OracleID:        "0bd67481-6bd9-48d6-92bd-8933b5ea1eae",
		Name:            "Peregrine Drake",
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Peregrine Drake — untap up to five lands",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						untapped := 0
						for _, c := range g.BattlefieldCardsForEffect() {
							if untapped >= 5 {
								break
							}
							if !c.IsLand() || c.Controller != item.Controller || !c.Tapped {
								continue
							}
							if err := (UntapTarget{Target: c.InstanceID}).Apply(ctx); err != nil {
								return err
							}
							untapped++
						}
						return nil
					})
			},
		}},
	})
}
