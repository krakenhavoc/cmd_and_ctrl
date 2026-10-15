package protocol

import (
	"testing"

	"github.com/google/uuid"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // Purphoros, Impact Tremors
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/oracletext"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ability_text_view_test.go — #3830: a stack item carries its ability's
// printed text in `ability_text`, the label stays the engine's, and a
// source the table cannot see has none.

const (
	purphorosOracle     = "4fdbbec2-e921-4b63-958d-f9ba1e417197"
	impactTremorsOracle = "9242cd3e-1a71-4700-8182-9c1005616033"

	purphorosTrigger = "Whenever another creature you control enters, Purphoros deals 2 damage to each opponent."
	purphorosPump    = "{2}{R}: Creatures you control get +1/+0 until end of turn."
	tremorsTrigger   = "Whenever a creature you control enters, this enchantment deals 1 damage to each opponent."
)

// useAbilityTextSource answers oracletext with the two cards' printed
// text, as the Scryfall index does in production.
func useAbilityTextSource(t *testing.T) {
	t.Helper()
	texts := map[string]oracletext.Card{
		purphorosOracle: {Name: "Purphoros, God of the Forge", Faces: []oracletext.Face{{
			Name: "Purphoros, God of the Forge",
			Text: "Indestructible\nAs long as your devotion to red is less than five, Purphoros isn't a creature.\n" +
				purphorosTrigger + "\n" + purphorosPump,
		}}},
		impactTremorsOracle: {Name: "Impact Tremors", Faces: []oracletext.Face{{Name: "Impact Tremors", Text: tremorsTrigger}}},
	}
	oracletext.SetSource(func(id uuid.UUID) (oracletext.Card, bool) {
		c, ok := texts[id.String()]
		return c, ok
	})
	t.Cleanup(func() { oracletext.SetSource(nil) })
}

// stackItemFrom is the item (on the stack or waiting) from a source.
func stackItemFrom(t *testing.T, v GameView, source uuid.UUID) StackItemView {
	t.Helper()
	for _, it := range append(append([]StackItemView(nil), v.StackItems...), v.PendingTriggers...) {
		if it.SourceCardID == source.String() {
			return it
		}
	}
	t.Fatalf("no stack item from %s", source)
	return StackItemView{}
}

func TestAbilityTextOfTriggeredAbilities(t *testing.T) {
	useAbilityTextSource(t)
	g := buildMainPhaseGame(t)
	me := g.Seats[0]
	purphoros := catalogPermanent(g, me.ID, "Purphoros, God of the Forge", "Legendary Enchantment Creature — God", purphorosOracle)
	tremors := catalogPermanent(g, me.ID, "Impact Tremors", "Enchantment", impactTremorsOracle)

	enterCreature(t, g, me.ID, "Grizzly Bears")

	for _, c := range []struct {
		source      uuid.UUID
		label, text string
	}{
		{purphoros, "Purphoros, God of the Forge — 2 damage to each opponent", purphorosTrigger},
		{tremors, "Impact Tremors — 1 damage to each opponent", tremorsTrigger},
	} {
		// Every seat reads the same text: the source is public.
		for _, viewer := range g.Seats {
			it := stackItemFrom(t, FilterViewFor(ViewOfGame(g), viewer.ID.String()), c.source)
			if it.AbilityText != c.text {
				t.Errorf("%s to %s: ability_text = %q, want %q", c.label, viewer.Name, it.AbilityText, c.text)
			}
			if it.Label != c.label {
				t.Errorf("label = %q, want the engine's %q unchanged", it.Label, c.label)
			}
		}
	}
}

func TestAbilityTextOfAnActivatedAbilityIsItsCostLine(t *testing.T) {
	useAbilityTextSource(t)
	g := buildMainPhaseGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	purphoros := catalogPermanent(g, me.ID, "Purphoros, God of the Forge", "Legendary Enchantment Creature — God", purphorosOracle)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, purphoros, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := stackItemFrom(t, ViewOfGame(g), purphoros).AbilityText; got != purphorosPump {
		t.Errorf("ability_text = %q, want %q", got, purphorosPump)
	}
}

// An item with no catalog row of its source's own (here one announced
// by hand, the sandbox shape; a delayed trigger, a copy and a granted
// row are the same) keeps only its label, which the client shows.
func TestAbilityTextFallsBackWithNoRowOfItsOwn(t *testing.T) {
	useAbilityTextSource(t)
	g := buildMainPhaseGame(t)
	me := g.Seats[0]
	purphoros := catalogPermanent(g, me.ID, "Purphoros, God of the Forge", "Legendary Enchantment Creature — God", purphorosOracle)
	const label = "Purphoros, God of the Forge — 2 damage to each opponent"
	if err := g.AnnounceTrigger(me.ID, purphoros, game.AbilityParams{Label: label}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}
	it := stackItemFrom(t, ViewOfGame(g), purphoros)
	if it.AbilityText != "" || it.Label != label {
		t.Errorf("item = label %q ability_text %q; want the label and no text", it.Label, it.AbilityText)
	}
}

// A face-down source never reveals its text, to anyone.
func TestAbilityTextOfAFaceDownSourceIsNone(t *testing.T) {
	useAbilityTextSource(t)
	g := buildMainPhaseGame(t)
	me := g.Seats[0]
	tremors := catalogPermanent(g, me.ID, "Impact Tremors", "Enchantment", impactTremorsOracle)
	enterCreature(t, g, me.ID, "Grizzly Bears")
	if got := stackItemFrom(t, ViewOfGame(g), tremors).AbilityText; got != tremorsTrigger {
		t.Fatalf("face up: ability_text = %q, want %q", got, tremorsTrigger)
	}

	g.WithWriteLock(func() {
		c, ok := g.LookupCardPtrForEffect(tremors)
		if !ok {
			t.Fatal("Impact Tremors left the battlefield")
		}
		c.FaceDown, c.FaceDownKind = true, game.FaceDownManifested
	})
	for _, viewer := range append([]string{""}, me.ID.String(), g.Seats[1].ID.String()) {
		if got := stackItemFrom(t, FilterViewFor(ViewOfGame(g), viewer), tremors).AbilityText; got != "" {
			t.Errorf("face down, viewer %q: ability_text = %q, want none", viewer, got)
		}
	}
}
