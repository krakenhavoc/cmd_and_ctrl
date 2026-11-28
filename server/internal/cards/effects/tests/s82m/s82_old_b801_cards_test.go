package testss82m

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b801_cards_test.go — Mirage's card draw and library cards, played
// from the real catalog entries (S82 OLD batch 801).

const (
	b801SoulRendOracle   = "80d233cc-7748-41a2-bfd6-2189d0136900"
	b801DesiresOracle    = "a5827d57-ae8b-44b0-a81b-3d191a0ba2fa"
	b801VaultOracle      = "b628150b-08a1-4ea3-978d-60255dfb0b7e"
	b801GuardianOracle   = "d72b69aa-fb61-441e-9b62-4534d7a32a69"
	b801AleatoryOracle   = "da4cab2d-0068-440f-ba7c-f2d4e5fa976e"
	b801DazzlingOracle   = "de96ef93-a2e0-409d-9d25-fcb2405039b5"
	b801FlareOracle      = "ff8fae94-3a34-4bc6-aaa2-7ab7d2e2d0dd"
	b801BearTypeLine     = "Creature — Bear"
	b801GuardianTypeLine = "Creature — Gargoyle"
)

func TestS82OldB801CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b801SoulRendOracle: "Soul Rend",
		b801DesiresOracle:  "Unfulfilled Desires",
		b801VaultOracle:    "Phyrexian Vault",
		b801GuardianOracle: "Harbor Guardian",
		b801AleatoryOracle: "Aleatory",
		b801DazzlingOracle: "Dazzling Beauty",
		b801FlareOracle:    "Flare",
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

// b801DrawsAtTheNextUpkeep asserts `p` draws one card at the next turn's
// upkeep and not before.
func b801DrawsAtTheNextUpkeep(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	before := p.Hand.Size()
	next := (g.Turn.ActiveSeat + 1) % len(g.Seats)
	effectstest.AdvanceToUpkeepOf(t, g, next)
	if p.Hand.Size() != before {
		t.Fatalf("drew before the next upkeep's trigger resolved: hand %d -> %d", before, p.Hand.Size())
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p.Hand.Size() != before+1 {
		t.Errorf("hand %d -> %d at the next upkeep, want +1", before, p.Hand.Size())
	}
}

func b801Kinds(g *game.Game) string {
	out := ""
	for _, c := range g.PendingChoices {
		if c != nil {
			out += string(c.Kind) + " "
		}
	}
	return out + "stack=" + fmt.Sprint(g.Stack.Size())
}

func TestSoulRendDestroysOnlyAWhiteCreatureAndDrawsLater(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	green := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Bear", TypeLine: b801BearTypeLine, Colors: []string{"G"}, Power: 2, Toughness: 2})

	// A non-white creature is a legal target and survives; the draw still comes.
	effectstest.CastCatalogSpell(t, g, "Soul Rend", "Instant", b801SoulRendOracle, effectstest.CardTarget(green))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(green) {
		t.Fatal("a green creature was destroyed")
	}
	b801DrawsAtTheNextUpkeep(t, g, me)

	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	effectstest.SeedLibrary(me2, "A", "B", "C", "D")
	white := effectstest.ApaPush(g2, opp2.ID, opp2.ID, game.Card{Name: "Knight", TypeLine: b801BearTypeLine, Colors: []string{"W"}, Power: 2, Toughness: 2})
	effectstest.CastCatalogSpell(t, g2, "Soul Rend", "Instant", b801SoulRendOracle, effectstest.CardTarget(white))
	effectstest.PassPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(white) {
		t.Error("a white creature survived")
	}
	b801DrawsAtTheNextUpkeep(t, g2, me2)
}

func TestFlareDealsOneDamageAndDrawsLater(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Flare", "Instant", b801FlareOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life %d -> %d, want -1", life, opp.Life)
	}
	b801DrawsAtTheNextUpkeep(t, g, me)
}

func TestUnfulfilledDesiresLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	id := effectstest.PushCatalogPermanent(g, me.ID, "Unfulfilled Desires", "Enchantment", b801DesiresOracle, false)
	me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	life, hand := me.Life, me.Hand.Size()

	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no mana: accepted")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Life != life-1 {
		t.Errorf("life %d -> %d, want -1 paid as a cost", life, me.Life)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var discard *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == me.ID && c.Kind == game.PendingChoiceChooseCards {
			discard = c
		}
	}
	if discard == nil {
		t.Fatalf("no discard prompt after the draw: %s", b801Kinds(g))
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d before the discard, want +1", hand, me.Hand.Size())
	}
}

func TestPhyrexianVaultSacrificesACreatureToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	vault := effectstest.PushCatalogPermanent(g, me.ID, "Phyrexian Vault", "Artifact", b801VaultOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", b801BearTypeLine, 2, 2)
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})

	if err := g.ActivateCatalogAbility(me.ID, vault, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{land}}); err == nil {
		t.Fatal("sacrificed a land to a creature cost: accepted")
	}
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, vault, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature was not sacrificed")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, me.Hand.Size())
	}
}

func TestHarborGuardianLetsTheDefenderDraw(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.SeedLibrary(opp, "A", "B", "C", "D")
		guardian := effectstest.PushCatalogPermanent(g, me.ID, "Harbor Guardian", b801GuardianTypeLine, b801GuardianOracle, false)
		hand := opp.Hand.Size()
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(guardian, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		effectstest.LockInAttacks(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerConfirmFor(t, g, opp, yes)
		effectstest.PassPriorityAroundTable(t, g)
		want := hand
		if yes {
			want++
		}
		if opp.Hand.Size() != want {
			t.Errorf("yes=%v: the defender's hand %d -> %d, want %d", yes, hand, opp.Hand.Size(), want)
		}
	}
}

func TestDazzlingBeautyBlocksAnUnblockedAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	bear := effectstest.B12Creature(g, me.ID, "Bear", b801BearTypeLine, 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, them.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	cast := func() error {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Dazzling Beauty", TypeLine: "Instant", OracleID: b801DazzlingOracle, Owner: me.ID, Controller: me.ID})
		return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(bear)})
	}
	if err := cast(); err == nil {
		t.Fatal("cast before the declare blockers step: accepted")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	effectstest.FinishBlockDeclarations(t, g)
	if err := cast(); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	life := them.Life
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	if them.Life != life {
		t.Errorf("the defender took %d combat damage from a blocked attacker", life-them.Life)
	}
	b801DrawsAtTheNextUpkeep(t, g, me)
}

func TestAleatoryPumpsOnAWonFlipAndAlwaysDraws(t *testing.T) {
	won, lost := 0, 0
	for table := 0; table < 40 && (won == 0 || lost == 0); table++ {
		g := effectstest.NewCatalogGame(t)
		me, them := g.Seats[0], g.Seats[1]
		effectstest.SeedLibrary(me, "A", "B", "C", "D")
		bear := effectstest.B12Creature(g, me.ID, "Bear", b801BearTypeLine, 2, 2)
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(bear, them.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		cast := func() error {
			id := uuid.New()
			me.Hand.PushTop(game.Card{InstanceID: id, Name: "Aleatory", TypeLine: "Instant", OracleID: b801AleatoryOracle, Owner: me.ID, Controller: me.ID})
			return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(bear)})
		}
		if err := cast(); err == nil {
			t.Fatal("cast before blockers were declared: accepted")
		}
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		effectstest.FinishBlockDeclarations(t, g)
		if err := cast(); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
		if ask == nil {
			t.Fatal("no coin call was asked")
		}
		if err := g.ResolveCoinCall(ask.ID, me.ID, "heads"); err != nil {
			t.Fatalf("ResolveCoinCall: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		evs := effectstest.RandomEvents(g, game.EventFlipCoin)
		p, tough := effectstest.PtOf(t, g, bear)
		if evs[len(evs)-1].Won {
			won++
			if p != 3 || tough != 3 {
				t.Errorf("won flip: the bear is %d/%d, want 3/3", p, tough)
			}
		} else {
			lost++
			if p != 2 || tough != 2 {
				t.Errorf("lost flip: the bear is %d/%d, want 2/2", p, tough)
			}
		}
		b801DrawsAtTheNextUpkeep(t, g, me)
	}
	if won == 0 || lost == 0 {
		t.Fatalf("40 tables gave %d wins and %d losses", won, lost)
	}
}
