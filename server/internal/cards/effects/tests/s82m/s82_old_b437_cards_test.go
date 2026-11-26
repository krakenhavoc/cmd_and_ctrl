package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b437_cards_test.go — Avacyn Restored, batch 437 (#5570, S82 OLD).
// Every test plays the real catalog entry through the engine.

const (
	oldB437Ghoul      = "12827076-450d-42b1-98ca-9f9279a539fc"
	oldB437Skaab      = "8fda6c20-04dc-4b60-9313-896a032ee5cb"
	oldB437Lumberknot = "97e0190d-a9a5-432d-9694-7576477bf65e"
	oldB437Bonfire    = "13822dcf-0639-4852-9a7f-e1d55719e82c"
	oldB437Revenge    = "6292edfc-d75b-465e-9e10-9d259551b1a5"
	oldB437Vanishment = "a2ae1a1b-5c4f-46ed-90da-bd50ca914e7c"
	oldB437Aggravate  = "192aa7b2-a819-4ffd-bfbf-a270271898e6"
	oldB437Malcontent = "eedaf405-4aa2-426a-af82-b0f92de7b6c3"
	oldB437Protector  = "33a5ad42-cd81-4724-9f68-7f32a1c8a047"
	oldB437Tyrant     = "ce1bf2c4-93d5-41e3-a450-6b95fad7dd21"
	oldB437Executor   = "e2bfac84-0e2f-4f3e-8d87-a555c26093a6"
	oldB437Bear       = "Creature — Bear"
	oldB437Human      = "Creature — Human"
)

func TestOldB437CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB437Ghoul: "Hunted Ghoul", oldB437Skaab: "Havengul Skaab", oldB437Lumberknot: "Flowering Lumberknot",
		oldB437Bonfire: "Bonfire of the Damned", oldB437Revenge: "Revenge of the Hunted", oldB437Vanishment: "Vanishment",
		oldB437Aggravate: "Aggravate", oldB437Malcontent: "Kessig Malcontents",
		oldB437Protector: "Midvast Protector", oldB437Tyrant: "Tyrant of Discord", oldB437Executor: "Undead Executioner",
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

func TestOldB437MiracleCardsDeclareTheirMiracleCost(t *testing.T) {
	for oracle, cost := range map[string]string{
		oldB437Bonfire: "{X}{R}", oldB437Revenge: "{G}", oldB437Vanishment: "{U}",
	} {
		spec, _ := effects.Lookup(oracle)
		found := false
		for _, ac := range spec.AlternativeCosts {
			if ac.Key == game.AltCostKeyMiracle && ac.ManaCost == cost {
				found = true
			}
		}
		if !found {
			t.Errorf("%s has no miracle %s", spec.Name, cost)
		}
	}
}

func TestOldB437HuntedGhoulCantBlockHumansButBlocksOthers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ghoul := effectstest.B12Push(g, opp.ID, "Hunted Ghoul", "Creature — Zombie", oldB437Ghoul, 1, 2)
	human := effectstest.BrCreature(g, me.ID, "Human", oldB437Human, 2, 2)
	bear := effectstest.BrCreature(g, me.ID, "Bear", oldB437Bear, 2, 2)
	effectstest.BrAttack(t, g, human, bear)
	effectstest.BrRefusal(t, g.DeclareBlockers([]game.BlockDeclaration{{Blocker: ghoul, Attacker: human}}), game.BlockReasonCantBlockAttacker)
	if err := g.DeclareBlockers([]game.BlockDeclaration{{Blocker: ghoul, Attacker: bear}}); err != nil {
		t.Fatalf("the Ghoul blocking a Bear: %v", err)
	}
}

func TestOldB437HavengulSkaabReturnsAnotherCreatureWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	skaab := effectstest.PushCatalogPermanent(g, me.ID, "Havengul Skaab", "Creature — Zombie Horror", oldB437Skaab, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB437Bear, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, skaab)
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if p == nil {
		t.Fatal("the attack trigger asked for no creature")
	}
	if p.ChooseMin != 1 {
		t.Error("the bounce is mandatory")
	}
	for _, id := range p.ChooseCards {
		if id == skaab {
			t.Error("the Skaab offered itself: the clause says another creature")
		}
	}
	if err := g.ResolveOwnPermanents(p.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	if g.Battlefield.Contains(bear) || !me.Hand.Contains(bear) {
		t.Error("the Bear did not return to its owner's hand")
	}
	if !g.Battlefield.Contains(skaab) {
		t.Error("the Skaab left the battlefield")
	}
}

func TestOldB437HavengulSkaabAloneBouncesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	skaab := effectstest.PushCatalogPermanent(g, me.ID, "Havengul Skaab", "Creature — Zombie Horror", oldB437Skaab, false)
	effectstest.DeclareAttack(t, g, opp.ID, skaab)
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) != nil {
		t.Error("asked for a creature with no other creature to return")
	}
	if !g.Battlefield.Contains(skaab) {
		t.Error("the Skaab returned itself")
	}
}

func TestOldB437VanishmentPutsANonlandPermanentOnTopOfItsOwnersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB437Bear, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Vanishment", "Instant", oldB437Vanishment, effectstest.CardTargets(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the permanent is still on the battlefield")
	}
	if top, err := opp.Library.Top(); err != nil || top.InstanceID != bear {
		t.Errorf("the top of the library is %q, want the Bear", top.Name)
	}
}

func TestOldB437VanishmentCantTargetALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	land := effectstest.PushCatalogPermanent(g, g.Seats[1].ID, "Forest", "Basic Land — Forest", "", false)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Vanishment", "Instant", oldB437Vanishment, effectstest.CardTargets(land)); err == nil {
		t.Error("Vanishment targeted a land")
	}
}

func TestOldB437VanishmentCastForItsMiracleCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[2]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB437Bear, 2, 2)
	van := effectstest.DrawMiracleInOwnDrawStep(t, g, 1, "Vanishment", "Instant", "{4}{U}", oldB437Vanishment)
	effectstest.RevealAndResolve(t, g, me)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, van, game.CastSpellParams{
		Strict: true, AlternativeCost: game.AltCostKeyMiracle, Targets: effectstest.CardTargets(bear),
	}); err != nil {
		t.Fatalf("the miracle cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if top, err := opp.Library.Top(); g.Battlefield.Contains(bear) || err != nil || top.InstanceID != bear {
		t.Error("the Bear is not on top of its owner's library")
	}
}

func TestOldB437RevengeOfTheHuntedBoostsAndForcesBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB437Bear, 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", oldB437Bear, 1, 1)
	idle := effectstest.B12Creature(g, opp.ID, "Idle", oldB437Bear, 1, 1)
	effectstest.CastCatalogSpell(t, g, "Revenge of the Hunted", "Sorcery", oldB437Revenge, effectstest.CardTargets(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 8 || tough != 8 {
		t.Errorf("the target is %d/%d, want 8/8", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the target has no trample")
	}
	effectstest.BrAttack(t, g, bear)
	must := effectstest.MustBlockNow(g)
	if must[blocker] != bear || must[idle] != bear {
		t.Errorf("creatures able to block it are not all forced to: %v", must)
	}
}

func TestOldB437MidvastProtectorGivesProtectionFromAChosenColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB437Bear, 2, 2)
	effectstest.CastCreatureCard(t, g, "Midvast Protector", "Creature — Human Wizard", oldB437Protector, 2, 3)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: bear}, true)
	effectstest.AnswerColor(t, g, me.ID, "R")
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, game.ProtectionFromColor("R")) {
		t.Error("the target did not gain protection from red")
	}
}

func TestOldB437UndeadExecutionerMayShrinkATargetWhenItDies(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		dier := effectstest.B12Push(g, me.ID, "Undead Executioner", "Creature — Zombie", oldB437Executor, 2, 2)
		target := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
		effectstest.AdvanceToMain(t, g)
		effectstest.KillCreature(t, g, me.ID, dier)
		effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: target}, yes)
		want := 3
		if yes {
			want = 1
		}
		if p, tough := effectstest.EffectivePTOf(t, g, target); p != want || tough != want {
			t.Errorf("yes=%v: the target is %d/%d, want %d/%d", yes, p, tough, want, want)
		}
	}
}

func TestOldB437KessigMalcontentsDealsDamageEqualToHumans(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Creature(g, me.ID, "Human A", oldB437Human, 1, 1)
	effectstest.B12Creature(g, me.ID, "Human B", oldB437Human, 1, 1)
	effectstest.B12Creature(g, me.ID, "Bear", oldB437Bear, 1, 1)
	effectstest.B12Creature(g, opp.ID, "Their Human", oldB437Human, 1, 1)
	life := opp.Life
	effectstest.CastCreatureCard(t, g, "Kessig Malcontents", "Creature — Human Warrior", oldB437Malcontent, 3, 1)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}, true)
	// Two Humans and the Malcontents itself.
	if opp.Life != life-3 {
		t.Errorf("life %d → %d, want 3 damage (the Humans you control, itself included)", life, opp.Life)
	}
}

func TestOldB437AggravateDamagesEachCreatureOfTheTargetPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", oldB437Bear, 2, 3)
	b := effectstest.B12Creature(g, opp.ID, "B", oldB437Bear, 2, 3)
	mine := effectstest.B12Creature(g, me.ID, "Mine", oldB437Bear, 2, 3)
	effectstest.CastCatalogSpell(t, g, "Aggravate", "Instant", oldB437Aggravate, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if got := effectstest.DamageOn(t, g, id); got != 1 {
			t.Errorf("their creature has %d damage, want 1", got)
		}
	}
	if effectstest.DamageOn(t, g, mine) != 0 {
		t.Error("your own creature was damaged")
	}
}

func TestOldB437AggravateForcesTheDamagedCreaturesToAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Creature(g, me.ID, "Mine", oldB437Bear, 2, 3)
	effectstest.CastCatalogSpell(t, g, "Aggravate", "Instant", oldB437Aggravate, []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.FaMustAttack(t, g) {
		t.Error("a creature dealt damage this way is not required to attack")
	}

	g = effectstest.NewCatalogGame(t)
	effectstest.B12Creature(g, g.Seats[0].ID, "Mine", oldB437Bear, 2, 3)
	if effectstest.FaMustAttack(t, g) {
		t.Error("without Aggravate the creature is required to attack")
	}
}

func TestOldB437BonfireOfTheDamnedDealsXToThePlayerAndTheirCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Wall", 0, 9)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Wall", 0, 9)
	life := opp.Life
	effectstest.AssistCardCastX(t, g, "Bonfire of the Damned", "Sorcery", oldB437Bonfire, "{X}{X}{R}", []string{"R", "C", "C", "C", "C"}, 2,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d → %d, want 2 damage", life, opp.Life)
	}
	if got := effectstest.DamageOn(t, g, theirs); got != 2 {
		t.Errorf("their creature has %d damage, want 2", got)
	}
	if effectstest.DamageOn(t, g, mine) != 0 {
		t.Error("your own creature was damaged")
	}
}

func TestOldB437TyrantOfDiscordStopsAtALandAndSweepsNonlands(t *testing.T) {
	// Only lands: the first sacrifice is a land, so the process ends there.
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 3; i++ {
		effectstest.PushCatalogPermanent(g, opp.ID, "Forest", "Basic Land — Forest", "", false)
	}
	effectstest.CastCreatureCard(t, g, "Tyrant of Discord", "Creature — Elemental", oldB437Tyrant, 7, 7)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}, true)
	if got := oldB437Controlled(g, opp.ID); got != 2 {
		t.Errorf("the opponent controls %d permanents, want 2 (one land sacrificed, the process stopped)", got)
	}

	// Only nonlands: each sacrifice repeats the process until nothing is left.
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	for i := 0; i < 3; i++ {
		effectstest.B12Creature(g, opp.ID, "Bear", oldB437Bear, 1, 1)
	}
	effectstest.CastCreatureCard(t, g, "Tyrant of Discord", "Creature — Elemental", oldB437Tyrant, 7, 7)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}, true)
	if got := oldB437Controlled(g, opp.ID); got != 0 {
		t.Errorf("the opponent controls %d permanents, want 0", got)
	}
}

func TestOldB437FloweringLumberknotNeedsASoulbondPartnerToAttack(t *testing.T) {
	for _, tc := range []struct {
		name    string
		partner []string
		pair    bool
		may     bool
	}{
		{"unpaired", nil, false, false},
		{"paired with a creature lacking soulbond", nil, true, false},
		{"paired with a soulbond creature", []string{game.KeywordSoulbond}, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			knot := effectstest.PushCatalogPermanent(g, me.ID, "Flowering Lumberknot", "Creature — Treefolk", oldB437Lumberknot, false)
			partner := effectstest.BrCreature(g, me.ID, "Partner", oldB437Bear, 2, 2, tc.partner...)
			if tc.pair {
				g.WithWriteLock(func() {
					if !g.PairForEffect(me.ID, knot, partner) {
						t.Error("the pair was refused")
					}
				})
			}
			effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
			err := effectstest.DeclareResult(g, knot, opp.ID)
			if tc.may && err != nil {
				t.Errorf("the attack was refused: %v", err)
			}
			if !tc.may && err == nil {
				t.Error("the attack was allowed")
			}
		})
	}
}

func TestOldB437FloweringLumberknotBlockHalf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	knot := effectstest.B12Push(g, opp.ID, "Flowering Lumberknot", "Creature — Treefolk", oldB437Lumberknot, 5, 5)
	partner := effectstest.BrCreature(g, opp.ID, "Bonded", oldB437Bear, 2, 2, game.KeywordSoulbond)
	attacker := effectstest.BrCreature(g, me.ID, "Bear", oldB437Bear, 2, 2)
	effectstest.BrAttack(t, g, attacker)
	err := g.DeclareBlockers([]game.BlockDeclaration{{Blocker: knot, Attacker: attacker}})
	effectstest.BrRefusal(t, err, game.BlockReasonCantBlockAttacker)
	g.WithWriteLock(func() {
		if !g.PairForEffect(opp.ID, knot, partner) {
			t.Error("the pair was refused")
		}
	})
	if err := g.DeclareBlockers([]game.BlockDeclaration{{Blocker: knot, Attacker: attacker}}); err != nil {
		t.Errorf("a paired Lumberknot may not block: %v", err)
	}
}

// oldB437Controlled counts the permanents a player controls.
func oldB437Controlled(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == player {
			n++
		}
	}
	return n
}
