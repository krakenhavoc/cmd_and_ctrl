package testsl

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestBatterskullMakesAGermAndAttachesItselfToIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	skull, germ := effectstest.CastLivingWeapon(t, g, "Batterskull", effectstest.BatterskullOracle)

	if host := effectstest.AttachmentHostOf(t, g, skull); host.Kind != game.TargetCard || host.ID != germ {
		t.Fatalf("Batterskull AttachedTo = %+v, want the Germ %s", host, germ)
	}
	// A 0/0 that survives is the whole point: the +4/+4 arrived in
	// the same resolution, so the CR 704.5f check never saw a 0/0.
	if got := effectstest.EffectivePower(t, g, germ); got != 4 {
		t.Errorf("Germ power %d, want 4", got)
	}
	if got := effectstest.EffectiveToughness(t, g, germ); got != 4 {
		t.Errorf("Germ toughness %d, want 4", got)
	}
	ab := effectstest.EffectiveAbilities(t, g, germ)
	for _, want := range []string{"vigilance", "lifelink"} {
		if !effectstest.ContainsString(ab, want) {
			t.Errorf("Germ abilities %v missing %q", ab, want)
		}
	}
}

// The Germ is a 0/0 and stays alive only while something is on it.
// Bouncing Batterskull with its own {3} is the printed way to prove
// that, and it exercises the bounce ability at the same time.
func TestBatterskullsBounceLeavesTheGermToDieAsA00(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	skull, germ := effectstest.CastLivingWeapon(t, g, "Batterskull", effectstest.BatterskullOracle)

	// Ability 0 is the bounce; ability 1 is equip.
	if err := g.ActivateCatalogAbility(me.ID, skull, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("the {3} bounce: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(skull) {
		t.Error("Batterskull is still on the battlefield after its own bounce")
	}
	if g.Battlefield.Contains(germ) {
		t.Error("the Germ survived with nothing attached to it — it is a 0/0")
	}
}

// Batterbone and Kaldra Compleat run the same helper and differ only
// in what they grant, which is what makes LivingWeapon worth sharing.
func TestBatterboneAndKaldraCompleatUseTheSameLivingWeapon(t *testing.T) {
	t.Run("Batterbone", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		bone, germ := effectstest.CastLivingWeapon(t, g, "Batterbone", effectstest.BatterboneOracle)
		if host := effectstest.AttachmentHostOf(t, g, bone); host.ID != germ {
			t.Fatalf("Batterbone AttachedTo = %+v, want the Germ", host)
		}
		if got := effectstest.EffectivePower(t, g, germ); got != 1 {
			t.Errorf("Germ power %d, want 1", got)
		}
	})

	t.Run("Kaldra Compleat", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		kaldra, germ := effectstest.CastLivingWeapon(t, g, "Kaldra Compleat", effectstest.KaldraCompleatOracle)
		if host := effectstest.AttachmentHostOf(t, g, kaldra); host.ID != germ {
			t.Fatalf("Kaldra Compleat AttachedTo = %+v, want the Germ", host)
		}
		if got := effectstest.EffectivePower(t, g, germ); got != 5 {
			t.Errorf("Germ power %d, want 5", got)
		}
		ab := effectstest.EffectiveAbilities(t, g, germ)
		// Haste is the one that makes living weapon plus Kaldra an
		// attack the turn it lands.
		for _, want := range []string{"first strike", "trample", "indestructible", "haste"} {
			if !effectstest.ContainsString(ab, want) {
				t.Errorf("Germ abilities %v missing %q", ab, want)
			}
		}
	})
}

// The Germ template is what the cards print. Spelled out here because
// a token row is data and a typo in it is invisible everywhere else.
func TestGermTokenTemplateMatchesThePrintedText(t *testing.T) {
	germ := effects.TokenCard(effects.LivingWeaponGermToken)
	if germ.Power != 0 || germ.Toughness != 0 {
		t.Errorf("Germ is %d/%d, want 0/0", germ.Power, germ.Toughness)
	}
	if len(germ.Colors) != 1 || germ.Colors[0] != "B" {
		t.Errorf("Germ colors %v, want black", germ.Colors)
	}
	if !germ.HasSubtype("Phyrexian") || !germ.HasSubtype("Germ") {
		t.Errorf("Germ type line %q, want a Phyrexian Germ", germ.TypeLine)
	}
}
