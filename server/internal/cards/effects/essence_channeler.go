package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Essence Channeler — Creature — Bat Cleric {1}{W}, 2/1 (Bloomburrow,
// #4467):
//
//	"As long as you've lost life this turn, this creature has flying and
//	 vigilance.
//	 Whenever you gain life, put a +1/+1 counter on this creature.
//	 When this creature dies, put its counters on target creature you
//	 control."
//
// The static is a layer-6 keyword grant gated on the controller's life lost
// this turn (the turn tally), woken by every life change
// (StaticAbility.DependsOnEvents). The counters are read as the creature last
// existed on the battlefield (CR 603.10a, 608.2h) and every kind it had moves
// to the target, which must still be a legal target as the ability resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1473580d-7cb2-48f5-a77d-08f49244c5e8",
		Name:         "Essence Channeler",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:           "As long as you've lost life this turn, this creature has flying and vigilance.",
			Layer:           game.Layer6Ability,
			DependsOnEvents: []game.EventKind{game.EventChangeLife},
			AppliesTo:       selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				if g.TurnTallyFor(source.Controller).LifeLost <= 0 {
					return
				}
				c.Abilities = game.AppendKeywordAbility(c.Abilities, "flying")
				c.Abilities = game.AppendKeywordAbility(c.Abilities, "vigilance")
			},
		}},
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventChangeLife},
				AppliesTo: YouGainedLife,
				Key:       "Essence Channeler — put a +1/+1 counter on this creature",
				Effect:    Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}),
			},
			Targeting(WhenThisDies("Essence Channeler — put its counters on target creature you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					info, ok := ctx.SourcePermanent()
					if !ok {
						return nil
					}
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						for kind, n := range info.Counters {
							if n <= 0 {
								continue
							}
							if err := (AddCounter{Target: t.ID, Kind: kind, N: n}).Apply(ctx); err != nil {
								return err
							}
						}
						return nil
					}
					return nil
				}),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}
