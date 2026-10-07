package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dragonspark Reactor — Artifact {1}{R}:
//
//	"Whenever this artifact or another artifact you control enters, put
//	 a charge counter on this artifact.
//	 {4}, Sacrifice this artifact: It deals damage equal to the number
//	 of charge counters on it to target player and that much damage to
//	 up to one target creature."
//
// One trigger that watches artifact entries under the Reactor's
// controller, itself included (it is on the battlefield when its own
// entry event is harvested). The activated ability is a two-clause
// target statement, a player and then up to one creature, answered in
// printed order. The sacrifice is a cost, so the number is the
// Reactor's last-known charge counters (CR 608.2h), read before either
// damage is dealt: the creature takes the same amount the player does.
// If one of the two targets is gone at resolution the other still takes
// its damage (CR 608.2b); only a fully illegal set fizzles it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a8acb0b0-7265-42f6-8143-62d0c02d5a28",
		Name:         "Dragonspark Reactor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsArtifact()
			},
			Key: "Dragonspark Reactor — put a charge counter on it",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return AddCounter{Target: item.SourceCardID, Kind: game.CounterCharge, N: 1}.Apply(NewContext(g, item))
			},
		}},
		Activated: []ActivatedAbility{{
			Label: "{4}, Sacrifice this artifact: It deals damage equal to the number of charge counters on it to target player and that much damage to up to one target creature.",
			Cost:  Plus(ManaCost("{4}"), SacrificeThis()),
			Targets: Clauses(
				TargetPlayer("target player"),
				TargetCreature("up to one target creature").WithCount(0, 1),
			),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				info, ok := ctx.SourcePermanent()
				if !ok {
					return nil
				}
				n := info.Counters[game.CounterCharge]
				if n <= 0 {
					return nil
				}
				return g.DamageInstanceForEffect(func() error {
					if p, ok := ctx.ClauseTarget(0); ok {
						if err := (DealDamage{Source: item.SourceCardID, Target: p.ID, Amount: n}).Apply(ctx); err != nil {
							return err
						}
					}
					for _, c := range ctx.ClauseTargets(1) {
						if err := (DealDamage{Source: item.SourceCardID, Target: c.ID, Amount: n}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				})
			},
		}},
	})
}
