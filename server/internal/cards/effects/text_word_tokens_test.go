package effects

import (
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// text_word_tokens_test.go — #4005, ADR 0160's slice-4 follow-up: the
// two classes of printed word slice 3 left unread. A token description
// ("create a 1/1 red Goblin creature token") and a cast-permission filter
// or condition that names a creature type are words of the text of the
// object whose clause they are (CR 612.2), so a Sleight of Mind or an
// Artificial Evolution reaches them. TestTokenDescriptionsReadThroughText
// is the lint; game.TestTokenThroughText… the engine half.

const (
	twtKrenkoOracle       = "68418069-f615-40ef-ae0d-764192acae00"
	twtStridehangarOracle = "9070c98b-fd01-4eeb-a4ec-fc464946c7c0"
)

// twtTokens are the tokens `owner` controls, by id.
func twtTokens(g *game.Game, owner uuid.UUID) []game.Card {
	var out []game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == owner && c.IsToken() {
				out = append(out, c)
			}
		}
	})
	return out
}

// Krenko, Mob Boss's "create X 1/1 red Goblin creature tokens" under a
// Sleight of Mind (red → blue) and an Artificial Evolution (Goblin → Elf)
// makes blue Elves named Elf (CR 111.4), counting the Elves Krenko's text
// now names — Krenko himself, whose type line says Elf too.
func TestTextWordTokensKrenkoMakesBlueElves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	krenko := pushCatalogPermanent(g, me.ID, "Krenko, Mob Boss", "Legendary Creature — Goblin Warrior", twtKrenkoOracle, false)
	twhSleight(t, g, krenko, twRed, twBlue)
	twxEvolve(t, g, krenko, "Goblin", "Elf")

	if err := g.ActivateCatalogAbility(me.ID, krenko, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	tokens := twtTokens(g, me.ID)
	if len(tokens) != 1 {
		t.Fatalf("%d tokens, want 1 (Krenko is the one Elf)", len(tokens))
	}
	tok := gaLayered(t, g, tokens[0].InstanceID)
	if !tok.HasColor("U") || tok.HasColor("R") {
		t.Errorf("token colours %v: want blue, not red", tok.Effective().Colors)
	}
	if !tok.HasSubtype("Elf") || tok.HasSubtype("Goblin") {
		t.Errorf("token subtypes %v: want an Elf and not a Goblin", tok.Effective().Subtypes)
	}
	if tok.Name != "Elf" {
		t.Errorf("the token is named %q, want Elf", tok.Name)
	}
}

// Stridehangar Automaton's "an additional 1/1 colorless Thopter artifact
// creature token" is a word of the Automaton's text, read by the
// replacement through its source: under an Artificial Evolution (Thopter
// → Elf) the extra token is an Elf.
func TestTextWordTokensStridehangarAddsAnElf(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	automaton := b12Push(g, me.ID, "Stridehangar Automaton", "Artifact Creature — Construct", twtStridehangarOracle, 2, 2)
	twxEvolve(t, g, automaton, "Thopter", "Elf")

	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, TreasureToken(), 1); err != nil {
			t.Fatalf("create: %v", err)
		}
	})
	var names []string
	for _, c := range twtTokens(g, me.ID) {
		names = append(names, c.Name)
		if c.Name == "Treasure" {
			continue
		}
		if !c.HasSubtype("Elf") || c.HasSubtype("Thopter") || !c.IsArtifact() {
			t.Errorf("the extra token is a %q, want an Elf artifact creature", c.TypeLine)
		}
	}
	slices.Sort(names)
	if !slices.Equal(names, []string{"Elf", "Treasure"}) {
		t.Errorf("tokens %v, want a Treasure and an Elf", names)
	}
}

// Gisa and Geralf's "you may cast a Zombie creature spell from your
// graveyard" under an Artificial Evolution (Zombie → Elf) opens Elf
// creature spells and shuts Zombies: the standing permission is derived
// through the permanent's text, and its label says Elf.
func TestTextWordPermissionGisaAndGeralfCastsAnElf(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gisa := pushCatalogPermanent(g, me.ID, "Gisa and Geralf", "Legendary Creature — Human Wizard", wrGisaAndGeralfOracle, false)
	zombie := ptGraveCard(me, "Walker", "Creature — Zombie")
	elf := ptGraveCard(me, "Dead Elf", "Creature — Elf")
	twxEvolve(t, g, gisa, "Zombie", "Elf")

	var label string
	g.ReadSnapshot(func() {
		for _, c := range me.Graveyard.Cards {
			if c.InstanceID == elf {
				if perm := g.CastPermissionForLocked(me.ID, c, game.ZoneGraveyard); perm != nil {
					label = perm.Label
				}
			}
		}
	})
	if !strings.Contains(label, "an Elf creature spell") || strings.Contains(label, "Zombie") {
		t.Errorf("the permission reads %q, want the Elf the text now names", label)
	}
	if err := muldrothaCast(g, me, zombie, ""); err == nil {
		t.Error("a Zombie was cast from the graveyard under \"cast an Elf creature spell\"")
	}
	if err := muldrothaCast(g, me, elf, ""); err != nil {
		t.Fatalf("an Elf from the graveyard: %v", err)
	}
}

// Deeproot Historian's "Merfolk and Druid cards in your graveyard have
// retrace" under an Artificial Evolution (Merfolk → Goblin) reaches a
// Goblin card and no longer a Merfolk one; the Druid half is untouched.
func TestTextWordPermissionDeeprootHistorianRetracesAGoblin(t *testing.T) {
	g, me, historian, land := historianTable(t)
	twxEvolve(t, g, historian, "Merfolk", "Goblin")
	merfolk := graveyardCardOf(g, me, "Gravebound Merfolk", "Creature — Merfolk Wizard", "{1}{U}", "")
	goblin := graveyardCardOf(g, me, "Gravebound Goblin", "Creature — Goblin", "{1}{R}", "")
	druid := graveyardCardOf(g, me, "Gravebound Druid", "Creature — Elf Druid", "{1}{G}", "")

	payMana(t, g, me, "{R}{G}")
	if len(retraceMoves(g, me.ID, merfolk)) != 0 {
		t.Error("a Merfolk card is offered retrace under \"Goblin and Druid cards\"")
	}
	if len(retraceMoves(g, me.ID, druid)) == 0 {
		t.Error("the Druid half no longer reaches a Druid card")
	}
	if err := retraceCast(g, me, goblin, land); err != nil {
		t.Fatalf("retrace of a Goblin card: %v", err)
	}
}

// Rooftop Storm's "Zombie creature spells you cast" under an Artificial
// Evolution (Zombie → Elf): the free cast is offered to an Elf creature
// spell and not to a Zombie.
func TestTextWordPermissionRooftopStormFreesElves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	storm := pushCatalogPermanent(g, me.ID, "Rooftop Storm", "Enchantment", wrRooftopStormOracle, false)
	twxEvolve(t, g, storm, "Zombie", "Elf")
	zombie := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: zombie, Name: "Big Zombie", TypeLine: "Creature — Zombie", ManaCost: "{6}{B}", Owner: me.ID, Controller: me.ID})
	elf := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: elf, Name: "Big Elf", TypeLine: "Creature — Elf", ManaCost: "{6}{G}", Owner: me.ID, Controller: me.ID})

	if offerWithKey(grantedCardOffers(g, me.ID, zombie, game.ZoneHand), GrantedAltCostZeroZombieCreatures) != nil {
		t.Error("a Zombie creature spell is offered the free cast under \"Elf creature spells\"")
	}
	if offerWithKey(grantedCardOffers(g, me.ID, elf, game.ZoneHand), GrantedAltCostZeroZombieCreatures) == nil {
		t.Fatal("no free-cast offer for an Elf creature spell")
	}
	if err := g.CastSpell(me.ID, elf, game.CastSpellParams{Strict: true, AlternativeCost: GrantedAltCostZeroZombieCreatures}); err != nil {
		t.Fatalf("cast the Elf for {0}: %v", err)
	}
}

// Robber of the Rich's "during any turn you attacked with a Rogue" under
// an Artificial Evolution (Rogue → Elf) is fixed as "an Elf" when the
// trigger grants it (CR 608.2h). The Robber's own type line reads Elf
// too, so its attack opens the card.
func TestTextWordPermissionRobberOfTheRichNeedsAnElf(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	robber := b12Push(g, me.ID, "Robber of the Rich", "Creature — Human Archer Rogue", robberOfTheRichOracle, 2, 2)
	twxEvolve(t, g, robber, "Rogue", "Elf")
	g.WithWriteLock(func() { me.Hand.Cards = nil; opp.Hand.Cards = nil })
	for i := 0; i < 3; i++ {
		opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Their Card", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	}
	theirs := topLibraryCard(opp, "Their Bolt", "Instant")

	declareAttack(t, g, opp.ID, robber)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(theirs) {
		t.Fatal("nothing was exiled")
	}
	perm := exiledPermission(g, theirs)
	if perm.Condition == nil || perm.Condition.Subtype != "Elf" {
		t.Fatalf("condition %+v, want \"attacked with an Elf\"", perm.Condition)
	}
	if !permissionLive(g, perm, me.ID) {
		t.Error("an Elf (the Robber) attacked this turn and the card is shut")
	}
}
