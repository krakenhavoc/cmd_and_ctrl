package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// adr0108_pr9_cards_test.go — ADR 0108 PR 9 (#1905): the cards the
// redirection kind (ModRedirectDamage) and its primitive unblock, each
// against the damage it must move and the damage it must leave alone.

const (
	pr9Beacon            = "a1ed2274-7774-4c65-a95b-28ae5e225994"
	pr9HarmsWay          = "4c34a882-3786-4d3c-9ca4-04fa85d5e51b"
	pr9ReflectDamage     = "21565b0d-f814-49ca-8613-f40040c4ba6c"
	pr9EyeForAnEye       = "22647b1a-5a7c-41b5-b820-b2e9f49c7aad"
	pr9KorChant          = "5b3f6817-5d7a-4d83-ad1d-df75b4e1970b"
	pr9Pariah            = "2c5c8250-1860-42a1-a335-071f54830d37"
	pr9PariahsShield     = "f2103ab8-a183-4db8-98dd-4146217b5125"
	pr9PalisadeGiant     = "f38fe1e9-8997-4b63-9109-7513034bac88"
	pr9TreacherousLink   = "b7902113-9ddf-4d81-a76a-b54b8a36c154"
	pr9NomadsEnKor       = "b8d395a3-0bfe-45c3-bb3a-820d4f235b88"
	pr9RefractionTrap    = "0fc0554d-da45-49d2-941b-3f6c6458e823"
	pr9VeteranBodyguard  = "d29078c0-1fb8-437a-81d1-bb319f646941"
	pr9HarshJudgment     = "1fa272bf-8759-4750-b40b-8e2f6972d570"
	pr9JadeMonolith      = "1e105ab7-fb10-4cfd-ac2f-5e11488cf1b0"
	pr9MirrorwoodTreefol = "fec51ab9-484f-46a8-b2b0-772a61d89e41"
)

// Beacon of Destiny: the next instance from the chosen source is dealt to
// the Beacon; the one after is dealt to you.
func TestADR0108PR9BeaconOfDestinyTakesTheNextInstance(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beacon := apaPush(g, me.ID, me.ID, apaCreature("Beacon of Destiny", pr9Beacon, 1, 9))
	src := pr7Creature(g, opp.ID, "Pinger", 1, "R")
	pr7Activate(t, g, me.ID, beacon, 0, game.ActivateAbilityParams{})
	pr7Choose(t, g, me.ID, src)
	life := me.Life
	pr7Hit(t, g, src, me.ID, 3)
	if me.Life != life || pr6Marked(g, beacon) != 3 {
		t.Fatalf("life %d (want %d), Beacon damage %d (want 3)", me.Life, life, pr6Marked(g, beacon))
	}
	pr7Hit(t, g, src, me.ID, 2)
	if me.Life != life-2 {
		t.Fatalf("life %d, want %d: the next instance is dealt to you", me.Life, life-2)
	}
}

// Harm's Way: 2 of the chosen source's 5 damage to you is dealt to the
// target instead; the rest is dealt to you.
func TestADR0108PR9HarmsWayRedirectsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pr7Creature(g, opp.ID, "Dragon", 5, "R")
	castCatalogSpell(t, g, "Harm's Way", "Instant", pr9HarmsWay, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	pr7Choose(t, g, me.ID, src)
	mine, theirs := me.Life, opp.Life
	pr7Hit(t, g, src, me.ID, 5)
	if me.Life != mine-3 || opp.Life != theirs-2 {
		t.Fatalf("life me %d (want %d), opponent %d (want %d)", me.Life, mine-3, opp.Life, theirs-2)
	}
	pr7Hit(t, g, src, me.ID, 1)
	if me.Life != mine-4 {
		t.Fatalf("life %d, want %d: the charge is spent", me.Life, mine-4)
	}
}

// Reflect Damage: the chosen source's next damage, to anything, is dealt
// to its controller.
func TestADR0108PR9ReflectDamageSendsItHome(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pr7Creature(g, opp.ID, "Dragon", 5, "R")
	castCatalogSpell(t, g, "Reflect Damage", "Instant", pr9ReflectDamage, nil)
	passPriorityAroundTable(t, g)
	pr7Choose(t, g, me.ID, src)
	mine, theirs := me.Life, opp.Life
	pr7Hit(t, g, src, me.ID, 4)
	if me.Life != mine || opp.Life != theirs-4 {
		t.Fatalf("life me %d (want %d), opponent %d (want %d)", me.Life, mine, opp.Life, theirs-4)
	}
}

// Eye for an Eye: the damage to you is still dealt, and Eye for an Eye
// deals that much to the source's controller.
func TestADR0108PR9EyeForAnEye(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pr7Creature(g, opp.ID, "Dragon", 5, "R")
	castCatalogSpell(t, g, "Eye for an Eye", "Instant", pr9EyeForAnEye, nil)
	passPriorityAroundTable(t, g)
	pr7Choose(t, g, me.ID, src)
	mine, theirs := me.Life, opp.Life
	pr7Hit(t, g, src, me.ID, 3)
	if me.Life != mine-3 || opp.Life != theirs-3 {
		t.Fatalf("life me %d (want %d), opponent %d (want %d)", me.Life, mine-3, opp.Life, theirs-3)
	}
}

// Kor Chant: the chosen source's damage to the first target is dealt to
// the second, all turn, still as the source's damage.
func TestADR0108PR9KorChantMovesTheSourcesDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pr7Creature(g, me.ID, "Mine", 2, "W")
	theirs := pr7Creature(g, opp.ID, "Theirs", 2, "G")
	src := pr7Creature(g, opp.ID, "Pinger", 1, "R")
	other := pr7Creature(g, opp.ID, "Other", 1, "R")
	castCatalogSpell(t, g, "Kor Chant", "Instant", pr9KorChant, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0},
		{Kind: game.TargetCard, ID: theirs, Slot: 1},
	})
	passPriorityAroundTable(t, g)
	pr7Choose(t, g, me.ID, src)
	pr6Damage(t, g, src, mine, 1)
	pr6Damage(t, g, src, mine, 1)
	pr6Damage(t, g, other, mine, 1)
	if pr6Marked(g, mine) != 1 || pr6Marked(g, theirs) != 2 {
		t.Fatalf("mine %d (want 1: the other source's), theirs %d (want 2)", pr6Marked(g, mine), pr6Marked(g, theirs))
	}
}

// Pariah: damage to you is dealt to the enchanted creature. Pariah's
// Shield unattached does nothing (its ruling).
func TestADR0108PR9PariahAndAnUnattachedShield(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := pr7Creature(g, me.ID, "Host", 1, "W")
	pushAuraOnLand(t, g, me.ID, "Pariah", pr9Pariah, host)
	pushCatalogPermanent(g, me.ID, "Pariah's Shield", "Artifact — Equipment", pr9PariahsShield, false)
	src := pr7Creature(g, opp.ID, "Pinger", 1, "R")
	life := me.Life
	pr7Hit(t, g, src, me.ID, 3)
	if me.Life != life || pr6Marked(g, host) != 3 {
		t.Fatalf("life %d (want %d), host damage %d (want 3)", me.Life, life, pr6Marked(g, host))
	}
}

// Two Pariahs on two creatures are two different redirections: the
// affected player orders them (CR 616.1, the Pariah ruling), and the
// damage goes to the one chosen first, never split.
func TestADR0108PR9TwoPariahsAreOrdered(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pr7Creature(g, me.ID, "A", 1, "W")
	b := pr7Creature(g, me.ID, "B", 1, "W")
	pushAuraOnLand(t, g, me.ID, "Pariah", pr9Pariah, a)
	pushAuraOnLand(t, g, me.ID, "Pariah", pr9Pariah, b)
	src := pr7Creature(g, opp.ID, "Pinger", 1, "R")
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(src, me.ID, 3); err != nil {
			t.Fatal(err)
		}
	})
	var order *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceReplacementOrder {
			order = c
		}
	}
	if order == nil || order.Chooser != me.ID || len(order.ReplacementEffectIDs) != 2 {
		t.Fatalf("want a CR 616 ordering prompt for me between the two Pariahs, got %+v", order)
	}
	ids := order.ReplacementEffectIDs
	if err := g.ResolveReplacementOrder(order.ID, me.ID, []game.ReplacementEffectID{ids[1], ids[0]}); err != nil {
		t.Fatal(err)
	}
	if pr6Marked(g, a)+pr6Marked(g, b) != 3 || (pr6Marked(g, a) != 0 && pr6Marked(g, b) != 0) {
		t.Fatalf("a %d, b %d: want all 3 on one creature", pr6Marked(g, a), pr6Marked(g, b))
	}
}

// Palisade Giant: damage to you and your other permanents is dealt to the
// Giant; damage to the Giant stays on it.
func TestADR0108PR9PalisadeGiant(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	giant := apaPush(g, me.ID, me.ID, apaCreature("Palisade Giant", pr9PalisadeGiant, 2, 9))
	bear := pr7Creature(g, me.ID, "Bear", 2, "G")
	src := pr7Creature(g, opp.ID, "Pinger", 1, "R")
	life := me.Life
	pr7Hit(t, g, src, me.ID, 2)
	pr6Damage(t, g, src, bear, 1)
	pr6Damage(t, g, src, giant, 1)
	if me.Life != life || pr6Marked(g, bear) != 0 || pr6Marked(g, giant) != 4 {
		t.Fatalf("life %d (want %d), bear %d (want 0), giant %d (want 4)", me.Life, life, pr6Marked(g, bear), pr6Marked(g, giant))
	}
}

// Treacherous Link: damage to the enchanted creature is dealt to its
// controller instead.
func TestADR0108PR9TreacherousLink(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := pr7Creature(g, opp.ID, "Theirs", 2, "G")
	pushAuraOnLand(t, g, me.ID, "Treacherous Link", pr9TreacherousLink, theirs)
	src := pr7Creature(g, me.ID, "Pinger", 1, "R")
	life := opp.Life
	pr6Damage(t, g, src, theirs, 3)
	if pr6Marked(g, theirs) != 0 || opp.Life != life-3 {
		t.Fatalf("creature %d (want 0), controller life %d (want %d)", pr6Marked(g, theirs), opp.Life, life-3)
	}
}

// The en-Kor row: 1 of the damage to the en-Kor is dealt to the target;
// the rest stays where it was.
func TestADR0108PR9EnKorRedirectsOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nomads := apaPush(g, me.ID, me.ID, apaCreature("Nomads en-Kor", pr9NomadsEnKor, 1, 9))
	other := pr7Creature(g, me.ID, "Other", 1, "W")
	src := pr7Creature(g, opp.ID, "Pinger", 1, "R")
	pr7Activate(t, g, me.ID, nomads, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: other}}})
	pr6Damage(t, g, src, nomads, 3)
	if pr6Marked(g, nomads) != 2 || pr6Marked(g, other) != 1 {
		t.Fatalf("nomads %d (want 2), other %d (want 1)", pr6Marked(g, nomads), pr6Marked(g, other))
	}
}

// Jade Monolith: the chosen source's next damage to the target creature
// is dealt to the activator.
func TestADR0108PR9JadeMonolith(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	monolith := pushCatalogPermanent(g, me.ID, "Jade Monolith", "Artifact", pr9JadeMonolith, false)
	bear := pr7Creature(g, me.ID, "Bear", 2, "G")
	src := pr7Creature(g, opp.ID, "Pinger", 1, "R")
	pr7Activate(t, g, me.ID, monolith, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	pr7Choose(t, g, me.ID, src)
	life := me.Life
	pr6Damage(t, g, src, bear, 2)
	if pr6Marked(g, bear) != 0 || me.Life != life-2 {
		t.Fatalf("bear %d (want 0), life %d (want %d)", pr6Marked(g, bear), me.Life, life-2)
	}
}

// Mirrorwood Treefolk: its next damage is dealt to the target, and a
// target that has gone leaves the damage on the Treefolk (its ruling).
func TestADR0108PR9MirrorwoodTreefolk(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tree := apaPush(g, me.ID, me.ID, apaCreature("Mirrorwood Treefolk", pr9MirrorwoodTreefol, 2, 9))
	src := pr7Creature(g, opp.ID, "Pinger", 1, "R")
	life := opp.Life
	pr7Activate(t, g, me.ID, tree, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	pr6Damage(t, g, src, tree, 3)
	if pr6Marked(g, tree) != 0 || opp.Life != life-3 {
		t.Fatalf("treefolk %d (want 0), opponent life %d (want %d)", pr6Marked(g, tree), opp.Life, life-3)
	}
}

// Veteran Bodyguard: an unblocked attacker's damage to its controller is
// dealt to the untapped Bodyguard; tapped, it is not.
func TestADR0108PR9VeteranBodyguard(t *testing.T) {
	for _, tapped := range []bool{false, true} {
		g := newCatalogGame(t)
		atk, def := g.Seats[0], g.Seats[1]
		guard := apaPush(g, def.ID, def.ID, apaCreature("Veteran Bodyguard", pr9VeteranBodyguard, 2, 9))
		if tapped {
			findBattlefieldCardForTest(g, guard).Tapped = true
		}
		attacker := pr7Creature(g, atk.ID, "Raider", 3, "R")
		life := def.Life
		advanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(attacker, def.ID); err != nil {
			t.Fatal(err)
		}
		advanceTo(t, g, game.StepCombatDamage)
		switch {
		case !tapped && (def.Life != life || pr6Marked(g, guard) != 3):
			t.Fatalf("untapped: life %d (want %d), guard %d (want 3)", def.Life, life, pr6Marked(g, guard))
		case tapped && (def.Life != life-3 || pr6Marked(g, guard) != 0):
			t.Fatalf("tapped: life %d (want %d), guard %d (want 0)", def.Life, life-3, pr6Marked(g, guard))
		}
	}
}

// Refraction Trap: the trap cost is offered only once an opponent has cast
// a red instant or sorcery this turn; the shield prevents 3 of the chosen
// source's damage and Refraction Trap deals that much to its target.
func TestADR0108PR9RefractionTrap(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cond := anOpponentCastAnInstantOrSorceryOfColor("R")
	if cond(g, me.ID) {
		t.Fatal("the trap is offered before anyone cast anything")
	}
	g.SpellsCastThisTurn = map[uuid.UUID]game.CastTally{opp.ID: {Total: 1, Noncreature: 1, InstantSorceryColors: "R"}}
	if !cond(g, me.ID) || cond(g, opp.ID) {
		t.Fatal("the trap reads the opponents' red instants and sorceries only")
	}
	src := pr7Creature(g, opp.ID, "Dragon", 5, "R")
	castCatalogSpell(t, g, "Refraction Trap", "Instant — Trap", pr9RefractionTrap, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	pr7Choose(t, g, me.ID, src)
	mine, theirs := me.Life, opp.Life
	pr7Hit(t, g, src, me.ID, 5)
	if me.Life != mine-2 || opp.Life != theirs-3 {
		t.Fatalf("life me %d (want %d), opponent %d (want %d)", me.Life, mine-2, opp.Life, theirs-3)
	}
}

// Harsh Judgment: a spell of the chosen colour that would deal damage to
// you deals it to its controller instead; a creature of that colour does
// not.
func TestADR0108PR9HarshJudgment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	judgment := pushCatalogPermanent(g, me.ID, "Harsh Judgment", "Enchantment", pr9HarshJudgment, false)
	findBattlefieldCardForTest(g, judgment).ChosenColor = "R"
	creature := pr7Creature(g, opp.ID, "Red Creature", 2, "R")
	bolt := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: bolt, Name: "Bolt", TypeLine: "Instant", Colors: []string{"R"}, Owner: opp.ID, Controller: opp.ID})
	mine, theirs := me.Life, opp.Life
	pr7Hit(t, g, creature, me.ID, 2)
	pr7Hit(t, g, bolt, me.ID, 3)
	if me.Life != mine-2 || opp.Life != theirs-3 {
		t.Fatalf("life me %d (want %d), opponent %d (want %d)", me.Life, mine-2, opp.Life, theirs-3)
	}
}

// Flanking (CR 702.25a): each blocker without flanking gets -1/-1 until
// end of turn; one with flanking does not.
func TestADR0108PR9ZhalfirinCrusaderFlanking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crusader := apaPush(g, me.ID, me.ID, apaCreature("Zhalfirin Crusader", "9e60c102-412e-4956-a7bb-a2cd737d6692", 2, 2))
	plain := apaPush(g, opp.ID, opp.ID, apaCreature("Plain", "", 2, 2))
	flanker := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Flanker", TypeLine: "Creature — Test", Power: 2, Toughness: 2, Keywords: []string{"flanking"}})
	g.WithWriteLock(func() {
		for i, b := range []uuid.UUID{plain, flanker} {
			g.EmitEvent(game.Event{Kind: game.EventBlock, Actor: opp.ID, CardID: b, Target: crusader, Amount: i + 1})
		}
	})
	passPriorityAroundTable(t, g)
	if got := apaLive(g, plain).CurrentToughness(); got != 1 {
		t.Fatalf("the blocker without flanking has toughness %d, want 1", got)
	}
	if got := apaLive(g, flanker).CurrentToughness(); got != 2 {
		t.Fatalf("the blocker with flanking has toughness %d, want 2", got)
	}
}

// Bushido 1 (CR 702.45a): "becomes blocked" once, and "blocks" once
// however many attackers it blocks.
func TestADR0108PR9OpalEyeBushido(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	opal := apaPush(g, me.ID, me.ID, apaCreature("Opal-Eye, Konda's Yojimbo", "e4acac65-112b-48fc-bea3-44747c1389a3", 1, 4))
	a := pr7Creature(g, opp.ID, "A", 2, "R")
	b := pr7Creature(g, opp.ID, "B", 2, "R")
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBlock, Actor: me.ID, CardID: opal, Target: a, Amount: 1})
		g.EmitEvent(game.Event{Kind: game.EventBlock, Actor: me.ID, CardID: opal, Target: b, Amount: 2})
	})
	passPriorityAroundTable(t, g)
	if got := apaLive(g, opal).CurrentPower(); got != 2 {
		t.Fatalf("power %d after blocking two attackers, want 2: blocks triggers once", got)
	}
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBecomesBlocked, Actor: opp.ID, Source: opal, CardID: opal, Target: opal})
	})
	passPriorityAroundTable(t, g)
	if got := apaLive(g, opal).CurrentPower(); got != 3 {
		t.Fatalf("power %d after becoming blocked, want 3", got)
	}
}
