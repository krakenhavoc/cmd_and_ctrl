package effects

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// single_graveyard_cards_test.go — #1807 (ADR 0106 §5), the cards that
// target "from a single graveyard". The engine half (the sameness rule
// and its four readers) is pinned in game/target_sameness_test.go;
// these are the cards on top of it.

const (
	oracleDigsiteConservator = "ced2ae56-988a-4c60-93d0-85a2261be72c"
	oracleCarrionBeetles     = "d64a51be-2dd1-485f-98fa-c93d4c296e68"
	oracleRagDealer          = "493acba2-92b0-43a4-9f4b-d49a91ca77f5"
	oracleDecompose          = "12c2b6bb-05f8-4291-b9a5-c4ddb7a69b9e"
	oracleFamishedGhoul      = "22c74e4b-b1fe-4a35-acd8-fa4840af3c06"
	oracleGriffnautTracker   = "6eff5e17-946b-4433-9f48-88f103844c42"
	oracleScarabFeast        = "8b9043d3-a1c6-4f49-9c49-ef78bbfbd4ac"
	oracleRapidDecay         = "2c769dfc-79be-4fcb-840c-b72cafdd0a23"
	oracleShredMemory        = "16cae855-d93a-445d-9cac-dbc5a7c18cb1"
	oracleArashinSunshield   = "1175d482-d8a2-467f-bc50-2f4b241966bb"
	oracleGravegouger        = "2677c0d9-1251-478e-b157-072859c593a9"
	oracleSoulShackledZombie = "f81ca3a5-356d-4b5d-afc2-a33aa774816b"
	oracleEbonyCharm         = "2f3237fe-9e07-4aad-8bc8-7ef4b510dd36"
	oracleRatsFeast          = "327c5ca2-4eeb-49cb-84fb-f6a4b06b24bc"
	oracleSereneRemembrance  = "f980c8b4-4cd7-42d2-82d5-96e3275b2647"
	oracleUnlicensedHearse   = "c640654c-487e-4a2c-aced-126ed835b78f"
	oracleWasteManagement    = "8d4e0866-d8f5-4eb6-a0fd-3fa9d4b9cf4a"
	oracleKayaOrzhovUsurper  = "7dd4a1a1-d5f4-4ac7-a9f6-34af411f070b"
	oracleCeaseDesist        = "ba4d644f-1931-4fc4-aed5-681a476a5a58"
	oraclePushPull           = "7e050495-bed7-43b9-abce-866c61beb1da"
)

func TestSingleGraveyardCardsAreRegistered(t *testing.T) {
	for _, tc := range []struct {
		oracle, name string
		full         bool
	}{
		{oracleDigsiteConservator, "Digsite Conservator", true},
		{oracleCarrionBeetles, "Carrion Beetles", true},
		{oracleRagDealer, "Rag Dealer", true},
		{oracleDecompose, "Decompose", true},
		{oracleFamishedGhoul, "Famished Ghoul", true},
		{oracleGriffnautTracker, "Griffnaut Tracker", true},
		{oracleScarabFeast, "Scarab Feast", true},
		{oracleRapidDecay, "Rapid Decay", true},
		{oracleShredMemory, "Shred Memory", true},
		{oracleArashinSunshield, "Arashin Sunshield", true},
		{oracleGravegouger, "Gravegouger", true},
		{oracleSoulShackledZombie, "Soul-Shackled Zombie", true},
		{oracleEbonyCharm, "Ebony Charm", true},
		{oracleRatsFeast, "Rats' Feast", true},
		{oracleSereneRemembrance, "Serene Remembrance", true},
		{oracleUnlicensedHearse, "Unlicensed Hearse", true},
		{oracleWasteManagement, "Waste Management", true},
		{oracleKayaOrzhovUsurper, "Kaya, Orzhov Usurper", true},
		{oracleCeaseDesist, "Cease", true},
		{oracleCeaseDesist + "#1", "Desist", true},
		{oraclePushPull, "Push", true},
		{oraclePushPull + "#1", "Pull", false},
		{pestilentCauldronOracleID, "Pestilent Cauldron", true},
		{pestilentCauldronOracleID + "#1", "Restorative Burst", true},
	} {
		spec, ok := Lookup(tc.oracle)
		if !ok || spec.Name != tc.name {
			t.Errorf("%s: registered = %v as %q", tc.name, ok, spec.Name)
			continue
		}
		if full := spec.Completeness == CompletenessFull; full != tc.full {
			t.Errorf("%s: Completeness = %s, want full=%v", tc.name, spec.Completeness, tc.full)
		}
	}
}

func sgCards(ids ...uuid.UUID) []game.TargetRef {
	out := make([]game.TargetRef, 0, len(ids))
	for _, id := range ids {
		out = append(out, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	return out
}

func sgInExile(g *game.Game, ids ...uuid.UUID) bool {
	for _, id := range ids {
		if !g.Exile.Contains(id) {
			return false
		}
	}
	return true
}

// --- spells -------------------------------------------------------

// Decompose: three from one graveyard is legal and all three go; one
// from each of two graveyards is refused with the rule's sentence.
func TestDecomposeExilesFromOneGraveyardOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := batch01GraveyardCard(opp, "A", "Creature — Bear")
	b := batch01GraveyardCard(opp, "B", "Sorcery")
	c := batch01GraveyardCard(opp, "C", "Instant")
	mine := batch01GraveyardCard(me, "Mine", "Creature — Bear")

	err := castCatalogSpellErr(t, g, "Decompose", "Sorcery", oracleDecompose, sgCards(a, mine))
	if !errors.Is(err, game.ErrIllegalTarget) || !strings.Contains(err.Error(), "come from a single graveyard") {
		t.Fatalf("two graveyards: err = %v, want the single-graveyard refusal", err)
	}

	castCatalogSpell(t, g, "Decompose", "Sorcery", oracleDecompose, sgCards(a, b, c))
	passPriorityAroundTable(t, g)
	if !sgInExile(g, a, b, c) {
		t.Error("all three cards from the one graveyard are exiled")
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("the other graveyard is untouched")
	}
}

// A card that leaves in response is skipped; the rest still go
// (CR 608.2b).
func TestScarabFeastSkipsATargetThatLeft(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := batch01GraveyardCard(opp, "A", "Creature — Bear")
	b := batch01GraveyardCard(opp, "B", "Sorcery")
	castCatalogSpell(t, g, "Scarab Feast", "Instant", oracleScarabFeast, sgCards(a, b))
	g.WithWriteLock(func() { opp.Graveyard.Remove(b) })
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(a) {
		t.Error("the survivor is exiled")
	}
}

// Rats' Feast: X targets, all from one graveyard.
func TestRatsFeastExilesXFromOneGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := batch01GraveyardCard(opp, "A", "Sorcery")
	b := batch01GraveyardCard(opp, "B", "Sorcery")
	mine := batch01GraveyardCard(me, "Mine", "Sorcery")
	b10AddMana(me, "B", "B", "B")
	if err := castX(t, g, "Rats' Feast", "Sorcery", oracleRatsFeast, 2, sgCards(a, mine)); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("X=2 across two graveyards: err = %v, want ErrIllegalTarget", err)
	}
	if err := castX(t, g, "Rats' Feast", "Sorcery", oracleRatsFeast, 2, sgCards(a, b)); err != nil {
		t.Fatalf("X=2 from one graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !sgInExile(g, a, b) {
		t.Error("both cards are exiled")
	}
}

// Serene Remembrance goes into its owner's library with its targets,
// each into its owner's.
func TestSereneRemembranceShufflesItselfAndTheTargetsIn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := batch01GraveyardCard(opp, "A", "Creature — Bear")
	b := batch01GraveyardCard(opp, "B", "Sorcery")
	spell := castCatalogSpell(t, g, "Serene Remembrance", "Sorcery", oracleSereneRemembrance, sgCards(a, b))
	passPriorityAroundTable(t, g)
	if !opp.Library.Contains(a) || !opp.Library.Contains(b) {
		t.Error("the targets are in their owner's library")
	}
	if !me.Library.Contains(spell) {
		t.Errorf("Serene Remembrance is in its owner's library, not the graveyard (graveyard has it: %v)", me.Graveyard.Contains(spell))
	}
}

// Shred Memory's transmute finds a card with the same mana value (2),
// at sorcery speed, from the hand.
func TestShredMemoryTransmutesForAManaValueTwoCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	shred := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: shred, Name: "Shred Memory", TypeLine: "Instant", ManaCost: "{1}{B}",
		OracleID: oracleShredMemory, Owner: me.ID, Controller: me.ID,
	})
	two := pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Two Drop", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Owner: me.ID, Controller: me.ID})
	pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Three Drop", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Owner: me.ID, Controller: me.ID})
	b10AddMana(me, "B", "B", "B")
	if err := g.ActivateCatalogAbility(me.ID, shred, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("transmute: %v", err)
	}
	if !me.Graveyard.Contains(shred) {
		t.Fatal("Shred Memory is discarded to pay for transmute")
	}
	passPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceSearchLibrary {
			if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{two}); err != nil {
				t.Fatalf("ResolveSearchLibrary: %v", err)
			}
		}
	}
	if !me.Hand.Contains(two) {
		t.Error("the mana value 2 card is in hand")
	}
}

// Waste Management unkicked: up to two from one graveyard, a Rogue per
// creature card exiled. Kicked: the whole of one player's graveyard.
func TestWasteManagementMakesARoguePerCreatureCardOnEitherBranch(t *testing.T) {
	t.Run("unkicked", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		bear := batch01GraveyardCard(opp, "Bear", "Creature — Bear")
		sorc := batch01GraveyardCard(opp, "Sorc", "Sorcery")
		castCatalogSpell(t, g, "Waste Management", "Instant", oracleWasteManagement, sgCards(bear, sorc))
		passPriorityAroundTable(t, g)
		if !sgInExile(g, bear, sorc) {
			t.Fatal("both cards are exiled")
		}
		if n := onBattlefieldNamed(g, "Rogue"); n != 1 {
			t.Errorf("Rogues = %d, want 1 for the one creature card", n)
		}
		_ = me
	})
	t.Run("kicked", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		for i := 0; i < 3; i++ {
			batch01GraveyardCard(opp, "Bear", "Creature — Bear")
		}
		batch01GraveyardCard(opp, "Sorc", "Sorcery")
		advanceToMain(t, g)
		id := uuid.New()
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Waste Management", TypeLine: "Instant",
			OracleID: oracleWasteManagement, Owner: me.ID, Controller: me.ID,
		})
		b10AddMana(me, "B", "B", "B", "B")
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{
			OptionalCosts: []int{0},
			Targets:       []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		}); err != nil {
			t.Fatalf("kicked cast: %v", err)
		}
		passPriorityAroundTable(t, g)
		if opp.Graveyard.Size() != 0 {
			t.Errorf("the target player's graveyard is exiled, %d cards left", opp.Graveyard.Size())
		}
		if n := onBattlefieldNamed(g, "Rogue"); n != 3 {
			t.Errorf("Rogues = %d, want 3", n)
		}
	})
}

// Ebony Charm's second mode is the family's clause on a bullet.
func TestEbonyCharmSecondModeExilesFromOneGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := batch01GraveyardCard(opp, "A", "Sorcery")
	mine := batch01GraveyardCard(me, "Mine", "Sorcery")
	advanceToMain(t, g)
	cast := func(targets ...uuid.UUID) error {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Ebony Charm", TypeLine: "Instant", OracleID: oracleEbonyCharm, Owner: me.ID, Controller: me.ID})
		return g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{1}, Targets: sgCards(targets...)})
	}
	if err := cast(a, mine); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("two graveyards on the mode: err = %v, want ErrIllegalTarget", err)
	}
	if err := cast(a); err != nil {
		t.Fatalf("one graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(a) || g.Exile.Contains(mine) {
		t.Error("the chosen card is exiled and nothing else")
	}
}

// Cease: the cards from one graveyard, then any player gains 2 and
// draws.
func TestCeaseExilesThenTheTargetPlayerGainsAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := batch01GraveyardCard(opp, "A", "Sorcery")
	advanceToMain(t, g)
	card := mdfcCard(me.ID, oracleCeaseDesist, game.LayoutSplit,
		game.Face{Name: "Cease", TypeLine: "Instant", ManaCost: "{0}"},
		game.Face{Name: "Desist", TypeLine: "Sorcery", ManaCost: "{0}"},
	)
	me.Hand.PushTop(card)
	life, hand := me.Life, me.Hand.Size()
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: a},
		{Kind: game.TargetPlayer, ID: me.ID, Slot: 1},
	}}); err != nil {
		t.Fatalf("cast Cease: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(a) {
		t.Error("the card is exiled")
	}
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2", life, me.Life)
	}
	// The hand lost Cease and drew one.
	if me.Hand.Size() != hand {
		t.Errorf("hand %d → %d, want the same (Cease left, a card was drawn)", hand, me.Hand.Size())
	}
}

// Pull: two creature cards from one graveyard enter under the caster's
// control with haste, and are sacrificed at the next end step.
func TestPullReanimatesFromOneGraveyardWithHasteThenSacrifices(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := batch01GraveyardCard(opp, "Bear A", "Creature — Bear")
	b := batch01GraveyardCard(opp, "Bear B", "Creature — Bear")
	mine := batch01GraveyardCard(me, "My Bear", "Creature — Bear")
	advanceToMain(t, g)
	cast := func(targets ...uuid.UUID) error {
		card := mdfcCard(me.ID, oraclePushPull, game.LayoutSplit,
			game.Face{Name: "Push", TypeLine: "Sorcery", ManaCost: "{0}"},
			game.Face{Name: "Pull", TypeLine: "Sorcery", ManaCost: "{0}"},
		)
		me.Hand.PushTop(card)
		return g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, Targets: sgCards(targets...)})
	}
	if err := cast(a, mine); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("two graveyards: err = %v, want ErrIllegalTarget", err)
	}
	if err := cast(a, b); err != nil {
		t.Fatalf("cast Pull: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		c, ok := g.LookupCardForEffect(id)
		if !ok || !onBattlefield(g, id) || c.Controller != me.ID {
			t.Fatalf("%s is on the battlefield under the caster's control", id)
		}
		if !bearHasKeyword(t, g, id, "haste") {
			t.Errorf("%s has haste", c.Name)
		}
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if onBattlefield(g, a) || onBattlefield(g, b) {
		t.Error("both are sacrificed at the end step")
	}
	if !opp.Graveyard.Contains(a) || !opp.Graveyard.Contains(b) {
		t.Error("sacrificed, they go to their owner's graveyard")
	}
}

// Push destroys only a TAPPED creature.
func TestPushDestroysATappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	untapped := b12Creature(g, opp.ID, "Untapped", "Creature — Bear", 2, 2)
	tapped := b12Creature(g, opp.ID, "Tapped", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(tapped) })
	advanceToMain(t, g)
	cast := func(target uuid.UUID) error {
		card := mdfcCard(me.ID, oraclePushPull, game.LayoutSplit,
			game.Face{Name: "Push", TypeLine: "Sorcery", ManaCost: "{0}"},
			game.Face{Name: "Pull", TypeLine: "Sorcery", ManaCost: "{0}"},
		)
		me.Hand.PushTop(card)
		return g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Targets: sgCards(target)})
	}
	if err := cast(untapped); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("an untapped creature: err = %v, want ErrIllegalTarget", err)
	}
	if err := cast(tapped); err != nil {
		t.Fatalf("cast Push: %v", err)
	}
	passPriorityAroundTable(t, g)
	if onBattlefield(g, tapped) {
		t.Error("the tapped creature is destroyed")
	}
}

// Restorative Burst: up to two cards back to hand, everyone gains 4,
// and it exiles itself.
func TestRestorativeBurstReturnsGivesLifeAndExilesItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := batch01GraveyardCard(me, "Forest", "Basic Land — Forest")
	advanceToMain(t, g)
	card := mdfcCard(me.ID, pestilentCauldronOracleID, game.LayoutModalDFC,
		game.Face{Name: "Pestilent Cauldron", TypeLine: "Artifact", ManaCost: "{0}"},
		game.Face{Name: "Restorative Burst", TypeLine: "Sorcery", ManaCost: "{0}"},
	)
	me.Hand.PushTop(card)
	myLife, oppLife := me.Life, opp.Life
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, Targets: sgCards(land)}); err != nil {
		t.Fatalf("cast Restorative Burst: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(land) {
		t.Error("the land card returns to hand")
	}
	if me.Life != myLife+4 || opp.Life != oppLife+4 {
		t.Errorf("life: me %d → %d, opponent %d → %d, want +4 each", myLife, me.Life, oppLife, opp.Life)
	}
	if !g.Exile.Contains(card.InstanceID) {
		t.Error("Restorative Burst exiles itself")
	}
}

// --- activated abilities -----------------------------------------

func TestCarrionBeetlesAndRagDealerExileUpToThree(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Carrion Beetles", oracleCarrionBeetles},
		{"Rag Dealer", oracleRagDealer},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			src := pushCatalogPermanent(g, me.ID, tc.name, "Creature — Insect", tc.oracle, false)
			a := batch01GraveyardCard(opp, "A", "Sorcery")
			b := batch01GraveyardCard(opp, "B", "Sorcery")
			mine := batch01GraveyardCard(me, "Mine", "Sorcery")
			advanceToMain(t, g)
			b10AddMana(me, "B", "B", "B")
			err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{Targets: sgCards(a, mine)})
			if !errors.Is(err, game.ErrIllegalTarget) {
				t.Fatalf("two graveyards: err = %v, want ErrIllegalTarget", err)
			}
			if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{Targets: sgCards(a, b)}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			passPriorityAroundTable(t, g)
			if !sgInExile(g, a, b) {
				t.Error("both are exiled")
			}
		})
	}
}

// Famished Ghoul is sacrificed as the cost and its ability still
// resolves (CR 113.7a).
func TestFamishedGhoulResolvesAfterItsSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ghoul := pushCatalogPermanent(g, me.ID, "Famished Ghoul", "Creature — Zombie", oracleFamishedGhoul, false)
	a := batch01GraveyardCard(opp, "A", "Sorcery")
	b := batch01GraveyardCard(opp, "B", "Sorcery")
	advanceToMain(t, g)
	b10AddMana(me, "B", "B")
	if err := g.ActivateCatalogAbility(me.ID, ghoul, 0, game.ActivateAbilityParams{Targets: sgCards(a, b)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if onBattlefield(g, ghoul) {
		t.Fatal("the Ghoul is sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	if !sgInExile(g, a, b) {
		t.Error("both are exiled")
	}
}

// Digsite Conservator: the sorcery-speed sacrifice exiles four from one
// graveyard; dying asks for {4}, and declining discovers nothing.
func TestDigsiteConservatorExilesFourFromOneGraveyardAndAsksToDiscover(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gnome := pushCatalogPermanent(g, me.ID, "Digsite Conservator", "Artifact Creature — Gnome", oracleDigsiteConservator, false)
	var four []uuid.UUID
	for i := 0; i < 4; i++ {
		four = append(four, batch01GraveyardCard(opp, "Card", "Sorcery"))
	}
	mine := batch01GraveyardCard(me, "Mine", "Sorcery")
	advanceToMain(t, g)
	err := g.ActivateCatalogAbility(me.ID, gnome, 0, game.ActivateAbilityParams{Targets: sgCards(four[0], four[1], four[2], mine)})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("two graveyards: err = %v, want ErrIllegalTarget", err)
	}
	if !onBattlefield(g, gnome) {
		t.Fatal("a refused activation pays nothing")
	}
	if err := g.ActivateCatalogAbility(me.ID, gnome, 0, game.ActivateAbilityParams{Targets: sgCards(four...)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	ask := answerPayUnless(t, g, me.ID, false)
	if ask.PayCost != "{4}" {
		t.Errorf("the dies trigger asks for {4}, got %q", ask.PayCost)
	}
	passPriorityAroundTable(t, g)
	if !sgInExile(g, four...) {
		t.Error("all four are exiled")
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("the other graveyard is untouched")
	}
}

// The sacrifice is sorcery-speed only (CR 602.5d).
func TestDigsiteConservatorActivatesOnlyAsASorcery(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gnome := pushCatalogPermanent(g, me.ID, "Digsite Conservator", "Artifact Creature — Gnome", oracleDigsiteConservator, false)
	a := batch01GraveyardCard(opp, "A", "Sorcery")
	// The harness parks on the turn-1 draw step: not a main phase.
	if err := g.ActivateCatalogAbility(me.ID, gnome, 0, game.ActivateAbilityParams{Targets: sgCards(a)}); err == nil {
		t.Fatal("activated outside a main phase")
	}
}

// Pestilent Cauldron's exact four: four from one graveyard exiles and
// draws; four split over two graveyards is refused.
func TestPestilentCauldronNeedsFourFromOneGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cauldron := pushCatalogPermanent(g, me.ID, "Pestilent Cauldron", "Artifact", pestilentCauldronOracleID, false)
	var theirs []uuid.UUID
	for i := 0; i < 4; i++ {
		theirs = append(theirs, batch01GraveyardCard(opp, "Card", "Sorcery"))
	}
	mine := batch01GraveyardCard(me, "Mine", "Sorcery")
	advanceToMain(t, g)
	b10AddMana(me, "B", "B", "B", "B")
	err := g.ActivateCatalogAbility(me.ID, cauldron, 2, game.ActivateAbilityParams{Targets: sgCards(theirs[0], theirs[1], theirs[2], mine)})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("four over two graveyards: err = %v, want ErrIllegalTarget", err)
	}
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, cauldron, 2, game.ActivateAbilityParams{Targets: sgCards(theirs...)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !sgInExile(g, theirs...) {
		t.Error("all four are exiled")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want a card drawn", hand, me.Hand.Size())
	}
}

// Unlicensed Hearse is as big as the number of cards it has exiled.
func TestUnlicensedHearseCountsTheCardsExiledWithIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hearse := pushCatalogPermanent(g, me.ID, "Unlicensed Hearse", "Artifact — Vehicle", oracleUnlicensedHearse, false)
	a := batch01GraveyardCard(opp, "A", "Sorcery")
	b := batch01GraveyardCard(opp, "B", "Sorcery")
	advanceToMain(t, g)
	if p := effectivePower(t, g, hearse); p != 0 {
		t.Errorf("power before anything is exiled = %d, want 0", p)
	}
	if err := g.ActivateCatalogAbility(me.ID, hearse, 0, game.ActivateAbilityParams{Targets: sgCards(a, b)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, hearse), effectiveToughness(t, g, hearse); p != 2 || tough != 2 {
		t.Errorf("P/T = %d/%d, want 2/2", p, tough)
	}
	// A card that leaves exile no longer counts (CR 607.2a).
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(a) })
	if p := effectivePower(t, g, hearse); p != 1 {
		t.Errorf("power after one card left exile = %d, want 1", p)
	}
}

// Kaya's +1 gains 2 only when a creature card was exiled; the −5 deals
// and gains the number of cards the player owns in exile.
func TestKayaOrzhovUsurper(t *testing.T) {
	t.Run("+1 with a creature card", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		kaya := pushCatalogWalker(g, me.ID, "Kaya, Orzhov Usurper", oracleKayaOrzhovUsurper, 3)
		bear := batch01GraveyardCard(opp, "Bear", "Creature — Bear")
		advanceToMain(t, g)
		life := me.Life
		if err := g.ActivateCatalogAbility(me.ID, kaya, 0, game.ActivateAbilityParams{Targets: sgCards(bear)}); err != nil {
			t.Fatalf("+1: %v", err)
		}
		passPriorityAroundTable(t, g)
		if me.Life != life+2 {
			t.Errorf("life %d → %d, want +2", life, me.Life)
		}
	})
	t.Run("+1 without one", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		kaya := pushCatalogWalker(g, me.ID, "Kaya, Orzhov Usurper", oracleKayaOrzhovUsurper, 3)
		sorc := batch01GraveyardCard(opp, "Sorc", "Sorcery")
		advanceToMain(t, g)
		life := me.Life
		if err := g.ActivateCatalogAbility(me.ID, kaya, 0, game.ActivateAbilityParams{Targets: sgCards(sorc)}); err != nil {
			t.Fatalf("+1: %v", err)
		}
		passPriorityAroundTable(t, g)
		if !g.Exile.Contains(sorc) || me.Life != life {
			t.Errorf("the sorcery is exiled and no life is gained (life %d → %d)", life, me.Life)
		}
	})
	t.Run("−5", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		kaya := pushCatalogWalker(g, me.ID, "Kaya, Orzhov Usurper", oracleKayaOrzhovUsurper, 6)
		for i := 0; i < 3; i++ {
			g.Exile.PushTop(game.Card{InstanceID: uuid.New(), Name: "Exiled", Owner: opp.ID, Controller: opp.ID})
		}
		g.Exile.PushTop(game.Card{InstanceID: uuid.New(), Name: "Mine", Owner: me.ID, Controller: me.ID})
		advanceToMain(t, g)
		myLife, oppLife := me.Life, opp.Life
		if err := g.ActivateCatalogAbility(me.ID, kaya, 2, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		}); err != nil {
			t.Fatalf("−5: %v", err)
		}
		passPriorityAroundTable(t, g)
		if opp.Life != oppLife-3 || me.Life != myLife+3 {
			t.Errorf("opponent %d → %d (want −3), me %d → %d (want +3)", oppLife, opp.Life, myLife, me.Life)
		}
	})
}

// --- entry triggers -----------------------------------------------

// sgAnswerTargets answers the open pick_target prompt with refs and
// returns the error.
func sgAnswerTargets(t *testing.T, g *game.Game, ids ...uuid.UUID) error {
	t.Helper()
	ch := pickTargetPrompt(t, g)
	return g.ResolvePickTargets(ch.ID, ch.Chooser, sgCards(ids...))
}

// Griffnaut Tracker's entry trigger refuses two graveyards, then
// exiles two from one.
func TestGriffnautTrackerEntryExilesFromOneGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := batch01GraveyardCard(opp, "A", "Sorcery")
	b := batch01GraveyardCard(opp, "B", "Sorcery")
	mine := batch01GraveyardCard(me, "Mine", "Sorcery")
	castHoldCreature(t, g, "Griffnaut Tracker", "Creature — Human Detective", oracleGriffnautTracker, 2, 3)
	if err := sgAnswerTargets(t, g, a, mine); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("two graveyards: err = %v, want ErrIllegalTarget", err)
	}
	if err := sgAnswerTargets(t, g, a, b); err != nil {
		t.Fatalf("one graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !sgInExile(g, a, b) || !me.Graveyard.Contains(mine) {
		t.Error("the two picked cards are exiled and nothing else")
	}
}

// Soul-Shackled Zombie drains only when a creature card was exiled.
func TestSoulShackledZombieDrainsOnlyForACreatureCard(t *testing.T) {
	for _, tc := range []struct {
		name     string
		typeLine string
		drain    bool
	}{
		{"a creature card", "Creature — Bear", true},
		{"no creature card", "Sorcery", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			card := batch01GraveyardCard(opp, "Card", tc.typeLine)
			myLife, oppLife := me.Life, opp.Life
			castHoldCreature(t, g, "Soul-Shackled Zombie", "Creature — Zombie", oracleSoulShackledZombie, 4, 2)
			if err := sgAnswerTargets(t, g, card); err != nil {
				t.Fatalf("answer: %v", err)
			}
			passPriorityAroundTable(t, g)
			drained := opp.Life == oppLife-2 && me.Life == myLife+2
			if drained != tc.drain {
				t.Errorf("drained = %v (me %d → %d, opponent %d → %d), want %v", drained, myLife, me.Life, oppLife, opp.Life, tc.drain)
			}
		})
	}
}

// Gravegouger returns the cards it exiled when it leaves.
func TestGravegougerReturnsTheExiledCardsWhenItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := batch01GraveyardCard(opp, "A", "Sorcery")
	b := batch01GraveyardCard(opp, "B", "Sorcery")
	gouger := castHoldCreature(t, g, "Gravegouger", "Creature — Nightmare Horror", oracleGravegouger, 3, 3)
	if err := sgAnswerTargets(t, g, a, b); err != nil {
		t.Fatalf("answer: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !sgInExile(g, a, b) {
		t.Fatal("both are exiled")
	}
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(gouger) })
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(a) || !opp.Graveyard.Contains(b) {
		t.Error("the exiled cards are back in their owner's graveyard")
	}
}

// Arashin Sunshield's tapper.
func TestArashinSunshieldTapsATargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shield := pushCatalogPermanent(g, me.ID, "Arashin Sunshield", "Creature — Human Warrior", oracleArashinSunshield, false)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	b10AddMana(me, "W")
	if err := g.ActivateCatalogAbility(me.ID, shield, 0, game.ActivateAbilityParams{Targets: sgCards(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	c, _ := g.LookupCardForEffect(bear)
	if !c.Tapped {
		t.Error("the creature is tapped")
	}
}
