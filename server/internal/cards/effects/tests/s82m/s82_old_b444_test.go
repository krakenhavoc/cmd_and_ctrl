package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Innistrad batch 444: Curses and two creature Auras.

const (
	oldB444DeathsHold   = "7194cf15-69fc-46f0-987a-2cea8b498f50"
	oldB444StalkedPrey  = "7971a2a6-f89f-46a5-9008-96cfa49de41a"
	oldB444BloodyTome   = "5480f855-80b7-45c7-94a9-071c7b5d5ccf"
	oldB444Oblivion     = "10b16b2b-7d7c-4f6f-932a-1f3bd30d33f1"
	oldB444PiercedHeart = "f22c3624-6bd5-4761-b8ad-ac301ad9351c"
	oldB444Grimace      = "99e03fa0-8e0c-4b5d-880f-d7de201d5f41"
	oldB444Wreath       = "b11c31ac-9a52-4faa-b03e-5a78129fd43f"
)

func TestOldB444CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB444DeathsHold: "Curse of Death's Hold", oldB444StalkedPrey: "Curse of Stalked Prey",
		oldB444BloodyTome: "Curse of the Bloody Tome", oldB444Oblivion: "Curse of Oblivion",
		oldB444PiercedHeart: "Curse of the Pierced Heart", oldB444Grimace: "Skeletal Grimace", oldB444Wreath: "Wreath of Geists",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q, completeness %v, caveats %v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// oldB444Curse casts the Curse at `target` from seat 0 and lets it resolve.
func oldB444Curse(t *testing.T, g *game.Game, name, oracle string, target uuid.UUID) uuid.UUID {
	t.Helper()
	id := effectstest.CastCatalogSpell(t, g, name, "Enchantment — Aura Curse", oracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: target}})
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok || c.AttachedTo.Kind != game.TargetPlayer || c.AttachedTo.ID != target {
		t.Fatalf("%s did not enchant the player: on battlefield %v, attached to %+v", name, ok, c.AttachedTo)
	}
	return id
}

func TestOldB444CurseOfDeathsHoldShrinksOnlyTheEnchantedPlayersCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Beast", 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Beast", 3, 3)
	other := effectstest.B12Creature(g, g.Seats[2].ID, "Other", "Creature — Beast", 3, 3)
	oldB444Curse(t, g, "Curse of Death's Hold", oldB444DeathsHold, opp.ID)
	for id, want := range map[uuid.UUID]int{theirs: 2, mine: 3, other: 3} {
		if p, tough := effectstest.EffectivePTOf(t, g, id); p != want || tough != want {
			t.Errorf("creature %s is %d/%d, want %d/%d", id, p, tough, want, want)
		}
	}
	// A creature that comes in later gets it too.
	late := effectstest.B12Creature(g, opp.ID, "Late", "Creature — Beast", 2, 2)
	if p, tough := effectstest.EffectivePTOf(t, g, late); p != 1 || tough != 1 {
		t.Errorf("a creature that arrived later is %d/%d, want 1/1", p, tough)
	}
}

func TestOldB444CurseOfStalkedPreyGrowsOnlyCreaturesThatHitTheEnchantedPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	hitter := effectstest.B12Creature(g, me.ID, "Hitter", "Creature — Beast", 2, 2)
	oldB444Curse(t, g, "Curse of Stalked Prey", oldB444StalkedPrey, opp.ID)
	effectstest.DealCombatDamageToPlayer(g, hitter, third.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, hitter); got != 0 {
		t.Fatalf("hitting another player gave %d counters", got)
	}
	effectstest.DealCombatDamageToPlayer(g, hitter, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, hitter); got != 1 {
		t.Fatalf("hitting the enchanted player gave %d counters, want 1", got)
	}
}

func TestOldB444CurseOfTheBloodyTomeMillsTwoAtTheEnchantedPlayersUpkeepOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB444Curse(t, g, "Curse of the Bloody Tome", oldB444BloodyTome, opp.ID)
	mine, theirs := me.Library.Size(), opp.Library.Size()
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := theirs - opp.Library.Size(); got != 2 {
		t.Errorf("the enchanted player milled %d cards, want 2 (the draw step has not come yet)", got)
	}
	if me.Library.Size() != mine {
		t.Errorf("the Curse's controller milled %d cards", mine-me.Library.Size())
	}
}

func TestOldB444CurseOfOblivionLetsThePlayerChooseTwoGraveyardCardsToExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	var ids []uuid.UUID
	for _, n := range []string{"A", "B", "C"} {
		id := uuid.New()
		opp.Graveyard.PushTop(game.Card{InstanceID: id, Name: n, TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
		ids = append(ids, id)
	}
	oldB444Curse(t, g, "Curse of Oblivion", oldB444Oblivion, opp.ID)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, opp.ID)
	if c == nil || c.ChooseMin != 2 || c.ChooseMax != 2 {
		t.Fatalf("want a prompt for exactly two cards, answered by the enchanted player: %+v", c)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{ids[0], ids[1]}); err == nil {
		t.Error("the Curse's controller answered for the enchanted player")
	}
	if err := g.ResolveChooseCards(c.ID, opp.ID, []uuid.UUID{ids[0]}); err == nil {
		t.Error("one card was accepted for \"two cards\"")
	}
	if err := g.ResolveChooseCards(c.ID, opp.ID, []uuid.UUID{ids[1], ids[2]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(ids[1]) || !g.Exile.Contains(ids[2]) || !opp.Graveyard.Contains(ids[0]) {
		t.Error("the two chosen cards should be exiled and the third left")
	}
}

func TestOldB444CurseOfTheBloodyTomeAndOblivionWithAnEmptyGraveyardDoNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	oldB444Curse(t, g, "Curse of Oblivion", oldB444Oblivion, opp.ID)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LatestChooseCardsFor(g, opp.ID); c != nil {
		t.Errorf("asked to exile from an empty graveyard: %+v", c)
	}
}

func TestOldB444CurseOfThePiercedHeartPingsThePlayerOrAPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB444Curse(t, g, "Curse of the Pierced Heart", oldB444PiercedHeart, opp.ID)
	life := opp.Life
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Fatalf("life %d → %d, want 1 damage with no planeswalker to choose", life, opp.Life)
	}
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{game.CounterLoyalty: 4}, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, walker, game.CounterLoyalty); got != 3 {
		t.Errorf("loyalty %d, want 3 after the Curse's controller chose the planeswalker", got)
	}
	if opp.Life != life-1 {
		t.Errorf("the player took the damage as well: life %d, want %d", opp.Life, life-1)
	}
}

func TestOldB444SkeletalGrimacePumpsAndGrantsRegeneration(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.AuraBear(g, me.ID)
	effectstest.AuraCast(t, g, "Skeletal Grimace", oldB444Grimace, bear)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("enchanted Bear is %d/%d, want 3/3", p, tough)
	}
	idx, ref := effectstest.GrantedActivatedIndex(t, g, bear)
	if idx < 0 {
		t.Fatal("the enchanted creature has no granted ability")
	}
	effectstest.FloatMana(t, g, me, "{B}")
	effectstest.B16Activate(t, g, me.ID, bear, idx, game.ActivateAbilityParams{Ref: ref, Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentsKeyedForEffect([]uuid.UUID{bear}, game.KeyedTail{}) })
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the regeneration shield did not save the creature")
	}
}

func TestOldB444WreathOfGeistsCountsYourCreatureCardsInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.AuraBear(g, me.ID)
	push := func(p *game.Player, typeLine string) uuid.UUID {
		id := uuid.New()
		p.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Card", TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
		return id
	}
	push(me, "Creature — Bear")
	push(me, "Creature — Bear")
	push(me, "Sorcery")
	push(opp, "Creature — Bear")
	effectstest.AuraCast(t, g, "Wreath of Geists", oldB444Wreath, bear)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("with two creature cards of yours in the graveyard the Bear is %d/%d, want 4/4", p, tough)
	}
	golem := effectstest.B12Creature(g, me.ID, "Golem", "Artifact Creature — Golem", 1, 1)
	effectstest.KillCreature(t, g, me.ID, golem)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("after a third creature card the Bear is %d/%d, want 5/5", p, tough)
	}
	// "Your" graveyard is the Aura's controller's, even on an opponent's creature.
	theirs := effectstest.AuraBear(g, opp.ID)
	effectstest.AuraCast(t, g, "Wreath of Geists", oldB444Wreath, theirs)
	if p, tough := effectstest.EffectivePTOf(t, g, theirs); p != 5 || tough != 5 {
		t.Errorf("an opponent's creature under my Wreath is %d/%d, want 5/5 (my three creature cards)", p, tough)
	}
}
