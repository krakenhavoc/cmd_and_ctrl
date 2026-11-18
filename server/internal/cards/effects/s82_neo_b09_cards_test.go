package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_neo_b09_cards_test.go — Kamigawa: Neon Dynasty, batch 9 (S82): the
// "Samurai or Warrior attacks alone" cards. Each test plays the real catalog
// entry through the engine.

const (
	neoB09NorikaOracle    = "6f143c8f-cb5c-4abc-a01b-b6813f5a524c"
	neoB09HeikoOracle     = "a4d67f2d-6b0f-4feb-bf7d-90ab4f2e46ab"
	neoB09RaiyuuOracle    = "9028f726-50c7-49f3-8386-e584415dcba2"
	neoB09SubduerOracle   = "bedf1dd5-b133-455b-bc3d-c7b1f469d91f"
	neoB09PeerlessOracle  = "fe3b0393-dc50-4fa6-9b3f-1175a6a825e4"
	neoB09ExemplarOracle  = "8e0f4e22-0677-4246-a4c5-6fcc74cb54a3"
	neoB09SelflessOracle  = "20c9fb06-3941-456f-ba02-4fbde9deb1e2"
	neoB09CaptainOracle   = "32e7385e-525e-4925-8cfd-bfea21fcd1bf"
	neoB09RoninOracle     = "4b4c76e1-c480-4d10-93d4-483ef2debc4b"
	neoB09KatanaOracle    = "13a9b9fd-20fa-4734-9d95-3ab41ab9c7a0"
	neoB09TemperedOracle  = "f4aedfd3-4414-493c-bbba-75db301c0737"
	neoB09IronsoulOracle  = "a613b00e-d6e8-4470-a7dc-5eab5e48dce0"
	neoB09SamuraiTypeLine = "Creature — Human Samurai"
)

func TestNeoB09EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		neoB09NorikaOracle:   "Norika Yamazaki, the Poet",
		neoB09HeikoOracle:    "Heiko Yamazaki, the General",
		neoB09RaiyuuOracle:   "Raiyuu, Storm's Edge",
		neoB09SubduerOracle:  "Imperial Subduer",
		neoB09PeerlessOracle: "Peerless Samurai",
		neoB09ExemplarOracle: "Eiganjo Exemplar",
		neoB09SelflessOracle: "Selfless Samurai",
		neoB09CaptainOracle:  "Asari Captain",
		neoB09RoninOracle:    "Akki Ronin",
		neoB09KatanaOracle:   "Ancestral Katana",
		neoB09TemperedOracle: "Tempered in Solitude",
		neoB09IronsoulOracle: "Ironsoul Enforcer",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// neoB09Table is a table with the card under test on the battlefield for the
// active player and an opponent.
func neoB09Table(t *testing.T, name, typeLine, oracle string, p, tough int) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := b12Push(g, me.ID, name, typeLine, oracle, p, tough)
	return g, me, opp, id
}

// --- the shared lone-attack condition ------------------------------------------------

func TestNeoB09TriggerNeedsASamuraiOrWarriorAttackingAlone(t *testing.T) {
	cases := []struct {
		name      string
		typeLine  string
		second    bool
		wantPower int
	}{
		{"a lone Samurai", "Creature — Human Samurai", false, 3},
		{"a lone Warrior", "Creature — Human Warrior", false, 3},
		{"a lone non-Samurai, non-Warrior", "Creature — Human Cleric", false, 2},
		{"a Samurai attacking beside another creature", "Creature — Human Samurai", true, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g, me, opp, _ := neoB09Table(t, "Eiganjo Exemplar", "Enchantment Creature — Human Samurai", neoB09ExemplarOracle, 2, 1)
			atk := b12Creature(g, me.ID, "Attacker", tc.typeLine, 2, 2)
			// The Exemplar itself does not attack; tap it so it cannot be
			// swept in by mistake.
			attackers := []uuid.UUID{atk}
			if tc.second {
				attackers = append(attackers, b12Creature(g, me.ID, "Friend", "Creature — Bear", 2, 2))
			}
			declareAttack(t, g, opp.ID, attackers...)
			passPriorityAroundTable(t, g)
			if got := hobA4Power(t, g, atk); got != tc.wantPower {
				t.Errorf("attacker power %d, want %d", got, tc.wantPower)
			}
		})
	}
}

func TestNeoB09TriggerIgnoresAnOpponentsLoneSamurai(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, opp.ID, "Eiganjo Exemplar", "Enchantment Creature — Human Samurai", neoB09ExemplarOracle, 2, 1)
	atk := b12Creature(g, me.ID, "Attacker", "Creature — Human Samurai", 2, 2)
	declareAttack(t, g, opp.ID, atk)
	passPriorityAroundTable(t, g)
	if got := hobA4Power(t, g, atk); got != 2 {
		t.Errorf("power %d, want 2: the opponent's Exemplar watches their own creatures", got)
	}
}

// --- Eiganjo Exemplar ----------------------------------------------------------------

func TestNeoB09EiganjoExemplarPumpsItselfAttackingAlone(t *testing.T) {
	g, _, opp, ex := neoB09Table(t, "Eiganjo Exemplar", "Enchantment Creature — Human Samurai", neoB09ExemplarOracle, 2, 1)
	declareAttack(t, g, opp.ID, ex)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, ex), effectiveToughness(t, g, ex); p != 3 || tough != 2 {
		t.Errorf("Exemplar is %d/%d, want 3/2", p, tough)
	}
}

// --- Asari Captain -------------------------------------------------------------------

func TestNeoB09AsariCaptainPumpsForEachSamuraiOrWarrior(t *testing.T) {
	g, me, opp, _ := neoB09Table(t, "Asari Captain", "Creature — Human Samurai", neoB09CaptainOracle, 4, 3)
	atk := b12Creature(g, me.ID, "Warrior", "Creature — Human Warrior", 1, 1)
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Their Samurai", "Creature — Human Samurai", 2, 2)
	declareAttack(t, g, opp.ID, atk)
	passPriorityAroundTable(t, g)
	// The Captain and the Warrior are mine; the opposing Samurai and the Bear
	// do not count.
	if got := hobA4Power(t, g, atk); got != 3 {
		t.Errorf("Warrior has power %d, want 3 (1 + 2)", got)
	}
}

// --- Selfless Samurai ----------------------------------------------------------------

func TestNeoB09SelflessSamuraiGivesTheLoneAttackerLifelink(t *testing.T) {
	g, _, opp, sam := neoB09Table(t, "Selfless Samurai", "Creature — Fox Samurai", neoB09SelflessOracle, 2, 2)
	declareAttack(t, g, opp.ID, sam)
	passPriorityAroundTable(t, g)
	if !gl6HasKeyword(t, g, sam, "lifelink") {
		t.Error("the lone attacker did not gain lifelink")
	}
}

func TestNeoB09SelflessSamuraiSacrificesForIndestructible(t *testing.T) {
	g, me, _, sam := neoB09Table(t, "Selfless Samurai", "Creature — Fox Samurai", neoB09SelflessOracle, 2, 2)
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, sam, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: sam}},
	}); err == nil {
		t.Fatal("the Samurai named itself as the other creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, sam, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(sam) {
		t.Error("the Samurai was not sacrificed")
	}
	if !gl6HasKeyword(t, g, bear, "indestructible") {
		t.Error("the Bear did not gain indestructible")
	}
}

// --- Peerless Samurai ----------------------------------------------------------------

func TestNeoB09PeerlessSamuraiMakesTheNextSpellCheaper(t *testing.T) {
	g, me, opp, sam := neoB09Table(t, "Peerless Samurai", neoB09SamuraiTypeLine, neoB09PeerlessOracle, 2, 3)
	declareAttack(t, g, opp.ID, sam)
	passPriorityAroundTable(t, g)
	if len(nsPromises(me)) != 1 {
		t.Fatalf("the trigger made no promise: %+v", me.Statics)
	}
	spell := game.Card{Name: "Test Spell", TypeLine: "Sorcery", ManaCost: "{2}{R}", Owner: me.ID}
	if got := nsPrice(t, g, me, spell); got.Generic != 1 {
		t.Errorf("the next spell costs {1} less: %+v", got)
	}
	if got := nsPrice(t, g, opp, spell); got.Generic != 2 {
		t.Errorf("an opponent's spell is not discounted: %+v", got)
	}
}

// --- Imperial Subduer ----------------------------------------------------------------

func TestNeoB09ImperialSubduerTapsACreatureYouDontControl(t *testing.T) {
	g, me, opp, sam := neoB09Table(t, "Imperial Subduer", neoB09SamuraiTypeLine, neoB09SubduerOracle, 3, 2)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, sam)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the trigger asked for no target")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: mine}}); err == nil {
		t.Error("a creature its controller controls was accepted")
	}
	answerTriggerTargets(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if !tappedOnBattlefield(t, g, theirs) {
		t.Error("the opposing creature is not tapped")
	}
}

// --- Raiyuu --------------------------------------------------------------------------

func TestNeoB09RaiyuuUntapsAndAddsACombatOnlyInTheFirst(t *testing.T) {
	g, _, opp, ray := neoB09Table(t, "Raiyuu, Storm's Edge", "Legendary Creature — Human Samurai", neoB09RaiyuuOracle, 3, 3)
	declareAttack(t, g, opp.ID, ray)
	passPriorityAroundTable(t, g)
	if isTapped(g, ray) {
		t.Error("Raiyuu is still tapped")
	}
	advanceTo(t, g, game.StepEndCombat)
	if next := stepOnce(t, g); next.Step != game.StepBeginCombat {
		t.Errorf("after the first combat %+v, want another combat", next)
	}
	// The additional combat is not the first: attacking alone again untaps
	// but adds nothing.
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(ray, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if isTapped(g, ray) {
		t.Error("Raiyuu is still tapped after the second combat's trigger")
	}
	advanceTo(t, g, game.StepEndCombat)
	if next := stepOnce(t, g); next.Step != game.StepPostcombatMain {
		t.Errorf("after the additional combat %+v, want the main phase", next)
	}
}

// --- Akki Ronin ----------------------------------------------------------------------

func TestNeoB09AkkiRoninRummages(t *testing.T) {
	for _, discard := range []bool{true, false} {
		name := "declined"
		if discard {
			name = "discarded"
		}
		t.Run(name, func(t *testing.T) {
			g, me, opp, ronin := neoB09Table(t, "Akki Ronin", "Creature — Goblin Samurai", neoB09RoninOracle, 1, 3)
			hand, graves := me.Hand.Size(), me.Graveyard.Size()
			pitch := me.Hand.Cards[0].InstanceID
			declareAttack(t, g, opp.ID, ronin)
			passPriorityAroundTable(t, g)
			if discard {
				answerDiscard(t, g, me.ID, pitch)
			} else {
				answerDiscard(t, g, me.ID)
			}
			passPriorityAroundTable(t, g)
			if got := me.Hand.Size(); got != hand {
				t.Errorf("hand %d, want %d", got, hand)
			}
			want := graves
			if discard {
				want++
			}
			if got := me.Graveyard.Size(); got != want {
				t.Errorf("graveyard %d, want %d", got, want)
			}
		})
	}
}

// --- Tempered in Solitude ------------------------------------------------------------

func TestNeoB09TemperedInSolitudeExilesTheTopCardToPlayThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Tempered in Solitude", "Enchantment", neoB09TemperedOracle, 0, 0)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	top := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: top, Name: "Top Spell", TypeLine: "Artifact", ManaCost: "{0}", Owner: me.ID, Controller: me.ID})
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(top) {
		t.Fatal("the top card was not exiled")
	}
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("cast the exiled card: %v", err)
	}
}

func TestNeoB09TemperedInSolitudeNeedsALoneAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Tempered in Solitude", "Enchantment", neoB09TemperedOracle, 0, 0)
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	exiled := g.Exile.Size()
	declareAttack(t, g, opp.ID, a, b)
	passPriorityAroundTable(t, g)
	if g.Exile.Size() != exiled {
		t.Error("two attackers still exiled a card")
	}
}

// --- Norika / Heiko ------------------------------------------------------------------

func TestNeoB09NorikaAndHeikoGrantACastFromTheGraveyard(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle, wrongType, rightType string
	}{
		{"Norika Yamazaki, the Poet", "Legendary Creature — Human Samurai", neoB09NorikaOracle, "Artifact", "Enchantment"},
		{"Heiko Yamazaki, the General", "Legendary Creature — Human Samurai", neoB09HeikoOracle, "Enchantment", "Artifact"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g, me, opp, lord := neoB09Table(t, tc.name, tc.typeLine, tc.oracle, 3, 2)
			wrong := ptGraveCard(me, "Wrong Card", tc.wrongType)
			right := ptGraveCard(me, "Right Card", tc.rightType)
			declareAttack(t, g, opp.ID, lord)
			pick := latestPickTarget(g, me.ID)
			if pick == nil {
				t.Fatal("the trigger asked for no target")
			}
			if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: wrong}}); err == nil {
				t.Error("a card of the wrong type was accepted")
			}
			answerTriggerTargets(t, g, me.ID, right)
			passPriorityAroundTable(t, g)
			advanceToMain(t, g)
			if err := g.CastSpell(me.ID, wrong, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
				t.Error("the untargeted card could be cast")
			}
			if err := g.CastSpell(me.ID, right, game.CastSpellParams{FromZone: "graveyard"}); err != nil {
				t.Fatalf("cast the targeted card: %v", err)
			}
			passPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(right) {
				t.Error("the card did not resolve onto the battlefield")
			}
		})
	}
}

// --- Ironsoul Enforcer ---------------------------------------------------------------

func TestNeoB09IronsoulEnforcerReturnsAnArtifactWhenItOrACommanderAttacksAlone(t *testing.T) {
	t.Run("itself", func(t *testing.T) {
		g, me, opp, ens := neoB09Table(t, "Ironsoul Enforcer", "Artifact Creature — Human Samurai", neoB09IronsoulOracle, 4, 4)
		rock := ptGraveCard(me, "Dead Rock", "Artifact")
		declareAttack(t, g, opp.ID, ens)
		answerTriggerTargets(t, g, me.ID, rock)
		passPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(rock) {
			t.Error("the artifact did not return")
		}
	})
	t.Run("a commander", func(t *testing.T) {
		g, me, opp, _ := neoB09Table(t, "Ironsoul Enforcer", "Artifact Creature — Human Samurai", neoB09IronsoulOracle, 4, 4)
		rock := ptGraveCard(me, "Dead Rock", "Artifact")
		cmdr := b12Creature(g, me.ID, "Commander", "Legendary Creature — Bear", 2, 2)
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == cmdr {
					g.Battlefield.Cards[i].IsCommander = true
				}
			}
		})
		declareAttack(t, g, opp.ID, cmdr)
		answerTriggerTargets(t, g, me.ID, rock)
		passPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(rock) {
			t.Error("the artifact did not return")
		}
	})
	t.Run("an ordinary creature", func(t *testing.T) {
		g, me, opp, _ := neoB09Table(t, "Ironsoul Enforcer", "Artifact Creature — Human Samurai", neoB09IronsoulOracle, 4, 4)
		rock := ptGraveCard(me, "Dead Rock", "Artifact")
		bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		declareAttack(t, g, opp.ID, bear)
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(rock) {
			t.Error("an ordinary creature's lone attack returned an artifact")
		}
	})
}

// --- Ancestral Katana ----------------------------------------------------------------

func TestNeoB09AncestralKatanaAttachesForOneMana(t *testing.T) {
	for _, pay := range []bool{true, false} {
		name := "declined"
		if pay {
			name = "paid"
		}
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			katana := seedEquipment(g, me.ID, "Ancestral Katana", neoB09KatanaOracle)
			sam := b12Creature(g, me.ID, "Samurai", neoB09SamuraiTypeLine, 2, 2)
			declareAttack(t, g, opp.ID, sam)
			passPriorityAroundTable(t, g)
			if !hasPayUnlessFor(g, me.ID) {
				t.Fatal("no {1} offer")
			}
			if pay {
				floatMana(t, g, me, "{C}")
			}
			answerPayUnless(t, g, me.ID, pay)
			passPriorityAroundTable(t, g)
			host := attachmentHostOf(t, g, katana)
			if !pay {
				if host.Kind == game.TargetCard {
					t.Errorf("declined, but the Katana is attached to %s", host.ID)
				}
				return
			}
			if host.Kind != game.TargetCard || host.ID != sam {
				t.Fatalf("host %+v, want the Samurai", host)
			}
			if p, tough := effectivePower(t, g, sam), effectiveToughness(t, g, sam); p != 4 || tough != 3 {
				t.Errorf("the Samurai is %d/%d, want 4/3", p, tough)
			}
		})
	}
}

func TestNeoB09AncestralKatanaDoesNotOfferForANonSamurai(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedEquipment(g, me.ID, "Ancestral Katana", neoB09KatanaOracle)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		t.Error("a Bear attacking alone offered the attach")
	}
}
