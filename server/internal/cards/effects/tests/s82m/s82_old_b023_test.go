package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB023Gust     = "011d4f47-c6fd-434b-98f2-2f558fbff40b"
	oldB023Despair  = "0f2d038e-6dab-4352-9ba7-ed9e063bb304"
	oldB023Reaper   = "70034860-5198-421f-871d-7c1676337b6e"
	oldB023Kogla    = "8b21062e-97b0-4967-96df-30e8309f4fba"
	oldB023Leveler  = "ba1d34ab-7529-4f47-a5e7-7d8f66d7b901"
	oldB023Paladin  = "f5975294-508a-453e-893a-2fbea2487d17"
	oldB023Tsabo    = "f98f1f71-99ec-4ffb-9d8e-acbab173f972"
	oldB023Scarecrw = "Creature — Scarecrow"
)

func TestOldB023CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB023Gust:    "Fracturing Gust",
		oldB023Despair: "Force of Despair",
		oldB023Reaper:  "Reaper King",
		oldB023Kogla:   "Kogla, the Titan Ape",
		oldB023Leveler: "Leveler",
		oldB023Paladin: "Northern Paladin",
		oldB023Tsabo:   "Tsabo Tavoc",
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

func TestOldB023FracturingGustDestroysArtifactsAndEnchantmentsAndGainsLifeForThoseDestroyed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	aura := effectstest.B12Permanent(g, opp.ID, "Aura", "Enchantment")
	creature := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	sturdy := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Sturdy Rock", TypeLine: "Artifact", Keywords: []string{"indestructible"}})
	effectstest.ToMainForCost(t, g)
	life := me.Life
	id := effectstest.HandCardFull(me, "Fracturing Gust", "Instant", "{2}{G/W}{G/W}{G/W}", oldB023Gust, []string{"G", "W"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, gone := range []uuid.UUID{rock, aura} {
		if g.Battlefield.Contains(gone) {
			t.Errorf("%s survived", gone)
		}
	}
	for name, kept := range map[string]uuid.UUID{"the creature": creature, "the land": land, "the indestructible artifact": sturdy} {
		if !g.Battlefield.Contains(kept) {
			t.Errorf("%s was destroyed", name)
		}
	}
	if got := me.Life - life; got != 4 {
		t.Errorf("gained %d life, want 4: two permanents were destroyed (not the indestructible one)", got)
	}
}

func TestOldB023ForceOfDespairDestroysOnlyCreaturesThatEnteredThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// Put straight on the battlefield with no entry event, so it is not in
	// this turn's tally of what entered.
	old := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: old, Name: "Old Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	effectstest.ToMainForCost(t, g)
	fresh := effectstest.TrekB31Enter(t, g, opp, game.Card{Name: "Fresh Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	mine := effectstest.TrekB31Enter(t, g, me, game.Card{Name: "Fresh Own Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	id := effectstest.HandCardFull(me, "Force of Despair", "Instant", "{1}{B}{B}", oldB023Despair, []string{"B"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fresh) || g.Battlefield.Contains(mine) {
		t.Error("a creature that entered this turn survived")
	}
	if !g.Battlefield.Contains(old) {
		t.Error("a creature from an earlier turn was destroyed")
	}
}

func TestOldB023ForceOfDespairPitchesABlackCardOnlyOffYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	active := g.Seats[g.Turn.ActiveSeat]
	other := me
	if active.ID == me.ID {
		other = opp
	}
	black := effectstest.HandCardFull(other, "Black Card", "Instant", "{B}", "", []string{"B"})
	green := effectstest.HandCardFull(other, "Green Card", "Instant", "{G}", "", []string{"G"})
	cast := func(p *game.Player, pitch uuid.UUID) error {
		id := effectstest.HandCardFull(p, "Force of Despair", "Instant", "{1}{B}{B}", oldB023Despair, []string{"B"})
		return g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: "pitch", AltCostIDs: []uuid.UUID{pitch}})
	}
	if err := cast(active, effectstest.HandCardFull(active, "Own Black", "Instant", "{B}", "", []string{"B"})); err == nil {
		t.Fatal("pitching on your own turn was accepted")
	}
	if err := cast(other, green); err == nil {
		t.Error("a nonblack card was pitched")
	}
	if err := cast(other, black); err != nil {
		t.Fatalf("pitching a black card off your turn: %v", err)
	}
	if !g.Exile.Contains(black) || other.Hand.Contains(black) {
		t.Error("the black card was not exiled as the cost")
	}
}

func TestOldB023ReaperKingAnthemAndDestroyOnAnotherScarecrow(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	king := effectstest.B12Push(g, me.ID, "Reaper King", "Legendary Artifact Creature — Scarecrow", oldB023Reaper, 6, 6)
	mine := effectstest.B12Creature(g, me.ID, "Straw", oldB023Scarecrw, 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Straw", oldB023Scarecrw, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if got := effectstest.ApaLive(g, mine).CurrentPower(); got != 2 {
		t.Errorf("my other Scarecrow has power %d, want 2", got)
	}
	if got := effectstest.ApaLive(g, theirs).CurrentPower(); got != 1 {
		t.Errorf("an opponent's Scarecrow has power %d, want 1", got)
	}
	if got := effectstest.ApaLive(g, king).CurrentPower(); got != 6 {
		t.Errorf("the King has power %d, want 6: the bonus is for others", got)
	}
	if got := effectstest.ApaLive(g, bear).CurrentPower(); got != 2 {
		t.Errorf("a non-Scarecrow has power %d, want 2", got)
	}
	victim := effectstest.B12Permanent(g, opp.ID, "Doomed Rock", "Artifact")
	fresh := effectstest.HandCardFull(me, "New Straw", oldB023Scarecrw, "", "", nil)
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == fresh {
				me.Hand.Cards[i].Power, me.Hand.Cards[i].Toughness = 1, 1
			}
		}
	})
	if err := g.CastSpell(me.ID, fresh, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the targeted permanent survived")
	}
	if !g.Battlefield.Contains(fresh) {
		t.Error("the entering Scarecrow was not on the battlefield")
	}
}

func TestOldB023KoglaFightsOnEntryAndBouncesAHumanForIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	kogla := effectstest.CastGraveyardCopier(t, g, "Kogla, the Titan Ape", "Legendary Creature — Ape", oldB023Kogla, 7, 6)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the creature Kogla fought survived")
	}
	if c := effectstest.ApaLive(g, kogla); c.DamageMarked != 2 {
		t.Errorf("Kogla has %d damage, want 2 from the fight", c.DamageMarked)
	}
	human := effectstest.B12Creature(g, me.ID, "Farmer", "Creature — Human Peasant", 1, 1)
	notHuman := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.ApaMana(me, "G", "G")
	if err := g.ActivateCatalogAbility(me.ID, kogla, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: notHuman}}}); err == nil {
		t.Fatal("a non-Human was accepted")
	}
	effectstest.ApaActivate(t, g, me, kogla, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: human}}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(human) || !me.Hand.Contains(human) {
		t.Error("the Human was not returned to its owner's hand")
	}
	if !effectstest.HasEffectiveKeyword(t, g, kogla, "indestructible") {
		t.Error("Kogla did not gain indestructible")
	}
}

func TestOldB023KoglaAttackDestroysADefendersArtifactOrEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kogla := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Kogla, the Titan Ape", OracleID: oldB023Kogla,
		TypeLine: "Legendary Creature — Ape", Power: 7, Toughness: 6})
	theirRock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	otherRock := effectstest.B12Permanent(g, g.Seats[2].ID, "Third Party Rock", "Artifact")
	effectstest.AdvanceToMain(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, kogla)
	if p := effectstest.LatestPickTarget(g, me.ID); p == nil {
		t.Fatal("the attack trigger asked for no target")
	}
	if err := g.ResolvePickTarget(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: otherRock}); err == nil {
		t.Fatal("a permanent of a player who was not attacked was accepted")
	}
	effectstest.PickTriggerTarget(t, g, me.ID, theirRock)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirRock) {
		t.Error("the defender's artifact survived")
	}
	if !g.Battlefield.Contains(otherRock) {
		t.Error("another player's artifact was destroyed")
	}
}

func TestOldB023LevelerExilesYourWholeLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	for range 5 {
		effectstest.SeedLibraryTop(me, "Card", "Sorcery", "")
	}
	mine, theirs := me.Library.Size(), opp.Library.Size()
	if mine == 0 || theirs == 0 {
		t.Fatal("the libraries start empty")
	}
	exiled := g.Exile.Size()
	effectstest.CastGraveyardCopier(t, g, "Leveler", "Artifact Creature — Juggernaut", oldB023Leveler, 10, 10)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != 0 {
		t.Errorf("%d cards remain in the library, want 0", me.Library.Size())
	}
	if opp.Library.Size() != theirs {
		t.Errorf("an opponent's library changed: %d -> %d", theirs, opp.Library.Size())
	}
	if got := g.Exile.Size() - exiled; got != mine {
		t.Errorf("%d cards were exiled, want the %d from the library", got, mine)
	}
	if me.Eliminated {
		t.Error("the player was eliminated by exiling their library")
	}
}

func TestOldB023NorthernPaladinDestroysABlackPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	paladin := effectstest.B12Push(g, me.ID, "Northern Paladin", "Creature — Human Knight", oldB023Paladin, 3, 3)
	black := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Dark Rock", TypeLine: "Artifact", Colors: []string{"B"}})
	white := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Bright Rock", TypeLine: "Artifact", Colors: []string{"W"}})
	effectstest.ApaMana(me, "W", "W")
	if err := g.ActivateCatalogAbility(me.ID, paladin, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: white}}}); err == nil {
		t.Fatal("a nonblack permanent was accepted")
	}
	effectstest.ApaActivate(t, g, me, paladin, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: black}}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(black) {
		t.Error("the black permanent survived")
	}
	if !g.Battlefield.Contains(white) {
		t.Error("the white permanent was destroyed")
	}
}

func TestOldB023TsaboTavocDestroysALegendaryCreatureAndHasProtectionFromThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	tsabo := effectstest.B12Push(g, me.ID, "Tsabo Tavoc", "Legendary Creature — Phyrexian Horror", oldB023Tsabo, 7, 4)
	legend := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Legend", TypeLine: "Legendary Creature — Human", Power: 2, Toughness: 2})
	plain := effectstest.B12Creature(g, opp.ID, "Plain Bear", "Creature — Bear", 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, tsabo, "first strike") {
		t.Error("Tsabo has no first strike")
	}
	found := false
	for _, q := range game.ProtectionQualities(effectstest.ApaLive(g, tsabo)) {
		if q.Kind == game.ProtectionQualityLegendaryCreatures {
			found = true
		}
	}
	if !found {
		t.Error("Tsabo has no protection from legendary creatures")
	}
	effectstest.ApaMana(me, "B", "B")
	if err := g.ActivateCatalogAbility(me.ID, tsabo, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: plain}}}); err == nil {
		t.Fatal("a nonlegendary creature was accepted")
	}
	effectstest.ApaActivate(t, g, me, tsabo, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: legend}}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(legend) {
		t.Error("the legendary creature survived")
	}
}
