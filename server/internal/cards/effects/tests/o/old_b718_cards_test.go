package testso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	b718GhituFireEater     = "3e792338-4a21-4a06-ad94-39931118e06c"
	b718Parch              = "7366374a-db19-4fb0-9c35-a6e466a9bbc3"
	b718LastDitchEffort    = "cdd8d7cf-28f2-4a4d-a373-40d48bff615b"
	b718Rivalry            = "da77cf1b-eaeb-4a60-aa16-a1217ce8b550"
	b718AngelsTrumpet      = "f187f766-3696-419e-9865-58023a46446b"
	b718Repopulate         = "3a225d39-c1f8-4ab0-98ec-dad63bf268db"
	b718MultanisPresence   = "adad1be7-8d3d-4857-a3b0-be33db4c5746"
	b718HarmonicConverge   = "400adeec-11b4-4a9c-9466-dcab524d87ea"
	b718BurstOfEnergy      = "b795ecfd-31ac-4a2a-9679-97b088932d93"
	b718YavimayaScion      = "6e06b5cc-444c-49c9-b08a-81ef8d97c9ab"
	b718WeatherseedFaeries = "904b2caf-e083-4917-8841-95f48d90dc0c"
)

func b718Ready(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, id).SummonedThisTurn = false })
}

func b718Lands(g *game.Game, p *game.Player, n int) {
	for i := 0; i < n; i++ {
		effectstest.B12Permanent(g, p.ID, "Mountain", "Basic Land — Mountain")
	}
}

func TestB718GhituFireEaterDealsItsPowerThenLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eater := effectstest.B12Push(g, me.ID, "Ghitu Fire-Eater", "Creature — Human Nomad", b718GhituFireEater, 2, 2)
	b718Ready(g, eater)
	before := opp.Life
	effectstest.P7Activate(t, g, me, eater, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	if effects.OnBattlefield(g, eater) {
		t.Error("it was not sacrificed")
	}
	if opp.Life != before-2 {
		t.Errorf("life %d, want %d", opp.Life, before-2)
	}
}

func TestB718GhituFireEaterReadsItsPowerAsItLastWas(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eater := effectstest.B12Push(g, me.ID, "Ghitu Fire-Eater", "Creature — Human Nomad", b718GhituFireEater, 2, 2)
	b718Ready(g, eater)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(eater, game.CounterPlusOne, 2); err != nil {
			t.Fatal(err)
		}
	})
	before := opp.Life
	effectstest.P7Activate(t, g, me, eater, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	if opp.Life != before-4 {
		t.Errorf("life %d, want %d (a 4/4 as it left)", opp.Life, before-4)
	}
}

func TestB718GhituFireEaterNeedsToTap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eater := effectstest.B12Push(g, me.ID, "Ghitu Fire-Eater", "Creature — Human Nomad", b718GhituFireEater, 2, 2)
	b718Ready(g, eater)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, eater).Tapped = true })
	err := g.ActivateCatalogAbility(me.ID, eater, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	if err == nil {
		t.Fatal("a tapped creature paid {T}")
	}
}

func TestB718ParchDealsTwoToAnyTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	before := opp.Life
	effectstest.CastModal(t, g, "Parch", "Instant", b718Parch, []int{0}, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("life %d, want %d", opp.Life, before-2)
	}
}

func TestB718ParchDealsFourToABlueCreatureOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	blue := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Faerie", TypeLine: "Creature — Faerie", Power: 1, Toughness: 4, Colors: []string{"U"}})
	green := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 4, Colors: []string{"G"}})
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Parch", TypeLine: "Instant", OracleID: b718Parch, Owner: active.ID, Controller: active.ID})
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Modes: []int{1}, Targets: effectstest.CardTarget(green)}); err == nil {
		t.Fatal("the blue-creature mode targeted a green creature")
	}
	effectstest.CastModal(t, g, "Parch", "Instant", b718Parch, []int{1}, effectstest.CardTarget(blue))
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, blue) {
		t.Error("the blue creature survived 4 damage")
	}
	if !effects.OnBattlefield(g, green) {
		t.Error("the green creature was destroyed")
	}
}

func TestB718LastDitchEffortDealsOneDamagePerCreatureSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, me.ID, "C", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	before := opp.Life
	effectstest.CastModal(t, g, "Last-Ditch Effort", "Instant", b718LastDitchEffort, nil, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil {
		t.Fatal("no sacrifice prompt")
	}
	if len(pick.ChooseCards) != 3 || effectstest.HasID(pick.ChooseCards, theirs) {
		t.Fatalf("offered %v, want exactly my three creatures", pick.ChooseCards)
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, []uuid.UUID{a, b}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, a) || effects.OnBattlefield(g, b) || !effects.OnBattlefield(g, c) {
		t.Error("the sacrifice was not exactly the two picked creatures")
	}
	if opp.Life != before-2 {
		t.Errorf("life %d, want %d", opp.Life, before-2)
	}
}

func TestB718LastDitchEffortMaySacrificeNone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	before := opp.Life
	effectstest.CastModal(t, g, "Last-Ditch Effort", "Instant", b718LastDitchEffort, nil, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil {
		t.Fatal("no sacrifice prompt")
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, a) || opp.Life != before {
		t.Errorf("a creature was sacrificed or damage dealt: life %d", opp.Life)
	}
}

// b718AtUpkeepOf walks to a seat's upkeep (past the opening one, whose
// beginning has gone) and settles what its beginning put on the stack.
func b718AtUpkeepOf(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; i < 300; i++ {
		if g.Turn.Step == game.StepUpkeep && g.Turn.ActiveSeat == seat && i > 0 {
			break
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestB718RivalryDamagesThePlayerWithStrictlyMostLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, p1, p2 := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Rivalry", b718Rivalry, "Enchantment", 0, 0))
	b718Lands(g, p1, 3)
	b718Lands(g, p2, 2)
	b718Lands(g, me, 1)
	b4, b1 := p1.Life, p2.Life
	b718AtUpkeepOf(t, g, 1)
	if p1.Life != b4-2 {
		t.Errorf("the player with the most lands is at %d, want %d", p1.Life, b4-2)
	}
	b718AtUpkeepOf(t, g, 2)
	if p2.Life != b1 {
		t.Errorf("a player who does not control the most lands took damage: %d, want %d", p2.Life, b1)
	}
}

func TestB718RivalryNeedsMoreThanEveryOtherPlayerNotATie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, p1, p2 := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Rivalry", b718Rivalry, "Enchantment", 0, 0))
	b718Lands(g, p1, 3)
	b718Lands(g, p2, 3)
	before := p1.Life
	b718AtUpkeepOf(t, g, 1)
	if p1.Life != before {
		t.Errorf("a tie for most lands took damage: %d", p1.Life)
	}
}

func TestB718AngelsTrumpetGivesEveryCreatureVigilance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Angel's Trumpet", OracleID: b718AngelsTrumpet, TypeLine: "Artifact"})
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{mine, theirs} {
		if c := effectstest.ApaLive(g, id); c == nil || !game.HasKeyword(c, "vigilance") {
			t.Error("a creature lacks vigilance")
		}
	}
}

func TestB718AngelsTrumpetTapsTheActivePlayersIdleCreaturesAndHurtsThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Angel's Trumpet", OracleID: b718AngelsTrumpet, TypeLine: "Artifact"})
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 1, 9)
	idleA := effectstest.B12Creature(g, me.ID, "IdleA", "Creature — Bear", 2, 2)
	idleB := effectstest.B12Creature(g, me.ID, "IdleB", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	before := me.Life
	effectstest.BrAttack(t, g, attacker)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{idleA, idleB} {
		if c := effectstest.ApaLive(g, id); c == nil || !c.Tapped {
			t.Error("an idle creature was not tapped")
		}
	}
	if c := effectstest.ApaLive(g, attacker); c == nil || c.Tapped {
		t.Error("the creature that attacked was tapped (it has vigilance)")
	}
	if c := effectstest.ApaLive(g, theirs); c == nil || c.Tapped {
		t.Error("another player's creature was tapped on my end step")
	}
	if me.Life != before-2 {
		t.Errorf("life %d, want %d: one damage for each of the two creatures tapped", me.Life, before-2)
	}
}

func TestB718RepopulateShufflesOnlyCreatureCardsIntoTheLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	var creatures []uuid.UUID
	for i := 0; i < 2; i++ {
		id := uuid.New()
		creatures = append(creatures, id)
		opp.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Bear", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})
	}
	spell := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Bolt", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	libBefore := opp.Library.Size()
	effectstest.CastModal(t, g, "Repopulate", "Instant", b718Repopulate, nil, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Library.Contains(creatures[0]) || !opp.Library.Contains(creatures[1]) {
		t.Error("a creature card is not in the library")
	}
	if opp.Library.Size() != libBefore+2 {
		t.Errorf("library %d, want %d", opp.Library.Size(), libBefore+2)
	}
	if !opp.Graveyard.Contains(spell) {
		t.Error("the noncreature card left the graveyard")
	}
}

func TestB718MultanisPresenceDrawsWhenASpellYouCastIsCountered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Multani's Presence", b718MultanisPresence, "Enchantment", 0, 0))
	for i := 0; i < 3; i++ {
		me.Library.PushTop(game.NewCard("Filler", me.ID))
	}
	spell := effectstest.B18CastFromHand(t, g, me, "Opt", "Instant", "", game.CastSpellParams{})
	hand := effectstest.HandCount(g, me)
	g.WithWriteLock(func() {
		if err := g.CounterTargetForEffect(spell); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandCount(g, me); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

func TestB718MultanisPresenceIgnoresAnOpponentsCounteredSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.StCard("Multani's Presence", b718MultanisPresence, "Enchantment", 0, 0))
	for i := 0; i < 3; i++ {
		opp.Library.PushTop(game.NewCard("Filler", opp.ID))
	}
	spell := effectstest.B18CastFromHand(t, g, me, "Opt", "Instant", "", game.CastSpellParams{})
	hand := effectstest.HandCount(g, opp)
	g.WithWriteLock(func() {
		if err := g.CounterTargetForEffect(spell); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandCount(g, opp); got != hand {
		t.Errorf("hand %d, want %d: it drew for an opponent's spell", got, hand)
	}
}

func TestB718HarmonicConvergencePutsEveryEnchantmentOnTopOfItsOwnersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Ench A", TypeLine: "Enchantment"})
	b := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Ench B", TypeLine: "Enchantment"})
	stolen := effectstest.ApaPush(g, opp.ID, me.ID, game.Card{Name: "Ench C", TypeLine: "Enchantment"})
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastModal(t, g, "Harmonic Convergence", "Instant", b718HarmonicConverge, nil, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no ordering prompt for an owner of two enchantments")
	}
	if err := g.ResolvePutInLibrary(pick.ID, me.ID, nil, []uuid.UUID{b, a}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, a) || effects.OnBattlefield(g, b) || effects.OnBattlefield(g, stolen) {
		t.Fatal("an enchantment stayed on the battlefield")
	}
	if !effects.OnBattlefield(g, bear) {
		t.Error("a creature left")
	}
	n := len(me.Library.Cards)
	if me.Library.Cards[n-1].InstanceID != b || me.Library.Cards[n-2].InstanceID != a {
		t.Error("the owner's chosen order (B above A) was not kept")
	}
	if got := opp.Library.Cards[len(opp.Library.Cards)-1].InstanceID; got != stolen {
		t.Error("the stolen enchantment did not go to its owner's library")
	}
	if me.Library.Contains(stolen) {
		t.Error("the stolen enchantment went to its controller's library")
	}
}

func TestB718BurstOfEnergyUntapsAPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.B12Permanent(g, opp.ID, "Mountain", "Basic Land — Mountain")
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, land).Tapped = true })
	_ = me
	effectstest.CastModal(t, g, "Burst of Energy", "Instant", b718BurstOfEnergy, nil, effectstest.CardTarget(land))
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardForTest(g, land); c == nil || c.Tapped {
		t.Error("the permanent is still tapped")
	}
}

func TestB718WeatherseedFaeriesCantBeTargetedByARedSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	fae := effectstest.B12Push(g, opp.ID, "Weatherseed Faeries", "Creature — Faerie", b718WeatherseedFaeries, 2, 1)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, fae).Colors = []string{"U"} })
	if c := effectstest.ApaLive(g, fae); c == nil || !game.HasKeyword(c, "flying") || !game.HasKeyword(c, "protection from red") {
		t.Fatal("it lacks flying or protection from red")
	}
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Parch", TypeLine: "Instant", OracleID: b718Parch,
		ManaCost: "{1}{R}", Colors: []string{"R"}, Owner: active.ID, Controller: active.ID})
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Modes: []int{0}, Targets: effectstest.CardTarget(fae)}); err == nil {
		t.Fatal("a red spell targeted a creature with protection from red")
	}
}

func TestB718YavimayaScionHasProtectionFromArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	scion := effectstest.B12Push(g, me.ID, "Yavimaya Scion", "Creature — Treefolk", b718YavimayaScion, 4, 4)
	if c := effectstest.ApaLive(g, scion); c == nil || !game.HasKeyword(c, "protection from artifacts") {
		t.Fatal("it lacks protection from artifacts")
	}
}
