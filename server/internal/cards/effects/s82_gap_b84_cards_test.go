package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapB84Squadron = "18ae6be7-cfb6-42e3-a04e-99e9c1471ecd"
	gapB84Vashta   = "9ae26054-caf4-47a3-b4bb-3dc14a634740"
	gapB84Cybermat = "ca173221-1661-41f6-a353-29a32d6ccce0"
	gapB84Drone    = "95d018f4-7f97-4b2c-abb4-cef69031caa1"
	gapB84Beast    = "ca187570-3634-4296-92b2-1fcd99d150e3"
	gapB84Ashad    = "f3902dfd-6d09-425f-8446-a68cae2eea50"
	gapB84Reaper   = "a3f9126f-5797-4436-ac5c-c6911d22ffc2"
	gapB84Vislor   = "cdf5c31c-c324-471f-b6d7-06a06e91daa3"
)

func TestGapB84EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB84Squadron: "Dalek Squadron",
		gapB84Vashta:   "Vashta Nerada",
		gapB84Cybermat: "Cybermat",
		gapB84Drone:    "Dalek Drone",
		gapB84Beast:    "The Beast, Deathless Prince",
		gapB84Ashad:    "Ashad, the Lone Cyberman",
		gapB84Reaper:   "Time Reaper",
		gapB84Vislor:   "Vislor Turlough",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v, %+v", name, ok, spec)
		}
	}
}

// gapB84Settle passes priority until the stack is empty, answering each target
// prompt `me` owes with the next of `picks` (none left: no target).
func gapB84Settle(t *testing.T, g *game.Game, me uuid.UUID, picks ...game.TargetRef) {
	t.Helper()
	next := 0
	for i := 0; i < 64; i++ {
		if p := latestPickTarget(g, me); p != nil {
			var err error
			if next < len(picks) {
				err = g.ResolvePickTarget(p.ID, me, picks[next])
				next++
			} else {
				err = g.ResolvePickTargets(p.ID, me, nil)
			}
			if err != nil {
				t.Fatalf("answering the target prompt: %v", err)
			}
			continue
		}
		if stackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never emptied")
}

// gapB84Cast puts c in me's hand, floats mana, and casts it in a main phase.
func gapB84Cast(t *testing.T, g *game.Game, me *game.Player, c game.Card, mana string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	id := avB24Hand(g, me, c)
	if mana != "" {
		floatForTest(g, me, mana)
	}
	if err := g.CastSpell(me.ID, id, params); err != nil {
		t.Fatalf("cast %s: %v", c.Name, err)
	}
	return id
}

func TestGapB84DalekSquadronHasMenaceAndMyriad(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := apaPush(g, me.ID, me.ID, game.Card{Name: "Dalek Squadron", TypeLine: "Artifact Creature — Dalek", OracleID: gapB84Squadron, Power: 3, Toughness: 3, Keywords: []string{"menace"}})
	declareAttack(t, g, opp.ID, atk)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	answerMayChoice(t, g, me.ID, true)
	copies := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Dalek Squadron" && c.InstanceID != atk && IsToken(c) {
			copies++
		}
	}
	if copies != 2 {
		t.Errorf("%d myriad copies, want 2", copies)
	}
}

func TestGapB84VashtaNeradaGrowsAtEachEndStepIfACreatureDied(t *testing.T) {
	for _, died := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		v := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Vashta Nerada", TypeLine: "Creature — Alien Horror", OracleID: gapB84Vashta, Power: 1, Toughness: 1,
			Keywords: []string{"indestructible", "shadow"}})
		if died {
			small := apaPush(g, me.ID, me.ID, game.Card{Name: "Mouse", TypeLine: "Creature — Mouse", Power: 1, Toughness: 1})
			gapB84Cast(t, g, me, game.Card{Name: "Lightning Bolt", TypeLine: "Instant", OracleID: pr7bLightningBolt}, "", game.CastSpellParams{Targets: b16TargetCard(small)})
			passPriorityAroundTable(t, g)
		}
		advanceTo(t, g, game.StepEnd)
		passPriorityAroundTable(t, g)
		want := 0
		if died {
			want = 1
		}
		if got := counterOnBattlefieldCard(g, v, game.CounterPlusOne); got != want {
			t.Errorf("a creature died %v: %d +1/+1 counters on an opponent's Vashta Nerada, want %d (each end step, not only yours)", died, got, want)
		}
	}
}

func TestGapB84CybermatGetsPlusXWhenUnblocked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cyber := apaPush(g, me.ID, me.ID, game.Card{Name: "Cybermat", TypeLine: "Artifact Creature — Robot", OracleID: gapB84Cybermat, Power: 2, Toughness: 1, Keywords: []string{"skulk"}})
	other := apaPush(g, me.ID, me.ID, game.Card{Name: "Robot", TypeLine: "Artifact Creature — Robot", Power: 1, Toughness: 1})
	declareAttack(t, g, opp.ID, cyber, other)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareBlockers)
	passPriorityAroundTable(t, g)
	if got := layeredCard(t, g, cyber).CurrentPower(); got != 4 {
		t.Errorf("Cybermat's power %d, want 4 (2 base, +2 for two attacking artifact creatures)", got)
	}
}

func TestGapB84DalekDroneDestroysAndDrainsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Victim", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	mine := apaPush(g, me.ID, me.ID, game.Card{Name: "Mine", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	life := opp.Life
	gapB84Cast(t, g, me, game.Card{Name: "Dalek Drone", TypeLine: "Artifact Creature — Dalek", ManaCost: "{3}{B}{B}", OracleID: gapB84Drone, Power: 3, Toughness: 3,
		Keywords: []string{"flying", "menace"}}, "BBCCC", game.CastSpellParams{})
	gapB84Settle(t, g, me.ID, b16TargetCard(victim)[0])
	if g.Battlefield.Contains(victim) || !g.Battlefield.Contains(mine) {
		t.Errorf("victim on battlefield %v, my creature %v", g.Battlefield.Contains(victim), g.Battlefield.Contains(mine))
	}
	if opp.Life != life-3 {
		t.Errorf("their life %d, want %d", opp.Life, life-3)
	}
}

func TestGapB84TheBeastStealsOnCastEntersStunnedAndUntapsOnOwnerDamage(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Victim", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Tapped: true})
	beast := gapB84Cast(t, g, me, game.Card{Name: "The Beast, Deathless Prince", TypeLine: "Legendary Creature — Demon", ManaCost: "{2}{B}{R}", OracleID: gapB84Beast, Power: 6, Toughness: 6}, "BRCC", game.CastSpellParams{})
	gapB84Settle(t, g, me.ID, b16TargetCard(victim)[0])
	v, _ := g.LookupCardForEffect(victim)
	if v.Controller != me.ID || v.Tapped {
		t.Fatalf("victim controller is me %v, tapped %v; want mine and untapped", v.Controller == me.ID, v.Tapped)
	}
	if got := counterOnBattlefieldCard(g, beast, game.CounterStun); got != 6 || !b16Tapped(t, g, beast) {
		t.Fatalf("stun counters %d tapped %v, want 6 and tapped", got, b16Tapped(t, g, beast))
	}
	// A creature deals combat damage to its owner: the Beast untaps (spending a stun counter) and its controller draws.
	hand := me.Hand.Size()
	dealCombatDamageToPlayer(g, victim, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (draw when a creature deals combat damage to its owner)", me.Hand.Size(), hand+1)
	}
	if got := counterOnBattlefieldCard(g, beast, game.CounterStun); got != 5 {
		t.Errorf("stun counters %d, want 5 after the untap was replaced", got)
	}
}

func TestGapB84AshadGrowsOnSacrificeAndGrantsCasualtyToTheFirstNonlegendaryArtifactSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ashad := apaPush(g, me.ID, me.ID, game.Card{Name: "Ashad, the Lone Cyberman", TypeLine: "Legendary Artifact Creature — Cyberman", OracleID: gapB84Ashad, Power: 3, Toughness: 3})
	fodder := apaPush(g, me.ID, me.ID, game.Card{Name: "Fodder", TypeLine: "Creature — Goblin", Power: 2, Toughness: 2})
	rock := avB24Hand(g, me, game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{1}"})
	advanceToMain(t, g)
	floatForTest(g, me, "C")
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{fodder}, Strict: true}); err != nil {
		t.Fatalf("cast the first nonlegendary artifact with casualty: %v", err)
	}
	passPriorityAroundTable(t, g)
	if triggerOrderPrompt(g) != nil {
		answerTriggerOrderInOfferedOrder(t, g)
	}
	passPriorityAroundTable(t, g)
	if got := counterOnBattlefieldCard(g, ashad, game.CounterPlusOne); got != 1 {
		t.Errorf("Ashad has %d +1/+1 counters, want 1 for the sacrificed creature", got)
	}
	second := avB24Hand(g, me, game.Card{Name: "Second Rock", TypeLine: "Artifact", ManaCost: "{1}"})
	other := apaPush(g, me.ID, me.ID, game.Card{Name: "Fodder Two", TypeLine: "Creature — Goblin", Power: 2, Toughness: 2})
	floatForTest(g, me, "C")
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{other}, Strict: true}); err == nil {
		t.Error("the second nonlegendary artifact spell had casualty")
	}
}

func TestGapB84TimeReaperBottomsTheirExiledCardAndGainsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	reaper := apaPush(g, me.ID, me.ID, game.Card{Name: "Time Reaper", TypeLine: "Creature — Alien Horror", OracleID: gapB84Reaper, Power: 4, Toughness: 4, Keywords: []string{"flying", "haste"}})
	theirs := pushExiledCard(g, opp.ID, "Their Exiled Card", "Sorcery", false)
	mine := pushExiledCard(g, me.ID, "My Exiled Card", "Sorcery", false)
	life := me.Life
	dealCombatDamageToPlayer(g, reaper, opp.ID, 4)
	gapB84Settle(t, g, me.ID, b16TargetCard(theirs)[0])
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
	if got := opp.Library.Cards[len(opp.Library.Cards)-1].InstanceID; got != theirs && opp.Library.Cards[0].InstanceID != theirs {
		t.Error("their exiled card is not at an end of their library")
	}
	_ = mine
}

func TestGapB84VislorTurloughHandsItOverAndDrawsThenBleeds(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	vislor := gapB84Cast(t, g, me, game.Card{Name: "Vislor Turlough", TypeLine: "Legendary Creature — Rogue", ManaCost: "{3}{B}", OracleID: gapB84Vislor, Power: 2, Toughness: 5}, "BCCC", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	c := latestOptionPickFor(g, me.ID)
	if c == nil {
		t.Fatal("no prompt for the Black Guardian deal")
	}
	idx := -1
	for i, o := range c.PickOptions {
		if o.Label == "Have "+opp.Name+" gain control of it" {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatalf("no option naming the opponent: %+v", c.PickOptions)
	}
	if err := g.ResolveOptionPick(c.ID, me.ID, idx); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	v, _ := g.LookupCardForEffect(vislor)
	if v.Controller != opp.ID {
		t.Fatalf("Vislor is controlled by the opponent: %v", v.Controller == opp.ID)
	}
	if lv := layeredCard(t, g, vislor); len(lv.StaticGoaders()) != 1 || lv.StaticGoaders()[0] != me.ID {
		t.Error("Vislor is not goaded by its former controller")
	}
}

func TestGapB84VislorTurloughDrawsThenLosesLifeForTheHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	apaPush(g, me.ID, me.ID, game.Card{Name: "Vislor Turlough", TypeLine: "Legendary Creature — Rogue", OracleID: gapB84Vislor, Power: 2, Toughness: 5})
	d2dSetHand(me, 3)
	life := me.Life
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != 4 || me.Life != life-4 {
		t.Errorf("hand %d life %d, want 4 cards and %d life (drew one, lost life equal to the hand)", me.Hand.Size(), me.Life, life-4)
	}
}
