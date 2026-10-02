package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// adr0107_pr7_cards_test.go — ADR 0107 PR 7 (#1860): the cards the
// next-damage shield unblocks, each against the damage it must stop and
// the damage it must not.

const (
	pr7CoPRed            = "df2738fe-9cd1-4347-8808-105fcfde1190"
	pr7CoPShadow         = "07c42f46-800e-4d2a-b844-c125ab93190d"
	pr7RoPLands          = "605dd1b5-8946-4bd0-b31b-e99bf63a1bc4"
	pr7GreaterRealm      = "b03eb0c4-89a4-420d-8a23-1a868a07d9cf"
	pr7StoryCircle       = "7071aee8-b5ca-4be5-9ba0-2df7e3af303b"
	pr7CircleOfSolace    = "14ebedd1-7a2d-413b-9b76-8610511292a4"
	pr7KithkinArmor      = "35e2730e-b3aa-4080-aae8-7b3ad1e59848"
	pr7SamiteBlessing    = "a524ae69-2594-467d-af0a-27b9984d4300"
	pr7Mercenaries       = "9d96671c-7a98-48ec-b479-bc451e18264d"
	pr7PilgrimOfJustice  = "959b672e-d4e4-49a1-874a-5148f040fadd"
	pr7MartyrsCause      = "72d1c789-e79f-4f15-80d0-981d4120b085"
	pr7BoneMask          = "ddfc7dac-f921-4529-935c-6c588890c522"
	pr7DeflectingPalm    = "dc5dffc8-fac5-4956-bac4-1ad2cc16f6be"
	pr7HonorablePassage  = "66805334-1015-4b41-ba9b-116de94b0744"
	pr7Shadowbane        = "9e9ec33b-b923-40e5-81a8-e1e78df63ec7"
	pr7AweStrike         = "994bcde3-4a76-403d-a6e1-88609a13a99c"
	pr7DazzlingReflect   = "44db1ff5-7082-4da4-8f4f-4a2d2e3d7666"
	pr7NewWayForward     = "429368e6-3de1-4a44-a062-86cbbb73e243"
	pr7InterventionPact  = "9f9882f5-2338-4881-b418-b15348a462b4"
	pr7ReverseDamage     = "eaaf7c30-f463-4115-a40e-7dc717063413"
	pr7Invulnerability   = "2c964157-d414-4d4e-822b-12ce61e882ff"
	pr7HaazdaShieldMate  = "49b689a3-9197-4a04-a62f-218b245d6e23"
	pr7PrismaticCircle   = "022dae2e-7fc3-486e-9168-59652d9ab21c"
	pr7ChoArrimAlchemist = "1e1d4d81-4a45-41b5-a781-ae2227b4f4c0"
)

// pr7Creature puts a creature of `colors` on the battlefield for `owner`.
func pr7Creature(g *game.Game, owner uuid.UUID, name string, power int, colors ...string) uuid.UUID {
	return apaPush(g, owner, owner, game.Card{Name: name, TypeLine: "Creature — Test", Power: power, Toughness: 4, Colors: colors})
}

// pr7Activate has `who` activate row `index` of `source` in permissive
// mode and passes priority until a choice is open or the stack is empty.
func pr7Activate(t *testing.T, g *game.Game, who, source uuid.UUID, index int, params game.ActivateAbilityParams) {
	t.Helper()
	if err := g.ActivateCatalogAbility(who, source, index, params); err != nil {
		t.Fatalf("activate row %d: %v", index, err)
	}
	passPriorityAroundTable(t, g)
}

// pr7SourcePrompt is the open choose_source prompt.
func pr7SourcePrompt(t *testing.T, g *game.Game) *game.PendingChoice {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseSource {
			return c
		}
	}
	t.Fatal("no choose_source prompt open")
	return nil
}

// pr7Choose answers the open choose_source prompt with `source`.
func pr7Choose(t *testing.T, g *game.Game, chooser, source uuid.UUID) {
	t.Helper()
	c := pr7SourcePrompt(t, g)
	if err := g.ResolveChooseSource(c.ID, chooser, []uuid.UUID{source}); err != nil {
		t.Fatalf("choose source: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func pr7Offers(c *game.PendingChoice, id uuid.UUID) bool {
	for _, x := range c.ChooseCards {
		if x == id {
			return true
		}
	}
	return false
}

// pr7Hit deals `n` damage from `source` to the player `to`. Each call is
// one damage instruction and so one instance of damage from a source
// (CR 615.8, ADR 0108 PR 0); the shields' "prevented this way" follow-ups
// run as it ends (CR 615.5).
//
// It then runs the priority boundary, ADR 0107's catch-all flush point.
func pr7Hit(t *testing.T, g *game.Game, source, to uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(source, to, n); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
}

// pr7Pridemate puts an Ajani's Pridemate ("Whenever you gain life, put a
// +1/+1 counter on this creature") on the battlefield for `owner`: one
// counter per life-gain EVENT.
func pr7Pridemate(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushCatalogPermanent(g, owner, "Ajani's Pridemate", "Creature — Cat Soldier", "95e94dea-5ac0-4d6f-adec-ca147aee861f", false)
}

func pr7Counters(g *game.Game, id uuid.UUID) int {
	return findBattlefieldCardForTest(g, id).Counters[game.CounterPlusOne]
}

// "The damage prevented this way" is the INSTANCE's total (CR 615.5,
// 615.8): a 5-power trampler blocked by two, under Awe Strike, is three
// events of combat damage and one instance, so its controller gains 5
// life in one event — one Pridemate trigger, not three.
func TestPR7AweStrikeTotalsATramplersSplitDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pridemate := pr7Pridemate(g, me.ID)
	trampler := apaPush(g, me.ID, me.ID, game.Card{Name: "Trampler", TypeLine: "Creature — Beast", Power: 5, Toughness: 5,
		Colors: []string{"G"}, Keywords: []string{"trample"}})
	b1 := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Blocker One", TypeLine: "Creature — Test", Power: 0, Toughness: 2})
	b2 := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Blocker Two", TypeLine: "Creature — Test", Power: 0, Toughness: 2})
	castCatalogSpell(t, g, "Awe Strike", "Instant", pr7AweStrike, []game.TargetRef{{Kind: game.TargetCard, ID: trampler}})
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(trampler, opp.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	for _, b := range []uuid.UUID{b1, b2} {
		if err := g.DeclareBlocker(b, trampler); err != nil {
			t.Fatal(err)
		}
	}
	life, theirs := me.Life, opp.Life
	advanceTo(t, g, game.StepCombatDamage)
	for i := 0; i < 4; i++ {
		var c *game.PendingChoice
		for _, p := range g.PendingChoices {
			if p != nil && p.Kind == game.PendingChoiceDamageAssignment {
				c = p
			}
		}
		if c == nil {
			break
		}
		// Lethal to each 2-toughness blocker (CR 702.19c) and the
		// rest to the player: three events.
		entries := []game.DamageAssignmentEntry{{BlockerID: b1, Amount: 2}, {BlockerID: b2, Amount: 2}}
		if err := g.ResolveDamageAssignment(c.ID, c.Chooser, entries, 1); err != nil {
			t.Fatalf("assign: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if damageMarkedOn(g, b1) != 0 || damageMarkedOn(g, b2) != 0 || opp.Life != theirs {
		t.Fatalf("the trampler's damage got through: %d, %d, opponent %d → %d",
			damageMarkedOn(g, b1), damageMarkedOn(g, b2), theirs, opp.Life)
	}
	if me.Life != life+5 {
		t.Errorf("life %d → %d, want +5", life, me.Life)
	}
	if got := pr7Counters(g, pridemate); got != 1 {
		t.Errorf("Pridemate has %d counters, want 1: one life-gain event", got)
	}
}

// The same for a follow-up split across two events of one instance:
// Reverse Damage gains the total once. One instance is one damage
// instruction (ADR 0108 PR 0), so the two events are dealt inside one
// damage-instance scope; TestPR0ReverseDamageMeetsOnlyTheFirstInstance
// is the two-instruction case.
func TestPR7ReverseDamageGainsOnceForASplitInstance(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pridemate := pr7Pridemate(g, me.ID)
	src := pr7Creature(g, opp.ID, "Src", 6, "R")
	castCatalogSpell(t, g, "Reverse Damage", "Instant", pr7ReverseDamage, nil)
	passPriorityAroundTable(t, g)
	pr7Choose(t, g, me.ID, src)
	life := me.Life
	g.WithWriteLock(func() {
		_ = g.DamageInstanceForEffect(func() error {
			_ = g.DealDamageToPlayerForEffect(src, me.ID, 4)
			return g.DealDamageToPlayerForEffect(src, me.ID, 2)
		})
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if me.Life != life+6 {
		t.Errorf("life %d → %d, want +6 in one gain", life, me.Life)
	}
	if got := pr7Counters(g, pridemate); got != 1 {
		t.Errorf("Pridemate has %d counters, want 1", got)
	}
}

// Every card of the batch is registered with its printed rows.
func TestPR7CardsRegisterTheirRows(t *testing.T) {
	cases := map[string]int{
		pr7CoPRed: 1, pr7CoPShadow: 1, pr7RoPLands: 2, pr7GreaterRealm: 1, pr7StoryCircle: 1,
		pr7CircleOfSolace: 1, pr7KithkinArmor: 1, pr7Mercenaries: 1, pr7PilgrimOfJustice: 1,
		pr7MartyrsCause: 1, pr7BoneMask: 1, pr7HaazdaShieldMate: 1, pr7PrismaticCircle: 1,
		pr7ChoArrimAlchemist: 1,
	}
	for oracle, n := range cases {
		if got := len(game.ActivatedAbilitiesForCard(game.Card{OracleID: oracle})); got != n {
			t.Errorf("%s: %d activated rows, want %d", oracle, got, n)
		}
	}
	if rows := game.ActivatedAbilitiesForCard(game.Card{OracleID: pr7Mercenaries}); len(rows) == 1 && !rows[0].AnyPlayer {
		t.Error("Mercenaries' row is an any-player row")
	}
}

// Circle of Protection: Red end to end: the prompt offers only red
// sources, the chosen one's next damage is prevented, and only once.
func TestPR7CircleOfProtectionRed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cop := pushCatalogPermanent(g, me.ID, "Circle of Protection: Red", "Enchantment", pr7CoPRed, false)
	red := pr7Creature(g, opp.ID, "Red Dragon", 5, "R")
	green := pr7Creature(g, opp.ID, "Green Bear", 2, "G")
	pr7Activate(t, g, me.ID, cop, 0, game.ActivateAbilityParams{})
	c := pr7SourcePrompt(t, g)
	if !pr7Offers(c, red) || pr7Offers(c, green) || c.Chooser != me.ID {
		t.Fatalf("prompt for %v offers %v: want the red creature and not the green one", c.Chooser, c.ChooseCards)
	}
	pr7Choose(t, g, me.ID, red)
	life := me.Life
	pr7Hit(t, g, green, me.ID, 2)
	pr7Hit(t, g, red, me.ID, 5)
	if me.Life != life-2 {
		t.Fatalf("life %d → %d, want only the green 2", life, me.Life)
	}
	// Play moves on (a new step is a new event batch): the next damage
	// from the dragon is a new instance.
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatal(err)
	}
	pr7Hit(t, g, red, me.ID, 5)
	if me.Life != life-7 {
		t.Errorf("life %d → %d, want the red source's second instance dealt", life, me.Life)
	}
}

// Circle of Protection: Shadow names a creature with shadow.
func TestPR7CircleOfProtectionShadowOffersShadowCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cop := pushCatalogPermanent(g, me.ID, "Circle of Protection: Shadow", "Enchantment", pr7CoPShadow, false)
	shade := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Shade", TypeLine: "Creature — Test", Power: 2, Toughness: 2, Keywords: []string{"shadow"}})
	plain := pr7Creature(g, opp.ID, "Plain", 2, "B")
	pr7Activate(t, g, me.ID, cop, 0, game.ActivateAbilityParams{})
	c := pr7SourcePrompt(t, g)
	if !pr7Offers(c, shade) || pr7Offers(c, plain) {
		t.Errorf("offers %v: want the shadow creature only", c.ChooseCards)
	}
}

// Story Circle: the colour chosen as it entered narrows the choice.
func TestPR7StoryCircleUsesTheChosenColour(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	circle := pushCatalogPermanent(g, me.ID, "Story Circle", "Enchantment", pr7StoryCircle, false)
	findBattlefieldCardForTest(g, circle).ChosenColor = "B"
	black := pr7Creature(g, opp.ID, "Black", 3, "B")
	red := pr7Creature(g, opp.ID, "Red", 3, "R")
	pr7Activate(t, g, me.ID, circle, 0, game.ActivateAbilityParams{})
	c := pr7SourcePrompt(t, g)
	if !pr7Offers(c, black) || pr7Offers(c, red) {
		t.Fatalf("offers %v: want the black creature only", c.ChooseCards)
	}
	pr7Choose(t, g, me.ID, black)
	life := me.Life
	pr7Hit(t, g, black, me.ID, 3)
	if me.Life != life {
		t.Errorf("life %d → %d, want the black source's damage prevented", life, me.Life)
	}
}

// Circle of Solace: no choice; the first creature of the chosen type
// to deal damage to you is the one.
func TestPR7CircleOfSolaceShieldsAgainstTheChosenType(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	circle := pushCatalogPermanent(g, me.ID, "Circle of Solace", "Enchantment", pr7CircleOfSolace, false)
	findBattlefieldCardForTest(g, circle).NamedTribe = "Goblin"
	goblin := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Goblin", TypeLine: "Creature — Goblin", Power: 2, Toughness: 2})
	elf := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Elf", TypeLine: "Creature — Elf", Power: 2, Toughness: 2})
	pr7Activate(t, g, me.ID, circle, 0, game.ActivateAbilityParams{})
	if len(g.PendingChoices) != 0 {
		t.Fatal("Circle of Solace chooses no source")
	}
	life := me.Life
	pr7Hit(t, g, elf, me.ID, 1)
	pr7Hit(t, g, goblin, me.ID, 2)
	if me.Life != life-1 {
		t.Errorf("life %d → %d, want the Elf's 1 dealt and the Goblin's 2 prevented", life, me.Life)
	}
}

// Mercenaries: any player may activate it, and "you" is the activator.
func TestPR7MercenariesShieldsTheActivator(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	merc := pushCatalogPermanent(g, me.ID, "Mercenaries", "Creature — Human Mercenary", pr7Mercenaries, false)
	pr7Activate(t, g, opp.ID, merc, 0, game.ActivateAbilityParams{})
	theirs, mine := opp.Life, me.Life
	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(merc, me.ID, 3)
		_ = g.DealDamageToPlayerForEffect(merc, opp.ID, 3)
	})
	if opp.Life != theirs || me.Life != mine-3 {
		t.Errorf("activator %d → %d (want unchanged), controller %d → %d (want -3)", theirs, opp.Life, mine, me.Life)
	}
}

// Pilgrim of Justice: a red source's next damage, to anything.
func TestPR7PilgrimShieldsAnything(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pilgrim := pushCatalogPermanent(g, me.ID, "Pilgrim of Justice", "Creature — Human Cleric", pr7PilgrimOfJustice, false)
	red := pr7Creature(g, opp.ID, "Red", 3, "R")
	bear := pr7Creature(g, opp.ID, "Bear", 2, "G")
	pr7Activate(t, g, me.ID, pilgrim, 0, game.ActivateAbilityParams{})
	pr7Choose(t, g, me.ID, red)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(red, bear, 3) })
	if got := damageMarkedOn(g, bear); got != 0 {
		t.Errorf("the opponent's own bear took %d, want 0: the shield protects anything", got)
	}
	if findBattlefieldCardForTest(g, pilgrim) != nil {
		t.Error("the Pilgrim is sacrificed to pay")
	}
}

// Martyr's Cause: a target, and a creature sacrificed to pay.
func TestPR7MartyrsCauseShieldsItsTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cause := pushCatalogPermanent(g, me.ID, "Martyr's Cause", "Enchantment", pr7MartyrsCause, false)
	fodder := pr7Creature(g, me.ID, "Fodder", 1, "W")
	knight := pr7Creature(g, me.ID, "Knight", 2, "W")
	src := pr7Creature(g, opp.ID, "Src", 3, "R")
	pr7Activate(t, g, me.ID, cause, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder},
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: knight}},
	})
	pr7Choose(t, g, me.ID, src)
	life := me.Life
	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(src, knight, 3)
		_ = g.DealDamageToPlayerForEffect(src, me.ID, 3)
	})
	if damageMarkedOn(g, knight) != 0 || me.Life != life-3 {
		t.Errorf("knight %d damage (want 0), life %d → %d (want -3: only the target is shielded)",
			damageMarkedOn(g, knight), life, me.Life)
	}
}

// Kithkin Armor: the Aura is gone by resolution, and its creature is
// still the one shielded.
func TestPR7KithkinArmorShieldsTheCreatureItEnchanted(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	knight := pr7Creature(g, me.ID, "Knight", 2, "W")
	armor := pushCatalogPermanent(g, me.ID, "Kithkin Armor", "Enchantment — Aura", pr7KithkinArmor, false)
	g.WithWriteLock(func() {
		findBattlefieldCardForTest(g, armor).AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: knight}
	})
	src := pr7Creature(g, opp.ID, "Src", 3, "R")
	pr7Activate(t, g, me.ID, armor, 0, game.ActivateAbilityParams{})
	pr7Choose(t, g, me.ID, src)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(src, knight, 3) })
	if got := damageMarkedOn(g, knight); got != 0 {
		t.Errorf("the enchanted creature took %d, want 0", got)
	}
}

// Samite Blessing gives the enchanted creature the row.
func TestPR7SamiteBlessingGrantsTheRow(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	knight := pr7Creature(g, me.ID, "Knight", 2, "W")
	blessing := pushCatalogPermanent(g, me.ID, "Samite Blessing", "Enchantment — Aura", pr7SamiteBlessing, false)
	g.WithWriteLock(func() {
		findBattlefieldCardForTest(g, blessing).AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: knight}
	})
	live := apaLive(g, knight)
	if n := len(game.ActivatedAbilitiesForCard(*live)); n != 1 {
		t.Errorf("the enchanted creature has %d activated rows, want the granted one", n)
	}
}

// Deflecting Palm: the damage prevented is dealt to the source's
// controller, by Deflecting Palm.
func TestPR7DeflectingPalmSendsTheDamageBack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pr7Creature(g, opp.ID, "Src", 4, "G")
	castCatalogSpell(t, g, "Deflecting Palm", "Instant", pr7DeflectingPalm, nil)
	passPriorityAroundTable(t, g)
	pr7Choose(t, g, me.ID, src)
	mine, theirs := me.Life, opp.Life
	pr7Hit(t, g, src, me.ID, 4)
	if me.Life != mine || opp.Life != theirs-4 {
		t.Errorf("me %d → %d (want unchanged), them %d → %d (want -4)", mine, me.Life, theirs, opp.Life)
	}
}

// Honorable Passage: only damage from a RED source comes back.
func TestPR7HonorablePassageReturnsOnlyRedDamage(t *testing.T) {
	for _, tc := range []struct {
		color string
		back  int
	}{{"R", 3}, {"G", 0}} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		src := pr7Creature(g, opp.ID, "Src", 3, tc.color)
		castCatalogSpell(t, g, "Honorable Passage", "Instant", pr7HonorablePassage,
			[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
		passPriorityAroundTable(t, g)
		pr7Choose(t, g, me.ID, src)
		mine, theirs := me.Life, opp.Life
		pr7Hit(t, g, src, me.ID, 3)
		if me.Life != mine || opp.Life != theirs-tc.back {
			t.Errorf("%s source: me %d → %d (want unchanged), them %d → %d (want -%d)",
				tc.color, mine, me.Life, theirs, opp.Life, tc.back)
		}
	}
}

// Shadowbane covers your creatures, and gains life only for a black
// source.
func TestPR7ShadowbaneCoversYourCreaturesAndGainsForBlack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pr7Creature(g, me.ID, "Mine", 2, "W")
	src := pr7Creature(g, opp.ID, "Src", 3, "B")
	castCatalogSpell(t, g, "Shadowbane", "Instant", pr7Shadowbane, nil)
	passPriorityAroundTable(t, g)
	pr7Choose(t, g, me.ID, src)
	life := me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(src, mine, 3) })
	g.RunStateChecksForTest()
	if damageMarkedOn(g, mine) != 0 || me.Life != life+3 {
		t.Errorf("creature %d damage (want 0), life %d → %d (want +3)", damageMarkedOn(g, mine), life, me.Life)
	}
}

// Awe Strike: the target creature's next damage is prevented, and you
// gain that much.
func TestPR7AweStrike(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pr7Creature(g, opp.ID, "Src", 5, "R")
	castCatalogSpell(t, g, "Awe Strike", "Instant", pr7AweStrike, []game.TargetRef{{Kind: game.TargetCard, ID: src}})
	passPriorityAroundTable(t, g)
	life := me.Life
	pr7Hit(t, g, src, me.ID, 5)
	if me.Life != life+5 {
		t.Errorf("life %d → %d, want the 5 prevented and 5 gained", life, me.Life)
	}
}

// Dazzling Reflection: life equal to the power now, and the creature's
// next damage prevented.
func TestPR7DazzlingReflection(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pr7Creature(g, opp.ID, "Src", 4, "R")
	life := me.Life
	castCatalogSpell(t, g, "Dazzling Reflection", "Instant", pr7DazzlingReflect, []game.TargetRef{{Kind: game.TargetCard, ID: src}})
	passPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Fatalf("life %d → %d, want +4 for the power", life, me.Life)
	}
	pr7Hit(t, g, src, me.ID, 4)
	if me.Life != life+4 {
		t.Errorf("life %d, want the creature's next damage prevented", me.Life)
	}
}

// Reverse Damage gains the amount prevented; Invulnerability returns
// with buyback.
func TestPR7ReverseDamageGainsWhatItPrevented(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pr7Creature(g, opp.ID, "Src", 6, "R")
	castCatalogSpell(t, g, "Reverse Damage", "Instant", pr7ReverseDamage, nil)
	passPriorityAroundTable(t, g)
	pr7Choose(t, g, me.ID, src)
	life := me.Life
	pr7Hit(t, g, src, me.ID, 6)
	if me.Life != life+6 {
		t.Errorf("life %d → %d, want +6", life, me.Life)
	}
}

// Bone Mask exiles the top of its controller's library, as many cards as
// it prevented.
func TestPR7BoneMaskExilesWhatItPrevented(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mask := pushCatalogPermanent(g, me.ID, "Bone Mask", "Artifact", pr7BoneMask, false)
	src := pr7Creature(g, opp.ID, "Src", 3, "R")
	pr7Activate(t, g, me.ID, mask, 0, game.ActivateAbilityParams{})
	pr7Choose(t, g, me.ID, src)
	lib, exiled := len(me.Library.Cards), len(g.Exile.Cards)
	pr7Hit(t, g, src, me.ID, 3)
	if len(me.Library.Cards) != lib-3 || len(g.Exile.Cards) != exiled+3 {
		t.Errorf("library %d → %d, exile %d → %d: want three cards moved", lib, len(me.Library.Cards), exiled, len(g.Exile.Cards))
	}
}

// New Way Forward: a reflexive trigger deals the damage back and draws.
func TestPR7NewWayForwardStrikesBackAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pr7Creature(g, opp.ID, "Src", 2, "R")
	castCatalogSpell(t, g, "New Way Forward", "Instant", pr7NewWayForward, nil)
	passPriorityAroundTable(t, g)
	pr7Choose(t, g, me.ID, src)
	hand, theirs := len(me.Hand.Cards), opp.Life
	pr7Hit(t, g, src, me.ID, 2)
	passPriorityAroundTable(t, g)
	if opp.Life != theirs-2 || len(me.Hand.Cards) != hand+2 {
		t.Errorf("them %d → %d (want -2), hand %d → %d (want +2)", theirs, opp.Life, hand, len(me.Hand.Cards))
	}
}

// Intervention Pact schedules its upkeep payment, and shields.
func TestPR7InterventionPactSchedulesThePayment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pr7Creature(g, opp.ID, "Src", 2, "R")
	castCatalogSpell(t, g, "Intervention Pact", "Instant", pr7InterventionPact, nil)
	passPriorityAroundTable(t, g)
	pr7Choose(t, g, me.ID, src)
	found := false
	for _, d := range g.DelayedTriggers {
		if d != nil && d.Controller == me.ID && d.Body.Key() == "pact/pay-or-lose" {
			found = true
		}
	}
	if !found {
		t.Error("no pact payment scheduled")
	}
	life := me.Life
	pr7Hit(t, g, src, me.ID, 2)
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2", life, me.Life)
	}
}

// After damage that can't be prevented, a "prevented this way" card does
// nothing and its shield is still there (CR 615.12, 609.7b).
func TestPR7FollowUpUnderUnpreventableDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pr7Creature(g, opp.ID, "Src", 4, "G")
	castCatalogSpell(t, g, "Deflecting Palm", "Instant", pr7DeflectingPalm, nil)
	passPriorityAroundTable(t, g)
	pr7Choose(t, g, me.ID, src)
	mine, theirs := me.Life, opp.Life
	g.WithWriteLock(func() {
		_ = g.DealMarkedDamageForEffect(src, nil, me.ID, 4, game.DamageMarks{CantBePrevented: true})
	})
	g.RunStateChecksForTest()
	if me.Life != mine-4 || opp.Life != theirs {
		t.Fatalf("me %d → %d (want -4), them %d → %d (want unchanged)", mine, me.Life, theirs, opp.Life)
	}
	pr7Hit(t, g, src, me.ID, 4)
	if me.Life != mine-4 || opp.Life != theirs-4 {
		t.Errorf("the unspent shield should take the next damage and send it back: me %d, them %d", me.Life, opp.Life)
	}
}
