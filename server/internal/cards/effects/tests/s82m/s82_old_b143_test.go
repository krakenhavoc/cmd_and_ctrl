package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b143_test.go — Throne of Eldraine, S82 older-sets batch 143
// (#5516): damage. Each test plays the real catalog entry.

const (
	oldB143MeleeOracle      = "1b78aa71-b554-45dc-a737-52cd74c3e7f1"
	oldB143JoustOracle      = "8584e96f-a3e7-475c-bfd1-2c9a72bee774"
	oldB143StrokeOracle     = "9c348bc7-e6e1-40a2-9798-6d2c4a2d8e72"
	oldB143CarahOracle      = "c3da0550-b816-4aed-8174-ddde199b6589"
	oldB143TrebuchetOracle  = "d8dd601f-3615-4cef-b837-52329cb361f8"
	oldB143QueenOracle      = "f5511802-56fe-4922-83d2-ae723e1beab7"
	oldB143HeadhunterOracle = "fad0118b-d046-4ee7-87d9-e6799f70a7a8"
)

func TestOldB143EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB143MeleeOracle: "Redcap Melee", oldB143JoustOracle: "Joust", oldB143StrokeOracle: "Sundering Stroke",
		oldB143CarahOracle: "Syr Carah, the Bold", oldB143TrebuchetOracle: "Brimstone Trebuchet",
		oldB143QueenOracle: "Queen of Ice", oldB143QueenOracle + "#1": "Rage of Winter",
		oldB143HeadhunterOracle: "Elite Headhunter",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Redcap Melee ---------------------------------------------------------

func oldB143Melee(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID) {
	t.Helper()
	effectstest.ToMainForCost(t, g)
	effectstest.B10AddMana(me, "R")
	id := effectstest.HandCardFull(me, "Redcap Melee", "Instant", "{R}", oldB143MeleeOracle, []string{"R"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(target)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB143RedcapMeleeMakesYouSacrificeALandOnlyForANonredPermanent(t *testing.T) {
	for _, red := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		colors := []string{"G"}
		if red {
			colors = []string{"R"}
		}
		victim := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Victim", TypeLine: "Creature — Ogre",
			Colors: colors, Power: 3, Toughness: 4, Owner: opp.ID, Controller: opp.ID})
		land := effectstest.BlbB1Push(g, me.ID, "Mountain", "Basic Land — Mountain", "", 0, 0)
		oldB143Melee(t, g, me, victim)
		if !red {
			effectstest.AnswerSacrifice(t, g, me.ID, land)
			effectstest.PassPriorityAroundTable(t, g)
		}
		if g.Battlefield.Contains(victim) {
			t.Errorf("red=%v: the 3/4 survived 4 damage", red)
		}
		if got := g.Battlefield.Contains(land); got != red {
			t.Errorf("red=%v: land still on the battlefield = %v, want %v", red, got, red)
		}
	}
}

func TestOldB143RedcapMeleeRefusesAPlayerTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	effectstest.B10AddMana(me, "R")
	id := effectstest.HandCardFull(me, "Redcap Melee", "Instant", "{R}", oldB143MeleeOracle, []string{"R"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err == nil {
		t.Error("Redcap Melee targeted a player")
	}
}

// --- Joust ----------------------------------------------------------------

func TestOldB143JoustPumpsOnlyAKnightBeforeTheFight(t *testing.T) {
	for _, knight := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		tl := "Creature — Bear"
		if knight {
			tl = "Creature — Human Knight"
		}
		mine := effectstest.B12Creature(g, me.ID, "Mine", tl, 2, 2)
		theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Ogre", 2, 3)
		effectstest.ToMainForCost(t, g)
		effectstest.B10AddMana(me, "R", "R")
		id := effectstest.HandCardFull(me, "Joust", "Sorcery", "{1}{R}", oldB143JoustOracle, []string{"R"})
		targets := []game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}}
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: targets}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := g.Battlefield.Contains(theirs); got == knight {
			t.Errorf("knight=%v: their 2/3 on the battlefield = %v (a 4/3 Knight kills it, a 2/2 does not)", knight, got)
		}
		if got := g.Battlefield.Contains(mine); got != knight {
			t.Errorf("knight=%v: my creature on the battlefield = %v (a 4/3 Knight survives 2, a 2/2 dies)", knight, got)
		}
	}
}

func TestOldB143JoustRefusesTwoOfMyCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	effectstest.ToMainForCost(t, g)
	effectstest.B10AddMana(me, "R", "R")
	id := effectstest.HandCardFull(me, "Joust", "Sorcery", "{1}{R}", oldB143JoustOracle, []string{"R"})
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: targets}); err == nil {
		t.Error("Joust took a second creature I control as 'target creature you don't control'")
	}
}

// --- Sundering Stroke -----------------------------------------------------

func oldB143Stroke(t *testing.T, g *game.Game, me *game.Player, red int, params game.CastSpellParams) error {
	t.Helper()
	effectstest.ToMainForCost(t, g)
	for i := 0; i < red; i++ {
		effectstest.B10AddMana(me, "R")
	}
	for i := red; i < 7; i++ {
		effectstest.B10AddMana(me, "G")
	}
	id := effectstest.HandCardFull(me, "Sundering Stroke", "Sorcery", "{6}{R}", oldB143StrokeOracle, []string{"R"})
	params.Strict = true
	return g.CastSpell(me.ID, id, params)
}

func TestOldB143SunderingStrokeDividesSevenDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Ogre", 1, 4)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Ogre", 1, 5)
	life := opp.Life
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}, effectstest.PlayerTarget(opp.ID)}
	if err := oldB143Stroke(t, g, me, 1, game.CastSpellParams{Targets: targets, Distribution: map[uuid.UUID]int{a: 4, b: 2, opp.ID: 1}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) {
		t.Error("the 4-toughness creature survived its 4")
	}
	if !g.Battlefield.Contains(b) {
		t.Error("the 5-toughness creature died to 2")
	}
	if opp.Life != life-1 {
		t.Errorf("life %d, want %d", opp.Life, life-1)
	}
}

func TestOldB143SunderingStrokeWithSevenRedDealsSevenToEach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Ogre", 1, 7)
	life := opp.Life
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: a}, effectstest.PlayerTarget(opp.ID)}
	if err := oldB143Stroke(t, g, me, 7, game.CastSpellParams{Targets: targets, Distribution: map[uuid.UUID]int{a: 4, opp.ID: 3}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) {
		t.Error("the 7-toughness creature survived 7")
	}
	if opp.Life != life-7 {
		t.Errorf("life %d, want %d", opp.Life, life-7)
	}
}

func TestOldB143SunderingStrokeWithSixRedIsStillDivided(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Ogre", 1, 7)
	life := opp.Life
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: a}, effectstest.PlayerTarget(opp.ID)}
	if err := oldB143Stroke(t, g, me, 6, game.CastSpellParams{Targets: targets, Distribution: map[uuid.UUID]int{a: 4, opp.ID: 3}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(a) || opp.Life != life-3 {
		t.Errorf("six red mana: creature alive=%v life %d, want alive and %d", g.Battlefield.Contains(a), opp.Life, life-3)
	}
}

func TestOldB143SunderingStrokeRefusesAWrongDivisionAndFourTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Ogre", 1, 7)
	if err := oldB143Stroke(t, g, me, 1, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}, effectstest.PlayerTarget(opp.ID)},
		Distribution: map[uuid.UUID]int{a: 4, opp.ID: 2}}); err == nil {
		t.Error("a division summing to 6 was accepted")
	}
	three := g.Seats[2]
	four := []game.TargetRef{{Kind: game.TargetCard, ID: a}, effectstest.PlayerTarget(opp.ID), effectstest.PlayerTarget(three.ID), effectstest.PlayerTarget(me.ID)}
	if err := oldB143Stroke(t, g, me, 1, game.CastSpellParams{Targets: four,
		Distribution: map[uuid.UUID]int{a: 3, opp.ID: 2, three.ID: 1, me.ID: 1}}); err == nil {
		t.Error("four targets were accepted")
	}
}

// --- Syr Carah, the Bold --------------------------------------------------

func oldB143Carah(g *game.Game, me *game.Player) uuid.UUID {
	effectstest.SeedSearchLibrary(me, game.Card{Name: "Top", TypeLine: "Basic Land — Forest"}, game.Card{Name: "Next", TypeLine: "Basic Land — Forest"})
	return effectstest.B12Push(g, me.ID, "Syr Carah, the Bold", "Legendary Creature — Human Knight", oldB143CarahOracle, 3, 3)
}

func TestOldB143SyrCarahPingExilesTheTopCardOnlyWhenItHitsAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	carah := oldB143Carah(g, me)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	before := len(g.Exile.Cards)
	if err := g.ActivateCatalogAbility(me.ID, carah, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("ping the bear: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.Exile.Cards) != before {
		t.Fatal("damage to a creature exiled a card")
	}
	effectstest.UntapForTest(g, carah)
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, carah, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("ping the player: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Fatalf("life %d, want %d", opp.Life, life-1)
	}
	if len(g.Exile.Cards) != before+1 {
		t.Fatalf("exile has %d cards, want %d", len(g.Exile.Cards), before+1)
	}
	if perm := g.CastPermissionOnCardByIDForEffect(g.Exile.Cards[len(g.Exile.Cards)-1].InstanceID); perm == nil {
		t.Error("no permission to play the exiled card")
	}
}

func TestOldB143SyrCarahTriggersOnYourSpellsDamageToAPlayerNotAnOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB143Carah(g, me)
	before := len(g.Exile.Cards)
	if err := oldB143Stroke(t, g, me, 1, game.CastSpellParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}, Distribution: map[uuid.UUID]int{opp.ID: 7}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.Exile.Cards) != before+1 {
		t.Fatalf("your sorcery's damage to a player: exile has %d, want %d", len(g.Exile.Cards), before+1)
	}
}

func TestOldB143SyrCarahTriggersOnCombatDamageToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	carah := oldB143Carah(g, me)
	before := len(g.Exile.Cards)
	effectstest.DealCombatDamageToPlayer(g, carah, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.Exile.Cards) != before+1 {
		t.Errorf("exile has %d, want %d", len(g.Exile.Cards), before+1)
	}
}

func TestOldB143SyrCarahIgnoresAnotherCreaturesDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB143Carah(g, me)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	before := len(g.Exile.Cards)
	effectstest.DealCombatDamageToPlayer(g, other, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.Exile.Cards) != before {
		t.Error("another creature's damage triggered Syr Carah")
	}
}

// --- Brimstone Trebuchet --------------------------------------------------

func TestOldB143BrimstoneTrebuchetPingsEachOpponentAndUntapsForAKnight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	treb := effectstest.B12Push(g, me.ID, "Brimstone Trebuchet", "Artifact Creature — Wall", oldB143TrebuchetOracle, 1, 3)
	for _, kw := range []string{"defender", "reach"} {
		if !effectstest.HasEffectiveKeyword(t, g, treb, kw) {
			t.Errorf("no %s", kw)
		}
	}
	effectstest.AdvanceToMain(t, g)
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats[1:] {
		lives[s.ID] = s.Life
	}
	if err := g.ActivateCatalogAbility(me.ID, treb, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, s := range g.Seats[1:] {
		if s.Life != lives[s.ID]-1 {
			t.Errorf("%s life %d, want %d", s.ID, s.Life, lives[s.ID]-1)
		}
	}
	if !effectstest.IsTapped(g, treb) {
		t.Fatal("the Trebuchet did not tap")
	}
	if err := g.ActivateCatalogAbility(me.ID, treb, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("activated while tapped")
	}
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.IsTapped(g, treb) {
		t.Error("a non-Knight untapped the Trebuchet")
	}
	effectstest.EnterCard(t, g, g.Seats[1].ID, game.Card{Name: "Their Knight", TypeLine: "Creature — Human Knight", Power: 2, Toughness: 2})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.IsTapped(g, treb) {
		t.Error("an opponent's Knight untapped the Trebuchet")
	}
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Knight", TypeLine: "Creature — Human Knight", Power: 2, Toughness: 2})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IsTapped(g, treb) {
		t.Error("a Knight entering under your control did not untap the Trebuchet")
	}
}

// --- Queen of Ice // Rage of Winter ---------------------------------------

func TestOldB143QueenOfIceTapsAndFreezesACreatureItDamaged(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	queen := effectstest.B12Push(g, me.ID, "Queen of Ice", "Creature — Human Noble Wizard", oldB143QueenOracle, 2, 3)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Ogre", 1, 5)
	effectstest.DealCombatDamage(g, me.ID, queen, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardByID(g, victim); c == nil || !c.Tapped {
		t.Fatal("the damaged creature was not tapped")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceTo(t, g, game.StepUpkeep)
	if c := effectstest.FindBattlefieldCardByID(g, victim); c == nil || !c.Tapped {
		t.Error("the damaged creature untapped in its controller's next untap step")
	}
}

func TestOldB143QueenOfIceIgnoresDamageToAPlayerAndNoncombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	queen := effectstest.B12Push(g, me.ID, "Queen of Ice", "Creature — Human Noble Wizard", oldB143QueenOracle, 2, 3)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Ogre", 1, 5)
	effectstest.DealNoncombatDamage(g, me.ID, queen, victim)
	effectstest.DealCombatDamageToPlayer(g, queen, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardByID(g, victim); c == nil || c.Tapped {
		t.Error("noncombat damage, or damage to a player, tapped a creature")
	}
}

func TestOldB143RageOfWinterTapsAndFreezesAndGoesOnAnAdventure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Ogre", 3, 3)
	me, id := effectstest.HobA6HandAdventure(t, g, oldB143QueenOracle,
		effectstest.HobA6Face("Queen of Ice", "Creature — Human Noble Wizard", "{2}{U}", 2, 3, "U"),
		effectstest.HobA6Face("Rage of Winter", "Sorcery — Adventure", "{1}{U}", 0, 0, "U"))
	effectstest.ToMainForCost(t, g)
	effectstest.B10AddMana(me, "U", "U")
	effectstest.HobA6MustCast(t, g, me, id, 1, effectstest.CardTarget(victim))
	if c := effectstest.FindBattlefieldCardByID(g, victim); c == nil || !c.Tapped {
		t.Fatal("the target was not tapped")
	}
	effectstest.HobA6WentOnAnAdventure(t, g, id)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceTo(t, g, game.StepUpkeep)
	if c := effectstest.FindBattlefieldCardByID(g, victim); c == nil || !c.Tapped {
		t.Error("the target untapped in its controller's next untap step")
	}
}

// --- Elite Headhunter -----------------------------------------------------

func TestOldB143EliteHeadhunterSacrificesAnotherCreatureOrArtifactToDealTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hunter := effectstest.B12Push(g, me.ID, "Elite Headhunter", "Creature — Human Knight", oldB143HeadhunterOracle, 2, 3)
	if !effectstest.HasEffectiveKeyword(t, g, hunter, "menace") {
		t.Error("no menace")
	}
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	land := effectstest.B12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Ogre", 2, 2)
	effectstest.AdvanceToMain(t, g)
	activate := func(sac uuid.UUID) error {
		effectstest.B10AddMana(me, "B", "R", "B")
		return g.ActivateCatalogAbility(me.ID, hunter, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(victim), SacrificeIDs: []uuid.UUID{sac}})
	}
	for _, bad := range []uuid.UUID{hunter, land} {
		if err := activate(bad); err == nil {
			t.Errorf("sacrificed %s", bad)
		}
	}
	if err := activate(rock); err != nil {
		t.Fatalf("sacrifice an artifact: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || g.Battlefield.Contains(victim) {
		t.Error("the artifact stayed or the 2/2 survived 2 damage")
	}
	if !g.Battlefield.Contains(fodder) {
		t.Error("the creature was sacrificed instead")
	}
}
