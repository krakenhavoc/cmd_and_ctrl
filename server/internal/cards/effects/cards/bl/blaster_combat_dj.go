package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blaster, Combat DJ // Blaster, Morale Booster (#3195, ADR 0152) — a
// More Than Meets the Eye card, and one of the two with no living metal.
//
// Front face, Legendary Artifact Creature — Robot {3}{R}{G}, 3/3:
//
//	"More Than Meets the Eye {1}{R}{G}
//	 Other nontoken artifact creatures and Vehicles you control have
//	 modular 1. (They enter with an additional +1/+1 counter on them.
//	 When they die, you may put their +1/+1 counters on target artifact
//	 creature.)
//	 Whenever you put one or more +1/+1 counters on Blaster, convert
//	 it."
//
// Back face, Legendary Artifact (not a creature, no power or toughness):
//
//	"Modular 3
//	 {X}, {T}: Move X +1/+1 counters from Blaster onto another target
//	 artifact. That artifact gains haste until end of turn. If Blaster
//	 has no +1/+1 counters on it, convert it. Activate only as a
//	 sorcery."
//
// Modular is the engine's keyword (game/modular.go): the entry counters
// and the dies trigger are derived from the LAST-KNOWN ability list, so a
// granted instance works for both, and the entry look-ahead sees a
// static from a Blaster already on the battlefield. A card cast converted
// enters as the back face, so its own "Modular 3" gives it three counters.
// The grant is a layer-6 static over the controller's other nontoken
// artifact creatures and Vehicles; a non-creature Vehicle is covered by
// the second noun. "Move X" moves as many as it holds when X is more
// (CR 122.1's removal cannot exceed the counters present); the convert
// reads the counters left after the move. "You put" is the placer named
// on the counter event (ADR 0056).
//
// No simplification.
const BlasterOracleID = "59c2af34-a2e5-4c67-af2d-a782e41a8e1e"

func init() {
	effects.Register(effects.Spec{
		OracleID:         BlasterOracleID,
		Name:             "Blaster, Combat DJ",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.MoreThanMeetsTheEye("{1}{R}{G}")},
		Static: []game.StaticAbility{{
			Label: "Other nontoken artifact creatures and Vehicles you control have modular 1.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID != source.InstanceID && target.Controller == source.Controller &&
					!target.IsToken() && target.IsArtifact() && (target.IsCreature() || target.HasSubtype("Vehicle"))
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				effects.AppendKeywordsTo(c, []string{"modular 1"})
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCounterPlaced, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Label == game.CounterPlusOne && effects.B11CountersWerePlaced(ev, source.InstanceID, ev.Label, g) &&
					effects.B11CounterPlacerOf(ev, g) == source.Controller
			}, "Blaster, Combat DJ — convert it", effects.Do(effects.ConvertThis{})),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        BlasterOracleID + "#1",
		Name:            "Blaster, Morale Booster",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"modular 3"},
		Activated: []effects.ActivatedAbility{{
			Label:        "{X}, {T}: Move X +1/+1 counters from Blaster onto another target artifact. That artifact gains haste until end of turn. If Blaster has no +1/+1 counters on it, convert it. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{X}"), effects.TapCost()),
			SorcerySpeed: true,
			Targets:      effects.Another(effects.TargetPermanent("another target artifact", effects.Artifact())),
			Effect:       blasterMoveCounters,
		}},
	})
}

func blasterMoveCounters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	x := ctx.X()
	moved := false
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard || t.ID == uuid.Nil {
			continue
		}
		if effects.OnBattlefield(g, item.SourceCardID) && !effects.SourceIsNewObject(g, item) && x > 0 {
			blaster, _ := g.LookupCardForEffect(item.SourceCardID)
			n := min(x, blaster.Counters[game.CounterPlusOne])
			if n > 0 {
				if err := (effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: -n}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: n}).Apply(ctx); err != nil {
					return err
				}
			}
		}
		if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"haste"},
			Label: "Blaster, Morale Booster — gains haste until end of turn"}).Apply(ctx); err != nil {
			return err
		}
		moved = true
		break
	}
	if !moved {
		return nil // the target is gone: the whole ability does nothing (CR 608.2b)
	}
	if now, ok := g.LookupCardForEffect(item.SourceCardID); ok && effects.OnBattlefield(g, item.SourceCardID) && now.Counters[game.CounterPlusOne] == 0 {
		return effects.ConvertThis{}.Apply(ctx)
	}
	return nil
}
