package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Theros Beyond Death batch 135 (S82 OLD, #5514).

const (
	oldB135Allure     = "252d9b25-45a1-40ca-b563-4cdefc791a44"
	oldB135Libation   = "32c84176-cf11-4010-b1cc-ed570a8c3e54"
	oldB135Repeal     = "3f453aca-e14f-4460-9334-f0c3f4256a0e"
	oldB135Design     = "85c20c39-7b9c-4661-9372-e659b321773f"
	oldB135Scorn      = "d2497a5e-ff2b-4024-9b14-634304d583fe"
	oldB135Altar      = "17d41bbf-6a84-48c4-baac-e45f8ed86b78"
	oldB135Petitioner = "63d92a82-d77d-4c47-a49b-bde6a8c9dd8d"
	oldB135Lampad     = "e0f23e7e-2858-4f5e-9dd5-8a0ba934a048"
)

func TestOldB135CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB135Allure: "Allure of the Unknown", oldB135Libation: "Pharika's Libation", oldB135Repeal: "Mystic Repeal",
		oldB135Design: "Klothys's Design", oldB135Scorn: "Sea God's Scorn", oldB135Altar: "Altar of the Pantheon",
		oldB135Petitioner: "Setessan Petitioner", oldB135Lampad: "Lampad of Death's Vigil",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB135CastPetitioner(t *testing.T, g *game.Game) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.RfdHandCard(me, "Setessan Petitioner", "Creature — Human Druid", "{1}{G}{G}", oldB135Petitioner, 2, 2)
	effectstest.ToMainForCost(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB135PetitionerGainsLifeEqualToDevotionToGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Green Thing", TypeLine: "Creature — Elf", ManaCost: "{G}{G}{G}", Power: 1, Toughness: 1})
	before := me.Life
	oldB135CastPetitioner(t, g)
	if got := me.Life - before; got != 5 {
		t.Errorf("gained %d life, want 5 (three from the Elf and two from the Petitioner)", got)
	}
}

func TestOldB135AltarAddsOneToEveryDevotion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Altar of the Pantheon", OracleID: oldB135Altar, TypeLine: "Artifact", ManaCost: "{3}"})
	before := me.Life
	oldB135CastPetitioner(t, g)
	if got := me.Life - before; got != 3 {
		t.Errorf("gained %d life, want 3 (two from the Petitioner and one from the Altar)", got)
	}
	var devotion int
	g.WithWriteLock(func() { devotion = effects.DevotionToColors(g, me.ID, "B", "R") })
	if devotion != 1 {
		t.Errorf("devotion to black and red is %d, want 1 (the Altar's bonus, once)", devotion)
	}
}

func TestOldB135AltarMakesManaAndGainsLifeOnlyWithAGodDemigodOrLegendaryEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	altar := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Altar of the Pantheon", OracleID: oldB135Altar, TypeLine: "Artifact", ManaCost: "{3}"})
	tapAltar := func() {
		t.Helper()
		if err := g.ActivateManaAbility(me.ID, altar, 0, game.ManaAbilityParams{Colors: []string{"W"}}); err != nil {
			t.Fatalf("tap the Altar: %v", err)
		}
	}
	before := me.Life
	tapAltar()
	if got := me.Life - before; got != 0 {
		t.Errorf("gained %d life with nothing to enable it, want 0", got)
	}
	if n := len(effectstest.PoolColors(me)); n != 1 {
		t.Errorf("the Altar made %d mana, want 1", n)
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == altar {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Legendary Ench", TypeLine: "Legendary Enchantment"})
	before = me.Life
	tapAltar()
	if got := me.Life - before; got != 1 {
		t.Errorf("gained %d life with a legendary enchantment, want 1", got)
	}
}

func TestOldB135KlothyssDesignPumpsYourCreaturesByDevotionToGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Green Thing", TypeLine: "Creature — Elf", ManaCost: "{G}{G}", Power: 1, Toughness: 1})
	other := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Klothys's Design", "Sorcery", oldB135Design, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, other); p != 4 || tt != 4 {
		t.Errorf("your creature is %d/%d, want 4/4 (+2/+2)", p, tt)
	}
	if p, tt := effectstest.PtOf(t, g, theirs); p != 2 || tt != 2 {
		t.Errorf("their creature is %d/%d, want 2/2", p, tt)
	}
}

func TestOldB135LibationMakesAnOpponentSacrificeACreatureOrAnEnchantment(t *testing.T) {
	for mode, kind := range map[int]string{0: "Creature — Bear", 1: "Enchantment"} {
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[1]
		victim := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Victim", TypeLine: kind, Power: 1, Toughness: 1})
		other := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Bystander", TypeLine: "Artifact"})
		id := effectstest.RfdHandCard(g.Seats[0], "Pharika's Libation", "Instant", "{2}{B}", oldB135Libation, 0, 0)
		effectstest.ToMainForCost(t, g)
		if err := g.CastSpell(g.Seats[0].ID, id, game.CastSpellParams{Modes: []int{mode}, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
			t.Fatalf("mode %d: CastSpell: %v", mode, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerSacrifice(t, g, opp.ID, victim)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(victim) || !g.Battlefield.Contains(other) {
			t.Errorf("mode %d: victim on battlefield %v, bystander %v", mode, g.Battlefield.Contains(victim), g.Battlefield.Contains(other))
		}
	}
}

func TestOldB135MysticRepealPutsAnEnchantmentOnTheBottomOfItsOwnersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	ench := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Their Enchantment", TypeLine: "Enchantment"})
	effectstest.CastCatalogSpell(t, g, "Mystic Repeal", "Instant", oldB135Repeal, []game.TargetRef{{Kind: game.TargetCard, ID: ench}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) {
		t.Fatal("the enchantment is still on the battlefield")
	}
	if n := len(opp.Library.Cards); n == 0 || opp.Library.Cards[0].InstanceID != ench {
		t.Errorf("the enchantment is not at the bottom of its owner's library (size %d)", n)
	}
}

func TestOldB135SeaGodsScornReturnsUpToThreeCreaturesAndEnchantments(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 1, 1)
	b := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "B", TypeLine: "Enchantment"})
	c := effectstest.B12Creature(g, me.ID, "C", "Creature — Bear", 1, 1)
	d := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "D", TypeLine: "Artifact"})
	effectstest.CastCatalogSpell(t, g, "Sea God's Scorn", "Sorcery", oldB135Scorn, []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}, {Kind: game.TargetCard, ID: c}})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c} {
		if g.Battlefield.Contains(id) {
			t.Errorf("a target stayed on the battlefield")
		}
	}
	if !g.Battlefield.Contains(d) {
		t.Error("an untargeted artifact left")
	}
	if !opp.Hand.Contains(a) || !opp.Hand.Contains(b) || !me.Hand.Contains(c) {
		t.Error("the targets did not go to their owners' hands")
	}
}

func TestOldB135LampadSacrificesACreatureToDrainOneLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lampad := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Lampad of Death's Vigil", OracleID: oldB135Lampad, TypeLine: "Enchantment Creature — Nymph", ManaCost: "{1}{B}", Power: 1, Toughness: 3})
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}")
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		lives[s.ID] = s.Life
	}
	if err := g.ActivateCatalogAbility(me.ID, lampad, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature was not sacrificed")
	}
	if got := me.Life - lives[me.ID]; got != 1 {
		t.Errorf("you gained %d life, want 1", got)
	}
	for _, s := range g.Seats {
		if s.ID != me.ID && lives[s.ID]-s.Life != 1 {
			t.Errorf("an opponent lost %d life, want 1", lives[s.ID]-s.Life)
		}
	}
}

func TestOldB135AllureOpponentExilesANonlandRestGoesToHandAndTheyMayCastIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	var ids []uuid.UUID
	g.WithWriteLock(func() {
		me.Library.Cards = nil
		for i := 0; i < 4; i++ {
			me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Creature — Bear", ManaCost: "{1}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
		}
		for i := 0; i < 2; i++ {
			me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Plains", TypeLine: "Basic Land — Plains", Owner: me.ID, Controller: me.ID})
		}
		for _, c := range me.Library.Cards {
			ids = append(ids, c.InstanceID)
		}
	})
	handBefore := effectstest.HandSize(me)
	effectstest.CastCatalogSpell(t, g, "Allure of the Unknown", "Sorcery", oldB135Allure, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	effectstest.PassPriorityAroundTable(t, g)
	var pick uuid.UUID
	if p := effectstest.LatestChooseCards(g, opp.ID); p == nil {
		t.Fatal("the opponent was offered no card to exile")
	} else {
		for _, id := range p.ChooseCards {
			if c, ok := g.LookupCardForEffect(id); ok && c.IsLand() {
				t.Errorf("a land was offered to exile")
			}
		}
		pick = p.ChooseCards[0]
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, pick)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.AdZone(g, pick); got != game.ZoneExile {
		t.Errorf("the picked card is in zone %v, want exile", got)
	}
	if got := effectstest.HandSize(me) - handBefore; got != 5 {
		t.Errorf("you put %d cards into your hand, want the other five", got)
	}
	_ = ids
}
