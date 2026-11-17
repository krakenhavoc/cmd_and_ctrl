package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_one_b18_helpers.go — shared vocabulary for the S82 ONE batch 18
// cards (Melira, the Living Cure and the other Legends).

// meliraReturnListedBody is the body of Melira's delayed trigger: "return
// that card to the battlefield under its owner's control". The listed
// card is returned only if it is still in a graveyard (CR 603.7c), to the
// battlefield under its owner's control (ReturnFromGraveyard's zero
// Controller).
var meliraReturnListedBody = game.SimpleDelayedBody("melira/return-listed-to-the-battlefield", meliraReturnListedToBattlefield)

func meliraReturnListedToBattlefield(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard {
			continue
		}
		if z := g.FindCardZoneForEffect(t.ID); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// s82OneB18AddArtifactType is "<it> is an artifact in addition to its other
// types" (CR 205.1b): the layer 4 Apply of Encroaching Mycosynth, and
// idempotent.
func s82OneB18AddArtifactType(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
	for _, t := range c.Types {
		if t == "Artifact" {
			return
		}
	}
	c.Types = append(c.Types, "Artifact")
}

// s82OneB18NonlandPermanentType reports whether a card or spell has a
// permanent card type (CR 110.4) and isn't a land: the "nonland permanent"
// of "nonland permanent cards" and "permanent spells" (a land is never a
// spell, CR 305.9).
func s82OneB18NonlandPermanentType(c *game.Card) bool {
	if c.HasCardType("land") {
		return false
	}
	for _, t := range []string{"artifact", "creature", "enchantment", "planeswalker", "battle"} {
		if c.HasCardType(t) {
			return true
		}
	}
	return false
}

// s82OneB18IsBasicSphereOrLocusLand is "a basic, Sphere, or Locus land":
// a land that is basic or has the land type Sphere or Locus (CR 205.3i).
func s82OneB18IsBasicSphereOrLocusLand(c game.Card) bool {
	if !c.IsLand() {
		return false
	}
	return IsBasicLand(c) || c.HasSubtypeWord("Sphere") || c.HasSubtypeWord("Locus")
}

// s82OneB18NineDifferentlyNamedLands is Monument to Perfection's gate:
// nine or more lands with different names among the basic, Sphere, and
// Locus lands `controller` controls, read as the ability is activated.
func s82OneB18NineDifferentlyNamedLands(g *game.Game, controller, _ uuid.UUID) bool {
	names := map[string]bool{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && s82OneB18IsBasicSphereOrLocusLand(c) {
			names[c.Name] = true
		}
	}
	return len(names) >= 9
}

// riaIvorMitesBody is the follow-up of Ria Ivor's shield: "If damage is
// prevented this way, create that many 1/1 colorless Phyrexian Mite
// artifact creature tokens with toxic 1 and 'This token can't block.'"
// The prevented amount rides the body (CR 615.5), and the tokens are
// created by the shield's controller. An on-disk key.
var riaIvorMitesBody = game.DelayedBody("prevention/ria-ivor-create-mites", riaIvorMites)

func riaIvorMites(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	if p.Amount <= 0 {
		return nil
	}
	return CreateToken{Template: PhyrexianMiteToken(), N: p.Amount}.Apply(NewContext(g, item))
}
