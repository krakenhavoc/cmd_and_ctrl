package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hydro-Man, Fluid Felon — Legendary Creature — Elemental Villain {U}{U},
// 2/2 (Marvel's Spider-Man, #4460):
//
//		"Whenever you cast a blue spell, if Hydro-Man is a creature, he gets
//		 +1/+1 until end of turn.
//		 At the beginning of your end step, untap Hydro-Man. Until your next
//		 turn, he becomes a land and gains "{T}: Add {U}." (He's not a creature
//		 during that time.)"
//
//	  - The first trigger has an intervening "if" (CR 603.4): Hydro-Man's
//	    being a creature is read through the layers as the spell is cast and
//	    again as the trigger resolves, so a cast while he is a land does
//	    nothing, and one that finds him turned back into a creature on
//	    resolution does. The colour is the spell's, read through the
//	    source's text (CR 612.2).
//	  - The end-step ability untaps him, then registers one layer 4 / layer 6
//	    effect until your next turn (CR 611.2a, pinned to the object, CR
//	    400.7): "becomes a land" with no "in addition" REPLACES his card
//	    types (CR 205.1a), so he is no longer a creature and his creature
//	    subtypes go with the type (Kitesail Larcenist's SetTypesMod); his
//	    supertype stays (CR 205.4b). The granted mana ability is an ordinary
//	    one for a land, so summoning sickness does not stop it while he is
//	    not a creature. The effect ends as your next turn begins, before the
//	    untap step, so he untaps as a creature again.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a93a8edf-2879-4cae-b3b9-b46298270823",
		Name:         "Hydro-Man, Fluid Felon",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: hydroManManaGrant,
			Mana: []ManaAbility{{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{U}",
				Label:    "{T}: Add {U}",
			}},
			Text: "{T}: Add {U}.",
		}},
		Triggered: []game.TriggeredAbility{
			InterveningIf(hydroManIsACreature,
				On(game.EventCast, AllOf(YouCast(OfColor("U")), hydroManSourceIsACreature),
					"Hydro-Man, Fluid Felon — he gets +1/+1 until end of turn",
					thisGetsUntilEndOfTurn(1, 1, "Hydro-Man — +1/+1 until end of turn"))),
			AtYourEndStep("Hydro-Man, Fluid Felon — untap him; until your next turn he becomes a land and gains {T}: Add {U}",
				hydroManBecomesALand),
		},
	})
}

const hydroManManaGrant = "hydro-man-fluid-felon/land-mana"

// hydroManSourceIsACreature is the trigger-time half of the intervening if.
func hydroManSourceIsACreature(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return source.IsCreature()
}

// hydroManIsACreature is the resolution-time half.
var hydroManIsACreature = CondBoard{
	Name: "Hydro-Man is a creature",
	Test: func(ctx *Context) bool {
		c, ok := ctx.Game.LookupCardForEffect(ctx.Source())
		return ok && onBattlefield(ctx.Game, ctx.Source()) && c.IsCreature()
	},
}

// hydroManBecomesALand untaps Hydro-Man and turns him into a mana land
// until his controller's next turn.
func hydroManBecomesALand(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	src := ctx.Source()
	if !onBattlefield(g, src) {
		return nil
	}
	if err := (UntapTarget{Target: src}).Apply(ctx); err != nil {
		return err
	}
	return ScopedEffectFor{
		Target: src,
		Mods: []game.Mod{
			game.SetTypesMod([]string{"Land"}, nil),
			game.GrantAbilitiesMod(hydroManManaGrant),
		},
		Duration: DurationUntilYourNextTurn(ctx, item.Controller),
		Label:    "Hydro-Man — becomes a land until your next turn",
	}.Apply(ctx)
}
