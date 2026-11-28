package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b721_cards_test.go — Urza's Saga enchantments, S82 OLD batch 721
// (#5624). Every test plays the real catalog entry through the engine.

const (
	b721Spider   = "b0709570-a145-4c3c-ba56-f06d25753855"
	b721Titan    = "ce06fe83-a8c5-46b7-b30c-b8189d217741"
	b721Ancients = "dc430e1b-d57f-4035-b518-9bc2e26a6c04"
	b721Presence = "e1aad679-93ec-420e-881b-35ebc99763a2"
	b721Caryatid = "eabcb1bd-5317-4e01-8f18-94f01c12c852"
	b721Gargoyle = "efef2d33-5354-4fce-880e-cb91fa60f730"
	b721Field    = "fbf7a7f4-9915-4a53-84b3-e55e9f487323"
	b721Veil     = "fc6278e5-7eb1-4590-880e-fe9efa0ec4d9"
)

func TestB721EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b721Spider: "Hidden Spider", b721Titan: "Opal Titan", b721Ancients: "Hidden Ancients",
		b721Presence: "Presence of the Master", b721Caryatid: "Opal Caryatid", b721Gargoyle: "Opal Gargoyle",
		b721Field: "Energy Field", b721Veil: "Veil of Birds",
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

// b721Spell casts a free spell from `p`'s hand and lets the stack settle.
func b721Spell(t *testing.T, g *game.Game, p *game.Player, typeLine string, colors []string, keywords ...string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: "Test Spell", TypeLine: typeLine, Colors: colors, Keywords: keywords,
		Power: 1, Toughness: 1, Owner: p.ID, Controller: p.ID})
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", typeLine, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

// b721Table is a game at the active seat's main phase: `me` casts, `owner`
// (another seat) controls the permanent under test.
func b721Table(t *testing.T, oracle, name string) (g *game.Game, me, owner *game.Player, perm uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, owner = g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	perm = effectstest.ApaPush(g, owner.ID, owner.ID, effectstest.StCard(name, oracle, "Enchantment", 0, 0))
	return
}

func b721IsStillEnchantment(g *game.Game, id uuid.UUID) bool {
	c := effectstest.ApaLive(g, id)
	return c != nil && c.HasCardType("enchantment") && !c.HasCardType("creature")
}

func b721RequireCreature(t *testing.T, g *game.Game, id uuid.UUID, subtype string, p, tough int) {
	t.Helper()
	c := effectstest.ApaLive(g, id)
	if c == nil || !c.HasCardType("creature") || c.HasCardType("enchantment") || !c.HasSubtype(subtype) {
		t.Fatalf("not a %s creature (and no longer an enchantment): %+v", subtype, c)
	}
	if gp, gt := effectstest.PtOf(t, g, id); gp != p || gt != tough {
		t.Errorf("it is %d/%d, want %d/%d", gp, gt, p, tough)
	}
}

func TestB721HiddenEnchantmentsWakeOnlyOnTheirOpponentCast(t *testing.T) {
	bear := func(t *testing.T, g *game.Game, p *game.Player) {
		b721Spell(t, g, p, "Creature — Bear", []string{"G"})
	}
	ench := func(t *testing.T, g *game.Game, p *game.Player) { b721Spell(t, g, p, "Enchantment", []string{"W"}) }
	rock := func(t *testing.T, g *game.Game, p *game.Player) { b721Spell(t, g, p, "Artifact", nil) }
	for _, tc := range []struct {
		name, oracle, subtype string
		p, tough              int
		keyword               string
		ignored, wakes        func(*testing.T, *game.Game, *game.Player)
	}{
		{"Opal Caryatid", b721Caryatid, "Soldier", 2, 2, "", rock, bear},
		{"Opal Gargoyle", b721Gargoyle, "Gargoyle", 2, 2, "flying", ench, bear},
		{"Hidden Ancients", b721Ancients, "Treefolk", 5, 5, "", bear, ench},
		{"Veil of Birds", b721Veil, "Bird", 1, 1, "flying", func(*testing.T, *game.Game, *game.Player) {}, rock},
		{"Hidden Spider", b721Spider, "Spider", 3, 5, "reach", bear,
			func(t *testing.T, g *game.Game, p *game.Player) {
				b721Spell(t, g, p, "Creature — Bird", []string{"W"}, "flying")
			}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _, perm := b721Table(t, tc.oracle, tc.name)
			// The caster's own copy is the "you don't wake it" control.
			mine := effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard(tc.name, tc.oracle, "Enchantment", 0, 0))
			tc.ignored(t, g, me)
			if !b721IsStillEnchantment(g, perm) {
				t.Fatalf("it woke on a spell it does not watch for")
			}
			tc.wakes(t, g, me)
			if b721IsStillEnchantment(g, perm) {
				t.Fatalf("an opponent's qualifying spell did not wake it")
			}
			b721RequireCreature(t, g, perm, tc.subtype, tc.p, tc.tough)
			if tc.keyword != "" && !effectstest.HasKeywordOnBattlefield(t, g, perm, tc.keyword) {
				t.Errorf("it lacks %s", tc.keyword)
			}
			if !b721IsStillEnchantment(g, mine) {
				t.Errorf("the caster's own copy woke on its own controller's spell")
			}
		})
	}
}

func TestB721OpalTitanGainsProtectionFromEachColorOfTheSpell(t *testing.T) {
	g, me, _, perm := b721Table(t, b721Titan, "Opal Titan")
	b721Spell(t, g, me, "Creature — Bear", []string{"R", "G"})
	b721RequireCreature(t, g, perm, "Giant", 4, 4)
	for _, c := range []string{"R", "G"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, perm, game.ProtectionFromColor(c)) {
			t.Errorf("no protection from %s", c)
		}
	}
	for _, c := range []string{"W", "U", "B"} {
		if effectstest.HasKeywordOnBattlefield(t, g, perm, game.ProtectionFromColor(c)) {
			t.Errorf("protection from %s, which the spell was not", c)
		}
	}
}

func TestB721OpalTitanGivesNoProtectionFromAColorlessSpell(t *testing.T) {
	g, me, _, perm := b721Table(t, b721Titan, "Opal Titan")
	b721Spell(t, g, me, "Artifact Creature — Golem", nil)
	b721RequireCreature(t, g, perm, "Giant", 4, 4)
	for _, c := range []string{"W", "U", "B", "R", "G"} {
		if effectstest.HasKeywordOnBattlefield(t, g, perm, game.ProtectionFromColor(c)) {
			t.Errorf("protection from %s off a colorless spell", c)
		}
	}
}

func TestB721OpalTitanIgnoresANoncreatureSpell(t *testing.T) {
	g, me, _, perm := b721Table(t, b721Titan, "Opal Titan")
	b721Spell(t, g, me, "Enchantment", []string{"W"})
	if !b721IsStillEnchantment(g, perm) {
		t.Fatal("a noncreature spell woke Opal Titan")
	}
}

func TestB721PresenceOfTheMasterCountersAnyPlayersEnchantmentSpell(t *testing.T) {
	for _, mine := range []bool{false, true} {
		g, me, _, perm := b721Table(t, b721Presence, "Presence of the Master")
		if mine {
			// Only the caster's own Presence: its controller's spell is countered too.
			g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(perm) })
			effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Presence of the Master", b721Presence, "Enchantment", 0, 0))
		}
		id := b721Spell(t, g, me, "Enchantment", []string{"W"})
		if g.Battlefield.Contains(id) || !me.Graveyard.Contains(id) {
			t.Errorf("mine=%v: the enchantment spell was not countered", mine)
		}
		other := b721Spell(t, g, me, "Creature — Bear", []string{"G"})
		if !g.Battlefield.Contains(other) {
			t.Errorf("mine=%v: a creature spell was countered", mine)
		}
	}
}

func TestB721EnergyFieldPreventsDamageFromSourcesYouDontControlOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Energy Field", b721Field, "Enchantment", 0, 0))
	theirs := effectstest.B12Creature(g, opp.ID, "Their Source", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Source", "Creature — Bear", 2, 2)
	life := me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(theirs, me.ID, 3) })
	if me.Life != life {
		t.Errorf("damage from an opponent's source went through: life %d, want %d", me.Life, life)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(mine, me.ID, 2) })
	if me.Life != life-2 {
		t.Errorf("damage from my own source: life %d, want %d", me.Life, life-2)
	}
	// It protects only its controller.
	oppLife := opp.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(theirs, opp.ID, 1) })
	if opp.Life != oppLife-1 {
		t.Errorf("the opponent was protected: life %d, want %d", opp.Life, oppLife-1)
	}
}

func TestB721EnergyFieldIsSacrificedWhenACardOfYoursHitsYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	field := effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Energy Field", b721Field, "Enchantment", 0, 0))
	// A token dying is not a card, and an opponent's card goes to their graveyard.
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Soldier", TypeLine: "Token Creature — Soldier",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, me.ID, token)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(field) {
		t.Fatal("a token, or a card that went to an opponent's graveyard, sacrificed Energy Field")
	}
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(field) {
		t.Error("Energy Field stayed after a card was put into its controller's graveyard")
	}
}

func TestB721EnergyFieldIsSacrificedWhenYouDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	field := effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Energy Field", b721Field, "Enchantment", 0, 0))
	effectstest.PutInHand(me, game.Card{InstanceID: uuid.New(), Name: "My Card", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(field) {
		t.Error("Energy Field stayed after its controller discarded a card")
	}
}
