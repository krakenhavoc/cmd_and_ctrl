package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b106_cards_test.go — Core Set 2021, batch 106 (#5511). Every test
// plays the real catalog entry through the engine.

const (
	m21B106FelineSovereign  = "05f0feb6-6783-49fd-acf6-68b44aaff846"
	m21B106KeralKeep        = "0a865453-ddc8-4403-8f1e-5752e3582458"
	m21B106SiegeStriker     = "12dd664c-a279-4adc-b6cb-bec22eedca5c"
	m21B106Masticore        = "1306c72f-fc6b-470b-a95e-0364e0e30ac7"
	m21B106Incinerator      = "515cee6a-a2e1-4eeb-8e1d-293724313b14"
	m21B106SanctumShattered = "5e39d311-efb5-4f51-82cd-e8079e32907a"
	m21B106VolcanicGeyser   = "846a4f9c-d955-403f-8a08-5c7c3d32e180"
	m21B106LeafkinAvenger   = "adff8c13-8909-47d0-acad-ba63597f2e96"
	m21B106DireFleet        = "ee08435b-a2c1-491a-87cb-e60789b09c25"
	m21B106ChandraNalaar    = "f4da7a43-e9d5-47ea-b2f0-3cf55a156f4a"

	m21B106Bear = "Creature — Bear"
	m21B106Cat  = "Creature — Cat"
)

func m21B106Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func TestM21B106EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		m21B106FelineSovereign:  "Feline Sovereign",
		m21B106KeralKeep:        "Keral Keep Disciples",
		m21B106SiegeStriker:     "Siege Striker",
		m21B106Masticore:        "Sparkhunter Masticore",
		m21B106Incinerator:      "Chandra's Incinerator",
		m21B106SanctumShattered: "Sanctum of Shattered Heights",
		m21B106VolcanicGeyser:   "Volcanic Geyser",
		m21B106LeafkinAvenger:   "Leafkin Avenger",
		m21B106DireFleet:        "Dire Fleet Warmonger",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestM21B106PrintedKeywords(t *testing.T) {
	for _, c := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		keywords               []string
	}{
		{"Siege Striker", "Creature — Human Soldier", m21B106SiegeStriker, 1, 1, []string{"double strike"}},
		{"Sparkhunter Masticore", "Artifact Creature — Masticore", m21B106Masticore, 3, 4, []string{"protection from planeswalkers"}},
		{"Chandra's Incinerator", "Creature — Elemental", m21B106Incinerator, 6, 6, []string{"trample"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, c.name, c.typeLine, c.oracle, c.power, c.toughness)
			effectstest.AssertKeywords(t, g, id, c.keywords...)
		})
	}
}

// --- Feline Sovereign -----------------------------------------------------------

func TestM21B106FelineSovereignPumpsOnlyYourOtherCats(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := m21B106Seats(g)
	sov := effectstest.B12Push(g, me.ID, "Feline Sovereign", m21B106Cat, m21B106FelineSovereign, 2, 3)
	mine := effectstest.B12Creature(g, me.ID, "Mine", m21B106Cat, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", m21B106Cat, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", m21B106Bear, 2, 2)
	for _, c := range []struct {
		name string
		id   uuid.UUID
		p, t int
		pro  bool
	}{{"the Sovereign", sov, 2, 3, false}, {"your other Cat", mine, 3, 3, true}, {"their Cat", theirs, 2, 2, false}, {"your Bear", bear, 2, 2, false}} {
		if p, tough := effectstest.PtOf(t, g, c.id); p != c.p || tough != c.t {
			t.Errorf("%s is %d/%d, want %d/%d", c.name, p, tough, c.p, c.t)
		}
		if got := effectstest.HasEffectiveKeyword(t, g, c.id, "protection from Dogs"); got != c.pro {
			t.Errorf("%s has protection from Dogs: %v, want %v", c.name, got, c.pro)
		}
	}
}

func TestM21B106FelineSovereignDestroysAnArtifactOrEnchantmentOfThePlayerHit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := m21B106Seats(g)
	third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "Feline Sovereign", m21B106Cat, m21B106FelineSovereign, 2, 3)
	cat := effectstest.B12Creature(g, me.ID, "Mine", m21B106Cat, 2, 2)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	aura := effectstest.B12Permanent(g, opp.ID, "Ward", "Enchantment")
	other := effectstest.B12Permanent(g, third.ID, "Other Rock", "Artifact")
	effectstest.DeclareAttack(t, g, opp.ID, cat)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.HobAnswerTargets(t, g, me.ID, other, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact of the player who was hit survived")
	}
	if !g.Battlefield.Contains(other) || !g.Battlefield.Contains(aura) {
		t.Error("only the one chosen target is destroyed, and only from the player who was hit")
	}
}

func TestM21B106FelineSovereignIsOneTriggerForSeveralCats(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := m21B106Seats(g)
	sov := effectstest.B12Push(g, me.ID, "Feline Sovereign", m21B106Cat, m21B106FelineSovereign, 2, 3)
	cat := effectstest.B12Creature(g, me.ID, "Mine", m21B106Cat, 2, 2)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	rock2 := effectstest.B12Permanent(g, opp.ID, "Rock Two", "Artifact")
	effectstest.DeclareAttack(t, g, opp.ID, sov, cat)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.HobAnswerTargets(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) == g.Battlefield.Contains(rock2) {
		t.Error("two Cats hitting together destroy exactly one permanent (one trigger)")
	}
}

// --- Keral Keep Disciples -------------------------------------------------------

func TestM21B106KeralKeepPingsEachOpponentForAChandraLoyaltyAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := m21B106Seats(g)
	effectstest.B12Push(g, me.ID, "Keral Keep Disciples", "Creature — Human Monk", m21B106KeralKeep, 4, 3)
	chandra := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Chandra Nalaar",
		OracleID: m21B106ChandraNalaar, TypeLine: "Legendary Planeswalker — Chandra", Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterLoyalty: 6}})
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		lives[s.ID] = s.Life
	}
	if err := g.ActivateCatalogAbility(me.ID, chandra, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		want := lives[s.ID] - 1
		if s.ID == me.ID {
			want = lives[s.ID]
		}
		if s.ID == opp.ID {
			want = lives[s.ID] - 2 // the +1's own damage as well
		}
		if s.Life != want {
			t.Errorf("%s is at %d, want %d", s.Name, s.Life, want)
		}
	}
}

func TestM21B106KeralKeepIgnoresOtherPlaneswalkers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := m21B106Seats(g)
	effectstest.B12Push(g, me.ID, "Keral Keep Disciples", "Creature — Human Monk", m21B106KeralKeep, 4, 3)
	lili := effectstest.PushWalkerForTest(g, me.ID, "Liliana of the Veil", effectstest.LilianaOfTheVeilOracle, 3)
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		lives[s.ID] = s.Life
	}
	if err := g.ActivateCatalogAbility(me.ID, lili, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerAllPrompts(t, g, me.ID)
	for _, s := range g.Seats {
		if s.Life != lives[s.ID] {
			t.Errorf("%s lost life to a non-Chandra loyalty ability", s.Name)
		}
	}
}

// --- Siege Striker --------------------------------------------------------------

func TestM21B106SiegeStrikerGrowsForEachCreatureTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := m21B106Seats(g)
	striker := effectstest.B12Push(g, me.ID, "Siege Striker", "Creature — Human Soldier", m21B106SiegeStriker, 1, 1)
	a := effectstest.B12Creature(g, me.ID, "Bear A", m21B106Bear, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", m21B106Bear, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, striker)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, striker); p != 3 || tough != 3 {
		t.Errorf("the Striker is %d/%d after tapping two, want 3/3", p, tough)
	}
	if !effectstest.TappedForTest(t, g, a) || !effectstest.TappedForTest(t, g, b) {
		t.Error("the chosen creatures were not tapped")
	}
}

func TestM21B106SiegeStrikerMayTapNone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := m21B106Seats(g)
	striker := effectstest.B12Push(g, me.ID, "Siege Striker", "Creature — Human Soldier", m21B106SiegeStriker, 1, 1)
	a := effectstest.B12Creature(g, me.ID, "Bear A", m21B106Bear, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, striker)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, striker); p != 1 {
		t.Errorf("power %d after tapping nothing, want 1", p)
	}
	if effectstest.TappedForTest(t, g, a) {
		t.Error("a creature was tapped though none was chosen")
	}
}

// --- Sparkhunter Masticore ------------------------------------------------------

func TestM21B106MasticoreNeedsADiscardToCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := m21B106Seats(g)
	effectstest.AdvanceToMain(t, g)
	for me.Hand.Size() > 0 {
		_, _ = me.Hand.PopTop()
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Sparkhunter Masticore", TypeLine: "Artifact Creature — Masticore", OracleID: m21B106Masticore,
		ManaCost: "{3}", Power: 3, Toughness: 4, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast with no card to discard")
	}
	filler := effectstest.EoHand(me, "Filler", "Instant")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{DiscardIDs: []uuid.UUID{filler}}); err != nil {
		t.Fatalf("cast with a discard: %v", err)
	}
	if me.Hand.Contains(filler) || !me.Graveyard.Contains(filler) {
		t.Error("the discard was not paid")
	}
}

func TestM21B106MasticorePingsOnlyAPlaneswalkerAndGainsIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := m21B106Seats(g)
	mast := effectstest.B12Push(g, me.ID, "Sparkhunter Masticore", "Artifact Creature — Masticore", m21B106Masticore, 3, 4)
	walker := effectstest.PushWalkerForTest(g, opp.ID, "Walker", effectstest.LilianaOfTheVeilOracle, 5)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", m21B106Bear, 2, 2)
	effectstest.ApaMana(me, "R")
	if err := g.ActivateCatalogAbility(me.ID, mast, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Fatal("the ping targeted a creature")
	}
	effectstest.ApaActivate(t, g, me, mast, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(walker)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyOf(g, walker); got != 4 {
		t.Errorf("loyalty %d after one ping, want 4", got)
	}
	if effectstest.HasEffectiveKeyword(t, g, mast, "indestructible") {
		t.Fatal("indestructible before paying for it")
	}
	effectstest.ApaMana(me, "R", "R", "R")
	effectstest.ApaActivate(t, g, me, mast, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, mast, "indestructible") {
		t.Error("no indestructible after {3}")
	}
}

func TestM21B106MasticoreCantBeDamagedByAPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := m21B106Seats(g)
	mast := effectstest.B12Push(g, opp.ID, "Sparkhunter Masticore", "Artifact Creature — Masticore", m21B106Masticore, 3, 4)
	chandra := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Chandra Nalaar",
		OracleID: m21B106ChandraNalaar, TypeLine: "Legendary Planeswalker — Chandra", Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterLoyalty: 6}})
	if err := g.ActivateCatalogAbility(me.ID, chandra, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(mast), XValue: 3}); err == nil {
		t.Error("a planeswalker targeted a creature with protection from planeswalkers")
	}
}

// --- Chandra's Incinerator ------------------------------------------------------

func TestM21B106IncineratorCostsLessByNoncombatDamageToOpponents(t *testing.T) {
	for _, dealt := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := m21B106Seats(g)
		effectstest.AdvanceToMain(t, g)
		if dealt {
			effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
			effectstest.PassPriorityAroundTable(t, g)
		}
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Chandra's Incinerator", TypeLine: "Creature — Elemental", OracleID: m21B106Incinerator,
			ManaCost: "{5}{R}", Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID})
		effectstest.ApaMana(me, "R", "R", "R")
		err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
		if dealt && err != nil {
			t.Errorf("3 noncombat damage dealt to an opponent: {2}{R} should cast it: %v", err)
		}
		if !dealt && err == nil {
			t.Error("cast for {R}{R}{R} with no damage dealt")
		}
	}
}

func TestM21B106IncineratorRedirectsNoncombatDamageToACreatureOrPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := m21B106Seats(g)
	effectstest.B12Push(g, me.ID, "Chandra's Incinerator", "Creature — Elemental", m21B106Incinerator, 6, 6)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", m21B106Bear, 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "Mine", m21B106Bear, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.HobAnswerTargets(t, g, me.ID, mine, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the opponent's creature took no damage")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("my own creature was hit")
	}
}

// --- Volcanic Geyser, Sanctum, Avenger, Warmonger -------------------------------

func TestM21B106VolcanicGeyserDealsX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := m21B106Seats(g)
	life := opp.Life
	effectstest.CastXSpell(t, g, "Volcanic Geyser", "Instant", m21B106VolcanicGeyser, "{X}{R}{R}", 4, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("life %d, want %d", opp.Life, life-4)
	}
}

func TestM21B106SanctumOfShatteredHeightsDamagesByShrines(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := m21B106Seats(g)
	sanctum := effectstest.SeedPermanentWithOracle(g, me.ID, "Sanctum of Shattered Heights", "Legendary Enchantment — Shrine", m21B106SanctumShattered)
	effectstest.B12Permanent(g, me.ID, "Other Shrine", "Enchantment — Shrine")
	effectstest.B12Permanent(g, opp.ID, "Their Shrine", "Enchantment — Shrine")
	big := effectstest.B12Creature(g, opp.ID, "Big", m21B106Bear, 4, 4)
	small := effectstest.B12Creature(g, opp.ID, "Small", m21B106Bear, 2, 2)
	spell := effectstest.EoHand(me, "Sorcery", "Sorcery")
	land := effectstest.EoHand(me, "Forest", "Basic Land — Forest")
	effectstest.ApaMana(me, "R")
	if err := g.ActivateCatalogAbility(me.ID, sanctum, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(small), DiscardIDs: []uuid.UUID{spell}}); err == nil {
		t.Fatal("a nonland, non-Shrine card paid the discard")
	}
	me.ManaPool = game.ManaPool{}
	effectstest.ApaMana(me, "R")
	effectstest.ApaActivate(t, g, me, sanctum, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(small), DiscardIDs: []uuid.UUID{land}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("X = 2 Shrines you control should kill a 2/2")
	}
	if !g.Battlefield.Contains(big) {
		t.Error("the other creature should be untouched")
	}
}

func TestM21B106LeafkinAvengerAddsGForEachBigCreatureAndShootsPlayers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := m21B106Seats(g)
	avenger := effectstest.B12Push(g, me.ID, "Leafkin Avenger", "Creature — Elemental Druid", m21B106LeafkinAvenger, 4, 3)
	effectstest.B12Creature(g, me.ID, "Big", m21B106Bear, 5, 5)
	effectstest.B12Creature(g, me.ID, "Small", m21B106Bear, 3, 3)
	effectstest.B12Creature(g, opp.ID, "Their Big", m21B106Bear, 6, 6)
	if err := g.ActivateManaAbility(me.ID, avenger, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	if got := me.ManaPool.Total(); got != 2 {
		t.Errorf("added %d mana, want 2 (the Avenger and the 5/5)", got)
	}
	me.ManaPool = game.ManaPool{}
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, avenger).Tapped = false })
	life := opp.Life
	effectstest.ApaMana(me, "R", "R", "R", "R", "R", "R", "R", "R")
	effectstest.ApaActivate(t, g, me, avenger, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("life %d, want %d (damage equal to its power)", opp.Life, life-4)
	}
}

func TestM21B106DireFleetWarmongerSacrificesForABoost(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, _ := m21B106Seats(g)
		warmonger := effectstest.B12Push(g, me.ID, "Dire Fleet Warmonger", "Creature — Orc Pirate", m21B106DireFleet, 3, 3)
		bear := effectstest.B12Creature(g, me.ID, "Bear", m21B106Bear, 2, 2)
		effectstest.AdvanceTo(t, g, game.StepBeginCombat)
		effectstest.PassPriorityAroundTable(t, g)
		if yes {
			effectstest.AnswerOwnPermanents(t, g, me.ID, bear)
		} else {
			effectstest.AnswerOwnPermanents(t, g, me.ID)
		}
		effectstest.PassPriorityAroundTable(t, g)
		p, tough := effectstest.PtOf(t, g, warmonger)
		if yes {
			if p != 5 || tough != 5 || !effectstest.HasEffectiveKeyword(t, g, warmonger, "trample") || g.Battlefield.Contains(bear) {
				t.Errorf("after sacrificing: %d/%d, trample %v, bear alive %v", p, tough, effectstest.HasEffectiveKeyword(t, g, warmonger, "trample"), g.Battlefield.Contains(bear))
			}
		} else if p != 3 || tough != 3 || !g.Battlefield.Contains(bear) {
			t.Errorf("declined: %d/%d, bear alive %v", p, tough, g.Battlefield.Contains(bear))
		}
	}
}
