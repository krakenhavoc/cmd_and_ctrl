package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const jukaiNaturalistOracle = "c4d1a78f-0cca-4042-862c-8f322ee25d50"

func TestJukaiNaturalistIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(jukaiNaturalistOracle)
	if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Jukai Naturalist: %+v %v", spec, ok)
	}
}

func TestJukaiNaturalistDiscountsOnlyYourEnchantmentSpells(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, cost, pool string
		wantErr                    bool
	}{
		{"enchantment is {1} cheaper", "Enchantment", "{1}{W}", "W", false},
		{"enchantment at full price", "Enchantment", "{1}{W}", "", true},
		{"a creature is not discounted", "Creature — Bear", "{1}{G}", "G", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.PushCatalogPermanent(g, me.ID, "Jukai Naturalist", "Enchantment Creature — Human Monk", jukaiNaturalistOracle, false)
			effectstest.AdvanceToMain(t, g)
			effectstest.FloatForTest(g, me, tc.pool)
			id := uuid.New()
			g.WithWriteLock(func() {
				me.Hand.PushTop(game.Card{InstanceID: id, Name: "Test spell", TypeLine: tc.typeLine, ManaCost: tc.cost, Owner: me.ID, Controller: me.ID})
			})
			err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
			if (err != nil) != tc.wantErr {
				t.Fatalf("cast err = %v, want error %v", err, tc.wantErr)
			}
		})
	}
}

func TestJukaiNaturalistHasLifelink(t *testing.T) {
	spec, _ := effects.Lookup(jukaiNaturalistOracle)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "lifelink" {
		t.Fatalf("keywords = %v", spec.PrintedKeywords)
	}
}
