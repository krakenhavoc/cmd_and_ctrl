package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Uthros Research Craft — Artifact — Spacecraft, {2}{U}, 0/8:
//
//	Station (Tap another creature you control: Put charge counters
//	equal to its power on this Spacecraft. Station only as a sorcery.
//	It's an artifact creature at 12+.)
//	3+ | Whenever you cast an artifact spell, draw a card. Put a
//	charge counter on this Spacecraft.
//	12+ | Flying
//	This Spacecraft gets +1/+0 for each artifact you control.
//
// The station card with a gated TRIGGER, which is the ADR 0071 shape
// none of the other proof cards exercise: below three charge counters
// the cast trigger does not exist (TriggeredAbility.ActiveWhen), so an
// artifact cast at two counters draws nothing. From three on, every
// artifact spell draws and ticks the Craft one step closer to twelve
// by itself.
//
// "Gets +1/+0 for each artifact you control" is printed below the
// threshold bars, so it is always on — a layer-7c modification that
// counts the Craft itself (it is an artifact you control). It only
// matters once the Craft is a creature at 12+, and until then a
// noncreature permanent with a modified power is simply not read by
// anything.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e1c9783a-1d1b-40d7-872e-0ca11b229ce6",
		Name:         "Uthros Research Craft",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Station()},
		Triggered:    []game.TriggeredAbility{uthrosResearchCraftCastTrigger()},
		Static: append(
			SpacecraftAt(12, 0, 8),
			ThresholdKeywords(12, "flying"),
			game.StaticAbility{
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: selfOnly,
				Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
					c.Power += artifactsControlledBy(g, source)
				},
			},
		),
	})
}

// uthrosResearchCraftCastTrigger is the "3+" line: an ordinary cast
// trigger (YouCast, Sai, Master Thopterist's shape) behind a charge-counter
// gate. The charge counter it adds goes on the SOURCE, through the
// counter window, so Doubling Season doubles it — and only while the
// Craft is still the object that triggered: one that left and came
// back is a new object (CR 400.7, #1418) and gets nothing.
func uthrosResearchCraftCastTrigger() game.TriggeredAbility {
	t := On(game.EventCast, YouCast(Artifact()), "Uthros Research Craft — draw a card and add a charge counter",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			if err := (DrawCards{N: 1}).Apply(ctx); err != nil {
				return err
			}
			if g.AbilitySourceGoneForEffect(item) {
				return nil
			}
			return AddCounter{Target: item.SourceCardID, Kind: game.CounterCharge, N: 1}.Apply(ctx)
		})
	t.ActiveWhen = AtChargeCounters(3)
	return t
}
