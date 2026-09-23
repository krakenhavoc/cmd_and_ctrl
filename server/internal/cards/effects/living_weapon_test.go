package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// living_weapon_test.go pins the keyword itself rather than the three
// cards that carry it. What is worth asserting is exactly the part
// that would break silently:
//
//   - the Germ is made AND the Equipment lands on it, in one
//     resolution, so the 0/0 is never a 0/0 with nothing on it when
//     state-based actions next run;
//   - the pump the host gets is the Equipment's, so the same helper
//     really is serving three different cards;
//   - the Germ dies the moment the Equipment leaves, which is the
//     other half of the same rule and is what Batterskull's {3}
//     bounce does on purpose.

const (
	batterskullOracle    = "d12e5ce0-5705-4c35-9a93-b883db52c80c"
	batterboneOracle     = "389d459d-446a-4b84-82b2-a30fe6ced11f"
	kaldraCompleatOracle = "7359e82b-db79-488d-a1d4-75a00f12a4cf"
)

// castLivingWeapon casts the named Equipment from the active seat's
// hand and settles the spell and the living weapon trigger it queues.
// Returns the Equipment and the Germ it made.
func castLivingWeapon(t *testing.T, g *game.Game, name, oracle string) (equipment, germ uuid.UUID) {
	t.Helper()
	equipment = castCatalogSpell(t, g, name, equipTypeLine, oracle, nil)
	passPriorityAroundTable(t, g)
	germ = findBattlefieldByName(g, "Phyrexian Germ")
	if germ == uuid.Nil {
		t.Fatalf("%s made no Germ token", name)
	}
	return equipment, germ
}

func TestBatterskullMakesAGermAndAttachesItselfToIt(t *testing.T) {
	g := newCatalogGame(t)
	skull, germ := castLivingWeapon(t, g, "Batterskull", batterskullOracle)

	if host := attachmentHostOf(t, g, skull); host.Kind != game.TargetCard || host.ID != germ {
		t.Fatalf("Batterskull AttachedTo = %+v, want the Germ %s", host, germ)
	}
	// A 0/0 that survives is the whole point: the +4/+4 arrived in
	// the same resolution, so the CR 704.5f check never saw a 0/0.
	if got := effectivePower(t, g, germ); got != 4 {
		t.Errorf("Germ power %d, want 4", got)
	}
	if got := effectiveToughness(t, g, germ); got != 4 {
		t.Errorf("Germ toughness %d, want 4", got)
	}
	ab := effectiveAbilities(t, g, germ)
	for _, want := range []string{"vigilance", "lifelink"} {
		if !containsString(ab, want) {
			t.Errorf("Germ abilities %v missing %q", ab, want)
		}
	}
}

// The Germ is a 0/0 and stays alive only while something is on it.
// Bouncing Batterskull with its own {3} is the printed way to prove
// that, and it exercises the bounce ability at the same time.
func TestBatterskullsBounceLeavesTheGermToDieAsA00(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	skull, germ := castLivingWeapon(t, g, "Batterskull", batterskullOracle)

	// Ability 0 is the bounce; ability 1 is equip.
	if err := g.ActivateCatalogAbility(me.ID, skull, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("the {3} bounce: %v", err)
	}
	passPriorityAroundTable(t, g)

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
		g := newCatalogGame(t)
		bone, germ := castLivingWeapon(t, g, "Batterbone", batterboneOracle)
		if host := attachmentHostOf(t, g, bone); host.ID != germ {
			t.Fatalf("Batterbone AttachedTo = %+v, want the Germ", host)
		}
		if got := effectivePower(t, g, germ); got != 1 {
			t.Errorf("Germ power %d, want 1", got)
		}
	})

	t.Run("Kaldra Compleat", func(t *testing.T) {
		g := newCatalogGame(t)
		kaldra, germ := castLivingWeapon(t, g, "Kaldra Compleat", kaldraCompleatOracle)
		if host := attachmentHostOf(t, g, kaldra); host.ID != germ {
			t.Fatalf("Kaldra Compleat AttachedTo = %+v, want the Germ", host)
		}
		if got := effectivePower(t, g, germ); got != 5 {
			t.Errorf("Germ power %d, want 5", got)
		}
		ab := effectiveAbilities(t, g, germ)
		// Haste is the one that makes living weapon plus Kaldra an
		// attack the turn it lands.
		for _, want := range []string{"first strike", "trample", "indestructible", "haste"} {
			if !containsString(ab, want) {
				t.Errorf("Germ abilities %v missing %q", ab, want)
			}
		}
	})
}

// The Germ template is what the cards print. Spelled out here because
// a token row is data and a typo in it is invisible everywhere else.
func TestGermTokenTemplateMatchesThePrintedText(t *testing.T) {
	germ := TokenCard(livingWeaponGermToken)
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
