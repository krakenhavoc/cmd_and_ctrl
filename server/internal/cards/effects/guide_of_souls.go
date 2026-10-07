package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Guide of Souls — Creature — Human Cleric {W}, 1/2:
//
//	"Whenever another creature you control enters, you gain 1 life and
//	 get {E} (an energy counter).
//	 Whenever you attack, you may pay {E}{E}{E}. When you do, put two
//	 +1/+1 counters and a flying counter on target attacking creature. It
//	 becomes an Angel in addition to its other types."
//
// ADR 0129 §3 (#1995): "whenever you attack" is one trigger per attack
// declaration (OncePerBatch). The energy is paid as it resolves (CR
// 118.12); "when you do" is a reflexive trigger (CR 603.12) that targets
// as it goes on the stack. The Angel type has no stated duration, so it
// lasts as long as the creature stays on the battlefield (CR 611.2a,
// CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b304ac72-7f40-40de-b8d6-6392909b6029",
		Name:         "Guide of Souls",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 1},
		Triggered: []game.TriggeredAbility{
			WheneverAnotherCreatureEntersUnderYourControl("Guide of Souls — you gain 1 life and get {E}",
				gainLifeAndGetEnergy(1, 1)),
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, "Guide of Souls — you may pay {E}{E}{E}",
				mayPayEnergyThen("Guide of Souls", 3, "make an attacking creature an Angel",
					func(g *game.Game, item *game.StackItem) error {
						return WhenYouDo("Guide of Souls — two +1/+1 counters and a flying counter on target attacking creature; it becomes an Angel",
							guideOfSoulsAngelBody).Apply(NewContext(g, item))
					}))),
		},
	})
}

// guideOfSoulsAngel is the reflexive body: the counters, then the Angel
// type, on the target if it is still a legal one.
func guideOfSoulsAngel(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := g.AddCounterByForEffect(item.Controller, t.ID, game.CounterPlusOne, 2); err != nil {
			return err
		}
		if err := g.AddCounterByForEffect(item.Controller, t.ID, game.CounterFlying, 1); err != nil {
			return err
		}
		if err := (ScopedEffectFor{
			Target:   t.ID,
			Mods:     []game.Mod{game.AddSubtypesMod("Angel")},
			Duration: game.IndefiniteDuration(),
			Label:    "Guide of Souls — an Angel in addition to its other types",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
