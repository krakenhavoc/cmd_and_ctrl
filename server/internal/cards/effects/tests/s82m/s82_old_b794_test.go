package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b794_test.go — Mirage, batch 794 (#5633). Every test plays the real
// catalog entry through the engine.

const (
	oldB794Minion   = "5cf1d7c3-6ebe-440b-88c8-33b7f6aa0d7c"
	oldB794Cage     = "72a1e724-eb32-4781-b44d-3b1ed02a5a8b"
	oldB794Hex      = "7c69cb2b-14f1-4215-aba8-c107493c3036"
	oldB794Bane     = "814adb99-6100-4c36-96d9-763bc20dca5d"
	oldB794Caliph   = "8434bdb1-286b-484e-904e-946d074614e7"
	oldB794Manti    = "9360d30c-4556-4b41-bd50-08b4367e5efb"
	oldB794Flame    = "975226a4-2d03-4993-9337-cfd6595011bd"
	oldB794Storm    = "9b3b0bef-cdbb-41c2-b8f1-ddda9903b43a"
	oldB794Sting    = "a500313b-35e3-4ebc-9144-e9486784757b"
	oldB794Retribut = "b08b348c-cf74-4aea-ab88-38201e04e42c"
	oldB794Unicorn  = "b437c963-def5-40d6-b567-ae9f1d2a0fa5"
)

func TestOldB794EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB794Minion: "Shauku's Minion", oldB794Cage: "Misers' Cage", oldB794Hex: "Kaervek's Hex",
		oldB794Bane: "Builder's Bane", oldB794Caliph: "Emberwilde Caliph", oldB794Manti: "Sawback Manticore",
		oldB794Flame: "Flame Elemental", oldB794Storm: "Tropical Storm", oldB794Sting: "Unyaro Bee Sting",
		oldB794Retribut: "Divine Retribution", oldB794Unicorn: "Benevolent Unicorn",
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

func oldB794Card(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func oldB794Colored(g *game.Game, owner uuid.UUID, name string, p, tough int, colors []string, kws ...string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name,
		TypeLine: "Creature — Test", Power: p, Toughness: tough, Colors: colors, Keywords: kws, Owner: owner, Controller: owner})
}

// oldB794Cast casts a catalogued spell from the active seat's hand and resolves it.
func oldB794Cast(t *testing.T, g *game.Game, name, typeLine, oracle string, x int, targets []game.TargetRef) {
	t.Helper()
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	c := game.NewCard(name, me.ID)
	c.TypeLine, c.OracleID = typeLine, oracle
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Targets: targets, XValue: x}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB794ShaukusMinionHurtsWhiteCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	m := effectstest.BlbB1Push(g, me.ID, "Shauku's Minion", "Creature — Human Minion", oldB794Minion, 2, 2)
	white := oldB794Colored(g, opp.ID, "Knight", 2, 3, []string{"W"})
	green := oldB794Colored(g, opp.ID, "Bear", 2, 3, []string{"G"})
	effectstest.FloatMana(t, g, me, "{B}{R}")
	if err := g.ActivateCatalogAbility(me.ID, m, 0, game.ActivateAbilityParams{Strict: true, Targets: oldB794Card(green)}); err == nil {
		t.Fatal("targeted a green creature")
	}
	effectstest.B16Activate(t, g, me.ID, m, 0, game.ActivateAbilityParams{Strict: true, Targets: oldB794Card(white)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageOn(t, g, white); got != 2 {
		t.Errorf("the white creature has %d damage, want 2", got)
	}
}

func TestOldB794MisersCageNeedsFiveCards(t *testing.T) {
	for _, tc := range []struct {
		name string
		hand int
		want int
	}{{"five cards", 5, 2}, {"four cards", 4, 0}, {"seven cards", 7, 2}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			effectstest.PushCatalogPermanent(g, me.ID, "Misers' Cage", "Artifact", oldB794Cage, false)
			opp.Hand.Cards = nil
			for i := 0; i < tc.hand; i++ {
				effectstest.PushCatalogHandCard(opp, "Filler", "Sorcery", "")
			}
			before := effectstest.LifeOf(g, opp.ID)
			effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.LifeOf(g, opp.ID); got != before-tc.want {
				t.Errorf("life %d, want %d", got, before-tc.want)
			}
		})
	}
}

func TestOldB794KaerveksHexDamagesNonblackAndGreenExtra(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	black := oldB794Colored(g, me.ID, "Zombie", 2, 5, []string{"B"})
	red := oldB794Colored(g, me.ID, "Goblin", 2, 5, []string{"R"})
	green := oldB794Colored(g, me.ID, "Bear", 2, 5, []string{"G"})
	bg := oldB794Colored(g, me.ID, "Dryad", 2, 5, []string{"B", "G"})
	oldB794Cast(t, g, "Kaervek's Hex", "Sorcery", oldB794Hex, 0, nil)
	for _, tc := range []struct {
		name string
		id   uuid.UUID
		want int
	}{{"black", black, 0}, {"red", red, 1}, {"green", green, 2}, {"black-green", bg, 1}} {
		if got := effectstest.DamageOn(t, g, tc.id); got != tc.want {
			t.Errorf("%s creature has %d damage, want %d", tc.name, got, tc.want)
		}
	}
}

func TestOldB794TropicalStormHitsFlyersForXAndBlueForOneMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ground := oldB794Colored(g, me.ID, "Bear", 2, 9, []string{"G"})
	flyer := oldB794Colored(g, me.ID, "Bird", 2, 9, []string{"W"}, "flying")
	blueFlyer := oldB794Colored(g, me.ID, "Drake", 2, 9, []string{"U"}, "flying")
	blueGround := oldB794Colored(g, me.ID, "Merfolk", 2, 9, []string{"U"})
	oldB794Cast(t, g, "Tropical Storm", "Sorcery", oldB794Storm, 3, nil)
	for _, tc := range []struct {
		name string
		id   uuid.UUID
		want int
	}{{"ground", ground, 0}, {"flyer", flyer, 3}, {"blue flyer", blueFlyer, 4}, {"blue ground", blueGround, 1}} {
		if got := effectstest.DamageOn(t, g, tc.id); got != tc.want {
			t.Errorf("%s creature has %d damage, want %d", tc.name, got, tc.want)
		}
	}
}

func TestOldB794BuildersBaneDestroysXArtifactsAndHurtsTheirControllers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a1 := effectstest.PushCatalogPermanent(g, opp.ID, "Rock", "Artifact", "", false)
	a2 := effectstest.PushCatalogPermanent(g, opp.ID, "Pebble", "Artifact", "", false)
	a3 := effectstest.PushCatalogPermanent(g, me.ID, "Mine", "Artifact", "", false)
	keep := effectstest.PushCatalogPermanent(g, opp.ID, "Bystander", "Artifact", "", false)
	myLife, oppLife := effectstest.LifeOf(g, me.ID), effectstest.LifeOf(g, opp.ID)
	oldB794Cast(t, g, "Builder's Bane", "Sorcery", oldB794Bane, 3, []game.TargetRef{
		{Kind: game.TargetCard, ID: a1}, {Kind: game.TargetCard, ID: a2}, {Kind: game.TargetCard, ID: a3}})
	for _, id := range []uuid.UUID{a1, a2, a3} {
		if g.Battlefield.Contains(id) {
			t.Errorf("artifact %s survived", id)
		}
	}
	if !g.Battlefield.Contains(keep) {
		t.Error("the untargeted artifact was destroyed")
	}
	if got := effectstest.LifeOf(g, opp.ID); got != oppLife-2 {
		t.Errorf("opponent life %d, want %d", got, oppLife-2)
	}
	if got := effectstest.LifeOf(g, me.ID); got != myLife-1 {
		t.Errorf("your life %d, want %d", got, myLife-1)
	}
}

func TestOldB794BuildersBaneRefusesAnUntargetedCountOfX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a1 := effectstest.PushCatalogPermanent(g, opp.ID, "Rock", "Artifact", "", false)
	c := game.NewCard("Builder's Bane", me.ID)
	c.TypeLine, c.OracleID = "Sorcery", oldB794Bane
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Strict: true, XValue: 2, Targets: oldB794Card(a1)}); err == nil {
		t.Error("cast with X=2 and one target")
	}
}

func TestOldB794EmberwildeCaliphMakesYouLoseLifeForDamageItDeals(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	c := effectstest.BlbB1Push(g, me.ID, "Emberwilde Caliph", "Creature — Djinn", oldB794Caliph, 4, 4)
	before := effectstest.LifeOf(g, me.ID)
	effectstest.P2Ping(t, g, c, opp.ID, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != before-4 {
		t.Errorf("life %d, want %d", got, before-4)
	}
}

func TestOldB794SawbackManticoreShootsOnlyWhileInCombatAndOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	m := effectstest.BlbB1Push(g, me.ID, "Sawback Manticore", "Creature — Manticore", oldB794Manti, 2, 4)
	other := oldB794Colored(g, me.ID, "Bear", 2, 5, []string{"G"})
	effectstest.FloatMana(t, g, me, "{R}{R}")
	if err := g.ActivateCatalogAbility(me.ID, m, 1, game.ActivateAbilityParams{Strict: true, Targets: oldB794Card(other)}); err == nil {
		t.Fatal("activated while neither attacking nor blocking")
	}
	effectstest.Pr7bAttack(t, g, opp.ID, m, other)
	effectstest.FloatMana(t, g, me, "{R}{R}")
	effectstest.B16Activate(t, g, me.ID, m, 1, game.ActivateAbilityParams{Strict: true, Targets: oldB794Card(other)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageOn(t, g, other); got != 2 {
		t.Errorf("the target has %d damage, want 2", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, m, 1, game.ActivateAbilityParams{Strict: true, Targets: oldB794Card(other)}); err == nil {
		t.Error("activated a second time this turn")
	}
}

func TestOldB794SawbackManticoreGainsFlyingForFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	m := effectstest.BlbB1Push(g, me.ID, "Sawback Manticore", "Creature — Manticore", oldB794Manti, 2, 4)
	effectstest.FloatMana(t, g, me, "{R}{R}{R}{R}")
	effectstest.B16Activate(t, g, me.ID, m, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.P1g2HasKeyword(g, m, "flying") {
		t.Error("no flying")
	}
}

func TestOldB794FlameElementalSacrificesToDealItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	f := effectstest.BlbB1Push(g, me.ID, "Flame Elemental", "Creature — Elemental", oldB794Flame, 3, 2)
	victim := oldB794Colored(g, opp.ID, "Bear", 2, 3, []string{"G"})
	effectstest.FloatMana(t, g, me, "{R}")
	effectstest.B16Activate(t, g, me.ID, f, 0, game.ActivateAbilityParams{Strict: true, Targets: oldB794Card(victim)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(f) {
		t.Error("the Elemental was not sacrificed")
	}
	if g.Battlefield.Contains(victim) {
		t.Error("the target survived 3 damage")
	}
}

func TestOldB794UnyaroBeeStingDealsTwoToAnyTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	before := effectstest.LifeOf(g, opp.ID)
	oldB794Cast(t, g, "Unyaro Bee Sting", "Sorcery", oldB794Sting, 0, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	if got := effectstest.LifeOf(g, opp.ID); got != before-2 {
		t.Errorf("life %d, want %d", got, before-2)
	}
}

func TestOldB794DivineRetributionDealsDamageEqualToTheAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a := oldB794Colored(g, me.ID, "Ogre", 2, 9, []string{"R"})
	b := oldB794Colored(g, me.ID, "Goblin", 2, 9, []string{"R"})
	c := oldB794Colored(g, me.ID, "Imp", 2, 9, []string{"B"})
	effectstest.Pr7bAttack(t, g, opp.ID, a, b, c)
	card := game.NewCard("Divine Retribution", me.ID)
	card.TypeLine, card.OracleID = "Instant", oldB794Retribut
	me.Hand.PushTop(card)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Strict: true, Targets: oldB794Card(a)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageOn(t, g, a); got != 3 {
		t.Errorf("the target has %d damage, want 3", got)
	}
}

func TestOldB794DivineRetributionRefusesANonAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	idle := oldB794Colored(g, opp.ID, "Bear", 2, 2, []string{"G"})
	card := game.NewCard("Divine Retribution", me.ID)
	card.TypeLine, card.OracleID = "Instant", oldB794Retribut
	me.Hand.PushTop(card)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Strict: true, Targets: oldB794Card(idle)}); err == nil {
		t.Error("targeted a creature that is not attacking")
	}
}

func TestOldB794BenevolentUnicornTakesOneOffEverySpellsDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.BlbB1Push(g, me.ID, "Benevolent Unicorn", "Creature — Unicorn", oldB794Unicorn, 1, 2)
	before := effectstest.LifeOf(g, opp.ID)
	oldB794Cast(t, g, "Unyaro Bee Sting", "Sorcery", oldB794Sting, 0, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	if got := effectstest.LifeOf(g, opp.ID); got != before-1 {
		t.Errorf("the spell dealt %d, want 1", before-got)
	}
}

func TestOldB794BenevolentUnicornLeavesAbilityDamageAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.BlbB1Push(g, me.ID, "Benevolent Unicorn", "Creature — Unicorn", oldB794Unicorn, 1, 2)
	src := effectstest.BlbB1Push(g, me.ID, "Pinger", "Creature — Test", "", 1, 1)
	before := effectstest.LifeOf(g, opp.ID)
	effectstest.P2Ping(t, g, src, opp.ID, 2)
	if got := effectstest.LifeOf(g, opp.ID); got != before-2 {
		t.Errorf("a creature's damage was %d, want 2", before-got)
	}
}
