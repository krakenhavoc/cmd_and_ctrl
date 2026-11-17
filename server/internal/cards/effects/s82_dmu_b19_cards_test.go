package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	dmuB19Ayesha   = "59fcd09c-4633-457e-b36a-488e94c1c941"
	dmuB19Marhault = "f49b31ce-910e-446a-b7a7-4274d71e961d"
	dmuB19Jasmine  = "138201c0-556d-4c76-b383-90fc9dad0fe0"
	dmuB19Moira    = "48060406-dd5a-4df5-84e9-3e797f0248cd"
	dmuB19Ohabi    = "e403fadb-537c-41ac-924a-450f0eabe2c4"
	dmuB19Ramses   = "170182ed-af1d-442a-8868-d74a96e17f03"
	dmuB19Shanid   = "70e03c58-194f-443d-ab64-646dd14d2f53"
	dmuB19Tetsuo   = "1111e042-0bce-4819-ab42-9912aed4d930"
	dmuB19Tor      = "b44d5457-b4d6-4521-a25a-8ef96adba461"
	dmuB19Xira     = "18d0e6e1-3824-465d-9245-21e0174f551f"
)

func TestDmuB19CardsAreFullWithNoCaveats(t *testing.T) {
	want := map[string]string{
		dmuB19Ayesha: "Ayesha Tanaka, Armorer", dmuB19Marhault: "General Marhault Elsdragon",
		dmuB19Jasmine: "Jasmine Boreal of the Seven", dmuB19Moira: "Moira, Urborg Haunt",
		dmuB19Ohabi: "Ohabi Caleria", dmuB19Ramses: "Ramses, Assassin Lord",
		dmuB19Shanid: "Shanid, Sleepers' Scourge", dmuB19Tetsuo: "Tetsuo, Imperial Champion",
		dmuB19Tor: "Tor Wauki the Younger", dmuB19Xira: "Xira, the Golden Sting",
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

func TestDmuB19MoiraReturnsACreatureThatDiedThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	moira := b12Push(g, me.ID, "Moira, Urborg Haunt", "Legendary Creature — Spirit Wizard", dmuB19Moira, 3, 2)
	if !hasKeywordOf(g, moira, "menace") {
		t.Error("Moira has no menace")
	}
	fresh := b12Creature(g, me.ID, "Fresh Corpse", "Creature — Bear", 2, 2)
	killCreature(t, g, me.ID, fresh)
	old := pushGraveyardCardTyped(me, "Old Corpse", "Creature — Bear")
	dealCombatDamageToPlayer(g, moira, opp.ID, 3)
	answerTriggerTargets(t, g, me.ID, fresh)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(fresh) {
		t.Error("the creature that died this turn did not return")
	}
	if g.Battlefield.Contains(old) {
		t.Error("a creature from an earlier turn must stay put")
	}
}

func TestDmuB19MoiraCannotTargetACardThatDidNotDieThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	moira := b12Push(g, me.ID, "Moira, Urborg Haunt", "Legendary Creature — Spirit Wizard", dmuB19Moira, 3, 2)
	old := pushGraveyardCardTyped(me, "Old Corpse", "Creature — Bear")
	dealCombatDamageToPlayer(g, moira, opp.ID, 3)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(old) {
		t.Error("a card that was milled or died earlier must not return")
	}
}

// dmuB19Spend casts a hand card with the pool already filled.
func dmuB19Library(g *game.Game, p *game.Player) int {
	n := 0
	g.ReadSnapshot(func() { n = p.Library.Size() })
	return n
}

func TestDmuB19ShanidGrantsMenaceAndDrawsOnLegendaryCastsAndLandPlays(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	shanid := b12Push(g, me.ID, "Shanid, Sleepers' Scourge", "Legendary Creature — Human Knight", dmuB19Shanid, 2, 4)
	legend := b12Push(g, me.ID, "Legend", "Legendary Creature — Bear", "", 2, 2)
	plain := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	theirs := b12Push(g, g.Seats[1].ID, "Their Legend", "Legendary Creature — Bear", "", 2, 2)
	if !hasKeywordOf(g, shanid, "menace") {
		t.Error("Shanid has no menace")
	}
	if !hasKeywordOf(g, legend, "menace") {
		t.Error("another legendary creature you control has no menace")
	}
	if hasKeywordOf(g, plain, "menace") || hasKeywordOf(g, theirs, "menace") {
		t.Error("a nonlegendary creature, or a legendary one an opponent controls, got menace")
	}

	life, lib := me.Life, dmuB19Library(g, me)
	castCatalogSpell(t, g, "Legendary Spell", "Legendary Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if me.Life != life-1 || dmuB19Library(g, me) != lib-1 {
		t.Errorf("legendary cast: life %d->%d, library %d->%d; want -1 life and one card drawn", life, me.Life, lib, dmuB19Library(g, me))
	}

	life, lib = me.Life, dmuB19Library(g, me)
	b12PlayFromHand(t, g, "Legendary Land", "Legendary Land", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if me.Life != life-1 || dmuB19Library(g, me) != lib-1 {
		t.Errorf("legendary land play: life %d->%d, library %d->%d; want -1 life and one card drawn", life, me.Life, lib, dmuB19Library(g, me))
	}

	life, lib = me.Life, dmuB19Library(g, me)
	b12PlayFromHand(t, g, "Plain Land", "Land", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if me.Life != life || dmuB19Library(g, me) != lib {
		t.Error("a nonlegendary land play must not trigger Shanid")
	}
}

func TestDmuB19ShanidDoesNotTriggerOnALegendaryLandPutOntoTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Shanid, Sleepers' Scourge", "Legendary Creature — Human Knight", dmuB19Shanid, 2, 4)
	life := me.Life
	b12Push(g, me.ID, "Legendary Land", "Legendary Land", "", 0, 0)
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Error("a land put onto the battlefield was not played")
	}
}

func TestDmuB19RamsesAnthemsOtherAssassins(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ramses := b12Push(g, me.ID, "Ramses, Assassin Lord", "Legendary Creature — Human Assassin", dmuB19Ramses, 4, 4)
	mine := b12Creature(g, me.ID, "My Assassin", "Creature — Human Assassin", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, g.Seats[1].ID, "Their Assassin", "Creature — Human Assassin", 2, 2)
	for id, want := range map[uuid.UUID]int{ramses: 4, mine: 3, bear: 2, theirs: 2} {
		if p, _ := mkmMPPT(t, g, id); p != want {
			t.Errorf("power of %s = %d, want %d", id, p, want)
		}
	}
	if !hasKeywordOf(g, ramses, "deathtouch") {
		t.Error("Ramses has no deathtouch")
	}
}

func TestDmuB19RamsesWinsWhenAPlayerAttackedByYourAssassinLoses(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Ramses, Assassin Lord", "Legendary Creature — Human Assassin", dmuB19Ramses, 4, 4)
	assassin := b12Creature(g, me.ID, "My Assassin", "Creature — Human Assassin", 2, 2)
	declareAttack(t, g, opp.ID, assassin)
	if err := g.Concede(opp.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Outcome == nil || g.Outcome.Winner != me.ID {
		t.Fatalf("outcome %+v, want a win for the Assassin's controller", g.Outcome)
	}
}

func TestDmuB19RamsesWinsEvenWhenTheAttackingAssassinHasDied(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Ramses, Assassin Lord", "Legendary Creature — Human Assassin", dmuB19Ramses, 4, 4)
	assassin := b12Creature(g, me.ID, "My Assassin", "Creature — Human Assassin", 2, 2)
	declareAttack(t, g, opp.ID, assassin)
	killCreature(t, g, me.ID, assassin)
	if err := g.Concede(opp.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Outcome == nil || g.Outcome.Winner != me.ID {
		t.Fatalf("outcome %+v, want a win off the attack the dead Assassin made", g.Outcome)
	}
}

func TestDmuB19RamsesDoesNotWinForAPlayerWhoWasNotAttackedByAnAssassin(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	b12Push(g, me.ID, "Ramses, Assassin Lord", "Legendary Creature — Human Assassin", dmuB19Ramses, 4, 4)
	assassin := b12Creature(g, me.ID, "My Assassin", "Creature — Human Assassin", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, bear)
	if err := g.Concede(opp.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Outcome != nil {
		t.Fatalf("a player attacked only by a Bear lost and %+v ended the game", g.Outcome)
	}
	_ = assassin
	_ = other
}

func TestDmuB19RamsesDoesNotWinForAPlayerAnotherPlayersAssassinAttacked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	b12Push(g, third.ID, "Ramses, Assassin Lord", "Legendary Creature — Human Assassin", dmuB19Ramses, 4, 4)
	assassin := b12Creature(g, me.ID, "My Assassin", "Creature — Human Assassin", 2, 2)
	declareAttack(t, g, opp.ID, assassin)
	if err := g.Concede(opp.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Outcome != nil {
		t.Fatalf("Ramses won for a player an Assassin its controller did not control attacked: %+v", g.Outcome)
	}
}

func TestDmuB19MarhaultPumpsABlockedCreatureThreePerBlocker(t *testing.T) {
	for blockers, want := range map[int]int{1: 5, 2: 8} {
		g := newCatalogGame(t)
		b12Push(g, g.Seats[0].ID, "General Marhault Elsdragon", "Legendary Creature — Elf Warrior", dmuB19Marhault, 4, 4)
		bear := b12Creature(g, g.Seats[0].ID, "Bear", "Creature — Bear", 2, 2)
		rampageBlock(t, g, bear, rampageBlockers(g, blockers)...)
		if p, tt := mkmMPPT(t, g, bear); p != want || tt != want {
			t.Errorf("a creature blocked by %d is %d/%d, want %d/%[3]d", blockers, p, tt, want)
		}
	}
}

func TestDmuB19MarhaultPumpsItselfButNotCreaturesItsControllerDoesNotControl(t *testing.T) {
	g := newCatalogGame(t)
	general := b12Push(g, g.Seats[0].ID, "General Marhault Elsdragon", "Legendary Creature — Elf Warrior", dmuB19Marhault, 4, 4)
	rampageBlock(t, g, general, rampageBlockers(g, 2)...)
	if p, _ := mkmMPPT(t, g, general); p != 10 {
		t.Errorf("a blocked General has power %d, want 10", p)
	}

	g = newCatalogGame(t)
	b12Push(g, g.Seats[2].ID, "General Marhault Elsdragon", "Legendary Creature — Elf Warrior", dmuB19Marhault, 4, 4)
	bear := b12Creature(g, g.Seats[0].ID, "Bear", "Creature — Bear", 2, 2)
	rampageBlock(t, g, bear, rampageBlockers(g, 2)...)
	if p, _ := mkmMPPT(t, g, bear); p != 2 {
		t.Errorf("a creature the General's controller does not control got pumped to %d", p)
	}
}

func TestDmuB19JasmineManaPaysOnlyForCreatureSpellsWithNoAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	jasmine := b12Push(g, me.ID, "Jasmine Boreal of the Seven", "Legendary Creature — Human Druid", dmuB19Jasmine, 2, 4)
	advanceToMain(t, g)
	vanilla, flier, rock := uuid.New(), uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		known := map[uuid.UUID]bool{me.ID: true}
		me.Hand.PushTop(game.Card{InstanceID: vanilla, Name: "Vanilla Bear", TypeLine: "Creature — Bear", ManaCost: "{G}{W}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, KnownBy: known, NoAbilitiesText: game.PrintedYes})
		me.Hand.PushTop(game.Card{InstanceID: flier, Name: "Flying Bear", TypeLine: "Creature — Bear", ManaCost: "{G}{W}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, KnownBy: known, NoAbilitiesText: game.PrintedNo, Keywords: []string{"flying"}})
		me.Hand.PushTop(game.Card{InstanceID: rock, Name: "Vanilla Rock", TypeLine: "Artifact", ManaCost: "{G}{W}", Owner: me.ID, Controller: me.ID, KnownBy: known, NoAbilitiesText: game.PrintedYes})
	})
	if err := g.ActivateManaAbility(me.ID, jasmine, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := me.ManaPool.Total(); got != 2 || me.ManaPool.CountColor("G") != 1 || me.ManaPool.CountColor("W") != 1 {
		t.Fatalf("pool holds %d mana, want {G}{W}", got)
	}
	for _, id := range []uuid.UUID{flier, rock} {
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
			t.Fatalf("Jasmine's mana paid for %s", id)
		}
	}
	if me.ManaPool.Total() != 2 {
		t.Fatalf("pool holds %d mana after refused casts, want 2", me.ManaPool.Total())
	}
	if err := g.CastSpell(me.ID, vanilla, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the vanilla creature spell: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("pool still holds %d mana", me.ManaPool.Total())
	}
}

func TestDmuB19JasmineCreaturesWithNoAbilitiesCantBeBlockedByCreaturesWithAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Jasmine Boreal of the Seven", "Legendary Creature — Human Druid", dmuB19Jasmine, 2, 4)
	vanilla := apaPush(g, me.ID, me.ID, game.Card{Name: "Vanilla", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, NoAbilitiesText: game.PrintedYes})
	flier := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Flier", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	plain := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Plain", TypeLine: "Creature — Bear", Power: 1, Toughness: 1, NoAbilitiesText: game.PrintedYes})
	brAttack(t, g, vanilla)
	brRefusal(t, g.DeclareBlocker(flier, vanilla), game.BlockReasonCantBeBlockedBy)
	if err := g.DeclareBlocker(plain, vanilla); err != nil {
		t.Fatalf("a creature with no abilities may block: %v", err)
	}
}

func TestDmuB19JasmineOnlyBindsCreaturesYouControlThatHaveNoAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jasmine := b12Push(g, me.ID, "Jasmine Boreal of the Seven", "Legendary Creature — Human Druid", dmuB19Jasmine, 2, 4)
	armed := apaPush(g, me.ID, me.ID, game.Card{Name: "Armed", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, NoAbilitiesText: game.PrintedYes})
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(armed),
			[]game.Mod{game.AddKeywordsMod("flying")}, g.UntilEndOfTurnDuration(), "gains flying")
	})
	flier := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Flier", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	brAttack(t, g, armed, jasmine)
	if err := g.DeclareBlocker(flier, armed); err != nil {
		t.Fatalf("a creature that has gained an ability can be blocked by a creature with abilities: %v", err)
	}
	if err := g.DeclareBlocker(apaPush(g, opp.ID, opp.ID, game.Card{Name: "Second Flier", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}}), jasmine); err != nil {
		t.Fatalf("Jasmine has an ability herself and can be blocked: %v", err)
	}
}

func TestDmuB19JasmineCreatureThatLostAllAbilitiesCantBeBlockedByCreaturesWithAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Jasmine Boreal of the Seven", "Legendary Creature — Human Druid", dmuB19Jasmine, 2, 4)
	armed := apaPush(g, me.ID, me.ID, game.Card{Name: "Armed", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, NoAbilitiesText: game.PrintedNo, Keywords: []string{"flying"}})
	flier := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Flier", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(armed),
			[]game.Mod{game.LoseAllAbilitiesMod()}, g.UntilEndOfTurnDuration(), "loses all abilities")
	})
	brAttack(t, g, armed)
	brRefusal(t, g.DeclareBlocker(flier, armed), game.BlockReasonCantBeBlockedBy)
}

func TestDmuB19TorWaukiAddsOneToNoncombatDamageFromOtherSources(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tor := b12Push(g, me.ID, "Tor Wauki the Younger", "Legendary Creature — Human Archer", dmuB19Tor, 3, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	their2 := b12Creature(g, opp.ID, "Their Other Bear", "Creature — Bear", 2, 2)
	damage := func(src, victim uuid.UUID, n int, toPlayer bool) {
		g.WithWriteLock(func() {
			var err error
			if toPlayer {
				err = g.DealDamageToPlayerForEffect(src, victim, n)
			} else {
				err = g.DealDamageToCreatureForEffect(src, victim, n)
			}
			if err != nil {
				t.Fatal(err)
			}
		})
	}
	life := opp.Life
	damage(bear, opp.ID, 2, true)
	if opp.Life != life-3 {
		t.Errorf("noncombat damage from another source: life %d -> %d, want 3 lost", life, opp.Life)
	}
	life = opp.Life
	damage(tor, opp.ID, 2, true)
	if opp.Life != life-2 {
		t.Errorf("Tor Wauki's own damage: life %d -> %d, want 2 lost", life, opp.Life)
	}
	damage(bear, theirs, 1, false)
	if c := b12Card(t, g, theirs); c.DamageMarked != 2 {
		t.Errorf("damage to a creature: %d marked, want 1+1", c.DamageMarked)
	}
	life = opp.Life
	damage(theirs, me.ID, 2, true)
	if me.Life != 40-2 && opp.Life != life {
		t.Errorf("an opponent's source must not be increased: my life %d", me.Life)
	}
	_ = their2
}

func TestDmuB19TorWaukiDoesNotIncreaseCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Tor Wauki the Younger", "Legendary Creature — Human Archer", dmuB19Tor, 3, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	life := opp.Life
	attackWith(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("combat damage: life %d -> %d, want exactly 2", life, opp.Life)
	}
}

func TestDmuB19TorWaukiShootsForTwoWhenYouCastAnInstantOrSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Tor Wauki the Younger", "Legendary Creature — Human Archer", dmuB19Tor, 3, 3)
	life, oppLife := me.Life, opp.Life
	castCatalogSpell(t, g, "Spell", "Instant", "", nil)
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 {
		t.Errorf("opponent life %d -> %d, want 2 lost", oppLife, opp.Life)
	}
	if me.Life != life+2 {
		t.Errorf("lifelink: my life %d -> %d, want +2", life, me.Life)
	}
	oppLife = opp.Life
	castCatalogSpell(t, g, "Bear Spell", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife {
		t.Error("a creature spell triggered Tor Wauki")
	}
}

func TestDmuB19XiraMarksACreatureAndPaysWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	xira := b12Push(g, me.ID, "Xira, the Golden Sting", "Legendary Creature — Insect Assassin", dmuB19Xira, 3, 3)
	if !hasKeywordOf(g, xira, "flying") || !hasKeywordOf(g, xira, "haste") {
		t.Error("Xira lacks flying or haste")
	}
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, xira)
	answerPickTarget(t, g, victim)
	passPriorityAroundTable(t, g)
	if c := b12Card(t, g, victim); c.Counters["egg"] != 1 {
		t.Fatalf("egg counters on the target: %d, want 1", c.Counters["egg"])
	}
	hand := me.Hand.Size()
	killCreature(t, g, opp.ID, victim)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, me.Hand.Size())
	}
	if countOnBattlefield(g, "Insect", me.ID) != 1 {
		t.Errorf("Insect tokens = %d, want 1", countOnBattlefield(g, "Insect", me.ID))
	}
}

func TestDmuB19XiraPaysAfterXiraHasLeftAndOnlyOnce(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	xira := b12Push(g, me.ID, "Xira, the Golden Sting", "Legendary Creature — Insect Assassin", dmuB19Xira, 3, 3)
	victim := b12Creature(g, me.ID, "Victim", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, xira)
	answerPickTarget(t, g, victim)
	passPriorityAroundTable(t, g)
	killCreature(t, g, me.ID, xira)
	hand := me.Hand.Size()
	killCreature(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || countOnBattlefield(g, "Insect", me.ID) != 1 {
		t.Errorf("after Xira left: hand %d -> %d, Insects %d; want one card and one token", hand, me.Hand.Size(), countOnBattlefield(g, "Insect", me.ID))
	}
}

func TestDmuB19XiraPaysNothingWhenTheEggCounterIsGone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	xira := b12Push(g, me.ID, "Xira, the Golden Sting", "Legendary Creature — Insect Assassin", dmuB19Xira, 3, 3)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, xira)
	answerPickTarget(t, g, victim)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(victim, "egg", -1); err != nil {
			t.Fatal(err)
		}
	})
	hand := me.Hand.Size()
	killCreature(t, g, opp.ID, victim)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || countOnBattlefield(g, "Insect", me.ID) != 0 {
		t.Error("a creature that died without its egg counter paid out")
	}
}

func TestDmuB19XiraCantTargetACreatureThatAlreadyHasAnEggCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	xira := b12Push(g, me.ID, "Xira, the Golden Sting", "Legendary Creature — Insect Assassin", dmuB19Xira, 3, 3)
	marked := b12Creature(g, opp.ID, "Marked", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(marked, "egg", 1); err != nil {
			t.Fatal(err)
		}
	})
	declareAttack(t, g, opp.ID, xira)
	passPriorityAroundTable(t, g)
	if c := b12Card(t, g, marked); c.Counters["egg"] != 1 {
		t.Errorf("egg counters = %d, want it left at 1", c.Counters["egg"])
	}
	if triggerOnStack(g, xira) != nil {
		t.Error("Xira's trigger went on the stack with no legal target")
	}
}

func TestDmuB19TetsuoDoesNothingUnequipped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tetsuo := b12Push(g, me.ID, "Tetsuo, Imperial Champion", "Legendary Creature — Human Samurai", dmuB19Tetsuo, 3, 3)
	declareAttack(t, g, opp.ID, tetsuo)
	if triggerOnStack(g, tetsuo) != nil || modePickChoiceFor(g, me.ID) != nil {
		t.Error("Tetsuo triggered while unequipped")
	}
}

func dmuB19Equip(g *game.Game, owner uuid.UUID, host uuid.UUID, mv int) uuid.UUID {
	cost := map[int]string{0: "", 1: "{1}", 2: "{2}", 3: "{3}", 4: "{4}"}[mv]
	id := apaPush(g, owner, owner, game.Card{Name: "Test Blade", TypeLine: "Artifact — Equipment", ManaCost: cost})
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(id, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			panic(err)
		}
	})
	return id
}

func TestDmuB19TetsuoDealsDamageEqualToTheGreatestEquipmentValue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tetsuo := b12Push(g, me.ID, "Tetsuo, Imperial Champion", "Legendary Creature — Human Samurai", dmuB19Tetsuo, 3, 3)
	dmuB19Equip(g, me.ID, tetsuo, 1)
	dmuB19Equip(g, me.ID, tetsuo, 3)
	declareAttack(t, g, opp.ID, tetsuo)
	cmdr2S43ChooseMode(t, g, me.ID, 0)
	answerPickTargetPlayer(t, g, opp.ID)
	life := opp.Life
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d -> %d, want 3 lost (the greatest value, not the sum)", life, opp.Life)
	}
}

func TestDmuB19TetsuoDoesNothingIfTheEquipmentIsGoneAsItResolves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tetsuo := b12Push(g, me.ID, "Tetsuo, Imperial Champion", "Legendary Creature — Human Samurai", dmuB19Tetsuo, 3, 3)
	blade := dmuB19Equip(g, me.ID, tetsuo, 3)
	declareAttack(t, g, opp.ID, tetsuo)
	cmdr2S43ChooseMode(t, g, me.ID, 0)
	answerPickTargetPlayer(t, g, opp.ID)
	exileCreature(t, g, blade)
	life := opp.Life
	passPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("opponent life %d -> %d, want no damage once the Equipment has gone", life, opp.Life)
	}
}

func TestDmuB19TetsuoOffersAFreeInstantOrSorceryUpToTheEquipmentValue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tetsuo := b12Push(g, me.ID, "Tetsuo, Imperial Champion", "Legendary Creature — Human Samurai", dmuB19Tetsuo, 3, 3)
	dmuB19Equip(g, me.ID, tetsuo, 2)
	dmuB19Equip(g, me.ID, tetsuo, 1)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	cheap := handSpell(me, "Cheap Instant", "Instant", "{1}{R}")
	dear := handSpell(me, "Dear Sorcery", "Sorcery", "{2}{R}")
	creature := handSpell(me, "Cheap Bear", "Creature — Bear", "{1}")
	declareAttack(t, g, opp.ID, tetsuo)
	cmdr2S43ChooseMode(t, g, me.ID, 1)
	passPriorityAroundTable(t, g)
	offer := mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, cheap)
	if hasID(offer.DiscardOptions, dear) {
		t.Error("a mana value 3 sorcery was offered with Equipment of at most 2")
	}
	if hasID(offer.DiscardOptions, creature) {
		t.Error("a creature spell was offered")
	}
}

func TestDmuB19OhabiUntapsArchersInEachOtherPlayersUntapStepOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ohabi := b12Push(g, me.ID, "Ohabi Caleria", "Legendary Creature — Elf Archer", dmuB19Ohabi, 1, 3)
	if !hasKeywordOf(g, ohabi, "reach") {
		t.Error("Ohabi has no reach")
	}
	archer := b12Creature(g, me.ID, "Archer", "Creature — Human Archer", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{ohabi, archer, bear} {
			findBattlefieldCardForTest(g, id).Tapped = true
		}
	})
	advanceToPrecombatMainOf(t, g, 1)
	tapped := func(id uuid.UUID) bool {
		c, ok := battlefieldCard(g, id)
		return ok && c.Tapped
	}
	if tapped(ohabi) || tapped(archer) {
		t.Error("Archers were not untapped in another player's untap step")
	}
	if !tapped(bear) {
		t.Error("a non-Archer untapped in another player's untap step")
	}
}

func dmuB19Damage(g *game.Game, src, victim uuid.UUID, toPlayer bool) {
	g.WithWriteLock(func() {
		if toPlayer {
			_ = g.DealDamageToPlayerForEffect(src, victim, 1)
		} else {
			_ = g.DealDamageToCreatureForEffect(src, victim, 1)
		}
	})
}

func TestDmuB19OhabiOffersACardForTwoWhenYourArcherDamagesACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Ohabi Caleria", "Legendary Creature — Elf Archer", dmuB19Ohabi, 1, 3)
	archer := b12Creature(g, me.ID, "Archer", "Creature — Human Archer", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 3, 5)

	// Not an Archer, or not a creature on the receiving end: nothing is asked.
	dmuB19Damage(g, bear, victim, false)
	dmuB19Damage(g, archer, opp.ID, true)
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		t.Fatal("a non-Archer's damage, or damage to a player, offered the draw")
	}

	hand := me.Hand.Size()
	dmuB19Damage(g, archer, victim, false)
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("no {2} offer after an Archer damaged a creature")
	}
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Error("declining the payment drew a card")
	}

	dmuB19Damage(g, archer, victim, false)
	passPriorityAroundTable(t, g)
	b06AddMana(me, "C", "C")
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want one card for the {2}", hand, me.Hand.Size())
	}
}

func TestDmuB19OhabiDoesNotTriggerOnAnOpponentsArcher(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Ohabi Caleria", "Legendary Creature — Elf Archer", dmuB19Ohabi, 1, 3)
	theirArcher := b12Creature(g, opp.ID, "Their Archer", "Creature — Human Archer", 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 5)
	dmuB19Damage(g, theirArcher, mine, false)
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		t.Error("an opponent's Archer triggered Ohabi")
	}
}

func TestDmuB19AyeshaPutsArtifactsUpToHerPowerOntoTheBattlefieldTapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ayesha := b12Push(g, me.ID, "Ayesha Tanaka, Armorer", "Legendary Creature — Human Artificer", dmuB19Ayesha, 2, 4)
	g.WithWriteLock(func() { me.Library.Cards = nil })
	mk := func(name, typeLine, cost string) uuid.UUID {
		id := uuid.New()
		g.WithWriteLock(func() {
			me.Library.PushBottom(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: me.ID, Controller: me.ID})
		})
		return id
	}
	a := mk("Cheap Rock", "Artifact", "{1}")
	b := mk("Mid Rock", "Artifact", "{2}")
	dear := mk("Dear Rock", "Artifact", "{3}")
	bear := mk("Bear", "Creature — Bear", "{1}")
	deep := mk("Deep Card", "Artifact", "{1}")
	declareAttack(t, g, opp.ID, ayesha)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, a, b)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		c, ok := battlefieldCard(g, id)
		if !ok || !c.Tapped {
			t.Errorf("artifact %s: on the battlefield %v, tapped %v; want it there tapped", id, ok, ok && c.Tapped)
		}
	}
	if g.Battlefield.Contains(dear) || g.Battlefield.Contains(bear) {
		t.Error("an artifact above her power or a nonartifact entered")
	}
	if !me.Library.Contains(deep) || !me.Library.Contains(dear) || !me.Library.Contains(bear) {
		t.Error("cards not put onto the battlefield did not stay in the library")
	}
	bottom := me.Library.Cards[len(me.Library.Cards)-2:]
	if len(me.Library.Cards) != 3 || (bottom[0].InstanceID == deep && bottom[1].InstanceID == deep) {
		t.Errorf("library holds %d cards, want the three that were not put", len(me.Library.Cards))
	}
}

func TestDmuB19AyeshaCantBeBlockedWhileTheDefenderControlsThreeArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ayesha := b12Push(g, me.ID, "Ayesha Tanaka, Armorer", "Legendary Creature — Human Artificer", dmuB19Ayesha, 2, 4)
	blocker := b12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	b12Permanent(g, opp.ID, "Rock A", "Artifact")
	b12Permanent(g, opp.ID, "Rock B", "Artifact")
	b12Permanent(g, g.Seats[2].ID, "Someone Else's Rock", "Artifact")
	brAttack(t, g, ayesha)
	if err := g.DeclareBlocker(blocker, ayesha); err != nil {
		t.Fatalf("two artifacts on the defender's side (and one elsewhere) must not make her unblockable: %v", err)
	}

	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	ayesha = b12Push(g, me.ID, "Ayesha Tanaka, Armorer", "Legendary Creature — Human Artificer", dmuB19Ayesha, 2, 4)
	blocker = b12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	for _, n := range []string{"Rock A", "Rock B", "Rock C"} {
		b12Permanent(g, opp.ID, n, "Artifact")
	}
	brAttack(t, g, ayesha)
	brRefusal(t, g.DeclareBlocker(blocker, ayesha), game.BlockReasonCantBeBlocked)
}
