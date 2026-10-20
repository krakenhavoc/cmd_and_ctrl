package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sarkhan the Masterless — Legendary Planeswalker — Sarkhan {3}{R}{R},
// loyalty 5:
//
//	"Whenever a creature attacks you or a planeswalker you control, each
//	 Dragon you control deals 1 damage to that creature.
//	 +1: Until end of turn, each planeswalker you control becomes a 4/4
//	     red Dragon creature and gains flying.
//	 −3: Create a 4/4 red Dragon creature token with flying."
//
// The static trigger fires once per attacking creature whose attack
// names Sarkhan's controller or a planeswalker they control (a battle
// they protect is neither). On resolution each Dragon its controller
// controls THEN deals 1 damage to that creature, each Dragon its own
// source (so deathtouch, lifelink and source prevention see each), all
// in one damage instance. A creature that left combat or the
// battlefield is skipped.
//
// +1 locks the planeswalkers it changes as it resolves (CR 611.2c): a
// walker that arrives later stays a walker. Each becomes a creature in
// addition to its types (CR 205.1b), adds the Dragon subtype, is red
// (replacing its colours, layer 5) and has base 4/4 with flying. It is
// still a planeswalker, so damage dealt to it removes loyalty as well
// as being marked (CR 120.3c, 120.3d). A Sarkhan animated on its own
// turn is a Dragon too, so its trigger counts it.
//
// No simplification. Guffster's Party deck request (#4184).
func init() {
	Register(Spec{
		OracleID:     "67dbc5dd-6d05-43e7-83f7-70e10eb54807",
		Name:         "Sarkhan the Masterless",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, sarkhanAttacksYouOrYourPlaneswalker,
				"Sarkhan the Masterless — each Dragon you control deals 1 damage to the attacking creature",
				sarkhanDragonsPing),
		},
		Activated: []ActivatedAbility{
			{
				Label:  "+1: Until end of turn, each planeswalker you control becomes a 4/4 red Dragon creature and gains flying.",
				Cost:   LoyaltyCost(1),
				Effect: sarkhanAnimateWalkers,
			},
			{
				Label:  "−3: Create a 4/4 red Dragon creature token with flying.",
				Cost:   LoyaltyCost(-3),
				Effect: Do(CreateToken{Template: TokenCard("4/4 red Dragon with flying"), N: 1}),
			},
		},
	})
}

// sarkhanAttacksYouOrYourPlaneswalker is "a creature attacks you or a
// planeswalker you control".
func sarkhanAttacksYouOrYourPlaneswalker(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventAttack || ev.Target == uuid.Nil || ev.CardID == uuid.Nil {
		return false
	}
	if ev.Target == source.Controller {
		return true
	}
	pw, ok := g.LookupCardForEffect(ev.Target)
	return ok && pw.IsPlaneswalker() && pw.Controller == source.Controller && g.Battlefield.Contains(ev.Target)
}

// sarkhanDragonsPing is the trigger's resolution: every Dragon its
// controller controls deals 1 damage to the attacker.
func sarkhanDragonsPing(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	info, ok := ctx.TriggeringPermanent()
	if !ok || info.Left {
		return nil
	}
	attacker := ctx.Trigger().Event.CardID
	dragon := ctx.TextSubtype("Dragon")
	dragons := MatchingBattlefield(ctx, And(YouControl(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.HasSubtype(dragon)
	}))
	if len(dragons) == 0 {
		return nil
	}
	return g.DamageInstanceForEffect(func() error {
		for _, d := range dragons {
			if err := (DealDamage{Source: d.InstanceID, Target: attacker, Amount: 1}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}

// sarkhanAnimateWalkers is +1: the planeswalkers its controller controls
// now become 4/4 red Dragon creatures with flying until end of turn.
func sarkhanAnimateWalkers(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return ScopedEffectFor{
		Match: And(Planeswalker(), YouControl()),
		Mods: []game.Mod{
			game.AddTypesMod("Creature"),
			game.AddSubtypesMod(ctx.TextSubtypes("Dragon")...),
			game.SetColorsMod(ctx.TextColors("R")...),
			game.SetBasePowerMod(4),
			game.SetBaseToughnessMod(4),
			game.AddKeywordsMod("flying"),
		},
		Duration: DurationUntilEndOfTurn(ctx),
		Label:    "Sarkhan the Masterless — a 4/4 red Dragon creature with flying until end of turn",
	}.Apply(ctx)
}
