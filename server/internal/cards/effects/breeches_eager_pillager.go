package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// breechesEagerPillagerLabel is the trigger's stack label, and with it
// the key its "hasn't been chosen this turn" memory is kept under.
const breechesEagerPillagerLabel = "Breeches, Eager Pillager — a Pirate attacked"

// Breeches, Eager Pillager — Legendary Creature — Goblin Pirate
// {2}{R}, 3/3:
//
//	"First strike
//	 Whenever a Pirate you control attacks, choose one that hasn't been
//	 chosen this turn —
//	 • Create a Treasure token.
//	 • Target creature can't block this turn.
//	 • Exile the top card of your library. You may play it this turn."
//
// The pirate commander of #1112. The trigger is per Pirate ("a Pirate
// you control attacks", not "one or more"), so an attack with three
// Pirates is three triggers, and ChooseOneNotChosenThisTurn (ADR 0097)
// makes them take three different bullets. The 2023-11-10 ruling's
// fourth Pirate — "that instance of the ability is removed from the
// stack with no effect" — is the engine's exhausted-trigger drop, and
// two Breeches keep separate memories because the memory belongs to
// each Breeches' ability.
//
// "A Pirate you control" reads the attacker's effective subtypes, so a
// changeling counts and Breeches' own attack triggers it. "Can't block
// this turn" is the restriction bit for the rest of the turn; "you may
// play it this turn" is the impulse-exile grant, which lets a land be
// played as well as a spell cast.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "15361770-c6dc-4db7-b745-aa03e4b866ea",
		Name:            "Breeches, Eager Pillager",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			breechesEagerPillagerTrigger(),
		},
	})
}

func breechesEagerPillagerTrigger() game.TriggeredAbility {
	t := On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if !attackDeclaredByYou(ev, source.Controller) {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && c.HasSubtype("Pirate")
	}, breechesEagerPillagerLabel, func(_ *game.Game, _ *game.StackItem) error { return nil })
	t.Modes = ChooseOneNotChosenThisTurn(
		ModeDoing("Create a Treasure token.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return CreateToken{Controller: item.Controller, Template: TreasureToken(), N: 1}.Apply(ctx)
			}),
		ModeDoing("Target creature can't block this turn.", TargetCreature("target creature"),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				return RestrictUntilEOT{Target: t.ID, Restrictions: game.CantBlock}.Apply(ctx)
			}),
		ModeDoing("Exile the top card of your library. You may play it this turn.", nil,
			exileTopCardYouMayPlayThisTurn),
	)
	return t
}
