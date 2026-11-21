package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 RPR batch 12 — Phyrexian infect creatures, Sun's Zenith and artifacts
// (#5908).

const (
	rprSwarmlord      = "4fc2755c-ae7b-4147-a573-b33679a461e8"
	rprPlagueStinger  = "047e761e-7bbe-403b-86d7-cb9fb0b20b21"
	rprCorrupter      = "8ba5e97b-3eac-4291-8674-4b402a5554bc"
	rprWhiteZenith    = "ea4a9299-c2b3-4092-89dd-28142b203d60"
	rprDisfigure      = "77eafe49-b9c5-461d-89c5-cec217dd2974"
	rprDoor           = "b919cca5-88c0-4104-b54e-49644fbf78a5"
	rprConfrontation  = "31123cb5-2dd5-4f2b-a2fd-8f87095972c6"
	rprIvoryTower     = "d416a4ed-9f16-4a8b-8aee-03c630e1eb6c"
	rprKeeningStone   = "d0f06250-25a5-4003-8f3a-6d85ef32ea10"
	rprLiquimetalCoat = "f4bdc551-c2eb-4a34-a3e3-b4a017c925af"
	rprLodestoneGolem = "a698c9f5-325a-49f6-8902-eef09f4260c5"
)

func TestRprB12CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		rprSwarmlord: "Phyrexian Swarmlord", rprPlagueStinger: "Plague Stinger", rprCorrupter: "Viridian Corrupter",
		rprWhiteZenith: "White Sun's Zenith", rprDisfigure: "Disfigure", rprDoor: "Door to Nothingness",
		rprConfrontation: "Epic Confrontation", rprIvoryTower: "Ivory Tower", rprKeeningStone: "Keening Stone",
		rprLiquimetalCoat: "Liquimetal Coating", rprLodestoneGolem: "Lodestone Golem",
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

// rprCast casts a card from the active seat's hand at main phase with the
// given pool, strictly.
func rprCast(t *testing.T, g *game.Game, name, typeLine, oracle, cost string, params game.CastSpellParams, mana ...string) error {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ApaMana(me, mana...)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
			OracleID: oracle, Owner: me.ID, Controller: me.ID})
	})
	params.Strict = true
	return g.CastSpell(me.ID, id, params)
}

func TestRprDisfigureShrinksAndKills(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	small := effectstest.B27Push(g, opp.ID, "Elf", "Creature — Elf", "", "{G}", 2, 2)
	big := effectstest.B27Push(g, opp.ID, "Ox", "Creature — Ox", "", "{3}", 3, 3)
	if err := rprCast(t, g, "Disfigure", "Instant", rprDisfigure, "{B}", game.CastSpellParams{Targets: effectstest.CardTarget(small)}, "B"); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the 2/2 survived -2/-2")
	}
	if err := rprCast(t, g, "Disfigure", "Instant", rprDisfigure, "{B}", game.CastSpellParams{Targets: effectstest.CardTarget(big)}, "B"); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(big) {
		t.Fatal("the 3/3 died")
	}
	if p := effectstest.EffectivePowerWithCounters(t, g, big); p != 1 {
		t.Errorf("3/3 power %d, want 1", p)
	}
}

func TestRprDisfigureRefusesAnIllegalTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	land := effectstest.PushCatalogPermanent(g, g.Seats[1].ID, "Wastes", "Land", "", false)
	if err := rprCast(t, g, "Disfigure", "Instant", rprDisfigure, "{B}", game.CastSpellParams{Targets: effectstest.CardTarget(land)}, "B"); err == nil {
		t.Error("Disfigure targeted a land")
	}
}

func TestRprEpicConfrontationPumpsThenFights(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B27Push(g, me.ID, "Bear", "Creature — Bear", "", "{1}{G}", 2, 2)
	theirs := effectstest.B27Push(g, opp.ID, "Ox", "Creature — Ox", "", "{3}", 3, 3)
	if err := rprCast(t, g, "Epic Confrontation", "Sorcery", rprConfrontation, "{1}{G}",
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}}}, "G", "C"); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(theirs) {
		t.Error("the Ox survived 3 damage from the pumped Bear")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the 3/4 Bear died to 3 damage")
	}
}

func TestRprEpicConfrontationCannotTargetYourOwnSecondCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.B27Push(g, me.ID, "Bear", "Creature — Bear", "", "{1}{G}", 2, 2)
	b := effectstest.B27Push(g, me.ID, "Wolf", "Creature — Wolf", "", "{1}{G}", 2, 2)
	if err := rprCast(t, g, "Epic Confrontation", "Sorcery", rprConfrontation, "{1}{G}",
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}}, "G", "C"); err == nil {
		t.Error("fought a creature its controller controls")
	}
}

func TestRprWhiteSunsZenithMakesXCatsAndShufflesIntoTheLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	libBefore := me.Library.Size()
	if err := rprCast(t, g, "White Sun's Zenith", "Instant", rprWhiteZenith, "{X}{W}{W}{W}",
		game.CastSpellParams{XValue: 3}, "W", "W", "W", "C", "C", "C"); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldByName(g, "Cat"); n != 3 {
		t.Errorf("%d Cats, want 3", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Cat" && (c.Power != 2 || c.Toughness != 2) {
			t.Errorf("Cat is %d/%d, want 2/2", c.Power, c.Toughness)
		}
	}
	if me.Library.Size() != libBefore+1 {
		t.Errorf("library %d -> %d, want +1 (the Zenith)", libBefore, me.Library.Size())
	}
	for _, c := range me.Graveyard.Cards {
		if c.Name == "White Sun's Zenith" {
			t.Error("the Zenith went to the graveyard")
		}
	}
}

func TestRprWhiteSunsZenithNeedsWhiteWhiteWhite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	if err := rprCast(t, g, "White Sun's Zenith", "Instant", rprWhiteZenith, "{X}{W}{W}{W}",
		game.CastSpellParams{XValue: 1}, "W", "W", "C", "C"); err == nil {
		t.Error("cast without three white")
	}
}

func TestRprLiquimetalCoatingMakesAnyPermanentAnArtifactUntilEOT(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	coat := effectstest.PushCatalogPermanent(g, me.ID, "Liquimetal Coating", "Artifact", rprLiquimetalCoat, false)
	land := effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	effectstest.ApaActivate(t, g, me, coat, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(land)})
	if c := effectstest.ApaLive(g, land); c == nil || !c.IsArtifact() || !c.IsLand() {
		t.Errorf("the Forest is %+v, want an artifact land", c)
	}
	for i := 0; i < 20 && g.Turn.Step != game.StepCleanup; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatal(err)
	}
	if c := effectstest.ApaLive(g, land); c == nil || c.IsArtifact() {
		t.Error("the Forest is still an artifact after the turn ended")
	}
}

func TestRprLodestoneGolemTaxesNonartifactSpellsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, g.Seats[1].ID, "Lodestone Golem", "Artifact Creature — Golem", rprLodestoneGolem, false)
	target := effectstest.B27Push(g, me.ID, "Ox", "Creature — Ox", "", "{3}", 3, 3)
	if err := rprCast(t, g, "Disfigure", "Instant", rprDisfigure, "{B}", game.CastSpellParams{Targets: effectstest.CardTarget(target)}, "B"); err == nil {
		t.Error("a nonartifact spell was cast for its printed cost")
	}
	me.ManaPool.EmptyPool()
	if err := rprCast(t, g, "Ivory Tower", "Artifact", rprIvoryTower, "{1}", game.CastSpellParams{}, "C"); err != nil {
		t.Errorf("an artifact spell was taxed: %v", err)
	}
}

func TestRprLodestoneGolemTaxesTheNonartifactSpellByOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Lodestone Golem", "Artifact Creature — Golem", rprLodestoneGolem, false)
	target := effectstest.B27Push(g, g.Seats[1].ID, "Ox", "Creature — Ox", "", "{3}", 3, 3)
	if err := rprCast(t, g, "Disfigure", "Instant", rprDisfigure, "{B}", game.CastSpellParams{Targets: effectstest.CardTarget(target)}, "B", "C"); err != nil {
		t.Errorf("Disfigure for {1}{B} under the Golem: %v", err)
	}
}

func TestRprDoorToNothingnessEntersTappedAndMakesAPlayerLose(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	opp := effectstest.Top100Opponent(g)
	door := effectstest.PushCatalogPermanent(g, me.ID, "Door to Nothingness", "Artifact", rprDoor, false)
	// Pay the cost while it is untapped so the ability can be used.
	effectstest.ApaMana(me, "W", "W", "U", "U", "B", "B", "R", "R", "G", "G")
	effectstest.ApaActivate(t, g, me, door, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	if g.Battlefield.Contains(door) {
		t.Error("the Door was not sacrificed")
	}
	if !opp.Eliminated {
		t.Error("the target player did not lose")
	}
}

func TestRprDoorToNothingnessEntersTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	if err := rprCast(t, g, "Door to Nothingness", "Artifact", rprDoor, "{5}", game.CastSpellParams{}, "C", "C", "C", "C", "C"); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	id := effectstest.FindBattlefieldByName(g, "Door to Nothingness")
	effectstest.Top100AssertEnteredTapped(t, g, id, "Door to Nothingness")
}

func TestRprIvoryTowerGainsHandMinusFour(t *testing.T) {
	for _, tc := range []struct{ hand, gain int }{{3, 0}, {4, 0}, {7, 3}} {
		g := effectstest.NewCatalogGame(t)
		owner := g.Seats[1]
		effectstest.PushCatalogPermanent(g, owner.ID, "Ivory Tower", "Artifact", rprIvoryTower, false)
		g.WithWriteLock(func() {
			for owner.Hand.Size() > 0 {
				owner.Hand.Cards = owner.Hand.Cards[:0]
			}
			for i := 0; i < tc.hand; i++ {
				owner.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: "Land", Owner: owner.ID, Controller: owner.ID})
			}
		})
		life := owner.Life
		effectstest.AdvanceToUpkeepOf(t, g, 1)
		// The draw step has not happened yet: the hand is as seeded.
		effectstest.PassPriorityAroundTable(t, g)
		if owner.Life != life+tc.gain {
			t.Errorf("hand %d: life %d -> %d, want +%d", tc.hand, life, owner.Life, tc.gain)
		}
	}
}

func TestRprKeeningStoneMillsAsManyAsTheirGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	opp := effectstest.Top100Opponent(g)
	stone := effectstest.PushCatalogPermanent(g, me.ID, "Keening Stone", "Artifact", rprKeeningStone, false)
	g.WithWriteLock(func() {
		for i := 0; i < 3; i++ {
			opp.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Old", TypeLine: "Land", Owner: opp.ID, Controller: opp.ID})
		}
	})
	lib := opp.Library.Size()
	effectstest.ApaMana(me, "C", "C", "C", "C", "C")
	effectstest.ApaActivate(t, g, me, stone, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	if opp.Library.Size() != lib-3 {
		t.Errorf("library %d -> %d, want -3", lib, opp.Library.Size())
	}
	if len(opp.Graveyard.Cards) != 6 {
		t.Errorf("graveyard %d, want 6", len(opp.Graveyard.Cards))
	}
}

func TestRprPlagueStingerHasFlyingAndInfect(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.PushCatalogPermanent(g, g.Seats[0].ID, "Plague Stinger", "Creature — Phyrexian Insect Horror", rprPlagueStinger, false)
	c := effectstest.ApaLive(g, id)
	if c == nil || !game.HasKeyword(c, "flying") || !game.HasKeyword(c, "infect") {
		t.Errorf("Plague Stinger keywords: %+v", c)
	}
}

func TestRprViridianCorrupterDestroysAnArtifactOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	art := effectstest.PushCatalogPermanent(g, opp.ID, "Trinket", "Artifact", "", false)
	ownCreature := effectstest.B27Push(g, g.Seats[0].ID, "Bear", "Creature — Bear", "", "{1}{G}", 2, 2)
	if err := rprCast(t, g, "Viridian Corrupter", "Creature — Phyrexian Elf Shaman", rprCorrupter, "{1}{G}{G}",
		game.CastSpellParams{}, "G", "G", "C"); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.S12Drive(t, g, g.Seats[0].ID, art, ownCreature)
	if g.Battlefield.Contains(art) {
		t.Error("the artifact survived the Corrupter's enter trigger")
	}
	if effectstest.FindBattlefieldByName(g, "Viridian Corrupter") == uuid.Nil {
		t.Error("the Corrupter is not on the battlefield")
	}
}

func TestRprSwarmlordMakesAnInsectPerOpponentPoisonCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Phyrexian Swarmlord", "Creature — Phyrexian Insect Horror", rprSwarmlord, false)
	for seat, n := range map[int]int{1: 2, 2: 3, 0: 4} { // your own counters don't count
		if err := g.AddPlayerCounterByForEffect(me.ID, g.Seats[seat].ID, game.CounterPoison, n); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Phyrexian Insect" && c.Controller == me.ID {
			n++
			if v := effectstest.ApaLive(g, c.InstanceID); v == nil || !game.HasKeyword(v, "infect") {
				t.Error("an Insect token lacks infect")
			}
		}
	}
	if n != 5 {
		t.Errorf("%d Insect tokens, want 5", n)
	}
}
