package game

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// shield_recipient_filter_test.go — #2045 (ADR 0108 §7, amendment of
// 2026-10-06): a preventFromSource shield that protects a set — every
// creature, the creatures you control, players, the permanents of a
// type — read as the damage would be dealt (CR 611.2c, 615.1).

// recipientShield registers a preventFromSource shield for `controller`
// protecting the set `f` from sources matching `qs` (none: any source).
func recipientShield(t *testing.T, g *Game, controller uuid.UUID, f DamageRecipientFilter, qs ...PermanentQuery) {
	t.Helper()
	g.WithWriteLock(func() {
		s := DamageShield{Controller: controller, Queries: qs, Recipients: f, Label: "test recipient shield"}
		if !g.PreventDamageFromSourceThisTurnForEffect(s) {
			t.Fatal("no shield registered")
		}
	})
}

// strike deals 1 from `source` to each creature and returns the damage
// now marked on each.
func strike(t *testing.T, g *Game, source uuid.UUID, creatures ...uuid.UUID) []int {
	t.Helper()
	out := make([]int, len(creatures))
	g.WithWriteLock(func() {
		for _, c := range creatures {
			if err := g.DealDamageToCreatureForEffect(source, c, 1); err != nil {
				t.Fatal(err)
			}
		}
	})
	for i, c := range creatures {
		out[i] = findBattlefieldCard(g, c).DamageMarked
	}
	return out
}

func pushArtifactSource(g *Game, owner *Player, name string) uuid.UUID {
	c := NewCard(name, owner.ID)
	c.TypeLine = "Artifact"
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

var everyCreature = DamageRecipientFilter{Permanents: true, Match: []PermanentQuery{creatureQuery}}

// "To creatures" (Forfend): every creature, whoever controls it, one
// that entered after the shield included; no player.
func TestRecipientFilterEveryCreature(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushArtifactSource(g, opp, "Rod")
	mine := pushCombatant(t, g, me, "Mine", 2, 3)
	recipientShield(t, g, me.ID, everyCreature)
	theirs := pushCombatant(t, g, opp, "Late", 2, 3)
	freshLayers(g)
	if got := strike(t, g, src, mine, theirs); got[0] != 0 || got[1] != 0 {
		t.Fatalf("marked %v, want none: every creature is protected, a late one too (CR 611.2c)", got)
	}
	if lost := hits(t, g, me.ID, 1, src); lost != 1 {
		t.Fatalf("lost %d, want 1: a player is not a creature", lost)
	}
}

// "To creatures you control" (Divine Light): you are not protected, an
// opponent's creature is not, and control is read as the damage would be
// dealt.
func TestRecipientFilterCreaturesYouControlFollowsControl(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushArtifactSource(g, opp, "Rod")
	mine := pushCombatant(t, g, me, "Mine", 2, 5)
	theirs := pushCombatant(t, g, opp, "Theirs", 2, 5)
	recipientShield(t, g, me.ID, DamageRecipientFilter{Permanents: true, Match: []PermanentQuery{creatureQuery},
		Controller: RecipientControllerYou})
	if got := strike(t, g, src, mine, theirs); got[0] != 0 || got[1] != 1 {
		t.Fatalf("marked %v, want [0 1]: only my creature is protected", got)
	}
	if lost := hits(t, g, me.ID, 1, src); lost != 1 {
		t.Fatalf("lost %d, want 1: \"creatures you control\" does not protect you", lost)
	}
	// Swap control: the protected set follows (CR 611.2c).
	findBattlefieldCard(g, mine).Controller = opp.ID
	findBattlefieldCard(g, theirs).Controller = me.ID
	freshLayers(g)
	if got := strike(t, g, src, mine, theirs); got[0] != 1 || got[1] != 1 {
		t.Fatalf("marked %v, want [1 1]: the creature I lost is dealt damage, the one I took is protected", got)
	}
}

// "Damage that creatures would deal to players" (Chameleon Blur): a
// players-only set beside a source property.
func TestRecipientFilterPlayersComposesWithASourceProperty(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushCombatant(t, g, opp, "Bear", 2, 2)
	rod := pushArtifactSource(g, opp, "Rod")
	mine := pushCombatant(t, g, me, "Mine", 2, 5)
	freshLayers(g)
	recipientShield(t, g, me.ID, DamageRecipientFilter{Players: true}, creatureQuery)
	if lost := hits(t, g, me.ID, 1, bear, rod); lost != 1 {
		t.Fatalf("lost %d, want 1: the creature's damage prevented, the artifact's dealt", lost)
	}
	if lost := hits(t, g, opp.ID, 1, bear); lost != 0 {
		t.Fatalf("lost %d, want 0: every player is protected", lost)
	}
	if got := strike(t, g, bear, mine); got[0] != 1 {
		t.Fatalf("marked %v, want [1]: a creature is not a player", got)
	}
}

// "To artifact creatures" (Ethersworn Shieldmage): every query must
// match; "to Dogs you control" (Pack Leader) is a subtype.
func TestRecipientFilterMatchesEveryQuery(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushArtifactSource(g, opp, "Rod")
	golemCard := NewCard("Golem", opp.ID)
	golemCard.TypeLine = "Artifact Creature — Golem"
	golemCard.Power, golemCard.Toughness = 2, 3
	g.Battlefield.PushTop(golemCard)
	golem := golemCard.InstanceID
	bear := pushCombatant(t, g, opp, "Bear", 2, 3)
	freshLayers(g)
	recipientShield(t, g, me.ID, DamageRecipientFilter{Permanents: true,
		Match: []PermanentQuery{{Types: []string{"artifact"}}, creatureQuery}})
	if got := strike(t, g, src, golem, bear); got[0] != 0 || got[1] != 1 {
		t.Fatalf("marked %v, want [0 1]: the artifact creature protected, the plain creature not", got)
	}

	g2 := newActiveGame(t)
	me, opp = g2.Seats[0], g2.Seats[1]
	src = pushArtifactSource(g2, opp, "Rod")
	dogCard := NewCard("Dog", me.ID)
	dogCard.TypeLine = "Creature — Dog"
	dogCard.Power, dogCard.Toughness = 2, 3
	g2.Battlefield.PushTop(dogCard)
	theirDog := NewCard("Their Dog", opp.ID)
	theirDog.TypeLine = "Creature — Dog"
	theirDog.Power, theirDog.Toughness = 2, 3
	g2.Battlefield.PushTop(theirDog)
	cat := pushCombatant(t, g2, me, "Cat", 2, 3)
	freshLayers(g2)
	recipientShield(t, g2, me.ID, DamageRecipientFilter{Permanents: true,
		Match: []PermanentQuery{{Subtypes: []string{"Dog"}}}, Controller: RecipientControllerYou})
	if got := strike(t, g2, src, dogCard.InstanceID, theirDog.InstanceID, cat); got[0] != 0 || got[1] != 1 || got[2] != 1 {
		t.Fatalf("marked %v, want [0 1 1]: only my Dog is protected", got)
	}
}

// What registration and restore refuse.
func TestRecipientFilterRefusals(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	c := pushCombatant(t, g, me, "Mine", 2, 2)
	g.WithWriteLock(func() {
		bad := []DamageShield{
			{Controller: me.ID, ProtectPlayer: me.ID, Recipients: everyCreature},
			{Controller: me.ID, ProtectPermanent: c, Recipients: everyCreature},
			{Controller: me.ID, ProtectPermanent: c, AndDealtBy: true, Recipients: everyCreature},
			{Controller: me.ID, Amount: 3, Queries: []PermanentQuery{creatureQuery}, Recipients: everyCreature},
			{Controller: me.ID, Recipients: DamageRecipientFilter{Match: []PermanentQuery{creatureQuery}}},
			{Controller: me.ID, Recipients: DamageRecipientFilter{Players: true, Controller: RecipientControllerYou}},
			{Controller: me.ID, Recipients: DamageRecipientFilter{Permanents: true, Controller: "teammates"}},
			{Controller: me.ID, Recipients: DamageRecipientFilter{Permanents: true, Match: []PermanentQuery{{}}}},
		}
		for i, s := range bad {
			if g.PreventDamageFromSourceThisTurnForEffect(s) {
				t.Errorf("shield %d registered: %+v", i, s.Recipients)
			}
		}
	})
	f := []DamageRecipientFilter{everyCreature}
	for _, m := range []Mod{
		{Kind: ModPreventNextFromSource, Queries: []PermanentQuery{creatureQuery}, RecipientFilter: f},
		{Kind: ModPreventDamage, Amount: 1, RecipientFilter: f},
		{Kind: ModMultiplyDamage, Amount: 2, RecipientFilter: f},
		{Kind: ModPreventFromSource, RecipientFilter: []DamageRecipientFilter{{}}},
		{Kind: ModPreventFromSource, RecipientFilter: []DamageRecipientFilter{everyCreature, everyCreature}},
		{Kind: ModPreventFromSource, Player: me.ID, RecipientFilter: f},
		{Kind: ModPreventFromSource, Types: []string{"creature"}, RecipientFilter: f},
		{Kind: ModPreventFromSource, Amount: 2, Queries: []PermanentQuery{creatureQuery}, RecipientFilter: f},
	} {
		if nextFromSourceModProblem(m) == "" {
			t.Errorf("mod %+v passed the check", m)
		}
	}
	if p := recipientFilterScopeProblem(ScopeNone, []Mod{{Kind: ModPreventFromSource, RecipientFilter: f}}); p == "" {
		t.Error("a pinned record with a recipient filter passed the check")
	}
}

// A recipient shield is a restore point: it comes back and still reads
// its set, and a key inside the filter that this binary does not know is
// refused (ADR 0041 P4), not dropped.
func TestRecipientFilterSurvivesARestorePoint(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushArtifactSource(g, opp, "Rod")
	mine := pushCombatant(t, g, me, "Mine", 2, 5)
	theirs := pushCombatant(t, g, opp, "Theirs", 2, 5)
	freshLayers(g)
	recipientShield(t, g, me.ID, DamageRecipientFilter{Permanents: true, Match: []PermanentQuery{creatureQuery},
		Controller: RecipientControllerYou})
	_, restored := roundTrip(t, g)
	if got := strike(t, restored, src, mine, theirs); got[0] != 0 || got[1] != 1 {
		t.Fatalf("marked %v after the restore, want [0 1]", got)
	}

	for _, poke := range []func(mod map[string]any){
		func(mod map[string]any) { mod["recipientFilter"].([]any)[0].(map[string]any)["teammates"] = true },
		func(mod map[string]any) {
			mod["recipientFilter"].([]any)[0].(map[string]any)["match"].([]any)[0].(map[string]any)["Power"] = 3
		},
		func(mod map[string]any) {
			mod["recipientFilter"].([]any)[0].(map[string]any)["controller"] = "teammates"
		},
	} {
		raw, err := json.Marshal(g.CaptureSnapshot())
		if err != nil {
			t.Fatal(err)
		}
		var doc map[string]any
		if err := json.Unmarshal(raw, &doc); err != nil {
			t.Fatal(err)
		}
		effects := doc["scopedEffects"].([]any)
		poke(effects[0].(map[string]any)["mods"].([]any)[0].(map[string]any))
		raw, err = json.Marshal(doc)
		if err != nil {
			t.Fatal(err)
		}
		var snap GameSnapshot
		if err := json.Unmarshal(raw, &snap); err != nil {
			t.Fatal(err)
		}
		if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
			t.Fatalf("restore error %v, want ErrUnknownEffectKey", err)
		}
	}
}
