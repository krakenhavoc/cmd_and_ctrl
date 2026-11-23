package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b223_test.go — Commander Legends: Battle for Baldur's Gate
// creatures, GAP batch 223 (#5948). Every test plays the real catalog entry
// through the engine.

const (
	gap223Serpent    = "d60e14b4-f750-40cc-85b5-17c87ae8a584"
	gap223Tomb       = "4f429b99-6aca-409f-9d2e-5743d81f1f22"
	gap223Invoker    = "36652a44-f5bd-4bc0-abfd-fa3688853dfd"
	gap223Vhal       = "7253e7d1-b1df-4637-bf64-850be7fa5188"
	gap223Wharf      = "ac88975c-fb2c-441a-bdd4-e7de6ba865e7"
	gap223Eladrin    = "48ae84e7-125d-4018-afc1-cd78448d1a13"
	gap223Thay       = "cda1e3b5-ea80-4e83-b305-88993d307170"
	gap223Dragon     = "fff28b30-9903-4a0a-a723-31216abd457d"
	gap223Oji        = "7c4c556c-b5e3-493a-a7b0-38c55b7809ac"
	gap223Council    = "3dfb006f-6620-4602-ac3c-d40a5a15a981"
	gap223Archpriest = "a8730922-3f47-4c07-bdd5-4a0f557a3d3d"
)

func TestGap223EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		gap223Serpent: "Sword Coast Serpent", gap223Serpent + "#1": "Capsizing Wave", gap223Tomb: "Tomb of Horrors Adventurer",
		gap223Invoker: "Tymora's Invoker", gap223Vhal: "Vhal, Candlekeep Researcher", gap223Wharf: "Wharf Infiltrator",
		gap223Eladrin: "Winter Eladrin", gap223Thay: "Wizards of Thay", gap223Dragon: "Young Blue Dragon",
		gap223Dragon + "#1": "Sand Augury", gap223Oji: "Oji, the Exquisite Blade", gap223Council: "The Council of Four",
		gap223Archpriest: "Archpriest of Iona",
	}
	for oracle, name := range want {
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

func TestGap223SwordCoastSerpentIsUnblockableOnlyAfterANoncreatureSpell(t *testing.T) {
	for _, cast := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		serpent := effectstest.B12Push(g, me.ID, "Sword Coast Serpent", "Creature — Serpent Dragon", gap223Serpent, 6, 6)
		blocker := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
		effectstest.AdvanceToMain(t, g)
		if cast {
			effectstest.CastCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
			effectstest.PassPriorityAroundTable(t, g)
		}
		effectstest.BrAttack(t, g, serpent)
		err := g.DeclareBlocker(blocker, serpent)
		if cast && err == nil {
			t.Error("the Serpent was blocked after a noncreature spell was cast")
		}
		if !cast && err != nil {
			t.Errorf("the Serpent could not be blocked with no spell cast: %v", err)
		}
	}
}

func TestGap223CreatureSpellDoesNotMakeTheSerpentUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	serpent := effectstest.B12Push(g, me.ID, "Sword Coast Serpent", "Creature — Serpent Dragon", gap223Serpent, 6, 6)
	blocker := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Test Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.BrAttack(t, g, serpent)
	if err := g.DeclareBlocker(blocker, serpent); err != nil {
		t.Errorf("a creature spell made the Serpent unblockable: %v", err)
	}
}

func TestGap223CapsizingWaveBouncesACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Capsizing Wave", TypeLine: "Instant — Adventure", OracleID: gap223Serpent + "#1",
		ManaCost: "{1}{U}", Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "U", "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Error("the Bear did not return to its owner's hand")
	}
}

func TestGap223TombOfHorrorsAdventurerTakesTheInitiative(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastHoldCreature(t, g, "Tomb of Horrors Adventurer", "Creature — Elf Monk", gap223Tomb, 4, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.YouHaveTheInitiative(g, me.ID) {
		t.Error("the Adventurer did not take the initiative")
	}
}

func TestGap223TombOfHorrorsAdventurerCopiesOnlyTheSecondSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.Msc18Library(me, 30)
	effectstest.B12Push(g, me.ID, "Tomb of Horrors Adventurer", "Creature — Elf Monk", gap223Tomb, 4, 4)
	hand := me.Hand.Size()
	effectstest.Msc18DivinationCast(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Fatalf("drew %d on the first spell, want 2: no copy", got)
	}
	hand = me.Hand.Size()
	effectstest.Msc18DivinationCast(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 4 {
		t.Errorf("drew %d on the second spell, want 4: the Divination and one copy", got)
	}
}

func TestGap223TombOfHorrorsAdventurerCopiesTwiceAfterCompletingADungeon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.Msc18Library(me, 30)
	me.CompletedDungeons = append(me.CompletedDungeons, "lost-mine-of-phandelver")
	effectstest.B12Push(g, me.ID, "Tomb of Horrors Adventurer", "Creature — Elf Monk", gap223Tomb, 4, 4)
	effectstest.Msc18DivinationCast(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.Msc18DivinationCast(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 6 {
		t.Errorf("drew %d on the second spell, want 6: the Divination and two copies", got)
	}
}

func TestGap223TymorasInvokerDrawsTwoForEightMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "A", "B", "C")
	inv := effectstest.PushCatalogPermanent(g, me.ID, "Tymora's Invoker", "Creature — Orc Rogue", gap223Invoker, true)
	effectstest.ApaMana(me, "C", "C", "C", "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, inv, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated for seven mana")
	}
	effectstest.ApaMana(me, "C")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, inv, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
}

func TestGap223VhalAddsItsToughnessInColorlessAndCantCastFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	vhal := effectstest.PushCatalogPermanent(g, me.ID, "Vhal, Candlekeep Researcher", "Legendary Creature — Human Wizard", gap223Vhal, false)
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, vhal)
		c.Power, c.Toughness = 2, 3
	})
	effectstest.TapForMana(t, g, me.ID, vhal)
	if pool := effectstest.PoolColors(me); len(pool) != 3 {
		t.Fatalf("pool %v, want three {C} (Vhal's toughness)", pool)
	}
	// The mana cannot pay for a spell from the hand.
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Test Rock", TypeLine: "Artifact", ManaCost: "{3}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a spell was cast from the hand with Vhal's mana")
	}
}

func TestGap223WharfInfiltratorIsSkulkAndLootsOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B", "C")
	wharf := effectstest.B12Push(g, me.ID, "Wharf Infiltrator", "Creature — Human Horror", gap223Wharf, 1, 1)
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, wharf), "skulk") {
		t.Error("no skulk")
	}
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Ogre", 3, 3)
	effectstest.BrAttack(t, g, wharf)
	if err := g.DeclareBlocker(big, wharf); err == nil {
		t.Error("a creature with greater power blocked the Infiltrator")
	}
}

func TestGap223WharfInfiltratorMakesAnEldraziHorrorWhenYouPayAfterDiscardingACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Wharf Infiltrator", "Creature — Human Horror", gap223Wharf, 1, 1)
	bear := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Discarded Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "C", "C")
	effectstest.LengDiscard(t, g, me, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Eldrazi Horror"); n != 1 {
		t.Errorf("%d Eldrazi Horror tokens, want 1", n)
	}
}

func TestGap223WinterEladrinBouncesAnotherCreatureOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	eladrin := effectstest.CastHoldCreature(t, g, "Winter Eladrin", "Creature — Faerie Elf Wizard", gap223Eladrin, 2, 2)
	p := effectstest.LatestPickTarget(g, g.Seats[g.Turn.ActiveSeat].ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(p.ID, p.Chooser, game.TargetRef{Kind: game.TargetCard, ID: eladrin}); err == nil {
		t.Error("the Eladrin targeted itself")
	}
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !opp.Hand.Contains(theirs) {
		t.Error("their creature was not returned to its owner's hand")
	}
}

func TestGap223WizardsOfThayDiscountsInstantsAndSorceriesAndFlashesSorceries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Wizards of Thay", "Creature — Human Wizard", gap223Thay, false)
	bolt := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: effectstest.LightningBoltOracle,
		Owner: me.ID, Controller: me.ID})
	// {R} less {1}: a generic reduction cannot take the coloured symbol, so the
	// Bolt still costs {R}; a {1}{R} instant costs {R}.
	two := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: two, Name: "Two Cost Instant", TypeLine: "Instant", ManaCost: "{1}{R}", Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "R")
	if err := g.CastSpell(me.ID, two, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("the {1}{R} instant was not discounted to {R}: %v", err)
	}
	// A creature spell is not discounted.
	bear := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Two Cost Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "G")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a creature spell was discounted")
	}
	// Sorcery flash: cast a sorcery on the opponent's turn.
	g.Turn.ActiveSeat = 1
	sorc := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: sorc, Name: "Test Sorcery", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, sorc, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("a sorcery could not be cast at instant speed: %v", err)
	}
}

func TestGap223WizardsOfThayHasMyriad(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Push(g, me.ID, "Wizards of Thay", "Creature — Human Wizard", gap223Thay, 3, 3)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	copies := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Wizards of Thay" && c.InstanceID != attacker && effects.IsToken(c) {
			copies++
		}
	}
	if copies != 2 {
		t.Errorf("%d myriad copies, want 2", copies)
	}
}

func TestGap223SandAuguryScriesThenDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "A", "B", "C")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Sand Augury", TypeLine: "Sorcery — Adventure", OracleID: gap223Dragon + "#1",
		ManaCost: "{1}{U}", Owner: me.ID, Controller: me.ID})
	hand := me.Hand.Size() - 1
	effectstest.ApaMana(me, "U", "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ResolveAnyScryFor(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

func TestGap223OjiGainsLifeThenBlinksOnTheSecondSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "A", "B", "C")
	life := me.Life
	oji := effectstest.CastHoldCreature(t, g, "Oji, the Exquisite Blade", "Legendary Creature — Human Monk", gap223Oji, 2, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2", life, me.Life)
	}
	_ = oji
}

func TestGap223CouncilOfFourDrawsOnAnyPlayersSecondDrawOfTheirTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.Msc18Library(me, 20)
	effectstest.B12Push(g, me.ID, "The Council of Four", "Legendary Creature — Human Noble", gap223Council, 0, 8)
	hand := me.Hand.Size()
	effectstest.Msc18DivinationCast(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	// Divination draws two (the second draw triggers one more), and the
	// Divination was the active player's first spell: no Knight yet.
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("hand grew by %d, want 3 (two from Divination, one from the Council)", got)
	}
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Knight"); n != 0 {
		t.Errorf("%d Knights after the first spell, want 0", n)
	}
	effectstest.Msc18DivinationCast(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Knight"); n != 1 {
		t.Errorf("%d Knights after the second spell, want 1", n)
	}
}

func TestGap223ArchpriestOfIonaPowerIsYourPartySize(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	pri := effectstest.B12Push(g, me.ID, "Archpriest of Iona", "Creature — Human Cleric", gap223Archpriest, 0, 2)
	if p, tough := effectstest.PtOf(t, g, pri); p != 1 || tough != 2 {
		t.Fatalf("%d/%d alone, want 1/2 (itself is the Cleric)", p, tough)
	}
	effectstest.B12Creature(g, me.ID, "Rogue", "Creature — Human Rogue", 1, 1)
	effectstest.B12Creature(g, me.ID, "Warrior", "Creature — Human Warrior", 1, 1)
	if p, _ := effectstest.PtOf(t, g, pri); p != 3 {
		t.Errorf("power %d with Cleric, Rogue and Warrior, want 3", p)
	}
	effectstest.B12Creature(g, me.ID, "Second Rogue", "Creature — Human Rogue", 1, 1)
	if p, _ := effectstest.PtOf(t, g, pri); p != 3 {
		t.Errorf("power %d with a second Rogue, want still 3", p)
	}
}

func TestGap223ArchpriestOfIonaPumpsATargetWithAFullParty(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Archpriest of Iona", "Creature — Human Cleric", gap223Archpriest, 0, 2)
	effectstest.B12Creature(g, me.ID, "Rogue", "Creature — Human Rogue", 1, 1)
	effectstest.B12Creature(g, me.ID, "Warrior", "Creature — Human Warrior", 1, 1)
	wiz := effectstest.B12Creature(g, me.ID, "Wizard", "Creature — Human Wizard", 1, 1)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.AnswerPickTarget(t, g, wiz)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, wiz); p != 2 || tough != 2 {
		t.Errorf("%d/%d, want 2/2", p, tough)
	}
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, wiz), "flying") {
		t.Error("no flying")
	}
}

func TestGap223ArchpriestOfIonaDoesNothingWithoutAFullParty(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Archpriest of Iona", "Creature — Human Cleric", gap223Archpriest, 0, 2)
	rogue := effectstest.B12Creature(g, me.ID, "Rogue", "Creature — Human Rogue", 1, 1)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Error("the trigger asked for a target without a full party")
	}
	if pw, _ := effectstest.PtOf(t, g, rogue); pw != 1 {
		t.Errorf("Rogue power %d, want 1", pw)
	}
}

func TestGap223OjiBlinksACreatureOnTheSecondSpellOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.Msc18Library(me, 20)
	effectstest.B12Push(g, me.ID, "Oji, the Exquisite Blade", "Legendary Creature — Human Monk", gap223Oji, 2, 3)
	bear := effectstest.B12Creature(g, me.ID, "Marked Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, bear).Counters = map[string]int{game.CounterPlusOne: 1}
	})
	effectstest.Msc18DivinationCast(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("the first spell triggered Oji")
	}
	if c, ok := g.LookupCardForEffect(bear); !ok || c.Counters[game.CounterPlusOne] != 1 {
		t.Fatal("the Bear changed after the first spell")
	}
	effectstest.Msc18DivinationCast(t, g)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Marked Bear" {
			if c.Counters[game.CounterPlusOne] != 0 {
				t.Errorf("the Bear kept its counter: it was not blinked")
			}
			return
		}
	}
	t.Error("the Bear did not return to the battlefield")
}
