package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	vowB08Retrieval  = "8651da8e-9b2c-4530-a521-aea490ad88d4"
	vowB08DigUp      = "ae96fe7e-8d89-4f44-9cc4-3c9b124fdc4b"
	vowB08Retributn  = "83094dd2-b8e9-4753-a8a0-657b446def60"
	vowB08Flare      = "06a1c404-d76f-4206-8740-7ad7488fac33"
	vowB08Lunar      = "d2f693d9-f338-41d6-b7e2-ddaa5f4bcc3b"
	vowB08Grasp      = "d9763c75-75f7-4766-8cca-84c899796e3c"
	vowB08Path       = "d41987b0-2ea0-4cc7-8608-70c486db76eb"
	vowB08Portent    = "378567f8-7aef-42b6-853e-8591641052e8"
	vowB08InstantTyp = "Instant"
)

// vowB08Costs are the printed mana costs, so a hard cast is priced.
var vowB08Costs = map[string]string{
	vowB08Retrieval: "{U}", vowB08DigUp: "{G}", vowB08Retributn: "{1}{W}", vowB08Flare: "{1}{W}",
	vowB08Lunar: "{1}{U}", vowB08Grasp: "{1}{B}", vowB08Path: "{1}{B}{B}", vowB08Portent: "{1}{U}{U}",
}

// vowB08Cast seeds the card in the active seat's hand with the pool it
// needs and casts it strictly, so a refusal is a rules refusal and not a
// free cast the engine waved through.
func vowB08Cast(t *testing.T, g *game.Game, name, typeLine, oracle string, mana []game.ManaToken, p game.CastSpellParams) error {
	t.Helper()
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: vowB08Costs[oracle], Owner: me.ID, Controller: me.ID})
	me.ManaPool.EmptyPool()
	me.ManaPool.AddMana(mana...)
	p.Strict = true
	return g.CastSpell(me.ID, id, p)
}

func vowB08Mana(s string) []game.ManaToken {
	var out []game.ManaToken
	for _, r := range s {
		if r >= '1' && r <= '9' {
			for i := 0; i < int(r-'0'); i++ {
				out = append(out, game.ManaToken{Color: "C"})
			}
			continue
		}
		out = append(out, game.ManaToken{Color: string(r)})
	}
	return out
}

func vowB08Tgt(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func vowB08OnBF(g *game.Game, id uuid.UUID) bool { return g.Battlefield.Contains(id) }

func TestVowB08AlchemistsRetrievalCleaveWidensToAnyNonland(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	land := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")

	if err := vowB08Cast(t, g, "Alchemist's Retrieval", vowB08InstantTyp, vowB08Retrieval, vowB08Mana("U"),
		game.CastSpellParams{Targets: vowB08Tgt(theirs)}); err == nil {
		t.Fatal("hard-cast must not target a permanent you do not control")
	}
	if err := vowB08Cast(t, g, "Alchemist's Retrieval", vowB08InstantTyp, vowB08Retrieval, vowB08Mana("U"),
		game.CastSpellParams{Targets: vowB08Tgt(land)}); err == nil {
		t.Fatal("must not target a land")
	}
	// A cleave cast for {U} alone cannot pay {1}{U}.
	if err := vowB08Cast(t, g, "Alchemist's Retrieval", vowB08InstantTyp, vowB08Retrieval, vowB08Mana("U"),
		game.CastSpellParams{AlternativeCost: "cleave", Targets: vowB08Tgt(theirs)}); err == nil {
		t.Fatal("cleave costs {1}{U}; one blue is not enough")
	}
	if err := vowB08Cast(t, g, "Alchemist's Retrieval", vowB08InstantTyp, vowB08Retrieval, vowB08Mana("U"),
		game.CastSpellParams{Targets: vowB08Tgt(mine)}); err != nil {
		t.Fatalf("hard-cast on own creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if vowB08OnBF(g, mine) || !me.Hand.Contains(mine) {
		t.Error("own creature did not return to its owner's hand")
	}

	if err := vowB08Cast(t, g, "Alchemist's Retrieval", vowB08InstantTyp, vowB08Retrieval, vowB08Mana("UU"),
		game.CastSpellParams{AlternativeCost: "cleave", Targets: vowB08Tgt(theirs)}); err != nil {
		t.Fatalf("cleaved on an opponent's creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if vowB08OnBF(g, theirs) || !opp.Hand.Contains(theirs) {
		t.Error("opponent's creature did not return to its owner's hand")
	}
}

func TestVowB08DigUpFetchesBasicLandOrAnyCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedSearchLibrary(me,
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	_ = ids
	if err := vowB08Cast(t, g, "Dig Up", "Sorcery", vowB08DigUp, vowB08Mana("G"), game.CastSpellParams{}); err != nil {
		t.Fatalf("hard-cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Bolt") != uuid.Nil {
		t.Error("hard-cast Dig Up offered a nonland card")
	}
	answerSearchNamed(t, g, me.ID, "Forest")
	passPriorityAroundTable(t, g)
	if !vowB08HandHas(me, "Forest") {
		t.Error("the basic land did not reach the hand")
	}

	seedSearchLibrary(me,
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	if err := vowB08Cast(t, g, "Dig Up", "Sorcery", vowB08DigUp, vowB08Mana("GBBG"), game.CastSpellParams{AlternativeCost: "cleave"}); err != nil {
		t.Fatalf("cleaved: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerSearchNamed(t, g, me.ID, "Bolt")
	passPriorityAroundTable(t, g)
	if !vowB08HandHas(me, "Bolt") {
		t.Error("cleaved Dig Up did not fetch the nonland card")
	}
}

func vowB08HandHas(p *game.Player, name string) bool {
	for _, c := range p.Hand.Cards {
		if c.Name == name {
			return true
		}
	}
	return false
}

func TestVowB08FierceRetributionCleaveHitsNonAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	idle := b12Creature(g, opp.ID, "Idle", "Creature — Bear", 2, 2)
	attacker := b12Creature(g, opp.ID, "Raider", "Creature — Bear", 2, 2)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == attacker {
			g.Battlefield.Cards[i].AttackingTarget = me.ID
		}
	}
	if err := vowB08Cast(t, g, "Fierce Retribution", vowB08InstantTyp, vowB08Retributn, vowB08Mana("1W"),
		game.CastSpellParams{Targets: vowB08Tgt(idle)}); err == nil {
		t.Fatal("hard-cast must not target a creature that is not attacking")
	}
	if err := vowB08Cast(t, g, "Fierce Retribution", vowB08InstantTyp, vowB08Retributn, vowB08Mana("1W"),
		game.CastSpellParams{Targets: vowB08Tgt(attacker)}); err != nil {
		t.Fatalf("hard-cast on an attacker: %v", err)
	}
	passPriorityAroundTable(t, g)
	if vowB08OnBF(g, attacker) {
		t.Error("attacking creature survived")
	}
	if err := vowB08Cast(t, g, "Fierce Retribution", vowB08InstantTyp, vowB08Retributn, vowB08Mana("5W"),
		game.CastSpellParams{AlternativeCost: "cleave", Targets: vowB08Tgt(idle)}); err != nil {
		t.Fatalf("cleaved on a non-attacker: %v", err)
	}
	passPriorityAroundTable(t, g)
	if vowB08OnBF(g, idle) {
		t.Error("cleaved Fierce Retribution left the creature")
	}
}

func TestVowB08LanternFlareCountsCreaturesHardCastAndUsesXCleaved(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
	b12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
	b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	big := b12Creature(g, opp.ID, "Big", "Creature — Giant", 3, 3)
	life := me.Life
	if err := vowB08Cast(t, g, "Lantern Flare", vowB08InstantTyp, vowB08Flare, vowB08Mana("1W"),
		game.CastSpellParams{Targets: vowB08Tgt(big)}); err != nil {
		t.Fatalf("hard-cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	// Two creatures of mine: 2 damage, 2 life.
	if me.Life != life+2 {
		t.Errorf("life %d, want %d (X is the creatures you control = 2)", me.Life, life+2)
	}
	if c, ok := battlefieldCard(g, big); !ok || c.DamageMarked != 2 {
		t.Errorf("damage marked on the Giant: %v", ok)
	}
	// Cleaved with X=3: three damage and three life regardless of the board.
	life = me.Life
	if err := vowB08Cast(t, g, "Lantern Flare", vowB08InstantTyp, vowB08Flare, vowB08Mana("RWCCC"),
		game.CastSpellParams{AlternativeCost: "cleave", XValue: 3, Targets: vowB08Tgt(big)}); err != nil {
		t.Fatalf("cleaved X=3: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d (cleaved X=3)", me.Life, life+3)
	}
	if vowB08OnBF(g, big) {
		t.Error("the Giant took 2+3 damage and should have died")
	}
	// Refusal: a player is not a legal target.
	if err := vowB08Cast(t, g, "Lantern Flare", vowB08InstantTyp, vowB08Flare, vowB08Mana("1W"),
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Error("Lantern Flare targeted a player")
	}
}

func TestVowB08LunarRejectionCleaveWidensAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	wolf := b12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 2, 2)
	were := b12Creature(g, opp.ID, "Were", "Creature — Human Werewolf", 2, 2)
	if err := vowB08Cast(t, g, "Lunar Rejection", vowB08InstantTyp, vowB08Lunar, vowB08Mana("1U"),
		game.CastSpellParams{Targets: vowB08Tgt(bear)}); err == nil {
		t.Fatal("hard-cast must not target a Bear")
	}
	for _, id := range []uuid.UUID{wolf, were} {
		hand := me.Hand.Size()
		if err := vowB08Cast(t, g, "Lunar Rejection", vowB08InstantTyp, vowB08Lunar, vowB08Mana("1U"),
			game.CastSpellParams{Targets: vowB08Tgt(id)}); err != nil {
			t.Fatalf("hard-cast: %v", err)
		}
		passPriorityAroundTable(t, g)
		if vowB08OnBF(g, id) {
			t.Error("Wolf or Werewolf stayed on the battlefield")
		}
		if me.Hand.Size() != hand+1 {
			t.Errorf("hand %d, want %d (the spell was cast from it, then one draw)", me.Hand.Size(), hand+1)
		}
	}
	hand := me.Hand.Size()
	if err := vowB08Cast(t, g, "Lunar Rejection", vowB08InstantTyp, vowB08Lunar, vowB08Mana("3U"),
		game.CastSpellParams{AlternativeCost: "cleave", Targets: vowB08Tgt(bear)}); err != nil {
		t.Fatalf("cleaved on a Bear: %v", err)
	}
	passPriorityAroundTable(t, g)
	if vowB08OnBF(g, bear) || me.Hand.Size() != hand+1 {
		t.Error("cleaved Lunar Rejection did not bounce and draw")
	}
}

func TestVowB08ParasiticGraspHumanOrAnyCreatureAndGainsLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	human := b12Creature(g, opp.ID, "Peasant", "Creature — Human Peasant", 2, 3)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 3)
	if err := vowB08Cast(t, g, "Parasitic Grasp", vowB08InstantTyp, vowB08Grasp, vowB08Mana("1B"),
		game.CastSpellParams{Targets: vowB08Tgt(bear)}); err == nil {
		t.Fatal("hard-cast must not target a non-Human")
	}
	life := me.Life
	if err := vowB08Cast(t, g, "Parasitic Grasp", vowB08InstantTyp, vowB08Grasp, vowB08Mana("1B"),
		game.CastSpellParams{Targets: vowB08Tgt(human)}); err != nil {
		t.Fatalf("hard-cast on a Human: %v", err)
	}
	passPriorityAroundTable(t, g)
	if vowB08OnBF(g, human) || me.Life != life+3 {
		t.Errorf("Human alive=%v life %d, want dead and %d", vowB08OnBF(g, human), me.Life, life+3)
	}
	if err := vowB08Cast(t, g, "Parasitic Grasp", vowB08InstantTyp, vowB08Grasp, vowB08Mana("1BB"),
		game.CastSpellParams{AlternativeCost: "cleave", Targets: vowB08Tgt(bear)}); err != nil {
		t.Fatalf("cleaved on a Bear: %v", err)
	}
	passPriorityAroundTable(t, g)
	if vowB08OnBF(g, bear) || me.Life != life+6 {
		t.Errorf("cleaved: Bear alive=%v life %d, want dead and %d", vowB08OnBF(g, bear), me.Life, life+6)
	}
}

func TestVowB08PathOfPerilSweepsSmallCreaturesOrAll(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mk := func(owner uuid.UUID, name, cost string) uuid.UUID {
		return pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear", ManaCost: cost,
			Power: 2, Toughness: 2, Owner: owner, Controller: owner,
		})
	}
	small, token := mk(me.ID, "Small", "{1}{B}"), mk(opp.ID, "Token", "")
	big, bigger := mk(opp.ID, "Big", "{2}{G}"), mk(me.ID, "Bigger", "{4}")
	land := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")

	if err := vowB08Cast(t, g, "Path of Peril", "Sorcery", vowB08Path, vowB08Mana("1B"), game.CastSpellParams{}); err == nil {
		t.Fatal("{1}{B}{B} is not payable with {1}{B}")
	}
	if err := vowB08Cast(t, g, "Path of Peril", "Sorcery", vowB08Path, vowB08Mana("1BB"), game.CastSpellParams{}); err != nil {
		t.Fatalf("hard-cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if vowB08OnBF(g, small) || vowB08OnBF(g, token) {
		t.Error("mana value 2 or less (and a token) should have died")
	}
	if !vowB08OnBF(g, big) || !vowB08OnBF(g, bigger) || !vowB08OnBF(g, land) {
		t.Error("hard-cast Path of Peril destroyed more than mana value 2 or less")
	}
	if err := vowB08Cast(t, g, "Path of Peril", "Sorcery", vowB08Path, vowB08Mana("4WB"), game.CastSpellParams{AlternativeCost: "cleave"}); err != nil {
		t.Fatalf("cleaved: %v", err)
	}
	passPriorityAroundTable(t, g)
	if vowB08OnBF(g, big) || vowB08OnBF(g, bigger) {
		t.Error("cleaved Path of Peril left a creature")
	}
	if !vowB08OnBF(g, land) {
		t.Error("Path of Peril destroyed a land")
	}
}

func TestVowB08WingedPortentCountsFliersOrAllCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Flier", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Keywords: []string{"flying"}, Owner: me.ID, Controller: me.ID})
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Flier 2", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Keywords: []string{"flying"}, Owner: me.ID, Controller: me.ID})
	b12Creature(g, me.ID, "Walker", "Creature — Bear", 2, 2)
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Flier", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})

	hand := me.Hand.Size()
	if err := vowB08Cast(t, g, "Winged Portent", vowB08InstantTyp, vowB08Portent, vowB08Mana("1UU"), game.CastSpellParams{}); err != nil {
		t.Fatalf("hard-cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2 (one per creature with flying you control; the spell left the hand)", got)
	}
	hand = me.Hand.Size()
	if err := vowB08Cast(t, g, "Winged Portent", vowB08InstantTyp, vowB08Portent, vowB08Mana("4GU"), game.CastSpellParams{AlternativeCost: "cleave"}); err != nil {
		t.Fatalf("cleaved: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("cleaved drew %d, want 3 (every creature you control)", got)
	}
}
