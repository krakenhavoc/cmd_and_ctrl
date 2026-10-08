package decks

import (
	"sort"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// purpose_test.go — ADR 0126 §6: the curated decks' cards declare what
// they do.
//
// The heuristic bot reads the view and nothing else (ADR 0033 §3), and
// the view cannot tell Wrath of God from Divination: both are
// untargeted sorceries with a mana value. A card's declared Purpose is
// what tells them apart, so a curated instant or sorcery that declares
// none is priced as a blank. TestCuratedDeckPurposes holds every one of
// them to a declaration, or to a named reason it has none.
//
// The decks package has no type lines offline (the dump is not in the
// repo), so the instants and sorceries are listed by hand below, and
// realdump_purpose_manual_test.go checks the list against Scryfall.

// curatedInstantsAndSorceries is every instant and sorcery in the four
// curated decks, by name. A card that is in two decks is listed once.
var curatedInstantsAndSorceries = []string{
	// esper-control
	"Counterspell", "Negate", "An Offer You Can't Refuse", "Mana Drain", "Force of Negation",
	"Fierce Guardianship", "Absorb", "Mystic Confluence", "Swords to Plowshares", "Path to Exile",
	"Murder", "Hero's Downfall", "Go for the Throat", "Infernal Grasp", "Feed the Swarm",
	"Withering Torment", "Anguished Unmaking", "Mortify", "Despark", "Utter End", "Void Rend",
	"Generous Gift", "Stroke of Midnight", "Wrath of God", "Damnation", "Day of Judgment", "Damn",
	"Austere Command", "Farewell", "Cyclonic Rift", "Aetherize", "Read the Bones",
	"Ambition's Cost", "Divination", "Pull from Tomorrow", "Fact or Fiction",
	"Sphinx's Revelation", "Treasure Cruise", "Dig Through Time", "Diabolic Tutor",
	// izzet-aggro
	"Lightning Bolt", "Shock", "Abrade", "Arc Trail", "Blaze", "Izzet Charm", "Fiery Temper",
	"Prismari Command", "Chaos Warp", "Pyroblast", "Swan Song", "Wash Away", "Preordain",
	"Frantic Search", "Faithless Looting", "Big Score", "Unexpected Windfall", "Windfall",
	"Wheel of Fortune", "Vandalblast", "Rapid Hybridization", "Pongify",
	// mono-black-aristocrats
	"Reanimate", "Zombify", "Dread Return", "Stitch Together", "Living Death",
	"Rise of the Dark Realms", "Dread Summons", "Entomb", "Altar's Reap", "Deadly Dispute",
	"Night's Whisper", "Sign in Blood", "Exsanguinate", "Doom Blade", "Ashes to Ashes",
	"Demonic Tutor",
	// simic-ramp
	"Rampant Growth", "Nature's Lore", "Three Visits", "Farseek", "Cultivate", "Kodama's Reach",
	"Explosive Vegetation", "Harrow", "Hour of Promise", "Circuitous Route", "Titania's Command",
	"Overrun", "Return of the Wildspeaker", "Shamanic Revelation", "Eureka Moment", "Harmonize",
	"Beast Within", "Krosan Grip",
}

// valueIsTheirTarget is ADR 0126 §6's short, named list of curated
// spells whose value is what they target: removal, counterspells,
// targeted discard and reanimation. The heuristic prices a target
// already (targetsValue), so a purpose would say nothing it does not
// know. Keyed by name; the value is the class.
var valueIsTheirTarget = map[string]string{
	"Counterspell":              "counterspell",
	"Negate":                    "counterspell",
	"Swan Song":                 "counterspell",
	"An Offer You Can't Refuse": "counterspell",
	"Mana Drain":                "counterspell",
	"Wash Away":                 "counterspell",
	"Force of Negation":         "counterspell",
	"Fierce Guardianship":       "counterspell",
	"Absorb":                    "counterspell",
	"Pyroblast":                 "counterspell or removal, by mode",
	"Swords to Plowshares":      "removal",
	"Path to Exile":             "removal",
	"Doom Blade":                "removal",
	"Go for the Throat":         "removal",
	"Infernal Grasp":            "removal",
	"Feed the Swarm":            "removal",
	"Withering Torment":         "removal",
	"Anguished Unmaking":        "removal",
	"Mortify":                   "removal",
	"Despark":                   "removal",
	"Generous Gift":             "removal",
	"Stroke of Midnight":        "removal",
	"Murder":                    "removal",
	"Hero's Downfall":           "removal",
	"Utter End":                 "removal",
	"Void Rend":                 "removal",
	"Chaos Warp":                "removal",
	"Arc Trail":                 "removal or burn; its two picks are one clause of two targets dealt 2 and 1 by position, which a per-clause entry cannot say (ADR 0126's amendment of 2026-10-08)",
	"Rapid Hybridization":       "removal",
	"Pongify":                   "removal",
	"Beast Within":              "removal",
	"Krosan Grip":               "removal",
	"Ashes to Ashes":            "removal",
	"Reanimate":                 "reanimation of the target",
	"Zombify":                   "reanimation of the target",
	"Dread Return":              "reanimation of the target",
	"Stitch Together":           "reanimation of the target, or its return to hand",
}

// noPrintedAmount is the curated spells that do something in a priced
// class, or near it, but print no fixed amount a Purpose could hold:
// the amount is X or counted at resolution, the effect is symmetric, or
// it is a class ADR 0126 leaves out (rituals, pumps). Each is priced by
// the generic floors.
var noPrintedAmount = map[string]string{
	"Pull from Tomorrow":        "draws X",
	"Sphinx's Revelation":       "draws X",
	"Fact or Fiction":           "an opponent splits five cards; how many reach the hand is theirs to decide",
	"Dread Summons":             "each player mills X; the Zombies are counted at resolution",
	"Windfall":                  "each player discards a hand and draws as many as the largest; no fixed amount",
	"Wheel of Fortune":          "each player discards a hand and draws seven; its value is the hands, which no amount says",
	"Exsanguinate":              "drains X",
	"Blaze":                     "deals X damage to any target",
	"Living Death":              "a symmetric mass reanimation: a sweep purpose would price the sacrifice and miss the return",
	"Rise of the Dark Realms":   "mass reanimation from every graveyard",
	"Overrun":                   "a combat pump: its value is the attack (ADR 0126 Out of scope)",
	"Return of the Wildspeaker": "draws as many as the greatest power, or a pump",
	"Shamanic Revelation":       "draws one per creature, counted at resolution",
}

// curatedTargetPurposes is every curated spell whose value depends on
// what it does to its target (ADR 0126's amendment of 2026-10-08, owner
// answer 5, first half): a gift that is good or bad by whom it is
// aimed at, and burn, which kills or does not by how much it deals.
// Each declares a target entry (Purpose.Targets) on the slot named.
// Arc Trail and Blaze cannot, and say why on valueIsTheirTarget and
// noPrintedAmount.
var curatedTargetPurposes = map[string]string{
	"Prismari Command": "modes",
	"Sign in Blood":    "card",
	"Lightning Bolt":   "card",
	"Shock":            "card",
	"Fiery Temper":     "card",
	"Abrade":           "modes",
	"Izzet Charm":      "modes",
}

// targetEntriesIn counts the target entries a spell declares in `slot`:
// "card" for the spell's own statement, "modes" for its bullets'.
func targetEntriesIn(s effects.Spec, slot string) int {
	switch slot {
	case "card":
		return len(s.Purpose.Targets.List())
	case "modes":
		n := 0
		if s.Modes != nil {
			for _, o := range s.Modes.Options {
				n += len(o.Purpose.Targets.List())
			}
		}
		return n
	}
	return 0
}

// curatedPermanentPurposes is every curated permanent in a class ADR
// 0126 prices (ramp, draw, loot, tutor, wipe, death payoff, discard
// payoff) that
// declares a purpose, with the slot it declares it in. Mana rocks and
// dorks are not here: their mana abilities already say what they make.
var curatedPermanentPurposes = map[string]string{
	// enters effects, on the card
	"Baleful Strix":          "card",
	"Mulldrifter":            "card",
	"Solemn Simulacrum":      "card",
	"Corsair Captain":        "card",
	"Bastion of Remembrance": "card",
	"Vile Entomber":          "card",
	"Woe Strider":            "card",
	// activated rows
	"Mary Read and Anne Bonny": "activated, discard payoff",
	"Loran of the Third Path":  "activated",
	"Glint-Horn Buccaneer":     "activated, discard payoff",
	"Geier Reach Sanitarium":   "activated",
	"Warren Soultrader":        "activated",
	"Sakura-Tribe Elder":       "activated",
	"Burnished Hart":           "activated",
	"Spectral Sailor":          "activated",
	"Teferi, Master of Time":   "activated",
	// death payoffs, on the triggered row
	"Blood Artist":         "death payoff",
	"Midnight Reaper":      "death payoff",
	"Pitiless Plunderer":   "death payoff",
	"Grave Pact":           "death payoff",
	"Dictate of Erebos":    "death payoff",
	"Butcher of Malakir":   "death payoff",
	"Syr Konrad, the Grim": "death payoff",
	"Mirkwood Bats":        "death payoff",
	"Falkenrath Noble":     "death payoff",
	"Vindictive Vampire":   "death payoff",
	// discard payoffs, on the triggered row (ADR 0126's amendment of
	// 2026-10-06)
	"Marauding Mako":       "discard payoff",
	"Scrounging Skyray":    "discard payoff",
	"Magmakin Artillerist": "discard payoff",
}

// specDeclaresPurpose reports whether a spell declares what it does on
// the card, on a mode or on an alternative cost.
func specDeclaresPurpose(s effects.Spec) bool {
	if !s.Purpose.IsZero() {
		return true
	}
	if s.Modes != nil {
		for _, o := range s.Modes.Options {
			if !o.Purpose.IsZero() {
				return true
			}
		}
	}
	for _, a := range s.AlternativeCosts {
		if !a.Purpose.IsZero() {
			return true
		}
	}
	return false
}

// declaresIn reports whether the spec declares a purpose in `slot`. A
// comma-separated list of slots asks for every one of them.
func declaresIn(s effects.Spec, slot string) bool {
	if many := strings.Split(slot, ", "); len(many) > 1 {
		for _, one := range many {
			if !declaresIn(s, one) {
				return false
			}
		}
		return true
	}
	switch slot {
	case "card":
		return !s.Purpose.IsZero()
	case "activated":
		for _, a := range s.Activated {
			if !a.Purpose.IsZero() {
				return true
			}
		}
	case "death payoff":
		for _, t := range s.Triggered {
			if t.Purpose.DeathPayoff {
				return true
			}
		}
	case "discard payoff":
		for _, t := range s.Triggered {
			if t.Purpose.DiscardPayoff != nil {
				return true
			}
		}
	}
	return false
}

func TestCuratedDeckPurposes(t *testing.T) {
	spells := map[string]bool{}
	for _, n := range curatedInstantsAndSorceries {
		if spells[n] {
			t.Errorf("%s is listed twice in curatedInstantsAndSorceries", n)
		}
		spells[n] = true
	}
	for n := range valueIsTheirTarget {
		if _, dup := noPrintedAmount[n]; dup {
			t.Errorf("%s is on both exemption lists", n)
		}
	}

	inDeck := map[string]bool{}
	var missing []string
	for _, d := range All() {
		for _, c := range d.Cards() {
			if c.Basic {
				continue
			}
			inDeck[c.Name] = true
			spec, ok := effects.Lookup(c.OracleID)
			if !ok {
				continue // TestEveryCardResolvesToARegisteredSpec's failure, not this one's
			}
			if spells[c.Name] {
				declared := specDeclaresPurpose(spec)
				_, target := valueIsTheirTarget[c.Name]
				_, noAmount := noPrintedAmount[c.Name]
				switch {
				case declared && (target || noAmount):
					t.Errorf("%s (%s) declares a purpose and is also on an exemption list: take it off the list", c.Name, d.ID)
				case !declared && !target && !noAmount:
					missing = append(missing, c.Name+" ("+d.ID+")")
				}
			}
			if slot, ok := curatedPermanentPurposes[c.Name]; ok && !declaresIn(spec, slot) {
				t.Errorf("%s (%s) declares no purpose on its %s", c.Name, d.ID, slot)
			}
			if slot, ok := curatedTargetPurposes[c.Name]; ok && targetEntriesIn(spec, slot) == 0 {
				t.Errorf("%s (%s) declares no target entry on its %s (ADR 0126's amendment of 2026-10-08)", c.Name, d.ID, slot)
			}
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		t.Errorf("%d curated instant(s) or sorcery(ies) declare no Purpose (ADR 0126 §6): declare one on the "+
			"Spec, its modes or its alternative costs, or name why not in valueIsTheirTarget or noPrintedAmount:\n\t%s",
			len(missing), strings.Join(missing, "\n\t"))
	}

	// No stale entries: every listed card is in a curated deck.
	for _, list := range []struct {
		name  string
		names []string
	}{
		{"curatedInstantsAndSorceries", curatedInstantsAndSorceries},
		{"valueIsTheirTarget", sortedKeys(valueIsTheirTarget)},
		{"noPrintedAmount", sortedKeys(noPrintedAmount)},
		{"curatedPermanentPurposes", sortedKeys(curatedPermanentPurposes)},
		{"curatedTargetPurposes", sortedKeys(curatedTargetPurposes)},
	} {
		for _, n := range list.names {
			if !inDeck[n] {
				t.Errorf("%s lists %q, which is in no curated deck", list.name, n)
			}
		}
	}
	for n := range valueIsTheirTarget {
		if !spells[n] {
			t.Errorf("valueIsTheirTarget lists %q, which is not in curatedInstantsAndSorceries", n)
		}
	}
	for n := range curatedTargetPurposes {
		if _, listed := valueIsTheirTarget[n]; listed {
			t.Errorf("%s declares its target entries and is also on valueIsTheirTarget: take it off", n)
		}
	}
	for n := range noPrintedAmount {
		if !spells[n] {
			t.Errorf("noPrintedAmount lists %q, which is not in curatedInstantsAndSorceries", n)
		}
	}
}

func sortedKeys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
