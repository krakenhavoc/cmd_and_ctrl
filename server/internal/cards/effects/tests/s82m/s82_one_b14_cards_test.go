package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 ONE batch 14 (#5479): the Equipment-matters cards of Phyrexia: All
// Will Be One.

const (
	s82B14WarWhipOracle      = "d3142c6b-86e9-42b7-916c-ce7c34e7bd53"
	s82B14GlimmerLensOracle  = "bb7cebb7-a979-4879-b834-deab13f2a22c"
	s82B14WallbreakerOracle  = "f7f78825-10d6-441a-ae26-51b77cc271cd"
	s82B14AspirantOracle     = "644c785d-3061-44bc-9f78-bc939c565bf4"
	s82B14LightbringerOracle = "1f874f7f-b959-4de5-a5db-e65a90ac2f9d"
	s82B14SalvoOracle        = "a35b7fd3-fdef-4414-b720-978f7ce70bff"
	s82B14ReunitedOracle     = "e68fa0e6-9f55-498b-a13e-197cb30ca4aa"
	s82B14SleeperOracle      = "d4df1710-7402-45a7-966f-b51ddc087fbe"
	s82B14RebukeOracle       = "af05cc39-725b-4535-919a-ddbcb89c2210"
	s82B14SwordOracle        = "c6331ce3-21f7-4784-80fe-df1f541a4c46"
	s82B14GambitOracle       = "cf113674-f809-4b97-9e8d-e22385446ddb"
)

func TestS82OneB14EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		s82B14WarWhipOracle:      "Bladehold War-Whip",
		s82B14GlimmerLensOracle:  "Glimmer Lens",
		s82B14WallbreakerOracle:  "Hexplate Wallbreaker",
		s82B14AspirantOracle:     "Bladegraft Aspirant",
		s82B14LightbringerOracle: "Leonin Lightbringer",
		s82B14SalvoOracle:        "Rebel Salvo",
		s82B14ReunitedOracle:     "Resistance Reunited",
		s82B14SleeperOracle:      "Awaken the Sleeper",
		s82B14RebukeOracle:       "Molten Rebuke",
		s82B14SwordOracle:        "Sword of Forge and Frontier",
		s82B14GambitOracle:       "Goldwardens' Gambit",
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

// s82B14Cost is the generic part of the price `me` would pay for a spell.
func s82B14Cost(t *testing.T, g *game.Game, me *game.Player, name, typeLine, manaCost, oracle string) int {
	t.Helper()
	id := effectstest.HandCardFull(me, name, typeLine, manaCost, oracle, nil)
	var card game.Card
	for _, c := range me.Hand.Cards {
		if c.InstanceID == id {
			card = c
		}
	}
	price, err := g.PriceCast(me.ID, card, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("PriceCast %s: %v", name, err)
	}
	return price.Total.Generic
}

func TestS82OneB14BladeholdWarWhipGivesDoubleStrikeAndDiscountsOtherEquip(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	whip, rebel := effectstest.CastForMirrodin(t, g, "Bladehold War-Whip", s82B14WarWhipOracle)
	effectstest.ForMirrodinCheck(t, g, "the armed Rebel", rebel, 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, rebel, "double strike") {
		t.Error("the equipped Rebel has no double strike")
	}
	bear := effectstest.SeedBear(g, me.ID)
	hammer := effectstest.SeedEquipment(g, me.ID, "Loxodon Warhammer", effectstest.WarhammerOracle)
	if got := effectstest.EquipPrice(t, g, me.ID, hammer, bear); got != "{2}" {
		t.Errorf("another Equipment's equip: %s, want {2} (printed {3}, {1} less)", got)
	}
	if got := effectstest.EquipPrice(t, g, me.ID, whip, bear); got != "{3}{R}{W}" {
		t.Errorf("the War-Whip's own equip: %s, want the printed {3}{R}{W}", got)
	}
}

func TestS82OneB14BladegraftAspirantDiscountsEquipmentSpellsAndEquipThatTargetsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMainForCost(t, g)
	if got := s82B14Cost(t, g, me, "Some Equipment", effectstest.EquipTypeLine, "{3}", ""); got != 3 {
		t.Fatalf("without the Aspirant: %d generic, want 3", got)
	}
	asp := effectstest.B12Push(g, me.ID, "Bladegraft Aspirant", "Creature — Phyrexian Warrior", s82B14AspirantOracle, 2, 3)
	if got := s82B14Cost(t, g, me, "Some Equipment", effectstest.EquipTypeLine, "{3}", ""); got != 2 {
		t.Errorf("Equipment spell: %d generic, want 2", got)
	}
	if got := s82B14Cost(t, g, me, "Some Sorcery", "Sorcery", "{3}", ""); got != 3 {
		t.Errorf("non-Equipment spell: %d generic, want 3", got)
	}
	bear := effectstest.SeedBear(g, me.ID)
	hammer := effectstest.SeedEquipment(g, me.ID, "Loxodon Warhammer", effectstest.WarhammerOracle)
	if got := effectstest.EquipPrice(t, g, me.ID, hammer, asp); got != "{2}" {
		t.Errorf("equip onto the Aspirant: %s, want {2}", got)
	}
	if got := effectstest.EquipPrice(t, g, me.ID, hammer, bear); got != "{3}" {
		t.Errorf("equip onto another creature: %s, want {3}", got)
	}
}

func TestS82OneB14LeoninLightbringerGrowsWhileEquipped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	cat := effectstest.B12Push(g, me.ID, "Leonin Lightbringer", "Creature — Cat Rebel", s82B14LightbringerOracle, 3, 2)
	if p, th := effectstest.EffectivePTOf(t, g, cat); p != 3 || th != 2 {
		t.Fatalf("unequipped: %d/%d, want 3/2", p, th)
	}
	hammer := effectstest.SeedEquipment(g, me.ID, "Loxodon Warhammer", effectstest.WarhammerOracle)
	effectstest.EquipTo(t, g, me.ID, hammer, cat)
	// Warhammer gives +3/+0; Lightbringer adds +1/+1 while equipped.
	if p, th := effectstest.EffectivePTOf(t, g, cat); p != 7 || th != 3 {
		t.Errorf("equipped: %d/%d, want 7/3", p, th)
	}
	spec, _ := effects.Lookup(s82B14LightbringerOracle)
	if len(spec.Triggered) != 1 {
		t.Errorf("Lightbringer has %d triggered abilities, want its ward", len(spec.Triggered))
	}
}

func TestS82OneB14GlimmerLensDrawsOnlyWithAnotherAttacker(t *testing.T) {
	for _, tc := range []struct {
		name  string
		other bool
		want  int
	}{{"alone", false, 0}, {"with another attacker", true, 1}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			bear := effectstest.SeedBear(g, me.ID)
			lens := effectstest.SeedEquipment(g, me.ID, "Glimmer Lens", s82B14GlimmerLensOracle)
			effectstest.EquipTo(t, g, me.ID, lens, bear)
			attackers := []uuid.UUID{bear}
			if tc.other {
				attackers = append(attackers, effectstest.SeedBear(g, me.ID))
			}
			before := me.Hand.Size()
			effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
			for _, id := range attackers {
				if err := g.DeclareAttacker(id, opp.ID); err != nil {
					t.Fatalf("DeclareAttacker: %v", err)
				}
			}
			effectstest.LockInAttacks(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			if got := me.Hand.Size() - before; got != tc.want {
				t.Errorf("drew %d, want %d", got, tc.want)
			}
		})
	}
}

func TestS82OneB14HexplateWallbreakerUntapsAttackersAndAddsOneCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	friend := effectstest.SeedBear(g, me.ID)
	wall := effectstest.SeedEquipment(g, me.ID, "Hexplate Wallbreaker", s82B14WallbreakerOracle)
	effectstest.EquipTo(t, g, me.ID, wall, bear)
	if p, th := effectstest.EffectivePTOf(t, g, bear); p != 4 || th != 4 {
		t.Fatalf("equipped: %d/%d, want 4/4", p, th)
	}
	effectstest.DeclareAttack(t, g, opp.ID, bear, friend)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B13Tapped(t, g, bear) || effectstest.B13Tapped(t, g, friend) {
		t.Error("an attacking creature stayed tapped")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	if next := effectstest.StepOnce(t, g); next.Step != game.StepBeginCombat {
		t.Fatalf("after the first combat: %+v, want an additional combat", next)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker in the second combat: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B13Tapped(t, g, bear) {
		t.Error("the added combat's attack untapped the creature (not the first combat)")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	if next := effectstest.StepOnce(t, g); next.Step == game.StepBeginCombat {
		t.Error("the additional combat added a third")
	}
}

func TestS82OneB14SwordOfForgeAndFrontierPumpsProtectsAndExilesTwoToPlay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Two", TypeLine: "Sorcery"},
		game.Card{Name: "Third", TypeLine: "Sorcery"},
	)
	bear := effectstest.SeedBear(g, me.ID)
	sword := effectstest.SeedEquipment(g, me.ID, "Sword of Forge and Frontier", s82B14SwordOracle)
	effectstest.EquipTo(t, g, me.ID, sword, bear)
	if p, th := effectstest.EffectivePTOf(t, g, bear); p != 4 || th != 4 {
		t.Errorf("equipped: %d/%d, want 4/4", p, th)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "protection from red") || !effectstest.HasEffectiveKeyword(t, g, bear, "protection from green") {
		t.Error("the equipped creature lacks protection from red and from green")
	}
	effectstest.AttackWith(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(ids[0]) || !g.Exile.Contains(ids[1]) || g.Exile.Contains(ids[2]) {
		t.Fatalf("exile holds the top two only: %v %v %v", g.Exile.Contains(ids[0]), g.Exile.Contains(ids[1]), g.Exile.Contains(ids[2]))
	}
	if g.CastPermissionOnCardByIDForEffect(ids[0]) == nil || g.CastPermissionOnCardByIDForEffect(ids[1]) == nil {
		t.Error("an exiled card has no permission to be played")
	}
	if got := g.LandDropsRemainingFor(me.ID); got != 2 {
		t.Errorf("land drops left %d, want 2 (one extra)", got)
	}
}

func TestS82OneB14RebelSalvoCostsLessAndStripsIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	if got := s82B14Cost(t, g, me, "Rebel Salvo", "Instant", "{2}{R}", s82B14SalvoOracle); got != 2 {
		t.Fatalf("no Equipment: %d generic, want 2", got)
	}
	effectstest.SeedEquipment(g, me.ID, "Loxodon Warhammer", effectstest.WarhammerOracle)
	effectstest.SeedEquipment(g, me.ID, "Loxodon Warhammer", effectstest.WarhammerOracle)
	if got := s82B14Cost(t, g, me, "Rebel Salvo", "Instant", "{2}{R}", s82B14SalvoOracle); got != 0 {
		t.Errorf("two Equipment: %d generic, want 0", got)
	}
	god := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Stone Wall", TypeLine: effectstest.TestCreatureTypeLine,
		Power: 0, Toughness: 5, Keywords: []string{"indestructible"}, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Rebel Salvo", "Instant", s82B14SalvoOracle, []game.TargetRef{{Kind: game.TargetCard, ID: god}})
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(god) {
		t.Error("the indestructible creature survived 5 damage")
	}
}

func TestS82OneB14RebelSalvoHitsAPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	pw := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{game.CounterLoyalty: 6}, Owner: opp.ID, Controller: opp.ID,
	})
	if err := func() error {
		return effectstest.CastCatalogSpellErr(t, g, "Rebel Salvo", "Instant", s82B14SalvoOracle, []game.TargetRef{{Kind: game.TargetCard, ID: opp.ID}})
	}(); err == nil {
		t.Error("Rebel Salvo targeted a player")
	}
	effectstest.CastCatalogSpell(t, g, "Rebel Salvo", "Instant", s82B14SalvoOracle, []game.TargetRef{{Kind: game.TargetCard, ID: pw}})
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if got := effectstest.LoyaltyOf(g, pw); got != 1 {
		t.Errorf("the planeswalker has %d loyalty, want 1 (6 - 5)", got)
	}
}

func TestS82OneB14MoltenRebukeModes(t *testing.T) {
	setup := func() (*game.Game, uuid.UUID, uuid.UUID) {
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[1]
		bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Bear", TypeLine: effectstest.TestCreatureTypeLine,
			Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
		})
		eq := effectstest.SeedEquipment(g, opp.ID, "Loxodon Warhammer", effectstest.WarhammerOracle)
		return g, bear, eq
	}
	t.Run("both", func(t *testing.T) {
		g, bear, eq := setup()
		effectstest.B02cCastModalSpell(t, g, "Molten Rebuke", "Sorcery", s82B14RebukeOracle, []int{0, 1},
			[]game.TargetRef{effectstest.P1g3ModeCard(bear, 0), effectstest.P1g3ModeCard(eq, 1)})
		effectstest.PassPriorityAroundTable(t, g)
		g.RunStateChecksForTest()
		if g.Battlefield.Contains(bear) || g.Battlefield.Contains(eq) {
			t.Errorf("bear on battlefield %v, Equipment on battlefield %v; want both gone",
				g.Battlefield.Contains(bear), g.Battlefield.Contains(eq))
		}
	})
	t.Run("damage only", func(t *testing.T) {
		g, bear, eq := setup()
		effectstest.B02cCastModalSpell(t, g, "Molten Rebuke", "Sorcery", s82B14RebukeOracle, []int{0},
			[]game.TargetRef{effectstest.P1g3ModeCard(bear, 0)})
		effectstest.PassPriorityAroundTable(t, g)
		g.RunStateChecksForTest()
		if g.Battlefield.Contains(bear) || !g.Battlefield.Contains(eq) {
			t.Error("only the creature should die")
		}
	})
	t.Run("destroy only", func(t *testing.T) {
		g, bear, eq := setup()
		effectstest.B02cCastModalSpell(t, g, "Molten Rebuke", "Sorcery", s82B14RebukeOracle, []int{1},
			[]game.TargetRef{effectstest.P1g3ModeCard(eq, 0)})
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(bear) || g.Battlefield.Contains(eq) {
			t.Error("only the Equipment should be destroyed")
		}
	})
}

func TestS82OneB14ResistanceReunitedPumpsTargetAndProtectsEquippedCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	armed := effectstest.SeedBear(g, me.ID)
	bare := effectstest.SeedBear(g, me.ID)
	hammer := effectstest.SeedEquipment(g, me.ID, "Loxodon Warhammer", effectstest.WarhammerOracle)
	effectstest.EquipTo(t, g, me.ID, hammer, armed)
	effectstest.CastCatalogSpell(t, g, "Resistance Reunited", "Instant", s82B14ReunitedOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bare}})
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, bare); p != 4 || th != 4 {
		t.Errorf("target is %d/%d, want 4/4", p, th)
	}
	if !effectstest.HasEffectiveKeyword(t, g, armed, "indestructible") {
		t.Error("the equipped creature did not gain indestructible")
	}
	if effectstest.HasEffectiveKeyword(t, g, bare, "indestructible") {
		t.Error("an unequipped creature gained indestructible")
	}
	// The set is fixed as the spell resolves: equipping the other one
	// afterwards earns it nothing (CR 611.2c).
	effectstest.EquipTo(t, g, me.ID, hammer, bare)
	if effectstest.HasEffectiveKeyword(t, g, bare, "indestructible") {
		t.Error("a creature equipped after resolution gained indestructible")
	}
}

func TestS82OneB14AwakenTheSleeperStealsAndMayDestroyTheEquipment(t *testing.T) {
	for _, destroy := range []bool{true, false} {
		name := "keep"
		if destroy {
			name = "destroy"
		}
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			victim := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: "Bear", TypeLine: effectstest.TestCreatureTypeLine,
				Power: 2, Toughness: 2, Tapped: true, Owner: opp.ID, Controller: opp.ID,
			})
			eq := effectstest.SeedEquipment(g, opp.ID, "Loxodon Warhammer", effectstest.WarhammerOracle)
			g.WithWriteLock(func() {
				_ = g.AttachForEffect(eq, game.TargetRef{Kind: game.TargetCard, ID: victim})
			})
			effectstest.CastCatalogSpell(t, g, "Awaken the Sleeper", "Sorcery", s82B14SleeperOracle, []game.TargetRef{{Kind: game.TargetCard, ID: victim}})
			effectstest.PassPriorityAroundTable(t, g)
			c, _ := effectstest.BattlefieldCard(g, victim)
			if c.Controller != me.ID {
				t.Fatalf("controller is %s, want me", c.Controller)
			}
			if effectstest.B13Tapped(t, g, victim) || !effectstest.HasEffectiveKeyword(t, g, victim, "haste") {
				t.Error("the stolen creature should be untapped and have haste")
			}
			prompt := effectstest.ConfirmChoiceFor(g, me.ID)
			if prompt == nil {
				t.Fatal("no \"you may destroy\" prompt for an equipped creature")
			}
			if err := g.ResolveConfirm(prompt.ID, me.ID, destroy); err != nil {
				t.Fatalf("ResolveConfirm: %v", err)
			}
			if got := !g.Battlefield.Contains(eq); got != destroy {
				t.Errorf("Equipment destroyed = %v, want %v", got, destroy)
			}
		})
	}
}

func TestS82OneB14AwakenTheSleeperAsksNothingOfAnUnequippedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	victim := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: effectstest.TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Awaken the Sleeper", "Sorcery", s82B14SleeperOracle, []game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ConfirmChoiceFor(g, me.ID) != nil {
		t.Error("asked about destroying Equipment on an unequipped creature")
	}
}

func TestS82OneB14GoldwardensGambitMakesFiveHastyRebelsAndOffersEachAnEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	if got := s82B14Cost(t, g, me, "Goldwardens' Gambit", "Sorcery", "{6}{R}{R}", s82B14GambitOracle); got != 6 {
		t.Fatalf("no Equipment: %d generic, want 6", got)
	}
	hammer := effectstest.SeedEquipment(g, me.ID, "Loxodon Warhammer", effectstest.WarhammerOracle)
	if got := s82B14Cost(t, g, me, "Goldwardens' Gambit", "Sorcery", "{6}{R}{R}", s82B14GambitOracle); got != 5 {
		t.Errorf("one Equipment: %d generic, want 5", got)
	}
	effectstest.CastCatalogSpell(t, g, "Goldwardens' Gambit", "Sorcery", s82B14GambitOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var rebels []uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Rebel" {
				rebels = append(rebels, c.InstanceID)
			}
		}
	})
	if len(rebels) != 5 {
		t.Fatalf("%d Rebels, want 5", len(rebels))
	}
	for _, r := range rebels {
		effectstest.ForMirrodinCheck(t, g, "a Rebel", r, 2, 2, "haste")
		_ = r
	}
	// The first Rebel is offered the Hammer and takes it; the rest are
	// offered it too (it may be moved) and decline.
	effectstest.AnswerChooseCards(t, g, me.ID, hammer)
	host := effectstest.AttachmentHostOf(t, g, hammer)
	if host.Kind != game.TargetCard || !containsUUID(rebels, host.ID) {
		t.Fatalf("the Hammer is attached to %+v, want a Rebel", host)
	}
	if p, th := effectstest.EffectivePTOf(t, g, host.ID); p != 5 || th != 2 {
		t.Errorf("the armed Rebel is %d/%d, want 5/2", p, th)
	}
	for i := 0; i < 4; i++ {
		if effectstest.ChooseCardsChoiceFor(g, me.ID) == nil {
			t.Fatalf("no offer for Rebel %d", i+2)
		}
		effectstest.AnswerChooseCards(t, g, me.ID)
	}
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a sixth offer is open")
	}
}

func containsUUID(ids []uuid.UUID, id uuid.UUID) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}
