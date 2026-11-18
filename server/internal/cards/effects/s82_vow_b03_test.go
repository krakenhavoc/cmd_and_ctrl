package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	s82v03Guest      = "417e798d-9ca6-4cd0-adb4-f61d3becb201"
	s82v03Petal      = "219cf182-0942-4569-91e9-9bc9bb1d06d0"
	s82v03Servitor   = "5a3f9c4d-20a8-46aa-9cf3-3916f56a7906"
	s82v03Celebrants = "3c0be888-0d66-4bad-84f4-90c3934915d7"
	s82v03Epicure    = "cef82674-1f61-407b-801b-a76a3803e598"
	s82v03Fountain   = "abfa5753-8c02-4e2f-8f98-b0636907293a"
	s82v03Knife      = "db221948-fcb9-4ef9-ae69-70266235dd04"
	s82v03Betrayal   = "a2743cb6-564d-43f1-99cc-4ba3ceb2e282"
	s82v03Ritual     = "56c78349-e4fa-48ba-ae8e-d2312ae6edc6"
	s82v03Discussion = "aff9a568-2848-4614-a28b-b1ea84f7f0b7"
	s82v03Essence    = "e6e7b5ba-ea10-4f53-8afc-045aa1906c6b"
	s82v03Kiss       = "105cea6e-18e9-4dec-a533-0dca4d41320b"
	s82v03Prison     = "574230c7-b949-4c90-b107-c26458a6e469"
)

func s82v03Blood(g *game.Game, p *game.Player) int {
	n, _ := countTokensNamed(g, p.ID, "Blood")
	return n
}

func TestS82Vow03Registered(t *testing.T) {
	for oracle, name := range map[string]string{
		s82v03Guest: "Belligerent Guest", s82v03Petal: "Blood Petal Celebrant",
		s82v03Servitor: "Blood Servitor", s82v03Celebrants: "Falkenrath Celebrants",
		s82v03Epicure: "Voldaren Epicure", s82v03Fountain: "Blood Fountain",
		s82v03Knife: "Ceremonial Knife", s82v03Betrayal: "Bloody Betrayal",
		s82v03Ritual: "Grisly Ritual", s82v03Discussion: "Pointed Discussion",
		s82v03Essence: "Syphon Essence", s82v03Kiss: "Vampire's Kiss",
		s82v03Prison: "Sigarda's Imprisonment",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v spec=%+v", name, ok, spec.Name)
		}
	}
}

func TestS82Vow03BelligerentGuestMakesBloodOnCombatDamageToAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guest := pushCatalogPermanent(g, me.ID, "Belligerent Guest", "Creature — Vampire", s82v03Guest, false)
	if !hasKeywordOf(g, guest, "trample") {
		t.Error("no trample")
	}
	deal := func(target uuid.UUID, combat bool) {
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: guest, Target: target, Amount: 3, Combat: combat})
		})
		passPriorityAroundTable(t, g)
	}
	deal(opp.ID, false)
	if n := s82v03Blood(g, me); n != 0 {
		t.Fatalf("noncombat damage made %d Blood", n)
	}
	deal(opp.ID, true)
	if n := s82v03Blood(g, me); n != 1 {
		t.Errorf("Blood = %d, want 1", n)
	}
}

func TestS82Vow03BelligerentGuestInRealCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guest := pushCatalogPermanent(g, me.ID, "Belligerent Guest", "Creature — Vampire", s82v03Guest, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == guest {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 3, 2
			}
		}
	})
	life := opp.Life
	attackWith(t, g, opp.ID, guest)
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 || s82v03Blood(g, me) != 1 {
		t.Errorf("life %d→%d, Blood %d; want -3 and 1 Blood", life, opp.Life, s82v03Blood(g, me))
	}
}

func TestS82Vow03BloodPetalCelebrantFirstStrikeOnlyWhileAttackingAndBloodOnDeath(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cel := pushCatalogPermanent(g, me.ID, "Blood Petal Celebrant", "Creature — Vampire", s82v03Petal, false)
	if hasKeywordOf(g, cel, "first strike") {
		t.Fatal("first strike while not attacking")
	}
	declareAttack(t, g, opp.ID, cel)
	if !hasKeywordOf(g, cel, "first strike") {
		t.Error("no first strike while attacking")
	}
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	cel2 := pushCatalogPermanent(g2, me2.ID, "Blood Petal Celebrant", "Creature — Vampire", s82v03Petal, false)
	if err := g2.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me2.ID}, cel2); err != nil {
		t.Fatalf("kill: %v", err)
	}
	passPriorityAroundTable(t, g2)
	if n := s82v03Blood(g2, me2); n != 1 {
		t.Errorf("Blood after dying = %d, want 1", n)
	}
}

func TestS82Vow03EntersMakeBlood(t *testing.T) {
	for _, c := range []struct {
		name, typeLine, oracle string
		blood                  int
	}{
		{"Blood Servitor", "Artifact Creature — Construct", s82v03Servitor, 1},
		{"Falkenrath Celebrants", "Creature — Vampire", s82v03Celebrants, 2},
		{"Voldaren Epicure", "Creature — Vampire", s82v03Epicure, 1},
		{"Blood Fountain", "Artifact", s82v03Fountain, 1},
	} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		var lives []int
		for _, p := range g.Seats[1:] {
			lives = append(lives, p.Life)
		}
		toMainForCost(t, g)
		castFromSeat(t, g, me, c.name, c.typeLine, c.oracle, nil)
		passPriorityAroundTable(t, g)
		if n := s82v03Blood(g, me); n != c.blood {
			t.Errorf("%s: Blood = %d, want %d", c.name, n, c.blood)
		}
		wantDmg := 0
		if c.name == "Voldaren Epicure" {
			wantDmg = 1
		}
		for i, p := range g.Seats[1:] {
			if p.Life != lives[i]-wantDmg {
				t.Errorf("%s: opponent %d life %d→%d, want -%d", c.name, i+1, lives[i], p.Life, wantDmg)
			}
		}
	}
}

func TestS82Vow03BloodFountainReturnsUpToTwoCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fountain := pushCatalogPermanent(g, me.ID, "Blood Fountain", "Artifact", s82v03Fountain, false)
	a := b17GraveyardCard(me, "Dead A", "Creature — Bear", "{1}{G}")
	b := b17GraveyardCard(me, "Dead B", "Creature — Bear", "{1}{G}")
	c := b17GraveyardCard(me, "Dead C", "Creature — Bear", "{1}{G}")
	spell := b17GraveyardCard(me, "Not a creature", "Sorcery", "{1}")
	toMainForCost(t, g)
	// Strict: the {3}{B} is not in the pool.
	if err := g.ActivateCatalogAbility(me.ID, fountain, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}},
	}); err == nil {
		t.Fatal("activated with an empty mana pool")
	}
	for _, m := range []string{"{B}", "{C}", "{C}", "{C}"} {
		if err := g.AddManaForEffect(me.ID, uuid.Nil, m); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, fountain, 0, game.ActivateAbilityParams{
		Strict: true,
		Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}, {Kind: game.TargetCard, ID: c},
		},
	}); err == nil {
		t.Fatal("three targets accepted for 'up to two'")
	}
	if err := g.ActivateCatalogAbility(me.ID, fountain, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: spell}},
	}); err == nil {
		t.Fatal("a noncreature card was targeted")
	}
	if err := g.ActivateCatalogAbility(me.ID, fountain, 0, game.ActivateAbilityParams{
		Strict: true,
		Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b},
		},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) || me.Hand.Contains(c) {
		t.Error("expected exactly the two targeted creatures back in hand")
	}
	if g.Battlefield.Contains(fountain) {
		t.Error("the Fountain was not sacrificed")
	}
}

func TestS82Vow03CeremonialKnifePumpsAndMakesBloodOncePerCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	knife := pushCatalogPermanent(g, me.ID, "Ceremonial Knife", "Artifact — Equipment", s82v03Knife, false)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatal(err)
	}
	equipTo(t, g, me.ID, knife, bear)
	if p, tough := ptOf(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("equipped Bear %d/%d, want 3/2", p, tough)
	}
	life := opp.Life
	attackWith(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d→%d, want -3", life, opp.Life)
	}
	if n := s82v03Blood(g, me); n != 1 {
		t.Errorf("Blood = %d, want 1", n)
	}
	// The grant goes with the Equipment: a bare creature makes none.
	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	plain := b12Creature(g2, me2.ID, "Bear", "Creature — Bear", 2, 2)
	attackWith(t, g2, opp2.ID, plain)
	passPriorityAroundTable(t, g2)
	if n := s82v03Blood(g2, me2); n != 0 {
		t.Errorf("an unequipped creature made %d Blood", n)
	}
}

func TestS82Vow03BloodyBetrayalStealsUntapsHastesAndMakesBlood(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	castCatalogSpell(t, g, "Bloody Betrayal", "Sorcery", s82v03Betrayal, b16TargetCard(bear))
	passPriorityAroundTable(t, g)
	c := e2Card(t, g, bear)
	if c.Controller != me.ID || c.Tapped || !hasKeywordOf(g, bear, "haste") {
		t.Errorf("controller=%v tapped=%v haste=%v", c.Controller == me.ID, c.Tapped, hasKeywordOf(g, bear, "haste"))
	}
	if n := s82v03Blood(g, me); n != 1 {
		t.Errorf("Blood = %d, want 1", n)
	}
	// An illegal target is refused at cast.
	g2 := newCatalogGame(t)
	if err := g2.CastSpell(g2.Seats[0].ID, b13HandCard(g2.Seats[0], "Bloody Betrayal", "Sorcery", s82v03Betrayal),
		game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: g2.Seats[1].ID}}}); err == nil {
		t.Error("cast targeting a player")
	}
}

func TestS82Vow03GrislyRitualDestroysAndMakesTwoBlood(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Grisly Ritual", "Sorcery", s82v03Ritual, b16TargetCard(bear))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature survived")
	}
	if n := s82v03Blood(g, me); n != 2 {
		t.Errorf("Blood = %d, want 2", n)
	}
	g2 := newCatalogGame(t)
	land := b12Permanent(g2, g2.Seats[1].ID, "Plains", "Basic Land — Plains")
	if err := g2.CastSpell(g2.Seats[0].ID, b13HandCard(g2.Seats[0], "Grisly Ritual", "Sorcery", s82v03Ritual),
		game.CastSpellParams{Strict: true, Targets: b16TargetCard(land)}); err == nil {
		t.Error("cast targeting a land")
	}
}

func TestS82Vow03PointedDiscussionDrawsTwoLosesTwoMakesBlood(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	hand, life := me.Hand.Size(), me.Life
	castFromSeat(t, g, me, "Pointed Discussion", "Sorcery", s82v03Discussion, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2-0 {
		// The spell itself left the hand, so net +2 draws.
		t.Errorf("hand grew by %d, want 2", got)
	}
	if me.Life != life-2 {
		t.Errorf("life %d→%d, want -2", life, me.Life)
	}
	if n := s82v03Blood(g, me); n != 1 {
		t.Errorf("Blood = %d, want 1", n)
	}
}

func TestS82Vow03SyphonEssenceCountersCreatureOrPlaneswalkerSpellsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b13PlayAs(t, g, 1, "Their Bear", "Creature — Bear", "")
	if err := g.CastSpell(me.ID, b13HandCard(me, "Syphon Essence", "Instant", s82v03Essence),
		game.CastSpellParams{Targets: b16TargetCard(bear)}); err != nil {
		t.Fatalf("Syphon Essence: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Graveyard.Contains(bear) {
		t.Error("the creature spell was not countered")
	}
	if n := s82v03Blood(g, me); n != 1 {
		t.Errorf("Blood = %d, want 1", n)
	}
	// A noncreature spell is not a legal target.
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	sorc := b13PlayAs(t, g2, 1, "Their Sorcery", "Sorcery", "")
	if err := g2.CastSpell(me2.ID, b13HandCard(me2, "Syphon Essence", "Instant", s82v03Essence),
		game.CastSpellParams{Strict: true, Targets: b16TargetCard(sorc)}); err == nil {
		t.Error("targeted a sorcery spell")
	}
}

func TestS82Vow03VampiresKissDrainsTwoAndMakesTwoBlood(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life, theirs := me.Life, opp.Life
	castCatalogSpell(t, g, "Vampire's Kiss", "Sorcery", s82v03Kiss,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if me.Life != life+2 || opp.Life != theirs-2 {
		t.Errorf("life me %d→%d, them %d→%d", life, me.Life, theirs, opp.Life)
	}
	if n := s82v03Blood(g, me); n != 2 {
		t.Errorf("Blood = %d, want 2", n)
	}
}

func TestS82Vow03SigardasImprisonmentStopsAttackAndBlockThenExilesForBlood(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	aura := castCatalogSpell(t, g, "Sigarda's Imprisonment", "Enchantment — Aura", s82v03Prison, b16TargetCard(bear))
	passPriorityAroundTable(t, g)
	if r := restrictionsOf(t, g, bear); r&game.CantAttackOrBlock != game.CantAttackOrBlock {
		t.Fatalf("restrictions = %v, want can't attack or block", r)
	}
	// Strict: no {4}{W} in the pool.
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with an empty pool")
	}
	for _, m := range []string{"{W}", "{C}", "{C}", "{C}", "{C}"} {
		if err := g.AddManaForEffect(me.ID, uuid.Nil, m); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the enchanted creature was not exiled")
	}
	if n := s82v03Blood(g, me); n != 1 {
		t.Errorf("Blood = %d, want 1", n)
	}
}
