package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Omnath, Locus of Rage — Legendary Creature — Elemental,
// {3}{R}{R}{G}{G}, 5/5 (EDHREC rank 945):
//
//	"Landfall — Whenever a land you control enters, create a 5/5 red
//	 and green Elemental creature token.
//	 Whenever Omnath or another Elemental you control dies, Omnath
//	 deals 3 damage to any target."
//
// The landfall commander: every land is a 5/5, and every 5/5 that
// dies (or Omnath himself) is a Lightning Bolt with an extra point.
// Two triggers. Landfall is Tireless Provisioner's condition with a
// token. The dies trigger fires for Omnath's own death (the LTB
// harvest finds him in the graveyard and hands the ability his
// battlefield characteristics) and for any other Elemental you
// control — the tokens his own first ability makes are Elementals,
// which is the whole card — and it targets "any target", picked
// when it goes on the stack. The damage source is Omnath whether he
// is on the battlefield or in the graveyard, and he is red, so
// Torbran adds to it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1816eede-c5bd-49df-958f-a3af64cb2932",
		Name:         "Omnath, Locus of Rage",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Landfall("Omnath, Locus of Rage — create a 5/5 Elemental", Do(CreateToken{Template: b08RedGreenElementalToken(), N: 1})),
			{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if cardDied(ev, source) {
						return true
					}
					dead, ok := diedCreature(ev, g)
					return ok && dead.Controller == source.Controller && dead.HasSubtype("Elemental")
				},
				Targets: TargetAny(),
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Omnath, Locus of Rage — 3 damage to any target",
						func(g *game.Game, item *game.StackItem) error {
							if len(item.Targets) == 0 {
								return nil
							}
							return DealDamage{Source: item.SourceCardID, Target: item.Targets[0].ID, Amount: 3}.Apply(NewContext(g, item))
						})
				},
			},
		},
	})
}
