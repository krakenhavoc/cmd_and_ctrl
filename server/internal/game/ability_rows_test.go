package game

import (
	"reflect"
	"testing"
)

// ability_rows_test.go — #2219: the rows an art tile counts in its
// ability chips.

// abilityRowFieldsNotRows is every CardDef field that is NOT a source
// of ability rows, with the reason. Together with the Triggered and
// Activated lists, Static and staticRowSlots, it must name every field
// of CardDef exactly once (TestEveryCardDefSlotIsClassifiedForAbilityRows),
// so a slot added to CardDef is placed rather than silently dropped.
var abilityRowFieldsNotRows = map[string]string{
	"Resolve":                    "the spell's effect as it resolves, not an ability of the object",
	"AsEnters":                   "the CR 614.12 choice as it enters; its answer is on the wire as chosen_color and the rest",
	"AsTransformsInto":           "a transform's as-it-transforms clause, run once",
	"StartingLoyalty":            "a characteristic (loyalty), shown as the loyalty pip",
	"BattleDefense":              "a characteristic (defense), shown on the card",
	"TargetMode":                 "the spell's target hint",
	"Targets":                    "the spell's target clause",
	"Modes":                      "the spell's modes",
	"Purpose":                    "the catalog's declared purpose for the bot (ADR 0126 §6), already the card's own purpose field",
	"ManaAbilities":              "mana abilities: ADR 0105's drop pip and the mana menu are their surface",
	"PrintedKeywords":            "keywords: the keyword chips",
	"ManaTriggers":               "triggered MANA abilities (CR 605.1b): mana, like ManaAbilities",
	"AdditionalCost":             "the spell's own cost",
	"OptionalCosts":              "the spell's own optional costs",
	"AlternativeCosts":           "the spell's own alternative costs",
	"TapCost":                    "the spell's own tap cost (convoke and the like)",
	"Delve":                      "the spell's own delve, a keyword",
	"SpendOnly":                  "the spell's own spend-only clause on X (#2556): a restriction on how its cost is paid, shown by the cast surface",
	"SpendOnlySources":           "the spell's own spend-only-by-source clause (#2556): the same, for the mana's source",
	"SelfCostModifiers":          "what THIS spell costs, shown by the cast surface",
	"CastableZones":              "where the card may be cast from, shown by the cast surface",
	"SpecialActions":             "special actions (CR 116), not abilities: ADR 0105's star pip",
	"OpeningHand":                "an action taken from the opening hand before the game (CR 103.6, ADR 0133), not an ability of the permanent",
	"CastCondition":              "the spell's own cast gate",
	"CastConditionLabel":         "the spell's own cast gate's text",
	"CantBeCountered":            "a property of the spell on the stack, shown there",
	"CantBeCounteredIf":          "a property of the spell on the stack, shown there",
	"SpellDamageCantBePrevented": "a property of the spell's damage",
	"Emblem":                     "an emblem's presentation",
	"TokenText":                  "a token's printed text, already token_text",
	"GrantText":                  "a grant bundle's printed text, already granted_abilities",
	"XMatters":                   "an enumerator hint",
	"WantsDistinctColors":        "an auto-tapper hint",
	"WantsManaFrom":              "an auto-tapper hint",
}

func TestEveryCardDefSlotIsClassifiedForAbilityRows(t *testing.T) {
	placed := map[string]string{"Triggered": "triggered", "Activated": "activated", "Static": "static"}
	for _, s := range staticRowSlots {
		if prev, dup := placed[s.field]; dup {
			t.Errorf("CardDef.%s is placed twice (%s and static slot)", s.field, prev)
		}
		placed[s.field] = "static slot"
	}
	for f := range abilityRowFieldsNotRows {
		if prev, dup := placed[f]; dup {
			t.Errorf("CardDef.%s is both %s and not a row", f, prev)
		}
		placed[f] = "not a row"
	}
	typ := reflect.TypeOf(CardDef{})
	fields := map[string]bool{}
	for i := 0; i < typ.NumField(); i++ {
		name := typ.Field(i).Name
		fields[name] = true
		if _, ok := placed[name]; !ok {
			t.Errorf("CardDef.%s is not classified for the tile's ability rows (#2219): add it to staticRowSlots if it holds a static ability of the object, or to abilityRowFieldsNotRows with the reason it does not", name)
		}
	}
	for f := range placed {
		if !fields[f] {
			t.Errorf("%q is classified for ability rows but is not a CardDef field", f)
		}
	}
}

const (
	arOracle  = "ability-rows-fixture"
	arGated   = "ability-rows-fixture-gated"
	arBundle  = "ability-rows-fixture/bundle"
	arTrigger = "Row Fixture — draw a card"
)

func stubAbilityRowCatalog(t *testing.T) {
	t.Helper()
	defs := map[string]*CardDef{
		arOracle: {
			Triggered: []TriggeredAbility{
				{Key: arTrigger},
				// A keyword implemented as a trigger: on the keyword chips.
				{Key: "Row Fixture — prowess", Keyword: KeywordProwess},
			},
			Static: []StaticAbility{
				{Layer: Layer7PT},
				{Layer: Layer6Ability, Label: "Attacks each combat if able."},
				// The printed-keyword static the effects package
				// synthesises: on the keyword chips.
				{Layer: Layer6Ability, Keywords: []string{"flying"}},
				// The later-layer half of the first row's ability.
				{Layer: Layer7PT, ContinuesAfterRemoval: true},
			},
			CostModifiers: []CostModifier{{Label: "Row Fixture — creature spells you cast cost {1} less to cast."}},
			Replacements:  []ReplacementEffect{{}},
			Activated: []ActivatedAbilityShape{
				{Label: "{T}: Draw a card."},
				{Label: "Exhaust — {4}: Earthbend 4."},
				{Label: "Equip {2}", Equip: true},
				{Label: "Crew 3"},
				{Label: "Basic landcycling {1}{G}", Cycling: true},
				{Label: "Station (Tap another untapped creature you control: …)"},
			},
			ManaAbilities: []ManaAbilityShape{{TapCost: true, Produced: "{G}", Label: "Add {G}"}},
		},
		arGated: {
			Static:    []StaticAbility{{Layer: Layer7PT, ActiveWhen: ClassLevel(2), Label: "Level 2 anthem."}},
			Triggered: []TriggeredAbility{{Key: "Gated — solved", ActiveWhen: CaseSolved()}},
		},
		GrantKey(arBundle): {
			Triggered: []TriggeredAbility{{Key: "Granted Rite — when this creature dies, draw a card"}},
			GrantText: "When this creature dies, draw a card.",
		},
	}
	prev := CatalogLookup
	CatalogLookup = func(key string) *CardDef {
		if d, ok := defs[key]; ok {
			return d
		}
		return nil
	}
	t.Cleanup(func() { CatalogLookup = prev })
}

func rowsByKind(rows []AbilityRow) map[AbilityRowKind][]string {
	out := map[AbilityRowKind][]string{}
	for _, r := range rows {
		out[r.Kind] = append(out[r.Kind], r.Label)
	}
	return out
}

func TestAbilityRowsListNonKeywordRowsByKind(t *testing.T) {
	stubAbilityRowCatalog(t)
	c := Card{Name: "Row Fixture", OracleID: arOracle}
	got := rowsByKind(AbilityRowsOf(c))
	want := map[AbilityRowKind][]string{
		AbilityRowTriggered: {"Draw a card"},
		AbilityRowStatic: {
			"Power/toughness effect",
			"Attacks each combat if able.",
			"Replacement effect",
			"Creature spells you cast cost {1} less to cast.",
		},
		AbilityRowActivated: {"{T}: Draw a card.", "Exhaust — {4}: Earthbend 4."},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("rows = %#v\nwant %#v", got, want)
	}
	// The order on the wire is triggered, static, activated.
	rows := AbilityRowsOf(c)
	if rows[0].Kind != AbilityRowTriggered || rows[len(rows)-1].Kind != AbilityRowActivated {
		t.Errorf("rows out of kind order: %+v", rows)
	}
}

func TestAbilityRowsUncataloguedAndVanillaHaveNone(t *testing.T) {
	stubAbilityRowCatalog(t)
	if rows := AbilityRowsOf(Card{Name: "Grizzly Bears", OracleID: "not-in-the-catalog"}); rows != nil {
		t.Errorf("uncatalogued card rows = %+v, want none", rows)
	}
	if rows := AbilityRowsOf(Card{Name: "Nameless"}); rows != nil {
		t.Errorf("card with no key rows = %+v, want none", rows)
	}
}

// A keyword-only catalog card — its one static is the synthesised
// keyword row — has no rows: the keyword chips are the whole story.
func TestAbilityRowsKeywordOnlyCardHasNone(t *testing.T) {
	prev := CatalogLookup
	CatalogLookup = func(key string) *CardDef {
		if key == "ability-rows-keyword-only" {
			return &CardDef{
				PrintedKeywords: []string{"flying", "vigilance"},
				Static:          []StaticAbility{{Layer: Layer6Ability, Keywords: []string{"flying", "vigilance"}}},
			}
		}
		return nil
	}
	t.Cleanup(func() { CatalogLookup = prev })
	if rows := AbilityRowsOf(Card{Name: "Serra Angel", OracleID: "ability-rows-keyword-only"}); rows != nil {
		t.Errorf("keyword-only card rows = %+v, want none", rows)
	}
}

// Current abilities: a CR 613.1f removal takes the object's own rows
// and keeps a layered grant; a designation gate decides whether a gated
// row exists; a face-down permanent has none (CR 708.2).
func TestAbilityRowsFollowTheLayersAndTheGates(t *testing.T) {
	stubAbilityRowCatalog(t)

	removed := Card{Name: "Row Fixture", OracleID: arOracle}
	removed.effective = &Characteristic{AbilitiesRemoved: true}
	if rows := AbilityRowsOf(removed); rows != nil {
		t.Errorf("a permanent that lost all abilities lists %+v, want none", rows)
	}

	granted := Card{Name: "Bear", OracleID: "not-in-the-catalog"}
	granted.effective = &Characteristic{GrantedAbilities: []GrantedAbility{{Key: GrantKey(arBundle)}}}
	got := AbilityRowsOf(granted)
	if len(got) != 1 || got[0].Kind != AbilityRowTriggered || got[0].Label != "Granted Rite — when this creature dies, draw a card" {
		t.Errorf("a granted trigger lists %+v, want the bundle's row with its grantor's name kept", got)
	}

	removedThenGranted := Card{Name: "Row Fixture", OracleID: arOracle}
	removedThenGranted.effective = &Characteristic{AbilitiesRemoved: true, GrantedAbilities: []GrantedAbility{{Key: GrantKey(arBundle)}}}
	if got := AbilityRowsOf(removedThenGranted); len(got) != 1 || got[0].Kind != AbilityRowTriggered {
		t.Errorf("removal then a later grant lists %+v, want only the grant", got)
	}

	gated := Card{Name: "Gated", OracleID: arGated, TypeLine: "Enchantment — Class Case"}
	if rows := AbilityRowsOf(gated); rows != nil {
		t.Errorf("ungated rows listed: %+v", rows)
	}
	gated.ClassLevel = 2
	gated.Solved = true
	got = AbilityRowsOf(gated)
	if k := rowsByKind(got); len(k[AbilityRowStatic]) != 1 || len(k[AbilityRowTriggered]) != 1 {
		t.Errorf("level-2 solved rows = %+v, want the gated static and trigger", got)
	}

	faceDown := Card{Name: "Row Fixture", OracleID: arOracle, FaceDown: true, FaceDownKind: FaceDownMorphed}
	if rows := AbilityRowsOf(faceDown); rows != nil {
		t.Errorf("a face-down permanent lists %+v, want none (CR 708.2)", rows)
	}
}

func TestAbilityRowLabelDropsTheObjectsName(t *testing.T) {
	names := []string{"Inferno Titan", "Fire // Ice", "Fire", "Ice"}
	for raw, want := range map[string]string{
		"Inferno Titan — 3 damage divided":    "3 damage divided",
		"Inferno Titan: gain twice that much": "Gain twice that much",
		"Inferno Titan":                       "fallback",
		"":                                    "fallback",
		"Fire — 2 damage":                     "2 damage",
		"Other Card — keeps its grantor":      "Other Card — keeps its grantor",
		"{T}: Add {R}.":                       "{T}: Add {R}.",
		"exhaust — {4}: earthbend 4.":         "Exhaust — {4}: earthbend 4.",
		"Inferno Titanic — not the same name": "Inferno Titanic — not the same name",
	} {
		if got := abilityRowLabel(raw, names, "fallback"); got != want {
			t.Errorf("abilityRowLabel(%q) = %q, want %q", raw, got, want)
		}
	}
}

func TestIsKeywordActivatedAbility(t *testing.T) {
	for label, want := range map[string]bool{
		"Equip {2}":                            true,
		"Equip Halfling {1}":                   true,
		"Crew 3":                               true,
		"Mountaincycling {2}":                  true,
		"Station (Tap another untapped …)":     true,
		"Ninjutsu {1}{U}":                      true,
		"Commander ninjutsu {2}{U}{B}":         true,
		"Level up {2}":                         true,
		"Reinforce 2—{2}{W}":                   true,
		"{T}: Draw a card.":                    false,
		"Exhaust — {4}: Earthbend 4.":          false,
		"Channel — {1}{G}, Discard this card:": false,
		"+1: Draw a card.":                     false,
		"Sacrifice a creature: Scry 1.":        false,
		"Crewmate's {T}: nonsense":             false,
	} {
		if got := IsKeywordActivatedAbility(ActivatedAbilityShape{Label: label}); got != want {
			t.Errorf("IsKeywordActivatedAbility(%q) = %v, want %v", label, got, want)
		}
	}
}
