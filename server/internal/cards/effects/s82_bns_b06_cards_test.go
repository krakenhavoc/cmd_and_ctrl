package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	bnsB06HelixPinnacleOracle = "9ddc5a06-b4e1-459a-9e39-7cd1aa994fe1"
	bnsB06RootMazeOracle      = "2f4abc5e-5aad-455e-b1ec-137d7ac4a10c"
	bnsB06ParadoxHazeOracle   = "daff0e7b-08f2-435f-97b9-9b000e94ffa8"
	bnsB06SplinterTwinOracle  = "2df3187b-4b91-4de9-9b26-d48a30700cec"
	bnsB06CollectedCoOracle   = "0bbcb563-09b3-499d-a9cc-3a6d7501dd81"
	bnsB06CondemnOracle       = "377d31b5-0a48-4d5b-91bd-b5a62d4f7899"
	bnsB06ConsignOracle       = "16c39265-72d9-4c86-a961-8dd38ff9adcd"
	bnsB06GalvanicBlastOracle = "d211c03a-03cd-40c5-b7c8-ada352fd01a6"
	bnsB06ManamorphoseOracle  = "89c83a6d-f6c8-4984-a888-0db62dfb93b1"
	bnsB06WarpingWailOracle   = "c8b9fa02-0a0c-4a87-9a00-3c1c7e2f4bfd"
	bnsB06InstantType         = "Instant"
	bnsB06LightningBoltOracle = "4457ed35-7c10-48c8-9776-456485fdf070"
	bnsB06EnchantmentAuraType = "Enchantment — Aura"
)

func TestBnsB06EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		bnsB06HelixPinnacleOracle: "Helix Pinnacle",
		bnsB06RootMazeOracle:      "Root Maze",
		bnsB06ParadoxHazeOracle:   "Paradox Haze",
		bnsB06SplinterTwinOracle:  "Splinter Twin",
		bnsB06CollectedCoOracle:   "Collected Company",
		bnsB06CondemnOracle:       "Condemn",
		bnsB06ConsignOracle:       "Consign to Memory",
		bnsB06GalvanicBlastOracle: "Galvanic Blast",
		bnsB06ManamorphoseOracle:  "Manamorphose",
		bnsB06WarpingWailOracle:   "Warping Wail",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestBnsB06GalvanicBlastIsTwoUnlessMetalcraft(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := opp.Life
	castCatalogSpell(t, g, "Galvanic Blast", bnsB06InstantType, bnsB06GalvanicBlastOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Fatalf("without metalcraft: life %d → %d, want 2 damage", life, opp.Life)
	}
	for i := 0; i < 2; i++ {
		b12Permanent(g, me.ID, "Rock", "Artifact")
	}
	life = opp.Life
	castCatalogSpell(t, g, "Galvanic Blast", bnsB06InstantType, bnsB06GalvanicBlastOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Fatalf("two artifacts: life %d → %d, want 2 damage", life, opp.Life)
	}
	b12Permanent(g, me.ID, "Rock", "Artifact")
	life = opp.Life
	castCatalogSpell(t, g, "Galvanic Blast", bnsB06InstantType, bnsB06GalvanicBlastOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Fatalf("three artifacts: life %d → %d, want 4 damage", life, opp.Life)
	}
}

func TestBnsB06RootMazeTapsEveryArtifactAndLandButNotACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Root Maze", "Enchantment", bnsB06RootMazeOracle, 0, 0)
	advanceToMain(t, g)
	mineLand := b12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	if c, ok := battlefieldCard(g, mineLand); !ok || !c.Tapped {
		t.Errorf("a land you play enters tapped: %+v", c)
	}
	rock := avB1Hand(me, game.Card{Name: "Mind Stone", TypeLine: "Artifact", ManaCost: "{2}"})
	floatMana(t, g, me, "{C}{C}")
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the artifact: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, rock); !ok || !c.Tapped {
		t.Errorf("an artifact enters tapped: %+v", c)
	}
	bear := avB1Hand(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, ManaCost: "{1}{G}"})
	floatMana(t, g, me, "{G}{G}")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, bear); !ok || c.Tapped {
		t.Errorf("a plain creature enters untapped: %+v", c)
	}
}

func TestBnsB06ManamorphoseAddsTwoManaOfAnyColorsAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	spell := avB1Hand(me, game.Card{Name: "Manamorphose", TypeLine: bnsB06InstantType, OracleID: bnsB06ManamorphoseOracle})
	hand := me.Hand.Size()
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	for i := 0; i < 64; i++ {
		if pick := riderLatestManaPick(g, me.ID); pick != nil {
			if err := g.ResolveManaChoice(pick.ID, me.ID, "W"); err != nil {
				t.Fatalf("ResolveManaChoice: %v", err)
			}
			continue
		}
		if stackFullyEmpty(g) {
			break
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d → %d, want the card cast (-1) and one drawn (+1)", hand, me.Hand.Size())
	}
	if me.ManaPool.Total() != 2 {
		t.Errorf("pool %d, want two mana", me.ManaPool.Total())
	}
}

func TestBnsB06CondemnBottomsAnAttackerAndGivesItsControllerLifeEqualToToughness(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	attacker := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 4)
	idle := b12Creature(g, opp.ID, "Idle Ogre", "Creature — Ogre", 3, 4)
	g.WithWriteLock(func() {
		findBattlefieldCardForTest(g, attacker).AttackingTarget = g.Seats[0].ID
	})
	if err := castCatalogSpellErrStrict(t, g, "Condemn", bnsB06InstantType, bnsB06CondemnOracle, b16TargetCard(idle)); err == nil {
		t.Fatal("Condemn targeted a creature that is not attacking")
	}
	life := opp.Life
	libSize := opp.Library.Size()
	castCatalogSpell(t, g, "Condemn", bnsB06InstantType, bnsB06CondemnOracle, b16TargetCard(attacker))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Fatal("the attacker is gone from the battlefield")
	}
	if opp.Library.Size() != libSize+1 || opp.Library.Cards[0].InstanceID != attacker {
		t.Errorf("the attacker is on the bottom of its owner's library (size %d → %d)", libSize, opp.Library.Size())
	}
	if opp.Life != life+4 {
		t.Errorf("its controller gains life equal to its toughness: %d → %d, want +4", life, opp.Life)
	}
}

func TestBnsB06ConsignToMemoryTakesATriggeredAbilityOrAColorlessSpellOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	src := pushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	trigger := taAnnounce(t, g, opp.ID, src, "Their Rock — trigger", nil, nil)
	g.WithWriteLock(func() { g.StackMeta[trigger].Kind = game.StackItemTriggered })
	activated := taAnnounce(t, g, opp.ID, src, "Their Rock — activated", nil, nil)
	g.WithWriteLock(func() { g.StackMeta[activated].Kind = game.StackItemActivated })
	red := batch01OpponentCasts(t, g, opp, "Lightning Bolt", bnsB06LightningBoltOracle, "{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	legal := taLegalTargets(g, me.ID, bnsB06ConsignOracle)
	if !legal[trigger] {
		t.Error("Consign to Memory takes a triggered ability")
	}
	if legal[activated] {
		t.Error("Consign to Memory must not take an activated ability")
	}
	if legal[red] {
		t.Error("Consign to Memory must not take a coloured spell")
	}
}

func TestBnsB06ConsignToMemoryCountersAColorlessSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	rock := avB1Hand(me, game.Card{Name: "Mind Stone", TypeLine: "Artifact", ManaCost: "{2}"})
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the artifact: %v", err)
	}
	consign := avB1Hand(me, game.Card{Name: "Consign to Memory", TypeLine: bnsB06InstantType, ManaCost: "{U}", OracleID: bnsB06ConsignOracle})
	if err := g.CastSpell(me.ID, consign, game.CastSpellParams{Targets: cardRefs(rock)}); err != nil {
		t.Fatalf("Consign to Memory on the colorless spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !me.Graveyard.Contains(rock) {
		t.Error("the colorless spell is countered into its owner's graveyard")
	}
}

func TestBnsB06ConsignToMemoryReplicatesForOneGeneric(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	rock := avB1Hand(me, game.Card{Name: "Mind Stone", TypeLine: "Artifact", ManaCost: "{2}"})
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the artifact: %v", err)
	}
	consign := avB1Hand(me, game.Card{Name: "Consign to Memory", TypeLine: bnsB06InstantType, ManaCost: "{U}", OracleID: bnsB06ConsignOracle})
	floatMana(t, g, me, "{U}")
	if err := g.CastSpell(me.ID, consign, game.CastSpellParams{Targets: cardRefs(rock), OptionalCosts: []int{0}, Strict: true}); err == nil {
		t.Fatal("cast with one replicate on {U} alone")
	}
	floatMana(t, g, me, "{U}{R}")
	if err := g.CastSpell(me.ID, consign, game.CastSpellParams{Targets: cardRefs(rock), OptionalCosts: []int{0}, Strict: true}); err != nil {
		t.Fatalf("cast with one replicate: %v", err)
	}
	if it := triggerOnStack(g, consign); it == nil || it.XValue != 1 {
		t.Fatalf("replicate trigger = %+v, want one counting one payment", it)
	}
}

func TestBnsB06WarpingWailModes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tiny := b12Creature(g, opp.ID, "Tiny", "Creature — Elf", 1, 5)
	thin := b12Creature(g, opp.ID, "Thin", "Creature — Elf", 5, 1)
	big := b12Creature(g, opp.ID, "Big", "Creature — Ogre", 3, 3)
	if _, err := castModalErr(t, g, "Warping Wail", bnsB06InstantType, bnsB06WarpingWailOracle, []int{0}, cardRefs(big)); err == nil {
		t.Fatal("a 3/3 is neither power nor toughness 1 or less")
	}
	castModal(t, g, "Warping Wail", bnsB06InstantType, bnsB06WarpingWailOracle, []int{0}, cardRefs(tiny))
	passPriorityAroundTable(t, g)
	castModal(t, g, "Warping Wail", bnsB06InstantType, bnsB06WarpingWailOracle, []int{0}, cardRefs(thin))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tiny) || g.Battlefield.Contains(thin) || !g.Battlefield.Contains(big) {
		t.Error("exactly the power-or-toughness-1 creatures are exiled")
	}
	castModal(t, g, "Warping Wail", bnsB06InstantType, bnsB06WarpingWailOracle, []int{2}, nil)
	passPriorityAroundTable(t, g)
	if battlefieldNamed(g, me.ID, "Eldrazi Scion") != 1 {
		t.Error("the third mode makes one Eldrazi Scion")
	}
}

func TestBnsB06WarpingWailCountersASorceryButNotAnInstant(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bolt := batch01OpponentCasts(t, g, opp, "Lightning Bolt", bnsB06LightningBoltOracle, "{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	if _, err := castModalErr(t, g, "Warping Wail", bnsB06InstantType, bnsB06WarpingWailOracle, []int{1}, cardRefs(bolt)); err == nil {
		t.Fatal("an instant is not a sorcery spell")
	}
}

func TestBnsB06WarpingWailCostsColorlessNotGeneric(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	wail := avB1Hand(me, game.Card{Name: "Warping Wail", TypeLine: bnsB06InstantType, ManaCost: "{1}{C}", OracleID: bnsB06WarpingWailOracle})
	floatMana(t, g, me, "{R}{R}")
	if err := g.CastSpell(me.ID, wail, game.CastSpellParams{Modes: []int{2}, Strict: true}); err == nil {
		t.Fatal("{C} cannot be paid with red mana")
	}
	floatMana(t, g, me, "{C}")
	if err := g.CastSpell(me.ID, wail, game.CastSpellParams{Modes: []int{2}, Strict: true}); err != nil {
		t.Fatalf("{R}{R}{C} pays {1}{C}: %v", err)
	}
}

func TestBnsB06SplinterTwinCopiesTheCreatureWithHasteAndExilesTheTokenAtEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	host := b12Creature(g, me.ID, "Host", "Creature — Ogre", 3, 3)
	aura := pushCatalogPermanent(g, me.ID, "Splinter Twin", bnsB06EnchantmentAuraType, bnsB06SplinterTwinOracle, false)
	g.WithWriteLock(func() {
		findBattlefieldCardForTest(g, aura).AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: host}
	})
	advanceToMain(t, g)
	idx, ref := grantedActivatedIndex(t, g, host)
	if idx < 0 {
		t.Fatal("the enchanted creature has the granted ability")
	}
	b16Activate(t, g, me.ID, host, idx, game.ActivateAbilityParams{Ref: ref})
	passPriorityAroundTable(t, g)
	if n := battlefieldNamed(g, me.ID, "Host"); n != 2 {
		t.Fatalf("Host and its token copy: %d on the battlefield, want 2", n)
	}
	if !b16Tapped(t, g, host) {
		t.Error("the ability taps the enchanted creature")
	}
	var token game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Host" && c.InstanceID != host {
			token = c
		}
	}
	if !game.HasKeyword(&token, "haste") && !hasEffectiveKeyword(t, g, token.InstanceID, "haste") {
		t.Error("the token has haste")
	}
	if _, grant := grantedActivatedIndex(t, g, token.InstanceID); grant != "" {
		t.Error("the Aura's granted ability is not copied")
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if n := battlefieldNamed(g, me.ID, "Host"); n != 1 {
		t.Errorf("the token is exiled at the next end step: %d Hosts remain", n)
	}
}

func TestBnsB06ParadoxHazeGivesTheEnchantedPlayerAnExtraUpkeepOnlyOnTheirFirst(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Paradox Haze", TypeLine: bnsB06EnchantmentAuraType, OracleID: bnsB06ParadoxHazeOracle,
		Owner: me.ID, Controller: me.ID,
		AttachedTo: game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID},
	})
	upkeepsAhead := func() int {
		n := 0
		g.WithWriteLock(func() {
			for _, s := range g.UpcomingStepsForEffect() {
				if s.Step == game.StepUpkeep {
					n++
				}
			}
		})
		return n
	}
	// My own upkeep: the Haze watches the enchanted player only.
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if got := upkeepsAhead(); got != 1 {
		t.Fatalf("after the enchanted player's first upkeep, %d upkeep steps ahead, want 1", got)
	}
	// Run the added upkeep: it is their second, so no further step.
	for i := 0; i < 20 && g.Turn.Step != game.StepUpkeep || i == 0; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if got := upkeepsAhead(); got != 0 {
		t.Errorf("the second upkeep must not add another, %d ahead", got)
	}
}

func TestBnsB06HelixPinnacleWinsAtHundredTowerCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pin := pushCatalogPermanent(g, me.ID, "Helix Pinnacle", "Enchantment", bnsB06HelixPinnacleOracle, false)
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, pin, 0, game.ActivateAbilityParams{XValue: 3, Strict: true}); err != nil {
		t.Fatalf("activate with X=3: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, pin, "tower"); got != 3 {
		t.Fatalf("tower counters %d, want 3", got)
	}
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, pin).Counters["tower"] = 99 })
	advanceToUpkeepOf(t, g, 1)
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if g.State != game.StateActive {
		t.Fatalf("99 counters must not win")
	}
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, pin).Counters["tower"] = 100 })
	advanceToUpkeepOf(t, g, 1)
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if g.State == game.StateActive {
		t.Errorf("100 tower counters at your upkeep: you win the game (state %v)", g.State)
	}
}

func TestBnsB06CollectedCompanyPutsTwoSmallCreaturesOntoTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	me.Library.Cards = nil
	var ids []uuid.UUID
	for _, c := range []game.Card{
		{Name: "Big", TypeLine: "Creature — Ogre", ManaCost: "{4}{G}", Power: 5, Toughness: 5},
		{Name: "Elf B", TypeLine: "Creature — Elf", ManaCost: "{1}{G}", Power: 1, Toughness: 1},
		{Name: "Elf C", TypeLine: "Creature — Elf", ManaCost: "{2}", Power: 1, Toughness: 1},
		{Name: "Elf D", TypeLine: "Creature — Elf", ManaCost: "{G}", Power: 1, Toughness: 1},
		{Name: "Land", TypeLine: "Basic Land — Forest"},
		{Name: "Spell", TypeLine: "Sorcery", ManaCost: "{1}"},
		{Name: "Deep", TypeLine: "Creature — Elf", ManaCost: "{1}", Power: 1, Toughness: 1},
	} {
		c.InstanceID = uuid.New()
		c.Owner, c.Controller = me.ID, me.ID
		ids = append(ids, c.InstanceID)
		g.WithWriteLock(func() { me.Library.PushBottom(c) })
	}
	castCatalogSpell(t, g, "Collected Company", bnsB06InstantType, bnsB06CollectedCoOracle, nil)
	passPriorityAroundTable(t, g)
	pick := latestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("no creature pick was offered")
	}
	for _, id := range pick.ChooseCards {
		if id == ids[0] || id == ids[4] || id == ids[5] {
			t.Errorf("only creature cards with mana value 3 or less may be offered, got %v", id)
		}
	}
	if len(pick.ChooseCards) != 3 {
		t.Errorf("three eligible among the top six, got %d", len(pick.ChooseCards))
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[1], ids[2], ids[3]}); err == nil {
		t.Error("three creatures were accepted for up to two")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[1], ids[2]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ids[1]) || !g.Battlefield.Contains(ids[2]) {
		t.Error("both chosen creatures enter the battlefield")
	}
	if g.Battlefield.Contains(ids[3]) || g.Battlefield.Contains(ids[6]) {
		t.Error("nothing else enters")
	}
}
