package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	bnsB39SpriteNoble     = "bf3e3004-4133-4542-9ee3-2fb98e759ffa"
	bnsB39SquadCaptain    = "865812c2-1838-4015-bf0d-990af76f7000"
	bnsB39Valkyrie        = "b6b4c459-9c8b-4b95-8706-05d972b8daea"
	bnsB39StarPupil       = "37fd3a9d-74fe-4180-805a-b865e7228a77"
	bnsB39Courser         = "01317d6c-8c11-40e1-8f7b-d36050973ada"
	bnsB39Sentry          = "e0f676ab-f859-4d4f-abe0-d033ffd3c396"
	bnsB39Unicorn         = "c823d38c-e73c-4e09-ab20-c5e7de1f1c64"
	bnsB39StitchedDrake   = "10fd868f-45b3-4e2f-89d9-5c0b8a92f9fe"
	bnsB39Outfitter       = "8e646a6b-77ba-4f7d-8ee9-b0b084453f77"
	bnsB39Confessor       = "e7c200a9-4793-44c8-ba8d-6298dc83bbac"
	bnsB39SungracePegasus = "67f59250-d46a-4637-9493-165d23278817"
)

func TestBnsB39CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		bnsB39SpriteNoble: "Sprite Noble", bnsB39SquadCaptain: "Squad Captain",
		bnsB39Valkyrie: "Stalwart Valkyrie", bnsB39StarPupil: "Star Pupil",
		bnsB39Courser: "Starnheim Courser", bnsB39Sentry: "Steadfast Sentry",
		bnsB39Unicorn: "Steadfast Unicorn", bnsB39StitchedDrake: "Stitched Drake",
		bnsB39Outfitter: "Stone Haven Outfitter", bnsB39Confessor: "Stronghold Confessor",
		bnsB39SungracePegasus: "Sungrace Pegasus",
	} {
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

func bnsB39Flyer(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name,
		TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"},
		Owner: owner, Controller: owner})
}

func TestBnsB39SungracePegasusHasFlyingAndLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b12Push(g, me.ID, "Sungrace Pegasus", "Creature — Pegasus", bnsB39SungracePegasus, 1, 2)
	c := b12Card(t, g, id)
	if !game.HasKeyword(&c, "flying") || !game.HasKeyword(&c, "lifelink") {
		t.Fatalf("Sungrace Pegasus lacks flying or lifelink: %v", c.Effective().Abilities)
	}
}

func TestBnsB39SpriteNobleBoostsOtherFliersOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	noble := pushCatalogPermanent(g, me.ID, "Sprite Noble", "Creature — Faerie Noble", bnsB39SpriteNoble, false)
	flyer := bnsB39Flyer(g, me.ID, "My Bird")
	theirs := bnsB39Flyer(g, g.Seats[1].ID, "Their Bird")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if p, tough := effectivePTOf(t, g, flyer); p != 1 || tough != 2 {
		t.Errorf("my flier is %d/%d, want 1/2", p, tough)
	}
	if p, tough := effectivePTOf(t, g, theirs); p != 1 || tough != 1 {
		t.Errorf("their flier is %d/%d, want 1/1", p, tough)
	}
	if p, tough := effectivePTOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("the bear is %d/%d, want 2/2", p, tough)
	}
	if err := g.ActivateCatalogAbility(me.ID, noble, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, flyer); p != 2 || tough != 2 {
		t.Errorf("my flier is %d/%d after the tap ability, want 2/2", p, tough)
	}
	if p, _ := effectivePTOf(t, g, bear); p != 2 {
		t.Errorf("the bear got the bonus: power %d", p)
	}
	if p, _ := effectivePTOf(t, g, theirs); p != 1 {
		t.Errorf("their flier got the bonus: power %d", p)
	}
	if err := g.ActivateCatalogAbility(me.ID, noble, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("Sprite Noble activated while tapped")
	}
}

func TestBnsB39SquadCaptainEntersWithACounterPerOtherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	b12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
	id := b12PlayFromHand(t, g, "Squad Captain", "Creature — Human Soldier", bnsB39SquadCaptain, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if got := b12Counter(t, g, id, game.CounterPlusOne); got != 2 {
		t.Fatalf("Squad Captain has %d counters, want 2", got)
	}
}

func bnsB39CastFromHand(t *testing.T, g *game.Game, name, typeLine, cost, oracle string, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	params.Strict = true
	return castWithTapParams(t, g, name, typeLine, cost, oracle, params)
}

func TestBnsB39StalwartValkyrieAlternativeCostExilesACreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	fuel := pushGraveyardCardForTest(me, "Fuel")
	bnsB39Pool(me, "CW")
	id, err := bnsB39CastFromHand(t, g, "Stalwart Valkyrie", "Creature — Angel Warrior", "{3}{W}", bnsB39Valkyrie,
		game.CastSpellParams{AlternativeCost: "exile-creature-from-graveyard", AltCostIDs: []uuid.UUID{fuel}})
	if err != nil {
		t.Fatalf("alternative cast: %v", err)
	}
	if !g.Exile.Contains(fuel) {
		t.Error("the creature card was not exiled")
	}
	passPriorityAroundTable(t, g)
	if b12ZoneOf(g, id) != game.ZoneBattlefield {
		t.Errorf("Valkyrie is in %v, want the battlefield", b12ZoneOf(g, id))
	}
}

func TestBnsB39StalwartValkyrieAlternativeCostNeedsACreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bnsB39Pool(me, "CW")
	_, err := bnsB39CastFromHand(t, g, "Stalwart Valkyrie", "Creature — Angel Warrior", "{3}{W}", bnsB39Valkyrie,
		game.CastSpellParams{AlternativeCost: "exile-creature-from-graveyard"})
	if err == nil {
		t.Fatal("cast for the alternative cost with an empty graveyard")
	}
	// The printed cost can't be paid with two mana either.
	if _, err := bnsB39CastFromHand(t, g, "Stalwart Valkyrie", "Creature — Angel Warrior", "{3}{W}", bnsB39Valkyrie,
		game.CastSpellParams{}); err == nil {
		t.Fatal("cast for the printed cost with two mana")
	}
}

func TestBnsB39StitchedDrakeNeedsACreatureCardToExile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bnsB39Pool(me, "CUU")
	if _, err := bnsB39CastFromHand(t, g, "Stitched Drake", "Creature — Zombie Drake", "{1}{U}{U}", bnsB39StitchedDrake,
		game.CastSpellParams{}); err == nil {
		t.Fatal("cast with no creature card in the graveyard")
	}
	fuel := pushGraveyardCardForTest(me, "Fuel")
	id, err := bnsB39CastFromHand(t, g, "Stitched Drake", "Creature — Zombie Drake", "{1}{U}{U}", bnsB39StitchedDrake,
		game.CastSpellParams{ExileIDs: []uuid.UUID{fuel}})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	if !g.Exile.Contains(fuel) {
		t.Error("the creature card was not exiled")
	}
	passPriorityAroundTable(t, g)
	if b12ZoneOf(g, id) != game.ZoneBattlefield {
		t.Errorf("Stitched Drake is in %v, want the battlefield", b12ZoneOf(g, id))
	}
}

func TestBnsB39StarPupilHandsItsCountersOn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	pupil := b12PlayFromHand(t, g, "Star Pupil", "Creature — Human Wizard", bnsB39StarPupil, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if got := b12Counter(t, g, pupil, game.CounterPlusOne); got != 1 {
		t.Fatalf("enters with %d counters, want 1", got)
	}
	killCreature(t, g, me.ID, pupil)
	eoeABPickIfAsked(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if got := b12Counter(t, g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the bear has %d counters, want 1", got)
	}
}

func TestBnsB39SteadfastSentryPutsACounterOnATargetWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	sentry := b12Push(g, me.ID, "Steadfast Sentry", "Creature — Human Soldier", bnsB39Sentry, 3, 2)
	killCreature(t, g, me.ID, sentry)
	eoeABPickIfAsked(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if got := b12Counter(t, g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the bear has %d counters, want 1", got)
	}
}

func TestBnsB39SteadfastUnicornPumpsMyCreaturesOnMyTurnOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	unicorn := pushCatalogPermanent(g, me.ID, "Steadfast Unicorn", "Creature — Unicorn", bnsB39Unicorn, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
	bnsB39Pool(me, "CCCW")
	if err := g.ActivateCatalogAbility(me.ID, unicorn, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the bear is %d/%d, want 3/3", p, tough)
	}
	if p, tough := effectivePTOf(t, g, unicorn); p != 2 || tough != 2 {
		t.Errorf("the Unicorn is %d/%d, want 2/2", p, tough)
	}
	if p, tough := effectivePTOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("their bear is %d/%d, want 2/2", p, tough)
	}
	b := b12Card(t, g, bear)
	if !game.HasKeyword(&b, "vigilance") {
		t.Error("the bear did not gain vigilance")
	}
	// Not on another player's turn.
	g.Turn.ActiveSeat = 1
	bnsB39Pool(me, "CCCW")
	if err := g.ActivateCatalogAbility(me.ID, unicorn, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated during an opponent's turn")
	}
}

func TestBnsB39StrongholdConfessorKickerGivesTwoCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bnsB39Pool(me, "CCCB")
	kicked, err := bnsB39CastFromHand(t, g, "Stronghold Confessor", "Creature — Human Cleric", "{B}", bnsB39Confessor,
		game.CastSpellParams{OptionalCosts: []int{0}})
	if err != nil {
		t.Fatalf("kicked cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := b12Counter(t, g, kicked, game.CounterPlusOne); got != 2 {
		t.Errorf("kicked Confessor has %d counters, want 2", got)
	}
	bnsB39Pool(me, "B")
	plain, err := bnsB39CastFromHand(t, g, "Stronghold Confessor", "Creature — Human Cleric", "{B}", bnsB39Confessor,
		game.CastSpellParams{})
	if err != nil {
		t.Fatalf("plain cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := b12Counter(t, g, plain, game.CounterPlusOne); got != 0 {
		t.Errorf("unkicked Confessor has %d counters, want 0", got)
	}
}

func TestBnsB39StarnheimCourserDiscountsArtifactsAndEnchantments(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Starnheim Courser", "Creature — Pegasus", bnsB39Courser, 2, 2)
	for _, tc := range []struct {
		name, typeLine string
		ok             bool
	}{
		{"Rock", "Artifact", true},
		{"Ward", "Enchantment", true},
		{"Both", "Enchantment Artifact", true},
		{"Bear", "Creature — Bear", false},
	} {
		bnsB39Pool(me, "C")
		_, err := castWithTapParams(t, g, tc.name, tc.typeLine, "{2}", "", game.CastSpellParams{Strict: true})
		if (err == nil) != tc.ok {
			t.Errorf("%s with one mana: err %v, want success %v", tc.name, err, tc.ok)
		}
		passPriorityAroundTable(t, g)
		me.ManaPool.EmptyPool()
	}
}

func TestBnsB39StoneHavenOutfitterBoostsAndDrawsForEquippedCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Stone Haven Outfitter", "Creature — Kor Artificer Ally", bnsB39Outfitter, 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	plain := b12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	sword := b12Permanent(g, me.ID, "Sword", "Artifact — Equipment")
	if err := g.AttachForEffect(sword, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
		t.Fatalf("attach: %v", err)
	}
	if p, tough := effectivePTOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the equipped bear is %d/%d, want 3/3", p, tough)
	}
	if p, _ := effectivePTOf(t, g, plain); p != 2 {
		t.Errorf("the unequipped bear got the bonus: %d", p)
	}
	seedLibrary(me, "A", "B", "C")
	hand := me.Hand.Size()
	killCreature(t, g, me.ID, plain)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatalf("an unequipped creature dying drew a card")
	}
	killCreature(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d after an equipped creature died, want +1", hand, me.Hand.Size())
	}
}

// bnsB39Pool drops one mana of each letter into the pool: C for generic.
func bnsB39Pool(p *game.Player, colors string) {
	for _, c := range colors {
		p.ManaPool.AddMana(game.ManaToken{Color: string(c)})
	}
}
