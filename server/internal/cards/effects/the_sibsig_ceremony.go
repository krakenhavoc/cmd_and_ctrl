package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Sibsig Ceremony — Legendary Enchantment {B}{B}{B} (Tarkir:
// Dragonstorm, #4463):
//
//	"Creature spells you cast cost {2} less to cast.
//	 Whenever a creature you control enters, if you cast it, destroy that
//	 creature, then create a 2/2 black Zombie Druid creature token."
//
// The reduction is Duskwatch Recruiter's shape at {2}: generic mana only,
// never below zero. The trigger is Alex Wilder's read of the entering
// creature's cast record (CR 603.4's intervening "if": a creature put onto
// the battlefield without being cast does not trigger): a creature entering
// under your control that its controller cast. It destroys that creature as
// it is now (a creature that left, or became a new object, is not
// destroyed, CR 400.7), then creates the Zombie Druid either way: the token
// is the second half of the sentence, not a consequence of the destruction
// (an indestructible creature is still replaced by the token). The enchantment
// itself is no creature, so it never triggers off itself.
//
// No simplification.
func init() {
	const label = "The Sibsig Ceremony — destroy that creature, then create a 2/2 black Zombie Druid creature token"
	Register(Spec{
		OracleID:     "bbdb731d-c533-48ef-b82a-cf1a8ba36208",
		Name:         "The Sibsig Ceremony",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		CostModifiers: []game.CostModifier{
			CostsLess(2, "Creature spells you cast cost {2} less to cast.", YourSpell(), CreatureSpell()),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, sibsigCeremonyCreatureYouCast, label, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if info, ok := ctx.TriggeringPermanent(); ok && !info.Left {
					if err := (DestroyTarget{Target: ctx.Trigger().Object.Ref().ID}).Apply(ctx); err != nil {
						return err
					}
				}
				return CreateToken{Template: TokenCard("2/2 black Zombie Druid"), N: 1}.Apply(ctx)
			}),
		},
	})
}

// sibsigCeremonyCreatureYouCast is the trigger and its intervening "if":
// a creature entered under the source's controller's control and its
// controller cast it.
func sibsigCeremonyCreatureYouCast(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	if !ok || !c.IsCreature() {
		return false
	}
	cast, known := c.CastByItsController()
	return known && cast && c.Provenance.FromZone != ""
}
