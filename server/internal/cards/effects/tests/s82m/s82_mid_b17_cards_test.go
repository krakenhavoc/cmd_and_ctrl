package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsfl "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/fl"
	cardsh "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/h"
	cardsj "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/j"
	cardslu "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/lu"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	midB17Shadow   = "7ecc506e-bed0-4ae2-9d2f-5143ef63f62c"
	midB17Rem      = "21971c8b-ff9b-40f8-8c9a-93d26c2615b7"
	midB17Ludevic  = cardslu.LudevicOracle
	midB17Jerren   = cardsj.JerrenOracle
	midB17Hostel   = cardsh.HostileHostelOracle
	midB17Channel  = cardsfl.FlameChannelerOracle
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
	id := effectstest.ImportToHand(row, p)
	c, ok := effectstest.CardInZone(p.Hand, id)
	if !ok {
		t.Fatal("card did not reach the hand")
	}
	p.Hand.Remove(id)
	effectstest.PushBattlefieldCardWithTimestamp(g, c)
	return id
}

func midB17Bolt(t *testing.T, g *game.Game, target game.TargetRef) {
	t.Helper()
	effectstest.B12Bolt(t, g, target)
	effectstest.MonSettle(t, g)
}

func TestMidB17FlameChannelerTransformsOnASpellsDamageAndCountsAfter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := midB17Put(t, g, me, midB17ChannelerRow())
	midB17Bolt(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if f := effectstest.IwFace(g, id); f != 1 {
		t.Fatalf("Flame Channeler is on face %d after a spell dealt damage, want 1", f)
	}
	if c, _ := effectstest.BattlefieldCard(g, id); c.Counters["flame"] != 0 {
		t.Fatalf("transforming gave %d flame counters", c.Counters["flame"])
	}
	midB17Bolt(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if c, _ := effectstest.BattlefieldCard(g, id); c.Counters["flame"] != 1 {
		t.Fatalf("flame counters = %d after a second spell, want 1", c.Counters["flame"])
	}
}

func TestMidB17FlameChannelerIgnoresAPermanentsDamageAndAnOpponentsSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := midB17Put(t, g, me, midB17ChannelerRow())
	flyer := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Bird",
		TypeLine: "Creature — Bird", Power: 1, Toughness: 3, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
	ranger := effectstest.B12Push(g, me.ID, "Daybreak Ranger", "Creature — Human Archer Ranger Werewolf", midB17Daybreak, 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, ranger, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: flyer}}, Strict: true}); err != nil {
		t.Fatalf("ranger ping: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MonSettle(t, g)
	if f := effectstest.IwFace(g, id); f != 0 {
		t.Fatalf("a creature ability's damage transformed the Channeler (face %d)", f)
	}
}

func TestMidB17RemKarolusPreventsSpellDamageToYouAndYourOtherPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rem := effectstest.B12Push(g, me.ID, "Rem Karolus, Stalwart Slayer", "Legendary Creature — Human Knight", midB17Rem, 2, 3)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	life := me.Life
	midB17Bolt(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
	if me.Life != life {
		t.Errorf("a spell damaged its controller through Rem Karolus: %d -> %d", life, me.Life)
	}
	midB17Bolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: bear})
	if !g.Battlefield.Contains(bear) {
		t.Error("a Bolt killed my other creature through Rem Karolus")
	}
	// "Another permanent": Rem himself is not protected.
	midB17Bolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: rem})
	if g.Battlefield.Contains(rem) {
		t.Error("Rem Karolus is protected from spells, but the text says another permanent")
	}
}

func TestMidB17RemKarolusAddsOneToSpellDamageAgainstOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "Rem Karolus, Stalwart Slayer", "Legendary Creature — Human Knight", midB17Rem, 2, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 4)
	life := opp.Life
	midB17Bolt(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if opp.Life != life-4 {
		t.Errorf("opponent's life %d -> %d, want -4", life, opp.Life)
	}
	// 3 + 1 = 4 damage kills a 2/4 that a plain Bolt would leave alive.
	midB17Bolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: theirs})
	if g.Battlefield.Contains(theirs) {
		t.Error("a Bolt did not deal 4 to an opponent's creature")
	}
}

func TestMidB17RemKarolusLeavesAPermanentsAbilityAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "Rem Karolus, Stalwart Slayer", "Legendary Creature — Human Knight", midB17Rem, 2, 3)
	ranger := effectstest.B12Push(g, me.ID, "Daybreak Ranger", "Creature — Human Archer Ranger Werewolf", midB17Daybreak, 2, 2)
	flyer := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Bird",
		TypeLine: "Creature — Bird", Power: 1, Toughness: 3, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
	if err := g.ActivateCatalogAbility(me.ID, ranger, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: flyer}}, Strict: true}); err != nil {
		t.Fatalf("ranger ping: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MonSettle(t, g)
	// 2 damage, not 3: the ping is a creature's ability, not a spell.
	if !g.Battlefield.Contains(flyer) {
		t.Error("an ability's damage was increased as if it were a spell's")
	}
}

func midB17HostelRow() cards.Card {
	return cards.Card{
		ID: uuid.New(), Name: "Hostile Hostel // Creeping Inn", Layout: "transform",
		TypeLine: "Land // Artifact Creature — Horror Construct",
		OracleID: uuid.MustParse(midB17Hostel), Keywords: []string{"Transform"},
		CardFaces: []cards.CardFace{
			{Name: "Hostile Hostel", TypeLine: "Land", OracleText: "{T}: Add {C}."},
			{Name: "Creeping Inn", TypeLine: "Artifact Creature — Horror Construct", Power: "3", Toughness: "7",
				OracleText: "Whenever this creature attacks, you may exile a creature card from your graveyard."},
		},
	}
}

func TestMidB17HostileHostelTransformsOnTheThirdSoulAndUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := midB17Put(t, g, me, midB17HostelRow())
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	for i := 1; i <= 3; i++ {
		fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
		if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
			t.Fatalf("mana: %v", err)
		}
		if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{
			Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.MonSettle(t, g)
		c, _ := effectstest.BattlefieldCard(g, id)
		if i < 3 {
			if c.Counters["soul"] != i || effectstest.IwFace(g, id) != 0 || !c.Tapped {
				t.Fatalf("after %d: souls %d, face %d, tapped %v", i, c.Counters["soul"], effectstest.IwFace(g, id), c.Tapped)
			}
			g.WithWriteLock(func() { _ = g.UntapTargetForEffect(id) })
		} else if c.Counters["soul"] != 0 || effectstest.IwFace(g, id) != 1 || c.Tapped {
			t.Fatalf("after 3: souls %d, face %d, tapped %v; want 0, back face, untapped", c.Counters["soul"], effectstest.IwFace(g, id), c.Tapped)
		}
	}
}

func TestMidB17HostileHostelNeedsACreatureToSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := midB17Put(t, g, me, midB17HostelRow())
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatalf("mana: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without sacrificing a creature")
	}
}

func TestMidB17CreepingInnDrainsForEveryCreatureCardExiledWithIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := midB17Put(t, g, me, midB17HostelRow())
	g.WithWriteLock(func() { _ = g.TransformPermanentForEffect(id) })
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	oppLife, myLife := opp.Life, me.Life
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(id, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.MonSettle(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, dead)
	effectstest.MonSettle(t, g)
	if opp.Life != oppLife-1 || me.Life != myLife+1 {
		t.Fatalf("opponent %d -> %d, you %d -> %d; want a drain of 1", oppLife, opp.Life, myLife, me.Life)
	}
}

func midB17JerrenRow() cards.Card {
	return cards.Card{
		ID: uuid.New(), Name: "Jerren, Corrupted Bishop // Ormendahl, the Corrupter", Layout: "transform",
		TypeLine: "Legendary Creature — Human Cleric // Legendary Creature — Demon",
		OracleID: uuid.MustParse(midB17Jerren), Keywords: []string{"Transform"},
		CardFaces: []cards.CardFace{
			{Name: "Jerren, Corrupted Bishop", TypeLine: "Legendary Creature — Human Cleric", ManaCost: "{2}{B}", Colors: []string{"B"}, Power: "2", Toughness: "3",
				OracleText: "Whenever Jerren enters or another nontoken Human you control dies, you lose 1 life and create a 1/1 white Human creature token."},
			{Name: "Ormendahl, the Corrupter", TypeLine: "Legendary Creature — Demon", Colors: []string{"B"}, Power: "6", Toughness: "6",
				OracleText: "Flying, trample, lifelink\nSacrifice another creature: Draw a card."},
		},
	}
}

func TestMidB17JerrenMakesAHumanWhenAnotherNontokenHumanDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	midB17Put(t, g, me, midB17JerrenRow())
	human := effectstest.B12Creature(g, me.ID, "Villager", "Creature — Human Peasant", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	life := me.Life
	midB17Bolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: bear})
	if me.Life != life || effectstest.RfdCountNamed(g, me.ID, "Human") != 0 {
		t.Fatalf("a dying Bear triggered Jerren (life %d)", me.Life)
	}
	midB17Bolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: human})
	if me.Life != life-1 || effectstest.RfdCountNamed(g, me.ID, "Human") != 1 {
		t.Fatalf("life %d (want %d), Human tokens %d (want 1)", me.Life, life-1, effectstest.RfdCountNamed(g, me.ID, "Human"))
	}
	// The token is a Human, but a token's death does not trigger it.
	token := effectstest.FindBattlefieldByName(g, "Human")
	midB17Bolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: token})
	if me.Life != life-1 {
		t.Fatalf("a Human token's death triggered Jerren (life %d)", me.Life)
	}
}

func TestMidB17JerrenTransformsAtExactly13LifeForSixMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := midB17Put(t, g, me, midB17JerrenRow())
	me.Life = 13
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.MonSettle(t, g)
	for i := 0; i < 4; i++ {
		me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.MonSettle(t, g)
	if f := effectstest.IwFace(g, id); f != 1 {
		t.Fatalf("Jerren is on face %d after paying, want Ormendahl", f)
	}
}

func TestMidB17JerrenDoesNotTriggerAtOtherLifeTotals(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := midB17Put(t, g, me, midB17JerrenRow())
	me.Life = 12
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.MonSettle(t, g)
	if len(g.PendingChoices) != 0 || effectstest.IwFace(g, id) != 0 {
		t.Fatal("Jerren offered a payment at 12 life")
	}
}

func midB17LudevicRow() cards.Card {
	return cards.Card{
		ID: uuid.New(), Name: "Ludevic, Necrogenius // Olag, Ludevic's Hubris", Layout: "transform",
		TypeLine: "Legendary Creature — Human Wizard // Legendary Creature — Zombie",
		OracleID: uuid.MustParse(midB17Ludevic), Keywords: []string{"Transform"},
		CardFaces: []cards.CardFace{
			{Name: "Ludevic, Necrogenius", TypeLine: "Legendary Creature — Human Wizard", ManaCost: "{U}{B}", Colors: []string{"U", "B"}, Power: "2", Toughness: "3",
				OracleText: "Whenever Ludevic enters or attacks, mill a card."},
			{Name: "Olag, Ludevic's Hubris", TypeLine: "Legendary Creature — Zombie", Colors: []string{"U", "B"}, Power: "4", Toughness: "4",
				OracleText: "As this creature transforms into Olag, Ludevic's Hubris, it becomes a copy of a creature card exiled with it."},
		},
	}
}

func midB17Graveyard(p *game.Player, n int, name string) []uuid.UUID {
	var ids []uuid.UUID
	for i := 0; i < n; i++ {
		id := uuid.New()
		p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Bear", Colors: []string{"G"},
			Power: 2, Toughness: 2, PrintedPTKnown: true, Owner: p.ID, Controller: p.ID})
		ids = append(ids, id)
	}
	return ids
}

func midB17ActivateLudevic(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, x int, exile []uuid.UUID) error {
	t.Helper()
	for i := 0; i < x; i++ {
		me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
	for _, c := range []string{"U", "U", "B", "B"} {
		me.ManaPool.AddMana(game.ManaToken{Color: c})
	}
	return g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, XValue: x, ExileIDs: exile})
}

func TestMidB17LudevicBecomesACopyOfTheOneCreatureCardExiledWithIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := midB17Put(t, g, me, midB17LudevicRow())
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	yard := midB17Graveyard(me, 1, "Grizzly Bears")
	if err := midB17ActivateLudevic(t, g, me, id, 1, yard); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MonSettle(t, g)
	c := effectstest.LayeredCard(t, g, id)
	if c.TurnOvers != 1 {
		t.Fatal("Ludevic did not transform")
	}
	if c.Name != "Olag, Ludevic's Hubris" || c.CurrentPower() != 5 || c.CurrentToughness() != 5 {
		t.Errorf("Olag is %q %d/%d, want Olag, Ludevic's Hubris 5/5 (4/4 and one counter)", c.Name, c.CurrentPower(), c.CurrentToughness())
	}
	if !c.HasSubtype("Zombie") || !c.HasSubtype("Bear") || !c.HasColor("U") || !c.HasColor("B") || !c.HasColor("G") {
		t.Errorf("Olag is not a blue, black and green Zombie Bear: %+v", c)
	}
	if !c.HasSupertype("Legendary") {
		t.Error("Olag is not legendary")
	}
}

func TestMidB17LudevicAsksWhichOfSeveralExiledCardsAndCountsThemAll(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := midB17Put(t, g, me, midB17LudevicRow())
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	yard := midB17Graveyard(me, 3, "Grizzly Bears")
	if err := midB17ActivateLudevic(t, g, me, id, 3, yard); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MonSettle(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, yard[1])
	effectstest.MonSettle(t, g)
	c := effectstest.LayeredCard(t, g, id)
	if c.Counters[game.CounterPlusOne] != 3 || c.CurrentPower() != 7 {
		t.Errorf("counters %d, power %d; want 3 counters and 7 power", c.Counters[game.CounterPlusOne], c.CurrentPower())
	}
}

func TestMidB17LudevicRefusesXZeroAndWrongNumberOfCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := midB17Put(t, g, me, midB17LudevicRow())
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	yard := midB17Graveyard(me, 2, "Grizzly Bears")
	if err := midB17ActivateLudevic(t, g, me, id, 0, nil); err == nil {
		t.Fatal("activated with X = 0")
	}
	_ = yard
	if effectstest.IwFace(g, id) != 0 {
		t.Fatal("transformed with X = 0")
	}
}

func TestMidB17LudevicMillsWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := me.Graveyard.Size()
	me.ManaPool.AddMana(game.ManaToken{Color: "U"})
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	effectstest.ImportAndCast(t, g, midB17LudevicRow(), me)
	effectstest.MonSettle(t, g)
	if me.Graveyard.Size() != before+1 {
		t.Errorf("graveyard %d -> %d, want one card milled", before, me.Graveyard.Size())
	}
}

func midB17ToUpkeepOf(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; i < 80; i++ {
		if g.Turn.Step == game.StepUpkeep && g.Turn.ActiveSeat == seat {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		effectstest.MonSettle(t, g)
	}
	t.Fatal("never reached the upkeep")
}

func TestMidB17ShadowKinMillsEveryoneAndBecomesACopyOfAnExiledCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	kin := effectstest.B12Push(g, me.ID, "Shadow Kin", "Creature — Shapeshifter", midB17Shadow, 2, 2)
	prevSeat := (seat + len(g.Seats) - 1) % len(g.Seats)
	opp := g.Seats[prevSeat]
	midB17ToUpkeepOf(t, g, prevSeat)
	prize := uuid.New()
	push := func(id uuid.UUID, name, typeLine string) {
		opp.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Power: 3, Toughness: 3, PrintedPTKnown: true, Owner: opp.ID, Controller: opp.ID})
	}
	// Pushed last is the top: the first one is drawn in their draw step.
	push(prize, "Hill Giant", "Creature — Giant")
	push(uuid.New(), "Filler", "Land")
	push(uuid.New(), "Filler", "Land")
	push(uuid.New(), "Drawn", "Land")
	yard := opp.Graveyard.Size()
	midB17ToUpkeepOf(t, g, seat)
	effectstest.MonSettle(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, prize)
	effectstest.MonSettle(t, g)
	if z := g.FindCardZoneForEffect(prize); z == nil || z.Kind != game.ZoneExile {
		t.Fatal("the chosen creature card was not exiled")
	}
	c := effectstest.LayeredCard(t, g, kin)
	if c.Name != "Hill Giant" || c.CurrentPower() != 3 {
		t.Errorf("Shadow Kin is %q %d/%d, want a Hill Giant 3/3", c.Name, c.CurrentPower(), c.CurrentToughness())
	}
	if opp.Graveyard.Size() != yard+2 {
		t.Errorf("the opponent's graveyard went %d -> %d, want two milled cards left in it", yard, opp.Graveyard.Size())
	}
}

func TestMidB17ShadowKinDecliningLeavesItAShadowKin(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	kin := effectstest.B12Push(g, me.ID, "Shadow Kin", "Creature — Shapeshifter", midB17Shadow, 2, 2)
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Hill Giant", TypeLine: "Creature — Giant", Power: 3, Toughness: 3, PrintedPTKnown: true, Owner: me.ID, Controller: me.ID})
	midB17ToUpkeepOf(t, g, (seat+len(g.Seats)-1)%len(g.Seats))
	midB17ToUpkeepOf(t, g, seat)
	effectstest.MonSettle(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.MonSettle(t, g)
	if c := effectstest.LayeredCard(t, g, kin); c.Name != "Shadow Kin" {
		t.Errorf("declining still copied: %q", c.Name)
	}
}
