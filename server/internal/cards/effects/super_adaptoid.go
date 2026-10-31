package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Super-Adaptoid — Legendary Artifact Creature — Robot Villain {2}, */2
// (Marvel Super Heroes, #4455):
//
//	"Super-Adaptoid's power is equal to the number of legendary creatures
//	 you control.
//	 Whenever Super-Adaptoid enters or attacks, choose another target
//	 creature. If that creature has haste and Super-Adaptoid doesn't, put a
//	 haste counter on Super-Adaptoid. Do the same for flying, first strike,
//	 double strike, deathtouch, indestructible, lifelink, menace, reach,
//	 trample, and vigilance."
//
// The power is a layer 7a characteristic-defining ability counting the
// legendary creatures its controller has, itself included (Adeline's
// shape). The trigger has one mandatory target. As it resolves, each of the
// eleven keywords is judged in the printed order, one after another, against
// the creature as it is then and against Super-Adaptoid with the counters
// already given: a keyword counter (CR 122.1b) is a real ability, so a
// Super-Adaptoid that already has it, from a counter or otherwise, gets no
// second. A target that is no longer a legal one gives nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f93fa6ec-1f13-4109-aa60-11e148f7ba88",
		Name:         "Super-Adaptoid",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "Power is equal to the number of legendary creatures you control.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 0
				for i := range g.Battlefield.Cards {
					b := &g.Battlefield.Cards[i]
					if b.Controller == source.Controller && b.IsCreature() && b.IsLegendary() {
						n++
					}
				}
				c.Power = n
			},
		}},
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB, game.EventAttack},
				AppliesTo: Self,
				Key:       "Super-Adaptoid — copy keywords from another target creature as counters",
				TargetsFrom: AnotherTarget(func(other CardPredicate) *game.TargetSpec {
					return TargetCreature("another target creature", other)
				}),
				Effect: mshA8AdaptoidCopyKeywords,
			},
		},
	})
}

// mshA8AdaptoidKeywords are the eleven keywords, in printed order.
var mshA8AdaptoidKeywords = []string{
	game.CounterHaste, game.CounterFlying, game.CounterFirstStrike,
	game.CounterDoubleStrike, game.CounterDeathtouch, game.CounterIndestructible,
	game.CounterLifelink, game.CounterMenace, game.CounterReach,
	game.CounterTrample, game.CounterVigilance,
}

// mshA8AdaptoidCopyKeywords is the trigger's resolution.
func mshA8AdaptoidCopyKeywords(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target := firstLegalCardTarget(ctx)
	if target == uuid.Nil {
		return nil
	}
	self := ctx.Source()
	for _, kw := range mshA8AdaptoidKeywords {
		other, ok := g.LookupCardForEffect(target)
		if !ok || !other.IsCreature() || !game.HasKeyword(&other, kw) {
			continue
		}
		me, ok := g.LookupCardForEffect(self)
		if !ok || !onBattlefield(g, self) || game.HasKeyword(&me, kw) {
			continue
		}
		if err := (AddCounter{Target: self, Kind: kw, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
