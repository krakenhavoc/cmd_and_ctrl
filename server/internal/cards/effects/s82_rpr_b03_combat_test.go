package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	rprB03ReyavOracle  = "d1fc2549-40d5-41c8-b3ab-07ea13e9381d"
	rprB03ValdukOracle = "102b0a65-f2e0-4ef3-a9dc-4a85557b852e"
)

func rprB03Attach(g *game.Game, owner, host uuid.UUID, typeLine string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Attachment", TypeLine: typeLine,
		Owner: owner, Controller: owner, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: host},
	})
}

func TestReyavGivesDoubleStrikeToAnEquippedOrEnchantedAttackerOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Reyav, Master Smith", "Legendary Creature — Dwarf Artificer", rprB03ReyavOracle, 2, 2)
	equipped := b12Creature(g, me.ID, "Equipped", "Creature — Bear", 2, 2)
	enchanted := b12Creature(g, me.ID, "Enchanted", "Creature — Bear", 2, 2)
	bare := b12Creature(g, me.ID, "Bare", "Creature — Bear", 2, 2)
	rprB03Attach(g, me.ID, equipped, "Artifact — Equipment")
	rprB03Attach(g, me.ID, enchanted, "Enchantment — Aura")
	declareAttack(t, g, opp.ID, equipped, enchanted, bare)
	passPriorityAroundTable(t, g)
	if !effectiveAbilitiesContain(t, g, equipped, "double strike") {
		t.Error("the equipped attacker did not gain double strike")
	}
	if !effectiveAbilitiesContain(t, g, enchanted, "double strike") {
		t.Error("the enchanted attacker did not gain double strike")
	}
	if effectiveAbilitiesContain(t, g, bare, "double strike") {
		t.Error("a bare attacker gained double strike")
	}
}

func TestReyavIgnoresAnOpponentsEquippedAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, opp.ID, "Reyav, Master Smith", "Legendary Creature — Dwarf Artificer", rprB03ReyavOracle, 2, 2)
	mine := b12Creature(g, me.ID, "Equipped", "Creature — Bear", 2, 2)
	rprB03Attach(g, me.ID, mine, "Artifact — Equipment")
	declareAttack(t, g, opp.ID, mine)
	passPriorityAroundTable(t, g)
	if effectiveAbilitiesContain(t, g, mine, "double strike") {
		t.Error("an opponent's Reyav gave my attacker double strike")
	}
}

func TestValdukMakesAnElementalForEachAuraAndEquipmentAndExilesThemAtEnd(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	valduk := b12Push(g, me.ID, "Valduk, Keeper of the Flame", "Legendary Creature — Human Shaman", rprB03ValdukOracle, 3, 2)
	other := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	rprB03Attach(g, me.ID, valduk, "Artifact — Equipment")
	rprB03Attach(g, me.ID, valduk, "Enchantment — Aura")
	rprB03Attach(g, me.ID, other, "Artifact — Equipment")
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	toks := msc16TokensNamed(g, "Elemental")
	if len(toks) != 2 {
		t.Fatalf("%d Elementals, want 2 (one Aura and one Equipment on Valduk)", len(toks))
	}
	for _, tok := range toks {
		if p, tt := effectivePower(t, g, tok.InstanceID), effectiveToughness(t, g, tok.InstanceID); p != 3 || tt != 1 {
			t.Errorf("Elemental is %d/%d, want 3/1", p, tt)
		}
		if !effectiveAbilitiesContain(t, g, tok.InstanceID, "trample") || !effectiveAbilitiesContain(t, g, tok.InstanceID, "haste") {
			t.Error("Elemental lacks trample or haste")
		}
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := len(msc16TokensNamed(g, "Elemental")); got != 0 {
		t.Errorf("%d Elementals left after the end step, want 0 (exiled)", got)
	}
}

func TestValdukWithNothingAttachedMakesNoTokens(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Valduk, Keeper of the Flame", "Legendary Creature — Human Shaman", rprB03ValdukOracle, 3, 2)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if got := len(msc16TokensNamed(g, "Elemental")); got != 0 {
		t.Errorf("%d Elementals, want 0", got)
	}
}
