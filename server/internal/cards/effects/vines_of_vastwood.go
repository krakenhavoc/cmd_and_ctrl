package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vines of Vastwood — Instant {G} (EDHREC rank 4048):
//
//	"Kicker {G} (You may pay an additional {G} as you cast this
//	 spell.)
//	 Target creature can't be the target of spells or abilities your
//	 opponents control this turn. If this spell was kicked, that
//	 creature gets +4/+4 until end of turn."
//
// One mana to blank a removal spell, two to blank it and win the
// combat. It is the cheapest protection spell green has and the
// reason a Voltron deck can hold up a single Forest.
//
// # What the first clause is
//
// "Can't be the target of spells or abilities your opponents control"
// is NOT hexproof (#3860). Hexproof (CR 702.11b) refuses what the
// CREATURE'S CONTROLLER's opponents control; this clause refuses what
// the Vines CASTER's opponents control. The two agree when the Vines
// go on your own creature, which is nearly every game, and come apart
// when they go on somebody else's: the printed card then locks the
// creature's own controller out of targeting it, and hexproof would
// not. The first cut of this card narrowed the target to "a creature
// you control" so hexproof was the printed effect word for word, and
// shipped weaker than printed (the Vines on an opponent's creature,
// to stop that opponent sacrificing or bouncing it in response, or on a
// creature you are about to steal, were not castable).
//
// So the target is any creature, and the clause is its own continuous
// record (CantBeTargetedByYourOpponentsUntilEOT, game/targeting_bar.go),
// read live by the targeting gate: the creature's controller can still
// target it if they are the caster, an opponent's board wipe still kills
// it because a wipe targets nothing, and a spell that had ALREADY
// targeted it before the Vines resolved fizzles at resolution when the
// re-check finds the target illegal (CR 608.2b), as the card's ruling
// says. A hexproof waiver (Nowhere to Run) does not lift it: it waives
// hexproof and this is not hexproof.
//
// # Kicker
//
// Kicker {G} is the ordinary optional additional cost (CR 702.33):
// OptionalCosts: Kicker("{G}"), read back at resolution with
// ctx.WasKicked(). A kicked Vines gives the same creature +4/+4 until
// end of turn after the protection.
func init() {
	Register(Spec{
		OracleID:      "8998d211-5ca6-41d2-bc64-f51f70bd41e5",
		Name:          "Vines of Vastwood",
		Completeness:  CompletenessFull,
		Targets:       TargetCreature("target creature"),
		OptionalCosts: []game.AdditionalCost{Kicker("{G}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			if err := (CantBeTargetedByYourOpponentsUntilEOT{
				Target: item.Targets[0].ID,
				Label:  "Vines of Vastwood — can't be the target of spells or abilities your opponents control",
			}).Apply(ctx); err != nil {
				return err
			}
			if !ctx.WasKicked() {
				return nil
			}
			return BoostUntilEOT{Target: item.Targets[0].ID, Power: 4, Toughness: 4, Label: "Vines of Vastwood — kicked: +4/+4"}.Apply(ctx)
		},
	})
}
