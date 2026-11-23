package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pay No Heed: every instance of the chosen source's damage this turn,
// to anything, is prevented; another source's is not.
func TestADR0108PR6PayNoHeedPreventsTheSourceAllTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := effectstest.Pr7Creature(g, opp.ID, "Dragon", 5, "R")
	other := effectstest.Pr7Creature(g, opp.ID, "Other", 2, "R")
	bear := effectstest.Pr7Creature(g, me.ID, "Bear", 2, "G")
	effectstest.CastCatalogSpell(t, g, "Pay No Heed", "Instant", effectstest.Pr6PayNoHeed, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Pr7Choose(t, g, me.ID, dragon)
	life := me.Life
	effectstest.Pr7Hit(t, g, dragon, me.ID, 3)
	effectstest.Pr7Hit(t, g, dragon, me.ID, 2)
	effectstest.Pr6Damage(t, g, dragon, bear, 3)
	effectstest.Pr7Hit(t, g, other, me.ID, 1)
	if me.Life != life-1 || effectstest.Pr6Marked(g, bear) != 0 {
		t.Fatalf("life %d (want %d), bear damage %d (want 0)", me.Life, life-1, effectstest.Pr6Marked(g, bear))
	}
}

// Auriok Replica protects only you: the chosen source's damage to your
// creature is dealt.
func TestADR0108PR6AuriokReplicaProtectsOnlyYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	replica := effectstest.PushCatalogPermanent(g, me.ID, "Auriok Replica", "Artifact Creature — Cleric", effectstest.Pr6AuriokReplica, false)
	src := effectstest.Pr7Creature(g, opp.ID, "Pinger", 1, "R")
	bear := effectstest.Pr7Creature(g, me.ID, "Bear", 2, "G")
	effectstest.Pr7Activate(t, g, me.ID, replica, 0, game.ActivateAbilityParams{})
	effectstest.Pr7Choose(t, g, me.ID, src)
	life := me.Life
	effectstest.Pr7Hit(t, g, src, me.ID, 2)
	effectstest.Pr7Hit(t, g, src, me.ID, 2)
	effectstest.Pr6Damage(t, g, src, bear, 1)
	if me.Life != life || effectstest.Pr6Marked(g, bear) != 1 {
		t.Fatalf("life %d (want %d), bear damage %d (want 1)", me.Life, life, effectstest.Pr6Marked(g, bear))
	}
}

// Burrenton Forge-Tender offers only red sources (CR 615.9) and prevents
// the chosen one's damage to anything.
func TestADR0108PR6ForgeTenderOffersOnlyRedSources(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tender := effectstest.PushCatalogPermanent(g, me.ID, "Burrenton Forge-Tender", "Creature — Kithkin Wizard", effectstest.Pr6ForgeTender, false)
	red := effectstest.Pr7Creature(g, opp.ID, "Red", 3, "R")
	green := effectstest.Pr7Creature(g, opp.ID, "Green", 3, "G")
	bear := effectstest.Pr7Creature(g, me.ID, "Bear", 2, "G")
	effectstest.Pr7Activate(t, g, me.ID, tender, 0, game.ActivateAbilityParams{})
	c := effectstest.Pr7SourcePrompt(t, g)
	if effectstest.Pr7Offers(c, green) || !effectstest.Pr7Offers(c, red) {
		t.Fatalf("offered %v: want the red source and not the green one", c.ChooseCards)
	}
	effectstest.Pr7Choose(t, g, me.ID, red)
	life := me.Life
	effectstest.Pr7Hit(t, g, red, me.ID, 3)
	effectstest.Pr6Damage(t, g, red, bear, 3)
	if me.Life != life || effectstest.Pr6Marked(g, bear) != 0 {
		t.Fatalf("life %d (want %d), bear damage %d (want 0)", me.Life, life, effectstest.Pr6Marked(g, bear))
	}
}

// Prahv, Spires of Order: the land's shield.
func TestADR0108PR6PrahvShieldsAgainstTheSource(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	prahv := effectstest.PushCatalogPermanent(g, me.ID, "Prahv, Spires of Order", "Land", effectstest.Pr6Prahv, false)
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 3, "B")
	effectstest.Pr7Activate(t, g, me.ID, prahv, 0, game.ActivateAbilityParams{})
	effectstest.Pr7Choose(t, g, me.ID, src)
	life := me.Life
	effectstest.Pr7Hit(t, g, src, me.ID, 4)
	effectstest.Pr7Hit(t, g, src, me.ID, 4)
	if me.Life != life {
		t.Fatalf("life %d, want %d", me.Life, life)
	}
}

// Rith's Charm's third mode is the shield; its second makes three
// Saprolings.
func TestADR0108PR6RithsCharmModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 3, "B")
	effectstest.CastCatalogSpellWithModes(t, g, "Rith's Charm", "Instant", effectstest.Pr6RithsCharm, []int{2})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Pr7Choose(t, g, me.ID, src)
	life := me.Life
	effectstest.Pr7Hit(t, g, src, me.ID, 4)
	if me.Life != life {
		t.Fatalf("life %d, want %d", me.Life, life)
	}

	before := effectstest.CountBattlefieldNamed(g, me.ID, "Saproling")
	effectstest.CastCatalogSpellWithModes(t, g, "Rith's Charm", "Instant", effectstest.Pr6RithsCharm, []int{1})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Saproling") - before; got != 3 {
		t.Errorf("%d Saprolings made, want 3", got)
	}
}

// Dark Sphere prevents half of the next instance's damage, rounded down.
func TestADR0108PR6DarkSphereHalvesRoundedDown(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sphere := effectstest.PushCatalogPermanent(g, me.ID, "Dark Sphere", "Artifact", effectstest.Pr6DarkSphere, false)
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 7, "R")
	effectstest.Pr7Activate(t, g, me.ID, sphere, 0, game.ActivateAbilityParams{})
	effectstest.Pr7Choose(t, g, me.ID, src)
	life := me.Life
	effectstest.Pr7Hit(t, g, src, me.ID, 7)
	if me.Life != life-4 {
		t.Fatalf("life %d, want %d: 3 of 7 prevented", me.Life, life-4)
	}
	effectstest.Pr7Hit(t, g, src, me.ID, 2)
	if me.Life != life-6 {
		t.Fatalf("life %d, want %d: the next instance is dealt in full", me.Life, life-6)
	}
}

// Healing Grace: the next 3 damage from the chosen source to the target,
// across instances, and 3 life.
func TestADR0108PR6HealingGraceChargedAndLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 3, "R")
	effectstest.CastCatalogSpell(t, g, "Healing Grace", "Instant", effectstest.Pr6HealingGrace, []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Pr7Choose(t, g, me.ID, src)
	life := me.Life
	effectstest.Pr7Hit(t, g, src, me.ID, 2)
	effectstest.Pr7Hit(t, g, src, me.ID, 2)
	if me.Life != life-1 {
		t.Fatalf("life %d, want %d: 2 then 1 more prevented", me.Life, life-1)
	}
	effectstest.Pr7Hit(t, g, src, me.ID, 2)
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
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		src := effectstest.Pr7Creature(g, opp.ID, "Src", 3, tc.colour)
		effectstest.CastCatalogSpell(t, g, "Samite Ministration", "Instant", effectstest.Pr6SamiteMinistry, nil)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.Pr7Choose(t, g, me.ID, src)
		life := me.Life
		effectstest.Pr7Hit(t, g, src, me.ID, 3)
		effectstest.PassPriorityAroundTable(t, g)
		if me.Life != life+tc.gain {
			t.Errorf("%s source: life %d, want %d", tc.colour, me.Life, life+tc.gain)
		}
	}
}

// Shieldmage Advocate returns the opponent's card and shields the second
// target.
func TestADR0108PR6ShieldmageAdvocate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	adv := effectstest.PushCatalogPermanent(g, me.ID, "Shieldmage Advocate", "Creature — Human Cleric", effectstest.Pr6ShieldmageAdv, false)
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 3, "R")
	bear := effectstest.Pr7Creature(g, me.ID, "Bear", 2, "G")
	dead := game.Card{InstanceID: uuid.New(), Name: "Dead Thing", TypeLine: "Creature — Test", Owner: opp.ID, Controller: opp.ID}
	opp.Graveyard.PushTop(dead)
	effectstest.Pr7Activate(t, g, me.ID, adv, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: dead.InstanceID, Slot: 0},
		{Kind: game.TargetCard, ID: bear, Slot: 1},
	}})
	effectstest.Pr7Choose(t, g, me.ID, src)
	if !opp.Hand.Contains(dead.InstanceID) {
		t.Error("the opponent's card did not go back to their hand")
	}
	effectstest.Pr6Damage(t, g, src, bear, 3)
	effectstest.Pr6Damage(t, g, src, bear, 3)
	if effectstest.Pr6Marked(g, bear) != 0 {
		t.Fatalf("bear damage %d, want 0", effectstest.Pr6Marked(g, bear))
	}
}

// Protective Sphere prevents only sources sharing a colour of the mana
// spent; colourless mana prevents nothing.
func TestADR0108PR6ProtectiveSphereReadsTheManaSpent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sphere := effectstest.PushCatalogPermanent(g, me.ID, "Protective Sphere", "Enchantment", effectstest.Pr6ProtectiveSphere, false)
	red := effectstest.Pr7Creature(g, opp.ID, "Red", 3, "R")
	green := effectstest.Pr7Creature(g, opp.ID, "Green", 3, "G")
	g.WithWriteLock(func() {
		if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}"); err != nil {
			t.Fatal(err)
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, sphere, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.Pr7SourcePrompt(t, g)
	if effectstest.Pr7Offers(c, green) || !effectstest.Pr7Offers(c, red) {
		t.Fatalf("offered %v: want the red source only", c.ChooseCards)
	}
	effectstest.Pr7Choose(t, g, me.ID, red)
	life := me.Life
	effectstest.Pr7Hit(t, g, red, me.ID, 3)
	effectstest.Pr7Hit(t, g, green, me.ID, 1)
	if me.Life != life-1 {
		t.Fatalf("life %d, want %d", me.Life, life-1)
	}
}

// Mourner's Shield: the imprinted card's colours are the sources the
// shield can be against.
func TestADR0108PR6MournersShieldImprintsAColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := game.Card{InstanceID: uuid.New(), Name: "Black Thing", TypeLine: "Creature — Test", Colors: []string{"B"}, Owner: opp.ID, Controller: opp.ID}
	opp.Graveyard.PushTop(dead)
	shield := effectstest.CastCatalogSpell(t, g, "Mourner's Shield", "Artifact", effectstest.Pr6MournersShield, nil)
	for i := 0; i < 8 && !effectstest.InExile(g, dead.InstanceID); i++ {
		effectstest.PassPriorityAroundTable(t, g)
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
	if !effectstest.InExile(g, dead.InstanceID) {
		t.Fatal("the imprint exiled nothing")
	}
	black := effectstest.Pr7Creature(g, opp.ID, "Black", 3, "B")
	red := effectstest.Pr7Creature(g, opp.ID, "Red", 3, "R")
	effectstest.Pr7Activate(t, g, me.ID, shield, 0, game.ActivateAbilityParams{})
	c := effectstest.Pr7SourcePrompt(t, g)
	if effectstest.Pr7Offers(c, red) || !effectstest.Pr7Offers(c, black) {
		t.Fatalf("offered %v: want the black source only", c.ChooseCards)
	}
	effectstest.Pr7Choose(t, g, me.ID, black)
	life := me.Life
	effectstest.Pr7Hit(t, g, black, me.ID, 3)
	if me.Life != life {
		t.Fatalf("life %d, want %d", me.Life, life)
	}
}

// Hidden Retreat pays its library cost and prevents all of the targeted
// spell's damage, with no source prompt.
func TestADR0108PR6HiddenRetreatShieldsAgainstTheSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	retreat := effectstest.PushCatalogPermanent(g, me.ID, "Hidden Retreat", "Enchantment", "a116329a-343e-4f10-a122-38bf8b5ac2c8", false)
	card := p7Hand(me, "Card", "Sorcery", "{R}")
	bolt := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", "4457ed35-7c10-48c8-9776-456485fdf070",
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
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseSource {
			t.Fatal("a targeted source asks no choose_source")
		}
	}
	if opp.Life != life {
		t.Fatalf("opponent at %d, want %d: the Bolt's damage is prevented", opp.Life, life)
	}
}
