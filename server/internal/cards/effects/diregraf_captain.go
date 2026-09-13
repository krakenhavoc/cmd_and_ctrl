package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Diregraf Captain — Creature — Zombie Soldier {1}{U}{B}, 2/2 (EDHREC
// rank 2310):
//
//	"Deathtouch
//	 Other Zombie creatures you control get +1/+1.
//	 Whenever another Zombie you control dies, target opponent loses
//	 1 life."
//
// The Dimir Zombie lord. Deathtouch rides PrintedKeywords; the
// anthem is TribalAnthem over "other Zombies you control" — both
// words printed, so it buffs only your side; the drain is a targeted
// dies trigger: another Zombie the controller controlled died (the
// dead card is read post-move, so a changeling counts and a Zombie
// only through a layer effect does not — weaker, never stronger),
// the controller picks the opponent as the trigger goes on the
// stack, and that player loses 1 life (life loss, not damage) on
// resolution.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "70de24d9-c585-4bf6-ac2c-c5b4b7aa298c",
		Name:            "Diregraf Captain",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Zombie"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.CardID == source.InstanceID {
					return false
				}
				dead, ok := diedCreature(ev, g)
				return ok && dead.Controller == source.Controller && dead.HasSubtype("Zombie")
			},
			Targets: TargetPlayer("target opponent", Opponent()),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Diregraf Captain — target opponent loses 1 life",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							if t.Kind == game.TargetPlayer {
								return g.ChangePlayerLifeForEffect(item.SourceCardID, t.ID, -1)
							}
						}
						return nil
					})
			},
		}},
	})
}
