package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b447_test.go — Innistrad graveyard recursion and removal, S82 OLD
// batch 447 (#5572), played through the real catalog entries.

const (
	oldB447Lunge      = "1392cfbc-ea8a-4eb3-8680-ddd12a3bbb6d"
	oldB447Brink      = "704aaf92-590f-46d3-8b39-4ab84e1c766d"
	oldB447Shovel     = "868b5315-da38-4918-b8e4-70762811605b"
	oldB447Wish       = "888ccda6-896f-40f3-93a1-efc48254af34"
	oldB447Chant      = "b685fff3-f1b4-42d5-9b7b-e0885bcc92d2"
	oldB447Mire       = "44e9030d-dd49-46b4-a0eb-75a49b6638a2"
	oldB447MawHell    = "5f0e8fd5-9db9-4507-a42f-cb99080b89e2"
	oldB447Orb        = "7411a8f4-b736-482f-886e-5fa671e94e0d"
	oldB447Paraselene = "e14ac0f3-00d5-4a95-990c-b7ea157bb87e"
)

func oldB447Graveyard(p *game.Player, name, typeLine, cost string, power, toughness int) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID})
	return id
}

func oldB447Cast(t *testing.T, g *game.Game, me *game.Player, name, typeLine, cost, oracle string, params game.CastSpellParams) error {
	t.Helper()
	params.Strict = true
	_, err := effectstest.CastFromHandErr(g, me, name, typeLine, cost, oracle, params)
	return err
}

func TestOldB447EveryCardIsRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		oldB447Lunge: "Corpse Lunge", oldB447Brink: "Back from the Brink", oldB447Shovel: "Graveyard Shovel",
		oldB447Wish: "Make a Wish", oldB447Chant: "Ghoulcaller's Chant", oldB447Mire: "Maw of the Mire",
		oldB447MawHell: "Into the Maw of Hell", oldB447Orb: "Witchbane Orb", oldB447Paraselene: "Paraselene",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

func TestOldB447CorpseLungeDealsTheExiledCardsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	target := effectstest.B12Creature(g, opp.ID, "Target", "Creature — Giant", 1, 6)
	corpse := oldB447Graveyard(me, "Corpse", "Creature — Zombie", "{2}", 4, 1)
	notCreature := oldB447Graveyard(me, "Spell", "Sorcery", "{1}", 0, 0)
	effectstest.B06AddMana(me, "B", "C", "C")
	if err := oldB447Cast(t, g, me, "Corpse Lunge", "Instant", "{2}{B}", oldB447Lunge,
		game.CastSpellParams{Targets: effectstest.B16TargetCard(target)}); err == nil {
		t.Fatal("cast without exiling a creature card")
	}
	if err := oldB447Cast(t, g, me, "Corpse Lunge", "Instant", "{2}{B}", oldB447Lunge,
		game.CastSpellParams{Targets: effectstest.B16TargetCard(target), ExileIDs: []uuid.UUID{notCreature}}); err == nil {
		t.Fatal("exiled a card that is not a creature card")
	}
	if err := oldB447Cast(t, g, me, "Corpse Lunge", "Instant", "{2}{B}", oldB447Lunge,
		game.CastSpellParams{Targets: effectstest.B16TargetCard(target), ExileIDs: []uuid.UUID{corpse}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if !g.Exile.Contains(corpse) {
		t.Error("the creature card was not exiled by the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Pr6Marked(g, target); got != 4 {
		t.Errorf("the target has %d damage, want the exiled card's power, 4", got)
	}
}

func TestOldB447BackFromTheBrinkCopiesTheExiledCreatureForItsManaCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	brink := effectstest.PushCatalogPermanent(g, me.ID, "Back from the Brink", "Enchantment", oldB447Brink, false)
	beast := oldB447Graveyard(me, "Graveyard Beast", "Creature — Beast", "{2}{G}", 3, 3)
	big := oldB447Graveyard(me, "Big Beast", "Creature — Beast", "{4}{G}", 5, 5)
	land := oldB447Graveyard(me, "Forest", "Basic Land — Forest", "", 0, 0)
	effectstest.B06AddMana(me, "G", "C", "C")
	for name, id := range map[string]uuid.UUID{"a five-mana card with three mana": big, "a land card": land} {
		if err := g.ActivateCatalogAbility(me.ID, brink, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{id}}); err == nil {
			t.Fatalf("activated naming %s", name)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, brink, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{beast}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(beast) {
		t.Error("the creature card was not exiled")
	}
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Graveyard Beast"); got != 1 {
		t.Fatalf("%d copies of Graveyard Beast on the battlefield, want 1", got)
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Graveyard Beast" && (!c.IsToken() || c.Power != 3 || c.Toughness != 3) {
			t.Errorf("the copy is a token=%v %d/%d, want a token 3/3", c.IsToken(), c.Power, c.Toughness)
		}
	}
	if p := me.ManaPool.Total(); p != 0 {
		t.Errorf("%d mana left, want the {2}{G} spent", p)
	}
}

func TestOldB447GraveyardShovelExilesAndGainsLifeForACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	shovel := effectstest.PushCatalogPermanent(g, me.ID, "Graveyard Shovel", "Artifact", oldB447Shovel, false)
	creature := oldB447Graveyard(opp, "Their Corpse", "Creature — Zombie", "{1}", 1, 1)
	other := oldB447Graveyard(opp, "Their Spell", "Sorcery", "{1}", 0, 0)
	effectstest.B06AddMana(me, "C", "C")
	before := me.Life
	if err := g.ActivateCatalogAbility(me.ID, shovel, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, opp.ID, creature)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(creature) || !opp.Graveyard.Contains(other) {
		t.Error("the target player's pick was not the card exiled")
	}
	if me.Life != before+2 {
		t.Errorf("life %d, want %d after exiling a creature card", me.Life, before+2)
	}
	// A noncreature card gains nothing.
	g.WithWriteLock(func() { shovelUntap(g, shovel) })
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, shovel, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("second activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, opp.ID, other)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+2 {
		t.Errorf("life %d after exiling a noncreature card, want unchanged %d", me.Life, before+2)
	}
}

func shovelUntap(g *game.Game, id uuid.UUID) {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].Tapped = false
		}
	}
}

func TestOldB447MakeAWishReturnsTwoCardsAtRandom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	var ids []uuid.UUID
	for i := 0; i < 5; i++ {
		ids = append(ids, oldB447Graveyard(me, "Card", "Sorcery", "{1}", 0, 0))
	}
	effectstest.B06AddMana(me, "G", "C", "C", "C")
	handBefore := len(me.Hand.Cards)
	if err := oldB447Cast(t, g, me, "Make a Wish", "Sorcery", "{3}{G}", oldB447Wish, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	back := 0
	for _, id := range ids {
		if me.Hand.Contains(id) {
			back++
		}
	}
	if back != 2 || len(me.Hand.Cards) != handBefore+2 {
		t.Errorf("%d graveyard cards returned (hand %d -> %d), want exactly 2", back, handBefore, len(me.Hand.Cards))
	}
}

func TestOldB447MakeAWishWithOneCardReturnsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	only := oldB447Graveyard(me, "Only", "Sorcery", "{1}", 0, 0)
	effectstest.B06AddMana(me, "G", "C", "C", "C")
	if err := oldB447Cast(t, g, me, "Make a Wish", "Sorcery", "{3}{G}", oldB447Wish, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(only) {
		t.Error("the only card was not returned")
	}
}

func TestOldB447GhoulcallersChantBothModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	bear := oldB447Graveyard(me, "Bear", "Creature — Bear", "{1}{G}", 2, 2)
	zombieA := oldB447Graveyard(me, "Zombie A", "Creature — Zombie", "{1}{B}", 2, 2)
	zombieB := oldB447Graveyard(me, "Zombie B", "Creature — Zombie", "{1}{B}", 2, 2)
	effectstest.B06AddMana(me, "B", "B")
	if err := oldB447Cast(t, g, me, "Ghoulcaller's Chant", "Sorcery", "{B}", oldB447Chant, game.CastSpellParams{Modes: []int{1},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: zombieA, Slot: 0}, {Kind: game.TargetCard, ID: bear, Slot: 1}}}); err == nil {
		t.Fatal("a non-Zombie was accepted for the Zombie mode")
	}
	if err := oldB447Cast(t, g, me, "Ghoulcaller's Chant", "Sorcery", "{B}", oldB447Chant, game.CastSpellParams{Modes: []int{1},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: zombieA, Slot: 0}, {Kind: game.TargetCard, ID: zombieB, Slot: 1}}}); err != nil {
		t.Fatalf("cast the Zombie mode: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(zombieA) || !me.Hand.Contains(zombieB) || me.Hand.Contains(bear) {
		t.Error("the two Zombies were not both returned alone")
	}
	if err := oldB447Cast(t, g, me, "Ghoulcaller's Chant", "Sorcery", "{B}", oldB447Chant, game.CastSpellParams{Modes: []int{0},
		Targets: effectstest.B16TargetCard(bear)}); err != nil {
		t.Fatalf("cast the creature mode: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Error("the creature card was not returned")
	}
}

func TestOldB447MawOfTheMireDestroysALandAndGainsFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.B12Permanent(g, opp.ID, "Swamp", "Basic Land — Swamp")
	creature := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B06AddMana(me, "B", "C", "C", "C", "C")
	before := me.Life
	if err := oldB447Cast(t, g, me, "Maw of the Mire", "Sorcery", "{4}{B}", oldB447Mire, game.CastSpellParams{Targets: effectstest.B16TargetCard(creature)}); err == nil {
		t.Fatal("a creature was accepted as the land target")
	}
	if err := oldB447Cast(t, g, me, "Maw of the Mire", "Sorcery", "{4}{B}", oldB447Mire, game.CastSpellParams{Targets: effectstest.B16TargetCard(land)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land was not destroyed")
	}
	if me.Life != before+4 {
		t.Errorf("life %d, want %d", me.Life, before+4)
	}
}

func TestOldB447IntoTheMawOfHellDestroysALandAndDealsThirteen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.B12Permanent(g, opp.ID, "Mountain", "Basic Land — Mountain")
	titan := effectstest.B12Creature(g, opp.ID, "Titan", "Creature — Giant", 13, 14)
	effectstest.B06AddMana(me, "R", "R", "C", "C", "C", "C")
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: land, Slot: 0}, {Kind: game.TargetCard, ID: titan, Slot: 1}}
	if err := oldB447Cast(t, g, me, "Into the Maw of Hell", "Sorcery", "{4}{R}{R}", oldB447MawHell, game.CastSpellParams{Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: titan, Slot: 0}, {Kind: game.TargetCard, ID: land, Slot: 1}}}); err == nil {
		t.Fatal("the targets were accepted in the wrong slots")
	}
	if err := oldB447Cast(t, g, me, "Into the Maw of Hell", "Sorcery", "{4}{R}{R}", oldB447MawHell, game.CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land was not destroyed")
	}
	if got := effectstest.Pr6Marked(g, titan); got != 13 {
		t.Errorf("the creature has %d damage, want 13", got)
	}
}

func TestOldB447WitchbaneOrbDestroysYourCursesAndGivesYouHexproof(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Curse on Me", TypeLine: "Enchantment — Aura Curse",
		Owner: opp.ID, Controller: opp.ID, AttachedTo: game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}})
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Curse on Them", TypeLine: "Enchantment — Aura Curse",
		Owner: me.ID, Controller: me.ID, AttachedTo: game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}})
	ordinary := effectstest.B12Permanent(g, me.ID, "Plain Aura", "Enchantment — Aura")
	effectstest.CastCatalogSpell(t, g, "Witchbane Orb", "Artifact", oldB447Orb, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Error("the Curse attached to me survived")
	}
	if !g.Battlefield.Contains(theirs) || !g.Battlefield.Contains(ordinary) {
		t.Error("a Curse on another player or a non-Curse was destroyed")
	}
	spec, _ := effects.Lookup(oldB447Orb)
	if len(spec.PlayerKeywords) != 1 || spec.PlayerKeywords[0] != effects.KeywordHexproof {
		t.Errorf("player keywords %v, want hexproof", spec.PlayerKeywords)
	}
}

func TestOldB447ParaseleneDestroysEveryEnchantmentAndGainsLifeForEach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Permanent(g, me.ID, "Mine", "Enchantment")
	b := effectstest.B12Permanent(g, opp.ID, "Theirs", "Enchantment")
	c := effectstest.B12Permanent(g, opp.ID, "Theirs Too", "Enchantment — Aura")
	keep := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B06AddMana(me, "W", "C", "C")
	before := me.Life
	if err := oldB447Cast(t, g, me, "Paraselene", "Sorcery", "{2}{W}", oldB447Paraselene, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c} {
		if g.Battlefield.Contains(id) {
			t.Error("an enchantment survived")
		}
	}
	if !g.Battlefield.Contains(keep) {
		t.Error("the creature was destroyed")
	}
	if me.Life != before+3 {
		t.Errorf("life %d, want %d (one per enchantment)", me.Life, before+3)
	}
}
