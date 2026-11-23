package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b26_cards_test.go — Avatar: The Last Airbender Eternal, batch 26
// (#5075): every card cast, triggered or activated through the engine.

const (
	avB26HeartbeatOracle = "f15911a5-f597-452e-a015-b115339f28fc"
	avB26RitesOracle     = "7080fb43-5d93-41f5-87d3-bc1801805ea0"
	avB26JoinDanceOracle = "d8fc015c-7e6d-441d-bbeb-98ca74b21314"
	avB26KomaOracle      = "6ba0656f-e0d2-4735-9c66-3a00afa9fc23"
	avB26RhysOracle      = "f7252190-ad24-4ba3-a644-2790dd1d680d"
	avB26MeteoriteOracle = "0c3c3bcc-d485-4a01-92fe-8bf29c0fc926"
	avB26TreetopOracle   = "b53f216d-1592-4eee-b204-502a805fbc8c"
	avB26AppaOracle      = "72994a7a-7c0d-4f28-8cc0-28161d098d7f"
	avB26ReversalOracle  = "83ccbcb1-7cf0-41a0-8abd-18f4bc647b5f"
	avB26TurretOracle    = "2ce2a091-95b6-45e5-9097-77fea411162e"
)

func TestAvatarB26EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		avB26HeartbeatOracle: "Heartbeat of Spring",
		avB26RitesOracle:     "Rites of Flourishing",
		avB26JoinDanceOracle: "Join the Dance",
		avB26KomaOracle:      "Koma, Cosmos Serpent",
		avB26RhysOracle:      "Rhys the Redeemed",
		avB26MeteoriteOracle: "Meteorite",
		avB26TreetopOracle:   "Treetop Village",
		avB26AppaOracle:      "Appa, the Vigilant",
		avB26ReversalOracle:  "Katara's Reversal",
		avB26TurretOracle:    "Fire Nation Turret",
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

func avB26Tapped(t *testing.T, g *game.Game, id uuid.UUID, tapped bool) {
	t.Helper()
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = tapped
			}
		}
	})
}

// --- Heartbeat of Spring ------------------------------------------------------

func TestAvatarB26HeartbeatDoublesEveryPlayersLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Heartbeat of Spring", "Enchantment", avB26HeartbeatOracle, false)
	mine := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	theirs := effectstest.B12Permanent(g, them.ID, "Island", "Basic Land — Island")
	effectstest.TapForMana(t, g, me.ID, mine)
	if got := effectstest.SortedPool(me); len(got) != 2 || got[0] != "G" || got[1] != "G" {
		t.Errorf("my pool = %v, want two green", got)
	}
	effectstest.TapForMana(t, g, them.ID, theirs)
	if got := effectstest.SortedPool(them); len(got) != 2 || got[0] != "U" || got[1] != "U" {
		t.Errorf("their pool = %v, want two blue: 'a player', not 'you'", got)
	}
}

// --- Rites of Flourishing -----------------------------------------------------

func TestAvatarB26RitesGivesEveryPlayerAnExtraLandDrop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	before := 0
	g.ReadSnapshot(func() { before = g.EffectiveLandDropsLocked(opp) })
	effectstest.PushCatalogPermanent(g, me.ID, "Rites of Flourishing", "Enchantment", avB26RitesOracle, false)
	g.ReadSnapshot(func() {
		if n := g.EffectiveLandDropsLocked(opp); n != before+1 {
			t.Errorf("an opponent's land drops = %d, want %d: 'each player'", n, before+1)
		}
		if n := g.EffectiveLandDropsLocked(me); n != before+1 {
			t.Errorf("the controller's land drops = %d, want %d (one extra, not two)", n, before+1)
		}
	})
	if err := effectstest.B02cPlayLandRaw(t, g, "Forest"); err != nil {
		t.Fatalf("the turn's ordinary land drop: %v", err)
	}
	if err := effectstest.B02cPlayLandRaw(t, g, "Island"); err != nil {
		t.Fatalf("the extra land drop was refused: %v", err)
	}
	if err := effectstest.B02cPlayLandRaw(t, g, "Swamp"); err == nil {
		t.Error("a third land in one turn was allowed")
	}
}

func TestAvatarB26RitesEachPlayerDrawsAnExtraCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Rites of Flourishing", "Enchantment", avB26RitesOracle, false)
	// newCatalogGame parks on seat 0's draw step: the next draw step is
	// an opponent's.
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.AdvanceTo(t, g, game.StepDraw)
	active := g.Seats[g.Turn.ActiveSeat]
	if active.ID == me.ID {
		t.Fatal("setup: still the controller's turn")
	}
	handBefore := active.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if got := active.Hand.Size(); got != handBefore+1 {
		t.Errorf("the active opponent's hand is %d, want %d", got, handBefore+1)
	}
}

// --- Join the Dance -----------------------------------------------------------

func TestAvatarB26JoinTheDanceMakesTwoHumansAndFlashesBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Join the Dance", "Sorcery", avB26JoinDanceOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B43TokensNamed(g, me.ID, "Human"); n != 2 {
		t.Fatalf("Human tokens = %d, want 2", n)
	}
	id := effectstest.SeedGraveyardCard(t, g, "Join the Dance", "Sorcery", avB26JoinDanceOracle)
	effectstest.FloatMana(t, g, me, "{G}{W}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback", Strict: true}); err == nil {
		t.Fatal("flashback was paid with {G}{W}, not {3}{G}{W}")
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback", Strict: true}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B43TokensNamed(g, me.ID, "Human"); n != 4 {
		t.Errorf("Human tokens = %d, want 4", n)
	}
	if !g.Exile.Contains(id) {
		t.Error("the flashed-back sorcery was not exiled")
	}
}

// --- Koma, Cosmos Serpent -----------------------------------------------------

func TestAvatarB26KomaMakesACoilInEveryUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Koma, Cosmos Serpent", "Legendary Creature — Serpent", avB26KomaOracle, false)
	for want := 1; want <= 2; want++ {
		effectstest.AdvanceTo(t, g, game.StepEnd)
		effectstest.AdvanceTo(t, g, game.StepUpkeep)
		if g.Seats[g.Turn.ActiveSeat].ID == me.ID {
			t.Fatal("setup: expected an opponent's upkeep")
		}
		effectstest.PassPriorityAroundTable(t, g)
		if n := effectstest.B43TokensNamed(g, me.ID, "Koma's Coil"); n != want {
			t.Fatalf("Koma's Coils after %d opponent upkeeps = %d, want %d", want, n, want)
		}
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Koma's Coil" && (c.Power != 3 || c.Toughness != 3 || !c.IsCreature() || !effects.CardHasSubtype(c, "Serpent")) {
			t.Errorf("Koma's Coil is %d/%d %q", c.Power, c.Toughness, c.TypeLine)
		}
	}
}

func TestAvatarB26KomaCantBeCountered(t *testing.T) {
	spec, _ := effects.Lookup(avB26KomaOracle)
	if !spec.CantBeCountered {
		t.Error("Koma can be countered")
	}
}

func TestAvatarB26KomaSacrificesAnotherSerpentToLockOrSaveItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	koma := effectstest.PushCatalogPermanent(g, me.ID, "Koma, Cosmos Serpent", "Legendary Creature — Serpent", avB26KomaOracle, false)
	coil := effectstest.B12Creature(g, me.ID, "Koma's Coil", "Token Creature — Serpent", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rock := effectstest.PushCatalogPermanent(g, opp.ID, "Sol Ring", "Artifact", "6ad8011d-3471-4369-9d68-b264cc027487", false)

	if err := g.ActivateCatalogAbility(me.ID, koma, 0, game.ActivateAbilityParams{Strict: true, Modes: []int{1}, SacrificeIDs: []uuid.UUID{koma}}); err == nil {
		t.Fatal("Koma sacrificed itself")
	}
	if err := g.ActivateCatalogAbility(me.ID, koma, 0, game.ActivateAbilityParams{Strict: true, Modes: []int{1}, SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("a non-Serpent paid the cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, koma, 0, game.ActivateAbilityParams{
		Strict: true, Modes: []int{0}, SacrificeIDs: []uuid.UUID{coil},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
	}); err != nil {
		t.Fatalf("activate the lock: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(coil) {
		t.Error("the Coil was not sacrificed")
	}
	if c, _ := effectstest.BattlefieldCard(g, rock); !c.Tapped {
		t.Error("the target was not tapped")
	}
	avB26Tapped(t, g, rock, false)
	if err := g.ActivateManaAbility(opp.ID, rock, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("the locked Sol Ring's mana ability was activated this turn")
	}

	coil2 := effectstest.B12Creature(g, me.ID, "Sea Serpent", "Creature — Serpent", 5, 5)
	if err := g.ActivateCatalogAbility(me.ID, koma, 0, game.ActivateAbilityParams{Strict: true, Modes: []int{1}, SacrificeIDs: []uuid.UUID{coil2}}); err != nil {
		t.Fatalf("activate indestructible: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.EffectiveAbilitiesContain(t, g, koma, "indestructible") {
		t.Error("Koma did not gain indestructible")
	}
}

// --- Rhys the Redeemed --------------------------------------------------------

func TestAvatarB26RhysMakesAnElfWarriorThenCopiesEachCreatureToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	rhys := effectstest.PushCatalogPermanent(g, me.ID, "Rhys the Redeemed", "Legendary Creature — Elf Warrior", avB26RhysOracle, false)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, rhys, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without a green or white mana")
	}
	effectstest.FloatMana(t, g, me, "{W}")
	if err := g.ActivateCatalogAbility(me.ID, rhys, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate the token ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var elf game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Elf Warrior" && c.IsToken() {
			elf = c
		}
	}
	if elf.InstanceID == uuid.Nil || len(elf.Colors) != 2 || elf.Power != 1 {
		t.Fatalf("no 1/1 green and white Elf Warrior: %+v", elf)
	}
	effectstest.MscM11Token(t, g, me.ID, "1/1 white Human")
	g.WithWriteLock(func() {
		ctx := effects.NewContext(g, &game.StackItem{ID: uuid.New(), Controller: me.ID})
		if err := (effects.CreateToken{Template: effects.TreasureToken(), N: 1, Controller: me.ID}).Apply(ctx); err != nil {
			t.Fatalf("Treasure: %v", err)
		}
	})
	avB26Tapped(t, g, rhys, false)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{G}{W}")
	if err := g.ActivateCatalogAbility(me.ID, rhys, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate the copy ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B43TokensNamed(g, me.ID, "Elf Warrior"); n != 2 {
		t.Errorf("Elf Warriors = %d, want 2", n)
	}
	if n := effectstest.B43TokensNamed(g, me.ID, "Human"); n != 2 {
		t.Errorf("Humans = %d, want 2", n)
	}
	if n := effectstest.B43TokensNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("Treasures = %d, want 1: only creature tokens are copied", n)
	}
}

// --- Meteorite ----------------------------------------------------------------

func TestAvatarB26MeteoriteDealsTwoOnEntryAndTapsForAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.CastCatalogSpell(t, g, "Meteorite", "Artifact", avB26MeteoriteOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Meteorite did not enter")
	}
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 38 {
		t.Errorf("opponent life %d, want 38", opp.Life)
	}
	g.WithWriteLock(func() { me.ManaPool = nil })
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"B"}}); err != nil {
		t.Fatalf("tap Meteorite for {B}: %v", err)
	}
	if got := effectstest.SortedPool(me); len(got) != 1 || got[0] != "B" {
		t.Errorf("pool %v, want one mana", effectstest.SortedPool(me))
	}
}

// --- Treetop Village ----------------------------------------------------------

func TestAvatarB26TreetopVillageEntersTappedAndBecomesAnApe(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PlayLandFromHand(t, g, "Treetop Village", avB26TreetopOracle)
	effectstest.Top100AssertEnteredTapped(t, g, id, "Treetop Village")

	g2 := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g2)
	me := g2.Seats[g2.Turn.ActiveSeat]
	v := effectstest.PushCatalogPermanent(g2, me.ID, "Treetop Village", "Land", avB26TreetopOracle, false)
	effectstest.FloatMana(t, g2, me, "{C}{U}")
	if err := g2.ActivateCatalogAbility(me.ID, v, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("animated without green mana")
	}
	effectstest.FloatMana(t, g2, me, "{G}")
	if err := g2.ActivateCatalogAbility(me.ID, v, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("animate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	c := effectstest.BattlefieldCardPtr(t, g2, v)
	if !c.IsCreature() || !c.IsLand() || c.CurrentPower() != 3 || c.CurrentToughness() != 3 || !effects.CardHasSubtype(*c, "Ape") {
		t.Errorf("Treetop Village is %q %d/%d", c.TypeLine, c.CurrentPower(), c.CurrentToughness())
	}
	if !effectstest.EffectiveAbilitiesContain(t, g2, v, "trample") {
		t.Error("the Ape has no trample")
	}
}

// --- Appa, the Vigilant -------------------------------------------------------

func TestAvatarB26AppaPumpsTheTeamWhenAnAllyEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastAndResolveCreature(t, g, "Appa, the Vigilant", "Legendary Creature — Bison Ally", avB26AppaOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, bear); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("Bear is %d/%d after Appa entered, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, bear, "flying") || !effectstest.EffectiveAbilitiesContain(t, g, bear, "vigilance") {
		t.Error("the Bear did not gain flying and vigilance")
	}
	if c, _ := effectstest.BattlefieldCard(g, theirs); c.CurrentPower() != 2 {
		t.Error("an opponent's creature was pumped")
	}
	// Another Ally entering triggers it again; a non-Ally does not.
	effectstest.CastAndResolveCreature(t, g, "Plain Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, bear); c.CurrentPower() != 3 {
		t.Errorf("a non-Ally entering pumped the team: Bear power %d", c.CurrentPower())
	}
	effectstest.CastAndResolveCreature(t, g, "Some Ally", "Creature — Human Ally", "")
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, bear); c.CurrentPower() != 4 {
		t.Errorf("another Ally entering: Bear power %d, want 4", c.CurrentPower())
	}
}

// --- Katara's Reversal --------------------------------------------------------

func TestAvatarB26KatarasReversalCountersAndUntapsUpToFourEach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	src := effectstest.PushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	fired := false
	ability := effectstest.TaAnnounce(t, g, opp.ID, src, "Their Rock — do a thing", nil, &fired)
	bolt := effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	for _, id := range []uuid.UUID{bear, rock, land} {
		avB26Tapped(t, g, id, true)
	}
	if err := effectstest.CastCatalogSpellErr(t, g, "Katara's Reversal", "Instant", avB26ReversalOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: land, Slot: 1},
	}); err == nil {
		t.Fatal("a land was a legal untap target")
	}
	effectstest.CastCatalogSpell(t, g, "Katara's Reversal", "Instant", avB26ReversalOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: bolt, Slot: 0},
		{Kind: game.TargetCard, ID: ability, Slot: 0},
		{Kind: game.TargetCard, ID: bear, Slot: 1},
		{Kind: game.TargetCard, ID: rock, Slot: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) || me.Life != 40 {
		t.Errorf("the Bolt was not countered (life %d)", me.Life)
	}
	if fired {
		t.Error("the countered ability resolved")
	}
	for _, id := range []uuid.UUID{bear, rock} {
		if c, _ := effectstest.BattlefieldCard(g, id); c.Tapped {
			t.Errorf("%s is still tapped", c.Name)
		}
	}
	if c, _ := effectstest.BattlefieldCard(g, land); !c.Tapped {
		t.Error("an untargeted land was untapped")
	}
}

// --- Fire Nation Turret -------------------------------------------------------

func TestAvatarB26TurretPumpsAndGrantsFirebendingTwoAtCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Fire Nation Turret", "Artifact", avB26TurretOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, bear); c.CurrentPower() != 4 || c.CurrentToughness() != 2 {
		t.Errorf("Bear is %d/%d, want 4/2", c.CurrentPower(), c.CurrentToughness())
	}
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MkWantPool(t, me, map[string]int{"R": 2})
}

func TestAvatarB26TurretFiftyCountersDealFifty(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	turret := effectstest.PushCatalogPermanent(g, me.ID, "Fire Nation Turret", "Artifact", avB26TurretOracle, false)
	effectstest.FloatMana(t, g, me, "{R}")
	if err := g.ActivateCatalogAbility(me.ID, turret, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("charge: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, turret); c.Counters[game.CounterCharge] != 1 {
		t.Fatalf("charge counters = %d, want 1", c.Counters[game.CounterCharge])
	}
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == turret {
				g.Battlefield.Cards[i].Counters[game.CounterCharge] = 49
			}
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, turret, 1, game.ActivateAbilityParams{Strict: true, Targets: target}); err == nil {
		t.Fatal("activated with 49 charge counters")
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == turret {
				g.Battlefield.Cards[i].Counters[game.CounterCharge] = 50
			}
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, turret, 1, game.ActivateAbilityParams{Strict: true, Targets: target}); err != nil {
		t.Fatalf("activate with fifty: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != -10 {
		t.Errorf("opponent life %d, want -10", opp.Life)
	}
	if c, _ := effectstest.BattlefieldCard(g, turret); c.Counters[game.CounterCharge] != 0 {
		t.Errorf("%d charge counters left, want 0", c.Counters[game.CounterCharge])
	}
}
