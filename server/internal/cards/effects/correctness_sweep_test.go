package effects

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tapCard taps a battlefield permanent directly.
func tapCard(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
}

// --- Powerstone (#1727) -----------------------------------------------

// The token's mana pays for an artifact spell and for an activated
// ability, and refuses a nonartifact spell.
func TestPowerstoneManaPaysForArtifactsAndActivationsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	stone := pushToken(g, me.ID, PowerstoneToken())
	if err := g.ActivateManaAbility(me.ID, stone, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Powerstone: %v", err)
	}
	if len(me.ManaPool) != 1 {
		t.Fatalf("pool = %v, want one {C}", me.ManaPool)
	}
	cost, _ := game.ParseCost("{C}")

	artifact := game.Card{Name: "Sol Ring", TypeLine: "Artifact", ManaCost: "{1}"}
	creature := game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"}
	sorcery := game.Card{Name: "Divination", TypeLine: "Sorcery", ManaCost: "{2}{U}"}
	artifactCreature := game.Card{Name: "Myr Retriever", TypeLine: "Artifact Creature — Myr", ManaCost: "{3}"}

	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(artifact)) {
		t.Error("Powerstone mana refused an artifact spell")
	}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(artifactCreature)) {
		t.Error("Powerstone mana refused an artifact creature spell")
	}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(creature)) {
		t.Error("Powerstone mana refused an activated ability")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(creature)) {
		t.Error("Powerstone mana paid for a nonartifact creature spell — stronger than printed")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(sorcery)) {
		t.Error("Powerstone mana paid for a sorcery — stronger than printed")
	}
}

// --- Herd Heirloom (#1600) --------------------------------------------

func TestHerdHeirloomGrantsTrampleAndADrawTriggerToAPowerFourCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := pushCatalogPermanent(g, me.ID, "Herd Heirloom", "Artifact", herdHeirloomOracle, false)
	small := pushVanillaCreature(g, me.ID, "Bear", 3, 3)
	big := pushVanillaCreature(g, me.ID, "Ogre", 4, 4)

	if err := g.ActivateCatalogAbility(me.ID, rock, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: small}},
	}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("a power-3 creature: err = %v, want ErrIllegalTarget", err)
	}
	b16Activate(t, g, me.ID, rock, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: big}},
	})
	if !hasEffectiveKeyword(t, g, big, "trample") {
		t.Fatal("the target gains trample")
	}
	if hasEffectiveKeyword(t, g, small, "trample") {
		t.Error("only the target gains it")
	}

	hand := me.Hand.Size()
	attackWith(t, g, opp.ID, big)
	passPriorityAroundTable(t, g)
	if opp.Life != 40-4 {
		t.Errorf("opponent life %d, want 36", opp.Life)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want +1 from the granted combat-damage trigger", hand, me.Hand.Size())
	}
}

// "Spend this mana only to cast a creature spell" pays a creature
// SPELL and not a creature's activated ability (#2059). Before the fix
// the restriction was the type tag alone, which ManaSpendForAbility
// also satisfies when the ability's source is a creature — so the
// Heirloom's black paid Crypt Rats' {X}.
func TestHerdHeirloomManaCastsCreatureSpellsButNeverPaysACreaturesAbility(t *testing.T) {
	g, me, _ := spendTable(t)
	rock := pushCatalogPermanent(g, me.ID, "Herd Heirloom", "Artifact", herdHeirloomOracle, false)
	rats := pushCatalogPermanent(g, me.ID, "Crypt Rats", "Creature — Rat", cryptRatsOracle, false)

	if err := g.ActivateManaAbility(me.ID, rock, 0, game.ManaAbilityParams{Colors: []string{"B"}}); err != nil {
		t.Fatalf("tap the Heirloom for {B}: %v", err)
	}
	cost, _ := game.ParseCost("{B}")
	ratsCard, _ := battlefieldCard(g, rats)
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(ratsCard)) {
		t.Error("Herd Heirloom mana pays a creature's activated ability — stronger than printed")
	}
	refusedForMana(t, g.ActivateCatalogAbility(me.ID, rats, 0, game.ActivateAbilityParams{Strict: true, XValue: 1}),
		"Herd Heirloom's {B} for Crypt Rats' {X}")

	zombie := handSpell(me, "Black Creature", "Creature — Zombie", "{B}")
	if err := g.CastSpell(me.ID, zombie, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("Herd Heirloom's {B} for a creature spell: %v", err)
	}
}

// The #2059 guard: a mana ability whose restriction names a property
// of the OBJECT (a type, subtype, supertype or colour) must also name
// the PURPOSE — tags AND, and the object tags alone match an
// activation whose source has that property as readily as a spell. The
// one card that legitimately names no purpose prints both: Eldrazi
// Temple's "cast colorless Eldrazi spells or activate abilities of
// colorless Eldrazi". RestrictionsFunc lists (Cavern of Souls, Throne
// of Eldraine) are built per game and checked by their own card tests.
func TestRestrictedManaNamingAnObjectPropertyAlsoNamesAPurpose(t *testing.T) {
	castOrActivate := map[string]bool{"Eldrazi Temple": true}
	objectTag := func(tag string) bool {
		for _, p := range []string{"type:", "subtype:", "supertype:", "color:"} {
			if strings.HasPrefix(tag, p) {
				return true
			}
		}
		return tag == game.ManaRestrictColorless || tag == game.ManaRestrictMonocolored || tag == game.ManaRestrictMulticolored
	}
	for _, spec := range All() {
		for i, m := range spec.ManaAbilities {
			names, purpose := false, false
			for _, tag := range m.Restrictions {
				names = names || objectTag(tag)
				purpose = purpose || strings.HasPrefix(tag, "purpose:") || strings.HasPrefix(tag, "anyof:")
			}
			if names && !purpose && !castOrActivate[spec.Name] {
				t.Errorf("%s mana ability %d restricts %v with no purpose tag — add ManaRestrictCast (or ManaRestrictActivate)", spec.Name, i, m.Restrictions)
			}
		}
	}
}

// --- Midnight Clock (#1727) -------------------------------------------

func TestMidnightClockTwelfthCounterShufflesHandAndGraveyardDrawsSevenAndExilesItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	clock := pushCatalogPermanent(g, me.ID, "Midnight Clock", "Artifact", midnightClockOracle, false)
	if err := g.AddCounter(clock, "hour", 11); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	seedLibrary(me, "L1", "L2", "L3", "L4", "L5", "L6", "L7", "L8", "L9", "L10")
	me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Old Hand Card", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Old Grave Card", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})

	advanceToMain(t, g)
	b06AddMana(me, "U", "C", "C")
	b16Activate(t, g, me.ID, clock, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(clock) {
		t.Error("the Clock is exiled after the twelfth counter")
	}
	if !g.Exile.Contains(clock) {
		t.Error("the Clock should be in exile")
	}
	if got := me.Hand.Size(); got != 7 {
		t.Errorf("hand = %d, want 7", got)
	}
	if got := len(me.Graveyard.Cards); got != 0 {
		t.Errorf("graveyard = %d, want 0 (shuffled away)", got)
	}
}

// Only the crossing fires: an eleventh counter does nothing.
func TestMidnightClockEleventhCounterDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	clock := pushCatalogPermanent(g, me.ID, "Midnight Clock", "Artifact", midnightClockOracle, false)
	if err := g.AddCounter(clock, "hour", 10); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	advanceToMain(t, g)
	b06AddMana(me, "U", "C", "C")
	b16Activate(t, g, me.ID, clock, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(clock) {
		t.Error("eleven counters: the Clock stays")
	}
}

// --- "another" is object identity (#1738) ------------------------------

func TestUmbralCollarZealotCanSacrificeASecondZealotButNotItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	first := pushCatalogPermanent(g, me.ID, "Umbral Collar Zealot", "Creature — Human Cleric", umbralCollarZealotOracle, false)
	second := pushCatalogPermanent(g, me.ID, "Umbral Collar Zealot", "Creature — Human Cleric", umbralCollarZealotOracle, false)

	if err := g.ActivateCatalogAbility(me.ID, first, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{first},
	}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("sacrificing itself: err = %v, want ErrIllegalTarget", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, first, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{second},
	}); err != nil {
		t.Fatalf("sacrificing a second Zealot: %v", err)
	}
	if g.Battlefield.Contains(second) {
		t.Error("the second Zealot should be sacrificed")
	}
	if !g.Battlefield.Contains(first) {
		t.Error("the first Zealot stays")
	}
}

func TestSonicScrewdriverCanUntapAnotherScrewdriverButNotItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	first := pushCatalogPermanent(g, me.ID, "Sonic Screwdriver", "Artifact", sonicScrewdriverOracle, false)
	second := pushCatalogPermanent(g, me.ID, "Sonic Screwdriver", "Artifact", sonicScrewdriverOracle, false)
	tapCard(g, second)

	b06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, first, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: first}},
	}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("untapping itself: err = %v, want ErrIllegalTarget", err)
	}
	b16Activate(t, g, me.ID, first, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: second}},
	})
	if b16Tapped(t, g, second) {
		t.Error("the second Screwdriver should be untapped")
	}
}

// --- "up to X" (#1738) ---------------------------------------------------

func TestUpToXClauseAcceptsFewerTargetsButNeverMore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := pushVoltronCreature(g, me, "Mine A")
	b := pushVoltronCreature(g, me, "Mine B")
	advanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "March of Swirling Mist", TypeLine: "Instant",
		OracleID: marchOfSwirlingMistOracle, Owner: me.ID, Controller: me.ID})

	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 1, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b},
	}}); err == nil {
		t.Fatal("two targets with X = 1 should be refused")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 2, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: a},
	}}); err != nil {
		t.Fatalf("one target with X = 2 is \"up to X\": %v", err)
	}
}
