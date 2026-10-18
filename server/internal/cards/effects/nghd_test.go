package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// nghd_test.go — #4092 slice d: Phenax, God of Deception; Phyrexian
// Rager; Phyrexian Revoker; Crippling Fear; Hex.

const (
	nghdPhenaxOracle   = "046ccc8a-ec0d-4dc5-b41d-5e8a6d86812a"
	nghdRagerOracle    = "e409c9be-0c9a-43c3-adb4-8c47afc1d551"
	nghdRevokerOracle  = "786b3657-b595-4b38-83f2-d40915e51e2b"
	nghdFearOracle     = "e7a7ba65-ad14-41d5-889b-9bbeae9ab3f7"
	nghdHexOracle      = "89457fa8-b0e2-4277-b6a9-c6601c59b46d"
	nghdZulaportOracle = "76b003e0-15af-4f22-bdf2-1ade5430964a"
)

func nghdGraveyardSize(g *game.Game, p *game.Player) int {
	n := 0
	g.ReadSnapshot(func() { n = len(p.Graveyard.Cards) })
	return n
}

func nghdLibrarySize(g *game.Game, p *game.Player) int {
	n := 0
	g.ReadSnapshot(func() { n = len(p.Library.Cards) })
	return n
}

// TestNghdCardsAreFullyAutomated pins Ian's requirement for the deck:
// every card of the slice is registered, complete, and carries no caveat.
func TestNghdCardsAreFullyAutomated(t *testing.T) {
	for oracle, name := range map[string]string{
		nghdPhenaxOracle:  "Phenax, God of Deception",
		nghdRagerOracle:   "Phyrexian Rager",
		nghdRevokerOracle: "Phyrexian Revoker",
		nghdFearOracle:    "Crippling Fear",
		nghdHexOracle:     "Hex",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("%s registered as %q", oracle, spec.Name)
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v, caveats %v — want full with none", name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Phyrexian Rager ------------------------------------------------------

func TestNghdRagerDrawsACardAndLosesOneLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hand, life := len(me.Hand.Cards), me.Life
	castCatalogSpell(t, g, "Phyrexian Rager", "Creature — Phyrexian Horror", nghdRagerOracle, nil)
	if got := len(me.Hand.Cards); got != hand {
		t.Fatalf("the trigger resolved before anyone could respond: hand %d → %d", hand, got)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("hand %d → %d, want +1 (the Rager itself left the hand, the draw replaced it)", hand, got)
	}
	if me.Life != life-1 {
		t.Errorf("life %d → %d, want -1", life, me.Life)
	}
}

// Both halves are one resolution: at 1 life the card is still drawn (the
// event is logged) and the player is eliminated by the state-based action
// afterwards.
func TestNghdRagerAtOneLifeStillDrawsAndThenLoses(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Life = 1
	castCatalogSpell(t, g, "Phyrexian Rager", "Creature — Phyrexian Horror", nghdRagerOracle, nil)
	seen := len(g.Events)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	drew := false
	for _, ev := range g.Events[seen:] {
		if ev.Kind == game.EventDrawCard && ev.Actor == me.ID {
			drew = true
		}
	}
	if !drew {
		t.Error("no draw was logged: the trigger lost the player before it finished drawing")
	}
	if !me.Eliminated {
		t.Errorf("life = %d, eliminated = %v; the lost life should have ended the player", me.Life, me.Eliminated)
	}
}

// --- Hex ------------------------------------------------------------------

func nghdSixCreatures(g *game.Game, owner uuid.UUID, n int) []uuid.UUID {
	ids := make([]uuid.UUID, n)
	for i := range ids {
		ids[i] = pushTribalCreature(g, owner, "Bear", "Creature — Bear", 2, 2)
	}
	return ids
}

func TestNghdHexNeedsExactlySixTargets(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	five := nghdSixCreatures(g, opp.ID, 5)
	if err := castCatalogSpellErr(t, g, "Hex", "Sorcery", nghdHexOracle, cardTargets(five...)); err == nil {
		t.Fatal("Hex was cast at five targets")
	}
}

func TestNghdHexDestroysSixTargetCreaturesAndSparesTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	six := append(nghdSixCreatures(g, opp.ID, 4), nghdSixCreatures(g, me.ID, 2)...)
	bystander := pushTribalCreature(g, opp.ID, "Bystander", "Creature — Human", 1, 1)
	castCatalogSpell(t, g, "Hex", "Sorcery", nghdHexOracle, cardTargets(six...))
	for _, id := range six {
		if !battlefieldHas(g, id) {
			t.Fatal("a target died before Hex resolved")
		}
	}
	passPriorityAroundTable(t, g)
	for _, id := range six {
		if battlefieldHas(g, id) {
			t.Fatal("a target survived Hex")
		}
	}
	if !battlefieldHas(g, bystander) {
		t.Error("Hex destroyed a creature it did not target")
	}
	if got := nghdGraveyardSize(g, opp); got != 4 {
		t.Errorf("opponent graveyard = %d, want 4", got)
	}
}

// The six die TOGETHER: a Zulaport Cutthroat among them sees every one
// of your creatures die, itself included (CR 603.10a). A sequential loop
// that destroyed the Cutthroat first would drain once.
func TestNghdHexDestroysTheSixAsOneEvent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cutthroat := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Zulaport Cutthroat", TypeLine: "Creature — Human Rogue Ally",
		OracleID: nghdZulaportOracle, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	others := nghdSixCreatures(g, me.ID, 5)
	targets := append([]uuid.UUID{cutthroat}, others...)
	life, oppLife := me.Life, opp.Life
	castCatalogSpell(t, g, "Hex", "Sorcery", nghdHexOracle, cardTargets(targets...))
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-6 || me.Life != life+6 {
		t.Errorf("opponent %d → %d, caster %d → %d; want the Cutthroat to see all six deaths (-6 / +6)",
			oppLife, opp.Life, life, me.Life)
	}
}

func TestNghdHexLeavesAnIndestructibleTarget(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	five := nghdSixCreatures(g, opp.ID, 5)
	god := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Stubborn", TypeLine: "Creature — Avatar",
		Power: 2, Toughness: 2, Keywords: []string{"indestructible"}, Owner: opp.ID, Controller: opp.ID,
	})
	castCatalogSpell(t, g, "Hex", "Sorcery", nghdHexOracle, cardTargets(append(five, god)...))
	passPriorityAroundTable(t, g)
	if !battlefieldHas(g, god) {
		t.Error("an indestructible creature was destroyed")
	}
	for _, id := range five {
		if battlefieldHas(g, id) {
			t.Error("a destructible target survived")
		}
	}
}

// A target that left in response is skipped; the other five still die
// (CR 608.2b).
func TestNghdHexSkipsATargetThatLeft(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	six := nghdSixCreatures(g, opp.ID, 6)
	castCatalogSpell(t, g, "Hex", "Sorcery", nghdHexOracle, cardTargets(six...))
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(six[0]); err != nil {
			t.Fatalf("remove one target: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	for _, id := range six {
		if battlefieldHas(g, id) {
			t.Error("a remaining target survived Hex")
		}
	}
}

// --- Crippling Fear -------------------------------------------------------

func TestNghdCripplingFearShrinksEveryCreatureNotOfTheType(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elf := pushTribalCreature(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	bigElf := pushTribalCreature(g, opp.ID, "Elvish Archdruid", "Creature — Elf Druid", 2, 4)
	goblin := pushTribalCreature(g, me.ID, "Goblin Guide", "Creature — Goblin Scout", 2, 2)
	ogre := pushTribalCreature(g, opp.ID, "Ogre", "Creature — Ogre", 5, 5)
	changeling := pushTribalCreature(g, opp.ID, "Mirror Entity", "Creature — Shapeshifter", 1, 1, game.KeywordChangeling)

	castAndPause(t, g, "Crippling Fear", "Sorcery", nghdFearOracle)
	if effectiveToughness(t, g, ogre) != 5 {
		t.Fatal("creatures shrank before the type was chosen")
	}
	answerCreatureType(t, g, me.ID, "Elf")

	if battlefieldHas(g, goblin) {
		t.Error("a 2/2 Goblin survived -3/-3")
	}
	if got := effectiveToughness(t, g, ogre); got != 2 {
		t.Errorf("Ogre toughness = %d, want 2", got)
	}
	if got := effectivePower(t, g, ogre); got != 2 {
		t.Errorf("Ogre power = %d, want 2", got)
	}
	for _, tc := range []struct {
		name string
		id   uuid.UUID
		p, t int
	}{{"Elf", elf, 1, 1}, {"opposing Elf", bigElf, 2, 4}, {"changeling", changeling, 1, 1}} {
		if !battlefieldHas(g, tc.id) {
			t.Errorf("the %s died, but it is of the chosen type", tc.name)
			continue
		}
		if got := effectiveToughness(t, g, tc.id); got != tc.t {
			t.Errorf("the %s toughness = %d, want %d (untouched)", tc.name, got, tc.t)
		}
	}
}

func TestNghdCripplingFearChoosesAtResolutionAndLocksTheSet(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := pushTribalCreature(g, opp.ID, "Ogre", "Creature — Ogre", 5, 5)
	castAndPause(t, g, "Crippling Fear", "Sorcery", nghdFearOracle)
	answerCreatureType(t, g, me.ID, "Sliver")
	late := pushTribalCreature(g, opp.ID, "Latecomer", "Creature — Ogre", 5, 5)
	if got := effectiveToughness(t, g, ogre); got != 2 {
		t.Errorf("Ogre toughness = %d, want 2", got)
	}
	if got := effectiveToughness(t, g, late); got != 5 {
		t.Errorf("a creature that entered after resolution was shrunk: toughness %d (CR 611.2c)", got)
	}
}

func TestNghdCripplingFearWearsOffAtEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := pushTribalCreature(g, opp.ID, "Ogre", "Creature — Ogre", 5, 5)
	castAndPause(t, g, "Crippling Fear", "Sorcery", nghdFearOracle)
	answerCreatureType(t, g, me.ID, "Elf")
	if got := effectiveToughness(t, g, ogre); got != 2 {
		t.Fatalf("Ogre toughness = %d, want 2", got)
	}
	advanceToUpkeepOf(t, g, 1)
	if got := effectiveToughness(t, g, ogre); got != 5 {
		t.Errorf("Ogre toughness next turn = %d, want 5 back", got)
	}
}

// --- Phyrexian Revoker ----------------------------------------------------

func nghdRevoker(t *testing.T, g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	t.Helper()
	id := b21Push(g, owner, "Phyrexian Revoker", "Artifact Creature — Phyrexian Horror", nghdRevokerOracle, 2, 1)
	g.WithWriteLock(func() {
		g.QueueCardNameChoiceForEffect(owner, id, "Phyrexian Revoker — choose a nonland card name")
	})
	actNameFor(t, g, owner, name)
	return id
}

// The ETB prompt is the as-enters card-name prompt, owed by the caster.
func TestNghdRevokerAsksForANameAsItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Phyrexian Revoker", "Artifact Creature — Phyrexian Horror", nghdRevokerOracle, nil)
	passPriorityAroundTable(t, g)
	c := pendingOfKind(g, game.PendingChoiceCardName)
	if c == nil || c.Chooser != me.ID {
		t.Fatalf("no card-name prompt for the caster: %+v", g.PendingChoices)
	}
}

// Unlike Pithing Needle there is no mana exemption: naming Sol Ring
// stops it tapping for mana, and naming Krenko stops his ability, for
// every player.
func TestNghdRevokerStopsTheNamedSourceIncludingManaAbilities(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, them := g.Seats[0], g.Seats[1]

	nghdRevoker(t, g, me.ID, "Sol Ring")
	solRing := b21Push(g, me.ID, "Sol Ring", "Artifact", actSolRingOracle, 0, 0)
	theirRing := b21Push(g, them.ID, "Sol Ring", "Artifact", actSolRingOracle, 0, 0)
	krenko := actPushKrenko(g, me.ID)

	if err := g.ActivateManaAbility(me.ID, solRing, 0, game.ManaAbilityParams{}); !errors.Is(err, game.ErrCantActivate) {
		t.Errorf("the named Sol Ring's mana ability: err = %v, want ErrCantActivate — the Revoker has no mana exemption", err)
	}
	if err := g.ActivateManaAbility(them.ID, theirRing, 0, game.ManaAbilityParams{}); !errors.Is(err, game.ErrCantActivate) {
		t.Errorf("the opponent's named Sol Ring: err = %v, want ErrCantActivate", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, krenko, 0, game.ActivateAbilityParams{}); err != nil {
		t.Errorf("an unnamed source: err = %v, want nil", err)
	}
}

func TestNghdRevokerStopsANamedNonManaAbility(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, them := g.Seats[0], g.Seats[1]
	nghdRevoker(t, g, them.ID, "Krenko, Mob Boss")
	krenko := actPushKrenko(g, me.ID)
	if err := g.ActivateCatalogAbility(me.ID, krenko, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrCantActivate) {
		t.Errorf("a named Krenko: err = %v, want ErrCantActivate", err)
	}
}

// "A nonland card name": naming a land's name reaches no land.
func TestNghdRevokerCannotStopALand(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	nghdRevoker(t, g, me.ID, "Forest")
	forest := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Colors: nil,
		Owner: me.ID, Controller: me.ID,
	})
	if err := g.ActivateManaAbility(me.ID, forest, 0, game.ManaAbilityParams{}); err != nil {
		t.Errorf("a Forest named by the Revoker: err = %v, want nil — a land name is not a legal choice", err)
	}
}

func TestNghdNonlandFaceNamed(t *testing.T) {
	spellLand := game.Card{
		Name: "Valakut Stoneforge", TypeLine: "Land",
		Faces: []game.Face{
			{Name: "Valakut Awakening", TypeLine: "Instant"},
			{Name: "Valakut Stoneforge", TypeLine: "Land"},
		},
	}
	for _, tc := range []struct {
		name string
		card game.Card
		said string
		want bool
	}{
		{"a plain nonland", game.Card{Name: "Sol Ring", TypeLine: "Artifact"}, "sol ring", true},
		{"a plain land", game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"}, "Forest", false},
		{"another name", game.Card{Name: "Sol Ring", TypeLine: "Artifact"}, "Mana Vault", false},
		{"the spell face of a modal land", spellLand, "Valakut Awakening", true},
		{"the land face of a modal land", spellLand, "Valakut Stoneforge", false},
		{"a face-down object", game.Card{Name: "Sol Ring", TypeLine: "Artifact", FaceDown: true}, "Sol Ring", false},
	} {
		if got := nonlandFaceNamed(tc.card, tc.said); got != tc.want {
			t.Errorf("%s: nonlandFaceNamed(%q) = %v, want %v", tc.name, tc.said, got, tc.want)
		}
	}
}

// --- Phenax, God of Deception --------------------------------------------

func nghdPhenax(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Phenax, God of Deception",
		TypeLine: "Legendary Enchantment Creature — God", OracleID: nghdPhenaxOracle,
		ManaCost: "{3}{U}{B}", Power: 4, Toughness: 7, Colors: []string{"U", "B"},
		Owner: owner, Controller: owner,
	})
}

// nghdDevotion adds `pips` blue-black pips of devotion as one permanent
// per pip, so the count is exact.
func nghdDevotion(g *game.Game, owner uuid.UUID, pips int) {
	for i := 0; i < pips; i++ {
		pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Pip", TypeLine: "Artifact", ManaCost: "{U}",
			Owner: owner, Controller: owner,
		})
	}
}

func TestNghdPhenaxIsACreatureOnlyAtSevenDevotion(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	phenax := nghdPhenax(g, me.ID)

	if containsStr(effectiveTypes(t, g, phenax), "Creature") {
		t.Error("with only Phenax's own two pips, he should be an enchantment, not a creature")
	}
	if !effectiveAbilitiesContain(t, g, phenax, "indestructible") {
		t.Error("Phenax should always be indestructible")
	}
	nghdDevotion(g, me.ID, 4) // 2 + 4 = 6
	if containsStr(effectiveTypes(t, g, phenax), "Creature") {
		t.Error("at devotion 6, Phenax is still not a creature")
	}
	nghdDevotion(g, me.ID, 1) // 7
	if !containsStr(effectiveTypes(t, g, phenax), "Creature") {
		t.Error("at devotion 7, Phenax should be a creature")
	}
	if !effectiveAbilitiesContain(t, g, phenax, "indestructible") {
		t.Error("a creature Phenax is still indestructible")
	}
}

// Devotion is blue AND black, and a hybrid blue-black symbol is one.
func TestNghdPhenaxCountsBlackAndBlueAndNotOtherColours(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	phenax := nghdPhenax(g, me.ID)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Black Pips", TypeLine: "Artifact", ManaCost: "{B}{B}{B}",
		Owner: me.ID, Controller: me.ID,
	})
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hybrid", TypeLine: "Artifact", ManaCost: "{U/B}{U/B}",
		Owner: me.ID, Controller: me.ID,
	})
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Red Pips", TypeLine: "Artifact", ManaCost: "{R}{R}{R}{R}",
		Owner: me.ID, Controller: me.ID,
	})
	// 2 (Phenax) + 3 black + 2 hybrid = 7; the red pips count for nothing.
	if !containsStr(effectiveTypes(t, g, phenax), "Creature") {
		t.Error("devotion should be 7: Phenax 2, three black, two hybrid; red does not count")
	}
}

func nghdMillAbility(t *testing.T, g *game.Game, controller, creature uuid.UUID, victim uuid.UUID) error {
	t.Helper()
	idx, ref := grantedActivatedIndex(t, g, creature)
	if idx < 0 {
		return errors.New("the creature has no granted ability")
	}
	return g.ActivateCatalogAbility(controller, creature, idx, game.ActivateAbilityParams{
		Ref: ref, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: victim}},
	})
}

func TestNghdPhenaxGivesEveryCreatureYouControlAMillAbility(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nghdPhenax(g, me.ID)
	mine := pushTribalCreature(g, me.ID, "Wall", "Creature — Wall", 0, 5)
	theirs := pushTribalCreature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)

	if idx, _ := grantedActivatedIndex(t, g, theirs); idx >= 0 {
		t.Error("an opponent's creature got Phenax's ability")
	}
	if idx, _ := grantedActivatedIndex(t, g, mine); idx < 0 {
		t.Fatal("my creature did not get Phenax's ability")
	}
	lib, grave := nghdLibrarySize(g, opp), nghdGraveyardSize(g, opp)
	if err := nghdMillAbility(t, g, me.ID, mine, opp.ID); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !tappedForTest(t, g, mine) {
		t.Error("the ability's {T} cost did not tap the creature")
	}
	passPriorityAroundTable(t, g)
	if got := lib - nghdLibrarySize(g, opp); got != 5 {
		t.Errorf("target player milled %d, want 5 (the Wall's toughness)", got)
	}
	if got := nghdGraveyardSize(g, opp) - grave; got != 5 {
		t.Errorf("graveyard grew by %d, want 5", got)
	}
}

// While Phenax is only an enchantment he gives the ability to the
// creatures around him and has none himself; once he is a creature he
// has it too, and X is his own toughness.
func TestNghdPhenaxGetsItHimselfOnlyAsACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	phenax := nghdPhenax(g, me.ID)
	if idx, _ := grantedActivatedIndex(t, g, phenax); idx >= 0 {
		t.Error("Phenax has the ability while he is not a creature")
	}
	nghdDevotion(g, me.ID, 5)
	if idx, _ := grantedActivatedIndex(t, g, phenax); idx < 0 {
		t.Fatal("Phenax does not have the ability as a creature")
	}
	lib := nghdLibrarySize(g, opp)
	if err := nghdMillAbility(t, g, me.ID, phenax, opp.ID); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := lib - nghdLibrarySize(g, opp); got != 7 {
		t.Errorf("Phenax milled %d, want 7", got)
	}
}

func TestNghdPhenaxAbilityNeedsAnUntappedNonSickCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nghdPhenax(g, me.ID)
	sick := pushCatalogPermanent(g, me.ID, "Sick", "Creature — Wall", "", true)
	if err := nghdMillAbility(t, g, me.ID, sick, opp.ID); err == nil {
		t.Error("a summoning-sick creature used a {T} ability (CR 302.6)")
	}
	ready := pushTribalCreature(g, me.ID, "Ready", "Creature — Wall", 0, 2)
	findBattlefieldCardForTest(g, ready).Tapped = true
	if err := nghdMillAbility(t, g, me.ID, ready, opp.ID); err == nil {
		t.Error("a tapped creature paid {T} again")
	}
}

// X is read as the ability resolves: a boost in response counts.
func TestNghdPhenaxReadsToughnessAtResolution(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nghdPhenax(g, me.ID)
	wall := pushTribalCreature(g, me.ID, "Wall", "Creature — Wall", 0, 2)
	lib := nghdLibrarySize(g, opp)
	if err := nghdMillAbility(t, g, me.ID, wall, opp.ID); err != nil {
		t.Fatalf("activate: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(wall, "+1/+1", 2); err != nil {
			t.Fatalf("pump in response: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if got := lib - nghdLibrarySize(g, opp); got != 4 {
		t.Errorf("milled %d, want 4 (toughness 2 + two counters at resolution)", got)
	}
}

// A creature that left before the ability resolved mills its
// last-known toughness (CR 608.2h).
func TestNghdPhenaxUsesLastKnownToughnessIfTheCreatureLeft(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nghdPhenax(g, me.ID)
	wall := pushTribalCreature(g, me.ID, "Wall", "Creature — Wall", 0, 4)
	lib := nghdLibrarySize(g, opp)
	if err := nghdMillAbility(t, g, me.ID, wall, opp.ID); err != nil {
		t.Fatalf("activate: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(wall); err != nil {
			t.Fatalf("remove the creature in response: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if got := lib - nghdLibrarySize(g, opp); got != 4 {
		t.Errorf("milled %d, want 4 (its last-known toughness)", got)
	}
}

// Phenax leaving takes the ability from everything (CR 613.8a).
func TestNghdPhenaxLeavingTakesTheAbilityAway(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	phenax := nghdPhenax(g, me.ID)
	wall := pushTribalCreature(g, me.ID, "Wall", "Creature — Wall", 0, 2)
	if idx, _ := grantedActivatedIndex(t, g, wall); idx < 0 {
		t.Fatal("setup: no granted ability")
	}
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(phenax); err != nil {
			t.Fatalf("exile Phenax: %v", err)
		}
	})
	if idx, _ := grantedActivatedIndex(t, g, wall); idx >= 0 {
		t.Error("the ability outlived Phenax")
	}
}
