package testsru

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapB213ProwlerOracle   = "72e2aa4f-b4b4-42ab-adc0-d0ac09cd5f92"
	gapB213HezrouOracle    = "65d257c2-b3be-4244-a74c-bc4b5d7cedc3"
	gapB213StenchOracle    = gapB213HezrouOracle + "#1"
	gapB213DevourerOracle  = "a06fa827-900b-46d4-8f48-b4e15964951a"
	gapB213StrikeOracle    = "9be8e04f-f87f-4477-93f3-c546e9b346f0"
	gapB213MoldFolkOracle  = "d0ec32d9-1d4e-4b6f-bc95-addbb484b0d3"
	gapB213InvokerOracle   = "9bc9f027-8a2a-4447-8458-66ed40cb5b0d"
	gapB213ImpOracle       = "c412191f-46d7-4520-a76c-1f70a2e92ab6"
	gapB213NothicOracle    = "ec8c2927-d857-426e-b81b-c3b1aa81045a"
	gapB213AdventurerOracl = "1822f924-51ba-4f81-92cc-6c5d741a33f9"
	gapB213SarevokOracle   = "4e215ede-e4ac-482a-a2e1-cacb20affa70"
	gapB213SivrissOracle   = "ed15a3c4-110e-4f08-bf09-fabbf5b4eeb9"
)

func gapB213Mana(t *testing.T, g *game.Game, p *game.Player, n int) {
	t.Helper()
	effectstest.Acrb1AddMana(t, g, p.ID, strings.Repeat("{C}", n))
}

func TestGapB213EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB213ProwlerOracle:   "Guildsworn Prowler",
		gapB213HezrouOracle:    "Hezrou",
		gapB213StenchOracle:    "Demonic Stench",
		gapB213DevourerOracle:  "Intellect Devourer",
		gapB213StrikeOracle:    "Mardu Strike Leader",
		gapB213MoldFolkOracle:  "Mold Folk",
		gapB213InvokerOracle:   "Myrkul's Invoker",
		gapB213ImpOracle:       "Nefarious Imp",
		gapB213NothicOracle:    "Nothic",
		gapB213AdventurerOracl: "Ravenloft Adventurer",
		gapB213SarevokOracle:   "Sarevok, Deathbringer",
		gapB213SivrissOracle:   "Sivriss, Nightmare Speaker",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Myrkul's Invoker --------------------------------------------------

func TestGapB213InvokerPumpsAndGivesMenaceToYourCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	inv := effectstest.B12Push(g, me.ID, "Myrkul's Invoker", "Creature — Cat Rogue", gapB213InvokerOracle, 2, 3)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	gapB213Mana(t, g, me, 8)
	effectstest.B16Activate(t, g, me.ID, inv, 0, game.ActivateAbilityParams{})
	if p := effectstest.EffectivePower(t, g, mine); p != 4 {
		t.Errorf("my Bear power %d, want 4", p)
	}
	if p := effectstest.EffectivePower(t, g, inv); p != 4 {
		t.Errorf("the Invoker's power %d, want 4", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, mine, "menace") {
		t.Error("my Bear lacks menace")
	}
	if effectstest.EffectivePower(t, g, theirs) != 2 || effectstest.HasEffectiveKeyword(t, g, theirs, "menace") {
		t.Error("an opponent's creature was affected")
	}
}

func TestGapB213InvokerNeedsAllEightMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	inv := effectstest.B12Push(g, me.ID, "Myrkul's Invoker", "Creature — Cat Rogue", gapB213InvokerOracle, 2, 3)
	gapB213Mana(t, g, me, 7)
	if err := g.ActivateCatalogAbility(me.ID, inv, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with seven mana")
	}
}

// --- Mold Folk -----------------------------------------------------------

func TestGapB213MoldFolkSacrificesACreatureOrAnArtifactForACounter(t *testing.T) {
	for _, tc := range []struct{ name, typeLine string }{
		{"Bear", "Creature — Bear"}, {"Trinket", "Artifact"},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		folk := effectstest.B12Push(g, me.ID, "Mold Folk", "Creature — Fungus Warrior", gapB213MoldFolkOracle, 1, 1)
		fodder := effectstest.B12Permanent(g, me.ID, tc.name, tc.typeLine)
		gapB213Mana(t, g, me, 1)
		effectstest.B16Activate(t, g, me.ID, folk, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}})
		if n := effectstest.CounterCount(g, folk, game.CounterPlusOne); n != 1 {
			t.Errorf("%s: %d +1/+1 counters, want 1", tc.name, n)
		}
		if effectstest.B12ZoneOf(g, fodder) == game.ZoneBattlefield {
			t.Errorf("%s was not sacrificed", tc.name)
		}
	}
}

func TestGapB213MoldFolkCannotSacrificeItselfOrALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	folk := effectstest.B12Push(g, me.ID, "Mold Folk", "Creature — Fungus Warrior", gapB213MoldFolkOracle, 1, 1)
	land := effectstest.B12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	gapB213Mana(t, g, me, 1)
	for _, victim := range []uuid.UUID{folk, land} {
		if err := g.ActivateCatalogAbility(me.ID, folk, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{victim}}); err == nil {
			t.Errorf("sacrificed %s", victim)
		}
	}
}

// --- Mardu Strike Leader -------------------------------------------------

func TestGapB213StrikeLeaderMakesAWarriorWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	leader := effectstest.B12Push(g, me.ID, "Mardu Strike Leader", "Creature — Human Warrior", gapB213StrikeOracle, 3, 2)
	effectstest.DeclareAttack(t, g, opp.ID, leader)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B16CountNamed(g, "Warrior"); n != 1 {
		t.Fatalf("%d Warrior tokens, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Warrior" {
			if c.Power != 2 || c.Toughness != 1 || c.AttackingTarget != uuid.Nil {
				t.Errorf("token %d/%d attacking=%v, want a 2/1 that is not attacking", c.Power, c.Toughness, c.AttackingTarget != uuid.Nil)
			}
		}
	}
}

func TestGapB213StrikeLeaderHasDash(t *testing.T) {
	spec, ok := effects.Lookup(gapB213StrikeOracle)
	if !ok || len(spec.AlternativeCosts) != 1 || spec.AlternativeCosts[0].ManaCost != "{3}{B}" || !spec.AlternativeCosts[0].Dash {
		t.Fatalf("alternative costs %+v, want Dash {3}{B}", spec.AlternativeCosts)
	}
}

// --- Nefarious Imp -------------------------------------------------------

func TestGapB213ImpScriesOncePerBatchOfDepartures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Nefarious Imp", "Creature — Imp", gapB213ImpOracle, 2, 1)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.B18Kill(t, g, theirs)
	if effectstest.ScryChoiceFor(g, me.ID) != nil {
		t.Fatal("an opponent's permanent leaving made the Imp's controller scry")
	}
	effectstest.B18Kill(t, g, mine)
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Error("a permanent I controlled leaving did not scry 1")
	}
}

// --- Guildsworn Prowler --------------------------------------------------

func TestGapB213ProwlerDrawsWhenItDiesUnlessBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	prowler := effectstest.B12Push(g, me.ID, "Guildsworn Prowler", "Creature — Tiefling Rogue Assassin", gapB213ProwlerOracle, 2, 1)
	hand := me.Hand.Size()
	effectstest.B18Kill(t, g, prowler)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want +1 when it died unblocking", hand, got)
	}
}

func TestGapB213ProwlerDrawsNothingWhenItDiesBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Brute", "Creature — Ogre", 4, 4)
	prowler := effectstest.B12Push(g, opp.ID, "Guildsworn Prowler", "Creature — Tiefling Rogue Assassin", gapB213ProwlerOracle, 2, 1)
	effectstest.Pr7bAttack(t, g, opp.ID, attacker)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{prowler: attacker})
	hand := opp.Hand.Size()
	effectstest.Pr7bDamageStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, prowler) != game.ZoneGraveyard {
		t.Fatalf("the Prowler did not die (zone %s)", effectstest.B12ZoneOf(g, prowler))
	}
	if got := opp.Hand.Size(); got != hand {
		t.Errorf("hand %d -> %d: it drew although it was blocking", hand, got)
	}
}

// --- Nothic --------------------------------------------------------------

func TestGapB213NothicRollsAD20AndDrawsAndLosesLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	nothic := effectstest.B12Push(g, me.ID, "Nothic", "Creature — Horror", gapB213NothicOracle, 4, 3)
	hand, life := me.Hand.Size(), me.Life
	effectstest.B18Kill(t, g, nothic)
	drew := me.Hand.Size() - hand
	lost := life - me.Life
	if drew != lost || (drew != 1 && drew != 2 && drew != 7) {
		t.Errorf("drew %d and lost %d life, want equal amounts of 1, 2 or 7", drew, lost)
	}
}

// --- Hezrou // Demonic Stench -------------------------------------------

func TestGapB213HezrouShrinksEveryBlockerWhenYourCreatureBecomesBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Hezrou", "Creature — Frog Demon", gapB213HezrouOracle, 6, 6)
	attacker := effectstest.B12Creature(g, me.ID, "Brute", "Creature — Ogre", 3, 3)
	blocker := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.Pr7bAttack(t, g, opp.ID, attacker)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{blocker: attacker})
	if p, tgh := effectstest.EffectivePower(t, g, blocker), effectstest.EffectiveToughness(t, g, blocker); p != -1 || tgh != 3 {
		t.Errorf("the blocker is %d/%d, want -1/3", p, tgh)
	}
	if p := effectstest.EffectivePower(t, g, attacker); p != 3 {
		t.Errorf("the attacker is %d power, want 3: only blocking creatures shrink", p)
	}
}

func TestGapB213HezrouIgnoresAnOpponentsCreatureBecomingBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, opp.ID, "Hezrou", "Creature — Frog Demon", gapB213HezrouOracle, 6, 6)
	attacker := effectstest.B12Creature(g, me.ID, "Brute", "Creature — Ogre", 3, 3)
	blocker := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.Pr7bAttack(t, g, opp.ID, attacker)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{blocker: attacker})
	if tgh := effectstest.EffectiveToughness(t, g, blocker); tgh != 4 {
		t.Errorf("the blocker's toughness is %d, want 4: the attacker is not Hezrou's controller's", tgh)
	}
}

func TestGapB213DemonicStenchShrinksEveryCreatureThatBlockedThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Brute", "Creature — Ogre", 2, 3)
	blocker := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", "Creature — Bear", 2, 2)
	effectstest.Pr7bAttack(t, g, opp.ID, attacker)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{blocker: attacker})
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	spell := effectstest.B18CastFromHand(t, g, me, "Demonic Stench", "Instant — Adventure", gapB213StenchOracle, game.CastSpellParams{})
	_ = spell
	effectstest.PassPriorityAroundTable(t, g)
	if tgh := effectstest.EffectiveToughness(t, g, blocker); tgh != 3 {
		t.Errorf("the creature that blocked has toughness %d, want 3", tgh)
	}
	if tgh := effectstest.EffectiveToughness(t, g, bystander); tgh != 2 {
		t.Errorf("a creature that did not block has toughness %d, want 2", tgh)
	}
	if tgh := effectstest.EffectiveToughness(t, g, attacker); tgh != 3 {
		t.Errorf("the attacker has toughness %d, want 3 (it did not block)", tgh)
	}
}

// --- Sarevok, Deathbringer ----------------------------------------------

func TestGapB213SarevokDrainsTheActivePlayerWhenNothingLeft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Sarevok, Deathbringer", "Legendary Creature — Human Knight", gapB213SarevokOracle, 3, 4)
	lives := effectstest.B24Lives(g)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	after := effectstest.B24Lives(g)
	active := g.Turn.ActiveSeat
	for i := range after {
		want := lives[i]
		if i == active {
			want -= 3
		}
		if after[i] != want {
			t.Errorf("seat %d life %d, want %d (active seat %d)", i, after[i], want, active)
		}
	}
	_ = opp
}

func TestGapB213SarevokDoesNothingWhenAPermanentLeftThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Sarevok, Deathbringer", "Legendary Creature — Human Knight", gapB213SarevokOracle, 3, 4)
	land := effectstest.B12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	effectstest.B18Kill(t, g, land)
	lives := effectstest.B24Lives(g)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	for i, l := range effectstest.B24Lives(g) {
		if l != lives[i] {
			t.Errorf("seat %d life %d -> %d although a land left the battlefield", i, lives[i], l)
		}
	}
}

// --- Ravenloft Adventurer ------------------------------------------------

func TestGapB213AdventurerExilesAnOpponentsDyingCreatureWithAHitCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Ravenloft Adventurer", "Creature — Human Rogue Assassin", gapB213AdventurerOracl, 3, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.B18Kill(t, g, theirs)
	if z := effectstest.B12ZoneOf(g, theirs); z != game.ZoneExile {
		t.Fatalf("their creature is in %s, want exile", z)
	}
	if n := effectstest.CounterCountInExile(g, theirs, "hit"); n != 1 {
		t.Errorf("hit counters on it: %d, want 1", n)
	}
	effectstest.B18Kill(t, g, mine)
	if z := effectstest.B12ZoneOf(g, mine); z != game.ZoneGraveyard {
		t.Errorf("my own creature is in %s, want the graveyard", z)
	}
}

func TestGapB213AdventurerTakesTheInitiativeWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B13PlayAs(t, g, 0, "Ravenloft Adventurer", "Creature — Human Rogue Assassin", gapB213AdventurerOracl)
	effectstest.PassPriorityAroundTable(t, g)
	g.ReadSnapshot(func() {
		if !effects.YouHaveTheInitiative(g, me.ID) {
			t.Error("the controller did not take the initiative")
		}
	})
}

func TestGapB213AdventurerAttackDrainsOnlyAfterACompletedDungeon(t *testing.T) {
	for _, completed := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		adv := effectstest.B12Push(g, me.ID, "Ravenloft Adventurer", "Creature — Human Rogue Assassin", gapB213AdventurerOracl, 3, 4)
		for i := 0; i < 2; i++ {
			c := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 1, 1)
			effectstest.B18Kill(t, g, c)
		}
		if completed {
			me.CompletedDungeons = append(me.CompletedDungeons, "lost-mine-of-phandelver")
		}
		life := opp.Life
		effectstest.DeclareAttack(t, g, opp.ID, adv)
		effectstest.PassPriorityAroundTable(t, g)
		want := life
		if completed {
			want = life - 2
		}
		if opp.Life != want {
			t.Errorf("completed=%v: opponent life %d -> %d, want %d", completed, life, opp.Life, want)
		}
	}
}

// --- Intellect Devourer --------------------------------------------------

func TestGapB213DevourerExilesACardFromEachOpponentsHandUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	card := effectstest.B13HandCard(opp, "Lightning Helix", "Instant", "")
	hand := opp.Hand.Size()
	dev := effectstest.B13PlayAs(t, g, 0, "Intellect Devourer", "Creature — Horror", gapB213DevourerOracle)
	effectstest.PassPriorityAroundTable(t, g)
	asked := 0
	for _, seat := range g.Seats[1:] {
		pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, seat.ID)
		if pick == nil {
			continue
		}
		asked++
		ids := pick.ChooseCards[:1]
		if seat.ID == opp.ID {
			ids = []uuid.UUID{card}
		}
		if err := g.ResolveChooseCards(pick.ID, seat.ID, ids); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
	}
	if asked != len(g.Seats)-1 {
		t.Fatalf("%d opponents were asked which card to exile, want %d", asked, len(g.Seats)-1)
	}
	if z := effectstest.B12ZoneOf(g, card); z != game.ZoneExile {
		t.Fatalf("the chosen card is in %s, want exile", z)
	}
	if opp.Hand.Size() != hand-1 {
		t.Errorf("opponent's hand %d -> %d, want one fewer", hand, opp.Hand.Size())
	}
	effectstest.LeaveBattlefield(t, g, me.ID, dev)
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.B12ZoneOf(g, card); z != game.ZoneHand {
		t.Errorf("after the Devourer left the card is in %s, want its owner's hand", z)
	}
	_ = me
}

func TestGapB213DevourerOffersAnyColorPermissionOverItsExiledCards(t *testing.T) {
	spec, ok := effects.Lookup(gapB213DevourerOracle)
	if !ok || len(spec.CastPermissions) != 1 {
		t.Fatalf("permissions %+v", spec.CastPermissions)
	}
	p := spec.CastPermissions[0]
	if p.Zone != game.ZoneExile || !p.Filter.ExiledWithSource || !p.AnyColor {
		t.Errorf("permission %+v: want exile, exiled-with-source, any color", p)
	}
}

// --- Sivriss, Nightmare Speaker -----------------------------------------

func TestGapB213SivrissMillsAndReturnsTheCardUnlessThePlayerPays(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	siv := effectstest.B12Push(g, me.ID, "Sivriss, Nightmare Speaker", "Legendary Creature — Snake Cleric Warlock", gapB213SivrissOracle, 3, 3)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)
	top, err := me.Library.Top()
	if err != nil {
		t.Skip("no library")
	}
	hand := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, siv, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}})
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	if z := effectstest.B12ZoneOf(g, top.InstanceID); z != game.ZoneHand {
		t.Errorf("the milled card is in %s, want my hand after they declined", z)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, me.Hand.Size())
	}
}

func TestGapB213SivrissLeavesTheCardInTheGraveyardWhenThePlayerPaysThreeLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	siv := effectstest.B12Push(g, me.ID, "Sivriss, Nightmare Speaker", "Legendary Creature — Snake Cleric Warlock", gapB213SivrissOracle, 3, 3)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)
	hand := me.Hand.Size()
	life := make([]int, len(g.Seats))
	for i, s := range g.Seats {
		life[i] = s.Life
	}
	effectstest.B16Activate(t, g, me.ID, siv, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}})
	// The first opponent pays; the other two decline and their cards return.
	effectstest.AnswerPayUnless(t, g, g.Seats[1].ID, true)
	effectstest.AnswerPayUnless(t, g, g.Seats[2].ID, false)
	effectstest.AnswerPayUnless(t, g, g.Seats[3].ID, false)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d -> %d, want +2 (two declined, one paid)", hand, got)
	}
	if g.Seats[1].Life != life[1]-3 {
		t.Errorf("the paying opponent's life %d -> %d, want -3", life[1], g.Seats[1].Life)
	}
	for _, i := range []int{2, 3} {
		if g.Seats[i].Life != life[i] {
			t.Errorf("seat %d life changed", i)
		}
	}
}
