package testss82m

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b448_test.go — Innistrad damage, flashback and Equipment, S82 OLD
// batch 448 (#5572), played through the real catalog entries.

const (
	oldB448Punishment = "a484120f-1027-4fc0-b4ea-cba94c262203"
	oldB448Cultist    = "a7fe5473-fae5-490e-8600-b7e497b9989e"
	oldB448Thrower    = "f9697a59-46e2-4612-b30b-d041a38fdd30"
	oldB448Doll       = "f9c8f9e1-a3a8-47f7-8d29-f954d14ad37b"
	oldB448Grasp      = "0d96e4bc-ff4e-4f37-89b6-8a4e3c1069a0"
	oldB448Journey    = "75e9ac31-5993-4039-93b2-9380ced2499a"
	oldB448Clutches   = "d12b875c-0056-4a57-a7ff-09c73d00f8cc"
	oldB448Stake      = "1b26dbb7-0ae5-4279-87b6-1b722c30a386"
	oldB448Dagger     = "93e9b0d6-97a8-4357-bcd4-f82d81bad232"
	oldB448Blade      = "f0031ef3-d5f5-4114-97aa-4beabd11ff8d"
)

func TestOldB448EveryCardIsRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		oldB448Punishment: "Heretic's Punishment", oldB448Cultist: "Skirsdag Cultist", oldB448Thrower: "Rage Thrower",
		oldB448Doll: "Creepy Doll", oldB448Grasp: "Grasp of Phantoms", oldB448Journey: "Memory's Journey",
		oldB448Clutches: "Nightbird's Clutches", oldB448Stake: "Wooden Stake", oldB448Dagger: "Silver-Inlaid Dagger",
		oldB448Blade: "Trepanation Blade",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

func oldB448Library(p *game.Player, cost string, n int) {
	for i := 0; i < n; i++ {
		p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Library Card", TypeLine: "Sorcery", ManaCost: cost, Owner: p.ID, Controller: p.ID})
	}
}

func TestOldB448HereticsPunishmentMillsThreeAndDealsTheGreatestManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	pun := effectstest.PushCatalogPermanent(g, me.ID, "Heretic's Punishment", "Enchantment", oldB448Punishment, false)
	g.WithWriteLock(func() { me.Library.Cards = nil })
	oldB448Library(me, "{1}", 1)
	oldB448Library(me, "{4}{R}", 1)
	oldB448Library(me, "{2}", 1)
	oldB448Library(me, "{6}", 1) // the fourth card is not milled
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Giant", 1, 9)
	effectstest.B06AddMana(me, "R", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, pun, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without a target")
	}
	if err := g.ActivateCatalogAbility(me.ID, pun, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.B16TargetCard(foe)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Library.Cards) != 1 || len(me.Graveyard.Cards) != 3 {
		t.Errorf("library %d / graveyard %d, want 1 / 3 after milling three", len(me.Library.Cards), len(me.Graveyard.Cards))
	}
	if got := effectstest.Pr6Marked(g, foe); got != 6 {
		t.Errorf("the creature has %d damage, want the greatest mana value milled, 6 ({6} was on top)", got)
	}
}

func TestOldB448HereticsPunishmentCanHitAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	pun := effectstest.PushCatalogPermanent(g, me.ID, "Heretic's Punishment", "Enchantment", oldB448Punishment, false)
	g.WithWriteLock(func() { me.Library.Cards = nil })
	oldB448Library(me, "{3}", 3)
	effectstest.B06AddMana(me, "R", "C", "C", "C")
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, pun, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d -> %d, want -3", life, opp.Life)
	}
}

func TestOldB448SkirsdagCultistSacrificesACreatureForTwoDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	cultist := effectstest.PushCatalogPermanent(g, me.ID, "Skirsdag Cultist", "Creature — Human Shaman", oldB448Cultist, false)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 2, 2)
	effectstest.B06AddMana(me, "R")
	params := game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}
	if err := g.ActivateCatalogAbility(me.ID, cultist, 0, params); err == nil {
		t.Fatal("activated without naming a creature to sacrifice")
	}
	params.SacrificeIDs = []uuid.UUID{fodder}
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, cultist, 0, params); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	if opp.Life != life-2 {
		t.Errorf("life %d -> %d, want -2", life, opp.Life)
	}
	if effectstest.TappedOnBattlefield(t, g, cultist) != true {
		t.Error("the Cultist did not tap")
	}
}

func TestOldB448RageThrowerPingsWhenAnotherCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	thrower := effectstest.B12Push(g, me.ID, "Rage Thrower", "Creature — Human Shaman", oldB448Thrower, 4, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	life := opp.Life
	effectstest.B25Destroy(g, bear)
	effectstest.Vow1Pick(t, g, me, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	if opp.Life != life-2 {
		t.Errorf("life %d -> %d, want -2 after another creature died", life, opp.Life)
	}
	effectstest.B25Destroy(g, thrower)
	g.RunStateChecksForTest()
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("the Thrower triggered off its own death")
	}
}

func TestOldB448CreepyDollDestroysWhatItDamagesOnAWonFlip(t *testing.T) {
	won, lost := 0, 0
	for table := 0; table < 80 && (won == 0 || lost == 0); table++ {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		doll := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Creepy Doll", OracleID: oldB448Doll,
			TypeLine: "Artifact Creature — Construct", Power: 1, Toughness: 1})
		wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 5)
		effectstest.Pr7bAttack(t, g, opp.ID, doll)
		effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{wall: doll})
		effectstest.Pr7bDamageStep(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
		if ask == nil {
			t.Fatal("no coin call was asked")
		}
		flips := len(effectstest.RandomEvents(g, game.EventFlipCoin))
		if err := g.ResolveCoinCall(ask.ID, me.ID, "heads"); err != nil {
			t.Fatalf("ResolveCoinCall: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		evs := effectstest.RandomEvents(g, game.EventFlipCoin)
		if len(evs) != flips+1 {
			t.Fatalf("%d coins flipped, want 1", len(evs)-flips)
		}
		if !g.Battlefield.Contains(doll) {
			t.Fatal("the indestructible Doll left the battlefield")
		}
		if evs[len(evs)-1].Won {
			won++
			if g.Battlefield.Contains(wall) {
				t.Error("won the flip but the damaged creature was not destroyed")
			}
		} else {
			lost++
			if !g.Battlefield.Contains(wall) {
				t.Error("lost the flip but the damaged creature was destroyed")
			}
		}
	}
	if won == 0 || lost == 0 {
		t.Fatalf("never saw both outcomes (won %d, lost %d)", won, lost)
	}
}

func TestOldB448GraspOfPhantomsPutsACreatureOnTopOfItsOwnersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.B06AddMana(me, "U", "C", "C", "C")
	if _, err := effectstest.CastFromHandErr(g, me, "Grasp of Phantoms", "Sorcery", "{3}{U}", oldB448Grasp,
		game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast without a target")
	}
	if _, err := effectstest.CastFromHandErr(g, me, "Grasp of Phantoms", "Sorcery", "{3}{U}", oldB448Grasp,
		game.CastSpellParams{Strict: true, Targets: effectstest.B16TargetCard(bear)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the creature stayed on the battlefield")
	}
	if n := len(opp.Library.Cards); n == 0 || opp.Library.Cards[n-1].InstanceID != bear {
		t.Error("the creature is not on top of its owner's library")
	}
}

func TestOldB448GraspOfPhantomsFlashbackCostsEightAndExiles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	grasp := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: grasp, Name: "Grasp of Phantoms", TypeLine: "Sorcery", ManaCost: "{3}{U}", OracleID: oldB448Grasp, Owner: me.ID, Controller: me.ID})
	effectstest.B06AddMana(me, "U", "C", "C", "C")
	params := game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "flashback", Targets: effectstest.B16TargetCard(bear)}
	if err := g.CastSpell(me.ID, grasp, params); err == nil {
		t.Fatal("flashback was paid with four mana instead of {7}{U}")
	}
	effectstest.B06AddMana(me, "C", "C", "C", "C")
	if err := g.CastSpell(me.ID, grasp, params); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(grasp) {
		t.Error("the flashed-back card was not exiled")
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the creature stayed on the battlefield")
	}
}

func TestOldB448MemorysJourneyShufflesUpToThreeCardsIn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	var ids []uuid.UUID
	for i := 0; i < 4; i++ {
		id := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Card", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		ids = append(ids, id)
	}
	effectstest.B06AddMana(me, "U", "C")
	picks := []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID, Slot: 0}}
	for _, id := range ids {
		picks = append(picks, game.TargetRef{Kind: game.TargetCard, ID: id, Slot: 1})
	}
	if _, err := effectstest.CastFromHandErr(g, me, "Memory's Journey", "Instant", "{1}{U}", oldB448Journey,
		game.CastSpellParams{Strict: true, Targets: picks}); err == nil {
		t.Fatal("four cards were accepted for 'up to three'")
	}
	libBefore := len(me.Library.Cards)
	if _, err := effectstest.CastFromHandErr(g, me, "Memory's Journey", "Instant", "{1}{U}", oldB448Journey,
		game.CastSpellParams{Strict: true, Targets: picks[:4]}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Library.Cards) != libBefore+3 {
		t.Errorf("library went from %d to %d, want +3", libBefore, len(me.Library.Cards))
	}
	if !me.Graveyard.Contains(ids[3]) {
		t.Error("the unchosen card left the graveyard")
	}
}

func TestOldB448NightbirdsClutchesBarsUpToTwoCreaturesFromBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "C", "Creature — Bear", 2, 2)
	effectstest.B06AddMana(me, "R", "C")
	if _, err := effectstest.CastFromHandErr(g, me, "Nightbird's Clutches", "Sorcery", "{1}{R}", oldB448Clutches,
		game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}, {Kind: game.TargetCard, ID: c}}}); err == nil {
		t.Fatal("three creatures were accepted for 'up to two'")
	}
	if _, err := effectstest.CastFromHandErr(g, me, "Nightbird's Clutches", "Sorcery", "{1}{R}", oldB448Clutches,
		game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]bool{a: true, b: true, c: false} {
		if got := effectstest.RestrictionsOf(t, g, id)&game.CantBlock != 0; got != want {
			t.Errorf("creature %s can't block = %v, want %v", id, got, want)
		}
	}
}

func oldB448Equip(t *testing.T, g *game.Game, me *game.Player, equipment uuid.UUID, cost string, host uuid.UUID) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, strings.Split(cost, "")...)
	if err := g.ActivateCatalogAbility(me.ID, equipment, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.B16TargetCard(host)}); err != nil {
		t.Fatalf("equip: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB448SilverInlaidDaggerGivesPlusThreeToAHuman(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dagger := effectstest.PushCatalogPermanent(g, me.ID, "Silver-Inlaid Dagger", "Artifact — Equipment", oldB448Dagger, false)
	human := effectstest.B12Creature(g, me.ID, "Human", "Creature — Human Soldier", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	oldB448Equip(t, g, me, dagger, "CC", bear)
	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Errorf("the Bear has power %d, want 3 (+2/+0)", got)
	}
	oldB448Equip(t, g, me, dagger, "CC", human)
	if got := effectstest.EffectivePower(t, g, human); got != 4 {
		t.Errorf("the Human has power %d, want 4 (+3/+0)", got)
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 1 {
		t.Errorf("the Bear kept power %d after the Dagger moved", got)
	}
}

func TestOldB448WoodenStakeDestroysAVampireItFights(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stake := effectstest.PushCatalogPermanent(g, me.ID, "Wooden Stake", "Artifact — Equipment", oldB448Stake, false)
	host := effectstest.B12Creature(g, me.ID, "Hunter", "Creature — Human", 1, 1)
	vampire := effectstest.B12Creature(g, opp.ID, "Vampire", "Creature — Vampire", 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	oldB448Equip(t, g, me, stake, "C", host)
	if got := effectstest.EffectivePower(t, g, host); got != 2 {
		t.Errorf("the equipped creature has power %d, want 2", got)
	}
	// A Vampire blocking the equipped creature is destroyed.
	effectstest.Pr7bAttack(t, g, opp.ID, host)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{vampire: host})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(vampire) {
		t.Error("the blocking Vampire survived the Stake")
	}
	_ = bear
}

func TestOldB448WoodenStakeIgnoresANonVampire(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stake := effectstest.PushCatalogPermanent(g, me.ID, "Wooden Stake", "Artifact — Equipment", oldB448Stake, false)
	host := effectstest.B12Creature(g, me.ID, "Hunter", "Creature — Human", 1, 5)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	oldB448Equip(t, g, me, stake, "C", host)
	effectstest.Pr7bAttack(t, g, opp.ID, host)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{bear: host})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("a non-Vampire blocker was destroyed")
	}
}

func TestOldB448TrepanationBladeRevealsUntilALandAndPumpsByTheCount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	blade := effectstest.PushCatalogPermanent(g, me.ID, "Trepanation Blade", "Artifact — Equipment", oldB448Blade, false)
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Human", 1, 1)
	oldB448Equip(t, g, me, blade, "CC", host)
	g.WithWriteLock(func() { opp.Library.Cards = nil })
	// Top first: spell, spell, land, spell. Three cards are revealed.
	opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Deep", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID})
	opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spell B", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spell A", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	effectstest.Pr7bAttack(t, g, opp.ID, host)
	effectstest.PassPriorityAroundTable(t, g)
	if len(opp.Library.Cards) != 1 || len(opp.Graveyard.Cards) != 3 {
		t.Errorf("library %d / graveyard %d, want 1 / 3", len(opp.Library.Cards), len(opp.Graveyard.Cards))
	}
	if got := effectstest.EffectivePower(t, g, host); got != 4 {
		t.Errorf("the attacker has power %d, want 1 + 3 revealed cards", got)
	}
}

func TestOldB448TrepanationBladeWithNoLandRevealsTheWholeLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	blade := effectstest.PushCatalogPermanent(g, me.ID, "Trepanation Blade", "Artifact — Equipment", oldB448Blade, false)
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Human", 1, 1)
	oldB448Equip(t, g, me, blade, "CC", host)
	g.WithWriteLock(func() { opp.Library.Cards = nil })
	for i := 0; i < 2; i++ {
		opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spell", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	}
	effectstest.Pr7bAttack(t, g, opp.ID, host)
	effectstest.PassPriorityAroundTable(t, g)
	if len(opp.Library.Cards) != 0 || len(opp.Graveyard.Cards) != 2 {
		t.Errorf("library %d / graveyard %d, want 0 / 2", len(opp.Library.Cards), len(opp.Graveyard.Cards))
	}
	if got := effectstest.EffectivePower(t, g, host); got != 3 {
		t.Errorf("the attacker has power %d, want 1 + 2", got)
	}
}
