package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	rprB03ReyavOracle  = "d1fc2549-40d5-41c8-b3ab-07ea13e9381d"
	rprB03ValdukOracle = "102b0a65-f2e0-4ef3-a9dc-4a85557b852e"
)

func rprB03Attach(g *game.Game, owner, host uuid.UUID, typeLine string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Attachment", TypeLine: typeLine,
		Owner: owner, Controller: owner, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: host},
	})
}

func TestReyavGivesDoubleStrikeToAnEquippedOrEnchantedAttackerOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Reyav, Master Smith", "Legendary Creature — Dwarf Artificer", rprB03ReyavOracle, 2, 2)
	equipped := effectstest.B12Creature(g, me.ID, "Equipped", "Creature — Bear", 2, 2)
	enchanted := effectstest.B12Creature(g, me.ID, "Enchanted", "Creature — Bear", 2, 2)
	bare := effectstest.B12Creature(g, me.ID, "Bare", "Creature — Bear", 2, 2)
	rprB03Attach(g, me.ID, equipped, "Artifact — Equipment")
	rprB03Attach(g, me.ID, enchanted, "Enchantment — Aura")
	effectstest.DeclareAttack(t, g, opp.ID, equipped, enchanted, bare)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.EffectiveAbilitiesContain(t, g, equipped, "double strike") {
		t.Error("the equipped attacker did not gain double strike")
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, enchanted, "double strike") {
		t.Error("the enchanted attacker did not gain double strike")
	}
	if effectstest.EffectiveAbilitiesContain(t, g, bare, "double strike") {
		t.Error("a bare attacker gained double strike")
	}
}

func TestReyavIgnoresAnOpponentsEquippedAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, opp.ID, "Reyav, Master Smith", "Legendary Creature — Dwarf Artificer", rprB03ReyavOracle, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Equipped", "Creature — Bear", 2, 2)
	rprB03Attach(g, me.ID, mine, "Artifact — Equipment")
	effectstest.DeclareAttack(t, g, opp.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EffectiveAbilitiesContain(t, g, mine, "double strike") {
		t.Error("an opponent's Reyav gave my attacker double strike")
	}
}

func TestValdukMakesAnElementalForEachAuraAndEquipmentAndExilesThemAtEnd(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	valduk := effectstest.B12Push(g, me.ID, "Valduk, Keeper of the Flame", "Legendary Creature — Human Shaman", rprB03ValdukOracle, 3, 2)
	other := effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	rprB03Attach(g, me.ID, valduk, "Artifact — Equipment")
	rprB03Attach(g, me.ID, valduk, "Enchantment — Aura")
	rprB03Attach(g, me.ID, other, "Artifact — Equipment")
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	toks := effectstest.Msc16TokensNamed(g, "Elemental")
	if len(toks) != 2 {
		t.Fatalf("%d Elementals, want 2 (one Aura and one Equipment on Valduk)", len(toks))
	}
	for _, tok := range toks {
		if p, tt := effectstest.EffectivePower(t, g, tok.InstanceID), effectstest.EffectiveToughness(t, g, tok.InstanceID); p != 3 || tt != 1 {
			t.Errorf("Elemental is %d/%d, want 3/1", p, tt)
		}
		if !effectstest.EffectiveAbilitiesContain(t, g, tok.InstanceID, "trample") || !effectstest.EffectiveAbilitiesContain(t, g, tok.InstanceID, "haste") {
			t.Error("Elemental lacks trample or haste")
		}
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.Msc16TokensNamed(g, "Elemental")); got != 0 {
		t.Errorf("%d Elementals left after the end step, want 0 (exiled)", got)
	}
}

func TestValdukWithNothingAttachedMakesNoTokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Valduk, Keeper of the Flame", "Legendary Creature — Human Shaman", rprB03ValdukOracle, 3, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.Msc16TokensNamed(g, "Elemental")); got != 0 {
		t.Errorf("%d Elementals, want 0", got)
	}
}
