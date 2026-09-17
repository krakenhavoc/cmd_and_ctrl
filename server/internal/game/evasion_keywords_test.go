package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// evasionFixture makes a current-characteristic creature for the compact
// pair table. Layered integration is covered below through real statics.
func evasionFixture(name string, power int, colors []string, artifact bool, keywords ...string) *Card {
	c := &Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Test",
		Power: power, Toughness: 2, Colors: append([]string(nil), colors...),
	}
	e := c.printedCharacteristic()
	e.Abilities = append(e.Abilities, keywords...)
	if artifact {
		e.Types = append(e.Types, "Artifact")
	}
	c.effective = &e
	return c
}

func TestEvasionKeywordsBlockPairTable(t *testing.T) {
	g := newActiveGame(t)
	for _, tc := range []struct {
		name              string
		attacker, blocker *Card
		want              BlockReason
		wantSourceBlocker bool
	}{
		{"fear accepts black", evasionFixture("Fear", 2, nil, false, "fear"), evasionFixture("Black", 2, []string{"B"}, false), "", false},
		{"fear accepts colorless artifact", evasionFixture("Fear", 2, nil, false, "fear"), evasionFixture("Artifact", 2, nil, true), "", false},
		{"fear refuses other creature", evasionFixture("Fear", 2, nil, false, "fear"), evasionFixture("Green", 2, []string{"G"}, false), BlockReasonFear, false},
		{"intimidate accepts shared effective color", evasionFixture("Red", 2, []string{"R"}, false, "intimidate"), evasionFixture("Other Red", 2, []string{"R"}, false), "", false},
		{"intimidate accepts artifact", evasionFixture("Red", 2, []string{"R"}, false, "intimidate"), evasionFixture("Artifact", 2, nil, true), "", false},
		{"intimidate refuses a different color", evasionFixture("Red", 2, []string{"R"}, false, "intimidate"), evasionFixture("Blue", 2, []string{"U"}, false), BlockReasonIntimidate, false},
		{"intimidate colorless does not share color", evasionFixture("Colorless", 2, nil, false, "intimidate"), evasionFixture("Other Colorless", 2, nil, false), BlockReasonIntimidate, false},
		{"shadow attacker requires shadow blocker", evasionFixture("Shadow", 2, nil, false, "shadow"), evasionFixture("Ground", 2, nil, false), BlockReasonShadow, false},
		{"shadow blocker cannot block ordinary attacker", evasionFixture("Ground", 2, nil, false), evasionFixture("Shadow", 2, nil, false, "shadow"), BlockReasonShadow, true},
		{"shadow matches shadow", evasionFixture("Shadow", 2, nil, false, "shadow"), evasionFixture("Shadow", 2, nil, false, "shadow"), "", false},
		{"horsemanship requires horsemanship", evasionFixture("Horse", 2, nil, false, "horsemanship"), evasionFixture("Ground", 2, nil, false), BlockReasonHorsemanship, false},
		{"horsemanship matches horsemanship", evasionFixture("Horse", 2, nil, false, "horsemanship"), evasionFixture("Horse", 2, nil, false, "horsemanship"), "", false},
		{"skulk refuses greater power", evasionFixture("Skulk", 2, nil, false, "skulk"), evasionFixture("Big", 3, nil, false), BlockReasonSkulk, false},
		{"skulk allows equal power", evasionFixture("Skulk", 2, nil, false, "skulk"), evasionFixture("Equal", 2, nil, false), "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := g.BlockPairRefusalLocked(tc.attacker, tc.blocker)
			if r.Reason != tc.want {
				t.Fatalf("reason = %q, want %q", r.Reason, tc.want)
			}
			if !r.Legal() {
				wantSource := tc.attacker.InstanceID
				if tc.wantSourceBlocker {
					wantSource = tc.blocker.InstanceID
				}
				if r.Source != wantSource {
					t.Errorf("source = %s, want %s", r.Source, wantSource)
				}
				if got := g.blockRefusedErrorLocked(tc.attacker, tc.blocker, r).Sentence(uuid.Nil); got == "" {
					t.Error("refusal has no server-built sentence")
				}
			}
		})
	}
}

func TestEvasionKeywordsAreCanonical(t *testing.T) {
	for _, keyword := range []string{"fear", "intimidate", "shadow", "horsemanship", "skulk"} {
		if got, ok := CanonicalKeyword(" " + keyword + " "); !ok || got != keyword {
			t.Errorf("CanonicalKeyword(%q) = (%q, %v)", keyword, got, ok)
		}
	}
}

func TestEvasionKeywordsRefuseRealBlockDeclarations(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushKeywordCreature(t, g, g.Seats[0], 2, 2, "skulk")
	blocker := pushKeywordCreature(t, g, g.Seats[1], 3, 3)
	attackAndStepToBlocks(t, g, g.Seats[1].ID, attacker)
	beforeEvents := len(g.Events)
	err := g.DeclareBlocker(blocker, attacker)
	if !errors.Is(err, ErrIllegalBlock) {
		t.Fatalf("DeclareBlocker: got %v, want ErrIllegalBlock", err)
	}
	var refused *BlockRefusedError
	if !errors.As(err, &refused) || refused.Reason != BlockReasonSkulk || refused.Keyword != "skulk" {
		t.Fatalf("refused error = %#v", err)
	}
	if got := findCard(g, blocker).BlockingTarget; got != uuid.Nil {
		t.Errorf("refused block was stored against %s", got)
	}
	for _, ev := range g.Events[beforeEvents:] {
		if ev.Kind == EventBlock {
			t.Error("refused block emitted EventBlock")
		}
	}
}

func TestEvasionKeywordsReadLayeredCharacteristics(t *testing.T) {
	t.Run("grant then ability removal", func(t *testing.T) {
		g := newActiveGame(t)
		const (
			grant = "evasion-fear-grant"
			strip = "evasion-ability-removal"
		)
		withStaticAbilities(t, func(key string) []StaticAbility {
			switch key {
			case grant:
				return []StaticAbility{{
					Layer:     Layer6Ability,
					AppliesTo: func(target *Card, _ *Game, _ *Card) bool { return target.Name == "Attacker" },
					Apply:     func(c *Characteristic, _ *Card, _ *Game, _ *Card) { c.Abilities = append(c.Abilities, "fear") },
				}}
			case strip:
				return []StaticAbility{{
					Layer: Layer6Ability, RemovesAbilities: true,
					AppliesTo: func(target *Card, _ *Game, _ *Card) bool { return target.Name == "Attacker" },
				}}
			}
			return nil
		})
		me, them := g.Seats[0].ID, g.Seats[1].ID
		attacker := pushTypedTestCard(g, Card{Name: "Attacker", TypeLine: "Creature — Test", Power: 2, Toughness: 2, Owner: me, Controller: me})
		blocker := pushTypedTestCard(g, Card{Name: "Blocker", TypeLine: "Creature — Test", Power: 2, Toughness: 2, Colors: []string{"G"}, Owner: them, Controller: them})
		if r := pairRefusal(g, attacker, blocker); !r.Legal() {
			t.Fatalf("ungranted attacker refused with %q", r.Reason)
		}
		pushTypedTestCard(g, Card{Name: "Grant", TypeLine: "Enchantment", OracleID: grant, Owner: me, Controller: me})
		if r := pairRefusal(g, attacker, blocker); r.Reason != BlockReasonFear {
			t.Fatalf("granted fear refusal = %q, want fear", r.Reason)
		}
		pushTypedTestCard(g, Card{Name: "Strip", TypeLine: "Enchantment", OracleID: strip, Owner: me, Controller: me})
		if r := pairRefusal(g, attacker, blocker); !r.Legal() {
			t.Fatalf("removed fear still refused with %q", r.Reason)
		}
	})

	t.Run("intimidate reads effective color and type", func(t *testing.T) {
		for _, tc := range []struct {
			name, source string
			ability      StaticAbility
		}{
			{
				name: "color", source: "evasion-red-color",
				ability: StaticAbility{Layer: Layer5Color,
					AppliesTo: func(target *Card, _ *Game, _ *Card) bool { return target.Name == "Blocker" },
					Apply:     func(c *Characteristic, _ *Card, _ *Game, _ *Card) { c.Colors = []string{"R"} }},
			},
			{
				name: "artifact type", source: "evasion-artifact-type",
				ability: StaticAbility{Layer: Layer4Type,
					AppliesTo: func(target *Card, _ *Game, _ *Card) bool { return target.Name == "Blocker" },
					Apply:     func(c *Characteristic, _ *Card, _ *Game, _ *Card) { c.Types = append(c.Types, "Artifact") }},
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				g := newActiveGame(t)
				withStaticAbilities(t, func(key string) []StaticAbility {
					if key == tc.source {
						return []StaticAbility{tc.ability}
					}
					return nil
				})
				me, them := g.Seats[0].ID, g.Seats[1].ID
				attacker := pushTypedTestCard(g, Card{Name: "Attacker", TypeLine: "Creature — Test", Power: 2, Toughness: 2, Colors: []string{"R"}, Keywords: []string{"intimidate"}, Owner: me, Controller: me})
				blocker := pushTypedTestCard(g, Card{Name: "Blocker", TypeLine: "Creature — Test", Power: 2, Toughness: 2, Colors: []string{"U"}, Owner: them, Controller: them})
				if r := pairRefusal(g, attacker, blocker); r.Reason != BlockReasonIntimidate {
					t.Fatalf("printed state refusal = %q, want intimidate", r.Reason)
				}
				pushTypedTestCard(g, Card{Name: "Layer source", TypeLine: "Enchantment", OracleID: tc.source, Owner: me, Controller: me})
				if r := pairRefusal(g, attacker, blocker); !r.Legal() {
					t.Fatalf("effective %s did not permit block: %q", tc.name, r.Reason)
				}
			})
		}
	})

	t.Run("skulk reads current power", func(t *testing.T) {
		g := newActiveGame(t)
		const source = "evasion-power-reduction"
		withStaticAbilities(t, func(key string) []StaticAbility {
			if key != source {
				return nil
			}
			return []StaticAbility{{
				Layer: Layer7PT, SubLayer: SubLayer7C_Modify,
				AppliesTo: func(target *Card, _ *Game, _ *Card) bool { return target.Name == "Blocker" },
				Apply:     func(c *Characteristic, _ *Card, _ *Game, _ *Card) { c.Power -= 2 },
			}}
		})
		me, them := g.Seats[0].ID, g.Seats[1].ID
		attacker := pushTypedTestCard(g, Card{Name: "Attacker", TypeLine: "Creature — Test", Power: 2, Toughness: 2, Keywords: []string{"skulk"}, Owner: me, Controller: me})
		blocker := pushTypedTestCard(g, Card{Name: "Blocker", TypeLine: "Creature — Test", Power: 3, Toughness: 3, Owner: them, Controller: them})
		if r := pairRefusal(g, attacker, blocker); r.Reason != BlockReasonSkulk {
			t.Fatalf("printed power refusal = %q, want skulk", r.Reason)
		}
		pushTypedTestCard(g, Card{Name: "Reduction", TypeLine: "Enchantment", OracleID: source, Owner: me, Controller: me})
		if r := pairRefusal(g, attacker, blocker); !r.Legal() {
			t.Fatalf("effective lower power did not permit block: %q", r.Reason)
		}
	})
}
