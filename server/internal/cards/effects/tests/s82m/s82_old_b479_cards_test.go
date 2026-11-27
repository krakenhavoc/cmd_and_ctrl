package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b479_cards_test.go — Rise of the Eldrazi: creatures with keywords,
// batch 479 (#5581, S82 OLD). Every test plays the real catalog entry through
// the engine.

const (
	oldB479Guardians   = "11f7f6e4-2c39-4478-8c10-4473dbf921a2"
	oldB479Gloomhunter = "637b1a19-2925-422e-b867-f87537803071"
	oldB479Cub         = "9d448c85-9c93-4e9e-8aa0-0035fc94112e"
	oldB479Soulsurge   = "a98764e2-3cf6-40a4-aba5-34e2411623ab"
	oldB479Hydra       = "d4649dd8-9a30-4f7e-8b9d-8edee8b2dbd7"
	oldB479Abomination = "f4925e05-1332-4466-b2ea-9f6f4ccf7db4"
)

func TestOldB479CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB479Guardians:   "Soulbound Guardians",
		oldB479Gloomhunter: "Gloomhunter",
		oldB479Cub:         "Stomper Cub",
		oldB479Soulsurge:   "Soulsurge Elemental",
		oldB479Hydra:       "Khalni Hydra",
		oldB479Abomination: "Phantasmal Abomination",
	} {
		effectstest.RequireFullNoCaveats(t, oracle, name)
	}
}

func TestOldB479PrintedKeywordsAndSizes(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typ string
		power, toughness  int
		keywords          []string
	}{
		{"Soulbound Guardians", oldB479Guardians, "Creature — Kor Spirit", 4, 5, []string{"defender", "flying"}},
		{"Gloomhunter", oldB479Gloomhunter, "Creature — Bat", 2, 1, []string{"flying"}},
		{"Stomper Cub", oldB479Cub, "Creature — Beast", 5, 3, []string{"trample"}},
		{"Phantasmal Abomination", oldB479Abomination, "Creature — Illusion", 5, 5, []string{"defender"}},
		{"Khalni Hydra", oldB479Hydra, "Creature — Hydra", 8, 8, []string{"trample"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, tc.typ, tc.oracle, tc.power, tc.toughness)
			if p, tough := effectstest.EffectivePTOf(t, g, id); p != tc.power || tough != tc.toughness {
				t.Errorf("%d/%d, want %d/%d", p, tough, tc.power, tc.toughness)
			}
			for _, kw := range tc.keywords {
				if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
					t.Errorf("no %s", kw)
				}
			}
		})
	}
}

func TestOldB479DefendersCannotAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guard := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Soulbound Guardians", TypeLine: "Creature — Kor Spirit",
		OracleID: oldB479Guardians, Power: 4, Toughness: 5, Owner: me.ID, Controller: me.ID, Keywords: []string{"haste"}})
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	if err := g.DeclareAttacker(guard, opp.ID); err == nil {
		t.Error("a creature with defender attacked")
	}
}

func TestOldB479SoulsurgeElementalPowerIsTheNumberOfCreaturesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Soulsurge Elemental", TypeLine: "Creature — Elemental",
		OracleID: oldB479Soulsurge, Power: 0, Toughness: 1, Owner: me.ID, Controller: me.ID})
	if !effectstest.HasEffectiveKeyword(t, g, id, "first strike") {
		t.Error("no first strike")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, id); p != 1 || tough != 1 {
		t.Fatalf("alone it is %d/%d, want 1/1 (it counts itself)", p, tough)
	}
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Another Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if p, _ := effectstest.EffectivePTOf(t, g, id); p != 3 {
		t.Errorf("power %d with three of my creatures and one of theirs, want 3", p)
	}
}

func oldB479HydraPrice(t *testing.T, green, otherColour, theirGreen int) string {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < green; i++ {
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Green Bear", TypeLine: "Creature — Bear",
			Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	}
	for i := 0; i < otherColour; i++ {
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Red Bear", TypeLine: "Creature — Bear",
			Colors: []string{"R"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	}
	for i := 0; i < theirGreen; i++ {
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Green Bear", TypeLine: "Creature — Bear",
			Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	}
	id := effectstest.HandCardFull(me, "Khalni Hydra", "Creature — Hydra", "{G}{G}{G}{G}{G}{G}{G}{G}", oldB479Hydra, []string{"G"})
	var card game.Card
	for _, c := range me.Hand.Cards {
		if c.InstanceID == id {
			card = c
		}
	}
	price, err := g.PriceCast(me.ID, card, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("PriceCast: %v", err)
	}
	return price.Total.String()
}

func TestOldB479KhalniHydraCostsOneGreenLessForEachGreenCreatureYouControl(t *testing.T) {
	for _, tc := range []struct {
		green, other, theirs int
		want                 string
	}{
		{0, 0, 0, "{G}{G}{G}{G}{G}{G}{G}{G}"},
		{3, 2, 4, "{G}{G}{G}{G}{G}"},
		{8, 0, 0, ""},
	} {
		got := oldB479HydraPrice(t, tc.green, tc.other, tc.theirs)
		if got != tc.want && !(tc.want == "" && (got == "{0}" || got == "")) {
			t.Errorf("%d green, %d other, %d theirs: price %q, want %q", tc.green, tc.other, tc.theirs, got, tc.want)
		}
	}
}

func TestOldB479KhalniHydraNeverCostsLessThanZero(t *testing.T) {
	if got := oldB479HydraPrice(t, 10, 0, 0); got != "{0}" && got != "" {
		t.Errorf("ten green creatures: price %q, want nothing", got)
	}
}

func TestOldB479PhantasmalAbominationIsSacrificedWhenTargeted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	abom := effectstest.B12Push(g, opp.ID, "Phantasmal Abomination", "Creature — Illusion", oldB479Abomination, 5, 5)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.CardTarget(abom))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(abom) {
		t.Error("the Abomination survived being targeted")
	}
}
