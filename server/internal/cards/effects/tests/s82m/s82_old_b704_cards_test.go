package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b704_cards_test.go — Urza's Destiny damage cards, S82 OLD batch 704
// (#5621), played through the real catalog entries.

const (
	o704AetherSting   = "61105cb5-d7a1-4021-a006-dd1b947dfa68"
	o704Masticore     = "621476a4-44bf-4f55-adb1-880cdd213f38"
	o704CinderSeer    = "64b63797-9db0-41fa-b2eb-fbf51de0e337"
	o704Impatience    = "97c8f5e1-9fc1-4a62-a757-1c0454754cd3"
	o704Caltrops      = "b13f97b3-56ae-4131-a93b-b162c8ff7cca"
	o704GoblinFestval = "de21b90f-a268-46d6-b977-d5b0856176f3"
	o704Landslide     = "f15761cf-2ac5-4458-957c-15badb1dda50"
	o704ScentOfCinder = "ff50d002-834e-4b5e-8db0-b0c8b4efb21f"
)

func TestS82OldB704CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		o704AetherSting: "Aether Sting", o704Masticore: "Masticore", o704CinderSeer: "Cinder Seer",
		o704Impatience: "Impatience", o704Caltrops: "Caltrops", o704GoblinFestval: "Goblin Festival",
		o704Landslide: "Landslide", o704ScentOfCinder: "Scent of Cinder",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestS82OldB704AetherStingHitsOpponentsCastingCreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Aether Sting", "Enchantment", o704AetherSting, false)
	mine := me.Life
	effectstest.B14CastAt(t, g, me, "My Bear", "Creature — Bear", "", nil, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != mine {
		t.Errorf("the controller's own creature spell cost them life: %d", me.Life)
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	theirs := opp.Life
	effectstest.B14CastAt(t, g, opp, "Their Sorcery", "Sorcery", "", nil, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirs {
		t.Errorf("a noncreature spell hurt the opponent: %d", opp.Life)
	}
	effectstest.B14CastAt(t, g, opp, "Their Bear", "Creature — Bear", "", nil, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirs-1 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, theirs-1)
	}
}

func TestS82OldB704MasticoreUpkeepDiscardAndAbilities(t *testing.T) {
	for _, tc := range []struct {
		name          string
		hand          int
		discard, kept bool
	}{{"discard", 1, true, true}, {"decline", 1, false, false}, {"empty hand", 0, false, false}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		g.WithWriteLock(func() { me.Hand.Cards = nil })
		var card uuid.UUID
		if tc.hand > 0 {
			card = effectstest.B13HandCard(me, "Spare", "Sorcery", "")
		}
		mast := effectstest.B12Push(g, me.ID, "Masticore", "Artifact Creature — Masticore", o704Masticore, 4, 4)
		effectstest.AdvanceToUpkeepOf(t, g, 0)
		effectstest.PassPriorityAroundTable(t, g)
		if tc.hand > 0 {
			if tc.discard {
				effectstest.AnswerDiscard(t, g, me.ID, card)
			} else {
				effectstest.AnswerDiscard(t, g, me.ID)
			}
			effectstest.PassPriorityAroundTable(t, g)
		}
		if got := g.Battlefield.Contains(mast); got != tc.kept {
			t.Errorf("%s: the Masticore on the battlefield = %v, want %v", tc.name, got, tc.kept)
		}
	}
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mast := effectstest.B12Push(g, me.ID, "Masticore", "Artifact Creature — Masticore", o704Masticore, 4, 4)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AddPool(t, g, me.ID, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, mast, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("activate the damage ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, bear); got != 1 {
		t.Errorf("the target has %d damage, want 1", got)
	}
	effectstest.AddPool(t, g, me.ID, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, mast, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate regenerate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B14Kill(g, mast)
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(mast) {
		t.Error("the regeneration shield did not save the Masticore")
	}
}

// o704RedHand puts three red cards and a green one into the seat's hand and
// returns the red ones and the green one.
func o704RedHand(p *game.Player) (red []uuid.UUID, green uuid.UUID) {
	for i := 0; i < 3; i++ {
		id := uuid.New()
		p.Hand.PushTop(game.Card{InstanceID: id, Name: "Red Card", TypeLine: "Sorcery", Colors: []string{"R"}, Owner: p.ID, Controller: p.ID})
		red = append(red, id)
	}
	green = uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: green, Name: "Green Card", TypeLine: "Sorcery", Colors: []string{"G"}, Owner: p.ID, Controller: p.ID})
	return red, green
}

func o704AnswerReveal(t *testing.T, g *game.Game, me *game.Player, red []uuid.UUID, green uuid.UUID, pick int) {
	t.Helper()
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("the controller was not asked which red cards to reveal")
	}
	if effectstest.HasID(prompt.ChooseCards, green) || len(prompt.ChooseCards) != len(red) {
		t.Errorf("the offered cards are %v, want only the three red ones", prompt.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, red[:pick]...)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestS82OldB704CinderSeerDealsDamageForEachRevealedRedCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	seer := effectstest.B12Push(g, me.ID, "Cinder Seer", "Creature — Human Wizard", o704CinderSeer, 1, 1)
	red, green := o704RedHand(me)
	effectstest.AddPool(t, g, me.ID, "{C}{C}{R}")
	if err := g.ActivateCatalogAbility(me.ID, seer, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	life := opp.Life
	o704AnswerReveal(t, g, me, red, green, 2)
	if opp.Life != life-2 {
		t.Errorf("the target is at %d, want %d (two cards revealed)", opp.Life, life-2)
	}
	if !effectstest.TappedForTest(t, g, seer) {
		t.Error("the ability did not tap the Seer")
	}
}

func TestS82OldB704ScentOfCinderDealsDamageForEachRevealedRedCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	red, green := o704RedHand(me)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Scent of Cinder", "Sorcery", o704ScentOfCinder, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	o704AnswerReveal(t, g, me, red, green, 3)
	if opp.Life != life-3 {
		t.Errorf("the target is at %d, want %d (three cards revealed)", opp.Life, life-3)
	}
}

func TestS82OldB704ScentOfCinderRevealingNothingDealsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	red, green := o704RedHand(me)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Scent of Cinder", "Sorcery", o704ScentOfCinder, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	o704AnswerReveal(t, g, me, red, green, 0)
	if opp.Life != life {
		t.Errorf("revealing nothing dealt damage: %d, want %d", opp.Life, life)
	}
}

func TestS82OldB704ImpatienceHurtsAPlayerWhoCastNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Impatience", "Enchantment", o704Impatience, false)
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.B14CastAt(t, g, opp, "Their Sorcery", "Sorcery", "", nil, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	theirs := opp.Life
	effectstest.AdvanceToEndStepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirs {
		t.Errorf("a player who cast a spell took damage: %d, want %d", opp.Life, theirs)
	}
	// The next player cast nothing this turn and is dealt 2.
	next := g.Seats[2]
	nextLife := next.Life
	effectstest.AdvanceToEndStepOf(t, g, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if next.Life != nextLife-2 {
		t.Errorf("a player who cast nothing is at %d, want %d", next.Life, nextLife-2)
	}
}

func TestS82OldB704CaltropsDamagesEachAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Caltrops", "Artifact", o704Caltrops, false)
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Ogre", 3, 3)
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Rat", 1, 1)
	effectstest.BrAttack(t, g, big, small)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, big); got != 1 {
		t.Errorf("the 3/3 attacker has %d damage, want 1", got)
	}
	if g.Battlefield.Contains(small) {
		t.Error("the 1/1 attacker survived 1 damage")
	}
}

func TestS82OldB704GoblinFestivalPingsAndMayChangeHands(t *testing.T) {
	lost, won := 0, 0
	for table := 0; table < 40 && (lost == 0 || won == 0); table++ {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		fest := effectstest.PushCatalogPermanent(g, me.ID, "Goblin Festival", "Enchantment", o704GoblinFestval, false)
		effectstest.AddPool(t, g, me.ID, "{C}{C}")
		life := opp.Life
		if err := g.ActivateCatalogAbility(me.ID, fest, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if opp.Life != life-1 {
			t.Fatalf("the target is at %d, want %d", opp.Life, life-1)
		}
		ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
		if ask == nil {
			t.Fatal("no coin call was asked")
		}
		flips := len(effectstest.RandomEvents(g, game.EventFlipCoin))
		if err := g.ResolveCoinCall(ask.ID, me.ID, "heads"); err != nil {
			t.Fatalf("ResolveCoinCall: %v", err)
		}
		evs := effectstest.RandomEvents(g, game.EventFlipCoin)
		if len(evs) != flips+1 {
			t.Fatalf("%d coins were flipped, want 1", len(evs)-flips)
		}
		if evs[len(evs)-1].Won {
			won++
			if c, _ := effectstest.BattlefieldCard(g, fest); c.Controller != me.ID {
				t.Error("won the flip, but the Festival changed hands")
			}
			continue
		}
		lost++
		chosen := g.Seats[2]
		effectstest.AnswerChoosePlayer(t, g, me.ID, chosen)
		effectstest.PassPriorityAroundTable(t, g)
		if c, _ := effectstest.BattlefieldCard(g, fest); c.Controller != chosen.ID {
			t.Errorf("lost the flip: the Festival is controlled by %v, want the chosen opponent", c.Controller)
		}
	}
	if lost == 0 || won == 0 {
		t.Skipf("never saw both flip outcomes in 40 tables (won %d, lost %d)", won, lost)
	}
}

func TestS82OldB704LandslideSacrificesMountainsForThatMuchDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	var mountains []uuid.UUID
	for i := 0; i < 3; i++ {
		mountains = append(mountains, effectstest.B12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain"))
	}
	forest := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	theirs := effectstest.B12Permanent(g, opp.ID, "Mountain", "Basic Land — Mountain")
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Landslide", "Sorcery", o704Landslide, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if p == nil {
		t.Fatal("Landslide queued no sacrifice pick")
	}
	if effectstest.HasID(p.ChooseCards, forest) || effectstest.HasID(p.ChooseCards, theirs) || len(p.ChooseCards) != 3 {
		t.Errorf("the offered permanents are %v, want exactly your three Mountains", p.ChooseCards)
	}
	if p.ChooseMin != 0 {
		t.Errorf("any number: min %d, want 0", p.ChooseMin)
	}
	if err := g.ResolveOwnPermanents(p.ID, me.ID, mountains[:2]); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mountains[0]) || g.Battlefield.Contains(mountains[1]) || !g.Battlefield.Contains(mountains[2]) {
		t.Error("exactly the two chosen Mountains should be sacrificed")
	}
	if opp.Life != life-2 {
		t.Errorf("the target is at %d, want %d (two Mountains sacrificed)", opp.Life, life-2)
	}
}
