package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dodgy Jalopy — Artifact — Vehicle {2}{G}, */5 (Streets of New Capenna):
//
//	"Trample
//	 Dodgy Jalopy's power is equal to the greatest mana value among
//	 creatures you control.
//	 Crew 3
//	 Scavenge {2}{G} ({2}{G}, Exile this card from your graveyard: Put a
//	 number of +1/+1 counters equal to this card's power on target
//	 creature. Scavenge only as a sorcery.)"
//
// The power is a characteristic-defining ability (CR 604.3), a layer 7a
// self-only static. It functions in every zone (CR 113.6a, 208.2a), which
// is what Scavenge needs: the card is in the graveyard when it is activated
// and in exile when it resolves, and "this card's power" is then the
// greatest mana value among the creatures its owner controls at that
// moment. scavengeCounters reads the power through
// Game.PowerAnywhereForEffect, which runs the card's own layer 7a
// abilities where no layer pass does. A Vehicle is not a creature until
// crewed, so with no creature the power is 0 and Scavenge adds no counters
// (CR 107.1b); a crewed Jalopy counts its own mana value, 3.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "469d820d-b7a6-414a-af18-71c1b1e42da5",
		Name:            "Dodgy Jalopy",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{{
			Label:    "Power is equal to the greatest mana value among creatures you control.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power = greatestManaValueAmongCreaturesOf(g, source)
			},
		}},
		Activated: []ActivatedAbility{
			{
				Label:   "Crew 3",
				Cost:    CrewCost(3),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  CrewEffect("Dodgy Jalopy"),
			},
			Scavenge("{2}{G}"),
		},
	})
}

// greatestManaValueAmongCreaturesOf is the greatest mana value among the
// creatures `source`'s controller controls: its controller on the
// battlefield, its owner anywhere else (a card in a graveyard or exile has
// no controller, CR 108.3).
func greatestManaValueAmongCreaturesOf(g *game.Game, source *game.Card) int {
	if g == nil || source == nil {
		return 0
	}
	player := source.Owner
	if z := g.FindCardZoneForEffect(source.InstanceID); z != nil && z.Kind == game.ZoneBattlefield {
		player = source.Controller
	}
	if player == uuid.Nil {
		return 0
	}
	greatest := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() {
			if mv := c.ManaValue(); mv > greatest {
				greatest = mv
			}
		}
	}
	return greatest
}
