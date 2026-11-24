package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b073_test.go — Kaldheim Commander, batch 73 (#5506): token makers.
// Each test plays the real catalog entry and asserts what its text says.

const (
	oldB073Evangel   = "4fbe1c02-f5bf-47c7-80de-09a32d01e5db"
	oldB073Artisan   = "7dbf5f59-1246-4156-bb21-090cfddd9111"
	oldB073Cullers   = "b7d69422-b90b-4149-8c6d-eef636d5c6f5"
	oldB073Cultivate = "ca4c0945-7035-47dc-aba8-fbda351d9b7b"
	oldB073Cloudgoat = "e1c13a19-46cd-474b-a46b-a483be6b91bb"
	oldB073Monk      = "e5deac0b-4417-42cc-b145-db0afe34f6e0"
)

func oldB073Enter(t *testing.T, g *game.Game, p *game.Player, c game.Card) uuid.UUID {
	t.Helper()
	id := effectstest.PutInHand(p, c)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: p.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, id); err != nil {
		t.Fatalf("enter %s: %v", c.Name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func oldB073Named(g *game.Game, controller uuid.UUID, name string) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

func TestOldB073EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB073Evangel: "Evangel of Heliod", oldB073Artisan: "Arcane Artisan", oldB073Cullers: "Eyeblight Cullers",
		oldB073Cultivate: "Cultivator of Blades", oldB073Cloudgoat: "Cloudgoat Ranger", oldB073Monk: "Geist-Honored Monk",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s is %v with caveats %v, want Full with none", name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB073EvangelMakesASoldierPerDevotionToWhiteItselfIncluded(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "White Thing", TypeLine: "Creature — Human",
		ManaCost: "{1}{W}", Power: 1, Toughness: 1, Colors: []string{"W"}, Owner: me.ID, Controller: me.ID})
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Black Thing", TypeLine: "Creature — Zombie",
		ManaCost: "{B}{B}", Power: 1, Toughness: 1, Colors: []string{"B"}, Owner: me.ID, Controller: me.ID})
	oldB073Enter(t, g, me, game.Card{Name: "Evangel of Heliod", TypeLine: "Creature — Human Cleric", ManaCost: "{4}{W}{W}",
		OracleID: oldB073Evangel, Power: 1, Toughness: 3, Colors: []string{"W"}})
	// {W} + {W}{W} = 3.
	if n := len(oldB073Named(g, me.ID, "Soldier")); n != 3 {
		t.Errorf("%d Soldiers, want 3 (devotion to white)", n)
	}
}

func TestOldB073EvangelCountsOnlyTheControllersPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their White", TypeLine: "Creature — Human",
		ManaCost: "{W}{W}{W}", Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	oldB073Enter(t, g, me, game.Card{Name: "Evangel of Heliod", TypeLine: "Creature — Human Cleric", ManaCost: "{4}{W}{W}",
		OracleID: oldB073Evangel, Power: 1, Toughness: 3})
	if n := len(oldB073Named(g, me.ID, "Soldier")); n != 2 {
		t.Errorf("%d Soldiers, want 2 (only Evangel's own {W}{W})", n)
	}
}

func TestOldB073CullersDieMakesThreeElfWarriorsThenMillsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Eyeblight Cullers", "Creature — Elf Warrior", oldB073Cullers, false)
	lib := me.Library.Size()
	effectstest.KillCreature(t, g, me.ID, id)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(oldB073Named(g, me.ID, "Elf Warrior")); n != 3 {
		t.Errorf("%d Elf Warriors, want 3", n)
	}
	if got := lib - me.Library.Size(); got != 3 {
		t.Errorf("milled %d, want 3", got)
	}
}

func TestOldB073CultivatorFabricatesWithCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := oldB073Enter(t, g, me, game.Card{Name: "Cultivator of Blades", TypeLine: "Creature — Elf Artificer", ManaCost: "{3}{G}{G}",
		OracleID: oldB073Cultivate, Power: 1, Toughness: 1})
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, id, game.CounterPlusOne); n != 2 {
		t.Errorf("%d +1/+1 counters, want 2", n)
	}
}

func TestOldB073CultivatorAttackPumpsOthersByItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cult := effectstest.PushCatalogPermanent(g, me.ID, "Cultivator of Blades", "Creature — Elf Artificer", oldB073Cultivate, false)
	effectstest.AddCountersForTest(g, cult, game.CounterPlusOne, 2) // 3/3
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	idle := effectstest.PushVanillaCreature(g, me.ID, "Idle", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, cult, bear)
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("attacking Bear is %d/%d, want 5/5", p, tough)
	}
	if p, _ := effectstest.PtOf(t, g, cult); p != 3 {
		t.Errorf("the Cultivator is %d power, want 3 (it does not pump itself)", p)
	}
	if p, _ := effectstest.PtOf(t, g, idle); p != 2 {
		t.Errorf("a non-attacker is %d power, want 2", p)
	}
}

func TestOldB073CultivatorAttackPumpCanBeDeclined(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cult := effectstest.PushCatalogPermanent(g, me.ID, "Cultivator of Blades", "Creature — Elf Artificer", oldB073Cultivate, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, cult, bear)
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, bear); p != 2 {
		t.Errorf("Bear is %d power after declining, want 2", p)
	}
}

func TestOldB073CloudgoatMakesThreeKithkinAndTapsThreeForFlyingAndPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := oldB073Enter(t, g, me, game.Card{Name: "Cloudgoat Ranger", TypeLine: "Creature — Giant Warrior Ranger", ManaCost: "{3}{W}{W}",
		OracleID: oldB073Cloudgoat, Power: 3, Toughness: 3})
	kithkin := oldB073Named(g, me.ID, "Kithkin Soldier")
	if len(kithkin) != 3 {
		t.Fatalf("%d Kithkin Soldiers, want 3", len(kithkin))
	}
	effectstest.AdvanceToMain(t, g)
	// Two are not enough.
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, TapIDs: kithkin[:2]}); err == nil {
		t.Fatal("activated with only two Kithkin tapped")
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, TapIDs: kithkin}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, id); p != 5 {
		t.Errorf("Cloudgoat Ranger has %d power, want 5", p)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, id, "flying") {
		t.Error("Cloudgoat Ranger did not gain flying")
	}
	for _, k := range kithkin {
		if !effectstest.TappedNow(g, k) {
			t.Error("a Kithkin was not tapped for the cost")
		}
	}
}

func TestOldB073CloudgoatCostOnlyTapsKithkin(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Cloudgoat Ranger", "Creature — Giant Warrior Ranger", oldB073Cloudgoat, false)
	var bears []uuid.UUID
	for i := 0; i < 3; i++ {
		bears = append(bears, effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2))
	}
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, TapIDs: bears}); err == nil {
		t.Fatal("tapped three non-Kithkin creatures to pay")
	}
}

func TestOldB073MonkCountsCreaturesAndMakesTwoFlyingSpirits(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	id := oldB073Enter(t, g, me, game.Card{Name: "Geist-Honored Monk", TypeLine: "Creature — Human Monk", ManaCost: "{3}{W}{W}",
		OracleID: oldB073Monk})
	spirits := oldB073Named(g, me.ID, "Spirit")
	if len(spirits) != 2 {
		t.Fatalf("%d Spirits, want 2", len(spirits))
	}
	for _, s := range spirits {
		if !effectstest.HasKeywordOnBattlefield(t, g, s, "flying") {
			t.Error("a Spirit lacks flying")
		}
	}
	// Bear + Monk + two Spirits.
	if p, tough := effectstest.PtOf(t, g, id); p != 4 || tough != 4 {
		t.Errorf("the Monk is %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, id, "vigilance") {
		t.Error("the Monk lacks vigilance")
	}
	effectstest.KillCreature(t, g, me.ID, spirits[0])
	g.RunStateChecksForTest()
	if p, _ := effectstest.PtOf(t, g, id); p != 3 {
		t.Errorf("the Monk has %d power after a Spirit died, want 3", p)
	}
}

func oldB073ArtisanActivate(t *testing.T, g *game.Game, me *game.Player, artisan, target uuid.UUID) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{U}")
	if err := g.ActivateCatalogAbility(me.ID, artisan, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: target}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB073ArtisanCopiesAnExiledCreatureCardForTheTargetPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	artisan := effectstest.PushCatalogPermanent(g, me.ID, "Arcane Artisan", "Creature — Human Wizard", oldB073Artisan, false)
	ogre := effectstest.PutInHand(opp, game.Card{Name: "Ogre", TypeLine: "Creature — Ogre", Power: 4, Toughness: 4})
	oldB073ArtisanActivate(t, g, me, artisan, opp.ID)
	effectstest.AnswerChooseCards(t, g, opp.ID, ogre)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(ogre) {
		t.Fatal("the creature card was not exiled")
	}
	tokens := oldB073Named(g, opp.ID, "Ogre")
	if len(tokens) != 1 {
		t.Fatalf("%d Ogre tokens for the target player, want 1", len(tokens))
	}
	if c := effectstest.CardByID(g, tokens[0]); !c.IsToken() {
		t.Error("the copy is not a token")
	}
	if len(oldB073Named(g, me.ID, "Ogre")) != 0 {
		t.Error("the Artisan's controller got the token")
	}
}

func TestOldB073ArtisanNoncreatureCardMakesNoToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	artisan := effectstest.PushCatalogPermanent(g, me.ID, "Arcane Artisan", "Creature — Human Wizard", oldB073Artisan, false)
	spell := effectstest.PutInHand(me, game.Card{Name: "Shock", TypeLine: "Instant"})
	oldB073ArtisanActivate(t, g, me, artisan, me.ID)
	effectstest.AnswerChooseCards(t, g, me.ID, spell)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(spell) {
		t.Fatal("the instant was not exiled")
	}
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() {
			t.Errorf("a token %s was made for a noncreature card", c.Name)
		}
	}
}

func TestOldB073ArtisanTokensAreExiledAtTheNextEndStepAfterItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	artisan := effectstest.PushCatalogPermanent(g, me.ID, "Arcane Artisan", "Creature — Human Wizard", oldB073Artisan, false)
	ogre := effectstest.PutInHand(me, game.Card{Name: "Ogre", TypeLine: "Creature — Ogre", Power: 4, Toughness: 4})
	oldB073ArtisanActivate(t, g, me, artisan, me.ID)
	effectstest.AnswerChooseCards(t, g, me.ID, ogre)
	effectstest.PassPriorityAroundTable(t, g)
	tokens := oldB073Named(g, me.ID, "Ogre")
	if len(tokens) != 1 {
		t.Fatalf("%d Ogre tokens, want 1", len(tokens))
	}
	effectstest.KillCreature(t, g, me.ID, artisan)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(tokens[0]) {
		t.Fatal("the token left before the end step")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tokens[0]) {
		t.Error("the token stayed after the next end step began")
	}
}

func TestOldB073ArtisanTokensStayWhileItDoes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	artisan := effectstest.PushCatalogPermanent(g, me.ID, "Arcane Artisan", "Creature — Human Wizard", oldB073Artisan, false)
	ogre := effectstest.PutInHand(me, game.Card{Name: "Ogre", TypeLine: "Creature — Ogre", Power: 4, Toughness: 4})
	oldB073ArtisanActivate(t, g, me, artisan, me.ID)
	effectstest.AnswerChooseCards(t, g, me.ID, ogre)
	effectstest.PassPriorityAroundTable(t, g)
	tokens := oldB073Named(g, me.ID, "Ogre")
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if len(tokens) != 1 || !g.Battlefield.Contains(tokens[0]) {
		t.Error("the token left although the Artisan is still on the battlefield")
	}
}
