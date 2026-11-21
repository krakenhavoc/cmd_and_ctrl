package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b24_cards_test.go — Avatar: The Last Airbender batches 24 and 25
// (#4459, #5075).

const (
	avB24TophOracle       = "8c1c0117-6fc0-43c5-a4c8-6472de6503d9"
	avB24PalaceOracle     = "f2000fb8-39c6-4ad6-a020-5245faaa1eba"
	avB24TeaShopOracle    = "d9a24444-289f-473f-9985-8df275257555"
	avB24MomoOracle       = "a3c56e71-0de6-447a-a9b4-0971f56c2362"
	avB24BroughtOracle    = "f6ed46e2-08f9-4e89-aae3-e5dd7f30f007"
	avB24LitaOracle       = "715cb547-726f-4bc8-ad46-297d4dccb881"
	avB24ReleaseOracle    = "3ae456e4-cb35-47be-b654-051e288622fe"
	avB24ScoutOracle      = "f65cba6a-3ef4-4841-b2c1-578b2a2c3300"
	avB24DreamsOracle     = "b3f9da94-c822-49b8-a1fc-be82189206af"
	avB24LegionOracle     = "690bd111-5b32-4156-b650-9afad3d6ef3d"
	avB24StandstillOracle = "fea4cb32-0329-4199-9abe-bff0cc45882a"
	avB24ZenithOracle     = "038c0165-32b6-4e81-8180-604b49905207"
	avB24CruelOracle      = "d9353eb6-c7e7-4d85-a0c2-32b1d091881a"
	avB24DocksideOracle   = "697bcfe1-ecbf-42a1-bfc7-0766d48ca56b"
	avB24DefectorOracle   = "1582e16d-4825-4cc5-b12c-449c45e4e449"
	avB24VolleyOracle     = "0324941f-484c-4656-8d1c-66e919c2bc7c"
	avB24SearingOracle    = "902b82fd-bb18-4833-b427-af8c9751f870"
	avB24SpreeOracle      = "fe725b27-49ef-41bc-91f9-9cd9871ad863"
)

func TestAvatarB24CardsAreRegistered(t *testing.T) {
	for oracle, name := range map[string]string{
		avB24TophOracle:       "Toph, Hardheaded Teacher",
		avB24PalaceOracle:     "Fire Nation Palace",
		avB24TeaShopOracle:    "Jasmine Dragon Tea Shop",
		avB24MomoOracle:       "Momo, Friendly Flier",
		avB24BroughtOracle:    "Brought Back",
		avB24LitaOracle:       "Lita, Mechanical Engineer",
		avB24ReleaseOracle:    "Release to Memory",
		avB24ScoutOracle:      "Scout's Warning",
		avB24DreamsOracle:     "Three Dreams",
		avB24LegionOracle:     "Clone Legion",
		avB24StandstillOracle: "Standstill",
		avB24ZenithOracle:     "Black Sun's Zenith",
		avB24CruelOracle:      "Cruel Tutor",
		avB24DocksideOracle:   "Dockside Extortionist",
		avB24DefectorOracle:   "Humble Defector",
		avB24VolleyOracle:     "Rending Volley",
		avB24SearingOracle:    "Searing Blood",
		avB24SpreeOracle:      "Shattering Spree",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func avB24Count(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			n++
		}
	}
	return n
}

// --- Rending Volley -----------------------------------------------------------

func TestAvatarB24RendingVolleyKillsAWhiteOrBlueCreatureOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	blue := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Blue Wall", TypeLine: "Creature — Wall",
		Power: 0, Toughness: 4, Colors: []string{"U"}, Owner: opp.ID, Controller: opp.ID})
	green := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Green Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Colors: []string{"G"}, Owner: opp.ID, Controller: opp.ID})
	if err := effectstest.CastCatalogSpellErr(t, g, "Rending Volley", "Instant", avB24VolleyOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: green}}); err == nil {
		t.Fatal("Rending Volley targeted a green creature")
	}
	effectstest.CastCatalogSpell(t, g, "Rending Volley", "Instant", avB24VolleyOracle, []game.TargetRef{{Kind: game.TargetCard, ID: blue}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(blue) {
		t.Error("the blue 0/4 survived 4 damage")
	}
	if spec, _ := effects.Lookup(avB24VolleyOracle); !spec.CantBeCountered {
		t.Error("Rending Volley can be countered")
	}
}

// --- Cruel Tutor ----------------------------------------------------------------

func TestAvatarB24CruelTutorPutsTheCardOnTopAndCostsTwoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	needle := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Cruel Needle"})
	life, hand := me.Life, me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Cruel Tutor", "Sorcery", avB24CruelOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, needle)
	if top, err := me.Library.Top(); err != nil || top.InstanceID != needle {
		t.Error("the found card is not on top of the library")
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d: the card goes on top, not to hand", me.Hand.Size(), hand)
	}
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
}

// --- Black Sun's Zenith ------------------------------------------------------------

func TestAvatarB24BlackSunsZenithShrinksEveryCreatureAndShufflesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Giant", "Creature — Giant", 4, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	library := me.Library.Size()
	zenith := effectstest.AvB24Hand(g, me, game.Card{Name: "Black Sun's Zenith", TypeLine: "Sorcery", OracleID: avB24ZenithOracle})
	if err := g.CastSpell(me.ID, zenith, game.CastSpellParams{XValue: 2}); err != nil {
		t.Fatalf("cast X=2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, mine, game.CounterMinusOne); got != 2 {
		t.Errorf("my creature has %d -1/-1 counters, want 2", got)
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("the 2/2 survived two -1/-1 counters")
	}
	if !me.Library.Contains(zenith) || me.Graveyard.Contains(zenith) {
		t.Error("Black Sun's Zenith was not shuffled into its owner's library")
	}
	if me.Library.Size() != library+1 {
		t.Errorf("library %d, want %d", me.Library.Size(), library+1)
	}
}

// --- Searing Blood ------------------------------------------------------------------

func TestAvatarB24SearingBloodBurnsTheControllerWhenTheCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Searing Blood", "Instant", avB24SearingOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the 2/2 survived 2 damage")
	}
	if opp.Life != life-3 {
		t.Errorf("controller life %d, want %d", opp.Life, life-3)
	}
}

func TestAvatarB24SearingBloodWaitsForTheCreatureToDieThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	ogre := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Searing Blood", "Instant", avB24SearingOracle, []game.TargetRef{{Kind: game.TargetCard, ID: ogre}})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ogre) || opp.Life != life {
		t.Fatalf("a surviving creature: on battlefield %v, life %d (want %d)", g.Battlefield.Contains(ogre), opp.Life, life)
	}
	// Anything that kills it this turn fires the trigger.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(ogre) })
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("controller life %d after it died, want %d", opp.Life, life-3)
	}
}

// --- Brought Back ---------------------------------------------------------------------

func TestAvatarB24BroughtBackReturnsTwoPermanentsThatDiedThisTurnTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	old := effectstest.PushGraveyardCardTyped(me, "Old Bear", "Creature — Bear")
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, me.ID, "Mind Stone", "Artifact")
	g.WithWriteLock(func() {
		_ = g.DestroyPermanentForEffect(bear)
		_ = g.DestroyPermanentForEffect(rock)
	})
	if err := effectstest.CastCatalogSpellErr(t, g, "Brought Back", "Instant", avB24BroughtOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: old}}); err == nil {
		t.Fatal("Brought Back targeted a card that was already in the graveyard")
	}
	effectstest.CastCatalogSpell(t, g, "Brought Back", "Instant", avB24BroughtOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}, {Kind: game.TargetCard, ID: rock}})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{bear, rock} {
		c, ok := effectstest.BattlefieldCard(g, id)
		if !ok {
			t.Fatalf("%s did not return", id)
		}
		if !c.Tapped {
			t.Errorf("%s returned untapped", c.Name)
		}
	}
}

// --- Release to Memory --------------------------------------------------------------------

func TestAvatarB24ReleaseToMemoryExilesTheGraveyardForSpirits(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.AvB24Graveyard(opp, game.Card{Name: "Dead Bear", TypeLine: "Creature — Bear"})
	b := effectstest.AvB24Graveyard(opp, game.Card{Name: "Dead Wurm", TypeLine: "Creature — Wurm"})
	c := effectstest.AvB24Graveyard(opp, game.Card{Name: "Old Bolt", TypeLine: "Instant"})
	if err := effectstest.CastCatalogSpellErr(t, g, "Release to Memory", "Instant", avB24ReleaseOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}); err == nil {
		t.Fatal("Release to Memory targeted its caster")
	}
	effectstest.CastCatalogSpell(t, g, "Release to Memory", "Instant", avB24ReleaseOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c} {
		if opp.Graveyard.Contains(id) {
			t.Errorf("%s is still in the graveyard", id)
		}
	}
	if got := avB24Count(g, me.ID, "Spirit"); got != 2 {
		t.Errorf("%d Spirits, want 2 (one per creature card)", got)
	}
}

// --- Scout's Warning --------------------------------------------------------------------------

func TestAvatarB24ScoutsWarningGivesTheNextCreatureFlashAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Scout's Warning", "Instant", avB24ScoutOracle, nil)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
	effectstest.MkAdvanceTo(t, g, game.StepBeginCombat)
	first := effectstest.AvB24Hand(g, me, game.Card{Name: "Flash Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	second := effectstest.AvB24Hand(g, me, game.Card{Name: "Slow Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the next creature spell could not be cast at instant speed: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a second creature spell was cast at instant speed")
	}
}

// --- Three Dreams --------------------------------------------------------------------------------

func TestAvatarB24ThreeDreamsFindsAurasWithDifferentNames(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Rancor", TypeLine: "Enchantment — Aura"})
	b := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Rancor", TypeLine: "Enchantment — Aura"})
	c := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Pacifism", TypeLine: "Enchantment — Aura"})
	bear := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear"})
	effectstest.CastCatalogSpell(t, g, "Three Dreams", "Sorcery", avB24DreamsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.SearchChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no search prompt")
	}
	effectstest.BeginSearchIfAsked(t, g, pick)
	if err := g.ResolveSearchLibrary(pick.ID, me.ID, []uuid.UUID{a, b}); err == nil {
		t.Fatal("two Auras with the same name were accepted")
	}
	if err := g.ResolveSearchLibrary(pick.ID, me.ID, []uuid.UUID{bear}); err == nil {
		t.Fatal("a non-Aura was accepted")
	}
	if err := g.ResolveSearchLibrary(pick.ID, me.ID, []uuid.UUID{a, c}); err != nil {
		t.Fatalf("two Auras with different names: %v", err)
	}
	if !me.Hand.Contains(a) || !me.Hand.Contains(c) || me.Hand.Contains(b) {
		t.Error("the picked Auras are not in hand")
	}
}

// --- Clone Legion -------------------------------------------------------------------------------------

func TestAvatarB24CloneLegionCopiesEachCreatureTheTargetControls(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 4, 4)
	effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Clone Legion", "Sorcery", avB24LegionOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if avB24Count(g, me.ID, "Their Bear") != 1 || avB24Count(g, me.ID, "Their Giant") != 1 {
		t.Error("one token copy of each creature was not created under the caster's control")
	}
	if avB24Count(g, me.ID, "Their Rock") != 0 {
		t.Error("a noncreature was copied")
	}
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Their Giant" && (!game.IsTokenPtr(&c) || c.Power != 4) {
			t.Errorf("the Giant copy is %+v", c)
		}
	}
}

// --- Standstill -----------------------------------------------------------------------------------------

func TestAvatarB24StandstillIsSacrificedAndTheCastersOpponentsDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	still := effectstest.PushCatalogPermanent(g, me.ID, "Standstill", "Enchantment", avB24StandstillOracle, false)
	effectstest.AdvanceToMainOf(t, g, 1)
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Hand.Size()
	}
	spell := effectstest.AvB24Hand(g, opp, game.Card{Name: "Opt Spell", TypeLine: "Sorcery"})
	before[opp.ID]++
	if err := g.CastSpell(opp.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(still) {
		t.Fatal("Standstill was not sacrificed")
	}
	for _, p := range g.Seats {
		want := before[p.ID] + 3
		if p.ID == opp.ID {
			want = before[p.ID] - 1
		}
		if p.Hand.Size() != want {
			t.Errorf("seat %s hand %d, want %d", p.Name, p.Hand.Size(), want)
		}
	}
	if me.Hand.Size() != before[me.ID]+3 {
		t.Error("Standstill's controller is an opponent of the caster and draws too")
	}
}

// --- Dockside Extortionist --------------------------------------------------------------------------------

func TestAvatarB24DocksideExtortionistCountsOpponentsArtifactsAndEnchantments(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	effectstest.B12Permanent(g, opp.ID, "Aura Thing", "Enchantment")
	effectstest.B12Permanent(g, third.ID, "Both", "Artifact Enchantment")
	effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Dockside Extortionist", "Creature — Goblin Pirate", avB24DocksideOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := avB24Count(g, me.ID, "Treasure"); got != 3 {
		t.Errorf("%d Treasures, want 3 (the artifact enchantment counts once, mine not at all)", got)
	}
}

// --- Humble Defector ------------------------------------------------------------------------------------------

func TestAvatarB24HumbleDefectorDrawsTwoAndChangesHands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	defector := effectstest.PushCatalogPermanent(g, me.ID, "Humble Defector", "Creature — Human Rogue", avB24DefectorOracle, false)
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, defector, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+2)
	}
	c, ok := effectstest.BattlefieldCard(g, defector)
	if !ok || c.Controller != opp.ID {
		t.Fatal("the opponent did not gain control of the Defector")
	}
	// On my turn the new controller can't activate it: only during THEIR turn.
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(defector) })
	if err := g.ActivateCatalogAbility(opp.ID, defector, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}, Strict: true}); err == nil {
		t.Error("activated outside its controller's turn")
	}
}

// --- Shattering Spree ---------------------------------------------------------------------------------------------

func TestAvatarB24ShatteringSpreeReplicatesForRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Permanent(g, opp.ID, "Rock A", "Artifact")
	b := effectstest.B12Permanent(g, opp.ID, "Rock B", "Artifact")
	effectstest.AdvanceToMain(t, g)
	spree := effectstest.AvB24Hand(g, me, game.Card{Name: "Shattering Spree", TypeLine: "Sorcery", ManaCost: "{R}", OracleID: avB24SpreeOracle})
	effectstest.FloatMana(t, g, me, "{R}")
	if err := g.CastSpell(me.ID, spree, game.CastSpellParams{Targets: effectstest.B16TargetCard(a), OptionalCosts: []int{0}, Strict: true}); err == nil {
		t.Fatal("cast with one replicate on {R} alone")
	}
	effectstest.FloatMana(t, g, me, "{R}")
	if err := g.CastSpell(me.ID, spree, game.CastSpellParams{Targets: effectstest.B16TargetCard(a), OptionalCosts: []int{0}, Strict: true}); err != nil {
		t.Fatalf("cast with one replicate: %v", err)
	}
	if it := effectstest.TriggerOnStack(g, spree); it == nil || it.XValue != 1 {
		t.Fatalf("replicate trigger = %+v, want one counting one payment", it)
	}
	for i := 0; i < 64; i++ {
		if effectstest.StackFullyEmpty(g) && effectstest.LatestPickTarget(g, me.ID) == nil {
			break
		}
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: b}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Errorf("both artifacts should be destroyed: A %v, B %v", g.Battlefield.Contains(a), g.Battlefield.Contains(b))
	}
}

// --- Lita, Mechanical Engineer --------------------------------------------------------------------------------------

func TestAvatarB24LitaUntapsOtherArtifactCreaturesAndMakesAZeppelin(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lita := effectstest.PushCatalogPermanent(g, me.ID, "Lita, Mechanical Engineer", "Legendary Artifact Creature — Artificer", avB24LitaOracle, false)
	golem := effectstest.B12Creature(g, me.ID, "Golem", "Artifact Creature — Golem", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{W}")
	if err := g.ActivateCatalogAbility(me.ID, lita, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var zeppelin game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Zeppelin" && c.Controller == me.ID {
			zeppelin = c
		}
	}
	if zeppelin.InstanceID == uuid.Nil || zeppelin.Power != 5 || zeppelin.Toughness != 5 || !game.HasKeyword(&zeppelin, "flying") ||
		!zeppelin.HasSubtype("Vehicle") || zeppelin.IsCreature() {
		t.Fatalf("Zeppelin = %+v, want a 5/5 flying noncreature Vehicle", zeppelin)
	}
	effectstest.TapForTest(t, g, golem, bear)
	effectstest.MkAdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IsTapped(g, golem) {
		t.Error("the artifact creature stayed tapped")
	}
	if !effectstest.IsTapped(g, bear) {
		t.Error("a nonartifact creature was untapped")
	}
	if !effectstest.IsTapped(g, lita) {
		t.Error("Lita untapped herself: \"each other\"")
	}
}

// --- Jasmine Dragon Tea Shop ---------------------------------------------------------------------------------------------

func TestAvatarB24JasmineDragonTeaShopManaIsForAllies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	shop := effectstest.PushCatalogPermanent(g, me.ID, "Jasmine Dragon Tea Shop", "Land", avB24TeaShopOracle, false)
	if err := g.ActivateManaAbility(me.ID, shop, 1, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("ActivateManaAbility any colour: %v", err)
	}
	cost, _ := game.ParseCost("{G}")
	ally := game.Card{Name: "Test Ally", TypeLine: "Creature — Human Ally", ManaCost: "{G}"}
	bear := game.Card{Name: "Test Bear", TypeLine: "Creature — Bear", ManaCost: "{G}"}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(ally)) || !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(ally)) {
		t.Error("the mana was refused for an Ally spell or ability")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(bear)) || me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(bear)) {
		t.Error("the mana paid for a non-Ally: stronger than printed")
	}
	me.ManaPool = game.ManaPool{}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(shop) })
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, shop, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if avB24Count(g, me.ID, "Ally") != 1 {
		t.Error("no Ally token")
	}
}

// --- Fire Nation Palace ------------------------------------------------------------------------------------------------------

func TestAvatarB24FireNationPalaceGrantsFirebendingFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	palace := effectstest.PushCatalogPermanent(g, me.ID, "Fire Nation Palace", "Land", avB24PalaceOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{R}")
	if err := g.ActivateCatalogAbility(me.ID, palace, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(theirs), Strict: true}); err == nil {
		t.Fatal("targeted a creature an opponent controls")
	}
	if err := g.ActivateCatalogAbility(me.ID, palace, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MkWantPool(t, me, map[string]int{"R": 4})
}

func TestAvatarB24FireNationPalaceEntersTappedWithoutABasic(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PlayLandFromHand(t, g, "Fire Nation Palace", avB24PalaceOracle)
	effectstest.Top100AssertEnteredTapped(t, g, id, "Fire Nation Palace")
}

// --- Momo, Friendly Flier ------------------------------------------------------------------------------------------------------

func TestAvatarB24MomoDiscountsTheFirstNonLemurFlierEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	momo := effectstest.PushCatalogPermanent(g, me.ID, "Momo, Friendly Flier", "Legendary Creature — Lemur Bat Ally", avB24MomoOracle, false)
	effectstest.AdvanceToMain(t, g)
	lemur := effectstest.AvB24Hand(g, me, game.Card{Name: "Flying Lemur", TypeLine: "Creature — Lemur", ManaCost: "{1}{W}", Keywords: []string{"flying"}, Power: 1, Toughness: 1})
	bird := effectstest.AvB24Hand(g, me, game.Card{Name: "Bird", TypeLine: "Creature — Bird", ManaCost: "{1}{W}", Keywords: []string{"flying"}, Power: 1, Toughness: 1})
	bird2 := effectstest.AvB24Hand(g, me, game.Card{Name: "Bird Two", TypeLine: "Creature — Bird", ManaCost: "{1}{W}", Keywords: []string{"flying"}, Power: 1, Toughness: 1})

	// A Lemur is not discounted and does not use the discount up.
	effectstest.FloatMana(t, g, me, "{W}")
	if err := g.CastSpell(me.ID, lemur, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("a Lemur flier was discounted")
	}
	effectstest.FloatMana(t, g, me, "{W}")
	if err := g.CastSpell(me.ID, lemur, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("Lemur at full price: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, momo); !ok || c.CurrentPower() != 2 {
		t.Errorf("Momo did not get +1/+1 when another flier entered: %+v", c)
	}

	effectstest.FloatMana(t, g, me, "{W}")
	if err := g.CastSpell(me.ID, bird, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the first non-Lemur flier for {W}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FloatMana(t, g, me, "{W}")
	if err := g.CastSpell(me.ID, bird2, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("the second non-Lemur flier was discounted too")
	}
}

// --- Toph, Hardheaded Teacher ------------------------------------------------------------------------------------------------------

func TestAvatarB24TophDiscardsToReturnAnInstant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bolt := effectstest.AvB24Graveyard(me, game.Card{Name: "Old Bolt", TypeLine: "Instant"})
	junk := effectstest.AvB24Hand(g, me, game.Card{Name: "Junk", TypeLine: "Creature — Bear"})
	effectstest.CastCatalogSpell(t, g, "Toph, Hardheaded Teacher", "Legendary Creature — Human Warrior Ally", avB24TophOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bolt}); err != nil {
			t.Fatalf("pick: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.AnswerDiscard(t, g, me.ID, junk)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bolt) || !me.Graveyard.Contains(junk) {
		t.Errorf("bolt in hand %v, junk in graveyard %v", me.Hand.Contains(bolt), me.Graveyard.Contains(junk))
	}
}

func TestAvatarB24TophDeclinesTheDiscardAndReturnsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bolt := effectstest.AvB24Graveyard(me, game.Card{Name: "Old Bolt", TypeLine: "Instant"})
	effectstest.CastCatalogSpell(t, g, "Toph, Hardheaded Teacher", "Legendary Creature — Human Warrior Ally", avB24TophOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bolt}); err != nil {
			t.Fatalf("pick: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.AnswerDiscard(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(bolt) {
		t.Error("returned the instant without a discard")
	}
}

func TestAvatarB24TophEarthbendsOnEachSpellAndTwiceForALesson(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Toph, Hardheaded Teacher", "Legendary Creature — Human Warrior Ally", avB24TophOracle, false)
	forest := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	mountain := effectstest.B12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	effectstest.AdvanceToMain(t, g)

	cast := func(name, typeLine string, land uuid.UUID) {
		t.Helper()
		id := effectstest.AvB24Hand(g, me, game.Card{Name: name, TypeLine: typeLine})
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast %s: %v", name, err)
		}
		for i := 0; i < 8; i++ {
			p := effectstest.LatestPickTarget(g, me.ID)
			if p == nil {
				break
			}
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: land}); err != nil {
				t.Fatalf("pick land: %v", err)
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	cast("Plain Sorcery", "Sorcery", forest)
	if got := effectstest.CountersOn(g, forest, game.CounterPlusOne); got != 1 {
		t.Errorf("a plain spell: %d counters, want 1", got)
	}
	cast("A Lesson", "Sorcery — Lesson", mountain)
	if got := effectstest.CountersOn(g, mountain, game.CounterPlusOne); got != 2 {
		t.Errorf("a Lesson: %d counters, want 2", got)
	}
	if c, ok := effectstest.BattlefieldCard(g, mountain); !ok || !c.IsCreature() {
		t.Error("the land is not a creature")
	}
}
