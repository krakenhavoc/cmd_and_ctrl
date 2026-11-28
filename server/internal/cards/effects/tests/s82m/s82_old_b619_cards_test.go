package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b619_cards_test.go — S82 OLD batch 619 (#5607): Eighth Edition pump
// and combat tricks, enters and dies triggers, graveyard recursion. Every test
// plays the real catalog card through the engine.

const (
	oldB619Behemoth = "4364bef8-f7ec-487c-9b64-b55c4afb864e"
	oldB619Foratog  = "77f257b9-2047-40a3-85c2-00ad17deeb08"
	oldB619Shambler = "7a2722c6-04a0-4790-be45-dafb7832bb41"
	oldB619Dancer   = "9ae49b45-a38f-4bd6-aa08-b5a5da6e75a1"
	oldB619Sentry   = "ac6f490b-6cf7-4afc-910e-4c8b51fcb2de"
	oldB619Sky      = "4634fd4e-5b38-4cfb-ab73-bf3b14b3bdee"
	oldB619Wrath    = "740cafc6-9521-403a-ab5f-11582bf55af0"
	oldB619DeathPit = "d445032b-b243-4203-9c88-a0ac357fb9b5"
	oldB619Force    = "f4170db0-3adf-4744-b0ec-e889c713bb93"
	oldB619Hammer   = "3089bdd7-87e6-4d0e-ad98-c32e6f861c30"
	oldB619Skull    = "67601399-089b-44fe-82db-c319729c75bf"
	oldB619Forest   = "Basic Land — Forest"
)

func TestOldB619CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB619Behemoth: "Llanowar Behemoth", oldB619Foratog: "Foratog", oldB619Shambler: "Primeval Shambler",
		oldB619Dancer: "Sword Dancer", oldB619Sentry: "Mogg Sentry", oldB619Sky: "Shifting Sky",
		oldB619Wrath: "Wrath of Marit Lage", oldB619DeathPit: "Death Pit Offering", oldB619Force: "Primeval Force",
		oldB619Hammer: "Hammer of Bogardan", oldB619Skull: "Skull of Orm",
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

// oldB619Pay floats mana and activates ability idx Strict, then settles.
func oldB619Pay(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, idx int, mana string, params game.ActivateAbilityParams) error {
	t.Helper()
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if mana != "" {
		effectstest.DcMana(t, g, me.ID, mana)
	}
	params.Strict = true
	err := g.ActivateCatalogAbility(me.ID, id, idx, params)
	if err == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	return err
}

func TestOldB619LlanowarBehemothTapsAnyUntappedCreatureToGrow(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	beast := effectstest.B12Push(g, me.ID, "Llanowar Behemoth", "Creature — Elemental", oldB619Behemoth, 4, 4)
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	if err := oldB619Pay(t, g, me, beast, 0, "", game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated without naming a creature to tap")
	}
	if err := oldB619Pay(t, g, me, beast, 0, "", game.ActivateAbilityParams{TapIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("activate tapping the Bear: %v", err)
	}
	if p, tough := effectstest.PtOf(t, g, beast); p != 5 || tough != 5 {
		t.Errorf("Behemoth is %d/%d, want 5/5", p, tough)
	}
	if err := oldB619Pay(t, g, me, beast, 0, "", game.ActivateAbilityParams{TapIDs: []uuid.UUID{bear}}); err == nil {
		t.Error("tapped an already tapped creature for the cost")
	}
	if err := oldB619Pay(t, g, me, beast, 0, "", game.ActivateAbilityParams{TapIDs: []uuid.UUID{beast}}); err != nil {
		t.Fatalf("activate tapping itself: %v", err)
	}
	if p, tough := effectstest.PtOf(t, g, beast); p != 6 || tough != 6 {
		t.Errorf("Behemoth is %d/%d, want 6/6", p, tough)
	}
}

func TestOldB619ForatogSacrificesAForestForPlusTwoPlusTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	atog := effectstest.B12Push(g, me.ID, "Foratog", "Creature — Atog", oldB619Foratog, 1, 2)
	forest := effectstest.PushLandFor(g, me.ID, "Forest", oldB619Forest)
	island := effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	if err := oldB619Pay(t, g, me, atog, 0, "{G}", game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{island}}); err == nil {
		t.Fatal("sacrificed an Island to pay for Foratog")
	}
	if err := oldB619Pay(t, g, me, atog, 0, "{G}", game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{forest}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(forest) {
		t.Error("the Forest was not sacrificed")
	}
	if p, tough := effectstest.PtOf(t, g, atog); p != 3 || tough != 4 {
		t.Errorf("Foratog is %d/%d, want 3/4", p, tough)
	}
}

func TestOldB619PrimevalShamblerPumpsForEachBlackMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	shambler := effectstest.B12Push(g, me.ID, "Primeval Shambler", "Creature — Horror Mercenary", oldB619Shambler, 3, 3)
	for i := 0; i < 2; i++ {
		if err := oldB619Pay(t, g, me, shambler, 0, "{B}", game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate %d: %v", i, err)
		}
	}
	if p, tough := effectstest.PtOf(t, g, shambler); p != 5 || tough != 5 {
		t.Errorf("Shambler is %d/%d, want 5/5", p, tough)
	}
	if err := oldB619Pay(t, g, me, shambler, 0, "{R}", game.ActivateAbilityParams{}); err == nil {
		t.Error("paid for Shambler with red mana")
	}
}

func TestOldB619SwordDancerShrinksOnlyAnAttackingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dancer := effectstest.B12Push(g, me.ID, "Sword Dancer", "Creature — Human Rebel", oldB619Dancer, 1, 2)
	attacker := effectstest.B12Creature(g, opp.ID, "Raider", "Creature — Human", 3, 3)
	idle := effectstest.B12Creature(g, opp.ID, "Idle", "Creature — Human", 3, 3)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == attacker {
				g.Battlefield.Cards[i].AttackingTarget = me.ID
			}
		}
	})
	if err := oldB619Pay(t, g, me, dancer, 0, "{W}{W}", game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(idle)}); err == nil {
		t.Fatal("targeted a creature that is not attacking")
	}
	if err := oldB619Pay(t, g, me, dancer, 0, "{W}{W}", game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(attacker)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if p, tough := effectstest.PtOf(t, g, attacker); p != 2 || tough != 3 {
		t.Errorf("the attacker is %d/%d, want 2/3", p, tough)
	}
}

func TestOldB619MoggSentryGrowsWheneverAnOpponentCastsASpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sentry := effectstest.B12Push(g, me.ID, "Mogg Sentry", "Creature — Goblin Warrior", oldB619Sentry, 1, 1)
	effectstest.Batch01OpponentCasts(t, g, opp, "Cantrip", "", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, sentry); p != 3 || tough != 3 {
		t.Fatalf("Sentry is %d/%d after one opposing spell, want 3/3", p, tough)
	}
	effectstest.Batch01OpponentCasts(t, g, opp, "Cantrip", "", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, sentry); p != 5 || tough != 5 {
		t.Errorf("Sentry is %d/%d after two, want 5/5", p, tough)
	}
}

func TestOldB619MoggSentryIgnoresYourOwnSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sentry := effectstest.B12Push(g, me.ID, "Mogg Sentry", "Creature — Goblin Warrior", oldB619Sentry, 1, 1)
	effectstest.CastCatalogSpell(t, g, "Giant Growth", "Instant", effectstest.RprB13GiantGrowthOracle, effectstest.B16TargetCard(sentry))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, sentry); p != 4 || tough != 4 {
		t.Errorf("Sentry is %d/%d, want 4/4 (Giant Growth only)", p, tough)
	}
}

func TestOldB619ShiftingSkyMakesEveryNonlandPermanentTheChosenColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Goblin", "Creature — Goblin", 1, 1)
	rock := effectstest.B12Permanent(g, opp.ID, "Mind Stone", "Artifact")
	land := effectstest.PushLandFor(g, me.ID, "Forest", oldB619Forest)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == theirs {
				g.Battlefield.Cards[i].Colors = []string{"R"}
			}
		}
	})
	effectstest.CastCatalogSpell(t, g, "Shifting Sky", "Enchantment", oldB619Sky, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "U")
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"colorless creature": mine, "red creature": theirs, "artifact": rock} {
		if cols := effectstest.EffectiveColors(t, g, id); len(cols) != 1 || cols[0] != "U" {
			t.Errorf("%s is %v, want exactly blue", name, cols)
		}
	}
	if cols := effectstest.EffectiveColors(t, g, land); len(cols) != 0 {
		t.Errorf("the land is %v, want colorless", cols)
	}
}

func TestOldB619WrathOfMaritLageTapsRedCreaturesAndKeepsThemTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	red := effectstest.B12Creature(g, opp.ID, "Goblin", "Creature — Goblin", 1, 1)
	green := effectstest.B12Creature(g, opp.ID, "Elf", "Creature — Elf", 1, 1)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			switch g.Battlefield.Cards[i].InstanceID {
			case red:
				g.Battlefield.Cards[i].Colors = []string{"R"}
			case green:
				g.Battlefield.Cards[i].Colors = []string{"G"}
			}
		}
	})
	effectstest.CastCatalogSpell(t, g, "Wrath of Marit Lage", "Enchantment", oldB619Wrath, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, red) {
		t.Fatal("the red creature was not tapped")
	}
	if effectstest.B16Tapped(t, g, green) {
		t.Error("a green creature was tapped")
	}
	effectstest.TapCard(g, green)
	effectstest.B39NextTurnOf(t, g, 1)
	if !effectstest.B16Tapped(t, g, red) {
		t.Error("the red creature untapped during its controller's untap step")
	}
	if effectstest.B16Tapped(t, g, green) {
		t.Error("the green creature stayed tapped")
	}
	_ = me
}

func TestOldB619DeathPitOfferingSacrificesYourCreaturesAndPumpsLaterOnes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Goblin", "Creature — Goblin", 1, 1)
	effectstest.CastCatalogSpell(t, g, "Death Pit Offering", "Enchantment", oldB619DeathPit, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || !me.Graveyard.Contains(mine) {
		t.Error("my creature was not sacrificed")
	}
	if !g.Battlefield.Contains(theirs) {
		t.Error("an opponent's creature was sacrificed")
	}
	if p, tough := effectstest.PtOf(t, g, theirs); p != 1 || tough != 1 {
		t.Errorf("the opponent's creature is %d/%d, want unchanged 1/1", p, tough)
	}
	later := effectstest.B12Creature(g, me.ID, "Ox", "Creature — Ox", 2, 2)
	if p, tough := effectstest.PtOf(t, g, later); p != 4 || tough != 4 {
		t.Errorf("a later creature is %d/%d, want 4/4", p, tough)
	}
}

func TestOldB619PrimevalForceIsSacrificedWithoutThreeForests(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushLandFor(g, me.ID, "Forest", oldB619Forest)
	effectstest.PushLandFor(g, me.ID, "Forest", oldB619Forest)
	id := effectstest.B12PlayFromHand(t, g, "Primeval Force", "Creature — Elemental", oldB619Force, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) {
		t.Error("Primeval Force stayed with only two Forests")
	}
	if n := effectstest.NamedOnBattlefield(g, "Forest", me.ID); n != 2 {
		t.Errorf("%d Forests left, want the two untouched", n)
	}
}

func TestOldB619PrimevalForceStaysForThreeSacrificedForests(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	var forests []uuid.UUID
	for i := 0; i < 4; i++ {
		forests = append(forests, effectstest.PushLandFor(g, me.ID, "Forest", oldB619Forest))
	}
	id := effectstest.B12PlayFromHand(t, g, "Primeval Force", "Creature — Elemental", oldB619Force, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.AnswerPermanentPicks(t, g, me.ID, map[uuid.UUID][]uuid.UUID{me.ID: forests[:3]})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Primeval Force was sacrificed after paying three Forests")
	}
	for _, f := range forests[:3] {
		if g.Battlefield.Contains(f) {
			t.Error("a chosen Forest was not sacrificed")
		}
	}
	if !g.Battlefield.Contains(forests[3]) {
		t.Error("the fourth Forest was sacrificed")
	}
}

func TestOldB619PrimevalForceCanBeSacrificedInstead(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 3; i++ {
		effectstest.PushLandFor(g, me.ID, "Forest", oldB619Forest)
	}
	id := effectstest.B12PlayFromHand(t, g, "Primeval Force", "Creature — Elemental", oldB619Force, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) {
		t.Error("Primeval Force stayed after choosing to sacrifice it")
	}
	if n := effectstest.NamedOnBattlefield(g, "Forest", me.ID); n != 3 {
		t.Errorf("%d Forests left, want all three", n)
	}
}

func TestOldB619HammerOfBogardanDealsThreeAndComesBackOnlyInYourUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := opp.Life
	hammer := effectstest.CastCatalogSpell(t, g, "Hammer of Bogardan", "Sorcery", oldB619Hammer, effectstest.B16TargetPlayer(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Fatalf("life %d → %d, want 3 damage", life, opp.Life)
	}
	if !me.Graveyard.Contains(hammer) {
		t.Fatal("the Hammer is not in the graveyard")
	}
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	if err := oldB619Pay(t, g, me, hammer, 0, "{C}{C}{R}{R}{R}", game.ActivateAbilityParams{}); err == nil {
		t.Fatal("returned the Hammer outside the upkeep")
	}
	effectstest.B39NextTurnOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if err := oldB619Pay(t, g, me, hammer, 0, "{C}{C}{R}{R}{R}", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("return during my upkeep: %v", err)
	}
	if !me.Hand.Contains(hammer) {
		t.Error("the Hammer did not return to hand")
	}
}

func TestOldB619SkullOfOrmReturnsAnEnchantmentCardFromYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	skull := effectstest.B12Push(g, me.ID, "Skull of Orm", "Artifact", oldB619Skull, 0, 0)
	ench := effectstest.Batch01GraveyardCard(me, "Pacifism", "Enchantment — Aura")
	bear := effectstest.Batch01GraveyardCard(me, "Bear", "Creature — Bear")
	theirs := effectstest.Batch01GraveyardCard(opp, "Opposing Aura", "Enchantment — Aura")
	if err := oldB619Pay(t, g, me, skull, 0, "{C}{C}{C}{C}{C}", game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear)}); err == nil {
		t.Fatal("targeted a creature card")
	}
	if err := oldB619Pay(t, g, me, skull, 0, "{C}{C}{C}{C}{C}", game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(theirs)}); err == nil {
		t.Fatal("targeted a card in an opponent's graveyard")
	}
	if err := oldB619Pay(t, g, me, skull, 0, "{C}{C}{C}{C}{C}", game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(ench)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Hand.Contains(ench) {
		t.Error("the enchantment card did not return to hand")
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, skull); !tapped {
		t.Error("Skull of Orm was not tapped")
	}
}
