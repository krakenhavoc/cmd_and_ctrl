package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// General's Enforcer — Creature — Human Soldier {W}{B}, 2/3 (Fallout,
// #5948):
//
//	"Legendary Humans you control have indestructible.
//	 {2}{W}{B}: Exile target card from a graveyard. If it was a
//	 creature card, create a 1/1 white Human Soldier creature token."
//
// The first line is a layer 6 grant over the legendary Human creatures
// its controller controls, the Enforcer itself only if it is made one. The
// activated ability is Cling to Dust's exile whose clause waits for the
// card to have reached exile (ExileThenIfItWas): the type is read as it was
// before the exile, and a card a replacement kept in its graveyard makes
// no token. The token is the catalog's 1/1 white Human Soldier.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8e953251-aa5c-4aa3-8e69-1ff1f609f3ff",
		Name:         "General's Enforcer",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			Labelled(KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller &&
					target.IsLegendary() && target.HasSubtype(source.TextSubtype("Human"))
			}, "indestructible"), "Legendary Humans you control have indestructible."),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{W}{B}: Exile target card from a graveyard. If it was a creature card, create a 1/1 white Human Soldier creature token.",
			Cost:    ManaCost("{2}{W}{B}"),
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
			Targets: TargetCardInGraveyard("target card from a graveyard"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return ExileThenIfItWas{Target: t.ID, Was: WasCreatureCard, Keyed: generalsEnforcerToken}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}

// generalsEnforcerToken is "if it was a creature card, create a 1/1 white
// Human Soldier creature token": the Enforcer's controller makes it. An
// on-disk key: never renamed, never reused.
var generalsEnforcerToken = IfItWasTail("if-it-was/generals-enforcer-token", func(ctx *Context) error {
	return CreateToken{Template: TokenCard("1/1 white Human Soldier"), N: 1}.Apply(ctx)
}, nil)
