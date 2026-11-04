package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Romulan Bird-of-Prey — Artifact — Spacecraft {1}{U}, 4/4 (Star Trek,
// #4973):
//
//	"Cloaking Device — As long as this Spacecraft isn't a creature, it has
//	 hexproof.
//	 Whenever you cast a noncreature spell, if this Spacecraft isn't a
//	 creature, put two charge counters on it.
//	 Station (Tap another creature you control: Put charge counters equal
//	 to its power on this Spacecraft. Station only as a sorcery. It's an
//	 artifact creature at 8+.)
//	 8+ | Flying"
//
// Station is the ordinary keyword (station.go): at eight charge counters
// it is a 4/4 artifact creature with flying. Cloaking Device is a layer-6
// grant that reads the type layer 4 has already settled, so the hexproof
// goes away the moment the Spacecraft becomes a creature. The cast
// trigger's "if" is an intervening if (CR 603.4): checked as the spell is
// cast and again as the trigger resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e614f97b-da8e-4569-9bdd-f76371c66836",
		Name:         "Romulan Bird-of-Prey",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Station()},
		Static: append(append(
			SpacecraftAt(8, 4, 4),
			ThresholdKeywords(8, "flying")),
			romulanBirdOfPreyCloakingDevice()),
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, romulanBirdOfPreyCastWhileCloaked,
				"Romulan Bird-of-Prey — put two charge counters on it",
				romulanBirdOfPreyCharge),
		},
	})
}

// romulanBirdOfPreyCloakingDevice is "As long as this Spacecraft isn't a
// creature, it has hexproof."
func romulanBirdOfPreyCloakingDevice() game.StaticAbility {
	return game.StaticAbility{
		Label:     "Cloaking Device — As long as this Spacecraft isn't a creature, it has hexproof.",
		Layer:     game.Layer6Ability,
		AppliesTo: selfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			for _, t := range c.Types {
				if t == "Creature" {
					return
				}
			}
			appendKeywordsTo(c, []string{"hexproof"})
		},
	}
}

// romulanBirdOfPreyCastWhileCloaked is "Whenever you cast a noncreature
// spell, if this Spacecraft isn't a creature".
func romulanBirdOfPreyCastWhileCloaked(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	return YouCast(Noncreature())(ev, source, lki, g) && !source.IsCreature()
}

// romulanBirdOfPreyCharge re-checks the "if" as the trigger resolves
// (CR 603.4) and puts the counters on this permanent while it is still
// the object that triggered.
func romulanBirdOfPreyCharge(g *game.Game, item *game.StackItem) error {
	if !sourceIsStillThisPermanent(g, item) {
		return nil
	}
	c, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || c.IsCreature() {
		return nil
	}
	return AddCounter{Target: item.SourceCardID, Kind: game.CounterCharge, N: 2}.Apply(NewContext(g, item))
}
