package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Emrakul, the Exigent Doom — Legendary Creature — Eldrazi {10}, 12/12:
//
//	"When you cast this spell, untap all lands you control.
//	 Flying, trample
//	 Ward—Sacrifice three permanents.
//	 {3}, Exile this card from your hand: Target land gains "{T}: Add
//	 {C}{C}" until this card is cast from exile. You may cast this card
//	 for as long as it remains exiled."
//
// The cast trigger is WhenYouCastThisSpell (it fires from the stack and
// resolves above the spell, so a countered Emrakul still untaps), and
// untaps every land the CASTER controls as it resolves. Ward is the
// shared ward trigger with a sacrifice-three cost; the three may be any
// permanents, lands included.
//
// Gap, declared as a caveat: the last ability is an activated ability
// from the HAND whose payoff is a land that keeps a granted mana
// ability until this card is cast from exile, and whose second sentence
// is a standing permission to cast the exiled card. The hand-ability
// seam (adding-cards.md, "Abilities from the hand") has no
// "until this card is cast" duration for a granted ability. Leaving it
// out is weaker than printed, never stronger: Emrakul is simply a
// {10} spell you can only cast from hand.
func init() {
	Register(Spec{
		OracleID:     "4421ab7d-6d9b-4edd-b5a0-53a8ed84da6f",
		Name:         "Emrakul, the Exigent Doom",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The {3}, exile-from-hand ability that makes a land tap for {C}{C} and lets you cast Emrakul from exile isn't implemented, so Emrakul can only be cast from your hand.",
		},
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			WhenYouCastThisSpell("Emrakul, the Exigent Doom — untap all lands you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					var lands []game.Card
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.IsLand() && c.Tapped {
							lands = append(lands, c)
						}
					}
					for _, c := range lands {
						if err := (UntapTarget{Target: c.InstanceID}).Apply(ctx.asGroupMember()); err != nil {
							return err
						}
					}
					return nil
				}),
			Ward(WardSacrificeN(3, "three permanents", Permanent()), "Emrakul, the Exigent Doom — ward, sacrifice three permanents"),
		},
	})
}
