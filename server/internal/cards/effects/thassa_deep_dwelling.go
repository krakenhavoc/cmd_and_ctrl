package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thassa, Deep-Dwelling — Legendary Enchantment Creature — God
// {3}{U}, 6/5:
//
//	"Indestructible
//	 As long as your devotion to blue is less than five, Thassa isn't
//	 a creature.
//	 At the beginning of your end step, exile up to one other target
//	 creature you control, then return that card to the battlefield
//	 under your control.
//	 {3}{U}: Tap another target creature."
//
// THE GOD CLAUSE is Erebos's shape: a layer 4 self-static, re-read on
// every recompute, that removes the Creature type (and with it the
// creature subtypes — the shared notACreature helper, #670) while devotionTo
// blue (CR 700.5) is under five. Thassa's own {U} counts.
//
// THE END-STEP BLINK is an immediate flicker (Flicker) — exile and
// return resolve together, so the creature is a new object (CR 400.7)
// that re-fires its ETBs. It comes back "under YOUR control", so the
// trigger's controller is passed as Flicker's Controller: a creature
// you had stolen stays yours. "Up to one" is the (0, 1) count and
// "other" is AnotherTarget (by instance, through TargetsFrom), so
// Thassa never blinks herself and a same-named copy still can be.
//
// THE ACTIVATED ABILITY taps "another target creature". An activated
// ability's clause is static (ActivatedAbility has no TargetsFrom), so
// "another" is the catalog's by-name exclusion, as Dour Port-Mage and
// Reflection of Kiki-Jiki write it. Thassa is legendary: the only
// other permanent with her name that can share a battlefield with her
// is a non-legendary copy.
func init() {
	Register(Spec{
		OracleID:        "2396299a-c031-4020-b13e-1f9bf9d64511",
		Name:            "Thassa, Deep-Dwelling",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Static: []game.StaticAbility{{
			Layer: game.Layer4Type,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID &&
					devotionTo(g, source.Controller, "U") < 5
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				notACreature(c)
			},
		}},
		Triggered: []game.TriggeredAbility{thassaEndStepBlink()},
		Activated: []ActivatedAbility{{
			Label:   "{3}{U}: Tap another target creature.",
			Cost:    ManaCost("{3}{U}"),
			Targets: TargetCreature("another target creature", b03NotNamed("Thassa, Deep-Dwelling")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok || id == item.SourceCardID {
					return nil
				}
				return TapTarget{Target: id}.Apply(ctx)
			},
		}},
	})
}

// thassaEndStepBlink is "At the beginning of your end step, exile up
// to one other target creature you control, then return that card to
// the battlefield under your control."
func thassaEndStepBlink() game.TriggeredAbility {
	t := AtYourEndStep("Thassa, Deep-Dwelling — blink another creature you control",
		flickerFirstLegalTarget)
	t.TargetsFrom = AnotherTarget(func(other CardPredicate) *game.TargetSpec {
		return TargetCreature("up to one other target creature you control", YouControl(), other).WithCount(0, 1)
	})
	return t
}
