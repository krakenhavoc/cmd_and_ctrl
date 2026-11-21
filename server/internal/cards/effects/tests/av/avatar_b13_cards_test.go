package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b13_cards_test.go — Avatar: The Last Airbender batch 13 (#4459).

const (
	avB13SozinsCometOracle   = "bc0da02c-2b85-4ddc-9ce0-63934dd14164"
	avB13TigerDilloOracle    = "b6e82d4f-4fbf-495c-bebd-982d2a855ad0"
	avB13TreetopOracle       = "922176b8-7174-4248-92fb-b196ceb802c7"
	avB13TwinBladesOracle    = "88b076a4-46ff-4a41-95a5-2821f4d8944d"
	avB13TyLeeOracle         = "3726b9a0-67cb-4f45-ab50-62db50f93568"
	avB13WarBalloonOracle    = "021b495d-236f-4274-988d-2dfd7103c9ef"
	avB13YuyanArchersOracle  = "63e5180f-16d0-4524-94bf-ecbf3c417814"
	avB13ZukoOracle          = "3231ce2e-d4f8-44fc-872d-aeafa59aed4b"
	avB13AlliesAtLastOracle  = "c15e5e93-2eff-407d-ae75-437e36ef85ba"
	avB13BadgermoleOracle    = "2084a7f6-226b-4fb4-a30e-703f674d33b4"
	avB13TwinBladesTypeLine  = "Artifact — Equipment"
	avB13WarBalloonTypeLine  = "Artifact — Vehicle"
	avB13TyLeeTypeLine       = "Legendary Creature — Human Performer"
	avB13ZukoTypeLine        = "Legendary Creature — Human Noble"
	avB13BadgermoleTypeLine  = "Creature — Badger Mole"
	avB13TreetopTypeLine     = "Creature — Human Rebel Ally"
	avB13YuyanArchersTypeLne = "Creature — Human Archer"
)

func TestAvatarB13CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		avB13SozinsCometOracle:  "Sozin's Comet",
		avB13TigerDilloOracle:   "Tiger-Dillo",
		avB13TreetopOracle:      "Treetop Freedom Fighters",
		avB13TwinBladesOracle:   "Twin Blades",
		avB13TyLeeOracle:        "Ty Lee, Artful Acrobat",
		avB13WarBalloonOracle:   "War Balloon",
		avB13YuyanArchersOracle: "Yuyan Archers",
		avB13ZukoOracle:         "Zuko, Exiled Prince",
		avB13AlliesAtLastOracle: "Allies at Last",
		avB13BadgermoleOracle:   "Badgermole",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("%s registered as %q, want %q", oracle, spec.Name, name)
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s is not Full without caveats", name)
		}
	}
}

// avB13Perm seeds a permanent (catalog or not) old enough to attack and
// tap.
func avB13Perm(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

func avB13Seats(g *game.Game) (me, opp *game.Player) {
	seat := g.Turn.ActiveSeat
	return g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
}

// --- Sozin's Comet ---------------------------------------------------------------

func TestAvatarB13SozinsCometFirebendsOnlyTheCreaturesYouControlledAsItResolved(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := avB13Seats(g)
	early := avB13Perm(g, me.ID, "Early Bear", "Creature — Bear", "", 2, 2)
	theirs := avB13Perm(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "R", 5)
	if _, err := effectstest.AvB13Cast(t, g, game.Card{Name: "Sozin's Comet", TypeLine: "Sorcery", ManaCost: "{3}{R}{R}",
		OracleID: avB13SozinsCometOracle}, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast Sozin's Comet: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	late := avB13Perm(g, me.ID, "Late Bear", "Creature — Bear", "", 2, 2)
	_ = theirs

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{early, late} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.ManaPool.Total(); got != 5 {
		t.Fatalf("pool %d after two attackers, want 5 (firebending 5 on the creature that was there, none on the late one)", got)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	if got := me.ManaPool.Total(); got != 5 {
		t.Errorf("pool %d in the damage step, want 5 (firebending mana lasts until end of combat)", got)
	}
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	if got := me.ManaPool.Total(); got != 0 {
		t.Errorf("pool %d after combat, want 0", got)
	}
}

func TestAvatarB13SozinsCometCanBeForetoldAndCastForTwoR(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := avB13Seats(g)
	bear := avB13Perm(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	pay := append(effectstest.CloudDeckMana("C", 2), game.ManaToken{Color: "R"})
	effectstest.CloudDeckForetellAndCast(t, g, "Sozin's Comet", "Sorcery", avB13SozinsCometOracle, "{2}{R}", pay, nil)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.ManaPool.Total(); got != 5 {
		t.Errorf("pool %d after the foretold Comet, want 5", got)
	}
}

// --- Tiger-Dillo -----------------------------------------------------------------

func TestAvatarB13TigerDilloNeedsAnotherOfYourCreaturesWithPowerFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := avB13Seats(g)
	dillo := avB13Perm(g, me.ID, "Tiger-Dillo", "Creature — Cat Armadillo", avB13TigerDilloOracle, 4, 3)
	avB13Perm(g, opp.ID, "Their Giant", "Creature — Giant", "", 5, 5)
	small := avB13Perm(g, me.ID, "Small Bear", "Creature — Bear", "", 3, 3)

	if r := effectstest.AvB13Restrictions(g, dillo); r&game.CantAttack == 0 || r&game.CantBlock == 0 {
		t.Fatalf("restrictions %v with only its own power 4, an opponent's 5/5 and a 3/3: want can't attack or block", r)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(dillo, opp.ID); err == nil {
		t.Fatal("Tiger-Dillo attacked without another creature with power 4 or greater")
	}

	// A +1/+1 counter lifts the 3/3 to power 4: the finished power counts.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(small, game.CounterPlusOne, 1) })
	if r := effectstest.AvB13Restrictions(g, dillo); r&(game.CantAttack|game.CantBlock) != 0 {
		t.Fatalf("restrictions %v with a 4/4 beside it, want none", r)
	}
	if err := g.DeclareAttacker(dillo, opp.ID); err != nil {
		t.Errorf("Tiger-Dillo could not attack beside a 4-power creature: %v", err)
	}
}

// --- Treetop Freedom Fighters ----------------------------------------------------

func TestAvatarB13TreetopFreedomFightersHasHasteAndMakesAnAlly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := avB13Seats(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "R", 3)
	id, err := effectstest.AvB13Cast(t, g, game.Card{Name: "Treetop Freedom Fighters", TypeLine: avB13TreetopTypeLine,
		ManaCost: "{2}{R}", OracleID: avB13TreetopOracle, Power: 2, Toughness: 1}, game.CastSpellParams{Strict: true})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, id, "haste") {
		t.Error("Treetop Freedom Fighters has no haste")
	}
	var allies int
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Ally" && c.Controller == me.ID && c.Power == 1 && c.Toughness == 1 &&
			len(c.Colors) == 1 && c.Colors[0] == "W" {
			allies++
		}
	}
	if allies != 1 {
		t.Errorf("%d 1/1 white Ally tokens, want 1", allies)
	}
}

// --- Twin Blades -----------------------------------------------------------------

func TestAvatarB13TwinBladesAttachesOnEntryAndGrantsDoubleStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := avB13Seats(g)
	bear := avB13Perm(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := avB13Perm(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)

	// Flash: cast in the beginning of combat step.
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	blades := effectstest.HandCardForTest(me, "Twin Blades", avB13TwinBladesTypeLine, avB13TwinBladesOracle)
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == blades {
				me.Hand.Cards[i].ManaCost = "{2}{R}"
			}
		}
	})
	effectstest.FillPoolColored(me, "R", 3)
	if err := g.CastSpell(me.ID, blades, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("flash-cast Twin Blades in combat: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the entry trigger")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
		t.Fatal("the entry trigger targeted an opponent's creature")
	}
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)

	c, ok := effectstest.BattlefieldCardByID(g, blades)
	if !ok || c.AttachedTo.ID != bear {
		t.Fatalf("Twin Blades attached to %v, want the bear", c.AttachedTo)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "double strike") {
		t.Error("the bear has no double strike")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("equipped bear %d/%d, want 3/3", p, tough)
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "double strike") {
		t.Error("an untargeted creature gained double strike")
	}
}

// --- Ty Lee, Artful Acrobat ------------------------------------------------------

func TestAvatarB13TyLeePaysOneToStopABlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := avB13Seats(g)
	tyLee := avB13Perm(g, me.ID, "Ty Lee, Artful Acrobat", avB13TyLeeTypeLine, avB13TyLeeOracle, 3, 2)
	wall := avB13Perm(g, opp.ID, "Wall", "Creature — Wall", "", 0, 4)
	if !effectstest.HasEffectiveKeyword(t, g, tyLee, "prowess") {
		t.Error("Ty Lee has no prowess")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(tyLee, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FillPoolColored(me, "R", 1)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, wall)
	effectstest.PassPriorityAroundTable(t, g)
	if r := effectstest.AvB13Restrictions(g, wall); r&game.CantBlock == 0 {
		t.Errorf("the wall's restrictions %v, want can't block", r)
	}
	if got := me.ManaPool.Total(); got != 0 {
		t.Errorf("pool %d, want the {1} spent", got)
	}
}

func TestAvatarB13TyLeeDecliningChoosesNoTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := avB13Seats(g)
	tyLee := avB13Perm(g, me.ID, "Ty Lee, Artful Acrobat", avB13TyLeeTypeLine, avB13TyLeeOracle, 3, 2)
	wall := avB13Perm(g, opp.ID, "Wall", "Creature — Wall", "", 0, 4)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(tyLee, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a target prompt appeared after declining")
	}
	if r := effectstest.AvB13Restrictions(g, wall); r&game.CantBlock != 0 {
		t.Error("the wall can't block although nothing was paid")
	}
}

// --- War Balloon -----------------------------------------------------------------

func TestAvatarB13WarBalloonBecomesACreatureAtThreeFireCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := avB13Seats(g)
	balloon := avB13Perm(g, me.ID, "War Balloon", avB13WarBalloonTypeLine, avB13WarBalloonOracle, 4, 3)
	effectstest.AdvanceToMain(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, balloon, "flying") {
		t.Error("War Balloon has no flying")
	}
	for i := 1; i <= 3; i++ {
		if effectstest.IsCreatureNow(g, balloon) {
			t.Fatalf("War Balloon is a creature with %d fire counters", i-1)
		}
		effectstest.FillPoolColored(me, "C", 1)
		if err := g.ActivateCatalogAbility(me.ID, balloon, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if n := effectstest.CountersOn(g, balloon, "fire"); n != 3 {
		t.Fatalf("%d fire counters, want 3", n)
	}
	if !effectstest.IsCreatureNow(g, balloon) {
		t.Fatal("War Balloon is not a creature with three fire counters")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, balloon); p != 4 || tough != 3 {
		t.Errorf("War Balloon %d/%d, want 4/3", p, tough)
	}
	if err := g.ActivateCatalogAbility(me.ID, balloon, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the {1} ability activated with an empty pool")
	}
}

func TestAvatarB13WarBalloonCrewThreeNeedsThreePower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := avB13Seats(g)
	balloon := avB13Perm(g, me.ID, "War Balloon", avB13WarBalloonTypeLine, avB13WarBalloonOracle, 4, 3)
	small := avB13Perm(g, me.ID, "Small", "Creature — Test", "", 2, 2)
	big := avB13Perm(g, me.ID, "Big", "Creature — Test", "", 3, 3)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, balloon, 1, game.ActivateAbilityParams{Strict: true, CrewIDs: []uuid.UUID{small}}); err == nil {
		t.Fatal("Crew 3 was paid with power 2")
	}
	if err := g.ActivateCatalogAbility(me.ID, balloon, 1, game.ActivateAbilityParams{Strict: true, CrewIDs: []uuid.UUID{big}}); err != nil {
		t.Fatalf("Crew 3 with power 3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.IsCreatureNow(g, balloon) {
		t.Error("the crewed War Balloon is not a creature")
	}
}

// --- Yuyan Archers ---------------------------------------------------------------

func TestAvatarB13YuyanArchersLootsOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := avB13Seats(g)
	effectstest.AdvanceToMain(t, g)
	pitch := effectstest.HandCardForTest(me, "Pitch", "Instant", "")
	effectstest.FillPoolColored(me, "R", 2)
	id, err := effectstest.AvB13Cast(t, g, game.Card{Name: "Yuyan Archers", TypeLine: avB13YuyanArchersTypeLne,
		ManaCost: "{1}{R}", OracleID: avB13YuyanArchersOracle, Power: 3, Toughness: 1}, game.CastSpellParams{Strict: true})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.AnswerDiscard(t, g, me.ID, pitch)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InGraveyardOf(g, me.ID, pitch) {
		t.Error("the chosen card was not discarded")
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d after discarding one and drawing one, want %d", got, hand)
	}
	if !effectstest.HasEffectiveKeyword(t, g, id, "reach") {
		t.Error("Yuyan Archers has no reach")
	}
}

func TestAvatarB13YuyanArchersDecliningDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := avB13Seats(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "R", 2)
	if _, err := effectstest.AvB13Cast(t, g, game.Card{Name: "Yuyan Archers", TypeLine: avB13YuyanArchersTypeLne,
		ManaCost: "{1}{R}", OracleID: avB13YuyanArchersOracle, Power: 3, Toughness: 1}, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.AnswerDiscard(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d after declining, want %d", got, hand)
	}
}

// --- Zuko, Exiled Prince ---------------------------------------------------------

func TestAvatarB13ZukoFirebendsThreeAndImpulses(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := avB13Seats(g)
	zuko := avB13Perm(g, me.ID, "Zuko, Exiled Prince", avB13ZukoTypeLine, avB13ZukoOracle, 4, 3)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, zuko, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("{3} activated with an empty pool")
	}
	topCard, err := me.Library.Top()
	if err != nil {
		t.Fatalf("library top: %v", err)
	}
	top := topCard.InstanceID
	effectstest.FillPoolColored(me, "C", 3)
	if err := g.ActivateCatalogAbility(me.ID, zuko, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{3}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(top) {
		t.Fatal("the top card of the library was not exiled")
	}
	if effectstest.GrantedPermissionOn(g, me.ID, top, game.ZoneExile) == nil {
		t.Error("the exiled card is not playable")
	}

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(zuko, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.ManaPool.Total(); got != 3 {
		t.Errorf("pool %d after Zuko attacked, want 3 (firebending 3)", got)
	}
}

// --- Allies at Last --------------------------------------------------------------

func avB13AlliesAtLast(t *testing.T, g *game.Game, targets []game.TargetRef) error {
	t.Helper()
	_, err := effectstest.AvB13Cast(t, g, game.Card{Name: "Allies at Last", TypeLine: "Instant", ManaCost: "{2}{G}",
		OracleID: avB13AlliesAtLastOracle}, game.CastSpellParams{Strict: true, Targets: targets})
	return err
}

func TestAvatarB13AlliesAtLastCostsLessPerAllyAndBothSourcesDealDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := avB13Seats(g)
	a := avB13Perm(g, me.ID, "Ally A", "Creature — Human Ally", "", 2, 2)
	b := avB13Perm(g, me.ID, "Ally B", "Creature — Human Ally", "", 3, 3)
	victim := avB13Perm(g, opp.ID, "Victim", "Creature — Giant", "", 5, 5)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "G", 1)
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}, {Kind: game.TargetCard, ID: victim}}
	if err := avB13AlliesAtLast(t, g, targets); err != nil {
		t.Fatalf("cast for {G} with two Allies: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the 5/5 survived 2 + 3 damage")
	}
	if !g.Battlefield.Contains(a) || !g.Battlefield.Contains(b) {
		t.Error("a source was hurt: the damage is one-sided")
	}
}

func TestAvatarB13AlliesAtLastCannotBePaidShortWithoutAllies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := avB13Seats(g)
	bear := avB13Perm(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	victim := avB13Perm(g, opp.ID, "Victim", "Creature — Giant", "", 5, 5)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "G", 1)
	if err := avB13AlliesAtLast(t, g, []game.TargetRef{{Kind: game.TargetCard, ID: bear}, {Kind: game.TargetCard, ID: victim}}); err == nil {
		t.Fatal("Allies at Last was cast for {G} with no Allies")
	}
}

func TestAvatarB13AlliesAtLastCannotAimAtYourOwnCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := avB13Seats(g)
	bear := avB13Perm(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	mine := avB13Perm(g, me.ID, "Mine", "Creature — Bear", "", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "G", 3)
	if err := avB13AlliesAtLast(t, g, []game.TargetRef{{Kind: game.TargetCard, ID: bear}, {Kind: game.TargetCard, ID: mine}}); err == nil {
		t.Fatal("Allies at Last targeted your own creature as the recipient")
	}
}

// --- Badgermole -------------------------------------------------------------------

func TestAvatarB13BadgermoleEarthbendsTwoAndGivesCounteredCreaturesTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := avB13Seats(g)
	land := effectstest.PushEarthbendLand(g, me.ID, "Forest", "Basic Land — Forest")
	theirLand := effectstest.PushEarthbendLand(g, opp.ID, "Forest", "Basic Land — Forest")
	bear := avB13Perm(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "G", 5)
	mole, err := effectstest.AvB13Cast(t, g, game.Card{Name: "Badgermole", TypeLine: avB13BadgermoleTypeLine, ManaCost: "{4}{G}",
		OracleID: avB13BadgermoleOracle, Power: 4, Toughness: 4}, game.CastSpellParams{Strict: true})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for earthbend")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirLand}); err == nil {
		t.Fatal("earthbend targeted an opponent's land")
	}
	effectstest.PickTriggerTarget(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)

	p, tough, creatureLand, hasty := effectstest.EarthbentBody(t, g, land)
	if !creatureLand || !hasty || p != 2 || tough != 2 {
		t.Fatalf("earthbent land %d/%d creature=%v haste=%v, want a 2/2 land creature with haste", p, tough, creatureLand, hasty)
	}
	if !effectstest.HasEffectiveKeyword(t, g, land, "trample") {
		t.Error("the earthbent land with +1/+1 counters has no trample")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "trample") || effectstest.HasEffectiveKeyword(t, g, mole, "trample") {
		t.Error("a creature without +1/+1 counters has trample")
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1) })
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the bear did not gain trample with its first +1/+1 counter")
	}
}
