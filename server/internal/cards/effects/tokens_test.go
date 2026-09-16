package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tokens_test.go — S21 sub-PR 1: token templates carry their own
// keywords and mana abilities, so a token behaves like the card it
// prints as rather than a vanilla body.

func TestTokenTemplatesCarryTheirKeywords(t *testing.T) {
	cases := []struct {
		name  string
		tmpl  game.Card
		wants []string
	}{
		{"Spirit", TokenCard("1/1 colorless Spirit with flying"), []string{"flying"}},
		{"Faerie Rogue", FaerieRogueToken(), []string{"flying"}},
		{"Bird", TokenCard("2/2 colorless Bird with flying"), []string{"flying"}},
		{"Samurai", TokenCard("2/2 colorless Samurai with vigilance"), []string{"vigilance"}},
		{"Wurm (deathtouch half)", PhyrexianWurmDeathtouchToken(), []string{"deathtouch"}},
		{"Wurm (lifelink half)", PhyrexianWurmLifelinkToken(), []string{"lifelink"}},
	}
	for _, tc := range cases {
		for _, kw := range tc.wants {
			if !game.HasKeyword(&tc.tmpl, kw) {
				t.Errorf("%s: missing %q (abilities %v)", tc.name, kw, tc.tmpl.Effective().Abilities)
			}
		}
	}
	// The shared base template stays vanilla — the halves add to it.
	if len(TokenCard("3/3 colorless Phyrexian Wurm artifact").Keywords) != 0 {
		t.Errorf("base Wurm template should carry no keywords")
	}
}

func TestTreasureAndSpawnCarryTheirManaAbilities(t *testing.T) {
	tr := game.ManaAbilitiesForCard(TreasureToken())
	if len(tr) != 1 || !tr[0].TapCost || !tr[0].SacrificeCost || tr[0].Produced != "{W|U|B|R|G}" {
		t.Errorf("Treasure ability = %+v, want {T}+sac for any colour", tr)
	}
	sp := game.ManaAbilitiesForCard(EldraziSpawnToken())
	if len(sp) != 1 || sp[0].TapCost || !sp[0].SacrificeCost || sp[0].Produced != "{C}" {
		t.Errorf("Spawn ability = %+v, want sac-only for {C}", sp)
	}
}

// End to end through the engine: a Treasure on the battlefield can
// be cracked for mana, and cracking it sacrifices it.
func TestCrackingATreasureProducesManaAndSacrifices(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tok := TreasureToken()
	tok.InstanceID = uuid.New()
	tok.Owner, tok.Controller = me.ID, me.ID
	g.Battlefield.PushTop(tok)

	if err := g.ActivateManaAbility(me.ID, tok.InstanceID, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("crack the Treasure: %v", err)
	}
	if g.Battlefield.Contains(tok.InstanceID) {
		t.Errorf("cracked Treasure should have left the battlefield")
	}
	// Five-colour pipe → a colour choice for the controller rather
	// than mana straight into the pool.
	var pending *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceMana && c.Chooser == me.ID {
			pending = c
		}
	}
	if pending == nil {
		t.Fatalf("no colour choice queued for the any-colour Treasure")
	}
	if err := g.ResolveManaChoice(pending.ID, me.ID, "R"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if len(me.ManaPool) != 1 || me.ManaPool[0].Color != "R" {
		t.Errorf("mana pool = %+v, want one {R}", me.ManaPool)
	}
}

// Wurmcoil's dies-trigger now makes one of each half, as printed.
func TestWurmcoilMakesOneDeathtouchAndOneLifelinkWurm(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: "Wurmcoil Engine",
		TypeLine: "Artifact Creature — Phyrexian Wurm",
		OracleID: "d1a60f44-7696-49ee-91fb-cab5b3102962",
		Power:    6, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
	// A destroy from inside the lock doesn't run state checks; the
	// next priority pass drains the harvested dies-trigger.
	passPriorityAroundTable(t, g)

	var deathtouch, lifelink int
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Name != "Phyrexian Wurm" {
			continue
		}
		if game.HasKeyword(c, "deathtouch") {
			deathtouch++
		}
		if game.HasKeyword(c, "lifelink") {
			lifelink++
		}
	}
	if deathtouch != 1 || lifelink != 1 {
		t.Errorf("Wurms: %d deathtouch, %d lifelink; want one of each", deathtouch, lifelink)
	}
}
