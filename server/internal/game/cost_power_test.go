package game

import "testing"

// cost_power_test.go — ADR 0186 Decision 4 (#3549, seam #2705). The
// arithmetic of Card.CountedPower and the layer-dependency probe's view
// of Characteristic.CostPower. The cards that write the rule, and the
// payment that reads it, are tested in cards/effects, legal and protocol.

// countedFixture is a creature whose effective characteristic carries
// the given rules, with its printed power and toughness.
func countedFixture(power, toughness int, rules ...CostPowerRule) Card {
	c := Card{TypeLine: "Creature — Test", Power: power, Toughness: toughness}
	eff := Characteristic{Power: power, Toughness: toughness, CostPower: rules}
	c.effective = &eff
	return c
}

func TestCountedPowerArithmetic(t *testing.T) {
	both := CostPowerCrew | CostPowerSaddle
	tests := []struct {
		name  string
		card  Card
		use   CostPowerUse
		want  int
		power int // CurrentPower, which no rule may change
	}{
		{"no rule counts plain power", countedFixture(2, 2), CostPowerCrew, 2, 2},
		{"a +2 rule counts power plus two", countedFixture(1, 1, CostPowerRule{Uses: both, Bonus: 2}), CostPowerCrew, 3, 1},
		{"the same rule applies to saddle when it names it", countedFixture(1, 1, CostPowerRule{Uses: both, Bonus: 2}), CostPowerSaddle, 3, 1},
		{"a crew-only rule does not help a saddle", countedFixture(1, 1, CostPowerRule{Uses: CostPowerCrew, Bonus: 2}), CostPowerSaddle, 1, 1},
		{"a saddle-only rule does not help a crew", countedFixture(1, 1, CostPowerRule{Uses: CostPowerSaddle, Bonus: 2}), CostPowerCrew, 1, 1},
		{"toughness replaces power", countedFixture(0, 4, CostPowerRule{Uses: CostPowerCrew, UseToughness: true}), CostPowerCrew, 4, 0},
		{"toughness does not replace power for the other cost", countedFixture(0, 4, CostPowerRule{Uses: CostPowerCrew, UseToughness: true}), CostPowerSaddle, 0, 0},
		{"a negative power plus a bonus adds before it clamps", countedFixture(-1, 1, CostPowerRule{Uses: both, Bonus: 2}), CostPowerCrew, 1, 0},
		{"a negative power with no bonus clamps to zero", countedFixture(-3, 1), CostPowerCrew, 0, 0},
		{"two rules add up", countedFixture(1, 1, CostPowerRule{Uses: both, Bonus: 2}, CostPowerRule{Uses: both, Bonus: 2}), CostPowerCrew, 5, 1},
		{"use zero is plain power whatever the rules say", countedFixture(1, 1, CostPowerRule{Uses: both, Bonus: 2}), 0, 1, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.card.CountedPower(tt.use); got != tt.want {
				t.Errorf("CountedPower = %d, want %d", got, tt.want)
			}
			if got := tt.card.CurrentPower(); got != tt.power {
				t.Errorf("CurrentPower = %d, want %d (a rule never changes the creature's power)", got, tt.power)
			}
		})
	}
}

// P/T counters are part of the power a bonus is added to (CR 122.1a).
func TestCountedPowerStartsFromPowerWithCounters(t *testing.T) {
	c := countedFixture(1, 1, CostPowerRule{Uses: CostPowerCrew, Bonus: 2})
	c.Counters = map[string]int{CounterPlusOne: 2}
	if got := c.CountedPower(CostPowerCrew); got != 5 {
		t.Errorf("CountedPower = %d, want 1 + 2 counters + 2 bonus = 5", got)
	}
}

// The CR 613.8 dependency probe decides whether two layer passes agree
// by comparing their results; two passes that differ only in CostPower
// must read as different, or an ordering that changes a creature's crew
// value would be judged to change nothing.
func TestSameCharacteristicSeesCostPower(t *testing.T) {
	none := Characteristic{Power: 1, Toughness: 1}
	plus2 := Characteristic{Power: 1, Toughness: 1, CostPower: []CostPowerRule{{Uses: CostPowerCrew, Bonus: 2}}}
	plus2Again := Characteristic{Power: 1, Toughness: 1, CostPower: []CostPowerRule{{Uses: CostPowerCrew, Bonus: 2}}}
	plus3 := Characteristic{Power: 1, Toughness: 1, CostPower: []CostPowerRule{{Uses: CostPowerCrew, Bonus: 3}}}
	saddle2 := Characteristic{Power: 1, Toughness: 1, CostPower: []CostPowerRule{{Uses: CostPowerSaddle, Bonus: 2}}}
	tough := Characteristic{Power: 1, Toughness: 1, CostPower: []CostPowerRule{{Uses: CostPowerCrew, UseToughness: true}}}

	if !sameCharacteristic(plus2, plus2Again) {
		t.Error("two passes with equal CostPower read as different")
	}
	for name, other := range map[string]Characteristic{"none": none, "a different bonus": plus3, "a different use": saddle2, "toughness": tough} {
		if sameCharacteristic(plus2, other) {
			t.Errorf("a pass with +2 crew reads as the same as one with %s", name)
		}
	}
}

// clone() must copy the slice: the dependency probe applies a trial pass
// to a clone, and an append through a shared backing array would leave a
// fingerprint on the real pass.
func TestCharacteristicCloneCopiesCostPower(t *testing.T) {
	orig := Characteristic{CostPower: make([]CostPowerRule, 1, 4)}
	orig.CostPower[0] = CostPowerRule{Uses: CostPowerCrew, Bonus: 2}
	cl := orig.clone()
	cl.CostPower[0].Bonus = 9
	cl.CostPower = append(cl.CostPower, CostPowerRule{Uses: CostPowerCrew, Bonus: 1})
	if orig.CostPower[0].Bonus != 2 || len(orig.CostPower) != 1 {
		t.Errorf("a change to the clone reached the original: %+v", orig.CostPower)
	}
}
