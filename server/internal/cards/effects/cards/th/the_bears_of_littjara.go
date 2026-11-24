package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Bears of Littjara — Enchantment — Saga {1}{G}{U} (Kaldheim, #5505):
//
//	"I — Create a 2/2 blue Shapeshifter creature token with changeling.
//	 II — Any number of target Shapeshifter creatures you control have base
//	 power and toughness 4/4.
//	 III — Choose up to one target creature or planeswalker. Each creature
//	 with power 4 or greater you control deals damage equal to its power to
//	 that permanent."
//
// Chapter II sets base power and toughness (layer 7b, CR 613.4b) on each
// chosen Shapeshifter with no stated duration, so it lasts until the end of
// the game (CR 611.2a). A changeling is every creature type (CR 702.73a),
// so it is a legal pick. Chapter III reads each qualifying creature's power
// as the chapter resolves, then every one deals its damage at the same time
// as one damage instance (CR 120.2, CR 608.2c).
//
// No simplification.
func init() {
	const name = "The Bears of Littjara"
	effects.Register(effects.Spec{
		OracleID:     "697f3fd9-1577-4f69-bdf8-db1f673d3505",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "create a 2/2 blue Shapeshifter creature token with changeling"),
				effects.Do(effects.CreateToken{Template: effects.BlueShapeshifterToken(), N: 1})),
			effects.ChapterTriggerTargeting(2, effects.SagaChapterLabel(name, 2, "any number of target Shapeshifter creatures you control have base power and toughness 4/4"),
				effects.TargetCreature("any number of target Shapeshifter creatures you control", effects.HasSubtype("Shapeshifter"), effects.YouControl()).WithCount(0, 0),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (effects.ScopedEffectFor{
							Target:   t.ID,
							Mods:     game.SetBasePTMods(4, 4),
							Duration: game.IndefiniteDuration(),
							Label:    name + " — base power and toughness 4/4",
						}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
			effects.ChapterTriggerTargeting(3, effects.SagaChapterLabel(name, 3, "each creature with power 4 or greater you control deals damage equal to its power to up to one target creature or planeswalker"),
				effects.TargetPermanent("up to one target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())).WithCount(0, 1),
				bearsOfLittjaraBite),
		},
	})
}

// bearsOfLittjaraBite is chapter III: each creature with power 4 or greater
// its controller controls, powers read once up front, deals that much damage
// to the chosen permanent as one damage instance.
func bearsOfLittjaraBite(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	victim, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	type biter struct {
		id    uuid.UUID
		power int
	}
	var biters []biter
	for _, c := range effects.MatchingBattlefield(ctx, effects.And(effects.Creature(), effects.YouControl())) {
		if p := c.CurrentPower(); p >= 4 {
			biters = append(biters, biter{c.InstanceID, p})
		}
	}
	if len(biters) == 0 {
		return nil
	}
	return g.DamageInstanceForEffect(func() error {
		for _, b := range biters {
			if err := (effects.DealDamage{Source: b.id, Target: victim, Amount: b.power}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}
