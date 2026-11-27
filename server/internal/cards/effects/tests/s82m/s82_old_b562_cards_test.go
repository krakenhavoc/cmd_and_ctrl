package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b562_cards_test.go — Coldsnap's damage cards, played from the real
// catalog entries (S82 OLD batch 562).

const (
	b562SkredOracle       = "026318be-e3dd-4d5c-bd77-53ce23a2e7cc"
	b562TamanoaOracle     = "14440669-ddf0-46ac-b140-13948a8113b8"
	b562BloodpainterOracl = "259bba2f-8c68-4267-8f51-0b2144778653"
	b562FeastOracle       = "6f214306-1664-41ec-b999-1b7471bbcbd6"
	b562SpecterOracle     = "b6f4148d-8511-46b3-93ec-2f5eb1e4ffad"
	b562WolverineOracle   = "ce690da4-6d22-4013-aef3-596d331813e4"
	b562YetiOracle        = "e79e67db-3f2e-4002-9e6c-8f06ca9d8af4"
)

func TestS82OldB562CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b562SkredOracle:       "Skred",
		b562TamanoaOracle:     "Tamanoa",
		b562BloodpainterOracl: "Orcish Bloodpainter",
		b562FeastOracle:       "Feast of Flesh",
		b562SpecterOracle:     "Blizzard Specter",
		b562WolverineOracle:   "Karplusan Wolverine",
		b562YetiOracle:        "Stalking Yeti",
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

func b562Snow(g *game.Game, owner uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		effectstest.BlbB1Push(g, owner, "Snow-Covered Mountain", "Basic Snow Land — Mountain", "", 0, 0)
	}
}

// b562CastSkred casts Skred at `victim` with {R} floating and resolves it.
func b562CastSkred(t *testing.T, g *game.Game, me *game.Player, victim uuid.UUID) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{R}")
	card := effectstest.PushCatalogHandCard(me, "Skred", "Instant", b562SkredOracle)
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == card {
				me.Hand.Cards[i].ManaCost = "{R}"
			}
		}
	})
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
		t.Fatalf("cast Skred: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestSkredDealsDamageEqualToYourSnowPermanents(t *testing.T) {
	for _, snow := range []int{0, 1, 3} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		b562Snow(g, me.ID, snow)
		// Snow permanents an opponent controls do not count.
		b562Snow(g, opp.ID, 2)
		victim := effectstest.PushVanillaCreature(g, opp.ID, "Wall", 0, 9)
		b562CastSkred(t, g, me, victim)
		if got := effectstest.DamageOn(t, g, victim); got != snow {
			t.Errorf("%d snow permanents: Skred dealt %d", snow, got)
		}
	}
}

func TestTamanoaGainsLifeFromNoncreatureSourcesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.BlbB1Push(g, me.ID, "Tamanoa", "Creature — Spirit", b562TamanoaOracle, 2, 4)
	b562Snow(g, me.ID, 2)
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Wall", 0, 9)
	life := me.Life
	b562CastSkred(t, g, me, victim)
	if me.Life != life+2 {
		t.Fatalf("a noncreature source dealt 2: life %d -> %d, want +2", life, me.Life)
	}

	// A creature's damage (the Bloodpainter's ping) gains nothing.
	pyro := effectstest.BlbB1Push(g, me.ID, "Orcish Bloodpainter", "Creature — Orc Shaman", b562BloodpainterOracl, 2, 1)
	fodder := effectstest.PushVanillaCreature(g, me.ID, "Fodder", 1, 1)
	life = me.Life
	if err := g.ActivateCatalogAbility(me.ID, pyro, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder},
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("Bloodpainter: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("a creature's damage gained life: %d -> %d", life, me.Life)
	}
}

func TestOrcishBloodpainterSacrificesACreatureToPing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pyro := effectstest.BlbB1Push(g, me.ID, "Orcish Bloodpainter", "Creature — Orc Shaman", b562BloodpainterOracl, 2, 1)
	fodder := effectstest.PushVanillaCreature(g, me.ID, "Fodder", 1, 1)
	effectstest.AdvanceToMain(t, g)
	// Without a creature to sacrifice named, the activation is refused.
	if err := g.ActivateCatalogAbility(me.ID, pyro, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Fatal("activated with no sacrifice: accepted")
	}
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, pyro, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder},
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life %d -> %d, want -1", life, opp.Life)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	if !b562Tapped(t, g, pyro) {
		t.Error("the Bloodpainter was not tapped")
	}
}

func b562Tapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c := effectstest.FindBattlefieldCardByID(g, id)
	if c == nil {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Tapped
}

func TestFeastOfFleshCountsFeastsInAllGraveyards(t *testing.T) {
	for _, tc := range []struct{ inMine, inTheirs, want int }{{0, 0, 1}, {1, 0, 2}, {1, 2, 4}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		for i := 0; i < tc.inMine; i++ {
			effectstest.PushCatalogGraveyardCard(me, "Feast of Flesh", "Sorcery", b562FeastOracle, 0, 0)
		}
		for i := 0; i < tc.inTheirs; i++ {
			effectstest.PushCatalogGraveyardCard(opp, "Feast of Flesh", "Sorcery", b562FeastOracle, 0, 0)
		}
		victim := effectstest.PushVanillaCreature(g, opp.ID, "Wall", 0, 9)
		effectstest.AdvanceToMain(t, g)
		effectstest.FloatMana(t, g, me, "{B}")
		card := effectstest.PushCatalogHandCard(me, "Feast of Flesh", "Sorcery", b562FeastOracle)
		g.WithWriteLock(func() {
			for i := range me.Hand.Cards {
				if me.Hand.Cards[i].InstanceID == card {
					me.Hand.Cards[i].ManaCost = "{B}"
				}
			}
		})
		life := me.Life
		if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.DamageOn(t, g, victim); got != tc.want {
			t.Errorf("%d+%d in graveyards: damage %d, want %d", tc.inMine, tc.inTheirs, got, tc.want)
		}
		if me.Life != life+tc.want {
			t.Errorf("%d+%d in graveyards: life %d -> %d, want +%d", tc.inMine, tc.inTheirs, life, me.Life, tc.want)
		}
	}
}

func TestKarplusanWolverinePingsWhenBlocked(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[1]
		id := effectstest.NcdAttacker(g, "Karplusan Wolverine", b562WolverineOracle, 1, 1)
		wall := effectstest.PushVanillaCreature(g, opp.ID, "Wall", 0, 9)
		victim := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 5)
		effectstest.Pr7bAttack(t, g, opp.ID, id)
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		if err := g.DeclareBlocker(wall, id); err != nil {
			t.Fatal(err)
		}
		effectstest.FinishBlockDeclarations(t, g)
		for i := 0; i < 30; i++ {
			var open *game.PendingChoice
			for _, c := range g.PendingChoices {
				if c != nil {
					open = c
					break
				}
			}
			if open == nil {
				if len(g.StackMeta) == 0 && len(g.PendingTriggers) == 0 && i > 2 {
					break
				}
				if err := g.PassPriority(); err != nil {
					break
				}
				continue
			}
			switch open.Kind {
			case game.PendingChoicePickTarget:
				if err := g.ResolvePickTargets(open.ID, g.Seats[0].ID, []game.TargetRef{{Kind: game.TargetCard, ID: victim}}); err != nil {
					t.Fatalf("pick target: %v", err)
				}
			case game.PendingChoiceTriggerPrompt:
				if err := g.ResolveTriggerPrompt(open.ID, g.Seats[0].ID, yes); err != nil {
					t.Fatalf("trigger prompt: %v", err)
				}
			default:
				t.Fatalf("unexpected prompt %s", open.Kind)
			}
		}
		got := effectstest.DamageOn(t, g, victim)
		if yes && got != 1 || !yes && got != 0 {
			t.Errorf("yes=%v: the target has %d damage", yes, got)
		}
	}
}

func TestBlizzardSpecterModes(t *testing.T) {
	for mode := 0; mode < 2; mode++ {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		id := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Blizzard Specter", TypeLine: "Snow Creature — Specter", OracleID: b562SpecterOracle, Power: 2, Toughness: 3, Keywords: []string{"flying"}})
		keep := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
		handBefore := opp.Hand.Size()
		effectstest.Pr7bAttack(t, g, opp.ID, id)
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		effectstest.FinishBlockDeclarations(t, g)
		answeredMode := false
		for i := 0; i < 30; i++ {
			if c := effectstest.ModePickChoiceFor(g, me.ID); c != nil {
				effectstest.AnswerMode(t, g, c, me.ID, mode)
				answeredMode = true
				continue
			}
			var open *game.PendingChoice
			for _, c := range g.PendingChoices {
				if c != nil {
					open = c
					break
				}
			}
			if open != nil {
				switch open.Kind {
				case game.PendingChoiceOwnPermanents:
					if err := g.ResolveOwnPermanents(open.ID, opp.ID, []uuid.UUID{keep}); err != nil {
						t.Fatalf("pick permanent: %v", err)
					}
				case game.PendingChoiceChooseCards:
					c := open
					if err := g.ResolveChooseCards(c.ID, c.Chooser, []uuid.UUID{opp.Hand.Cards[0].InstanceID}); err != nil {
						t.Fatalf("discard: %v", err)
					}
				default:
					t.Fatalf("unexpected prompt %s", open.Kind)
				}
				continue
			}
			if len(g.StackMeta) == 0 && len(g.PendingTriggers) == 0 && answeredMode {
				break
			}
			if err := g.PassPriority(); err != nil {
				break
			}
		}
		if !answeredMode {
			t.Fatalf("mode %d: no mode pick", mode)
		}
		switch mode {
		case 0:
			if g.Battlefield.Contains(keep) {
				t.Errorf("mode 0: the Bear was not returned to hand")
			}
		case 1:
			if opp.Hand.Size() != handBefore-1 {
				t.Errorf("mode 1: hand %d -> %d, want one discard", handBefore, opp.Hand.Size())
			}
			if !g.Battlefield.Contains(keep) {
				t.Errorf("mode 1: the Bear left the battlefield")
			}
		}
	}
}

func TestStalkingYetiBitesBothWaysAndGoesBackToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b562Snow(g, me.ID, 3)
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Ogre", 2, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{R}{R}{R}{R}")
	card := effectstest.PushCatalogHandCard(me, "Stalking Yeti", "Snow Creature — Yeti", b562YetiOracle)
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == card {
				me.Hand.Cards[i].ManaCost = "{2}{R}{R}"
				me.Hand.Cards[i].Power, me.Hand.Cards[i].Toughness = 3, 3
			}
		}
	})
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerFirstPickTarget(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("3 damage did not kill the 2/3 Ogre")
	}
	if got := effectstest.DamageOn(t, g, card); got != 2 {
		t.Errorf("the Yeti took %d damage, want 2", got)
	}

	// {2}{S}: back to hand, at sorcery speed only.
	if err := g.ActivateCatalogAbility(me.ID, card, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("return: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(card) {
		t.Error("the Yeti is still on the battlefield")
	}
}
