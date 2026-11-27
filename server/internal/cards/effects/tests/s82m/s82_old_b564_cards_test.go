package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b564_cards_test.go — Coldsnap's recover cards, played from the real
// catalog entries (S82 OLD batch 564).

const (
	b564HarvestOracle    = "231ca538-b277-4c12-979f-57640a0e87f0"
	b564ResizeOracle     = "2c092ed4-4dbb-41ec-9e07-e59b9676802e"
	b564IcefallOracle    = "30baca42-50e2-45f8-8f99-8e2f9795b4da"
	b564BountyOracle     = "5c15d9b1-5ec7-4476-a3e6-539d899a9957"
	b564AssassinOracle   = "9660b7c4-7b04-4ec3-b867-299c141f77e6"
	b564ControvertOracle = "f7ad79e3-d128-4aee-8542-99fd3527f9d3"
	b564RotOracle        = "f9a272c5-6ebc-4d29-a47c-a2d570184196"
)

func TestS82OldB564CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b564HarvestOracle:    "Grim Harvest",
		b564ResizeOracle:     "Resize",
		b564IcefallOracle:    "Icefall",
		b564BountyOracle:     "Sun's Bounty",
		b564AssassinOracle:   "Garza's Assassin",
		b564ControvertOracle: "Controvert",
		b564RotOracle:        "Krovikan Rot",
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

// --- the spells -------------------------------------------------------------

func TestGrimHarvestReturnsACreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dead := effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	land := effectstest.PushGraveyardCardTyped(me, "Dead Land", "Land")
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Grim Harvest", "Instant", b564HarvestOracle, effectstest.CardTarget(land)); err == nil {
		t.Fatal("targeted a land card: accepted")
	}
	effectstest.CastCatalogSpell(t, g, "Grim Harvest", "Instant", b564HarvestOracle, effectstest.CardTarget(dead))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HandHasID(me, dead) {
		t.Error("the creature card is not in hand")
	}
}

func TestResizeGivesPlusThreePlusThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Resize", "Instant", b564ResizeOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("the Bear is %d/%d, want 5/5", p, tough)
	}
}

func TestIcefallDestroysAnArtifactOrALandOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	land := effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Icefall", "Sorcery", b564IcefallOracle, effectstest.CardTarget(bear)); err == nil {
		t.Fatal("targeted a creature: accepted")
	}
	for _, id := range []uuid.UUID{rock, land} {
		effectstest.CastCatalogSpell(t, g, "Icefall", "Sorcery", b564IcefallOracle, effectstest.CardTarget(id))
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(id) {
			t.Errorf("%s survived", id)
		}
	}
	_ = me
}

func TestSunsBountyGainsFourLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Sun's Bounty", "Instant", b564BountyOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("life %d -> %d, want +4", life, me.Life)
	}
}

func TestKrovikanRotNeedsPowerTwoOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	big := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	small := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Krovikan Rot", "Instant", b564RotOracle, effectstest.CardTarget(big)); err == nil {
		t.Fatal("targeted a 3-power creature: accepted")
	}
	effectstest.CastCatalogSpell(t, g, "Krovikan Rot", "Instant", b564RotOracle, effectstest.CardTarget(small))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) || !g.Battlefield.Contains(big) {
		t.Errorf("small on battlefield %v, big on battlefield %v", g.Battlefield.Contains(small), g.Battlefield.Contains(big))
	}
}

func TestGarzasAssassinSacrificesToDestroyANonblackCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	assassin := effectstest.BlbB1Push(g, me.ID, "Garza's Assassin", "Creature — Human Assassin", b564AssassinOracle, 2, 2)
	black := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Zombie", TypeLine: "Creature — Zombie", Colors: []string{"B"}, Power: 2, Toughness: 2})
	bear := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Colors: []string{"G"}, Power: 2, Toughness: 2})
	effectstest.AdvanceToMain(t, g)
	target := func(id uuid.UUID) game.ActivateAbilityParams {
		return game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(id)}
	}
	if err := g.ActivateCatalogAbility(me.ID, assassin, 0, target(black)); err == nil {
		t.Fatal("targeted a black creature: accepted")
	}
	if !g.Battlefield.Contains(assassin) {
		t.Fatal("a refused activation sacrificed the Assassin")
	}
	if err := g.ActivateCatalogAbility(me.ID, assassin, 0, target(bear)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || g.Battlefield.Contains(assassin) {
		t.Error("the Bear or the Assassin is still on the battlefield")
	}
}

func TestControvertCountersASpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: victim, Name: "Sorcery", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(opp.ID, victim, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the victim: %v", err)
	}
	cv := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: cv, Name: "Controvert", TypeLine: "Instant", OracleID: b564ControvertOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, cv, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(victim)}); err != nil {
		t.Fatalf("cast Controvert: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(victim) {
		t.Error("the victim was not countered into the graveyard")
	}
}

// --- Recover (CR 702.59a) ---------------------------------------------------------

// b564Die destroys `id` as an effect would.
func b564Die(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	g.RunStateChecksForTest()
}

// b564Answer walks the table to Recover's pay prompt for `who` and answers it,
// reporting whether one was asked.
func b564Answer(t *testing.T, g *game.Game, who *game.Player, pay bool) bool {
	t.Helper()
	for i := 0; i < 20; i++ {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == who.ID {
				if err := g.ResolvePayUnless(c.ID, who.ID, pay); err != nil {
					t.Fatalf("ResolvePayUnless: %v", err)
				}
				effectstest.PassPriorityAroundTable(t, g)
				return true
			}
		}
		if err := g.PassPriority(); err != nil {
			return false
		}
	}
	return false
}

func TestRecoverPaidReturnsTheCardToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	card := effectstest.PushCatalogGraveyardCard(me, "Sun's Bounty", "Instant", b564BountyOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{W}{W}")
	b564Die(t, g, bear)
	if !b564Answer(t, g, me, true) {
		t.Fatal("recover asked nothing")
	}
	if !effectstest.HandHasID(me, card) {
		t.Errorf("the card did not return to hand (graveyard %v, exile %v, pool %d)", me.Graveyard.Contains(card), effectstest.InExile(g, card), effectstest.PoolSize(g, me.ID))
	}
	if me.Graveyard.Contains(card) {
		t.Error("the card is still in the graveyard")
	}
}

func TestRecoverDeclinedExilesTheCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	card := effectstest.PushCatalogGraveyardCard(me, "Sun's Bounty", "Instant", b564BountyOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	b564Die(t, g, bear)
	if !b564Answer(t, g, me, false) {
		t.Fatal("recover asked nothing")
	}
	if !effectstest.InExile(g, card) {
		t.Error("the card was not exiled")
	}
	if effectstest.HandHasID(me, card) || me.Graveyard.Contains(card) {
		t.Error("the card is still in hand or the graveyard")
	}
}

func TestRecoverDoesNothingForAnOpponentsCreatureOrOnTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	card := effectstest.PushCatalogGraveyardCard(me, "Sun's Bounty", "Instant", b564BountyOracle, 0, 0)
	theirs := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{W}{W}")
	b564Die(t, g, theirs)
	b564Die(t, g, rock)
	if b564Answer(t, g, me, true) {
		t.Fatal("recover triggered for a creature that went to another graveyard, or for a noncreature")
	}
	if !me.Graveyard.Contains(card) {
		t.Error("the card left the graveyard")
	}
}

func TestRecoverTriggersForACreatureToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	card := effectstest.PushCatalogGraveyardCard(me, "Sun's Bounty", "Instant", b564BountyOracle, 0, 0)
	token := effectstest.PushToken(g, me.ID, game.Card{Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 1})
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{W}{W}")
	b564Die(t, g, token)
	if !b564Answer(t, g, me, true) {
		t.Fatal("recover asked nothing for a dying creature token")
	}
	if !effectstest.HandHasID(me, card) {
		t.Error("the card did not return to hand")
	}
}

func TestGarzasAssassinRecoverPaysHalfYourLifeRoundedUp(t *testing.T) {
	for _, tc := range []struct{ life, want int }{{40, 20}, {7, 4}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		me.Life = tc.life
		card := effectstest.PushCatalogGraveyardCard(me, "Garza's Assassin", "Creature — Human Assassin", b564AssassinOracle, 2, 2)
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		effectstest.AdvanceToMain(t, g)
		b564Die(t, g, bear)
		if !b564Answer(t, g, me, true) {
			t.Fatalf("life %d: recover asked nothing", tc.life)
		}
		if me.Life != tc.life-tc.want {
			t.Errorf("life %d -> %d, want to pay %d", tc.life, me.Life, tc.want)
		}
		if !effectstest.HandHasID(me, card) {
			t.Errorf("life %d: the Assassin did not return to hand", tc.life)
		}
	}
}

func TestGarzasAssassinRecoverDeclinedExilesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	card := effectstest.PushCatalogGraveyardCard(me, "Garza's Assassin", "Creature — Human Assassin", b564AssassinOracle, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	b564Die(t, g, bear)
	if !b564Answer(t, g, me, false) {
		t.Fatal("recover asked nothing")
	}
	if !effectstest.InExile(g, card) || me.Life != 40 {
		t.Errorf("exiled %v, life %d; want exiled and unchanged", effectstest.InExile(g, card), me.Life)
	}
}

func TestGarzasAssassinDoesNotRecoverItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	assassin := effectstest.BlbB1Push(g, me.ID, "Garza's Assassin", "Creature — Human Assassin", b564AssassinOracle, 2, 2)
	effectstest.AdvanceToMain(t, g)
	b564Die(t, g, assassin)
	if b564Answer(t, g, me, true) {
		t.Fatal("recover asked as the Assassin itself went to the graveyard")
	}
	if !me.Graveyard.Contains(assassin) {
		t.Error("the Assassin is not in the graveyard")
	}
}

// Two creatures put into the graveyard at once: the Assassin was not in the
// graveyard when the Bear went there, so its recover does not trigger for the
// Bear (CR 603.10a looks back only for leaves-the-battlefield abilities, and
// recover is not one).
func TestGarzasAssassinDyingWithAnotherCreatureDoesNotRecover(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	assassin := effectstest.BlbB1Push(g, me.ID, "Garza's Assassin", "Creature — Human Assassin", b564AssassinOracle, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{assassin, bear}) })
	g.RunStateChecksForTest()
	if b564Answer(t, g, me, true) {
		t.Fatal("recover triggered for a creature that died at the same time as the Assassin")
	}
	if !me.Graveyard.Contains(assassin) {
		t.Error("the Assassin is not in the graveyard")
	}
}
