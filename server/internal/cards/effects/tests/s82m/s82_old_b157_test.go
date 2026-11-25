package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b157_test.go — Core Set 2020, batch 157 (#5518): each test plays the
// real catalog entry through the real cast, activation and trigger paths.

const (
	oldB157SerpentOracle   = "6dc534e8-e22b-4e3c-aa30-c6e4c84f698c"
	oldB157PortalOracle    = "c30491b5-08fe-4694-9ac8-53be7b4b92ce"
	oldB157SalvagerOracle  = "c6dc70de-4fa5-466b-af28-0ced885d4e12"
	oldB157WakerootOracle  = "edc6db57-3519-4649-9c4a-5cf264cce26f"
	oldB157KethisOracle    = "1bf08ba0-4235-4dea-9722-ed991a3b6044"
	oldB157CavalierOracle  = "27c2b8ce-17b9-4f6c-8444-3bdc636dd413"
	oldB157BloodOracle     = "fc7439ec-32e3-47db-9ff4-6eb84cb33d89"
	oldB157ThoughtOracle   = "5f089ac6-9e92-4ec2-bf46-a0b08d1e2979"
	oldB157LegionOracle    = "7df8a8aa-1856-4fe6-a543-5de65da4f49f"
	oldB157CutthroatOracle = "e90e3f58-1197-41c1-81c5-008e946eb539"
	oldB157SweepOracle     = "0b34f471-40da-4c7a-81df-245fc3d14092"
	oldB157LeylineOracle   = "49a5c149-5067-4840-8385-c2989d7e5a34"
)

func TestOldB157CardsAreRegisteredComplete(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB157SerpentOracle:   "Frilled Sea Serpent",
		oldB157PortalOracle:    "Portal of Sanctuary",
		oldB157SalvagerOracle:  "Salvager of Ruin",
		oldB157WakerootOracle:  "Wakeroot Elemental",
		oldB157KethisOracle:    "Kethis, the Hidden Hand",
		oldB157CavalierOracle:  "Cavalier of Night",
		oldB157BloodOracle:     "Blood for Bones",
		oldB157ThoughtOracle:   "Thought Distortion",
		oldB157LegionOracle:    "Legion's End",
		oldB157CutthroatOracle: "Fathom Fleet Cutthroat",
		oldB157SweepOracle:     "Flame Sweep",
		oldB157LeylineOracle:   "Leyline of Combustion",
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

func oldB157Target(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func oldB157Activate(t *testing.T, g *game.Game, who, source uuid.UUID, idx int, params game.ActivateAbilityParams) error {
	t.Helper()
	params.Strict = true
	return g.ActivateCatalogAbility(who, source, idx, params)
}

func oldB157Graveyard(g *game.Game, owner uuid.UUID, name, typeLine, cost string, power, toughness int) uuid.UUID {
	return effectstest.B38Graveyard(g, owner, name, typeLine, cost, power, toughness)
}

func oldB157Zone(g *game.Game, id uuid.UUID) game.ZoneKind {
	var kind game.ZoneKind
	g.ReadSnapshot(func() {
		if z := g.FindCardZoneForEffect(id); z != nil {
			kind = z.Kind
		}
	})
	return kind
}

// --- Flame Sweep ---

func TestOldB157FlameSweepSparesOnlyYourFliers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myFlier := effectstest.B12Push(g, me.ID, "My Flier", "Creature — Bird", "", 1, 2)
	myBear := effectstest.B12Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	theirFlier := effectstest.B12Push(g, opp.ID, "Their Flier", "Creature — Bird", "", 1, 2)
	theirBig := effectstest.B12Push(g, opp.ID, "Their Giant", "Creature — Giant", "", 4, 4)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.InstanceID == myFlier || c.InstanceID == theirFlier {
				c.Keywords = append(c.Keywords, "flying")
			}
		}
	})
	effectstest.CastCatalogSpell(t, g, "Flame Sweep", "Instant", oldB157SweepOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(myFlier) {
		t.Error("a creature I control with flying was dealt damage")
	}
	if g.Battlefield.Contains(myBear) || g.Battlefield.Contains(theirFlier) {
		t.Error("a creature that should take 2 damage survived")
	}
	if got := effectstest.DamageMarkedOn(g, theirBig); got != 2 {
		t.Errorf("the 4/4 has %d damage marked, want 2", got)
	}
	if got := effectstest.DamageMarkedOn(g, myFlier); got != 0 {
		t.Errorf("my flier has %d damage marked, want 0", got)
	}
}

// --- Frilled Sea Serpent ---

func TestOldB157FrilledSeaSerpentCantBeBlockedThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	serpent := effectstest.PushCatalogPermanent(g, me.ID, "Frilled Sea Serpent", "Creature — Serpent", oldB157SerpentOracle, false)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	effectstest.B06AddMana(me, "U", "U", "U", "U", "U", "U", "U")
	if err := oldB157Activate(t, g, me.ID, serpent, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.BrAttack(t, g, serpent)
	if err := g.DeclareBlocker(blocker, serpent); err == nil {
		t.Error("a creature blocked the Serpent after it became unblockable")
	}
}

func TestOldB157FrilledSeaSerpentCanBeBlockedWithoutTheActivation(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	serpent := effectstest.PushCatalogPermanent(g, me.ID, "Frilled Sea Serpent", "Creature — Serpent", oldB157SerpentOracle, false)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	effectstest.BrAttack(t, g, serpent)
	if err := g.DeclareBlocker(blocker, serpent); err != nil {
		t.Errorf("an ordinary block was refused: %v", err)
	}
}

func TestOldB157FrilledSeaSerpentRefusesAPoolThatCantPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	serpent := effectstest.PushCatalogPermanent(g, me.ID, "Frilled Sea Serpent", "Creature — Serpent", oldB157SerpentOracle, false)
	effectstest.B06AddMana(me, "U", "U", "U")
	if err := oldB157Activate(t, g, me.ID, serpent, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("the ability was activated with 3 mana")
	}
}

// --- Portal of Sanctuary ---

func TestOldB157PortalReturnsTheCreatureAndEveryAuraToTheirOwnersHands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	portal := effectstest.PushCatalogPermanent(g, me.ID, "Portal of Sanctuary", "Artifact", oldB157PortalOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B15Aura(g, me.ID, bear)
	theirs := effectstest.B15Aura(g, opp.ID, bear)
	other := effectstest.B12Creature(g, me.ID, "Other Bear", "Creature — Bear", 2, 2)
	otherAura := effectstest.B15Aura(g, me.ID, other)
	effectstest.B06AddMana(me, "U")
	if err := oldB157Activate(t, g, me.ID, portal, 0, game.ActivateAbilityParams{Targets: oldB157Target(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if oldB157Zone(g, bear) != game.ZoneHand || oldB157Zone(g, mine) != game.ZoneHand {
		t.Errorf("the creature is in %v and my Aura in %v, want both in a hand", oldB157Zone(g, bear), oldB157Zone(g, mine))
	}
	if !opp.Hand.Contains(theirs) {
		t.Error("the opponent's Aura did not return to its owner's hand")
	}
	if !g.Battlefield.Contains(other) || !g.Battlefield.Contains(otherAura) {
		t.Error("a creature and Aura that were not chosen left the battlefield")
	}
}

func TestOldB157PortalCantTargetAnOpposingCreatureOrActOnTheirTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	portal := effectstest.PushCatalogPermanent(g, me.ID, "Portal of Sanctuary", "Artifact", oldB157PortalOracle, false)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.B06AddMana(me, "U")
	if err := oldB157Activate(t, g, me.ID, portal, 0, game.ActivateAbilityParams{Targets: oldB157Target(theirs)}); err == nil {
		t.Error("the Portal targeted a creature I don't control")
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.B06AddMana(me, "U")
	if err := oldB157Activate(t, g, me.ID, portal, 0, game.ActivateAbilityParams{Targets: oldB157Target(mine)}); err == nil {
		t.Error("the Portal was activated during an opponent's turn")
	}
}

// --- Salvager of Ruin ---

func TestOldB157SalvagerReturnsAPermanentCardThatDiedThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	salvager := effectstest.PushCatalogPermanent(g, me.ID, "Salvager of Ruin", "Artifact Creature — Construct", oldB157SalvagerOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Fallen Bear", "Creature — Bear", 2, 2)
	old := oldB157Graveyard(g, me.ID, "Old Bear", "Creature — Bear", "{1}{G}", 2, 2)
	effectstest.P1Destroy(t, g, bear)
	if err := oldB157Activate(t, g, me.ID, salvager, 0, game.ActivateAbilityParams{Targets: oldB157Target(old)}); err == nil {
		t.Error("a card that was in the graveyard before this turn was a legal target")
	}
	if err := oldB157Activate(t, g, me.ID, salvager, 0, game.ActivateAbilityParams{Targets: oldB157Target(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(salvager) {
		t.Error("the Salvager was not sacrificed")
	}
	if !me.Hand.Contains(bear) {
		t.Error("the creature that died this turn did not return to hand")
	}
}

// --- Wakeroot Elemental ---

func TestOldB157WakerootElementalUntapsALandAndMakesItAHastyElemental(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	wake := effectstest.PushCatalogPermanent(g, me.ID, "Wakeroot Elemental", "Creature — Elemental", oldB157WakerootOracle, false)
	land := effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == land {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	effectstest.B06AddMana(me, "G", "G", "G", "G", "G")
	if err := oldB157Activate(t, g, me.ID, wake, 0, game.ActivateAbilityParams{Targets: oldB157Target(land)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, land); got != 0 {
		t.Errorf("land damage %d", got)
	}
	if p, tt := effectstest.PtOf(t, g, land); p != 5 || tt != 5 {
		t.Errorf("the land is %d/%d, want 5/5", p, tt)
	}
	types := effectstest.EffectiveTypes(t, g, land)
	hasType := func(want string) bool {
		for _, ty := range types {
			if ty == want {
				return true
			}
		}
		return false
	}
	if !hasType("Creature") || !hasType("Land") {
		t.Errorf("types %v, want a land that is also a creature", types)
	}
	if !effectstest.HasEffectiveKeyword(t, g, land, "haste") {
		t.Error("the animated land has no haste")
	}
	if effectstest.IsTapped(g, land) {
		t.Error("the land was not untapped")
	}
	// It lasts past the turn, for as long as the land stays on the battlefield.
	effectstest.NextTurnOfMine(t, g)
	if p, _ := effectstest.PtOf(t, g, land); p != 5 {
		t.Errorf("the land is %d power on a later turn, want 5", p)
	}
}

func TestOldB157WakerootElementalCantTargetAnOpposingLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	wake := effectstest.PushCatalogPermanent(g, me.ID, "Wakeroot Elemental", "Creature — Elemental", oldB157WakerootOracle, false)
	land := effectstest.PushCatalogPermanent(g, opp.ID, "Forest", "Basic Land — Forest", "", false)
	effectstest.B06AddMana(me, "G", "G", "G", "G", "G")
	if err := oldB157Activate(t, g, me.ID, wake, 0, game.ActivateAbilityParams{Targets: oldB157Target(land)}); err == nil {
		t.Error("the ability targeted a land I don't control")
	}
}

// --- Kethis, the Hidden Hand ---

func TestOldB157KethisGrantsPlayFromTheGraveyardToLegendaryCardsLeftThere(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	kethis := effectstest.PushCatalogPermanent(g, me.ID, "Kethis, the Hidden Hand", "Legendary Creature — Elf Advisor", oldB157KethisOracle, false)
	a := oldB157Graveyard(g, me.ID, "Legend A", "Legendary Creature — Human", "{1}{G}", 2, 2)
	b := oldB157Graveyard(g, me.ID, "Legend B", "Legendary Creature — Human", "{1}{G}", 2, 2)
	c := oldB157Graveyard(g, me.ID, "Legend C", "Legendary Creature — Human", "{2}{G}", 3, 3)
	plain := oldB157Graveyard(g, me.ID, "Plain Bear", "Creature — Bear", "{1}{G}", 2, 2)
	if err := oldB157Activate(t, g, me.ID, kethis, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{a, b}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if oldB157Zone(g, a) != game.ZoneExile || oldB157Zone(g, b) != game.ZoneExile {
		t.Fatal("the two legendary cards were not exiled as the cost")
	}
	if effectstest.GrantedPermissionOn(g, me.ID, c, game.ZoneGraveyard) == nil {
		t.Error("the remaining legendary card has no permission to be played from the graveyard")
	}
	if effectstest.GrantedPermissionOn(g, me.ID, plain, game.ZoneGraveyard) != nil {
		t.Error("a nonlegendary card was given a permission")
	}
}

func TestOldB157KethisRefusesWithOneLegendaryCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	kethis := effectstest.PushCatalogPermanent(g, me.ID, "Kethis, the Hidden Hand", "Legendary Creature — Elf Advisor", oldB157KethisOracle, false)
	a := oldB157Graveyard(g, me.ID, "Legend A", "Legendary Creature — Human", "{1}{G}", 2, 2)
	if err := oldB157Activate(t, g, me.ID, kethis, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{a}}); err == nil {
		t.Error("the ability was paid with one card")
	}
}

func TestOldB157KethisMakesLegendarySpellsCostOneLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Kethis, the Hidden Hand", "Legendary Creature — Elf Advisor", oldB157KethisOracle, false)
	legend := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: legend, Name: "Legendary Bear", TypeLine: "Legendary Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	plain := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: plain, Name: "Plain Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	effectstest.B06AddMana(me, "G")
	if err := g.CastSpell(me.ID, legend, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("a legendary {1}{G} spell was not castable for {G}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B06AddMana(me, "G")
	if err := g.CastSpell(me.ID, plain, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a nonlegendary {1}{G} spell was castable for {G}")
	}
}

// --- Cavalier of Night ---

func oldB157Cavalier(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	return effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Cavalier of Night", TypeLine: "Creature — Elemental Knight", OracleID: oldB157CavalierOracle, Power: 4, Toughness: 5, Keywords: []string{"lifelink"}})
}

func TestOldB157CavalierSacrificesAnotherCreatureToDestroyOneOfTheirs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fodder := effectstest.PushVanillaCreature(g, me.ID, "Fodder", 1, 1)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	oldB157Cavalier(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayCounter(t, g, me.ID, fodder)
	effectstest.AnswerPickTarget(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	if g.Battlefield.Contains(victim) {
		t.Error("the opposing creature was not destroyed")
	}
}

func TestOldB157CavalierDecliningTheSacrificeDestroysNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fodder := effectstest.PushVanillaCreature(g, me.ID, "Fodder", 1, 1)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	oldB157Cavalier(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayCounter(t, g, me.ID, uuid.Nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(fodder) || !g.Battlefield.Contains(victim) {
		t.Error("declining the sacrifice still removed a creature")
	}
}

func TestOldB157CavalierCantSacrificeItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cav := oldB157Cavalier(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == me.ID {
			t.Errorf("a prompt %v was opened with no other creature to sacrifice", c.Kind)
		}
	}
	if !g.Battlefield.Contains(cav) {
		t.Error("the Cavalier left the battlefield")
	}
}

func TestOldB157CavalierReturnsACheapCreatureWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cav := effectstest.B12Push(g, me.ID, "Cavalier of Night", "Creature — Elemental Knight", oldB157CavalierOracle, 4, 5)
	cheap := oldB157Graveyard(g, me.ID, "Cheap Bear", "Creature — Bear", "{1}{G}", 2, 2)
	dear := oldB157Graveyard(g, me.ID, "Dear Bear", "Creature — Bear", "{3}{G}", 4, 4)
	effectstest.P1Destroy(t, g, cav)
	effectstest.AnswerPickTarget(t, g, cheap)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) {
		t.Error("the mana value 2 creature did not return")
	}
	if g.Battlefield.Contains(dear) {
		t.Error("the mana value 4 creature returned")
	}
}

// --- Blood for Bones ---

func TestOldB157BloodForBonesReturnsOneToTheBattlefieldAndAnotherToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	fodder := effectstest.PushVanillaCreature(g, me.ID, "Fodder", 1, 1)
	first := oldB157Graveyard(g, me.ID, "First Bear", "Creature — Bear", "{1}{G}", 2, 2)
	second := oldB157Graveyard(g, me.ID, "Second Bear", "Creature — Bear", "{2}{G}", 3, 3)
	spell := effectstest.HandSorcery(me, "Blood for Bones", "{3}{B}")
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == spell {
				me.Hand.Cards[i].OracleID = oldB157BloodOracle
			}
		}
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "B", "B", "B", "B")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.SlChoose(t, g, me.ID, first)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(first) {
		t.Fatal("the first creature card did not return to the battlefield")
	}
	effectstest.SlChoose(t, g, me.ID, second)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(second) {
		t.Error("the second creature card did not return to hand")
	}
	if !me.Graveyard.Contains(fodder) {
		t.Error("the sacrificed creature is not in the graveyard")
	}
}

func TestOldB157BloodForBonesRefusesWithoutACreatureToSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spell := effectstest.HandSorcery(me, "Blood for Bones", "{3}{B}")
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == spell {
				me.Hand.Cards[i].OracleID = oldB157BloodOracle
			}
		}
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "B", "B", "B", "B")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("the spell was cast with no creature to sacrifice")
	}
}

// --- Thought Distortion ---

func TestOldB157ThoughtDistortionExilesNoncreatureNonlandCardsFromHandAndGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	oppHand := func(name, typeLine string) uuid.UUID {
		id := uuid.New()
		opp.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: opp.ID, Controller: opp.ID})
		return id
	}
	spellInHand := oppHand("Bolt", "Instant")
	creatureInHand := oppHand("Bear", "Creature — Bear")
	landInHand := oppHand("Forest", "Basic Land — Forest")
	artifactInHand := oppHand("Rock", "Artifact")
	spellInYard := oldB157Graveyard(g, opp.ID, "Old Bolt", "Instant", "{R}", 0, 0)
	creatureInYard := oldB157Graveyard(g, opp.ID, "Old Bear", "Creature — Bear", "{1}{G}", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Thought Distortion", "Sorcery", oldB157ThoughtOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{spellInHand, artifactInHand, spellInYard} {
		if oldB157Zone(g, id) != game.ZoneExile {
			t.Errorf("a noncreature, nonland card is in zone %v, want exile", oldB157Zone(g, id))
		}
	}
	if !opp.Hand.Contains(creatureInHand) || !opp.Hand.Contains(landInHand) || !opp.Graveyard.Contains(creatureInYard) {
		t.Error("a creature or land card was exiled")
	}
}

func TestOldB157ThoughtDistortionCantBeCounteredAndOnlyTargetsOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.CastCatalogSpell(t, g, "Thought Distortion", "Sorcery", oldB157ThoughtOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}})
	var uncounterable bool
	g.ReadSnapshot(func() { uncounterable = g.SpellCantBeCounteredForEffect(id) })
	if !uncounterable {
		t.Error("Thought Distortion can be countered")
	}
	effectstest.PassPriorityAroundTable(t, g)
	err := effectstest.CastCatalogSpellErrStrict(t, g, "Thought Distortion", "Sorcery", oldB157ThoughtOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	if err == nil {
		t.Error("Thought Distortion targeted its own controller")
	}
}

// --- Legion's End ---

func TestOldB157LegionsEndExilesEveryCopyAndTheNamedCardsInHandAndGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "Goblin Token Maker", "Creature — Goblin", 1, 1)
	b := effectstest.B12Creature(g, opp.ID, "Goblin Token Maker", "Creature — Goblin", 1, 1)
	other := effectstest.B12Creature(g, opp.ID, "Other Goblin", "Creature — Goblin", 1, 1)
	handCopy := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: handCopy, Name: "Goblin Token Maker", TypeLine: "Creature — Goblin", ManaCost: "{R}", Owner: opp.ID, Controller: opp.ID})
	handOther := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: handOther, Name: "Other Goblin", TypeLine: "Creature — Goblin", ManaCost: "{R}", Owner: opp.ID, Controller: opp.ID})
	yardCopy := oldB157Graveyard(g, opp.ID, "Goblin Token Maker", "Creature — Goblin", "{R}", 1, 1)
	effectstest.CastCatalogSpell(t, g, "Legion's End", "Sorcery", oldB157LegionOracle, oldB157Target(a))
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, handCopy, yardCopy} {
		if oldB157Zone(g, id) != game.ZoneExile {
			t.Errorf("a card named Goblin Token Maker is in zone %v, want exile", oldB157Zone(g, id))
		}
	}
	if !g.Battlefield.Contains(other) || !opp.Hand.Contains(handOther) {
		t.Error("a differently named card was exiled")
	}
}

func TestOldB157LegionsEndRefusesAManaValueThreeCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	big := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Three Drop", TypeLine: "Creature — Bear", ManaCost: "{2}{G}",
		Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID,
	})
	err := effectstest.CastCatalogSpellErrStrict(t, g, "Legion's End", "Sorcery", oldB157LegionOracle, oldB157Target(big))
	if err == nil {
		t.Error("Legion's End targeted a creature with mana value 3")
	}
}

// --- Fathom Fleet Cutthroat ---

func TestOldB157CutthroatDestroysOnlyACreatureDealtDamageThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hurt := effectstest.B12Creature(g, opp.ID, "Hurt Giant", "Creature — Giant", 5, 5)
	fresh := effectstest.B12Creature(g, opp.ID, "Fresh Giant", "Creature — Giant", 5, 5)
	effectstest.Pr6Damage(t, g, me.ID, hurt, 1)
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Fathom Fleet Cutthroat", TypeLine: "Creature — Human Pirate", OracleID: oldB157CutthroatOracle, Power: 3, Toughness: 3})
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: fresh}); err == nil {
			t.Error("an undamaged creature was accepted as the target")
		}
	}
	effectstest.AnswerPickTarget(t, g, hurt)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(hurt) {
		t.Error("the damaged creature was not destroyed")
	}
	if !g.Battlefield.Contains(fresh) {
		t.Error("the undamaged creature was destroyed")
	}
}

// --- Leyline of Combustion ---

func oldB157Leyline(g *game.Game, me *game.Player) uuid.UUID {
	return effectstest.B12Push(g, me.ID, "Leyline of Combustion", "Enchantment", oldB157LeylineOracle, 0, 0)
}

func TestOldB157LeylineDamagesAnOpponentWhoTargetsYouAndYourCreatureOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB157Leyline(g, me)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 5, 5)
	effectstest.AdvanceToMainOf(t, g, 1)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Arc Trail", "Sorcery", effectstest.ArcTrailOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}, {Kind: game.TargetCard, ID: mine}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("the opponent's life went %d -> %d, want exactly 2 damage for one spell that targets me and my creature", life, opp.Life)
	}
}

func TestOldB157LeylineDamagesAnOpponentWhoTargetsOneOfYourPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB157Leyline(g, me)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 5, 5)
	effectstest.AdvanceToMainOf(t, g, 1)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", "4457ed35-7c10-48c8-9776-456485fdf070",
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("the opponent's life went %d -> %d, want 2 damage", life, opp.Life)
	}
}

func TestOldB157LeylineIgnoresYourOwnSpellsAndOtherPlayersPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB157Leyline(g, me)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 5, 5)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 5, 5)
	life, myLife := opp.Life, me.Life
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", "4457ed35-7c10-48c8-9776-456485fdf070",
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", "4457ed35-7c10-48c8-9776-456485fdf070",
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life || me.Life != myLife {
		t.Errorf("life changed: opponent %d -> %d, me %d -> %d", life, opp.Life, myLife, me.Life)
	}
}
