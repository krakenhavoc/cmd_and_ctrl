package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Feral Appetite — Enchantment {2}{G} (Secrets of Strixhaven Commander,
// #4834):
//
//	"Attacking Pests you control get +1/+0 and have deathtouch.
//	 {1}{G}: Exile target card from a graveyard. If a creature card is
//	 exiled this way, create a 1/1 black and green Pest creature token
//	 with "When this token dies, you gain 1 life.""
//
// The static is Blight Mound's, with deathtouch for menace: the layer 7c
// +1/+0 and the layer 6 grant, each narrowed to attacking Pests you
// control (blightMoundAttackingPests), so a Pest that stops attacking
// loses both at once.
//
// The activated ability is Deluge of the Dead's: the card's type is read
// before the exile, and the Pest waits for the exile and follows only a
// creature card that really left its graveyard (ExileThenIfItWas). Any
// graveyard, any card. The Pest is the printed PestToken, its dies
// trigger included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "adf731a1-fc4d-4260-bb85-535edb1d82e9",
		Name:         "Feral Appetite",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			blightMoundAttackingPests(attackingCreaturesYouControlGetPower(1), "Attacking Pests you control get +1/+0."),
			blightMoundAttackingPests(AttackingCreaturesHave("", nil, "deathtouch"), "Attacking Pests you control have deathtouch."),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{G}: Exile target card from a graveyard. If a creature card is exiled this way, create a 1/1 black and green Pest creature token.",
			Cost:    ManaCost("{1}{G}"),
			Targets: TargetCardInGraveyard("target card from a graveyard"),
			// #4559: priced by what the exile takes away; the token it
			// makes for a creature card is left unpriced.
			Purpose: ForTargets(ExilesFromGraveyard(0, "")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if err := (ExileThenIfItWas{
						Target: t.ID,
						Was:    WasCreatureCard,
						Keyed:  feralAppetiteIfCreature,
					}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}

// feralAppetiteIfCreature is the exile's keyed "if a creature card is
// exiled this way, create a Pest": the ability rides the tail, so the
// token is its controller's. An on-disk key: never renamed, never reused.
var feralAppetiteIfCreature = IfItWasTail("if-it-was/feral-appetite-pest", func(ctx *Context) error {
	return CreateToken{Template: PestToken(), N: 1}.Apply(ctx)
}, nil)
