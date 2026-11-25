package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b097_cards_test.go — Zendikar Rising, S82 older sets batch 097
// (#5508).

const (
	znrNissasZendikon = "1b8c4cbe-b2fc-4d52-9382-1e03c40003be"
	znrNahirisBinding = "8f674382-78b5-462c-9a7f-c5f613c317d4"
	znrBaseCamp       = "41fbf835-baee-4530-9155-e2c1b9045567"
	znrThrone         = "7e8198e9-0f3b-420b-ab09-74f13f4fd548"
	znrRelicRobber    = "85eafa00-51df-4070-9b2d-247457ba3364"
	znrGrakmaw        = "9c93acf3-d3da-4b4b-947f-ec8fc99d0d34"
	znrMindDrain      = "0a88d1b1-ac3c-41da-b3f5-b689af0e45d7"
	znrFamiliar       = "2a56fe0b-f8a9-499c-bce1-cec6cb3b66f0"
	znrNahiri         = "39bbc4cf-53fe-47dd-8ba3-1a59b1fa7dd1"
	znrRiptide        = "e682ad51-001c-4465-951e-43df96b7d3b9"
	znrColossus       = "ef7f6766-e878-40c6-aa9a-da7415cc757a"
	znrOrah           = "fe348f31-4530-4c59-9ee2-6839eab973da"

	// A kicker creature from the catalog (Kicker {2}), used as the spell cast.
	znrKickerSerpent = "c39c1604-3bae-454d-9985-85101e51ec6e"
)

func TestS82OldB097EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		znrNissasZendikon: "Nissa's Zendikon", znrNahirisBinding: "Nahiri's Binding", znrBaseCamp: "Base Camp",
		znrThrone: "Throne of Makindi", znrRelicRobber: "Relic Robber", znrGrakmaw: "Grakmaw, Skyclave Ravager",
		znrMindDrain: "Mind Drain", znrFamiliar: "Lullmage's Familiar", znrNahiri: "Nahiri, Heir of the Ancients",
		znrRiptide: "Risen Riptide", znrColossus: "Sea Gate Colossus", znrOrah: "Orah, Skyclave Hierophant",
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

func znrLand(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: owner, Controller: owner})
}

func TestS82OldB097NissasZendikonAnimatesTheLandAndReturnsItWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := znrLand(g, me.ID)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Nissa's Zendikon", effectstest.AuraTypeLine, znrNissasZendikon, effectstest.CardRefs(land))
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.B12Card(t, g, land)
	if !c.IsLand() || !c.IsCreature() || !c.HasSubtype("Elemental") {
		t.Fatalf("the enchanted land is %q, want a land Elemental creature", c.TypeLine)
	}
	if p, tt := effectstest.PtOf(t, g, land); p != 4 || tt != 4 {
		t.Errorf("the land is %d/%d, want 4/4", p, tt)
	}
	for _, kw := range []string{"reach", "haste"} {
		if !effectstest.HasEffectiveKeyword(t, g, land, kw) {
			t.Errorf("the land has no %s", kw)
		}
	}
	effectstest.KillCreature(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(land) {
		t.Error("the dead land did not return to its owner's hand")
	}
}

func TestS82OldB097NissasZendikonEnchantsOnlyALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	card := game.Card{InstanceID: uuid.New(), Name: "Nissa's Zendikon", TypeLine: effectstest.AuraTypeLine, OracleID: znrNissasZendikon, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(card)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Error("Nissa's Zendikon was cast on a creature")
	}
}

func TestS82OldB097NahirisBindingLocksACreatureOrPlaneswalkerOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	land := znrLand(g, me.ID)
	if err := effectstest.CastCatalogSpellErr(t, g, "Nahiri's Binding", effectstest.AuraTypeLine, znrNahirisBinding, effectstest.CardRefs(land)); err == nil {
		t.Error("the Aura was cast on a land")
	}
	effectstest.CastCatalogSpell(t, g, "Nahiri's Binding", effectstest.AuraTypeLine, znrNahirisBinding, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	got := effectstest.RestrictionsOf(t, g, bear)
	want := game.CantAttackOrBlock | game.CantActivate
	if got&want != want {
		t.Errorf("restrictions = %v, want can't attack or block and can't activate", got.Names())
	}
	walker := effectstest.PushWalkerForTest(g, me.ID, "Test Walker", "", 4)
	effectstest.CastCatalogSpell(t, g, "Nahiri's Binding", effectstest.AuraTypeLine, znrNahirisBinding, effectstest.CardRefs(walker))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.RestrictionsOf(t, g, walker)&game.CantActivate == 0 {
		t.Error("the enchanted planeswalker can still activate its abilities")
	}
}

func TestS82OldB097MindDrainDiscardsMillsDrainsAndGains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 3; i++ {
		effectstest.HandSpell(opp, "Card", "Sorcery", "{9}")
	}
	hand, lib, gy := opp.Hand.Size(), opp.Library.Size(), opp.Graveyard.Size()
	myLife, oppLife := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "Mind Drain", "Sorcery", znrMindDrain, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, opp.ID) == nil {
		t.Fatal("the opponent was not asked to discard")
	}
	effectstest.DiscardFromHand(t, g, opp.ID)
	g.SettleResolution()
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != hand-2 {
		t.Errorf("hand %d, want %d", opp.Hand.Size(), hand-2)
	}
	if opp.Library.Size() != lib-1 {
		t.Errorf("library %d, want %d", opp.Library.Size(), lib-1)
	}
	if opp.Graveyard.Size() != gy+3 {
		t.Errorf("graveyard %d, want %d (two discards and a mill)", opp.Graveyard.Size(), gy+3)
	}
	if opp.Life != oppLife-1 || me.Life != myLife+1 {
		t.Errorf("life: opponent %d (want %d), you %d (want %d)", opp.Life, oppLife-1, me.Life, myLife+1)
	}
}

func TestS82OldB097MindDrainTargetsOnlyAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	card := game.Card{InstanceID: uuid.New(), Name: "Mind Drain", TypeLine: "Sorcery", OracleID: znrMindDrain, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(card)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err == nil {
		t.Error("Mind Drain was cast targeting its own caster")
	}
}

func TestS82OldB097RelicRobberGivesTheVictimAConstructThatPingsThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	robber := effectstest.B12Push(g, me.ID, "Relic Robber", "Creature — Goblin Rogue", znrRelicRobber, 2, 2)
	effectstest.AttackWith(t, g, opp.ID, robber)
	effectstest.PassPriorityAroundTable(t, g)
	construct := effectstest.FindBattlefieldByName(g, "Goblin Construct")
	if construct == uuid.Nil {
		t.Fatal("no Goblin Construct token")
	}
	c := effectstest.B12Card(t, g, construct)
	if c.Controller != opp.ID {
		t.Errorf("the Construct is controlled by %v, want the damaged player", c.Controller)
	}
	if !c.IsArtifact() || !c.IsCreature() || !c.IsColorless() {
		t.Errorf("token %q is not a colorless artifact creature", c.TypeLine)
	}
	if p, tt := effectstest.PtOf(t, g, construct); p != 0 || tt != 1 {
		t.Errorf("the Construct is %d/%d, want 0/1", p, tt)
	}
	if effectstest.RestrictionsOf(t, g, construct)&game.CantBlock == 0 {
		t.Error("the Construct can block")
	}
}

func TestS82OldB097GrakmawEntersWithCountersGrowsAndLeavesAHydra(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	grakmaw := effectstest.CastCatalogSpell(t, g, "Grakmaw, Skyclave Ravager", "Legendary Creature — Hydra Horror", znrGrakmaw, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.B12Card(t, g, grakmaw); c.Counters[game.CounterPlusOne] != 3 {
		t.Fatalf("Grakmaw entered with %d +1/+1 counters, want 3", c.Counters[game.CounterPlusOne])
	}
	if p, tt := effectstest.PtOf(t, g, grakmaw); p != 3 || tt != 3 {
		t.Errorf("Grakmaw is %d/%d, want 3/3", p, tt)
	}
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, me.ID, plain)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.B12Card(t, g, grakmaw); c.Counters[game.CounterPlusOne] != 3 {
		t.Errorf("a creature with no counter died and Grakmaw has %d counters, want 3", c.Counters[game.CounterPlusOne])
	}
	counted := effectstest.B12Creature(g, me.ID, "Counted", "Creature — Bear", 1, 1)
	effectstest.FindBattlefieldCardForTest(g, counted).Counters = map[string]int{game.CounterPlusOne: 1}
	effectstest.KillCreature(t, g, me.ID, counted)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.B12Card(t, g, grakmaw); c.Counters[game.CounterPlusOne] != 4 {
		t.Fatalf("a creature with a +1/+1 counter died and Grakmaw has %d counters, want 4", c.Counters[game.CounterPlusOne])
	}
	effectstest.KillCreature(t, g, me.ID, grakmaw)
	effectstest.PassPriorityAroundTable(t, g)
	hydra := effectstest.FindBattlefieldByName(g, "Hydra")
	if hydra == uuid.Nil {
		t.Fatal("no Hydra token")
	}
	c := effectstest.B12Card(t, g, hydra)
	if !c.IsToken() || len(c.Colors) != 2 {
		t.Errorf("Hydra token %+v is not black and green", c)
	}
	if p, tt := effectstest.PtOf(t, g, hydra); p != 4 || tt != 4 {
		t.Errorf("the Hydra is %d/%d, want 4/4", p, tt)
	}
}

func znrCast(t *testing.T, g *game.Game, me *game.Player, kicked bool, cost string) error {
	t.Helper()
	var opt []int
	if kicked {
		opt = []int{0}
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Test Serpent", TypeLine: "Creature — Serpent", ManaCost: cost,
		OracleID: znrKickerSerpent, Owner: me.ID, Controller: me.ID})
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, OptionalCosts: opt})
}

func TestS82OldB097KickedCastGivesLullmagesFamiliarsControllerTwoLife(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		effectstest.PushCatalogPermanent(g, me.ID, "Lullmage's Familiar", "Creature — Beast", znrFamiliar, false)
		life := me.Life
		effectstest.AddPool(t, g, me.ID, "{U}{U}{U}{U}")
		if err := znrCast(t, g, me, kicked, "{U}"); err != nil {
			t.Fatalf("kicked=%v: cast: %v", kicked, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		want := life
		if kicked {
			want = life + 2
		}
		if me.Life != want {
			t.Errorf("kicked=%v: life %d, want %d", kicked, me.Life, want)
		}
	}
}

func TestS82OldB097KickedCastMakesRisenRiptideFiveFiveUntilEndOfTurn(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		riptide := effectstest.B12Push(g, me.ID, "Risen Riptide", "Creature — Elemental", znrRiptide, 0, 5)
		effectstest.AddPool(t, g, me.ID, "{U}{U}{U}{U}")
		if err := znrCast(t, g, me, kicked, "{U}"); err != nil {
			t.Fatalf("kicked=%v: cast: %v", kicked, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		wantP := 0
		if kicked {
			wantP = 5
		}
		if p, tt := effectstest.PtOf(t, g, riptide); p != wantP || tt != 5 {
			t.Errorf("kicked=%v: Risen Riptide is %d/%d, want %d/5", kicked, p, tt, wantP)
		}
		if kicked {
			effectstest.EndTurn(t, g)
			if p, _ := effectstest.PtOf(t, g, riptide); p != 0 {
				t.Errorf("the 5/5 outlasted the turn: power %d", p)
			}
		}
	}
}

func TestS82OldB097LullmagesFamiliarTapsForGreenOrBlue(t *testing.T) {
	spec, _ := effects.Lookup(znrFamiliar)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{G|U}" {
		t.Errorf("mana abilities %+v, want one that adds {G} or {U}", spec.ManaAbilities)
	}
}

func TestS82OldB097ThroneOfMakindiChargesAndPaysOnlyForKickedSpells(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g, me, _ := effectstest.SpendTable(t)
		var thrones []uuid.UUID
		for i := 0; i < 2; i++ {
			id := effectstest.PushCatalogPermanent(g, me.ID, "Throne of Makindi", "Land", znrThrone, false)
			effectstest.FindBattlefieldCardForTest(g, id).Counters = map[string]int{game.CounterCharge: 1}
			thrones = append(thrones, id)
		}
		for _, id := range thrones {
			effectstest.ActivateManaFor(t, g, me.ID, id, 1, game.ManaAbilityParams{})
			if effectstest.LatestChoiceOfKind(g, game.PendingChoiceMana) != nil {
				effectstest.ResolveManaPickAnyColor(t, g, me.ID)
			}
			if c := effectstest.B12Card(t, g, id); c.Counters[game.CounterCharge] != 0 {
				t.Errorf("a charge counter was not removed as a cost")
			}
		}
		err := znrCast(t, g, me, kicked, "{1}")
		if kicked && err != nil {
			t.Errorf("the mana could not pay for a kicked spell: %v", err)
		}
		if !kicked {
			effectstest.RefusedForMana(t, err, "the Throne's mana for an unkicked spell")
		}
	}
}

func TestS82OldB097ThroneOfMakindiPutsAChargeCounterOnItself(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	id := effectstest.PushCatalogPermanent(g, me.ID, "Throne of Makindi", "Land", znrThrone, false)
	effectstest.AddPool(t, g, me.ID, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.B12Card(t, g, id); c.Counters[game.CounterCharge] != 1 || !c.Tapped {
		t.Errorf("charge counters %d tapped %v, want 1 and tapped", c.Counters[game.CounterCharge], c.Tapped)
	}
}

func TestS82OldB097BaseCampEntersTappedAndRestrictsItsColoredMana(t *testing.T) {
	spec, _ := effects.Lookup(znrBaseCamp)
	if len(spec.ManaAbilities) != 2 || len(spec.ManaAbilities[0].Restrictions) != 0 || len(spec.ManaAbilities[1].Restrictions) != 1 {
		t.Fatalf("Base Camp's mana abilities are %+v, want an unrestricted {C} and one restricted any-colour", spec.ManaAbilities)
	}
	if len(spec.Replacements) != 1 {
		t.Error("Base Camp has no enters-tapped replacement")
	}
	for _, tc := range []struct {
		name, subtype string
		ok            bool
	}{
		{"cleric", "Cleric", true}, {"rogue", "Rogue", true}, {"warrior", "Warrior", true}, {"wizard", "Wizard", true}, {"bear", "Bear", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := effectstest.SpendTable(t)
			camp := effectstest.PushCatalogPermanent(g, me.ID, "Base Camp", "Land", znrBaseCamp, false)
			effectstest.ActivateManaFor(t, g, me.ID, camp, 1, game.ManaAbilityParams{})
			if effectstest.LatestChoiceOfKind(g, game.PendingChoiceMana) != nil {
				effectstest.ResolveManaPickAnyColor(t, g, me.ID)
			}
			id := uuid.New()
			me.Hand.PushTop(game.Card{InstanceID: id, Name: "Test", TypeLine: "Creature — " + tc.subtype, ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
			err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
			if tc.ok && err != nil {
				t.Errorf("Base Camp's mana could not pay for a %s spell: %v", tc.name, err)
			}
			if !tc.ok {
				effectstest.RefusedForMana(t, err, "Base Camp's mana for a "+tc.name)
			}
		})
	}
}

func TestS82OldB097SeaGateColossusCostsOneLessPerPartyMember(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	cost := func() int {
		return effectstest.SelfPriced(t, g, me, znrColossus, "Artifact Creature — Golem Warrior", "{7}", game.ZoneHand, nil).ManaValue()
	}
	if got := cost(); got != 7 {
		t.Errorf("with no party it costs %d, want 7", got)
	}
	effectstest.B12Creature(g, me.ID, "Cleric", "Creature — Human Cleric", 1, 1)
	effectstest.B12Creature(g, me.ID, "Second Cleric", "Creature — Human Cleric", 1, 1)
	effectstest.B12Creature(g, me.ID, "Rogue", "Creature — Human Rogue", 1, 1)
	if got := cost(); got != 5 {
		t.Errorf("with a Cleric and a Rogue (a second Cleric adds nothing) it costs %d, want 5", got)
	}
	effectstest.B12Creature(g, me.ID, "Warrior", "Creature — Human Warrior", 1, 1)
	effectstest.B12Creature(g, me.ID, "Wizard", "Creature — Human Wizard", 1, 1)
	if got := cost(); got != 3 {
		t.Errorf("with a full party it costs %d, want 3", got)
	}
}

func znrGraveyardCleric(p *game.Player, name, cost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Human Cleric", ManaCost: cost,
		Power: 1, Toughness: 1, Owner: p.ID, Controller: p.ID})
	return id
}

func TestS82OldB097OrahReturnsALesserClericWhenAClericDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Orah, Skyclave Hierophant", "Legendary Creature — Kor Cleric", znrOrah, 3, 3)
	cleric := effectstest.B12Creature(g, me.ID, "Acolyte", "Creature — Human Cleric", 1, 1)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, cleric).ManaCost = "{2}{W}" })
	cheap := znrGraveyardCleric(me, "Cheap Cleric", "{1}{W}")
	equal := znrGraveyardCleric(me, "Equal Cleric", "{2}{W}")
	pricey := znrGraveyardCleric(me, "Pricey Cleric", "{3}{W}")
	effectstest.KillCreature(t, g, me.ID, cleric)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickCard(t, g, me.ID, cheap)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if !g.Battlefield.Contains(cheap) {
		t.Fatal("the lesser Cleric did not return")
	}
	if g.Battlefield.Contains(equal) || g.Battlefield.Contains(pricey) {
		t.Error("a Cleric of equal or greater mana value returned")
	}
}

func TestS82OldB097OrahDoesNotTriggerForANonCleric(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Orah, Skyclave Hierophant", "Legendary Creature — Kor Cleric", znrOrah, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, bear).ManaCost = "{4}" })
	cheap := znrGraveyardCleric(me, "Cheap Cleric", "{1}{W}")
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cheap) {
		t.Error("a Bear dying brought a Cleric back")
	}
}

func TestS82OldB097OrahHerselfDyingReturnsALesserCleric(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	orah := effectstest.B12Push(g, me.ID, "Orah, Skyclave Hierophant", "Legendary Creature — Kor Cleric", znrOrah, 3, 3)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, orah).ManaCost = "{2}{W}{B}" })
	cheap := znrGraveyardCleric(me, "Cheap Cleric", "{1}{W}")
	effectstest.KillCreature(t, g, me.ID, orah)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickCard(t, g, me.ID, cheap)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if !g.Battlefield.Contains(cheap) {
		t.Error("Orah dying did not return the lesser Cleric")
	}
}

func TestS82OldB097NahiriPlusOneMakesAKorWarriorAndMayAttachAnEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	n := effectstest.PushWalkerForTest(g, me.ID, "Nahiri, Heir of the Ancients", znrNahiri, 4)
	sword := effectstest.B12Permanent(g, me.ID, "Sword", "Artifact — Equipment")
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, n, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	warrior := effectstest.FindBattlefieldByName(g, "Kor Warrior")
	if warrior == uuid.Nil {
		t.Fatal("no Kor Warrior token")
	}
	if c, _ := effectstest.BattlefieldCard(g, warrior); !c.IsCreature() || c.CurrentPower() != 1 || c.CurrentToughness() != 1 || !c.HasSubtype("Warrior") {
		t.Errorf("Kor Warrior is %+v, want a 1/1 Warrior creature", c)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, sword)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, sword); c.AttachedTo.ID != warrior {
		t.Errorf("the Equipment is attached to %v, want the Warrior", c.AttachedTo.ID)
	}
	if got := effectstest.LoyaltyOf(g, n); got != 5 {
		t.Errorf("loyalty %d, want 5", got)
	}
}

func TestS82OldB097NahiriMinusTwoTakesAWarriorOrEquipmentFromTheTopSix(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	n := effectstest.PushWalkerForTest(g, me.ID, "Nahiri, Heir of the Ancients", znrNahiri, 4)
	effectstest.AdvanceToMain(t, g)
	var warrior, sword, bear uuid.UUID
	g.WithWriteLock(func() {
		warrior, sword, bear = uuid.New(), uuid.New(), uuid.New()
		me.Library.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
		me.Library.PushTop(game.Card{InstanceID: sword, Name: "Sword", TypeLine: "Artifact — Equipment", Owner: me.ID, Controller: me.ID})
		me.Library.PushTop(game.Card{InstanceID: warrior, Name: "Warrior", TypeLine: "Creature — Human Warrior", Owner: me.ID, Controller: me.ID})
	})
	libBefore := me.Library.Size()
	effectstest.B16Activate(t, g, me.ID, n, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no take prompt")
	}
	for _, id := range pick.ChooseCards {
		if id == bear {
			t.Error("a Bear is offered")
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, sword)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(sword) {
		t.Error("the Equipment did not reach your hand")
	}
	if me.Hand.Contains(warrior) {
		t.Error("two cards were taken")
	}
	if me.Library.Size() != libBefore-1 {
		t.Errorf("library %d, want %d (one taken, the rest on the bottom)", me.Library.Size(), libBefore-1)
	}
	if got := effectstest.LoyaltyOf(g, n); got != 2 {
		t.Errorf("loyalty %d, want 2", got)
	}
}

func TestS82OldB097NahiriMinusThreeDealsTwiceTheEquipmentYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	n := effectstest.PushWalkerForTest(g, me.ID, "Nahiri, Heir of the Ancients", znrNahiri, 4)
	effectstest.B12Permanent(g, me.ID, "Sword", "Artifact — Equipment")
	effectstest.B12Permanent(g, me.ID, "Axe", "Artifact — Equipment")
	effectstest.B12Permanent(g, opp.ID, "Their Sword", "Artifact — Equipment")
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Ogre", 9, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, n, 2, game.ActivateAbilityParams{Targets: effectstest.CardRefs(big)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("the 4-toughness creature survived 4 damage (two Equipment you control, twice)")
	}
	if got := effectstest.LoyaltyOf(g, n); got != 1 {
		t.Errorf("loyalty %d, want 1", got)
	}
}
