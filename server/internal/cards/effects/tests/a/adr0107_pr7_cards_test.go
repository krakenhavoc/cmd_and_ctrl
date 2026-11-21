package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pr7Pridemate puts an Ajani's Pridemate ("Whenever you gain life, put a
// +1/+1 counter on this creature") on the battlefield for `owner`: one
// counter per life-gain EVENT.
func pr7Pridemate(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.PushCatalogPermanent(g, owner, "Ajani's Pridemate", "Creature — Cat Soldier", "95e94dea-5ac0-4d6f-adec-ca147aee861f", false)
}

// "The damage prevented this way" is the INSTANCE's total (CR 615.5,
// 615.8): a 5-power trampler blocked by two, under Awe Strike, is three
// events of combat damage and one instance, so its controller gains 5
// life in one event — one Pridemate trigger, not three.
func TestPR7AweStrikeTotalsATramplersSplitDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pridemate := pr7Pridemate(g, me.ID)
	trampler := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Trampler", TypeLine: "Creature — Beast", Power: 5, Toughness: 5,
		Colors: []string{"G"}, Keywords: []string{"trample"}})
	b1 := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Blocker One", TypeLine: "Creature — Test", Power: 0, Toughness: 2})
	b2 := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Blocker Two", TypeLine: "Creature — Test", Power: 0, Toughness: 2})
	effectstest.CastCatalogSpell(t, g, "Awe Strike", "Instant", effectstest.Pr7AweStrike, []game.TargetRef{{Kind: game.TargetCard, ID: trampler}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(trampler, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	for _, b := range []uuid.UUID{b1, b2} {
		if err := g.DeclareBlocker(b, trampler); err != nil {
			t.Fatal(err)
		}
	}
	life, theirs := me.Life, opp.Life
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
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
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DamageMarkedOn(g, b1) != 0 || effectstest.DamageMarkedOn(g, b2) != 0 || opp.Life != theirs {
		t.Fatalf("the trampler's damage got through: %d, %d, opponent %d → %d",
			effectstest.DamageMarkedOn(g, b1), effectstest.DamageMarkedOn(g, b2), theirs, opp.Life)
	}
	if me.Life != life+5 {
		t.Errorf("life %d → %d, want +5", life, me.Life)
	}
	if got := effectstest.Pr7Counters(g, pridemate); got != 1 {
		t.Errorf("Pridemate has %d counters, want 1: one life-gain event", got)
	}
}

// The same for a follow-up split across two events of one instance:
// Reverse Damage gains the total once. One instance is one damage
// instruction (ADR 0108 PR 0), so the two events are dealt inside one
// damage-instance scope; TestPR0ReverseDamageMeetsOnlyTheFirstInstance
// is the two-instruction case.
func TestPR7ReverseDamageGainsOnceForASplitInstance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pridemate := pr7Pridemate(g, me.ID)
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 6, "R")
	effectstest.CastCatalogSpell(t, g, "Reverse Damage", "Instant", effectstest.Pr7ReverseDamage, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Pr7Choose(t, g, me.ID, src)
	life := me.Life
	g.WithWriteLock(func() {
		_ = g.DamageInstanceForEffect(func() error {
			_ = g.DealDamageToPlayerForEffect(src, me.ID, 4)
			return g.DealDamageToPlayerForEffect(src, me.ID, 2)
		})
	})
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+6 {
		t.Errorf("life %d → %d, want +6 in one gain", life, me.Life)
	}
	if got := effectstest.Pr7Counters(g, pridemate); got != 1 {
		t.Errorf("Pridemate has %d counters, want 1", got)
	}
}

// Every card of the batch is registered with its printed rows.
func TestPR7CardsRegisterTheirRows(t *testing.T) {
	cases := map[string]int{
		effectstest.Pr7CoPRed: 1, effectstest.Pr7CoPShadow: 1, effectstest.Pr7RoPLands: 2, effectstest.Pr7GreaterRealm: 1, effectstest.Pr7StoryCircle: 1,
		effectstest.Pr7CircleOfSolace: 1, effectstest.Pr7KithkinArmor: 1, effectstest.Pr7Mercenaries: 1, effectstest.Pr7PilgrimOfJustice: 1,
		effectstest.Pr7MartyrsCause: 1, effectstest.Pr7BoneMask: 1, effectstest.Pr7HaazdaShieldMate: 1, effectstest.Pr7PrismaticCircle: 1,
		effectstest.Pr7ChoArrimAlchemist: 1,
	}
	for oracle, n := range cases {
		if got := len(game.ActivatedAbilitiesForCard(game.Card{OracleID: oracle})); got != n {
			t.Errorf("%s: %d activated rows, want %d", oracle, got, n)
		}
	}
	if rows := game.ActivatedAbilitiesForCard(game.Card{OracleID: effectstest.Pr7Mercenaries}); len(rows) == 1 && !rows[0].AnyPlayer {
		t.Error("Mercenaries' row is an any-player row")
	}
}

// Circle of Protection: Red end to end: the prompt offers only red
// sources, the chosen one's next damage is prevented, and only once.
func TestPR7CircleOfProtectionRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cop := effectstest.PushCatalogPermanent(g, me.ID, "Circle of Protection: Red", "Enchantment", effectstest.Pr7CoPRed, false)
	red := effectstest.Pr7Creature(g, opp.ID, "Red Dragon", 5, "R")
	green := effectstest.Pr7Creature(g, opp.ID, "Green Bear", 2, "G")
	effectstest.Pr7Activate(t, g, me.ID, cop, 0, game.ActivateAbilityParams{})
	c := effectstest.Pr7SourcePrompt(t, g)
	if !effectstest.Pr7Offers(c, red) || effectstest.Pr7Offers(c, green) || c.Chooser != me.ID {
		t.Fatalf("prompt for %v offers %v: want the red creature and not the green one", c.Chooser, c.ChooseCards)
	}
	effectstest.Pr7Choose(t, g, me.ID, red)
	life := me.Life
	effectstest.Pr7Hit(t, g, green, me.ID, 2)
	effectstest.Pr7Hit(t, g, red, me.ID, 5)
	if me.Life != life-2 {
		t.Fatalf("life %d → %d, want only the green 2", life, me.Life)
	}
	// Play moves on (a new step is a new event batch): the next damage
	// from the dragon is a new instance.
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatal(err)
	}
	effectstest.Pr7Hit(t, g, red, me.ID, 5)
	if me.Life != life-7 {
		t.Errorf("life %d → %d, want the red source's second instance dealt", life, me.Life)
	}
}

// Circle of Protection: Shadow names a creature with shadow.
func TestPR7CircleOfProtectionShadowOffersShadowCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cop := effectstest.PushCatalogPermanent(g, me.ID, "Circle of Protection: Shadow", "Enchantment", effectstest.Pr7CoPShadow, false)
	shade := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Shade", TypeLine: "Creature — Test", Power: 2, Toughness: 2, Keywords: []string{"shadow"}})
	plain := effectstest.Pr7Creature(g, opp.ID, "Plain", 2, "B")
	effectstest.Pr7Activate(t, g, me.ID, cop, 0, game.ActivateAbilityParams{})
	c := effectstest.Pr7SourcePrompt(t, g)
	if !effectstest.Pr7Offers(c, shade) || effectstest.Pr7Offers(c, plain) {
		t.Errorf("offers %v: want the shadow creature only", c.ChooseCards)
	}
}

// Story Circle: the colour chosen as it entered narrows the choice.
func TestPR7StoryCircleUsesTheChosenColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	circle := effectstest.PushCatalogPermanent(g, me.ID, "Story Circle", "Enchantment", effectstest.Pr7StoryCircle, false)
	effectstest.FindBattlefieldCardForTest(g, circle).ChosenColor = "B"
	black := effectstest.Pr7Creature(g, opp.ID, "Black", 3, "B")
	red := effectstest.Pr7Creature(g, opp.ID, "Red", 3, "R")
	effectstest.Pr7Activate(t, g, me.ID, circle, 0, game.ActivateAbilityParams{})
	c := effectstest.Pr7SourcePrompt(t, g)
	if !effectstest.Pr7Offers(c, black) || effectstest.Pr7Offers(c, red) {
		t.Fatalf("offers %v: want the black creature only", c.ChooseCards)
	}
	effectstest.Pr7Choose(t, g, me.ID, black)
	life := me.Life
	effectstest.Pr7Hit(t, g, black, me.ID, 3)
	if me.Life != life {
		t.Errorf("life %d → %d, want the black source's damage prevented", life, me.Life)
	}
}

// Circle of Solace: no choice; the first creature of the chosen type
// to deal damage to you is the one.
func TestPR7CircleOfSolaceShieldsAgainstTheChosenType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	circle := effectstest.PushCatalogPermanent(g, me.ID, "Circle of Solace", "Enchantment", effectstest.Pr7CircleOfSolace, false)
	effectstest.FindBattlefieldCardForTest(g, circle).NamedTribe = "Goblin"
	goblin := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Goblin", TypeLine: "Creature — Goblin", Power: 2, Toughness: 2})
	elf := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Elf", TypeLine: "Creature — Elf", Power: 2, Toughness: 2})
	effectstest.Pr7Activate(t, g, me.ID, circle, 0, game.ActivateAbilityParams{})
	if len(g.PendingChoices) != 0 {
		t.Fatal("Circle of Solace chooses no source")
	}
	life := me.Life
	effectstest.Pr7Hit(t, g, elf, me.ID, 1)
	effectstest.Pr7Hit(t, g, goblin, me.ID, 2)
	if me.Life != life-1 {
		t.Errorf("life %d → %d, want the Elf's 1 dealt and the Goblin's 2 prevented", life, me.Life)
	}
}

// Mercenaries: any player may activate it, and "you" is the activator.
func TestPR7MercenariesShieldsTheActivator(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	merc := effectstest.PushCatalogPermanent(g, me.ID, "Mercenaries", "Creature — Human Mercenary", effectstest.Pr7Mercenaries, false)
	effectstest.Pr7Activate(t, g, opp.ID, merc, 0, game.ActivateAbilityParams{})
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pilgrim := effectstest.PushCatalogPermanent(g, me.ID, "Pilgrim of Justice", "Creature — Human Cleric", effectstest.Pr7PilgrimOfJustice, false)
	red := effectstest.Pr7Creature(g, opp.ID, "Red", 3, "R")
	bear := effectstest.Pr7Creature(g, opp.ID, "Bear", 2, "G")
	effectstest.Pr7Activate(t, g, me.ID, pilgrim, 0, game.ActivateAbilityParams{})
	effectstest.Pr7Choose(t, g, me.ID, red)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(red, bear, 3) })
	if got := effectstest.DamageMarkedOn(g, bear); got != 0 {
		t.Errorf("the opponent's own bear took %d, want 0: the shield protects anything", got)
	}
	if effectstest.FindBattlefieldCardForTest(g, pilgrim) != nil {
		t.Error("the Pilgrim is sacrificed to pay")
	}
}

// Martyr's Cause: a target, and a creature sacrificed to pay.
func TestPR7MartyrsCauseShieldsItsTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cause := effectstest.PushCatalogPermanent(g, me.ID, "Martyr's Cause", "Enchantment", effectstest.Pr7MartyrsCause, false)
	fodder := effectstest.Pr7Creature(g, me.ID, "Fodder", 1, "W")
	knight := effectstest.Pr7Creature(g, me.ID, "Knight", 2, "W")
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 3, "R")
	effectstest.Pr7Activate(t, g, me.ID, cause, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder},
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: knight}},
	})
	effectstest.Pr7Choose(t, g, me.ID, src)
	life := me.Life
	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(src, knight, 3)
		_ = g.DealDamageToPlayerForEffect(src, me.ID, 3)
	})
	if effectstest.DamageMarkedOn(g, knight) != 0 || me.Life != life-3 {
		t.Errorf("knight %d damage (want 0), life %d → %d (want -3: only the target is shielded)",
			effectstest.DamageMarkedOn(g, knight), life, me.Life)
	}
}

// Kithkin Armor: the Aura is gone by resolution, and its creature is
// still the one shielded.
func TestPR7KithkinArmorShieldsTheCreatureItEnchanted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	knight := effectstest.Pr7Creature(g, me.ID, "Knight", 2, "W")
	armor := effectstest.PushCatalogPermanent(g, me.ID, "Kithkin Armor", "Enchantment — Aura", effectstest.Pr7KithkinArmor, false)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, armor).AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: knight}
	})
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 3, "R")
	effectstest.Pr7Activate(t, g, me.ID, armor, 0, game.ActivateAbilityParams{})
	effectstest.Pr7Choose(t, g, me.ID, src)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(src, knight, 3) })
	if got := effectstest.DamageMarkedOn(g, knight); got != 0 {
		t.Errorf("the enchanted creature took %d, want 0", got)
	}
}

// Samite Blessing gives the enchanted creature the row.
func TestPR7SamiteBlessingGrantsTheRow(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	knight := effectstest.Pr7Creature(g, me.ID, "Knight", 2, "W")
	blessing := effectstest.PushCatalogPermanent(g, me.ID, "Samite Blessing", "Enchantment — Aura", effectstest.Pr7SamiteBlessing, false)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, blessing).AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: knight}
	})
	live := effectstest.ApaLive(g, knight)
	if n := len(game.ActivatedAbilitiesForCard(*live)); n != 1 {
		t.Errorf("the enchanted creature has %d activated rows, want the granted one", n)
	}
}

// Deflecting Palm: the damage prevented is dealt to the source's
// controller, by Deflecting Palm.
func TestPR7DeflectingPalmSendsTheDamageBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 4, "G")
	effectstest.CastCatalogSpell(t, g, "Deflecting Palm", "Instant", effectstest.Pr7DeflectingPalm, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Pr7Choose(t, g, me.ID, src)
	mine, theirs := me.Life, opp.Life
	effectstest.Pr7Hit(t, g, src, me.ID, 4)
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
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		src := effectstest.Pr7Creature(g, opp.ID, "Src", 3, tc.color)
		effectstest.CastCatalogSpell(t, g, "Honorable Passage", "Instant", effectstest.Pr7HonorablePassage,
			[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.Pr7Choose(t, g, me.ID, src)
		mine, theirs := me.Life, opp.Life
		effectstest.Pr7Hit(t, g, src, me.ID, 3)
		if me.Life != mine || opp.Life != theirs-tc.back {
			t.Errorf("%s source: me %d → %d (want unchanged), them %d → %d (want -%d)",
				tc.color, mine, me.Life, theirs, opp.Life, tc.back)
		}
	}
}

// Shadowbane covers your creatures, and gains life only for a black
// source.
func TestPR7ShadowbaneCoversYourCreaturesAndGainsForBlack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.Pr7Creature(g, me.ID, "Mine", 2, "W")
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 3, "B")
	effectstest.CastCatalogSpell(t, g, "Shadowbane", "Instant", effectstest.Pr7Shadowbane, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Pr7Choose(t, g, me.ID, src)
	life := me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(src, mine, 3) })
	g.RunStateChecksForTest()
	if effectstest.DamageMarkedOn(g, mine) != 0 || me.Life != life+3 {
		t.Errorf("creature %d damage (want 0), life %d → %d (want +3)", effectstest.DamageMarkedOn(g, mine), life, me.Life)
	}
}

// Awe Strike: the target creature's next damage is prevented, and you
// gain that much.
func TestPR7AweStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 5, "R")
	effectstest.CastCatalogSpell(t, g, "Awe Strike", "Instant", effectstest.Pr7AweStrike, []game.TargetRef{{Kind: game.TargetCard, ID: src}})
	effectstest.PassPriorityAroundTable(t, g)
	life := me.Life
	effectstest.Pr7Hit(t, g, src, me.ID, 5)
	if me.Life != life+5 {
		t.Errorf("life %d → %d, want the 5 prevented and 5 gained", life, me.Life)
	}
}

// Dazzling Reflection: life equal to the power now, and the creature's
// next damage prevented.
func TestPR7DazzlingReflection(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 4, "R")
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Dazzling Reflection", "Instant", effectstest.Pr7DazzlingReflect, []game.TargetRef{{Kind: game.TargetCard, ID: src}})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Fatalf("life %d → %d, want +4 for the power", life, me.Life)
	}
	effectstest.Pr7Hit(t, g, src, me.ID, 4)
	if me.Life != life+4 {
		t.Errorf("life %d, want the creature's next damage prevented", me.Life)
	}
}

// Reverse Damage gains the amount prevented; Invulnerability returns
// with buyback.
func TestPR7ReverseDamageGainsWhatItPrevented(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 6, "R")
	effectstest.CastCatalogSpell(t, g, "Reverse Damage", "Instant", effectstest.Pr7ReverseDamage, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Pr7Choose(t, g, me.ID, src)
	life := me.Life
	effectstest.Pr7Hit(t, g, src, me.ID, 6)
	if me.Life != life+6 {
		t.Errorf("life %d → %d, want +6", life, me.Life)
	}
}

// Bone Mask exiles the top of its controller's library, as many cards as
// it prevented.
func TestPR7BoneMaskExilesWhatItPrevented(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mask := effectstest.PushCatalogPermanent(g, me.ID, "Bone Mask", "Artifact", effectstest.Pr7BoneMask, false)
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 3, "R")
	effectstest.Pr7Activate(t, g, me.ID, mask, 0, game.ActivateAbilityParams{})
	effectstest.Pr7Choose(t, g, me.ID, src)
	lib, exiled := len(me.Library.Cards), len(g.Exile.Cards)
	effectstest.Pr7Hit(t, g, src, me.ID, 3)
	if len(me.Library.Cards) != lib-3 || len(g.Exile.Cards) != exiled+3 {
		t.Errorf("library %d → %d, exile %d → %d: want three cards moved", lib, len(me.Library.Cards), exiled, len(g.Exile.Cards))
	}
}

// New Way Forward: a reflexive trigger deals the damage back and draws.
func TestPR7NewWayForwardStrikesBackAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 2, "R")
	effectstest.CastCatalogSpell(t, g, "New Way Forward", "Instant", effectstest.Pr7NewWayForward, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Pr7Choose(t, g, me.ID, src)
	hand, theirs := len(me.Hand.Cards), opp.Life
	effectstest.Pr7Hit(t, g, src, me.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirs-2 || len(me.Hand.Cards) != hand+2 {
		t.Errorf("them %d → %d (want -2), hand %d → %d (want +2)", theirs, opp.Life, hand, len(me.Hand.Cards))
	}
}

// Intervention Pact schedules its upkeep payment, and shields.
func TestPR7InterventionPactSchedulesThePayment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 2, "R")
	effectstest.CastCatalogSpell(t, g, "Intervention Pact", "Instant", effectstest.Pr7InterventionPact, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Pr7Choose(t, g, me.ID, src)
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
	effectstest.Pr7Hit(t, g, src, me.ID, 2)
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2", life, me.Life)
	}
}

// After damage that can't be prevented, a "prevented this way" card does
// nothing and its shield is still there (CR 615.12, 609.7b).
func TestPR7FollowUpUnderUnpreventableDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 4, "G")
	effectstest.CastCatalogSpell(t, g, "Deflecting Palm", "Instant", effectstest.Pr7DeflectingPalm, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Pr7Choose(t, g, me.ID, src)
	mine, theirs := me.Life, opp.Life
	g.WithWriteLock(func() {
		_ = g.DealMarkedDamageForEffect(src, nil, me.ID, 4, game.DamageMarks{CantBePrevented: true})
	})
	g.RunStateChecksForTest()
	if me.Life != mine-4 || opp.Life != theirs {
		t.Fatalf("me %d → %d (want -4), them %d → %d (want unchanged)", mine, me.Life, theirs, opp.Life)
	}
	effectstest.Pr7Hit(t, g, src, me.ID, 4)
	if me.Life != mine-4 || opp.Life != theirs-4 {
		t.Errorf("the unspent shield should take the next damage and send it back: me %d, them %d", me.Life, opp.Life)
	}
}
