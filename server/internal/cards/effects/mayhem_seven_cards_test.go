package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mayhem_seven_cards_test.go — the seven mayhem cards (#4628, #4455),
// each played through the real catalog. The seam's own tests are in
// mayhem_cards_test.go.

var mayhemCards = []struct {
	name   string
	oracle string
	cost   string
}{
	{"Abomination, World Ravager", mayhemAbominationOracle, "{4}{R}"},
	{"Chameleon, Master of Disguise", mayhemChameleonOracle, "{2}{U}"},
	{"Green Goblin, Back for More", mayhemGoblinOracle, "{3}{B}{B}"},
	{"Rocket-Powered Goblin Glider", mayhemGliderOracle, "{2}"},
	{"Sadistic Slash", mayhemSlashOracle, "{1}{B}"},
	{"Swarm, Being of Bees", mayhemSwarmOracle, "{B}"},
	{"The Squadron Sinister", mayhemSquadronOracle, "{3}{U}{R}"},
}

// Every card is Full, opens its graveyard only conditionally (so a card
// that was not discarded is not castable), and offers exactly one
// graveyard price: mayhem at the printed cost, with no exile clause.
func TestMayhemCardsDeclareTheirPrice(t *testing.T) {
	for _, c := range mayhemCards {
		t.Run(c.name, func(t *testing.T) {
			spec, ok := Lookup(c.oracle)
			if !ok {
				t.Fatalf("%s is not registered", c.name)
			}
			if spec.Name != c.name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
				t.Errorf("%s: name %q, completeness %s, caveats %v", c.name, spec.Name, spec.Completeness, spec.Caveats)
			}
			if game.CardCastableFromZone(c.oracle, game.ZoneGraveyard) {
				t.Errorf("%s opens its graveyard unconditionally", c.name)
			}
			offers := game.AlternativeCostsOfferedFromZone(c.oracle, game.ZoneGraveyard)
			if len(offers) != 1 || offers[0].Key != "mayhem" {
				t.Fatalf("%s graveyard offers = %+v, want one mayhem", c.name, offers)
			}
			o := offers[0]
			if o.ManaCost != c.cost || !o.RequiresDiscardedThisTurn || o.ExileOnLeavingStack || o.FromZone != game.ZoneGraveyard {
				t.Errorf("%s mayhem offer = %+v, want %s, discard-gated, no exile", c.name, o, c.cost)
			}
			if hand := game.AlternativeCostsOfferedFromZone(c.oracle, game.ZoneHand); len(hand) != 0 {
				t.Errorf("%s offers %+v from hand", c.name, hand)
			}
		})
	}
}

func mayhemCast(g *game.Game, me *game.Player, id uuid.UUID) error {
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "mayhem"})
}

// Abomination, World Ravager: a 10/10 with menace and trample for
// {4}{R} out of the graveyard.
func TestMayhemAbominationIsCastForFiveAndKeepsItsKeywords(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	id := mayhemHandCard(g, me, "Abomination, World Ravager", "Legendary Creature — Gamma Berserker Villain", "{7}{R}", mayhemAbominationOracle)
	mayhemDiscard(t, g, me, id)

	floatForTest(g, me, "CCCCB")
	if err := mayhemCast(g, me, id); err == nil {
		t.Fatal("{4}{R} paid with black mana")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	floatForTest(g, me, "CCCCR")
	if err := mayhemCast(g, me, id); err != nil {
		t.Fatalf("mayhem cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !onBattlefieldByID(g, id) {
		t.Fatal("Abomination did not enter")
	}
	for _, kw := range []string{"menace", "trample"} {
		if !effectiveAbilitiesContain(t, g, id, kw) {
			t.Errorf("Abomination lacks %s", kw)
		}
	}
}

// "Timing rules still apply": a sorcery-speed creature is not castable
// from the graveyard at instant speed, a flash creature is. Both were
// discarded this turn.
func TestMayhemTimingRulesStillApply(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	abom := mayhemHandCard(g, me, "Abomination, World Ravager", "Legendary Creature — Gamma Berserker Villain", "{7}{R}", mayhemAbominationOracle)
	swarm := mayhemHandCard(g, me, "Swarm, Being of Bees", "Legendary Creature — Insect Villain", "{2}{B}", mayhemSwarmOracle)
	mayhemDiscard(t, g, me, abom)
	mayhemDiscard(t, g, me, swarm)
	advanceTo(t, g, game.StepBeginCombat)

	floatForTest(g, me, "CCCCRB")
	if err := mayhemCast(g, me, abom); err == nil {
		t.Error("a sorcery-speed creature was mayhemed in the beginning of combat step")
	}
	if !me.Graveyard.Contains(abom) {
		t.Error("the refused cast moved the card")
	}
	if err := mayhemCast(g, me, swarm); err != nil {
		t.Fatalf("Swarm has flash: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !onBattlefieldByID(g, swarm) {
		t.Fatal("Swarm did not enter")
	}
	if !effectiveAbilitiesContain(t, g, swarm, "flying") {
		t.Error("Swarm lacks flying")
	}
	// {B} was all it cost.
	if got := len(me.ManaPool); got != 5 {
		t.Errorf("pool after Swarm's {B} holds %d tokens, want 5 of the 6 floated", got)
	}
}

// Chameleon enters as a copy of a creature you control, named
// Chameleon, Master of Disguise; only your own creatures are offered.
func TestMayhemChameleonCopiesYourCreatureUnderItsOwnName(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	mine := seedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 4, 4)
	theirs := seedCopyableCreature(g, opp.ID, "Runeclaw Bear", "Creature — Bear", 2, 2)
	id := mayhemHandCard(g, me, "Chameleon, Master of Disguise", "Legendary Creature — Human Shapeshifter Villain", "{3}{U}", mayhemChameleonOracle)
	mayhemDiscard(t, g, me, id)

	floatForTest(g, me, "CCU")
	if err := mayhemCast(g, me, id); err != nil {
		t.Fatalf("mayhem cast: %v", err)
	}
	for i := 0; i < 8 && copyPrompt(g) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	pc := copyPrompt(g)
	if pc == nil {
		t.Fatal("no copy prompt")
	}
	if len(pc.CopyOptions) != 1 || pc.CopyOptions[0] != mine {
		t.Fatalf("CopyOptions = %v, want only %s (%s is an opponent's)", pc.CopyOptions, mine, theirs)
	}
	resolveWithCopyChoice(t, g, mine)
	got := copyBattlefieldCard(t, g, id)
	if got.Name != "Chameleon, Master of Disguise" {
		t.Errorf("name = %q, want Chameleon, Master of Disguise", got.Name)
	}
	if got.CurrentPower() != 4 || got.CurrentToughness() != 4 || !got.HasSubtype("Bear") {
		t.Errorf("copy is %d/%d %q, want the 4/4 Bear it copied", got.CurrentPower(), got.CurrentToughness(), got.TypeLine)
	}
}

// Green Goblin: at the beginning of combat on its controller's turn it
// may discard a card, and then each opponent discards a card.
func TestMayhemGreenGoblinMakesEachOpponentDiscardIfYouDo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	pushCatalogPermanent(g, me.ID, "Green Goblin, Back for More", "Legendary Creature — Goblin Human Villain", mayhemGoblinOracle, false)
	hands := map[uuid.UUID]int{}
	for _, p := range g.Seats[1:] {
		hands[p.ID] = p.Hand.Size()
	}
	myHand := me.Hand.Size()
	pitch := me.Hand.Cards[0].InstanceID

	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID, pitch)
	for _, p := range g.Seats[1:] {
		if discardChoiceFor(g, p.ID) == nil {
			t.Fatalf("%s was not asked to discard", p.Name)
		}
		discardFromHand(t, g, p.ID)
	}
	if me.Hand.Size() != myHand-1 {
		t.Errorf("my hand = %d, want %d", me.Hand.Size(), myHand-1)
	}
	for _, p := range g.Seats[1:] {
		if p.Hand.Size() != hands[p.ID]-1 {
			t.Errorf("%s hand = %d, want %d", p.Name, p.Hand.Size(), hands[p.ID]-1)
		}
	}
}

func TestMayhemGreenGoblinDoesNothingToOpponentsWhenYouDecline(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	pushCatalogPermanent(g, me.ID, "Green Goblin, Back for More", "Legendary Creature — Goblin Human Villain", mayhemGoblinOracle, false)
	hands := map[uuid.UUID]int{}
	for _, p := range g.Seats[1:] {
		hands[p.ID] = p.Hand.Size()
	}

	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID) // up to one: none
	for _, p := range g.Seats[1:] {
		if discardChoiceFor(g, p.ID) != nil {
			t.Errorf("%s was asked to discard though the controller did not", p.Name)
		}
		if p.Hand.Size() != hands[p.ID] {
			t.Errorf("%s hand = %d, want %d", p.Name, p.Hand.Size(), hands[p.ID])
		}
	}
}

// The Glider's enters trigger needs the cast to have come from a
// graveyard: mayhem attaches it, the printed cost does not.
func TestMayhemGliderAttachesOnlyWhenCastFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)

	// From the graveyard: attached, +2/+0, flying and haste.
	glider := mayhemHandCard(g, me, "Rocket-Powered Goblin Glider", "Artifact — Equipment", "{3}", mayhemGliderOracle)
	mayhemDiscard(t, g, me, glider)
	floatForTest(g, me, "CC")
	if err := mayhemCast(g, me, glider); err != nil {
		t.Fatalf("mayhem cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if a := copyBattlefieldCard(t, g, glider); a.AttachedTo.ID != bear {
		t.Fatalf("the Glider is attached to %v, want the bear", a.AttachedTo)
	}
	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("bear is %d/%d, want 4/2", p, tough)
	}
	for _, kw := range []string{"flying", "haste"} {
		if !effectiveAbilitiesContain(t, g, bear, kw) {
			t.Errorf("the equipped bear lacks %s", kw)
		}
	}

	// From hand for the printed {3}: no trigger, no attachment.
	other := mayhemHandCard(g, me, "Rocket-Powered Goblin Glider", "Artifact — Equipment", "{3}", mayhemGliderOracle)
	floatForTest(g, me, "CCC")
	if err := g.CastSpell(me.ID, other, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("hand cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := chooseCardsChoiceFor(g, me.ID); c != nil {
		t.Errorf("a hand-cast Glider asked for a target: %+v", c)
	}
	if a := copyBattlefieldCard(t, g, other); a.AttachedTo.ID != uuid.Nil {
		t.Errorf("the hand-cast Glider attached itself to %v", a.AttachedTo)
	}
}

// The Squadron Sinister: other Villains you control get +2/+2 and have
// flying and haste; non-Villains, opposing Villains and the Squadron
// itself do not.
func TestMayhemSquadronSinisterLordsOnlyYourOtherVillains(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	ally := seedCopyableCreature(g, me.ID, "Goblin Henchman", "Creature — Goblin Villain", 1, 1)
	bear := seedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	foe := seedCopyableCreature(g, opp.ID, "Rival Henchman", "Creature — Human Villain", 1, 1)
	id := mayhemHandCard(g, me, "The Squadron Sinister", "Legendary Creature — Human Villain", "{5}{U}{R}", mayhemSquadronOracle)
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == id {
				me.Hand.Cards[i].Power, me.Hand.Cards[i].Toughness = 5, 5
			}
		}
	})
	mayhemDiscard(t, g, me, id)

	floatForTest(g, me, "CCCUR")
	if err := mayhemCast(g, me, id); err != nil {
		t.Fatalf("mayhem cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, ally); p != 3 || tough != 3 {
		t.Errorf("your other Villain is %d/%d, want 3/3", p, tough)
	}
	for _, kw := range []string{"flying", "haste"} {
		if !effectiveAbilitiesContain(t, g, ally, kw) {
			t.Errorf("your other Villain lacks %s", kw)
		}
		if effectiveAbilitiesContain(t, g, bear, kw) || effectiveAbilitiesContain(t, g, foe, kw) {
			t.Errorf("a non-Villain or opposing Villain has %s", kw)
		}
	}
	if p, _ := effectivePTOf(t, g, bear); p != 2 {
		t.Errorf("a non-Villain got the pump: power %d", p)
	}
	if p, _ := effectivePTOf(t, g, foe); p != 1 {
		t.Errorf("an opposing Villain got the pump: power %d", p)
	}
	if p, tough := effectivePTOf(t, g, id); p != 5 || tough != 5 {
		t.Errorf("the Squadron is %d/%d, want its printed 5/5 (\"other\")", p, tough)
	}
}
