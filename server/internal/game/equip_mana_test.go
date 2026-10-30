package game

import "testing"

// equip_mana_test.go — #4600, the engine half of Ronin, Shadow Stalker's
// "Spend this mana only to cast Equipment spells or activate equip
// abilities" (CR 106.6, CR 702.6). The card is in
// internal/cards/effects/ronin_shadow_stalker_test.go.

// roninRestriction is the restriction as Ronin prints it.
var roninRestriction = ManaRestrictAnyOf(
	[]string{ManaRestrictCast, ManaRestrictSubtype("Equipment")},
	[]string{ManaRestrictEquip},
)

// The purpose tag admits an activation of an equip ability and nothing
// else.
func TestManaRestrictEquipAdmitsOnlyAnEquipActivation(t *testing.T) {
	tags := []string{ManaRestrictEquip}
	for _, tc := range []struct {
		name string
		ctx  ManaSpendContext
		want bool
	}{
		{"an equip activation", ManaSpendContext{Purpose: SpendPurposeActivate, Equip: true}, true},
		{"an ordinary activation", ManaSpendContext{Purpose: SpendPurposeActivate}, false},
		{"an Equipment's other activation", ManaSpendContext{Purpose: SpendPurposeActivate, Subtypes: []string{"Equipment"}}, false},
		{"a power-up activation", ManaSpendContext{Purpose: SpendPurposeActivate, PowerUp: true}, false},
		{"a spell", ManaSpendContext{Purpose: SpendPurposeCast}, false},
		{"a spell that wrongly claims the bit", ManaSpendContext{Purpose: SpendPurposeCast, Equip: true}, false},
		{"an unlock", ManaSpendContext{Purpose: SpendPurposeUnlock, Equip: true}, false},
		{"a payment that says nothing", ManaSpendContext{}, false},
		{"a payment that says nothing but the bit", ManaSpendContext{Equip: true}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.ctx.allows(tags); got != tc.want {
				t.Errorf("allows = %v, want %v", got, tc.want)
			}
		})
	}
}

// Ronin's whole clause: an Equipment spell or an equip ability, nothing
// else.
func TestRoninRestrictionAdmitsEquipmentSpellsAndEquipAbilities(t *testing.T) {
	tags := []string{roninRestriction}
	for _, tc := range []struct {
		name string
		ctx  ManaSpendContext
		want bool
	}{
		{"an Equipment spell", ManaSpendContext{Purpose: SpendPurposeCast, Types: []string{"Artifact"}, Subtypes: []string{"Equipment"}}, true},
		{"a creature spell", ManaSpendContext{Purpose: SpendPurposeCast, Types: []string{"Creature"}, Subtypes: []string{"Bear"}}, false},
		{"a non-Equipment artifact spell", ManaSpendContext{Purpose: SpendPurposeCast, Types: []string{"Artifact"}}, false},
		{"an equip ability", ManaSpendContext{Purpose: SpendPurposeActivate, Equip: true}, true},
		{"an Equipment's other ability", ManaSpendContext{Purpose: SpendPurposeActivate, Types: []string{"Artifact"}, Subtypes: []string{"Equipment"}}, false},
		{"an ability that is not equip", ManaSpendContext{Purpose: SpendPurposeActivate}, false},
		{"a payment that says nothing", ManaSpendContext{}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.ctx.allows(tags); got != tc.want {
				t.Errorf("allows = %v, want %v", got, tc.want)
			}
		})
	}
}

// The pool solver spends the restricted mana through the same matcher.
func TestEquipOnlyManaIsSpentThroughTheSpendContext(t *testing.T) {
	cost, err := ParseCost("{1}{R}")
	if err != nil {
		t.Fatal(err)
	}
	pool := ManaPool{
		{Color: "R", Restrictions: []string{roninRestriction}},
		{Color: "R", Restrictions: []string{roninRestriction}},
	}
	if !pool.CanPayFor(cost, 0, ManaSpendContext{Purpose: SpendPurposeActivate, Equip: true}) {
		t.Error("the mana does not pay an equip activation")
	}
	if pool.CanPayFor(cost, 0, ManaSpendContext{Purpose: SpendPurposeActivate}) {
		t.Error("the mana pays an ordinary activation")
	}
	if pool.CanPayFor(cost, 0, ManaSpendContext{}) {
		t.Error("the mana pays a payment that names no purpose")
	}
	if !pool.CanPayFor(cost, 0, ManaSpendForCast(Card{Name: "Test Blade", TypeLine: "Artifact — Equipment"})) {
		t.Error("the mana does not pay an Equipment spell")
	}
	if pool.CanPayFor(cost, 0, ManaSpendForCast(Card{Name: "Test Bear", TypeLine: "Creature — Bear"})) {
		t.Error("the mana pays a creature spell")
	}
}
