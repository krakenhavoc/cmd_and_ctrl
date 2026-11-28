package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b670_cards_test.go — Prophecy: Damage, batch 670 (#5617). Each test
// plays the real catalog entry and asserts what its text says.

const (
	b670Apparition = "13cfc92c-142a-4016-b86f-ace6bda1a69c"
	b670Citadel    = "2233c079-5ca6-4992-8bd2-84dbf9ed53a9"
	b670Death      = "491606d8-e9c1-4f89-bbef-b7590d310097"
	b670Might      = "68d84ffa-5bd0-4ed9-a987-6cef13095438"
	b670Latulla    = "86b2879b-6c54-4d96-ab81-a3f2c87d42d9"
	b670Wing       = "c593f66d-0e9a-4d46-a9c6-ac7a70bf17ea"
	b670Fiend      = "d287c11e-0684-42ed-abdb-758852ed0502"
	b670Soul       = "de56a8d9-faeb-4e73-bdc1-cc81a6b6d05c"
)

func TestB670CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		b670Apparition: "Chilling Apparition", b670Citadel: "Citadel of Pain", b670Death: "Death Charmer",
		b670Might: "Avatar of Might", b670Latulla: "Latulla, Keldon Overseer", b670Wing: "Wing Storm",
		b670Fiend: "Plague Fiend", b670Soul: "Soul Charmer",
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

func b670Table(t *testing.T) (*game.Game, *game.Player, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp, other := effectstest.BlbB1Seats(g)
	effectstest.AdvanceToMain(t, g)
	return g, me, opp, other
}

func TestB670ChillingApparitionRegeneratesAndMakesThemDiscard(t *testing.T) {
	g, me, opp, _ := b670Table(t)
	fiend := effectstest.PushCatalogPermanent(g, me.ID, "Chilling Apparition", "Creature — Spirit", b670Apparition, false)
	effectstest.GiveHandCards(opp, 2)
	hand := len(opp.Hand.Cards)
	effectstest.DealCombatDamage(g, me.ID, fiend, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardOwed(g, opp.ID) != 1 {
		t.Fatal("the damaged player was not asked to discard a card")
	}
	effectstest.DiscardFromHand(t, g, opp.ID)
	if len(opp.Hand.Cards) != hand-1 {
		t.Errorf("hand %d, want %d", len(opp.Hand.Cards), hand-1)
	}
	effectstest.FloatForTest(g, me, "B")
	effectstest.ApaActivate(t, g, me, fiend, 0, game.ActivateAbilityParams{})
	effectstest.P1Destroy(t, g, fiend)
	if !effectstest.BattlefieldHas(g, fiend) {
		t.Error("the regeneration shield did not save it")
	}
}

func TestB670CitadelOfPainDamagesEachPlayerForTheirUntappedLands(t *testing.T) {
	g, me, opp, _ := b670Table(t)
	effectstest.PushCatalogPermanent(g, me.ID, "Citadel of Pain", "Enchantment", b670Citadel, false)
	for i := 0; i < 3; i++ {
		effectstest.GaPut(g, opp.ID, "Plains", "Basic Land — Plains", "")
	}
	tapped := effectstest.GaPut(g, opp.ID, "Island", "Basic Land — Island", "")
	effectstest.GaPut(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	myLife, theirLife := me.Life, opp.Life
	effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife-1 || opp.Life != theirLife {
		t.Errorf("my end step: life %d/%d, want %d/%d", me.Life, opp.Life, myLife-1, theirLife)
	}
	next := (g.Turn.ActiveSeat + 1) % len(g.Seats)
	effectstest.AdvanceToMainOf(t, g, next)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, tapped).Tapped = true })
	effectstest.AdvanceToEndStepOf(t, g, next)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife-1 || opp.Life != theirLife-3 {
		t.Errorf("their end step: life %d/%d, want %d/%d", me.Life, opp.Life, myLife-1, theirLife-3)
	}
}

// b670Charmer plays the named Charmer, hits the opponent's creature with
// combat damage and answers the {2} prompt.
func b670Charmer(t *testing.T, name, typeLine, oracle string, pay bool) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g, me, opp, _ := b670Table(t)
	charmer := effectstest.PushCatalogPermanent(g, me.ID, name, typeLine, oracle, false)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	effectstest.DealCombatDamage(g, me.ID, charmer, victim)
	effectstest.PassUntilTaxed(t, g, opp.ID)
	if pay {
		opp.ManaPool.AddMana(game.ManaToken{Color: "C"})
		opp.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
	effectstest.AnswerPayUnless(t, g, opp.ID, pay)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	return g, me, opp, victim
}

func TestB670DeathCharmerMakesTheCreaturesControllerLoseLifeUnlessTheyPay(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g, _, opp, _ := b670Charmer(t, "Death Charmer", "Creature — Worm Mercenary", b670Death, pay)
		want := 40
		if !pay {
			want = 38
		}
		if opp.Life != want {
			t.Errorf("pay=%v: life %d, want %d (starting %d)", pay, opp.Life, want, 40)
		}
		_ = g
	}
}

func TestB670PlagueFiendDestroysTheCreatureUnlessItsControllerPays(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g, _, _, victim := b670Charmer(t, "Plague Fiend", "Creature — Insect", b670Fiend, pay)
		if survived := effectstest.BattlefieldHas(g, victim); survived != pay {
			t.Errorf("pay=%v: the creature survived=%v", pay, survived)
		}
	}
}

func TestB670SoulCharmerGainsLifeUnlessTheCreaturesControllerPays(t *testing.T) {
	for _, pay := range []bool{false, true} {
		_, me, _, _ := b670Charmer(t, "Soul Charmer", "Creature — Human Rebel", b670Soul, pay)
		want := 40
		if !pay {
			want = 42
		}
		if me.Life != want {
			t.Errorf("pay=%v: life %d, want %d", pay, me.Life, want)
		}
	}
}

func TestB670CharmersIgnorePlayersAndNoncombatDamage(t *testing.T) {
	g, me, opp, _ := b670Table(t)
	charmer := effectstest.PushCatalogPermanent(g, me.ID, "Death Charmer", "Creature — Worm Mercenary", b670Death, false)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	effectstest.DealCombatDamage(g, me.ID, charmer, opp.ID)
	effectstest.DealNoncombatDamage(g, me.ID, charmer, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Error("asked to pay for damage to a player or noncombat damage")
	}
}

func TestB670AvatarOfMightCostsLessOnlyWhenAnOpponentHasFourMoreCreatures(t *testing.T) {
	for _, theirs := range []int{4, 5} {
		g, me, opp, _ := b670Table(t)
		effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
		for i := 0; i < theirs; i++ {
			effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
		}
		id := effectstest.HandCardFull(me, "Avatar of Might", "Creature — Avatar", "{6}{G}{G}", b670Might, []string{"G"})
		effectstest.FloatForTest(g, me, "GG")
		err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
		if theirs == 4 && err == nil {
			t.Error("three more creatures: {G}{G} paid for an eight-mana spell")
		}
		if theirs == 5 && err != nil {
			t.Errorf("four more creatures: {G}{G} was refused: %v", err)
		}
	}
}

func TestB670LatullaDealsXDamageForTwoDiscards(t *testing.T) {
	g, me, opp, _ := b670Table(t)
	latulla := effectstest.PushCatalogPermanent(g, me.ID, "Latulla, Keldon Overseer", "Legendary Creature — Human Spellshaper", b670Latulla, false)
	effectstest.GiveHandCards(me, 3)
	handBefore := len(me.Hand.Cards)
	cards := []uuid.UUID{me.Hand.Cards[0].InstanceID, me.Hand.Cards[1].InstanceID}
	life := opp.Life
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	effectstest.FloatForTest(g, me, "RR")
	if err := g.ActivateCatalogAbility(me.ID, latulla, 0, game.ActivateAbilityParams{Strict: true, XValue: 1,
		DiscardIDs: cards[:1], Targets: target}); err == nil {
		t.Fatal("one discard paid for a two-discard cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, latulla, 0, game.ActivateAbilityParams{Strict: true, XValue: 1,
		DiscardIDs: cards, Targets: target}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life %d, want %d", opp.Life, life-1)
	}
	if len(me.Hand.Cards) != handBefore-2 || !effectstest.PermanentTapped(g, latulla) {
		t.Errorf("hand %d (want %d), tapped %v", len(me.Hand.Cards), handBefore-2, effectstest.PermanentTapped(g, latulla))
	}
}

func TestB670WingStormDamagesEachPlayerForTheirFliers(t *testing.T) {
	g, me, opp, other := b670Table(t)
	effectstest.ReqCreature(g, opp.ID, "Bird", "", 1, 1, "flying")
	effectstest.ReqCreature(g, opp.ID, "Bird 2", "", 1, 1, "flying")
	effectstest.ReqCreature(g, opp.ID, "Bear", "", 2, 2)
	effectstest.ReqCreature(g, me.ID, "Bird 3", "", 1, 1, "flying")
	myLife, theirLife, otherLife := me.Life, opp.Life, other.Life
	effectstest.CastCatalogSpell(t, g, "Wing Storm", "Sorcery", b670Wing, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife-2 || opp.Life != theirLife-4 || other.Life != otherLife {
		t.Errorf("life %d/%d/%d, want %d/%d/%d", me.Life, opp.Life, other.Life, myLife-2, theirLife-4, otherLife)
	}
}
