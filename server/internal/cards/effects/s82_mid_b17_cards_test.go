package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	midB17Shadow   = "7ecc506e-bed0-4ae2-9d2f-5143ef63f62c"
	midB17Rem      = "21971c8b-ff9b-40f8-8c9a-93d26c2615b7"
	midB17Ludevic  = ludevicOracle
	midB17Jerren   = jerrenOracle
	midB17Hostel   = hostileHostelOracle
	midB17Channel  = flameChannelerOracle
	midB17Daybreak = "280624aa-5f9a-48fd-85ea-815c96c747b3"
)

func TestMidB17CardsAreFullWithNoCaveats(t *testing.T) {
	want := map[string]string{
		midB17Shadow: "Shadow Kin", midB17Rem: "Rem Karolus, Stalwart Slayer",
		midB17Ludevic: "Ludevic, Necrogenius", midB17Ludevic + "#1": "Olag, Ludevic's Hubris",
		midB17Jerren: "Jerren, Corrupted Bishop", midB17Jerren + "#1": "Ormendahl, the Corrupter",
		midB17Hostel: "Hostile Hostel", midB17Hostel + "#1": "Creeping Inn",
		midB17Channel: "Flame Channeler", midB17Channel + "#1": "Embodiment of Flame",
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

func midB17ChannelerRow() cards.Card {
	return cards.Card{
		ID: uuid.New(), Name: "Flame Channeler // Embodiment of Flame", Layout: "transform",
		TypeLine: "Creature — Human Wizard // Creature — Elemental Wizard",
		OracleID: uuid.MustParse(midB17Channel), Keywords: []string{"Transform"},
		CardFaces: []cards.CardFace{
			{Name: "Flame Channeler", TypeLine: "Creature — Human Wizard", ManaCost: "{1}{R}", Colors: []string{"R"}, Power: "2", Toughness: "2",
				OracleText: "When a spell you control deals damage, transform this creature."},
			{Name: "Embodiment of Flame", TypeLine: "Creature — Elemental Wizard", Colors: []string{"R"}, Power: "3", Toughness: "3",
				OracleText: "Whenever a spell you control deals damage, put a flame counter on this creature.\n{1}, Remove a flame counter from this creature: Exile the top card of your library. You may play that card this turn."},
		},
	}
}

// midB17Put puts an imported card onto p's battlefield, front face up.
func midB17Put(t *testing.T, g *game.Game, p *game.Player, row cards.Card) uuid.UUID {
	t.Helper()
	id := importToHand(row, p)
	c, ok := cardInZone(p.Hand, id)
	if !ok {
		t.Fatal("card did not reach the hand")
	}
	p.Hand.Remove(id)
	pushBattlefieldCardWithTimestamp(g, c)
	return id
}

func midB17Bolt(t *testing.T, g *game.Game, target game.TargetRef) {
	t.Helper()
	b12Bolt(t, g, target)
	monSettle(t, g)
}

func TestMidB17FlameChannelerTransformsOnASpellsDamageAndCountsAfter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := midB17Put(t, g, me, midB17ChannelerRow())
	midB17Bolt(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if f := iwFace(g, id); f != 1 {
		t.Fatalf("Flame Channeler is on face %d after a spell dealt damage, want 1", f)
	}
	if c, _ := battlefieldCard(g, id); c.Counters["flame"] != 0 {
		t.Fatalf("transforming gave %d flame counters", c.Counters["flame"])
	}
	midB17Bolt(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if c, _ := battlefieldCard(g, id); c.Counters["flame"] != 1 {
		t.Fatalf("flame counters = %d after a second spell, want 1", c.Counters["flame"])
	}
}

func TestMidB17FlameChannelerIgnoresAPermanentsDamageAndAnOpponentsSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := midB17Put(t, g, me, midB17ChannelerRow())
	flyer := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Bird",
		TypeLine: "Creature — Bird", Power: 1, Toughness: 3, Keywords: []string{"Flying"}, Owner: opp.ID, Controller: opp.ID})
	ranger := b12Push(g, me.ID, "Daybreak Ranger", "Creature — Human Archer Ranger Werewolf", midB17Daybreak, 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, ranger, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: flyer}}, Strict: true}); err != nil {
		t.Fatalf("ranger ping: %v", err)
	}
	passPriorityAroundTable(t, g)
	monSettle(t, g)
	if f := iwFace(g, id); f != 0 {
		t.Fatalf("a creature ability's damage transformed the Channeler (face %d)", f)
	}
}
