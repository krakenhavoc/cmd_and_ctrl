package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 Amonkhet batch 248 (#5535): enters and dies triggers, graveyard
// recursion and damage.

const (
	akh248Intervention = "02fe8ed4-3c5d-4875-b584-b33b819c47be"
	akh248Vizier       = "340e1078-5158-4199-ab43-976741780b2b"
	akh248Ambition     = "bf614953-bd62-4000-a960-e2c26e1f7bb0"
	akh248Solidarity   = "cbcc2577-7556-4174-83ff-9e00611412b3"
	akh248Scribe       = "019a3229-dd1c-4f70-a8fb-e33dbc54df39"
	akh248Excavation   = "5d8ff6a0-7037-4231-b6a9-ad97260bb267"
	akh248Wanderer     = "88154d64-221f-4742-bccb-b1b61a2e8bf0"
	akh248Neheb        = "092847a7-a238-4acf-9dcd-be6f3e6a6b93"
	akh248Hazoret      = "2c6934b6-67c4-457a-ae8f-59914fd8f256"
	akh248Hekma        = "c5f611c9-9e6e-4719-ad71-8686327a8564"

	akh248Bolt       = "4457ed35-7c10-48c8-9776-456485fdf070"
	akh248Archfiend  = "535e2af0-7b08-4026-b1bf-87626407dd42" // a card with cycling
	akh248CartoucheS = "b9176173-c4a1-45ae-8f34-bd6b31c141fe" // Cartouche of Solidarity
)

func TestAkh248CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		akh248Intervention: "Gideon's Intervention", akh248Vizier: "Pitiless Vizier", akh248Ambition: "Trial of Ambition",
		akh248Solidarity: "Trial of Solidarity", akh248Scribe: "Scribe of the Mindful", akh248Excavation: "Sacred Excavation",
		akh248Wanderer: "Dread Wanderer", akh248Neheb: "Neheb, the Worthy", akh248Hazoret: "Hazoret the Fervent",
		akh248Hekma: "Protection of the Hekma",
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

// akh248SetHand gives `p` exactly n cards in hand, through real draws and
// discards so the layer cache sees the hand change.
func akh248SetHand(t *testing.T, g *game.Game, p *game.Player, n int) {
	t.Helper()
	for p.Hand.Size() > n {
		g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(p.ID, 1) })
	}
	for p.Hand.Size() < n {
		effectstest.PushLibraryCardForTest(p, game.Card{Name: "Filler", TypeLine: "Land"})
		if err := g.DrawCard(p.ID); err != nil {
			t.Fatalf("draw: %v", err)
		}
	}
}

func TestAkh248GideonsInterventionBansTheNamedSpellForOpponentsAndPreventsItsDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Gideon's Intervention", "Enchantment", akh248Intervention, nil)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceCardName, me.ID)
	if prompt == nil {
		t.Fatalf("no card-name prompt as it entered: %+v", g.PendingChoices)
	}
	if err := g.ResolveCardNameChoice(prompt.ID, me.ID, "Lightning Bolt"); err != nil {
		t.Fatalf("ResolveCardNameChoice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// A source with the chosen name deals no damage to me or my permanents.
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", akh248Bolt, []game.TargetRef{effectstest.PlayerTarget(me.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", akh248Bolt, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life || !effectstest.BattlefieldHas(g, bear) {
		t.Errorf("life %d (want %d), Bear on battlefield %v: a named source's damage was not prevented", me.Life, life, effectstest.BattlefieldHas(g, bear))
	}
	// A different source's damage is not prevented.
	effectstest.CastCatalogSpell(t, g, "Burst Lightning", "Instant", effectstest.BurstLightningOracle, []game.TargetRef{effectstest.PlayerTarget(me.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-2 {
		t.Errorf("life %d after Burst Lightning, want %d", me.Life, life-2)
	}
	// On an opponent's turn the opponent can't cast the named spell.
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Lightning Bolt", "Instant", akh248Bolt, []game.TargetRef{effectstest.PlayerTarget(opp.ID)}); err == nil {
		t.Error("an opponent cast the named spell")
	}
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Burst Lightning", "Instant", effectstest.BurstLightningOracle, []game.TargetRef{effectstest.PlayerTarget(me.ID)}); err != nil {
		t.Errorf("an opponent could not cast a differently named spell: %v", err)
	}
}

func TestAkh248PitilessVizierGainsIndestructibleWhenYouCycleOrDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	vizier := effectstest.PushCatalogPermanent(g, me.ID, "Pitiless Vizier", "Creature — Minotaur Cleric", akh248Vizier, false)
	if effectstest.HasKeywordOf(g, vizier, "indestructible") {
		t.Fatal("the Vizier starts indestructible")
	}
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Drawn", TypeLine: "Land"})
	effectstest.CycleFromHand(t, g, "Archfiend of Ifnir", "Creature — Demon", akh248Archfiend, "{C}{C}")
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOf(g, vizier, "indestructible") {
		t.Error("the Vizier did not gain indestructible after a card was cycled")
	}
}

func TestAkh248PitilessVizierIgnoresAnOpponentsDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	vizier := effectstest.PushCatalogPermanent(g, me.ID, "Pitiless Vizier", "Creature — Minotaur Cleric", akh248Vizier, false)
	effectstest.S58HandCards(opp, 1)
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(opp.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasKeywordOf(g, vizier, "indestructible") {
		t.Error("an opponent's discard made the Vizier indestructible")
	}
}

func TestAkh248TrialOfAmbitionMakesTargetOpponentSacrificeAndReturnsForACartouche(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	keep := effectstest.B12Creature(g, opp.ID, "Keep", "Creature — Bear", 3, 3)
	lose := effectstest.B12Creature(g, opp.ID, "Lose", "Creature — Bear", 1, 1)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	trial := effectstest.CastCatalogSpell(t, g, "Trial of Ambition", "Enchantment", akh248Ambition, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, opp.ID, lose)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, lose) || !effectstest.BattlefieldHas(g, keep) || !effectstest.BattlefieldHas(g, mine) {
		t.Errorf("Lose %v, Keep %v, Mine %v: want the opponent's chosen creature gone and the rest kept",
			effectstest.BattlefieldHas(g, lose), effectstest.BattlefieldHas(g, keep), effectstest.BattlefieldHas(g, mine))
	}
	if _, ok := effectstest.BattlefieldCard(g, trial); !ok {
		t.Fatal("the Trial is not on the battlefield")
	}
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Cartouche of Solidarity", "Enchantment — Aura Cartouche", akh248CartoucheS, effectstest.CardTarget(host))
	effectstest.PassPriorityAroundTable(t, g)
	akh248OrderTriggers(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, trial); ok {
		t.Error("the Trial stayed after a Cartouche entered")
	}
	if !me.Hand.Contains(trial) {
		t.Error("the Trial did not return to its owner's hand")
	}
}

func akh248OrderTriggers(t *testing.T, g *game.Game) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerOrder {
			if err := g.ResolveTriggerOrder(c.ID, c.Chooser, append([]uuid.UUID(nil), c.TriggerOrderIDs...)); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			return
		}
	}
}

func TestAkh248TrialOfSolidarityPumpsYourCreaturesAndGivesVigilance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Trial of Solidarity", "Enchantment", akh248Solidarity, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 3 {
		t.Errorf("my Bear is %d/%d, want 4/3", p, tough)
	}
	if !effectstest.HasKeywordOf(g, bear, "vigilance") {
		t.Error("my Bear did not gain vigilance")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, theirs); p != 2 || tough != 2 || effectstest.HasKeywordOf(g, theirs, "vigilance") {
		t.Errorf("their Bear is %d/%d, vigilance %v: it must be untouched", p, tough, effectstest.HasKeywordOf(g, theirs, "vigilance"))
	}
}

func TestAkh248ScribeOfTheMindfulReturnsAnInstantOrSorceryAndNothingElse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	scribe := effectstest.PushCatalogPermanent(g, me.ID, "Scribe of the Mindful", "Creature — Human Cleric", akh248Scribe, false)
	spell := effectstest.PushCatalogGraveyardCard(me, "Old Spell", "Instant", "", 0, 0)
	bear := effectstest.PushCatalogGraveyardCard(me, "Old Bear", "Creature — Bear", "", 2, 2)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, scribe, 0, game.ActivateAbilityParams{
		Targets: effectstest.CardTarget(bear), Strict: true}); err == nil {
		t.Fatal("the Scribe targeted a creature card")
	}
	if err := g.ActivateCatalogAbility(me.ID, scribe, 0, game.ActivateAbilityParams{
		Targets: effectstest.CardTarget(spell), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(spell) {
		t.Error("the instant did not return to hand")
	}
	if effectstest.BattlefieldHas(g, scribe) {
		t.Error("the Scribe was not sacrificed")
	}
}

func TestAkh248SacredExcavationReturnsUpToTwoCardsWithCycling(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.PushCatalogGraveyardCard(me, "Archfiend of Ifnir", "Creature — Demon", akh248Archfiend, 5, 4)
	b := effectstest.PushCatalogGraveyardCard(me, "Archfiend of Ifnir", "Creature — Demon", akh248Archfiend, 5, 4)
	c := effectstest.PushCatalogGraveyardCard(me, "Archfiend of Ifnir", "Creature — Demon", akh248Archfiend, 5, 4)
	plain := effectstest.PushCatalogGraveyardCard(me, "Plain Bear", "Creature — Bear", "", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Sacred Excavation", "Sorcery", akh248Excavation, effectstest.CardTarget(plain)); err == nil {
		t.Error("Sacred Excavation targeted a card with no cycling")
	}
	effectstest.CastCatalogSpell(t, g, "Sacred Excavation", "Sorcery", akh248Excavation, effectstest.CardTargets(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) {
		t.Error("the two targeted cards did not return to hand")
	}
	if !me.Graveyard.Contains(c) || !me.Graveyard.Contains(plain) {
		t.Error("an untargeted card left the graveyard")
	}
}

func TestAkh248DreadWandererReturnsEnteredTappedAtSorcerySpeedWithOneOrFewerCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushCatalogGraveyardCard(me, "Dread Wanderer", "Creature — Zombie Jackal", akh248Wanderer, 2, 1)
	akh248SetHand(t, g, me, 2)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{B}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the Wanderer returned with two cards in hand")
	}
	akh248SetHand(t, g, me, 1)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate with one card in hand: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.BattlefieldHas(g, id) {
		t.Fatal("the Wanderer did not return to the battlefield")
	}
	if !effectstest.TappedNow(g, id) {
		t.Error("the Wanderer entered untapped")
	}
}

func TestAkh248DreadWandererOnlyAsASorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	id := effectstest.PushCatalogGraveyardCard(me, "Dread Wanderer", "Creature — Zombie Jackal", akh248Wanderer, 2, 1)
	me.Hand.Cards = nil
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{B}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the Wanderer returned outside a main phase")
	}
}

func TestAkh248NehebGrantsFirstStrikePumpsMinotaursAndMakesEveryoneDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	neheb := effectstest.PushCatalogPermanent(g, me.ID, "Neheb, the Worthy", "Legendary Creature — Minotaur Warrior", akh248Neheb, false)
	other := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Other Minotaur", TypeLine: "Creature — Minotaur", Power: 2, Toughness: 2})
	bear := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	theirs := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Their Minotaur", TypeLine: "Creature — Minotaur", Power: 2, Toughness: 2})
	akh248SetHand(t, g, me, 3)
	nehebBase, _ := effectstest.EffectivePTOf(t, g, neheb)
	if !effectstest.HasKeywordOf(g, other, "first strike") {
		t.Error("another Minotaur lacks first strike")
	}
	if effectstest.HasKeywordOf(g, bear, "first strike") || effectstest.HasKeywordOf(g, theirs, "first strike") {
		t.Error("a non-Minotaur or an opponent's Minotaur has first strike")
	}
	if p, _ := effectstest.EffectivePTOf(t, g, other); p != 2 {
		t.Errorf("a Minotaur has %d power with three cards in hand, want 2", p)
	}
	akh248SetHand(t, g, me, 1)
	if p, _ := effectstest.EffectivePTOf(t, g, other); p != 4 {
		t.Errorf("a Minotaur has %d power with one card in hand, want 4", p)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, neheb); p != nehebBase+2 {
		t.Errorf("Neheb has %d power with one card in hand, want %d (he is a Minotaur)", p, nehebBase+2)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, bear); p != 2 {
		t.Errorf("a Bear has %d power, want 2", p)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, theirs); p != 2 {
		t.Errorf("their Minotaur has %d power, want 2", p)
	}
}

func TestAkh248NehebMakesEachPlayerDiscardWhenItHitsAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	neheb := effectstest.PushCatalogPermanent(g, me.ID, "Neheb, the Worthy", "Legendary Creature — Minotaur Warrior", akh248Neheb, false)
	akh248SetHand(t, g, me, 2)
	akh248SetHand(t, g, opp, 2)
	effectstest.DealCombatDamage(g, me.ID, neheb, opp.ID)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range []*game.Player{me, opp} {
		if c := effectstest.DiscardChoiceFor(g, p.ID); c != nil {
			effectstest.AnswerDiscard(t, g, p.ID, p.Hand.Cards[0].InstanceID)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != 1 || opp.Hand.Size() != 1 {
		t.Errorf("hands %d and %d, want 1 each (each player discarded a card)", me.Hand.Size(), opp.Hand.Size())
	}
}

func TestAkh248HazoretCantAttackOrBlockUnlessYouHaveOneOrFewerCardsInHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	hazoret := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Hazoret the Fervent", OracleID: akh248Hazoret,
		TypeLine: "Legendary Creature — God", Power: 5, Toughness: 4})
	akh248SetHand(t, g, me, 2)
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	effectstest.WantAttackRefused(t, g, hazoret, them.ID, "more than one card in hand")
	akh248SetHand(t, g, me, 1)
	if err := effectstest.DeclareResult(g, hazoret, them.ID); err != nil {
		t.Errorf("attack with one card in hand: %v", err)
	}
}

func TestAkh248HazoretBlocksOnlyWithOneOrFewerCardsInHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	attacker, defender := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, attacker.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	hazoret := effectstest.B12Push(g, defender.ID, "Hazoret the Fervent", "Legendary Creature — God", akh248Hazoret, 5, 4)
	effectstest.S58HandCards(defender, 2)
	effectstest.B12Creature(g, defender.ID, "Elf", "Creature — Elf", 1, 1)
	effectstest.BrAttack(t, g, bear)
	effectstest.BrRefusal(t, g.Clone().DeclareBlocker(hazoret, bear), game.BlockReasonCantBlockAttacker)
	defender.Hand.Cards = nil
	if err := g.DeclareBlocker(hazoret, bear); err != nil {
		t.Errorf("block with an empty hand: %v", err)
	}
}

func TestAkh248HazoretDiscardsACardToDealTwoToEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	hazoret := effectstest.PushCatalogPermanent(g, me.ID, "Hazoret the Fervent", "Legendary Creature — God", akh248Hazoret, false)
	akh248SetHand(t, g, me, 3)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{R}"); err != nil {
		t.Fatal(err)
	}
	discard := me.Hand.Cards[0].InstanceID
	if err := g.ActivateCatalogAbility(me.ID, hazoret, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{discard}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(discard) {
		t.Error("the discarded card is not in the graveyard")
	}
	for _, p := range g.Seats {
		want := lives[p.ID]
		if p.ID != me.ID {
			want -= 2
		}
		if p.Life != want {
			t.Errorf("seat %s life %d, want %d", p.ID, p.Life, want)
		}
	}
}

func TestAkh248ProtectionOfTheHekmaPreventsOneDamageFromAnOpponentsSourceToYouOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Protection of the Hekma", "Enchantment", akh248Hekma, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// My own source is not an opponent's: full damage.
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", akh248Bolt, []game.TargetRef{effectstest.PlayerTarget(me.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-3 {
		t.Fatalf("life %d after my own Bolt, want %d", me.Life, life-3)
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 4, 4)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	life = me.Life
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", akh248Bolt, []game.TargetRef{effectstest.PlayerTarget(me.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-2 {
		t.Errorf("life %d after an opponent's Bolt, want %d (3 damage, 1 prevented)", me.Life, life-2)
	}
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", akh248Bolt, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageOn(t, g, bear); got != 3 {
		t.Errorf("Bear has %d damage, want 3: the prevention is for damage to you, not your permanents", got)
	}
}
