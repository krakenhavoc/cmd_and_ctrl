package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_dmu_b07_cards_test.go — Dominaria United, S82 batch 7 (#5487): kicker
// with a delayed or replaced outcome. Each test plays the real catalog entry
// and asserts what its text says.

const (
	dmuB07AtrocityOracle    = "0dcd13ab-37f9-4d66-8227-1e72d920f297"
	dmuB07SkyriderOracle    = "eb1ab332-5549-4ff4-8afa-b385a236fb9b"
	dmuB07HydromancerOracle = "9a708f82-a4a1-47f9-8ff3-65608269eebd"
	dmuB07AmplifierOracle   = "a965898b-9c68-46cb-82a5-a1e11397f68f"
	dmuB07StrikeTeamOracle  = "6465f93d-e6bf-4a7b-86f3-5a2376f8b127"
	dmuB07WarLeechOracle    = "b92f2a4c-36dc-40a6-a874-7405fc97fb7a"
	dmuB07ThrallOracle      = "5bf1c26b-12eb-4be7-b9d6-f8ad75a778aa"
	dmuB07RestorationOracle = "df587772-516f-4ae8-aefe-e6fc63f8b365"
	dmuB07VortexOracle      = "21989fad-9672-46d4-b7d5-101906e7aed2"
	dmuB07FrenzyOracle      = "a0218a14-8301-4484-aed5-90334349620e"
	dmuB07UnsummonOracle    = "837182db-1bf3-4a2c-bd01-1af9d9873561"
)

func TestS82DmuB07EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		dmuB07AtrocityOracle:    "Balduvian Atrocity",
		dmuB07SkyriderOracle:    "Cleaving Skyrider",
		dmuB07HydromancerOracle: "Elvish Hydromancer",
		dmuB07AmplifierOracle:   "Ghitu Amplifier",
		dmuB07StrikeTeamOracle:  "Keldon Strike Team",
		dmuB07WarLeechOracle:    "Monstrous War-Leech",
		dmuB07ThrallOracle:      "In Thrall to the Pit",
		dmuB07RestorationOracle: "Sheoldred's Restoration",
		dmuB07VortexOracle:      "Rona's Vortex",
		dmuB07FrenzyOracle:      "Warhost's Frenzy",
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
		if len(spec.OptionalCosts) == 0 {
			t.Errorf("%s has no kicker", name)
		}
	}
}

// dmuB07Cast puts the card into the active player's hand and casts it where
// the game stands, kicked when optional is {0}. Strict refuses what the
// table would.
func dmuB07Cast(g *game.Game, name, typeLine, oracle string, targets []game.TargetRef, kicked, strict bool) (uuid.UUID, error) {
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: me.ID, Controller: me.ID})
	var optional []int
	if kicked {
		optional = []int{0}
	}
	return id, g.CastSpell(me.ID, id, game.CastSpellParams{Targets: targets, OptionalCosts: optional, Strict: strict})
}

func dmuB07Target(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func dmuB07Mainphase(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	return g, me, opp
}

func TestS82DmuB07BalduvianAtrocityReturnsASmallCreatureHastyAndSacrificesIt(t *testing.T) {
	g, me, _ := dmuB07Mainphase(t)
	small, big := uuid.New(), uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: small, Name: "Small Elf", TypeLine: "Creature — Elf", ManaCost: "{2}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: big, Name: "Big Elf", TypeLine: "Creature — Elf", ManaCost: "{2}{G}{G}", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})

	// Unkicked: nothing triggers (an intervening "if", CR 603.4).
	if _, err := dmuB07Cast(g, "Balduvian Atrocity", "Creature — Phyrexian Berserker", dmuB07AtrocityOracle, nil, false, false); err != nil {
		t.Fatalf("unkicked cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || len(g.PendingTriggers) != 0 || !me.Graveyard.Contains(small) {
		t.Fatal("an unkicked Atrocity triggered")
	}

	atrocity, err := dmuB07Cast(g, "Balduvian Atrocity", "Creature — Phyrexian Berserker", dmuB07AtrocityOracle, nil, true, false)
	if err != nil {
		t.Fatalf("kicked cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, atrocity, "menace") {
		t.Error("no menace")
	}
	effectstest.B04WaitForPick(t, g, me.ID)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: big}); err == nil {
		t.Fatal("a mana value 4 creature card was accepted as the target")
	}
	effectstest.PickCard(t, g, me.ID, small)
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, small); !ok || c.Controller != me.ID {
		t.Fatal("the creature card did not return to the battlefield")
	}
	if !effectstest.HasEffectiveKeyword(t, g, small, "haste") {
		t.Error("the returned creature has no haste")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, small); ok || !me.Graveyard.Contains(small) {
		t.Error("the returned creature was not sacrificed at the beginning of the next end step")
	}
	if _, ok := effectstest.BattlefieldCard(g, atrocity); !ok {
		t.Error("the Atrocity itself was sacrificed")
	}
}

func TestS82DmuB07CleavingSkyriderDealsDamageEqualToAttackersOnlyWhenKicked(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g, me, opp := dmuB07Mainphase(t)
		a := effectstest.PushVanillaCreature(g, me.ID, "Bear A", 2, 2)
		b := effectstest.PushVanillaCreature(g, me.ID, "Bear B", 2, 2)
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		for _, id := range []uuid.UUID{a, b} {
			if err := g.DeclareAttacker(id, opp.ID); err != nil {
				t.Fatalf("DeclareAttacker: %v", err)
			}
		}
		effectstest.LockInAttacks(t, g)
		before := opp.Life
		sky, err := dmuB07Cast(g, "Cleaving Skyrider", "Creature — Human Warrior", dmuB07SkyriderOracle, nil, kicked, false)
		if err != nil {
			t.Fatalf("kicked=%v cast (flash, in combat): %v", kicked, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.HasEffectiveKeyword(t, g, sky, "flying") {
			t.Error("no flying")
		}
		if !kicked {
			if effectstest.LatestPickTarget(g, me.ID) != nil || opp.Life != before {
				t.Error("an unkicked Skyrider dealt damage or asked for a target")
			}
			continue
		}
		effectstest.B04WaitForPick(t, g, me.ID)
		effectstest.PickTriggerTargetPlayer(t, g, me.ID, opp.ID)
		effectstest.PassPriorityAroundTable(t, g)
		if got := before - opp.Life; got != 2 {
			t.Errorf("damage to the opponent %d, want 2 (two attacking creatures)", got)
		}
	}
}

func TestS82DmuB07ElvishHydromancerCopiesAYourCreatureOnlyWhenKicked(t *testing.T) {
	g, me, _ := dmuB07Mainphase(t)
	bear := effectstest.B12Creature(g, me.ID, "Test Bear", "Creature — Bear", 2, 2)
	count := func() int {
		n := 0
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Name == "Test Bear" {
				n++
			}
		}
		return n
	}
	if _, err := dmuB07Cast(g, "Elvish Hydromancer", "Creature — Elf Wizard", dmuB07HydromancerOracle, nil, false, false); err != nil {
		t.Fatalf("unkicked cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || count() != 1 {
		t.Fatal("an unkicked Hydromancer made a copy")
	}
	if _, err := dmuB07Cast(g, "Elvish Hydromancer", "Creature — Elf Wizard", dmuB07HydromancerOracle, nil, true, false); err != nil {
		t.Fatalf("kicked cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if count() != 2 {
		t.Errorf("%d Test Bears on the battlefield, want 2 (the original and a token copy)", count())
	}
}

func TestS82DmuB07GhituAmplifierBouncesWhenKickedAndPumpsOnInstantsAndSorceries(t *testing.T) {
	g, me, opp := dmuB07Mainphase(t)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	amp, err := dmuB07Cast(g, "Ghitu Amplifier", "Creature — Human Wizard", dmuB07AmplifierOracle, nil, true, false)
	if err != nil {
		t.Fatalf("kicked cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, theirs); ok || !effectstest.InHandOf(opp, theirs) {
		t.Fatal("the opponent's creature was not returned to its owner's hand")
	}
	base := effectstest.EffectivePower(t, g, amp)
	other := effectstest.B12Creature(g, opp.ID, "Another Bear", "Creature — Bear", 2, 2)
	if _, err := dmuB07Cast(g, "Unsummon", "Instant", dmuB07UnsummonOracle, dmuB07Target(other), false, false); err != nil {
		t.Fatalf("instant: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, amp); got != base+2 {
		t.Errorf("power %d after casting an instant, want %d", got, base+2)
	}
}

func TestS82DmuB07GhituAmplifierUnkickedDoesNotBounce(t *testing.T) {
	g, me, opp := dmuB07Mainphase(t)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if _, err := dmuB07Cast(g, "Ghitu Amplifier", "Creature — Human Wizard", dmuB07AmplifierOracle, nil, false, false); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("an unkicked Amplifier asked for a target")
	}
	if _, ok := effectstest.BattlefieldCard(g, theirs); !ok {
		t.Error("the opponent's creature left")
	}
}

func TestS82DmuB07KeldonStrikeTeamMakesSoldiersWhenKickedAndGivesHasteOnlyTheTurnItEnters(t *testing.T) {
	g, me, _ := dmuB07Mainphase(t)
	bear := effectstest.B12Creature(g, me.ID, "Test Bear", "Creature — Bear", 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, bear, "haste") {
		t.Fatal("the bear had haste before the Strike Team")
	}
	soldiers := func() int {
		n := 0
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Name == "Soldier" && c.Controller == me.ID {
				n++
			}
		}
		return n
	}
	team, err := dmuB07Cast(g, "Keldon Strike Team", "Creature — Human Warrior", dmuB07StrikeTeamOracle, nil, false, false)
	if err != nil {
		t.Fatalf("unkicked cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if soldiers() != 0 {
		t.Error("an unkicked Strike Team made Soldiers")
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "haste") || !effectstest.HasEffectiveKeyword(t, g, team, "haste") {
		t.Error("creatures you control (itself included) lack haste the turn it entered")
	}
	if _, err := dmuB07Cast(g, "Keldon Strike Team", "Creature — Human Warrior", dmuB07StrikeTeamOracle, nil, true, false); err != nil {
		t.Fatalf("kicked cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if soldiers() != 2 {
		t.Errorf("%d Soldier tokens, want 2", soldiers())
	}
	// A Strike Team that did not enter this turn grants nothing.
	g2, me2, _ := dmuB07Mainphase(t)
	bear2 := effectstest.B12Creature(g2, me2.ID, "Test Bear", "Creature — Bear", 2, 2)
	effectstest.B12Push(g2, me2.ID, "Keldon Strike Team", "Creature — Human Warrior", dmuB07StrikeTeamOracle, 3, 1)
	if effectstest.HasEffectiveKeyword(t, g2, bear2, "haste") {
		t.Error("a Strike Team that entered on an earlier turn still grants haste")
	}
}

func TestS82DmuB07MonstrousWarLeechMillsFourWhenKickedAndSizesToTheGreatestManaValue(t *testing.T) {
	g, me, _ := dmuB07Mainphase(t)
	effectstest.SamiLibraryTopFirst(me,
		game.Card{Name: "One", TypeLine: "Sorcery", ManaCost: "{1}"},
		game.Card{Name: "Six", TypeLine: "Sorcery", ManaCost: "{5}{R}"},
		game.Card{Name: "Three", TypeLine: "Sorcery", ManaCost: "{2}{R}"},
		game.Card{Name: "Two", TypeLine: "Sorcery", ManaCost: "{1}{R}"},
		game.Card{Name: "Untouched", TypeLine: "Sorcery", ManaCost: "{9}"},
	)
	gy := me.Graveyard.Size()
	leech, err := dmuB07Cast(g, "Monstrous War-Leech", "Creature — Leech Horror", dmuB07WarLeechOracle, nil, true, false)
	if err != nil {
		t.Fatalf("kicked cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Graveyard.Size() - gy; got != 4 {
		t.Fatalf("%d cards milled, want 4", got)
	}
	if _, ok := effectstest.BattlefieldCard(g, leech); !ok {
		t.Fatal("the War-Leech did not survive")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, leech); p != 6 || tough != 6 {
		t.Errorf("%d/%d, want 6/6 (greatest mana value in the graveyard)", p, tough)
	}
}

func TestS82DmuB07MonstrousWarLeechUnkickedMillsNothing(t *testing.T) {
	g, me, _ := dmuB07Mainphase(t)
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Old", TypeLine: "Sorcery", ManaCost: "{3}", Owner: me.ID, Controller: me.ID})
	gy := me.Graveyard.Size()
	leech, err := dmuB07Cast(g, "Monstrous War-Leech", "Creature — Leech Horror", dmuB07WarLeechOracle, nil, false, false)
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Size() != gy {
		t.Error("an unkicked War-Leech milled cards")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, leech); p != 3 || tough != 3 {
		t.Errorf("%d/%d, want 3/3", p, tough)
	}
}

func TestS82DmuB07InThrallToThePitStealsAndOnlyKickedSacrificesIt(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g, me, opp := dmuB07Mainphase(t)
		theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == theirs {
					g.Battlefield.Cards[i].Tapped = true
				}
			}
		})
		if _, err := dmuB07Cast(g, "In Thrall to the Pit", "Sorcery", dmuB07ThrallOracle, dmuB07Target(theirs), kicked, false); err != nil {
			t.Fatalf("kicked=%v cast: %v", kicked, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		c, ok := effectstest.BattlefieldCard(g, theirs)
		if !ok || c.Controller != me.ID || c.Tapped {
			t.Fatalf("kicked=%v: the creature is not an untapped one of mine (%+v)", kicked, c)
		}
		if !effectstest.HasEffectiveKeyword(t, g, theirs, "haste") {
			t.Errorf("kicked=%v: no haste", kicked)
		}
		effectstest.AdvanceTo(t, g, game.StepEnd)
		effectstest.PassPriorityAroundTable(t, g)
		inYard := opp.Graveyard.Contains(theirs)
		if kicked && !inYard {
			t.Error("the kicked spell did not sacrifice the creature at the beginning of the next end step")
		}
		if !kicked && inYard {
			t.Error("the unkicked spell sacrificed the creature")
		}
	}
}

func TestS82DmuB07InThrallToThePitRefusesAnIllegalTarget(t *testing.T) {
	g, me, _ := dmuB07Mainphase(t)
	if _, err := dmuB07Cast(g, "In Thrall to the Pit", "Sorcery", dmuB07ThrallOracle, dmuB07Target(uuid.New()), false, true); err == nil {
		t.Error("a nonexistent target was accepted")
	}
	_ = me
}

func TestS82DmuB07SheoldredsRestorationGainsWhenKickedLosesOtherwiseAndExilesItself(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g, me, _ := dmuB07Mainphase(t)
		dead := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Giant", TypeLine: "Creature — Giant", ManaCost: "{3}{R}", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
		before := me.Life
		spell, err := dmuB07Cast(g, "Sheoldred's Restoration", "Sorcery", dmuB07RestorationOracle, dmuB07Target(dead), kicked, false)
		if err != nil {
			t.Fatalf("kicked=%v cast: %v", kicked, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if c, ok := effectstest.BattlefieldCard(g, dead); !ok || c.Controller != me.ID {
			t.Fatalf("kicked=%v: the creature did not return", kicked)
		}
		want := before - 4
		if kicked {
			want = before + 4
		}
		if me.Life != want {
			t.Errorf("kicked=%v: life %d, want %d", kicked, me.Life, want)
		}
		if !g.Exile.Contains(spell) || me.Graveyard.Contains(spell) {
			t.Errorf("kicked=%v: Sheoldred's Restoration was not exiled", kicked)
		}
	}
}

func TestS82DmuB07SheoldredsRestorationNeedsACreatureInYourGraveyard(t *testing.T) {
	g, me, opp := dmuB07Mainphase(t)
	theirs := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: theirs, Name: "Their Dead", TypeLine: "Creature — Giant", ManaCost: "{3}", Owner: opp.ID, Controller: opp.ID})
	if _, err := dmuB07Cast(g, "Sheoldred's Restoration", "Sorcery", dmuB07RestorationOracle, dmuB07Target(theirs), false, true); err == nil {
		t.Error("a creature card in an opponent's graveyard was accepted")
	}
	_ = me
}

func TestS82DmuB07RonasVortexBouncesOrPutsOnTheBottomWhenKicked(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g, me, opp := dmuB07Mainphase(t)
		theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		if _, err := dmuB07Cast(g, "Rona's Vortex", "Instant", dmuB07VortexOracle, dmuB07Target(theirs), kicked, false); err != nil {
			t.Fatalf("kicked=%v cast: %v", kicked, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if _, ok := effectstest.BattlefieldCard(g, theirs); ok {
			t.Fatalf("kicked=%v: the creature is still on the battlefield", kicked)
		}
		if kicked {
			bottom, err := opp.Library.Bottom()
			if err != nil || bottom.InstanceID != theirs || effectstest.InHandOf(opp, theirs) {
				t.Errorf("kicked: the creature is not on the bottom of its owner's library (%v)", err)
			}
		} else if !effectstest.InHandOf(opp, theirs) {
			t.Error("unkicked: the creature is not in its owner's hand")
		}
		_ = me
	}
}

func TestS82DmuB07RonasVortexRefusesYourOwnCreature(t *testing.T) {
	g, me, _ := dmuB07Mainphase(t)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if _, err := dmuB07Cast(g, "Rona's Vortex", "Instant", dmuB07VortexOracle, dmuB07Target(mine), false, true); err == nil {
		t.Error("a creature you control was accepted")
	}
}

func TestS82DmuB07WarhostsFrenzyPumpsAndOnlyKickedDrawsForEachDeath(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g, me, opp := dmuB07Mainphase(t)
		a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
		b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
		theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		pa := effectstest.EffectivePower(t, g, a)
		if _, err := dmuB07Cast(g, "Warhost's Frenzy", "Instant", dmuB07FrenzyOracle, nil, kicked, false); err != nil {
			t.Fatalf("kicked=%v cast: %v", kicked, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.EffectivePower(t, g, a); got != pa+2 {
			t.Errorf("kicked=%v: power %d, want %d", kicked, got, pa+2)
		}
		if got := effectstest.EffectivePower(t, g, theirs); got != 2 {
			t.Errorf("kicked=%v: an opponent's creature got the pump (%d)", kicked, got)
		}
		hand := me.Hand.Size()
		effectstest.RooncDestroy(t, g, a)
		effectstest.RooncDestroy(t, g, b)
		effectstest.RooncDestroy(t, g, theirs)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if kicked {
			want = 2
		}
		if got := me.Hand.Size() - hand; got != want {
			t.Errorf("kicked=%v: drew %d cards, want %d", kicked, got, want)
		}
	}
}
