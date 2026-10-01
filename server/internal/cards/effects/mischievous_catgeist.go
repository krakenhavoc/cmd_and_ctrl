package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mischievous Catgeist // Catlike Curiosity (#1855, ADR 0107 §4) — a
// disturb card whose back face is an Aura.
//
// Front face, Creature — Cat Spirit {1}{U}, 1/1:
//
//	"Whenever this creature deals combat damage to a player, draw a
//	 card.
//	 Disturb {2}{U} (You may cast this card from your graveyard
//	 transformed for its disturb cost.)"
//
// Back face, Enchantment — Aura:
//
//	"Enchant creature
//	 Enchanted creature has "Whenever this creature deals combat damage
//	 to a player, draw a card."
//	 If Catlike Curiosity would be put into a graveyard from anywhere,
//	 exile it instead."
//
// The Aura's quoted ability is a granted bundle (ADR 0093): the
// enchanted creature has the trigger, so its controller draws. The
// trigger is the front face's own, so the two faces cannot drift.
// Disturb is effects.Disturb; the exile clause is DisturbedExile.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         mischievousCatgeistOracleID,
		Name:             "Mischievous Catgeist",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Disturb("{2}{U}")},
		Triggered:        []game.TriggeredAbility{catgeistDraw("Mischievous Catgeist — draw a card")},
	})
	Register(Spec{
		OracleID:     mischievousCatgeistOracleID + "#1",
		Name:         "Catlike Curiosity",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Grants: []AbilityGrant{{
			Key:       catlikeCuriosityDraw,
			Triggered: []game.TriggeredAbility{catgeistDraw("Catlike Curiosity — draw a card")},
			Text:      "Whenever this creature deals combat damage to a player, draw a card.",
		}},
		Static:       []game.StaticAbility{GrantAbilitiesToAttached(catlikeCuriosityDraw)},
		Replacements: []game.ReplacementEffect{DisturbedExile("Catlike Curiosity")},
	})
}

const (
	mischievousCatgeistOracleID = "4d0a0027-53b3-45a1-8736-f0ac86b19342"
	catlikeCuriosityDraw        = "catlike-curiosity/draw"
)

// catgeistDraw is "Whenever this creature deals combat damage to a
// player, draw a card", printed on the front face and quoted on the
// back.
func catgeistDraw(label string) game.TriggeredAbility {
	return WheneverThisDealsCombatDamageToAPlayer(label, Do(DrawCards{N: 1}))
}
