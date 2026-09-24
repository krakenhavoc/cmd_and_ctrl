package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Teferi's Talent — Enchantment — Aura {3}{U}{U}:
//
//	"Enchant planeswalker
//	 Enchanted planeswalker has "[−12]: You get an emblem with 'You
//	 may activate loyalty abilities of planeswalkers you control on
//	 any player's turn any time you could cast an instant.'"
//	 Whenever you draw a card, put a loyalty counter on enchanted
//	 planeswalker."
//
// WHAT IS WIRED
//
//   - Enchant planeswalker, through EnchantPlaneswalker, so the Aura
//     falls off a host that stops being one (CR 704.5m).
//   - The draw trigger, once per card drawn. "Enchanted planeswalker"
//     is read off the Aura at RESOLUTION through SourcePermanent, so
//     an Aura destroyed in response still names the planeswalker it
//     was on (CR 608.2h); the counter lands only while that
//     planeswalker is still on the battlefield.
//
// WHAT IS NOT WIRED
//
//   - The granted −12. The emblem it makes is fully expressible since
//     #1275 — it is Teferi, Temporal Archmage's, built by
//     teferiTemporalArchmageEmblem and read through
//     `EmblemSpec.ActivationTimings` — but the ability that makes it
//     lives on ANOTHER permanent, and nothing lets a static grant an
//     activated ability to the object it is attached to
//     (docs/engine-seams.md, "Abilities granted to other permanents").
//     Omitted, not faked: the card is weaker than printed, never
//     stronger. The day that row closes, the grant's effect is
//     `CreateEmblem{Source: <this Aura>}` against an `Emblem:
//     teferiTemporalArchmageEmblem()` declared here, and this
//     caveat goes.
func init() {
	Register(Spec{
		OracleID:     "efc23669-1bbb-487f-94b3-cc07f3cd35f6",
		Name:         "Teferi's Talent",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The enchanted planeswalker doesn't gain the −12 that makes the emblem — only the loyalty counter for each card you draw happens.",
		},
		Targets: EnchantPlaneswalker(),
		Triggered: []game.TriggeredAbility{
			WheneverYouDraw("Teferi's Talent — put a loyalty counter on enchanted planeswalker",
				teferisTalentLoyaltyOnDraw),
		},
	})
}

// teferisTalentLoyaltyOnDraw is "put a loyalty counter on enchanted
// planeswalker". Package-level so the trigger captures nothing.
func teferisTalentLoyaltyOnDraw(g *game.Game, item *game.StackItem) error {
	aura, ok := NewContext(g, item).SourcePermanent()
	if !ok || aura.AttachedTo.Kind != game.TargetCard {
		return nil
	}
	host := aura.AttachedTo.ID
	if z := g.FindCardZoneForEffect(host); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	return g.AddCounterForEffect(host, game.CounterLoyalty, 1)
}
