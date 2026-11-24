package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b015_cards_test.go — Secret Lair Drop, "Life gain and loss"
// (S82 OLD batch 15, #5462): each card driven through a real cast, activation or
// attack.

const (
	oldB015ElixirOracle   = "204b71fc-4d5f-40dd-903d-56c733171d84"
	oldB015KolaghanOracle = "5ff89859-ddfd-48e5-8423-57b3956173ac"
	oldB015MohiamOracle   = "6d920b96-eb3e-447a-b6c1-34c3d9cb1403"
	oldB015GreymondOracle = "739a5220-50cd-4148-be58-1bbdeef2954e"
	oldB015AthreosOracle  = "af206aed-1f58-4f5a-ba42-0e0d1f282940"
	oldB015PrinceOracle   = "e7f3b4bb-ebf5-4e59-8f0b-aa26f78e9edc"
	oldB015LeechingOracle = "f104b361-8000-4eda-b530-1502528e09f9"
	oldB015BatwingOracle  = "fc8247f5-f7b2-4fc4-a9ea-90babad5fa42"
	oldB015ScionOracle    = "fd547632-a29a-4dad-a9da-155882391815"
)

// oldB015PickPlayers answers an open pick_target prompt, if any, with the players named.
func oldB015PickPlayers(t *testing.T, g *game.Game, chooser uuid.UUID, players ...uuid.UUID) {
	t.Helper()
	p := effectstest.LatestPickTarget(g, chooser)
	if p == nil {
		return
	}
	var refs []game.TargetRef
	for _, id := range players {
		refs = append(refs, game.TargetRef{Kind: game.TargetPlayer, ID: id})
	}
	if err := g.ResolvePickTargets(p.ID, chooser, refs); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
}

func TestS82OldB015AllAreFullWithNoCaveats(t *testing.T) {
	for _, id := range []string{oldB015ElixirOracle, oldB015KolaghanOracle, oldB015MohiamOracle, oldB015GreymondOracle,
		oldB015AthreosOracle, oldB015PrinceOracle, oldB015LeechingOracle, oldB015BatwingOracle, oldB015ScionOracle} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", id)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// oldB015Ready untaps a battlefield permanent and clears its summoning sickness.
func oldB015Ready(g *game.Game, id uuid.UUID) {
	effectstest.UntapForTest(g, id)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
}

// --- Elixir of Immortality ---------------------------------------------------

func TestS82OldB015ElixirGainsFiveAndShufflesItselfAndTheGraveyardIn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	elixir := effectstest.SeedPermanentWithOracle(g, me.ID, "Elixir of Immortality", "Artifact", oldB015ElixirOracle)
	oldB015Ready(g, elixir)
	a := effectstest.PushGraveyardCardTyped(me, "Dead A", "Creature — Bear")
	b := effectstest.PushGraveyardCardTyped(me, "Dead B", "Sorcery")
	life, lib := me.Life, me.Library.Size()
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, elixir, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+5 {
		t.Errorf("life = %d, want %d", me.Life, life+5)
	}
	if g.Battlefield.Contains(elixir) {
		t.Error("the Elixir stayed on the battlefield")
	}
	if me.Graveyard.Size() != 0 || me.Graveyard.Contains(a) || me.Graveyard.Contains(b) {
		t.Errorf("the graveyard still holds %d cards", me.Graveyard.Size())
	}
	if me.Library.Size() != lib+3 {
		t.Errorf("library = %d, want %d (two graveyard cards and the Elixir)", me.Library.Size(), lib+3)
	}
}

// --- Leeching Sliver ---------------------------------------------------------

func TestS82OldB015LeechingSliverDrainsTheDefenderForEachAttackingSliver(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	leech := effectstest.B12Push(g, me.ID, "Leeching Sliver", "Creature — Sliver", oldB015LeechingOracle, 1, 1)
	other := effectstest.B12Creature(g, me.ID, "Other Sliver", "Creature — Sliver", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	life := opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, leech, other, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("defender life = %d, want %d (one per attacking Sliver, none for the Bear)", opp.Life, life-2)
	}
}

// --- Scion of Draco ----------------------------------------------------------

func TestS82OldB015ScionCostsTwoLessPerBasicLandType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	id := effectstest.HandCardFull(me, "Scion of Draco", "Artifact Creature — Dragon", "{12}", oldB015ScionOracle, nil)
	var card game.Card
	for _, c := range me.Hand.Cards {
		if c.InstanceID == id {
			card = c
		}
	}
	price := func() int {
		p, err := g.PriceCast(me.ID, card, game.CastSpellParams{})
		if err != nil {
			t.Fatalf("PriceCast: %v", err)
		}
		return p.Total.Generic
	}
	if got := price(); got != 12 {
		t.Fatalf("with no lands: %d, want 12", got)
	}
	effectstest.B12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	if got := price(); got != 8 {
		t.Errorf("with two basic land types: %d, want 8", got)
	}
}

func TestS82OldB015ScionGivesEachCreatureAKeywordForItsColours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Scion of Draco", "Artifact Creature — Dragon", oldB015ScionOracle, 4, 4)
	mk := func(owner uuid.UUID, name string, colors ...string) uuid.UUID {
		return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name,
			TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: colors, Owner: owner, Controller: owner})
	}
	white, blue, black := mk(me.ID, "White", "W"), mk(me.ID, "Blue", "U"), mk(me.ID, "Black", "B")
	red, green, wb := mk(me.ID, "Red", "R"), mk(me.ID, "Green", "G"), mk(me.ID, "WhiteBlack", "W", "B")
	theirs := mk(opp.ID, "TheirWhite", "W")
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	for _, tc := range []struct {
		id  uuid.UUID
		kw  string
		has bool
	}{{white, "vigilance", true}, {white, "lifelink", false}, {blue, "hexproof", true}, {black, "lifelink", true},
		{red, "first strike", true}, {green, "trample", true}, {green, "vigilance", false},
		{wb, "vigilance", true}, {wb, "lifelink", true}, {theirs, "vigilance", false}} {
		if got := effectstest.EffectiveAbilitiesContain(t, g, tc.id, tc.kw); got != tc.has {
			t.Errorf("%v has %s = %v, want %v", tc.id, tc.kw, got, tc.has)
		}
	}
}

// --- Batwing Brume -----------------------------------------------------------

// oldB015Brume casts Batwing Brume paying the named mana, with `attackers`
// of mine attacking, and returns my attacker, the opponent's creature and the seats.
func oldB015Brume(t *testing.T, g *game.Game, pool string, attackers int) (me, opp *game.Player, mine, theirs uuid.UUID) {
	t.Helper()
	me, opp = effectstest.SeatsForCombat(g)
	mine = effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	theirs = effectstest.B12Creature(g, opp.ID, "Defender", "Creature — Bear", 2, 5)
	var attack []uuid.UUID
	for i := 0; i < attackers; i++ {
		attack = append(attack, effectstest.B12Creature(g, me.ID, "Extra", "Creature — Bear", 1, 1))
	}
	attack = append(attack, mine)
	effectstest.DeclareAttack(t, g, opp.ID, attack...)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Batwing Brume", TypeLine: "Instant", OracleID: oldB015BatwingOracle,
		ManaCost: "{1}{W/B}", Colors: []string{"W", "B"}, Owner: me.ID, Controller: me.ID})
	effectstest.AddPool(t, g, me.ID, pool)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return me, opp, mine, theirs
}

func TestS82OldB015BatwingBrumeWhitePreventsCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, mine, theirs := oldB015Brume(t, g, "{C}{W}", 1)
	if err := g.MarkCombatDamage(mine, theirs, 2); err != nil {
		t.Fatalf("MarkCombatDamage: %v", err)
	}
	if got := effectstest.DamageMarkedOn(g, theirs); got != 0 {
		t.Errorf("combat damage marked = %d, want 0", got)
	}
	if me.Life != 40 || opp.Life != 40 {
		t.Errorf("life %d/%d, want no change from the white half", me.Life, opp.Life)
	}
}

func TestS82OldB015BatwingBrumeBlackMakesEachPlayerLoseLifePerAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, mine, theirs := oldB015Brume(t, g, "{C}{B}", 1)
	if me.Life != 40-2 {
		t.Errorf("my life = %d, want %d (two attackers)", me.Life, 40-2)
	}
	if opp.Life != 40 {
		t.Errorf("the opponent's life = %d, want 40 (no attackers)", opp.Life)
	}
	if err := g.MarkCombatDamage(mine, theirs, 2); err != nil {
		t.Fatalf("MarkCombatDamage: %v", err)
	}
	if got := effectstest.DamageMarkedOn(g, theirs); got != 2 {
		t.Errorf("combat damage marked = %d, want 2 (black does not prevent it)", got)
	}
}

func TestS82OldB015BatwingBrumeWhiteAndBlackDoBoth(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, mine, theirs := oldB015Brume(t, g, "{W}{B}", 0)
	if me.Life != 40-1 {
		t.Errorf("my life = %d, want 39", me.Life)
	}
	if err := g.MarkCombatDamage(mine, theirs, 2); err != nil {
		t.Fatalf("MarkCombatDamage: %v", err)
	}
	if got := effectstest.DamageMarkedOn(g, theirs); got != 0 {
		t.Errorf("combat damage marked = %d, want 0", got)
	}
}

// --- Dragonlord Kolaghan -----------------------------------------------------

// oldB015CastFlash casts a free creature spell with flash named `name` for p.
func oldB015CastFlash(t *testing.T, g *game.Game, p *game.Player, name string) {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Bear", Keywords: []string{"flash"},
		Power: 1, Toughness: 1, Owner: p.ID, Controller: p.ID})
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
}

func TestS82OldB015KolaghanGivesHasteAndPunishesARepeatedName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	dragon := effectstest.B12Push(g, me.ID, "Dragonlord Kolaghan", "Legendary Creature — Elder Dragon", oldB015KolaghanOracle, 6, 5)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if !effectstest.EffectiveAbilitiesContain(t, g, bear, "haste") {
		t.Error("other creatures you control lack haste")
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, dragon, "haste") {
		t.Error("Kolaghan lacks haste")
	}
	effectstest.PushGraveyardCardTyped(opp, "Twin", "Creature — Bear")
	// A different name: nothing.
	life := opp.Life
	oldB015CastFlash(t, g, opp, "Other")
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("a different name cost %d life", life-opp.Life)
	}
	// The same name as a card in their own graveyard: 10 life.
	oldB015CastFlash(t, g, opp, "Twin")
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-10 {
		t.Errorf("life = %d, want %d", opp.Life, life-10)
	}
	// Your own repeated name does not trigger it.
	effectstest.PushGraveyardCardTyped(me, "Mine", "Creature — Bear")
	mylife := me.Life
	oldB015CastFlash(t, g, me, "Mine")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != mylife {
		t.Errorf("my own repeated name cost me %d life", mylife-me.Life)
	}
}

// --- Greymond, Avacyn's Stalwart ---------------------------------------------

func TestS82OldB015GreymondGrantsTheChosenPairToHumansAndPumpsAtFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	human := effectstest.B12Creature(g, me.ID, "Human One", "Creature — Human Soldier", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Greymond, Avacyn's Stalwart", TypeLine: "Legendary Creature — Human Soldier",
		OracleID: oldB015GreymondOracle, Power: 3, Toughness: 4, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID)
	if c == nil || len(c.PickOptions) != 3 {
		t.Fatalf("no three-way as-enters prompt: %+v", c)
	}
	// "first strike and lifelink"
	if err := g.ResolveOptionPick(c.ID, me.ID, 1); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	for _, tc := range []struct {
		who  uuid.UUID
		kw   string
		want bool
	}{{human, "first strike", true}, {human, "lifelink", true}, {human, "vigilance", false},
		{id, "first strike", true}, {bear, "first strike", false}} {
		if got := effectstest.EffectiveAbilitiesContain(t, g, tc.who, tc.kw); got != tc.want {
			t.Errorf("%v has %s = %v, want %v", tc.who, tc.kw, got, tc.want)
		}
	}
	// Two Humans (Greymond and one): no pump yet.
	if p, _ := effectstest.EffectivePTOf(t, g, human); p != 1 {
		t.Errorf("a Human has power %d with two Humans, want 1", p)
	}
	effectstest.B12Creature(g, me.ID, "Human Two", "Creature — Human", 1, 1)
	effectstest.B12Creature(g, me.ID, "Human Three", "Creature — Human", 1, 1)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if p, tough := effectstest.EffectivePTOf(t, g, human); p != 3 || tough != 3 {
		t.Errorf("a Human is %d/%d with four Humans, want 3/3", p, tough)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, bear); p != 2 {
		t.Errorf("a Bear has power %d, want 2", p)
	}
}

// --- Athreos, God of Passage -------------------------------------------------

func TestS82OldB015AthreosReturnsAnOwnedCreatureUnlessTheOpponentPays(t *testing.T) {
	for _, tc := range []struct {
		name string
		pay  bool
	}{{"the opponent lets it return", false}, {"the opponent pays 3 life", true}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			effectstest.B12Push(g, me.ID, "Athreos, God of Passage", "Legendary Enchantment Creature — God", oldB015AthreosOracle, 5, 4)
			victim := effectstest.B12Creature(g, me.ID, "Victim", "Creature — Bear", 2, 2)
			life := opp.Life
			effectstest.KillCreature(t, g, me.ID, victim)
			effectstest.PassPriorityAroundTable(t, g)
			oldB015PickPlayers(t, g, me.ID, opp.ID)
			effectstest.PassPriorityAroundTable(t, g)
			pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOptionPick, opp.ID)
			if pick == nil {
				t.Fatal("the opponent was not asked")
			}
			idx := 0
			if tc.pay {
				idx = 1
			}
			if err := g.ResolveOptionPick(pick.ID, opp.ID, idx); err != nil {
				t.Fatalf("answer: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if tc.pay {
				if opp.Life != life-3 || !me.Graveyard.Contains(victim) {
					t.Errorf("after paying: life %d (want %d), in graveyard %v", opp.Life, life-3, me.Graveyard.Contains(victim))
				}
				return
			}
			if opp.Life != life || !me.Hand.Contains(victim) {
				t.Errorf("after declining: life %d (want %d), in hand %v", opp.Life, life, me.Hand.Contains(victim))
			}
		})
	}
}

func TestS82OldB015AthreosIgnoresAnotherPlayersCreatureAndIsNoCreatureBelowSevenDevotion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ath := effectstest.B12Push(g, me.ID, "Athreos, God of Passage", "Legendary Enchantment Creature — God", oldB015AthreosOracle, 5, 4)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if c := effectstest.ApaLive(g, ath); c.IsCreature() {
		t.Error("Athreos is a creature with no devotion")
	}
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOptionPick, opp.ID) != nil {
		t.Error("another player's creature dying triggered Athreos")
	}
}

// --- Prince of Thralls -------------------------------------------------------

func TestS82OldB015PrinceOfThrallsTakesAnOpponentsPermanentUnlessTheyPay(t *testing.T) {
	for _, tc := range []struct {
		name string
		pay  bool
	}{{"the opponent lets it go", false}, {"the opponent pays 3 life", true}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			effectstest.B12Push(g, me.ID, "Prince of Thralls", "Creature — Demon", oldB015PrinceOracle, 7, 7)
			victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
			life := opp.Life
			effectstest.KillCreature(t, g, opp.ID, victim)
			effectstest.PassPriorityAroundTable(t, g)
			pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOptionPick, opp.ID)
			if pick == nil {
				t.Fatal("the opponent was not asked")
			}
			idx := 0
			if tc.pay {
				idx = 1
			}
			if err := g.ResolveOptionPick(pick.ID, opp.ID, idx); err != nil {
				t.Fatalf("answer: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			c, onField := effectstest.BattlefieldCardByID(g, victim)
			if tc.pay {
				if opp.Life != life-3 || onField {
					t.Errorf("after paying: life %d (want %d), on battlefield %v", opp.Life, life-3, onField)
				}
				return
			}
			if !onField || c.Controller != me.ID || opp.Life != life {
				t.Errorf("after declining: on battlefield %v under %v, life %d", onField, c.Controller, opp.Life)
			}
		})
	}
}

func TestS82OldB015PrinceOfThrallsIgnoresYourOwnPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Prince of Thralls", "Creature — Demon", oldB015PrinceOracle, 7, 7)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOptionPick, p.ID) != nil {
			t.Fatalf("a prompt was opened for %v after my own creature died", p.ID)
		}
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("my creature did not stay in my graveyard")
	}
}

// --- Mohiam, Imperial Truthsayer ---------------------------------------------

func TestS82OldB015MohiamLooksAtTheTopCardAndMayMillIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mo := effectstest.B12Push(g, me.ID, "Mohiam, Imperial Truthsayer", "Legendary Creature — Human Advisor", oldB015MohiamOracle, 1, 3)
	oldB015Ready(g, mo)
	top := effectstest.SeedLibraryTop(opp, "Top", "Sorcery", "{1}")
	if err := g.ActivateCatalogAbility(me.ID, mo, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(top) {
		t.Error("the top card was not put into its owner's graveyard")
	}
}

func TestS82OldB015MohiamEndStepExilesAndPlotsTheNonlandCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Mohiam, Imperial Truthsayer", "Legendary Creature — Human Advisor", oldB015MohiamOracle, 1, 3)
	spell := effectstest.SeedLibraryTop(opp, "Big Spell", "Sorcery", "{3}{R}")
	land := effectstest.SeedLibraryTop(third, "Island", "Basic Land — Island", "")
	oppLife, thirdLife := opp.Life, third.Life
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	oldB015PickPlayers(t, g, me.ID, opp.ID, third.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOf(g, spell) != game.ZoneExile || effectstest.ZoneOf(g, land) != game.ZoneExile {
		t.Fatalf("zones: spell %q, land %q, want both exiled", effectstest.ZoneOf(g, spell), effectstest.ZoneOf(g, land))
	}
	if opp.Life != oppLife-4 {
		t.Errorf("the owner of the {3}{R} card has %d life, want %d", opp.Life, oppLife-4)
	}
	if third.Life != thirdLife {
		t.Errorf("a land cost its owner %d life", thirdLife-third.Life)
	}
}
