package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bartz and Boko — Legendary Creature — Human Bird {3}{G}{G}, 4/3
// (Final Fantasy, #4462):
//
//	"Affinity for Birds (This spell costs {1} less to cast for each Bird
//	 you control.)
//	 When Bartz and Boko enters, each other Bird you control deals damage
//	 equal to its power to target creature an opponent controls."
//
// Affinity is the shared self cost modifier (CR 702.41a), counting Bird
// permanents its caster controls. The enters trigger targets as it goes
// on the stack; as it resolves, each OTHER Bird creature its controller
// controls deals damage equal to its power at that moment, each Bird
// being the source of its own damage (so deathtouch, lifelink and
// prevention see the Bird), all as one damage event (Chandra's Ignition's
// DamageInstanceForEffect). A Bird with no power deals none. "Bird" is a
// word of the card's text (CR 612.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2d995c47-e548-4fc4-9182-fef46af20ca5",
		Name:         "Bartz and Boko",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for Birds", HasSubtype("Bird")),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Bartz and Boko — each other Bird you control deals damage equal to its power to target creature an opponent controls",
					bartzAndBokoBirdsStrike),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}

// bartzAndBokoBirdsStrike has each other Bird its controller controls
// deal its power in damage to the target.
func bartzAndBokoBirdsStrike(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	bird := ctx.TextSubtype("Bird")
	type hit struct {
		source game.Card
		power  int
	}
	var hits []hit
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == item.SourceCardID || c.Controller != item.Controller || !c.IsCreature() || !c.HasSubtype(bird) {
			continue
		}
		if p := c.CurrentPower(); p > 0 {
			hits = append(hits, hit{source: c, power: p})
		}
	}
	if len(hits) == 0 {
		return nil
	}
	return g.DamageInstanceForEffect(func() error {
		for _, h := range hits {
			if err := (DealDamage{Source: h.source.InstanceID, Target: target, Amount: h.power}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}
