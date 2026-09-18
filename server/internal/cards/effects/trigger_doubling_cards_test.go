package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestTriggerDoublerCauseHelpersUseThePrintedCauseAndLKI(t *testing.T) {
	g := game.NewGame()
	controller := uuid.New()
	subject := uuid.New()
	q := game.TriggerDoublingQuery{
		Event:      game.Event{Kind: game.EventETB, CardID: subject},
		DoublerLKI: game.Characteristic{Controller: controller},
		SourceLKI:  game.Characteristic{Controller: controller},
		Subject:    subject,
		SubjectLKI: game.Characteristic{
			Types:      []string{"Creature"},
			Controller: controller,
		},
		HasSubject: true,
	}

	if !DoublesEntering(Creature()).Applies(g, q) {
		t.Fatal("creature entering should match Panharmonicon's cause")
	}
	q.Event = game.Event{Kind: game.EventZoneMove, CardID: subject, NewZone: game.ZoneGraveyard}
	if DoublesEntering(Creature()).Applies(g, q) {
		t.Fatal("a creature entering the graveyard is not an entering cause")
	}

	q.Event = game.Event{Kind: game.EventLTB, CardID: subject, NewZone: game.ZoneGraveyard}
	if !DoublesDying(Creature()).Applies(g, q) {
		t.Fatal("a creature's battlefield LKI dying should match Teysa")
	}
	q.Event.NewZone = game.ZoneExile
	if DoublesDying(Creature()).Applies(g, q) {
		t.Fatal("leaving for exile is not a dying cause")
	}

	q.Event = game.Event{Kind: game.EventAttack, CardID: subject}
	if !DoublesAttacking(Creature()).Applies(g, q) {
		t.Fatal("a creature attack should match Isshin")
	}
	q.Event.Kind = game.EventTapCard
	if DoublesAttacking(Creature()).Applies(g, q) {
		t.Fatal("a tap caused by attacking is not an attacking cause")
	}
}

func TestTeysaKarlovGivesYourCreatureTokensVigilanceAndLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Teysa Karlov", "Legendary Creature — Human Advisor", "644eeefd-e684-4ca8-8aef-a892ca130c07", false)
	token := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Soldier", TypeLine: "Token Creature — Soldier",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	for _, keyword := range []string{"vigilance", "lifelink"} {
		if !hasEffectiveKeyword(t, g, token, keyword) {
			t.Errorf("Teysa token missing %s", keyword)
		}
	}
}

func TestTriggerDoublerCardsDeclareTheirRealNames(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
	}{
		{"Panharmonicon", "76678885-3674-443d-b9a2-2a460cf6aac0"},
		{"Teysa Karlov", "644eeefd-e684-4ca8-8aef-a892ca130c07"},
		{"Isshin, Two Heavens as One", "65114758-9a75-43a7-96e8-0aa68faa6b24"},
		{"Cloud, Midgar Mercenary", "33d2584b-bf29-4c22-bd45-14ba2fb98c0e"},
	} {
		spec, ok := Lookup(tc.oracle)
		if !ok || spec.Name != tc.name || len(spec.TriggerDoublers) != 1 {
			t.Errorf("%s registration = ok %v name %q doublers %d", tc.name, ok, spec.Name, len(spec.TriggerDoublers))
		}
		if ok && spec.TriggerDoublers[0].Label != tc.name {
			t.Errorf("%s doubler label = %q", tc.name, spec.TriggerDoublers[0].Label)
		}
	}
}

func TestPanharmoniconDoublesYourWatcherOnlyForArtifactOrCreatureEntries(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Panharmonicon", "Artifact", "76678885-3674-443d-b9a2-2a460cf6aac0", false)
	pushCatalogPermanent(g, me.ID, "Impact Tremors", "Enchantment", "9242cd3e-1a71-4700-8182-9c1005616033", false)
	pushCatalogPermanent(g, opp.ID, "Impact Tremors", "Enchantment", "9242cd3e-1a71-4700-8182-9c1005616033", false)
	before := lifeOfOpponents(g)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 1) })
	passPriorityAroundTable(t, g)
	for i, life := range before {
		if got := g.Seats[i+1].Life; got != life-2 {
			t.Errorf("own Impact Tremors after own token: %d -> %d, want -2", life, got)
		}
	}

	// The opponent's watcher triggers for an opponent's creature, but the
	// Panharmonicon controller does not control that watcher, so it remains one.
	before = lifeOfOpponents(g)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, RedGoblinToken(), 1) })
	passPriorityAroundTable(t, g)
	for i, life := range before {
		want := life - 1
		if g.Seats[i+1].ID == opp.ID {
			want = life
		}
		if got := g.Seats[i+1].Life; got != want {
			t.Errorf("opponent Impact Tremors after opponent token: %d -> %d, want %d", life, got, want)
		}
	}

	// Ruin Crab is an all-land watcher. A basic land is not an artifact or
	// creature, so Panharmonicon must leave its single trigger alone.
	ruin := pushCatalogPermanent(g, me.ID, "Ruin Crab", "Creature — Crab", "8afc00d4-a1c6-4329-af2c-a7f58a0c33e7", false)
	_ = ruin
	beforeLibrary := opp.Library.Size()
	enterFromHand(t, g, me.ID, "Forest", "Basic Land — Forest", "")
	passPriorityAroundTable(t, g)
	if got := opp.Library.Size(); got != beforeLibrary-3 {
		t.Errorf("Ruin Crab landfall milled %d cards, want 3", beforeLibrary-got)
	}
}

func TestTeysaDoublesBloodArtistForAnimatedLandAndAWholeWipe(t *testing.T) {
	t.Run("land creature LKI", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		pushCatalogPermanent(g, me.ID, "Teysa Karlov", "Legendary Creature — Human Advisor", "644eeefd-e684-4ca8-8aef-a892ca130c07", false)
		artist := pushCatalogPermanent(g, me.ID, "Blood Artist", "Creature — Vampire", "310f141c-7f37-4729-aed6-dd9c09db448d", false)
		land := pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Animated Forest", TypeLine: "Land Creature — Forest Elemental",
			Power: 0, Toughness: 1, Owner: me.ID, Controller: me.ID,
		})
		g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(land) })
		var prompts []*game.PendingChoice
		for _, choice := range g.PendingChoices {
			if choice.Kind == game.PendingChoicePickTarget && choice.Source == artist {
				prompts = append(prompts, choice)
			}
		}
		if len(prompts) != 2 {
			t.Fatalf("land creature death prompts = %d, want 2", len(prompts))
		}
		for _, choice := range prompts {
			if err := g.ResolvePickTargets(choice.ID, me.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
				t.Fatal(err)
			}
		}
		passPriorityAroundTable(t, g)
		if opp.Life != 38 || me.Life != 42 {
			t.Errorf("Blood Artist after doubled animated-land death: lives %d/%d, want 38/42", opp.Life, me.Life)
		}
	})

	t.Run("Teysa dies in wipe", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		pushCatalogPermanent(g, me.ID, "Teysa Karlov", "Legendary Creature — Human Advisor", "644eeefd-e684-4ca8-8aef-a892ca130c07", false)
		artist := pushCatalogPermanent(g, me.ID, "Blood Artist", "Creature — Vampire", "310f141c-7f37-4729-aed6-dd9c09db448d", false)
		pushCatalogPermanent(g, me.ID, "Victim", "Creature — Bear", "", false)
		_ = castCatalogSpell(t, g, "Wrath of God", "Sorcery", "34515b16-c9a4-4f98-8c77-416a7a523407", nil)
		passPriorityAroundTable(t, g)
		prompts := 0
		for _, choice := range g.PendingChoices {
			if choice.Kind == game.PendingChoicePickTarget && choice.Source == artist {
				prompts++
			}
		}
		if prompts != 6 {
			t.Errorf("wipe Blood Artist prompts = %d, want 6 (three deaths doubled)", prompts)
		}
	})
}

func TestIsshinDoublesYourAttackWatcherButNotAnOpponents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Isshin, Two Heavens as One", "Legendary Creature — Human Samurai", "65114758-9a75-43a7-96e8-0aa68faa6b24", false)
	adeline := pushCatalogPermanent(g, me.ID, "Adeline, Resplendent Cathar", "Legendary Creature — Human Knight", "38515f89-348b-4cf3-b7bd-1f6fe4ce2fba", false)
	attacker := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Attacker", TypeLine: "Creature — Soldier", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventAttack, CardID: attacker, Actor: me.ID, Target: opp.ID})
	})
	if got := countTokensControlled(g, me.ID, "Human"); got != 0 {
		t.Fatalf("Adeline tokens before trigger resolution = %d", got)
	}
	passPriorityAroundTable(t, g)
	if got := countTokensControlled(g, me.ID, "Human"); got != 6 {
		t.Errorf("your Adeline tokens with Isshin = %d, want 6", got)
	}
	_ = adeline

	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Isshin, Two Heavens as One", "Legendary Creature — Human Samurai", "65114758-9a75-43a7-96e8-0aa68faa6b24", false)
	adeline = pushCatalogPermanent(g, opp.ID, "Adeline, Resplendent Cathar", "Legendary Creature — Human Knight", "38515f89-348b-4cf3-b7bd-1f6fe4ce2fba", false)
	attacker = pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Opponent Attacker", TypeLine: "Creature — Soldier", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventAttack, CardID: attacker, Actor: opp.ID, Target: me.ID})
	})
	passPriorityAroundTable(t, g)
	if got := countTokensControlled(g, opp.ID, "Human"); got != 3 {
		t.Errorf("opponent Adeline tokens with your Isshin = %d, want 3", got)
	}
	_ = adeline
}

func TestCloudDoublesEquippedCloudAndAttachedEquipmentTriggers(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*testing.T, *game.Game, *game.Player, *game.Player, uuid.UUID)
		want  int
	}{
		{"Cloud's own trigger while equipped", func(t *testing.T, g *game.Game, me, _ *game.Player, cloud uuid.UUID) {
			equipment := seedEquipment(g, me.ID, "Sword of the Animist", "d79cbc61-6c15-48ea-bbba-3cffb819ccba")
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == equipment {
					g.Battlefield.Cards[i].AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: cloud}
				}
			}
			g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, CardID: cloud, Actor: me.ID}) })
		}, 2},
		{"attached opponent Equipment trigger", func(_ *testing.T, g *game.Game, me, opp *game.Player, cloud uuid.UUID) {
			equipment := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Sword of the Animist", TypeLine: "Legendary Artifact — Equipment", OracleID: "d79cbc61-6c15-48ea-bbba-3cffb819ccba", Owner: opp.ID, Controller: opp.ID, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: cloud}})
			g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventAttack, CardID: cloud, Actor: me.ID}) })
			_ = equipment
		}, 2},
		{"unattached Equipment trigger", func(_ *testing.T, g *game.Game, me, opp *game.Player, cloud uuid.UUID) {
			pushCatalogPermanent(g, opp.ID, "Sword of the Animist", "Legendary Artifact — Equipment", "d79cbc61-6c15-48ea-bbba-3cffb819ccba", false)
			g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventAttack, CardID: cloud, Actor: me.ID}) })
		}, 0},
		{"attached Equipment death trigger", func(t *testing.T, g *game.Game, me, _ *game.Player, cloud uuid.UUID) {
			clamp := pushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: "Skullclamp", TypeLine: "Artifact — Equipment", OracleID: "65986c1b-8e51-4604-b685-d82fa7d1263a", Owner: me.ID, Controller: me.ID,
				AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: cloud},
			})
			g.WithWriteLock(func() {
				if err := g.DestroyPermanentForEffect(cloud); err != nil {
					t.Fatal(err)
				}
			})
			_ = clamp
		}, 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			cloud := pushCatalogPermanent(g, me.ID, "Cloud, Midgar Mercenary", "Legendary Creature — Human Soldier Mercenary", "33d2584b-bf29-4c22-bd45-14ba2fb98c0e", false)
			tc.setup(t, g, me, opp, cloud)
			got := 0
			for _, item := range g.PendingTriggers {
				if item != nil && item.SourceCardID != uuid.Nil {
					got++
				}
			}
			if got != tc.want {
				t.Errorf("pending Cloud-related triggers = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestCloudDoublesAnAttachedEquipmentLeavingTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cloud := pushCatalogPermanent(g, me.ID, "Cloud, Midgar Mercenary", "Legendary Creature — Human Soldier Mercenary", "33d2584b-bf29-4c22-bd45-14ba2fb98c0e", false)
	solemn := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Solemn Simulacrum", TypeLine: "Artifact Creature — Equipment",
		OracleID: "00c0543c-2a1f-4425-8283-4062d74a1637", Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: cloud},
	})
	before := me.Hand.Size()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(solemn); err != nil {
			t.Fatal(err)
		}
	})
	var prompts []*game.PendingChoice
	for _, choice := range g.PendingChoices {
		if choice.Kind == game.PendingChoiceTriggerPrompt && choice.Source == solemn {
			prompts = append(prompts, choice)
		}
	}
	if len(prompts) != 2 {
		t.Fatalf("attached Solemn death prompts = %d, want 2", len(prompts))
	}
	for _, choice := range prompts {
		if err := g.ResolveTriggerPrompt(choice.ID, me.ID, true); err != nil {
			t.Fatal(err)
		}
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 2 {
		t.Errorf("attached Solemn doubled death draw = %d, want 2", got)
	}
}

func countTokensControlled(g *game.Game, controller uuid.UUID, subtype string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.IsToken() && c.HasSubtype(subtype) {
			n++
		}
	}
	return n
}
