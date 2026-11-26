package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82OldB503CycledOrDies is one triggered ability with two conditions: "When
// you cycle this card and when this creature dies" (the Alara Reborn
// Sojourners). It watches from the graveyard, where the cycled card already is
// (CR 702.29c) and where the creature has just arrived (CR 700.4), and so
// fires once for either event.
func S82OldB503CycledOrDies(label string, effect Effect) game.TriggeredAbility {
	t := OnAny([]game.EventKind{game.EventCycle, game.EventLTB},
		func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
			if ev.Kind == game.EventCycle {
				return Self(ev, source, lki, g)
			}
			return ThisDied(ev, source, lki, g)
		}, label, effect)
	t.Zones = []game.ZoneKind{game.ZoneGraveyard}
	return t
}

// PrintedOldB503Dragon is Dragon Broodmother's token: a 1/1 red and green
// Dragon creature token with flying and devour 2 (CR 702.82a: as it enters, its
// controller may sacrifice any number of creatures, and it enters with twice
// that many +1/+1 counters). Devour is a replacement effect, so it lives on the
// template's replacement slot.
func PrintedOldB503Dragon() TokenTemplate {
	return TokenTemplate{
		Slug: "dragon-devour",
		Card: game.Card{
			Name:      "Dragon",
			TypeLine:  "Token Creature — Dragon",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"R", "G"},
			Keywords:  []string{"flying"},
		},
		Replacements: []game.ReplacementEffect{Devour("Dragon", 2)},
		Text:         "Flying\nDevour 2 (As this creature enters, you may sacrifice any number of creatures. It enters with twice that many +1/+1 counters on it.)",
	}
}

// S82OldB503DragonToken is a fresh copy of the Dragon token.
func S82OldB503DragonToken() game.Card { return TokenFromCatalog(PrintedOldB503Dragon) }

// S82OldB503Borderpost is the Alara Reborn Borderpost cycle:
//
//	"You may pay {1} and return a basic land you control to its owner's hand
//	 rather than pay this spell's mana cost.
//	 This artifact enters tapped.
//	 {T}: Add {A} or {B}."
//
// The alternative cost is {1} plus returning a basic land you control (CR 118.9),
// chosen as the spell is cast; the artifact enters tapped (a replacement on
// itself) and taps for either colour.
func S82OldB503Borderpost(oracleID, name, cost, a, b string) Spec {
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		AlternativeCosts: []game.AlternativeCost{{
			Key:      "borderpost",
			Label:    "Pay {1} and return a basic land you control to its owner's hand",
			ManaCost: "{1}",
			ReturnToHand: PermanentYouControl("a basic land you control", Land(),
				func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return IsBasicLand(c) }),
			PayLabel: "a basic land you control",
		}},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{" + a + "|" + b + "}",
			Label:    "{T}: Add {" + a + "} or {" + b + "}.",
		}},
	}
}
