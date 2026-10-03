package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// adr0108_pr6_cards_test.go — ADR 0108 PR 6 (#1904): the cards the
// not-one-use source shield (preventFromSource), Dark Sphere's half
// shield and the CR 615.7 division unblock, each against the damage it
// must stop and the damage it must not.

const (
	pr6PayNoHeed        = "71da5acb-78bc-453a-ae28-d22faa5a4e43"
	pr6AuriokReplica    = "783a62c0-3b06-4250-a1cb-5cf4042611aa"
	pr6ForgeTender      = "70bb275b-3458-4690-a50f-b231fbf0bccb"
	pr6Prahv            = "37ff5ba6-0763-4c73-85bf-66856e67b8f3"
	pr6RithsCharm       = "9dbccdc5-61f7-4a7d-bbbc-5cab393e24f7"
	pr6DarkSphere       = "397f53f7-f801-4442-a778-2f26ac246b62"
	pr6HealingGrace     = "e3b22777-4d5b-4f6c-a339-24d204bd007e"
	pr6SamiteMinistry   = "af669df1-3526-46ca-b90c-19b112b7ba44"
	pr6ShieldmageAdv    = "57188b3d-567c-4dff-8b93-7cc7a47894be"
	pr6ProtectiveSphere = "2037c4bf-bf0d-46ab-8b4b-61d9a50c431e"
	pr6MournersShield   = "34cadd9d-b8e6-422c-b0ec-a2420cd983ad"
)

// pr6Damage deals `n` from `source` to the permanent `to` as one
// instruction, then runs the priority boundary.
func pr6Damage(t *testing.T, g *game.Game, source, to uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(source, to, n); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
}

func pr6Marked(g *game.Game, id uuid.UUID) int {
	c := findBattlefieldCardForTest(g, id)
	if c == nil {
		return -1
	}
	return c.DamageMarked
}

// Pay No Heed: every instance of the chosen source's damage this turn,
// to anything, is prevented; another source's is not.
func TestADR0108PR6PayNoHeedPreventsTheSourceAllTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := pr7Creature(g, opp.ID, "Dragon", 5, "R")
	other := pr7Creature(g, opp.ID, "Other", 2, "R")
	bear := pr7Creature(g, me.ID, "Bear", 2, "G")
	castCatalogSpell(t, g, "Pay No Heed", "Instant", pr6PayNoHeed, nil)
	passPriorityAroundTable(t, g)
	pr7Choose(t, g, me.ID, dragon)
	life := me.Life
	pr7Hit(t, g, dragon, me.ID, 3)
	pr7Hit(t, g, dragon, me.ID, 2)
	pr6Damage(t, g, dragon, bear, 3)
	pr7Hit(t, g, other, me.ID, 1)
	if me.Life != life-1 || pr6Marked(g, bear) != 0 {
		t.Fatalf("life %d (want %d), bear damage %d (want 0)", me.Life, life-1, pr6Marked(g, bear))
	}
}

// Auriok Replica protects only you: the chosen source's damage to your
// creature is dealt.
func TestADR0108PR6AuriokReplicaProtectsOnlyYou(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	replica := pushCatalogPermanent(g, me.ID, "Auriok Replica", "Artifact Creature — Cleric", pr6AuriokReplica, false)
	src := pr7Creature(g, opp.ID, "Pinger", 1, "R")
	bear := pr7Creature(g, me.ID, "Bear", 2, "G")
	pr7Activate(t, g, me.ID, replica, 0, game.ActivateAbilityParams{})
	pr7Choose(t, g, me.ID, src)
	life := me.Life
	pr7Hit(t, g, src, me.ID, 2)
	pr7Hit(t, g, src, me.ID, 2)
	pr6Damage(t, g, src, bear, 1)
	if me.Life != life || pr6Marked(g, bear) != 1 {
		t.Fatalf("life %d (want %d), bear damage %d (want 1)", me.Life, life, pr6Marked(g, bear))
	}
}

// Burrenton Forge-Tender offers only red sources (CR 615.9) and prevents
// the chosen one's damage to anything.
func TestADR0108PR6ForgeTenderOffersOnlyRedSources(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tender := pushCatalogPermanent(g, me.ID, "Burrenton Forge-Tender", "Creature — Kithkin Wizard", pr6ForgeTender, false)
	red := pr7Creature(g, opp.ID, "Red", 3, "R")
	green := pr7Creature(g, opp.ID, "Green", 3, "G")
	bear := pr7Creature(g, me.ID, "Bear", 2, "G")
	pr7Activate(t, g, me.ID, tender, 0, game.ActivateAbilityParams{})
	c := pr7SourcePrompt(t, g)
	if pr7Offers(c, green) || !pr7Offers(c, red) {
		t.Fatalf("offered %v: want the red source and not the green one", c.ChooseCards)
	}
	pr7Choose(t, g, me.ID, red)
	life := me.Life
	pr7Hit(t, g, red, me.ID, 3)
	pr6Damage(t, g, red, bear, 3)
	if me.Life != life || pr6Marked(g, bear) != 0 {
		t.Fatalf("life %d (want %d), bear damage %d (want 0)", me.Life, life, pr6Marked(g, bear))
	}
}

// Prahv, Spires of Order: the land's shield.
func TestADR0108PR6PrahvShieldsAgainstTheSource(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	prahv := pushCatalogPermanent(g, me.ID, "Prahv, Spires of Order", "Land", pr6Prahv, false)
	src := pr7Creature(g, opp.ID, "Src", 3, "B")
	pr7Activate(t, g, me.ID, prahv, 0, game.ActivateAbilityParams{})
	pr7Choose(t, g, me.ID, src)
	life := me.Life
	pr7Hit(t, g, src, me.ID, 4)
	pr7Hit(t, g, src, me.ID, 4)
	if me.Life != life {
		t.Fatalf("life %d, want %d", me.Life, life)
	}
}

// Rith's Charm's third mode is the shield; its second makes three
// Saprolings.
func TestADR0108PR6RithsCharmModes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pr7Creature(g, opp.ID, "Src", 3, "B")
	castCatalogSpellWithModes(t, g, "Rith's Charm", "Instant", pr6RithsCharm, []int{2})
	passPriorityAroundTable(t, g)
	pr7Choose(t, g, me.ID, src)
	life := me.Life
	pr7Hit(t, g, src, me.ID, 4)
	if me.Life != life {
		t.Fatalf("life %d, want %d", me.Life, life)
	}

	before := countBattlefieldNamed(g, me.ID, "Saproling")
	castCatalogSpellWithModes(t, g, "Rith's Charm", "Instant", pr6RithsCharm, []int{1})
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Saproling") - before; got != 3 {
		t.Errorf("%d Saprolings made, want 3", got)
	}
}

// Dark Sphere prevents half of the next instance's damage, rounded down.
func TestADR0108PR6DarkSphereHalvesRoundedDown(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sphere := pushCatalogPermanent(g, me.ID, "Dark Sphere", "Artifact", pr6DarkSphere, false)
	src := pr7Creature(g, opp.ID, "Src", 7, "R")
	pr7Activate(t, g, me.ID, sphere, 0, game.ActivateAbilityParams{})
	pr7Choose(t, g, me.ID, src)
	life := me.Life
	pr7Hit(t, g, src, me.ID, 7)
	if me.Life != life-4 {
		t.Fatalf("life %d, want %d: 3 of 7 prevented", me.Life, life-4)
	}
	pr7Hit(t, g, src, me.ID, 2)
	if me.Life != life-6 {
		t.Fatalf("life %d, want %d: the next instance is dealt in full", me.Life, life-6)
	}
}

// Healing Grace: the next 3 damage from the chosen source to the target,
// across instances, and 3 life.
func TestADR0108PR6HealingGraceChargedAndLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pr7Creature(g, opp.ID, "Src", 3, "R")
	castCatalogSpell(t, g, "Healing Grace", "Instant", pr6HealingGrace, []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	passPriorityAroundTable(t, g)
	pr7Choose(t, g, me.ID, src)
	life := me.Life
	pr7Hit(t, g, src, me.ID, 2)
	pr7Hit(t, g, src, me.ID, 2)
	if me.Life != life-1 {
		t.Fatalf("life %d, want %d: 2 then 1 more prevented", me.Life, life-1)
	}
	pr7Hit(t, g, src, me.ID, 2)
	if me.Life != life-3 {
		t.Fatalf("life %d, want %d: the shield is spent", me.Life, life-3)
	}
}

// Samite Ministration: damage from a red source prevented gains that
// much life, through a trigger on the stack; a green source's prevented
// damage gains nothing.
func TestADR0108PR6SamiteMinistrationGainsFromBlackOrRed(t *testing.T) {
	for _, tc := range []struct {
		colour string
		gain   int
	}{{"R", 3}, {"B", 3}, {"G", 0}} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		src := pr7Creature(g, opp.ID, "Src", 3, tc.colour)
		castCatalogSpell(t, g, "Samite Ministration", "Instant", pr6SamiteMinistry, nil)
		passPriorityAroundTable(t, g)
		pr7Choose(t, g, me.ID, src)
		life := me.Life
		pr7Hit(t, g, src, me.ID, 3)
		passPriorityAroundTable(t, g)
		if me.Life != life+tc.gain {
			t.Errorf("%s source: life %d, want %d", tc.colour, me.Life, life+tc.gain)
		}
	}
}

// Shieldmage Advocate returns the opponent's card and shields the second
// target.
func TestADR0108PR6ShieldmageAdvocate(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	adv := pushCatalogPermanent(g, me.ID, "Shieldmage Advocate", "Creature — Human Cleric", pr6ShieldmageAdv, false)
	src := pr7Creature(g, opp.ID, "Src", 3, "R")
	bear := pr7Creature(g, me.ID, "Bear", 2, "G")
	dead := game.Card{InstanceID: uuid.New(), Name: "Dead Thing", TypeLine: "Creature — Test", Owner: opp.ID, Controller: opp.ID}
	opp.Graveyard.PushTop(dead)
	pr7Activate(t, g, me.ID, adv, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: dead.InstanceID, Slot: 0},
		{Kind: game.TargetCard, ID: bear, Slot: 1},
	}})
	pr7Choose(t, g, me.ID, src)
	if !opp.Hand.Contains(dead.InstanceID) {
		t.Error("the opponent's card did not go back to their hand")
	}
	pr6Damage(t, g, src, bear, 3)
	pr6Damage(t, g, src, bear, 3)
	if pr6Marked(g, bear) != 0 {
		t.Fatalf("bear damage %d, want 0", pr6Marked(g, bear))
	}
}

// Protective Sphere prevents only sources sharing a colour of the mana
// spent; colourless mana prevents nothing.
func TestADR0108PR6ProtectiveSphereReadsTheManaSpent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sphere := pushCatalogPermanent(g, me.ID, "Protective Sphere", "Enchantment", pr6ProtectiveSphere, false)
	red := pr7Creature(g, opp.ID, "Red", 3, "R")
	green := pr7Creature(g, opp.ID, "Green", 3, "G")
	g.WithWriteLock(func() {
		if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}"); err != nil {
			t.Fatal(err)
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, sphere, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := pr7SourcePrompt(t, g)
	if pr7Offers(c, green) || !pr7Offers(c, red) {
		t.Fatalf("offered %v: want the red source only", c.ChooseCards)
	}
	pr7Choose(t, g, me.ID, red)
	life := me.Life
	pr7Hit(t, g, red, me.ID, 3)
	pr7Hit(t, g, green, me.ID, 1)
	if me.Life != life-1 {
		t.Fatalf("life %d, want %d", me.Life, life-1)
	}
}

// Mourner's Shield: the imprinted card's colours are the sources the
// shield can be against.
func TestADR0108PR6MournersShieldImprintsAColour(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := game.Card{InstanceID: uuid.New(), Name: "Black Thing", TypeLine: "Creature — Test", Colors: []string{"B"}, Owner: opp.ID, Controller: opp.ID}
	opp.Graveyard.PushTop(dead)
	shield := castCatalogSpell(t, g, "Mourner's Shield", "Artifact", pr6MournersShield, nil)
	for i := 0; i < 8 && !inExile(g, dead.InstanceID); i++ {
		passPriorityAroundTable(t, g)
		for _, c := range g.PendingChoices {
			if c == nil {
				continue
			}
			switch c.Kind {
			case game.PendingChoiceTriggerPrompt:
				if err := g.ResolveTriggerPrompt(c.ID, c.Chooser, true); err != nil {
					t.Fatal(err)
				}
			case game.PendingChoicePickTarget:
				if err := g.ResolvePickTarget(c.ID, c.Chooser, game.TargetRef{Kind: game.TargetCard, ID: dead.InstanceID}); err != nil {
					t.Fatal(err)
				}
			}
			break
		}
	}
	if !inExile(g, dead.InstanceID) {
		t.Fatal("the imprint exiled nothing")
	}
	black := pr7Creature(g, opp.ID, "Black", 3, "B")
	red := pr7Creature(g, opp.ID, "Red", 3, "R")
	pr7Activate(t, g, me.ID, shield, 0, game.ActivateAbilityParams{})
	c := pr7SourcePrompt(t, g)
	if pr7Offers(c, red) || !pr7Offers(c, black) {
		t.Fatalf("offered %v: want the black source only", c.ChooseCards)
	}
	pr7Choose(t, g, me.ID, black)
	life := me.Life
	pr7Hit(t, g, black, me.ID, 3)
	if me.Life != life {
		t.Fatalf("life %d, want %d", me.Life, life)
	}
}

// Hidden Retreat pays its library cost and prevents all of the targeted
// spell's damage, with no source prompt.
func TestADR0108PR6HiddenRetreatShieldsAgainstTheSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	retreat := pushCatalogPermanent(g, me.ID, "Hidden Retreat", "Enchantment", "a116329a-343e-4f10-a122-38bf8b5ac2c8", false)
	card := p7Hand(me, "Card", "Sorcery", "{R}")
	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", "4457ed35-7c10-48c8-9776-456485fdf070",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, retreat, 0, game.ActivateAbilityParams{
		TopIDs:  []uuid.UUID{card},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}},
	}); err != nil {
		t.Fatalf("Hidden Retreat: %v", err)
	}
	if top, _ := me.Library.Top(); top.InstanceID != card {
		t.Error("the cost's card is not on top of the library")
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseSource {
			t.Fatal("a targeted source asks no choose_source")
		}
	}
	if opp.Life != life {
		t.Fatalf("opponent at %d, want %d: the Bolt's damage is prevented", opp.Life, life)
	}
}
