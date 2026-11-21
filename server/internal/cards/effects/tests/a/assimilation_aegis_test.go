package testsa

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsap "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/ap"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// aegisTable is a catalog game in seat 0's main phase with two creatures
// of seat 0's (Bear, Elk) and seat 1's Hill Giant.
type aegisTable struct {
	g                *game.Game
	me, opp          *game.Player
	bear, elk, giant uuid.UUID
}

func newAegisTable(t *testing.T) aegisTable {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	push := func(owner uuid.UUID, name, typeLine, cost string, p, tough int) uuid.UUID {
		return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: cost,
			Power: p, Toughness: tough, Owner: owner, Controller: owner,
		})
	}
	return aegisTable{
		g: g, me: me, opp: opp,
		bear:  push(me.ID, "Bear", "Creature — Bear", "{1}{G}", 2, 2),
		elk:   push(me.ID, "Elk", "Creature — Elk", "{2}{G}", 3, 1),
		giant: push(opp.ID, "Hill Giant", "Creature — Giant", "{3}{R}", 3, 3),
	}
}

// castAegis casts Assimilation Aegis and resolves it and its enters
// trigger, exiling `exile` (uuid.Nil: no target is picked).
func (a aegisTable) castAegis(t *testing.T, exile uuid.UUID) uuid.UUID {
	t.Helper()
	aegis := effectstest.CastCatalogSpell(t, a.g, "Assimilation Aegis", "Artifact — Equipment", cardsap.AssimilationAegisOracleID, nil)
	effectstest.PassPriorityAroundTable(t, a.g)
	if exile != uuid.Nil {
		effectstest.AnswerPickTarget(t, a.g, exile)
	}
	effectstest.PassPriorityAroundTable(t, a.g)
	if !a.g.Battlefield.Contains(aegis) {
		t.Fatal("Assimilation Aegis is not on the battlefield")
	}
	return aegis
}

// aegisLinkedExile puts a card into exile linked to the Aegis's enters
// row, as a second resolution of that trigger would (the ruling's
// "copied or triggered more than once").
func aegisLinkedExile(g *game.Game, aegis uuid.UUID, owner uuid.UUID, name, typeLine string, p, tough int) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		var epoch int
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == aegis {
				epoch = c.ObjectEpoch
			}
		}
		g.Exile.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, Power: p, Toughness: tough,
			Owner: owner, Controller: owner,
			ExiledWith:        game.PermissionCardRef{ID: aegis, Epoch: epoch},
			ExiledWithAbility: game.AbilityRef{Slot: game.AbilitySlotTriggered, Name: cardsap.AssimilationAegisExileKey},
		})
	})
	return id
}

// The enters trigger exiles the creature linked to this Aegis, and the
// card comes back when the Aegis leaves.
func TestAssimilationAegisExilesUntilItLeaves(t *testing.T) {
	a := newAegisTable(t)
	aegis := a.castAegis(t, a.giant)
	if !a.g.Exile.Contains(a.giant) {
		t.Fatal("the Hill Giant was not exiled")
	}
	if got := effectstest.AegisCardNow(a.g, a.giant); got.ExiledWith.ID != aegis || !got.ExiledWithAbility.IsRow(game.AbilitySlotTriggered, cardsap.AssimilationAegisExileKey) {
		t.Fatalf("exiled card's link = %+v / %+v, want the Aegis's enters row", got.ExiledWith, got.ExiledWithAbility)
	}
	effectstest.DestroyNow(t, a.g, aegis)
	a.g.RunStateChecksForTest()
	// It returns as a new object (CR 400.7), under its owner's control
	// (CR 610.3c).
	if effectstest.BattlefieldNamed(a.g, a.opp.ID, "Hill Giant") != 1 {
		t.Fatal("the Hill Giant did not return to its owner when the Aegis left")
	}
}

// Equipping makes the creature a copy of the exiled card; moving the
// Aegis ends that copy and starts one on the new creature; the copy
// keeps the creature's own counters.
func TestAssimilationAegisCopyFollowsTheAttachment(t *testing.T) {
	a := newAegisTable(t)
	aegis := a.castAegis(t, a.giant)
	a.g.WithWriteLock(func() {
		for i := range a.g.Battlefield.Cards {
			if a.g.Battlefield.Cards[i].InstanceID == a.bear {
				a.g.Battlefield.Cards[i].Counters = map[string]int{game.CounterPlusOne: 1}
			}
		}
	})

	effectstest.EquipTo(t, a.g, a.me.ID, aegis, a.bear)
	effectstest.AegisWantName(t, a.g, a.bear, "Hill Giant", 4)
	if c := effectstest.AegisCardNow(a.g, a.bear); c.Controller != a.me.ID || !c.HasSubtype("Giant") {
		t.Errorf("the copy is %+v, want my Giant", c.TypeLine)
	}

	effectstest.EquipTo(t, a.g, a.me.ID, aegis, a.elk)
	effectstest.AegisWantName(t, a.g, a.bear, "Bear", 3)
	effectstest.AegisWantName(t, a.g, a.elk, "Hill Giant", 3)

	// The Aegis leaving ends the copy and returns the card.
	effectstest.DestroyNow(t, a.g, aegis)
	a.g.RunStateChecksForTest()
	a.g.WithWriteLock(func() { a.g.RecomputeLayersIfStaleLocked() })
	effectstest.AegisWantName(t, a.g, a.elk, "Elk", 3)
	if effectstest.BattlefieldNamed(a.g, a.opp.ID, "Hill Giant") != 1 {
		t.Error("the Hill Giant did not return")
	}
}

// No card exiled with the Aegis, or only a noncreature card: the creature
// stays as it was.
func TestAssimilationAegisWithNoExiledCreatureCardCopiesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	aegis := effectstest.CastCatalogSpell(t, g, "Assimilation Aegis", "Artifact — Equipment", cardsap.AssimilationAegisOracleID, nil)
	effectstest.PassPriorityAroundTable(t, g) // no creature to target
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID,
	})
	aegisLinkedExile(g, aegis, g.Seats[1].ID, "Ornithopter Blueprint", "Artifact", 0, 0)
	effectstest.EquipTo(t, g, me.ID, aegis, bear)
	if host := effectstest.AttachmentHostOf(t, g, aegis); host.ID != bear {
		t.Fatalf("the Aegis is attached to %+v", host)
	}
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("asked to choose with no creature card exiled")
	}
	effectstest.AegisWantName(t, g, bear, "Bear", 2)
}

// Two creature cards exiled with it: the controller picks one, each time
// (the ruling), the bot can answer the pick, and the enumerator offers the
// equip.
func TestAssimilationAegisPicksAmongSeveralAndTheBotCanPlayIt(t *testing.T) {
	a := newAegisTable(t)
	aegis := a.castAegis(t, a.giant)
	drake := aegisLinkedExile(a.g, aegis, a.opp.ID, "Wind Drake", "Creature — Drake", 2, 2)
	aegisLinkedExile(a.g, aegis, a.opp.ID, "Ornithopter Blueprint", "Artifact", 0, 0)

	effectstest.FloatMana(t, a.g, a.me, "{C}{C}")
	offersEquip := false
	for _, m := range legal.EnumerateFor(a.g, a.me.ID) {
		if m.Type == "activate_ability" && m.Source == aegis {
			offersEquip = true
		}
	}
	if !offersEquip {
		t.Fatal("the enumerator offers no equip for the Aegis")
	}

	effectstest.EquipTo(t, a.g, a.me.ID, aegis, a.bear)
	pick := effectstest.ChooseCardsChoiceFor(a.g, a.me.ID)
	if pick == nil {
		t.Fatal("no pick among the two exiled creature cards")
	}
	offered := map[string]bool{}
	for _, m := range legal.EnumerateFor(a.g, a.me.ID) {
		var p struct {
			CardIDs []string `json:"card_ids"`
		}
		if json.Unmarshal(m.Params, &p) == nil && len(p.CardIDs) == 1 {
			offered[p.CardIDs[0]] = true
		}
	}
	if len(offered) != 2 || !offered[a.giant.String()] || !offered[drake.String()] {
		t.Fatalf("the bot is offered %v, want the Hill Giant and the Wind Drake only", offered)
	}
	effectstest.AnswerChooseCards(t, a.g, a.me.ID, drake)
	effectstest.PassPriorityAroundTable(t, a.g)
	effectstest.AegisWantName(t, a.g, a.bear, "Wind Drake", 2)

	// Moving it asks again, and a different card may be chosen.
	effectstest.EquipTo(t, a.g, a.me.ID, aegis, a.elk)
	effectstest.AnswerChooseCards(t, a.g, a.me.ID, a.giant)
	effectstest.PassPriorityAroundTable(t, a.g)
	effectstest.AegisWantName(t, a.g, a.bear, "Bear", 2)
	effectstest.AegisWantName(t, a.g, a.elk, "Hill Giant", 3)
}

// The Aegis destroyed with its attach trigger on the stack: the "for as
// long as" duration has ended before the copy would begin, so nothing is
// copied (CR 611.2b).
func TestAssimilationAegisGoneBeforeTheTriggerResolvesCopiesNothing(t *testing.T) {
	a := newAegisTable(t)
	aegis := a.castAegis(t, a.giant)
	effectstest.ActivateEquipWithoutResolving(t, a.g, a.me.ID, aegis, a.bear)
	// Pass one priority at a time until the equip has resolved and the
	// attach trigger is waiting on the stack.
	triggered := func() bool {
		for _, it := range a.g.StackMeta {
			if it != nil && it.Kind == game.StackItemTriggered && it.SourceCardID == aegis {
				return true
			}
		}
		return false
	}
	for i := 0; i < 16 && !triggered(); i++ {
		if err := a.g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !triggered() {
		t.Fatal("attaching the Aegis did not trigger it")
	}
	effectstest.DestroyNow(t, a.g, aegis)
	effectstest.PassPriorityAroundTable(t, a.g)
	effectstest.AegisWantName(t, a.g, a.bear, "Bear", 2)
}

// Strict refusals: the cast is paid for, the enters trigger takes only a
// creature, and equip takes only a creature you control.
func TestAssimilationAegisStrictRefusals(t *testing.T) {
	a := newAegisTable(t)
	id := effectstest.HandSpell(a.me, "Assimilation Aegis", "Artifact — Equipment", "{1}{W}{U}")
	a.g.WithWriteLock(func() {
		for i := range a.me.Hand.Cards {
			if a.me.Hand.Cards[i].InstanceID == id {
				a.me.Hand.Cards[i].OracleID = cardsap.AssimilationAegisOracleID
			}
		}
	})
	effectstest.RefusedForMana(t, a.g.CastSpell(a.me.ID, id, game.CastSpellParams{Strict: true}), "Assimilation Aegis with no mana")

	land := effectstest.PushBattlefieldCardWithTimestamp(a.g, game.Card{
		InstanceID: uuid.New(), Name: "Plains", TypeLine: "Basic Land — Plains", Owner: a.me.ID, Controller: a.me.ID,
	})
	aegis := effectstest.CastCatalogSpell(t, a.g, "Assimilation Aegis", "Artifact — Equipment", cardsap.AssimilationAegisOracleID, nil)
	effectstest.PassPriorityAroundTable(t, a.g)
	ch := effectstest.PickTargetPrompt(t, a.g)
	if err := a.g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: land}); err == nil {
		t.Error("the enters trigger took a land")
	}
	effectstest.AnswerPickTarget(t, a.g, a.giant)
	effectstest.PassPriorityAroundTable(t, a.g)

	abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: cardsap.AssimilationAegisOracleID})
	if err := a.g.ActivateCatalogAbility(a.me.ID, aegis, len(abilities)-1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}},
	}); err == nil {
		t.Error("equip targeted a land")
	}
	theirs := effectstest.PushVanillaCreature(a.g, a.opp.ID, "Their Bear", 2, 2)
	if err := a.g.ActivateCatalogAbility(a.me.ID, aegis, len(abilities)-1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
	}); err == nil {
		t.Error("equip targeted an opponent's creature")
	}
	if host := effectstest.AttachmentHostOf(t, a.g, aegis); host.ID != uuid.Nil {
		t.Errorf("the Aegis is attached to %+v after refused equips", host)
	}
}

// The attach trigger fires on an attach to a creature and nothing else:
// not a re-attach to the creature it is on (CR 701.3b), not another
// permanent's attach.
func TestAssimilationAegisAttachTriggerCondition(t *testing.T) {
	a := newAegisTable(t)
	aegis := a.castAegis(t, a.giant)
	var src game.Card
	a.g.ReadSnapshot(func() { src = *effectstest.FindBattlefieldCardForTest(a.g, aegis) })
	cases := []struct {
		name string
		ev   game.Event
		want bool
	}{
		{"onto a creature", game.Event{Kind: game.EventAttach, CardID: aegis, Target: a.bear}, true},
		{"onto the creature it is on", game.Event{Kind: game.EventAttach, CardID: aegis, Target: a.bear, AlreadyAttached: true}, false},
		{"another attachment", game.Event{Kind: game.EventAttach, CardID: uuid.New(), Target: a.bear}, false},
		{"onto a player", game.Event{Kind: game.EventAttach, CardID: aegis, Target: a.opp.ID}, false},
	}
	for _, tc := range cases {
		var got bool
		a.g.ReadSnapshot(func() { got = effects.ThisBecameAttachedToACreature(tc.ev, &src, game.Characteristic{}, a.g) })
		if got != tc.want {
			t.Errorf("%s: triggered = %v, want %v", tc.name, got, tc.want)
		}
	}
}
