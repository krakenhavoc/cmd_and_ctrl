package testss82m

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Theros Beyond Death batch 133 (S82 OLD, #5514).

const (
	oldB133Lion      = "0db3b29e-e21c-4240-8e5e-959ba6bcf706"
	oldB133Singer    = "2c0816bd-8ce2-4ddf-845a-bcdaf929fe28"
	oldB133Catoblep  = "4098ab29-bf90-4ec9-b547-05d304fc7798"
	oldB133Huntmstr  = "5ce273f2-a865-4934-afd4-cf6d5c4f1a4f"
	oldB133Berserker = "658d54d4-e967-490c-b958-3743b3b1ef45"
	oldB133Haktos    = "932a86e5-9640-4ef5-8264-2ce97bde33e2"
)

func TestOldB133CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB133Lion: "Bronzehide Lion", oldB133Singer: "Threnody Singer", oldB133Catoblep: "Blight-Breath Catoblepas",
		oldB133Huntmstr: "Nylea's Huntmaster", oldB133Berserker: "Rage-Scarred Berserker", oldB133Haktos: "Haktos the Unscarred",
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

// oldB133Cast casts a creature with a real mana cost and answers its entry
// trigger's target prompt with `pick`.
func oldB133Cast(t *testing.T, g *game.Game, name, typeLine, cost, oracle string, power, toughness int, pick uuid.UUID) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.RfdHandCard(me, name, typeLine, cost, oracle, power, toughness)
	effectstest.ToMainForCost(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 4; i++ {
		if effectstest.LatestPickTarget(g, me.ID) == nil {
			break
		}
		effectstest.PickCard(t, g, me.ID, pick)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB133ThrenodySingerShrinksByDevotionToBlue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Blue Thing", TypeLine: "Creature — Wizard", ManaCost: "{U}{U}", Power: 1, Toughness: 1})
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Giant", 5, 5)
	oldB133Cast(t, g, "Threnody Singer", "Creature — Siren", "{1}{U}", oldB133Singer, 1, 3, victim)
	// Devotion to blue: {U}{U} and the Singer's {U} = 3.
	if p, tt := effectstest.PtOf(t, g, victim); p != 2 || tt != 5 {
		t.Errorf("victim is %d/%d, want 2/5 (-3/-0)", p, tt)
	}
}

func TestOldB133CatoblepasShrinksByDevotionToBlack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Giant", 4, 4)
	oldB133Cast(t, g, "Blight-Breath Catoblepas", "Creature — Beast", "{4}{B}{B}", oldB133Catoblep, 3, 2, victim)
	_ = me
	// Only the Catoblepas's own {B}{B}: X = 2.
	if p, tt := effectstest.PtOf(t, g, victim); p != 2 || tt != 2 {
		t.Errorf("victim is %d/%d, want 2/2 (-2/-2)", p, tt)
	}
}

func TestOldB133HuntmasterPumpsByDevotionToGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ally := effectstest.B12Creature(g, me.ID, "Ally", "Creature — Bear", 2, 2)
	oldB133Cast(t, g, "Nylea's Huntmaster", "Creature — Centaur Shaman", "{3}{G}", oldB133Huntmstr, 4, 3, ally)
	if p, tt := effectstest.PtOf(t, g, ally); p != 3 || tt != 2 {
		t.Errorf("ally is %d/%d, want 3/2 (+1/+0 for one {G})", p, tt)
	}
}

func TestOldB133BerserkerGivesPlusOneAndIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ally := effectstest.B12Creature(g, me.ID, "Ally", "Creature — Bear", 2, 2)
	oldB133Cast(t, g, "Rage-Scarred Berserker", "Creature — Minotaur Berserker", "{4}{B}", oldB133Berserker, 5, 4, ally)
	if p, tt := effectstest.PtOf(t, g, ally); p != 3 || tt != 2 {
		t.Errorf("ally is %d/%d, want 3/2", p, tt)
	}
	c := effectstest.GaLayered(t, g, ally)
	if !game.HasKeyword(&c, "indestructible") {
		t.Error("ally did not gain indestructible")
	}
}

func TestOldB133BronzehideLionReturnsAsAnAuraThatProtectsItsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	lion := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Bronzehide Lion", OracleID: oldB133Lion,
		TypeLine: "Creature — Cat", ManaCost: "{G}{W}", Colors: []string{"G", "W"}, Power: 3, Toughness: 3})
	effectstest.RadDestroy(t, g, lion)
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 3 && effectstest.LatestPickTarget(g, me.ID) != nil; i++ {
		effectstest.PickCard(t, g, me.ID, bear)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.GaLayered(t, g, lion)
	if effectstest.AdZone(g, lion) != game.ZoneBattlefield || !c.IsAttachedTo(bear) || !c.IsAura() || c.IsCreature() {
		t.Fatalf("the Lion did not return as an Aura on the Bear: zone %v attached %+v", effectstest.AdZone(g, lion), c.AttachedTo)
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{G}{W}")
	var err error
	for idx := 0; idx < 3; idx++ {
		err = g.ActivateCatalogAbility(me.ID, lion, idx, game.ActivateAbilityParams{Strict: true})
		if err == nil {
			break
		}
	}
	if err != nil {
		t.Fatalf("activate the Aura's ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	b := effectstest.GaLayered(t, g, bear)
	if !game.HasKeyword(&b, "indestructible") {
		t.Error("the enchanted Bear did not gain indestructible")
	}
}

func TestOldB133BronzehideLionStaysDeadWithNoCreatureToEnchant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lion := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Bronzehide Lion", OracleID: oldB133Lion,
		TypeLine: "Creature — Cat", ManaCost: "{G}{W}", Power: 3, Toughness: 3})
	effectstest.RadDestroy(t, g, lion)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.AdZone(g, lion) != game.ZoneGraveyard {
		t.Errorf("the Lion is in zone %v, want the graveyard (CR 303.4g)", effectstest.AdZone(g, lion))
	}
}

func TestOldB133HaktosChoosesTwoThreeOrFourAsItEntersAndHasProtectionFromTheRest(t *testing.T) {
	for i := 0; i < 6; i++ {
		g := effectstest.NewCatalogGame(t)
		id := effectstest.RfdCast(t, g, "Haktos the Unscarred", "Legendary Creature — Human Warrior", oldB133Haktos, 6, 1)
		var n int
		g.WithWriteLock(func() { n = g.ChosenNumberOf(id) })
		if n < 2 || n > 4 {
			t.Fatalf("chosen number %d, want 2, 3 or 4", n)
		}
		c := effectstest.GaLayered(t, g, id)
		for _, kw := range []string{fmt.Sprintf("protection from mana value %d or less", n-1), fmt.Sprintf("protection from mana value %d or greater", n+1)} {
			if !game.HasKeyword(&c, kw) {
				t.Errorf("Haktos (chose %d) lacks %q", n, kw)
			}
		}
	}
}
