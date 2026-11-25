package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b148_cards_test.go — Throne of Eldraine, S82 older sets batch 148
// (#5516): Inquisitive Puppet, Thorn Mammoth, Rosethorn Acolyte // Seasonal
// Ritual, Bramblefort Fink, Giant's Skewer, Steelclaw Lance, Archon of
// Absolution, Fires of Invention, Resolute Rider, Lucky Clover, Enchanted
// Carriage and Moonlit Scavengers.

const (
	eld148Puppet    = "015004d7-989b-4549-9ffa-b062126d04d2"
	eld148Mammoth   = "d7337b13-5459-434a-aa56-e1547b875594"
	eld148Acolyte   = "639a8804-8513-4ba4-8683-ea1e2e5a9530"
	eld148Fink      = "e56a9b7b-ce3f-4f5c-87d0-6f1d50d98ac3"
	eld148Skewer    = "aaa8377e-21de-4c63-a26a-c25d986a272d"
	eld148Lance     = "bfe06fb5-83b5-4212-86f8-947dfba15b26"
	eld148Archon    = "0e3ac581-3d78-4879-9afb-3db69c5c47a5"
	eld148Fires     = "27e27681-9221-4e64-8c78-e153a86bc6e0"
	eld148Rider     = "2f8b2bca-991c-4dee-829b-c1cc7deaadfb"
	eld148Clover    = "370d67f6-8d43-46ca-ae6c-00799bb0bd0c"
	eld148Carriage  = "4f4a3c9e-80c4-4fa1-940d-2c57065901a3"
	eld148Scavenger = "8403ea46-e6ac-458c-9c17-4c8a81f21659"
)

func TestS82OldB148EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		eld148Puppet: "Inquisitive Puppet", eld148Mammoth: "Thorn Mammoth", eld148Acolyte: "Rosethorn Acolyte",
		eld148Acolyte + "#1": "Seasonal Ritual", eld148Fink: "Bramblefort Fink", eld148Skewer: "Giant's Skewer",
		eld148Lance: "Steelclaw Lance", eld148Archon: "Archon of Absolution", eld148Fires: "Fires of Invention",
		eld148Rider: "Resolute Rider", eld148Clover: "Lucky Clover", eld148Carriage: "Enchanted Carriage",
		eld148Scavenger: "Moonlit Scavengers",
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

// --- Inquisitive Puppet ---

func TestS82OldB148InquisitivePuppetScriesThenExilesItselfForAHuman(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	puppet := effectstest.CastHoldCreature(t, g, "Inquisitive Puppet", "Artifact Creature — Construct", eld148Puppet, 0, 2)
	if c := effectstest.ScryChoiceFor(g, me.ID); c == nil {
		t.Fatal("no scry prompt on entering")
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	if err := g.ActivateCatalogAbility(me.ID, puppet, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.AppInZone(g, puppet, game.ZoneExile) {
		t.Error("the Puppet was not exiled as a cost")
	}
	if got := effectstest.TmtB1Count(g, me.ID, "Human"); got != 1 {
		t.Errorf("%d Humans, want 1", got)
	}
}

// --- Thorn Mammoth ---

func TestS82OldB148ThornMammothFightsWhenItOrAnotherCreatureEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	victim := effectstest.B12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	mammoth := effectstest.CastHoldCreature(t, g, "Thorn Mammoth", "Creature — Elephant", eld148Mammoth, 6, 6)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.AppInZone(g, victim, game.ZoneBattlefield) {
		t.Error("the Mammoth did not fight its own entering")
	}
	second := effectstest.B12Push(g, opp.ID, "Second Bear", "Creature — Bear", "", 3, 3)
	effectstest.CastHoldCreature(t, g, "Another Creature", "Creature — Bear", "", 1, 1)
	effectstest.PickCard(t, g, me.ID, second)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.AppInZone(g, second, game.ZoneBattlefield) {
		t.Error("the Mammoth did not fight when another creature entered")
	}
	if !effectstest.AppInZone(g, mammoth, game.ZoneBattlefield) {
		t.Error("the Mammoth died")
	}
}

func TestS82OldB148ThornMammothDoesNotTriggerOnAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Thorn Mammoth", "Creature — Elephant", eld148Mammoth, 6, 6)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Their Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LatestPickTarget(g, me.ID); c != nil {
		t.Error("the Mammoth triggered on an opponent's creature")
	}
}

// --- Rosethorn Acolyte // Seasonal Ritual ---

func TestS82OldB148RosethornAcolyteTapsForAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	acolyte := effectstest.B12Push(g, me.ID, "Rosethorn Acolyte", "Creature — Elf Druid", eld148Acolyte, 2, 3)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, acolyte).SummonedThisTurn = false })
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.TapForMana(t, g, me.ID, acolyte)
	eld148AnswerManaPick(t, g, me)
}

func TestS82OldB148SeasonalRitualAddsOneManaOfAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.CastCatalogSpell(t, g, "Seasonal Ritual", "Sorcery — Adventure", eld148Acolyte+"#1", nil)
	effectstest.PassPriorityAroundTable(t, g)
	eld148AnswerManaPick(t, g, me)
}

// eld148AnswerManaPick answers the open "any color" pick with green and checks
// that exactly one green mana is then in the pool.
func eld148AnswerManaPick(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	pick := effectstest.PendingOfKind(g, game.PendingChoiceMana)
	if pick == nil {
		t.Fatal("no mana color pick is open")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "G"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if n, green := effectstest.PoolSize(g, me.ID), effectstest.PoolCount(me, "G"); n != 1 || green != 1 {
		t.Errorf("pool has %d mana, %d green, want one green", n, green)
	}
}

// --- Bramblefort Fink ---

func TestS82OldB148BramblefortFinkNeedsAnOkoToGrowToTenTen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	fink := effectstest.B12Push(g, me.ID, "Bramblefort Fink", "Creature — Ouphe", eld148Fink, 2, 2)
	effectstest.FillPool(me, 8)
	if err := g.ActivateCatalogAbility(me.ID, fink, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated without an Oko planeswalker")
	}
	effectstest.B12Push(g, me.ID, "Oko, Thief of Crowns", "Legendary Planeswalker — Oko", "", 0, 0)
	if err := g.ActivateCatalogAbility(me.ID, fink, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate with an Oko: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, fink); p != 10 || tough != 10 {
		t.Errorf("the Fink is %d/%d, want 10/10", p, tough)
	}
}

func TestS82OldB148BramblefortFinkIgnoresANonOkoPlaneswalkerAndAnOpponentsOko(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	fink := effectstest.B12Push(g, me.ID, "Bramblefort Fink", "Creature — Ouphe", eld148Fink, 2, 2)
	effectstest.B12Push(g, me.ID, "Jace", "Legendary Planeswalker — Jace", "", 0, 0)
	effectstest.B12Push(g, opp.ID, "Oko, Thief of Crowns", "Legendary Planeswalker — Oko", "", 0, 0)
	effectstest.FillPool(me, 8)
	if err := g.ActivateCatalogAbility(me.ID, fink, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("activated without an Oko of your own")
	}
}

// --- Giant's Skewer ---

func TestS82OldB148GiantsSkewerMakesAFoodWhenEquippedCreatureDamagesACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	skewer := effectstest.SeedEquipment(g, me.ID, "Giant's Skewer", eld148Skewer)
	bear := effectstest.B12Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	other := effectstest.B12Push(g, me.ID, "Other Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.B12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 20)
	effectstest.FillPool(me, 3)
	effectstest.EquipTo(t, g, me.ID, skewer, bear)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 3 {
		t.Errorf("the equipped Bear is %d/%d, want 4/3", p, tough)
	}
	effectstest.EmitDamageEventForTest(g, other, theirs, 2, true)
	effectstest.EmitDamageEventForTest(g, bear, theirs, 2, false)
	effectstest.EmitDamageEventForTest(g, bear, opp.ID, 2, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB1Count(g, me.ID, "Food"); got != 0 {
		t.Fatalf("%d Food from damage that does not count, want 0", got)
	}
	effectstest.EmitDamageEventForTest(g, bear, theirs, 4, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB1Count(g, me.ID, "Food"); got != 1 {
		t.Errorf("%d Food, want 1", got)
	}
}

// --- Steelclaw Lance ---

func TestS82OldB148SteelclawLanceEquipsKnightsForOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	lance := effectstest.SeedEquipment(g, me.ID, "Steelclaw Lance", eld148Lance)
	knight := effectstest.B12Push(g, me.ID, "A Knight", "Creature — Human Knight", "", 2, 2)
	bear := effectstest.B12Push(g, me.ID, "A Bear", "Creature — Bear", "", 2, 2)
	effectstest.FillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, lance, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Fatal("Equip Knight targeted a Bear")
	}
	if err := g.ActivateCatalogAbility(me.ID, lance, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(knight)}); err != nil {
		t.Fatalf("Equip Knight {1}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, knight); p != 4 || tough != 4 {
		t.Errorf("the equipped Knight is %d/%d, want 4/4", p, tough)
	}
}

func TestS82OldB148SteelclawLanceEquipsAnythingForThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	lance := effectstest.SeedEquipment(g, me.ID, "Steelclaw Lance", eld148Lance)
	bear := effectstest.B12Push(g, me.ID, "A Bear", "Creature — Bear", "", 2, 2)
	effectstest.FillPool(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, lance, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("Equip {3}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, bear); p != 4 {
		t.Errorf("the equipped Bear's power %d, want 4", p)
	}
}

// --- Archon of Absolution ---

func TestS82OldB148ArchonOfAbsolutionTaxesEachAttackerOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	a := effectstest.PushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	archon := effectstest.PushCatalogPermanent(g, opp.ID, "Archon of Absolution", "Creature — Archon", eld148Archon, false)
	effectstest.AssertKeywords(t, g, archon, "flying", "protection from white")
	p := g.PriceAttackDeclaration([]game.AttackDeclaration{{Attacker: a, Target: opp.ID}, {Attacker: b, Target: opp.ID}})
	if p.Cost != "{1}{1}" {
		t.Errorf("two attackers cost %q, want {1}{1}", p.Cost)
	}
	free := g.PriceAttackDeclaration([]game.AttackDeclaration{{Attacker: a, Target: me.ID}})
	if free.Cost != "" {
		t.Errorf("attacking the Archon's opponent costs %q, want nothing", free.Cost)
	}
}

// --- Fires of Invention ---

func TestS82OldB148FiresOfInventionOnlyYourTurnTwoSpellsAndFreeSpellsUpToYourLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Fires of Invention", "Enchantment", eld148Fires, 0, 0)
	for i := 0; i < 3; i++ {
		effectstest.B12Push(g, me.ID, "Mountain", "Basic Land — Mountain", "", 0, 0)
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	cheap := effectstest.HandCardFull(me, "Cheap Spell", "Sorcery", "{3}", "", nil)
	dear := effectstest.HandCardFull(me, "Dear Spell", "Sorcery", "{4}", "", nil)
	third := effectstest.HandCardFull(me, "Third Spell", "Sorcery", "{1}", "", nil)
	if err := g.CastSpell(me.ID, dear, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast a four-mana spell for free with three lands")
	}
	if err := g.CastSpell(me.ID, cheap, game.CastSpellParams{Strict: true, AlternativeCost: effects.OldB148GrantedAltCostFreeUpToLandCount}); err != nil {
		t.Fatalf("cast a three-mana spell for free with three lands: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	second := effectstest.HandCardFull(me, "Second Spell", "Sorcery", "{1}", "", nil)
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{Strict: true, AlternativeCost: effects.OldB148GrantedAltCostFreeUpToLandCount}); err != nil {
		t.Fatalf("second spell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, third, game.CastSpellParams{Strict: true, AlternativeCost: effects.OldB148GrantedAltCostFreeUpToLandCount}); err == nil {
		t.Error("cast a third spell in one turn")
	}
	_ = opp
}

func TestS82OldB148FiresOfInventionStopsYouCastingOnAnOpponentsTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Fires of Invention", "Enchantment", eld148Fires, 0, 0)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	card := effectstest.HandCardFull(me, "Instant", "Instant", "{0}", "", nil)
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("cast an instant on an opponent's turn")
	}
}

// --- Resolute Rider ---

func TestS82OldB148ResoluteRiderGainsLifelinkAndIndestructibleWithHybridMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	rider := effectstest.B12Push(g, me.ID, "Resolute Rider", "Creature — Human Knight", eld148Rider, 4, 2)
	effectstest.FillPoolColored(me, "W", 2)
	if err := g.ActivateCatalogAbility(me.ID, rider, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("lifelink: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertKeywords(t, g, rider, "lifelink")
	effectstest.FillPoolColored(me, "B", 3)
	if err := g.ActivateCatalogAbility(me.ID, rider, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("indestructible: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertKeywords(t, g, rider, "indestructible")
}

// --- Lucky Clover ---

func TestS82OldB148LuckyCloverCopiesAnAdventureSpellButNotTheCreatureHalf(t *testing.T) {
	const reaper = "d40e5a76-6c9a-47eb-b412-d3056cf36b3f"
	for _, adventure := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		effectstest.AdvanceToMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.B12Push(g, me.ID, "Lucky Clover", "Artifact", eld148Clover, 0, 0)
		effectstest.RevealHand(opp, effectstest.RhForest(), effectstest.RhBolt(), effectstest.PbDivination(),
			game.Card{Name: "A", TypeLine: "Sorcery"}, game.Card{Name: "B", TypeLine: "Sorcery"})
		id := uuid.New()
		card := game.Card{InstanceID: id, Owner: me.ID, Controller: me.ID, ManaCost: "{B}", Colors: []string{"B"}}
		effectstest.FloatForTest(g, me, "BB")
		var targets []game.TargetRef
		if adventure {
			card.OracleID, card.Name, card.TypeLine, card.Layout = reaper+"#1", "Harvest Fear", "Sorcery — Adventure", game.LayoutAdventure
			targets = []game.TargetRef{effectstest.PbPlayer(opp.ID)}
		} else {
			card.OracleID, card.Name, card.TypeLine, card.Power, card.Toughness = reaper, "Reaper of Night", "Creature — Specter", 4, 5
		}
		me.Hand.PushTop(card)
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: targets}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		// Let the Clover's trigger resolve (it is above the spell) and stop
		// with the copy, if any, still on the stack.
		for i := 0; i < 8 && eld148TriggerOnStack(g); i++ {
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority: %v", err)
			}
		}
		for i := 0; i < 3 && effectstest.CopyPromptFor(g, me.ID) != nil; i++ {
			c := effectstest.CopyPromptFor(g, me.ID)
			if err := g.ResolvePickTarget(c.ID, me.ID, effectstest.PbPlayer(opp.ID)); err != nil {
				t.Fatalf("copy target: %v", err)
			}
		}
		want := 0
		if adventure {
			want = 1
		}
		if got := effectstest.StackCopies(g); got != want {
			t.Errorf("adventure=%v: %d copies on the stack, want %d", adventure, got, want)
		}
	}
}

// eld148TriggerOnStack reports whether a triggered ability is on the stack.
func eld148TriggerOnStack(g *game.Game) bool {
	for _, it := range g.StackMeta {
		if it != nil && it.Kind == game.StackItemTriggered {
			return true
		}
	}
	return false
}

// --- Enchanted Carriage ---

func TestS82OldB148EnchantedCarriageMakesTwoMiceAndCrews(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	carriage := effectstest.CastHoldCreature(t, g, "Enchanted Carriage", "Artifact — Vehicle", eld148Carriage, 4, 4)
	if got := effectstest.TmtB1Count(g, me.ID, "Mouse"); got != 2 {
		t.Errorf("%d Mice, want 2", got)
	}
	if effectstest.LayeredCard(t, g, carriage).IsCreature() {
		t.Error("the Carriage is a creature before it is crewed")
	}
}

// --- Moonlit Scavengers ---

func TestS82OldB148MoonlitScavengersBouncesOnlyWithAnArtifactOrEnchantment(t *testing.T) {
	for _, with := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := effectstest.SeatsForCombat(g)
		effectstest.AdvanceToMain(t, g)
		if with {
			effectstest.B12Push(g, me.ID, "Trinket", "Artifact", "", 0, 0)
		}
		victim := effectstest.B12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
		effectstest.CastHoldCreature(t, g, "Moonlit Scavengers", "Creature — Merfolk Rogue", eld148Scavenger, 4, 5)
		if with {
			effectstest.PickCard(t, g, me.ID, victim)
			effectstest.PassPriorityAroundTable(t, g)
		}
		bounced := opp.Hand.Contains(victim)
		if bounced != with {
			t.Errorf("artifact=%v: bounced = %v, want %v", with, bounced, with)
		}
	}
}
