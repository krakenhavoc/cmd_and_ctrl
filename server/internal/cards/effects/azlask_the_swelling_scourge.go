package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Azlask, the Swelling Scourge — Legendary Creature — Eldrazi {3}, 2/2:
//
//	"Whenever Azlask or another colorless creature you control dies, you
//	 get an experience counter.
//	 {W}{U}{B}{R}{G}: Creatures you control get +X/+X until end of turn,
//	 where X is the number of experience counters you have. Scions and
//	 Spawns you control gain indestructible and annihilator 1 until end
//	 of turn."
//
// The dies trigger reads the creature's colours as it last existed on
// the battlefield (CR 603.10a), so a creature painted a colour by an
// effect does not count and one made colourless does. Azlask itself
// counts whatever its colour. The experience counter is a player
// counter placed by you.
//
// The activation reads X as it resolves (the 2024-06-07 ruling) and
// locks both sets then (CR 611.2c): the creatures you control get
// +X/+X, and the Scions and Spawns you control gain indestructible and
// annihilator 1 — the keyword the engine turns into an attack trigger
// (ADR 0113 §2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "55bc7f55-f73b-40b5-8912-9bf76c129ccc",
		Name:         "Azlask, the Swelling Scourge",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, azlaskOrAnotherColorlessCreatureYouControlDied,
				"Azlask, the Swelling Scourge — you get an experience counter", youGetAnExperienceCounter),
		},
		Activated: []ActivatedAbility{{
			Label: "{W}{U}{B}{R}{G}: Creatures you control get +X/+X until end of turn, where X is the number of experience counters you have. Scions and Spawns you control gain indestructible and annihilator 1 until end of turn.",
			Cost:  ManaCost("{W}{U}{B}{R}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				x := experienceCounters(g, item.Controller)
				if err := (BoostUntilEOT{
					Match: And(Creature(), YouControl()), Power: x, Toughness: x,
					Label: "Azlask, the Swelling Scourge — +X/+X until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{
					Match:    And(Or(HasSubtype("Scion"), HasSubtype("Spawn")), YouControl()),
					Keywords: []string{"indestructible", "annihilator 1"},
					Label:    "Azlask, the Swelling Scourge — indestructible and annihilator 1 until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}

// azlaskOrAnotherColorlessCreatureYouControlDied is "Whenever ~ or
// another colorless creature you control dies".
func azlaskOrAnotherColorlessCreatureYouControlDied(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	if ThisDied(ev, source, lki, g) {
		return true
	}
	dead, ok := diedCreature(ev, g)
	if !ok || leftUnderControlOf(ev, dead) != source.Controller {
		return false
	}
	return leftColorless(ev, dead)
}

// leftColorless reports whether the permanent an EventLTB names was
// colourless as it last existed on the battlefield (CR 603.10a), and
// falls back to the card as it now sits for an event with no
// last-known information.
func leftColorless(ev game.Event, c game.Card) bool {
	for _, color := range []string{"W", "U", "B", "R", "G"} {
		if leftAsColor(ev, c, color) {
			return false
		}
	}
	return true
}
