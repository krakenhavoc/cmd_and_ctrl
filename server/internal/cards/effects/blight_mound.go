package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blight Mound — Enchantment {2}{B} (Secrets of Strixhaven Commander,
// #4834):
//
//	"Attacking Pests you control get +1/+0 and have menace.
//	 Whenever a nontoken creature you control dies, create a 1/1 black
//	 and green Pest creature token with "When this token dies, you gain
//	 1 life.""
//
// The static is two layers: the 7c +1/+0 (Instigator Gang's
// attackingCreaturesYouControlGetPower) and the layer 6 menace grant
// (AttackingCreaturesHave), each narrowed to Pests. "Pests" is a word of
// the source's text (CR 612.2), so the subtype is read through
// TextSubtype. Both declare DependsOnAttackingStatus, so a Pest that
// stops attacking loses them at once.
//
// The dies trigger is Open the Graves's condition: a creature that left
// the battlefield for a graveyard under your control, read from its
// last-known information, and not a token. The Pest is the printed
// PestToken, its dies trigger included.
func init() {
	Register(Spec{
		OracleID:     "c4fe854e-8783-492e-bb2e-9f242ada8a5b",
		Name:         "Blight Mound",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			blightMoundAttackingPests(attackingCreaturesYouControlGetPower(1), "Attacking Pests you control get +1/+0."),
			blightMoundAttackingPests(AttackingCreaturesHave("", nil, "menace"), "Attacking Pests you control have menace."),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := diedCreature(ev, g)
				return ok && leftUnderControlOf(ev, dead) == source.Controller && !IsToken(dead)
			}, "Blight Mound — create a Pest", Do(CreateToken{Template: PestToken(), N: 1})),
		},
	})
}

// blightMoundAttackingPests narrows an attacking-creatures static to
// Pests, labelled with the printed half it is.
func blightMoundAttackingPests(s game.StaticAbility, label string) game.StaticAbility {
	s.Label = label
	base := s.AppliesTo
	s.AppliesTo = func(target *game.Card, g *game.Game, source *game.Card) bool {
		return base(target, g, source) && target.HasSubtype(source.TextSubtype("Pest"))
	}
	return s
}
