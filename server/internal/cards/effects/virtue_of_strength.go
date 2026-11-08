package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Virtue of Strength // Garenbrig Growth — Enchantment {5}{G}{G} //
// Sorcery — Adventure {G} (Wilds of Eldraine, #4471):
//
//	Virtue of Strength — "If you tap a basic land for mana, it produces
//	 three times as much of that mana instead."
//	Garenbrig Growth — "Return target creature or land card from your
//	 graveyard to your hand. (Then exile this card. You may cast the
//	 enchantment later from exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID
// and "#1"). The enchantment is Nyxbloom Ancient's CR 614 replacement on
// the mana produced (ManaProducedBecomes, three times, your taps only),
// narrowed to a producing object that is a land with the basic supertype,
// the way Damping Sphere reads the producer off the event. A nonbasic
// land, a creature or an artifact tapped for mana is left alone, and so is
// mana that is not from a tap (CR 106.12a). With Nyxbloom Ancient out too
// a Forest makes nine; multiplication commutes, so the order is never
// asked (ADR 0013 §5ab). Garenbrig Growth is Raise Dead's return with a
// creature-or-land filter. The Adventure lifecycle is the engine's (CR
// 715.3d).
//
// No simplification.
const virtueOfStrengthOracleID = "31af78dd-e962-4d9a-b696-f3048be03486"

func init() {
	Register(Spec{
		OracleID:     virtueOfStrengthOracleID,
		Name:         "Virtue of Strength",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{virtueOfStrengthTriplesBasics()},
	})
	Register(Spec{
		OracleID:     virtueOfStrengthOracleID + "#1",
		Name:         "Garenbrig Growth",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target creature or land card from your graveyard", YouOwn(), Or(Creature(), Land())),
		OnResolve:    returnTargetFromGraveyardToHand,
	})
}

// virtueOfStrengthTriplesBasics is "if you tap a basic land for mana, it
// produces three times as much of that mana instead": Nyxbloom Ancient's
// replacement, applying only when the producing object is a basic land.
func virtueOfStrengthTriplesBasics() game.ReplacementEffect {
	r := ManaProducedBecomes{Times: 3, Scope: ManaTapsByController,
		Label: "Virtue of Strength: a basic land you tap produces three times as much of that mana"}.Build()
	tripled := r.AppliesTo
	r.AppliesTo = func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
		if !tripled(ev, g, src) {
			return false
		}
		land, ok := g.LookupCardForEffect(ev.ManaSource)
		return ok && land.IsLand() && land.HasSupertype("basic")
	}
	return r
}
