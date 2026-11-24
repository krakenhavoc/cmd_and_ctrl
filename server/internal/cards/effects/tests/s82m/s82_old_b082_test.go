package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OLD batch 82 (S82, #5507): Commander Legends encore creatures. Every test
// plays the real catalog entry and asserts what its text says.

const (
	oldB082Soul      = "4d211447-9f18-4d23-8f19-cdf6ad3c68ba"
	oldB082Dreadlord = "51b9d8ea-b0df-4a57-804e-6d395e301242"
	oldB082FinClade  = "dad2965a-ceb4-42ac-bb62-de7eea500011"
	oldB082Swordjack = "77828ebf-58d2-4aca-ae6f-0efa50d5d640"
	oldB082Courier   = "92c0346d-c375-4851-9806-5d3b9b83425f"
	oldB082Kitesail  = "462269c7-ca16-4aa7-bf52-f7fcda25a05b"
	oldB082Coastline = "c93bef6e-b676-43df-9f0f-026aaff25604"
	oldB082Briarblad = "6680a5f7-3ac1-4550-a29a-9553ad7ae465"
)

func TestOldB082EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB082Soul: "Soul of Eternity", oldB082Dreadlord: "Elvish Dreadlord", oldB082FinClade: "Fin-Clade Fugitives",
		oldB082Swordjack: "Fathom Fleet Swordjack", oldB082Courier: "Kinsbaile Courier",
		oldB082Kitesail: "Kitesail Skirmisher", oldB082Coastline: "Coastline Marauders", oldB082Briarblad: "Briarblade Adept",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
		if len(spec.Activated) != 1 {
			t.Errorf("%s: %d activated abilities, want encore only", name, len(spec.Activated))
		}
	}
}

func TestOldB082EncoreMakesATokenPerOpponentForEveryCard(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, mana string
		power, toughness             int
	}{
		{"Soul of Eternity", "Creature — Avatar", oldB082Soul, "{C}{C}{C}{C}{C}{C}{C}{W}{W}", 0, 0},
		{"Elvish Dreadlord", "Creature — Zombie Elf", oldB082Dreadlord, "{C}{C}{C}{C}{C}{B}{B}", 3, 3},
		{"Fin-Clade Fugitives", "Creature — Elf Salamander Rogue", oldB082FinClade, "{C}{C}{C}{C}{G}", 7, 4},
		{"Fathom Fleet Swordjack", "Creature — Orc Pirate", oldB082Swordjack, "{C}{C}{C}{C}{C}{R}", 4, 3},
		{"Kinsbaile Courier", "Creature — Kithkin Soldier", oldB082Courier, "{C}{C}{W}", 2, 1},
		{"Kitesail Skirmisher", "Creature — Human Pirate", oldB082Kitesail, "{C}{C}{C}{C}{U}", 3, 1},
		{"Coastline Marauders", "Creature — Human Pirate", oldB082Coastline, "{C}{C}{C}{C}{R}{R}", 0, 3},
		{"Briarblade Adept", "Creature — Elf Assassin", oldB082Briarblad, "{C}{C}{C}{B}", 3, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			_, me := effectstest.S46Encore(t, g, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness, tc.mana)
			if tokens := effectstest.S46Tokens(g, me.ID, tc.name); len(tokens) != 3 {
				t.Fatalf("%d tokens, want one per opponent", len(tokens))
			}
		})
	}
}

func TestOldB082EncoreRefusedWithoutEnoughMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushCatalogGraveyardCard(me, "Kinsbaile Courier", "Creature — Kithkin Soldier", oldB082Courier, 2, 1)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("encore activated with no mana")
	}
}

// --- Soul of Eternity ----------------------------------------------------

func TestOldB082SoulOfEternityIsAsBigAsYourLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	soul := effectstest.B12Push(g, me.ID, "Soul of Eternity", "Creature — Avatar", oldB082Soul, 0, 0)
	p, tough := effectstest.EffectivePTOf(t, g, soul)
	if p != me.Life || tough != me.Life {
		t.Fatalf("%d/%d at %d life", p, tough, me.Life)
	}
	effectstest.StSetLife(t, g, me, 7)
	if p, tough = effectstest.EffectivePTOf(t, g, soul); p != 7 || tough != 7 {
		t.Errorf("%d/%d at 7 life, want 7/7", p, tough)
	}
}

// --- Elvish Dreadlord ----------------------------------------------------

func TestOldB082ElvishDreadlordShrinksEveryNonElfWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	lord := effectstest.B12Push(g, me.ID, "Elvish Dreadlord", "Creature — Zombie Elf", oldB082Dreadlord, 3, 3)
	elf := effectstest.B12Creature(g, opp.ID, "Elf", "Creature — Elf Druid", 4, 4)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.TrekB18BDestroy(t, g, lord)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, elf); p != 4 || tough != 4 {
		t.Errorf("the Elf is %d/%d, want 4/4", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, ogre); p != 1 || tough != 1 {
		t.Errorf("the Ogre is %d/%d, want 1/1", p, tough)
	}
	if g.Battlefield.Contains(mine) {
		t.Error("the controller's own non-Elf 2/2 survived the -3/-3")
	}
}

// --- Fin-Clade Fugitives -------------------------------------------------

func TestOldB082FinCladeFugitivesCantBeBlockedByPowerTwoOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	fugitive := effectstest.B12Push(g, me.ID, "Fin-Clade Fugitives", "Creature — Elf Salamander Rogue", oldB082FinClade, 7, 4)
	small := effectstest.PushVanillaCreature(g, opp.ID, "Small", 2, 2)
	big := effectstest.PushVanillaCreature(g, opp.ID, "Big", 3, 3)
	can := func(blocker uuid.UUID) (ok bool) {
		g.WithWriteLock(func() {
			ok = g.CanBlockLocked(effectstest.FindBattlefieldCardByID(g, fugitive), effectstest.FindBattlefieldCardByID(g, blocker))
		})
		return ok
	}
	if can(small) {
		t.Error("a power-2 creature can block the Fugitives")
	}
	if !can(big) {
		t.Error("a power-3 creature cannot block the Fugitives")
	}
}

// --- Fathom Fleet Swordjack ----------------------------------------------

func TestOldB082SwordjackHitsTheAttackedPlayerForEachArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	jack := effectstest.B12Push(g, me.ID, "Fathom Fleet Swordjack", "Creature — Orc Pirate", oldB082Swordjack, 4, 3)
	effectstest.PushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	effectstest.PushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	effectstest.PushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	before, otherBefore := opp.Life, other.Life
	effectstest.DeclareAttack(t, g, opp.ID, jack)
	effectstest.PassPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 2 {
		t.Errorf("the attacked player lost %d, want 2 (my two artifacts, not theirs)", got)
	}
	if other.Life != otherBefore {
		t.Errorf("an unattacked player lost %d", otherBefore-other.Life)
	}
}

// --- Kinsbaile Courier ---------------------------------------------------

func TestOldB082KinsbaileCourierPutsACounterOnTargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Kinsbaile Courier", "Creature — Kithkin Soldier", oldB082Courier, 2, 1)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickCard(t, g, me.ID, bear)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var n int
	g.ReadSnapshot(func() { n = effectstest.PlusOneCounters(g, bear) })
	if n != 1 {
		t.Errorf("the Bear has %d +1/+1 counters, want 1", n)
	}
}

// --- Kitesail Skirmisher -------------------------------------------------

func TestOldB082KitesailGivesFlyingToAnotherAttackerOfTheSamePlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	skirmisher := effectstest.B12Push(g, me.ID, "Kitesail Skirmisher", "Creature — Human Pirate", oldB082Kitesail, 3, 1)
	same := effectstest.B12Creature(g, me.ID, "Same Target", "Creature — Bear", 2, 2)
	elsewhere := effectstest.B12Creature(g, me.ID, "Elsewhere", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for id, defender := range map[uuid.UUID]uuid.UUID{skirmisher: opp.ID, same: opp.ID, elsewhere: third.ID} {
		if err := g.DeclareAttacker(id, defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the attack trigger asked for no target")
	}
	if !effectstest.HasID(p.PickTargetCards, same) || effectstest.HasID(p.PickTargetCards, elsewhere) || effectstest.HasID(p.PickTargetCards, skirmisher) {
		t.Fatalf("legal targets must be only another creature attacking the same player: %+v", p.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, same)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, same, "flying") {
		t.Error("the chosen attacker did not gain flying")
	}
	if effectstest.HasEffectiveKeyword(t, g, elsewhere, "flying") {
		t.Error("an attacker of another player gained flying")
	}
}

// --- Coastline Marauders -------------------------------------------------

func TestOldB082CoastlineMaraudersGetsPowerPerDefendersLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	m := effectstest.B12Push(g, me.ID, "Coastline Marauders", "Creature — Human Pirate", oldB082Coastline, 0, 3)
	for i := 0; i < 3; i++ {
		effectstest.PushCatalogPermanent(g, opp.ID, "Island", "Basic Land — Island", "", false)
	}
	for i := 0; i < 5; i++ {
		effectstest.PushCatalogPermanent(g, other.ID, "Forest", "Basic Land — Forest", "", false)
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Mountain", "Basic Land — Mountain", "", false)
	effectstest.DeclareAttack(t, g, opp.ID, m)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, m); p != 3 || tough != 3 {
		t.Errorf("Marauders are %d/%d, want 3/3 (only the defender's three lands count)", p, tough)
	}
}

// --- Briarblade Adept ----------------------------------------------------

func TestOldB082BriarbladeAdeptShrinksAnOpponentsCreatureWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	adept := effectstest.B12Push(g, me.ID, "Briarblade Adept", "Creature — Elf Assassin", oldB082Briarblad, 3, 4)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, adept)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil || !effectstest.HasID(p.PickTargetCards, theirs) || effectstest.HasID(p.PickTargetCards, mine) {
		t.Fatalf("legal targets must be only creatures an opponent controls: %+v", p)
	}
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if pw, tough := effectstest.EffectivePTOf(t, g, theirs); pw != 1 || tough != 1 {
		t.Errorf("their Bear is %d/%d, want 1/1", pw, tough)
	}
	if pw, tough := effectstest.EffectivePTOf(t, g, mine); pw != 2 || tough != 2 {
		t.Errorf("my Bear is %d/%d, want 2/2", pw, tough)
	}
}
